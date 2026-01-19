package target

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPTarget struct {
	url     string
	method  string
	headers map[string]string
	body    []byte
	client  *http.Client
}

type Result struct {
	Success    bool
	StatusCode int
	Latency    time.Duration
	Error      error
}

func NewHTTP(url, method string, headerSlice []string, body []byte, timeout time.Duration, disableKeepalive bool) (*HTTPTarget, error) {
	if url == "" {
		return nil, fmt.Errorf("URL is required")
	}
	if method == "" {
		method = "GET"
	}

	headers := make(map[string]string)
	for _, h := range headerSlice {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format: %q (expected 'Key: Value')", h)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("empty header key in: %q", h)
		}
		headers[key] = value
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DisableKeepAlives: disableKeepalive,
			TLSClientConfig: &tls.Config{
				// 默认不跳过证书验证
				InsecureSkipVerify: false,
			},
			// MaxIdleConns:        100,
			// IdleConnTimeout:     90 * time.Second,
		},
	}

	return &HTTPTarget{
		url:     url,
		method:  method,
		headers: headers,
		body:    body,
		client:  client,
	}, nil
}

func (t *HTTPTarget) Send(ctx context.Context) (*Result, error) {
	start := time.Now()

	var reqBody io.Reader
	if len(t.body) > 0 {
		reqBody = bytes.NewReader(t.body)
	}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(t.method), t.url, reqBody)
	if err != nil {
		return &Result{
			Success: false,
			Error:   err,
			Latency: time.Since(start),
		}, nil
	}

	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	resp, err := t.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return &Result{
			Success: false,
			Error:   err,
			Latency: latency,
		}, nil
	}
	defer resp.Body.Close()

	// 读取响应体（避免连接未释放）
	_, _ = io.Copy(io.Discard, resp.Body)

	// 成功：200-299
	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	return &Result{
		Success:    success,
		StatusCode: resp.StatusCode,
		Latency:    latency,
		Error:      nil,
	}, nil
}
