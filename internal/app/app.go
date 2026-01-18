package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/windlant/go-stress-test/internal/config"
	"github.com/windlant/go-stress-test/internal/reporter"
	"github.com/windlant/go-stress-test/internal/runner"
	"github.com/windlant/go-stress-test/internal/target"
)

// 一次压测运行的主流程
func Run(cfg *config.Config) error {
	// 验证配置
	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// 加载请求体
	body, err := loadRequestBody(cfg)
	if err != nil {
		return fmt.Errorf("failed to load request body: %w", err)
	}

	// 创建目标（Target）
	tgt, err := target.NewHTTP(cfg.URL, cfg.Method, cfg.Headers, body, cfg.Timeout, cfg.DisableKeepalive)
	if err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}

	// 创建并运行压测器
	results, err := runner.Run(tgt, cfg.Rate, cfg.Concurrency, cfg.Duration, cfg.Total, cfg.Quiet)
	if err != nil {
		return fmt.Errorf("runner failed: %w", err)
	}

	// 生成报告
	rep, err := reporter.New(results, cfg)
	if err != nil {
		return fmt.Errorf("failed to create reporter: %w", err)
	}
	if err := rep.Write(); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}
	// fmt.Printf("Starting stress test with config: %+v\n", cfg)
	return nil
}

// 检查配置逻辑一致性
func validateConfig(cfg *config.Config) error {
	if cfg.URL == "" {
		return errors.New("URL is required")
	}
	if cfg.Rate < 0 {
		return errors.New("rate must be >= 0")
	}
	if cfg.Concurrency <= 0 {
		return errors.New("concurrency must be > 0")
	}
	if cfg.Duration <= 0 && cfg.Total <= 0 {
		return errors.New("either duration or total must be > 0")
	}
	if cfg.Timeout <= 0 {
		return errors.New("timeout must be > 0")
	}
	return nil
}

// 加载body内容
func loadRequestBody(cfg *config.Config) ([]byte, error) {
	if cfg.BodyFile != "" {
		// 检查文件是否存在
		if _, err := os.Stat(cfg.BodyFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("body file not found: %s", cfg.BodyFile)
		} else if err != nil {
			return nil, fmt.Errorf("failed to open body file %w", err)
		}
		data, err := os.ReadFile(filepath.Clean(cfg.BodyFile))
		if err != nil {
			return nil, fmt.Errorf("read body file: %w", err)
		}
		return data, nil
	}
	return []byte(cfg.Body), nil
}
