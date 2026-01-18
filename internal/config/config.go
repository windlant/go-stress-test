package config

import (
	"time"
)

type Config struct {
	URL              string        // 目标URL
	Method           string        // 请求方法（GET/POST/PUT/DELETE 等）
	Body             string        // 请求体（可直接传字符串）
	BodyFile         string        // 从文件读取请求体，优先级高于 Body
	Headers          []string      // 自定义请求头，支持多次指定
	Rate             int           // 每秒请求数(RPS)
	Concurrency      int           // 并发数
	Duration         time.Duration // 压测持续时间
	Total            int           // 总请求数，若大于0则覆盖 Duration
	Timeout          time.Duration // 单个请求超时时间
	Output           string        // 报告输出位置：stdout（终端），report.json，report.csv
	Quiet            bool          // 静默模式（不显示实时进度）
	DisableKeepalive bool          // 禁用 HTTP Keep-Alive
}
