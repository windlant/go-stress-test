package reporter

import (
	"fmt"
	"io"
	"os"

	"github.com/windlant/go-stress-test/internal/config"
	"github.com/windlant/go-stress-test/internal/runner"
)

type Reporter struct {
	result *runner.Result
	cfg    *config.Config
	out    io.Writer
	closer io.Closer // 用于关闭文件（如果打开了）
}

// New 创建 reporter，根据 cfg.Output 决定输出目标
func New(result *runner.Result, cfg *config.Config) (*Reporter, error) {
	var out io.Writer = os.Stdout
	var closer io.Closer

	if cfg.Output != "" {
		file, err := os.Create(cfg.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to create output file %q: %w", cfg.Output, err)
		}
		out = file
		closer = file
	}

	return &Reporter{
		result: result,
		cfg:    cfg,
		out:    out,
		closer: closer,
	}, nil
}

// Write 生成并写入报告
func (r *Reporter) Write() error {
	defer func() {
		if r.closer != nil {
			r.closer.Close()
		}
	}()

	// 打印配置（仅在非 quiet 模式）
	if !r.cfg.Quiet {
		r.printConfig()
		fmt.Fprintln(r.out)
	}

	// 打印最终报告
	r.printFinalReport()

	return nil
}

// printConfig 打印压测配置
func (r *Reporter) printConfig() {
	fmt.Fprintf(r.out, "Started stress test with the following configuration:\n")
	fmt.Fprintf(r.out, "   URL:          %s\n", r.cfg.URL)
	fmt.Fprintf(r.out, "   Method:       %s\n", r.cfg.Method)
	if len(r.cfg.Headers) > 0 {
		fmt.Fprintf(r.out, "   Headers:      %v\n", r.cfg.Headers)
	}
	if len(r.cfg.Body) > 0 {
		bodyStr := string(r.cfg.Body)
		if len(bodyStr) > 50 {
			bodyStr = bodyStr[:50] + "..."
		}
		fmt.Fprintf(r.out, "   Body:         %q\n", bodyStr)
	}
	fmt.Fprintf(r.out, "   Concurrency:  %d workers\n", r.cfg.Concurrency)
	if r.cfg.Rate > 0 {
		fmt.Fprintf(r.out, "   Rate:         %d RPS\n", r.cfg.Rate)
	} else {
		fmt.Fprintf(r.out, "   Rate:         unlimited (burst)\n")
	}
	if r.cfg.Duration > 0 {
		fmt.Fprintf(r.out, "   Duration:     %v\n", r.cfg.Duration)
	}
	if r.cfg.Total > 0 {
		fmt.Fprintf(r.out, "   Total:        %d requests\n", r.cfg.Total)
	}
	fmt.Fprintf(r.out, "   Timeout:      %v\n", r.cfg.Timeout)
	fmt.Fprintf(r.out, "   Keep-alive:   %v\n", !r.cfg.DisableKeepalive)
	if r.cfg.Output != "" {
		fmt.Fprintf(r.out, "   Output:       %s\n", r.cfg.Output)
	}
}

// printFinalReport 打印最终结果
func (r *Reporter) printFinalReport() {
	fmt.Fprintf(r.out, "\nFinal Report\n")
	fmt.Fprintf(r.out, "==============\n")

	fmt.Fprintf(r.out, "Total Requests:    %d\n", r.result.TotalRequests)

	if r.result.TotalRequests > 0 {
		fmt.Fprintf(r.out, "Successful:        %d (%.2f%%)\n", r.result.Successful, r.result.SuccessRate*100)
		fmt.Fprintf(r.out, "Failed:            %d (%.2f%%)\n", r.result.Failed, float64(r.result.Failed)/float64(r.result.TotalRequests)*100)
	} else {
		fmt.Fprintf(r.out, "Successful:        0 (0.00%%)\n")
		fmt.Fprintf(r.out, "Failed:            0 (0.00%%)\n")
	}

	durationSec := r.result.Duration.Seconds()

	fmt.Fprintf(r.out, "Duration:          %.3f s\n", durationSec)
	fmt.Fprintf(r.out, "Avg RPS:           %.2f\n", r.result.RPS)

	if r.result.TotalRequests > 0 {
		fmt.Fprintf(r.out, "Avg Latency:       %.2f ms\n", r.result.AvgLatencyMs)
		fmt.Fprintf(r.out, "Min Latency:       %.2f ms\n", r.result.MinLatencyMs)
		fmt.Fprintf(r.out, "Max Latency:       %.2f ms\n", r.result.MaxLatencyMs)
		fmt.Fprintf(r.out, "P95 Latency:       %.2f ms\n", r.result.P95LatencyMs)
		fmt.Fprintf(r.out, "P99 Latency:       %.2f ms\n", r.result.P99LatencyMs)
	} else {
		fmt.Fprintf(r.out, "Latency:           N/A (no completed requests)\n")
	}
}
