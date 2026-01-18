package runner

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/windlant/go-stress-test/internal/target"
)

type Result struct {
	TotalRequests int
	Successful    int
	Failed        int
	SuccessRate   float64
	AvgLatencyMs  float64
	MinLatencyMs  float64
	MaxLatencyMs  float64
	P95LatencyMs  float64
	P99LatencyMs  float64
	RPS           float64
	Duration      time.Duration
}

func Run(
	tgt *target.HTTPTarget,
	rate, concurrency int,
	duration time.Duration,
	total int,
	quiet bool,
) (*Result, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	if duration > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), duration)
	} else if total > 0 {
		ctx, cancel = context.WithCancel(context.Background())
	} else {
		return nil, fmt.Errorf("either duration or total must be specified")
	}
	defer cancel()

	startTime := time.Now()

	// 共享状态
	var mu sync.Mutex
	totalReqs := 0
	success := 0
	failed := 0
	latencies := []time.Duration{}
	sentCount := 0

	// Worker 就绪计数（用于实时打印）
	var readyWorkers int64 = 0
	// 启动信号：关闭即开始
	startSignal := make(chan struct{})

	// 可选限流器
	var limiterChan <-chan time.Time
	if rate > 0 {
		limiter := time.NewTicker(time.Second / time.Duration(rate))
		defer limiter.Stop()
		limiterChan = limiter.C
	}

	// 实时打印（非 quiet 模式）
	var wgPrint sync.WaitGroup
	if !quiet {
		wgPrint.Add(1)
		go func() {
			defer wgPrint.Done()
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					mu.Lock()
					printProgress(startTime, totalReqs, success, failed, latencies, int(readyWorkers), concurrency)
					mu.Unlock()
					return
				case <-ticker.C:
					mu.Lock()
					printProgress(startTime, totalReqs, success, failed, latencies, int(readyWorkers), concurrency)
					mu.Unlock()
				}
			}
		}()
	}

	// 启动所有 worker
	var wgWorkers sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wgWorkers.Add(1)
		go func() {
			defer wgWorkers.Done()

			// 标记自己已就绪
			atomic.AddInt64(&readyWorkers, 1)

			// 等待启动信号
			<-startSignal

			// 开始压测循环
			for {
				if rate > 0 {
					select {
					case <-ctx.Done():
						return
					case <-limiterChan:
					}
				} else {
					select {
					case <-ctx.Done():
						return
					default:
					}
				}

				mu.Lock()
				if total > 0 && sentCount >= total {
					mu.Unlock()
					cancel()
					return
				}
				sentCount++
				mu.Unlock()

				res, _ := tgt.Send(ctx)
				if res == nil {
					continue
				}

				mu.Lock()
				totalReqs++
				if res.Success {
					success++
				} else {
					failed++
				}
				latencies = append(latencies, res.Latency)
				mu.Unlock()
			}
		}()
	}

	// 【关键】等待所有 worker 就绪后，再发出启动信号
	// 简单轮询（也可用 channel，但轮询足够轻量）
	for atomic.LoadInt64(&readyWorkers) < int64(concurrency) {
		time.Sleep(100 * time.Microsecond) // 避免忙等
	}
	close(startSignal) // 广播：所有 worker 开始！

	// 等待压测结束
	<-ctx.Done()
	wgWorkers.Wait()
	if !quiet {
		wgPrint.Wait()
	}

	// 计算最终结果
	endTime := time.Now()
	durationSec := endTime.Sub(startTime).Seconds()
	realDuration := time.Since(startTime)
	rps := float64(totalReqs) / durationSec

	var avgLatencyMs, minLatencyMs, maxLatencyMs, p95LatencyMs, p99LatencyMs float64
	var successRate float64

	if totalReqs > 0 {
		sortedLatencies := make([]time.Duration, len(latencies))
		copy(sortedLatencies, latencies)
		sort.Slice(sortedLatencies, func(i, j int) bool {
			return sortedLatencies[i] < sortedLatencies[j]
		})

		var sum float64
		min := sortedLatencies[0]
		max := sortedLatencies[len(sortedLatencies)-1]
		for _, d := range sortedLatencies {
			sum += float64(d.Milliseconds())
		}
		avgLatencyMs = sum / float64(len(sortedLatencies))
		minLatencyMs = float64(min.Milliseconds())
		maxLatencyMs = float64(max.Milliseconds())

		n := len(sortedLatencies)
		p95Idx := int(float64(n) * 0.95)
		if p95Idx >= n {
			p95Idx = n - 1
		}
		p99Idx := int(float64(n) * 0.99)
		if p99Idx >= n {
			p99Idx = n - 1
		}
		p95LatencyMs = float64(sortedLatencies[p95Idx].Milliseconds())
		p99LatencyMs = float64(sortedLatencies[p99Idx].Milliseconds())

		successRate = float64(success) / float64(totalReqs)
	}

	return &Result{
		TotalRequests: totalReqs,
		Successful:    success,
		Failed:        failed,
		SuccessRate:   successRate,
		AvgLatencyMs:  avgLatencyMs,
		MinLatencyMs:  minLatencyMs,
		MaxLatencyMs:  maxLatencyMs,
		P95LatencyMs:  p95LatencyMs,
		P99LatencyMs:  p99LatencyMs,
		RPS:           rps,
		Duration:      realDuration,
	}, nil
}

// printProgress 新增 workersReady 和 totalWorkers 参数
func printProgress(
	startTime time.Time,
	total, success, failed int,
	latencies []time.Duration,
	workersReady, totalWorkers int,
) {
	elapsed := time.Since(startTime).Seconds()
	if elapsed < 1.0 {
		elapsed = 1.0
	}
	currentRPS := float64(success) / elapsed

	var successRate float64
	if total > 0 {
		successRate = float64(success) / float64(total)
	}

	if total == 0 {
		fmt.Printf("Workers: %d/%d | RPS: %.1f | Total: %d | Success: %.1f%% | Failed: %d | Avg: 0ms | Min: 0ms | Max: 0ms | P95: 0ms | P99: 0ms\n",
			workersReady, totalWorkers, currentRPS, total, successRate*100, failed)
		return
	}

	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var sum float64
	min := sorted[0]
	max := sorted[len(sorted)-1]
	for _, d := range sorted {
		sum += float64(d.Milliseconds())
	}
	avg := sum / float64(len(sorted))

	n := len(sorted)
	p95Idx := int(float64(n) * 0.95)
	if p95Idx >= n {
		p95Idx = n - 1
	}
	p99Idx := int(float64(n) * 0.99)
	if p99Idx >= n {
		p99Idx = n - 1
	}
	p95 := float64(sorted[p95Idx].Milliseconds())
	p99 := float64(sorted[p99Idx].Milliseconds())

	fmt.Printf("Workers: %d/%d | RPS: %.1f | Total: %d | Success: %.1f%% | Failed: %d | Avg: %.1fms | Min: %dms | Max: %dms | P95: %.1fms | P99: %.1fms\n",
		workersReady, totalWorkers, currentRPS, total, successRate*100, failed, avg, min.Milliseconds(), max.Milliseconds(), p95, p99)
}
