// Package httpclient 提供统一的 HTTP 客户端，内置 Chrome UA、GBK 解码、重试。
// 所有 infra 数据源共用，避免被反爬识别为机器人。
package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// ChromeUA 模拟 Chrome 浏览器 User-Agent。
const ChromeUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

// uaTransport 包装 RoundTrip，自动注入 User-Agent 头。
type uaTransport struct {
	base http.RoundTripper
}

func (t *uaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", ChromeUA)
	}
	return t.base.RoundTrip(req)
}

// New 返回带 Chrome UA 的 HTTP 客户端，超时 30 秒。
func New() *http.Client {
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: &uaTransport{base: http.DefaultTransport},
	}
}

// NewWithTimeout 返回带 Chrome UA 的 HTTP 客户端，自定义超时。
func NewWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: &uaTransport{base: http.DefaultTransport},
	}
}

// NewWithRetry 返回带重试的 HTTP 客户端。请求失败时指数退避重试 maxRetries 次。
func NewWithRetry(maxRetries int) *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		Transport: &retryTransport{
			base:       &uaTransport{base: http.DefaultTransport},
			maxRetries: maxRetries,
		},
	}
}

// retryTransport 对 5xx 和网络错误做指数退避重试。
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<uint(attempt-1)*200) * time.Millisecond)
		}
		resp, err = t.base.RoundTrip(req)
		if err != nil {
			continue
		}
		if resp.StatusCode < 500 {
			return resp, nil
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("server error after %d retries", t.maxRetries)
}

// DecodeGBK 将 GBK 编码的字节流转为 UTF-8 字符串。
// 腾讯财经 API 返回 GBK 编码数据，需要转码。
func DecodeGBK(data []byte) (string, error) {
	utf8Data, _, err := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), data)
	if err != nil {
		return "", fmt.Errorf("decode GBK: %w", err)
	}
	return string(utf8Data), nil
}

// DecodeGBKReader 从 io.Reader 读取 GBK 数据并转为 UTF-8。
func DecodeGBKReader(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	return DecodeGBK(data)
}

// DecodeGBKResponse 从 HTTP 响应体读取 GBK 数据并转为 UTF-8。
func DecodeGBKResponse(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}
	return DecodeGBK(data)
}

// ReadBody 读取 HTTP 响应体为 UTF-8 字符串，自动检测 GBK 编码。
func ReadBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	// 检测 Content-Type 是否指定 GBK
	ct := resp.Header.Get("Content-Type")
	if bytes.Contains([]byte(ct), []byte("gbk")) || bytes.Contains([]byte(ct), []byte("GBK")) {
		return DecodeGBK(data)
	}

	return string(data), nil
}
