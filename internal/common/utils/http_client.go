package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"auth-perm/internal/common/errors"
)

// HTTPClient 通用的HTTP客户端
type HTTPClient struct {
	client    *http.Client
	baseURL   string
	timeout   time.Duration
	headers   map[string]string
	userAgent string
}

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	BaseURL   string
	Timeout   time.Duration
	UserAgent string
	Headers   map[string]string
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient(config HTTPClientConfig) *HTTPClient {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}

	// 设置默认User-Agent
	if config.UserAgent == "" {
		config.UserAgent = "auth-perm/1.0"
	}
	config.Headers["User-Agent"] = config.UserAgent

	return &HTTPClient{
		client:    client,
		baseURL:   config.BaseURL,
		timeout:   config.Timeout,
		headers:   config.Headers,
		userAgent: config.UserAgent,
	}
}

// HTTPResponse HTTP响应
type HTTPResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	RequestID  string
}

// Get 发送GET请求
func (c *HTTPClient) Get(ctx context.Context, endpoint string, params map[string]string) (*HTTPResponse, error) {
	return c.Request(ctx, http.MethodGet, endpoint, params, nil, nil)
}

// Post 发送POST请求
func (c *HTTPClient) Post(ctx context.Context, endpoint string, params map[string]string, body interface{}) (*HTTPResponse, error) {
	return c.Request(ctx, http.MethodPost, endpoint, params, body, nil)
}

// Put 发送PUT请求
func (c *HTTPClient) Put(ctx context.Context, endpoint string, params map[string]string, body interface{}) (*HTTPResponse, error) {
	return c.Request(ctx, http.MethodPut, endpoint, params, body, nil)
}

// Delete 发送DELETE请求
func (c *HTTPClient) Delete(ctx context.Context, endpoint string, params map[string]string) (*HTTPResponse, error) {
	return c.Request(ctx, http.MethodDelete, endpoint, params, nil, nil)
}

// Request 发送HTTP请求
func (c *HTTPClient) Request(ctx context.Context, method, endpoint string, params map[string]string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	// 构建URL
	fullURL := c.buildURL(endpoint, params)

	// 准备请求体
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, errors.WrapBizError(err, "Failed to marshal request body")
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, errors.WrapBizError(err, "Failed to create HTTP request")
	}

	// 设置请求头
	c.setHeaders(req, headers)

	// 设置Content-Type
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.WrapBizError(err, fmt.Sprintf("HTTP request failed: %s %s", method, fullURL))
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WrapBizError(err, "Failed to read response body")
	}

	// 创建响应对象
	httpResp := &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       respBody,
		RequestID:  resp.Header.Get("X-Request-ID"),
	}

	// 检查响应状态码
	if resp.StatusCode >= 400 {
		return httpResp, errors.NewBusinessError(fmt.Sprintf("HTTP request failed with status code %d: %s", resp.StatusCode, string(respBody)))
	}

	return httpResp, nil
}

// GetJSON 获取JSON响应并解析
func (c *HTTPClient) GetJSON(ctx context.Context, endpoint string, params map[string]string, result interface{}) error {
	resp, err := c.Get(ctx, endpoint, params)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp.Body, result); err != nil {
		return errors.WrapBizError(err, "Failed to unmarshal JSON response")
	}

	return nil
}

// PostJSON 发送JSON请求并解析响应
func (c *HTTPClient) PostJSON(ctx context.Context, endpoint string, params map[string]string, body, result interface{}) error {
	resp, err := c.Post(ctx, endpoint, params, body)
	if err != nil {
		return err
	}

	if result != nil {
		if err := json.Unmarshal(resp.Body, result); err != nil {
			return errors.WrapBizError(err, "Failed to unmarshal JSON response")
		}
	}

	return nil
}

// OAuthClient OAuth客户端
type OAuthClient struct {
	*HTTPClient
	accessToken string
}

// NewOAuthClient 创建OAuth客户端
func NewOAuthClient(config HTTPClientConfig, accessToken string) *OAuthClient {
	client := NewHTTPClient(config)
	return &OAuthClient{
		HTTPClient:  client,
		accessToken: accessToken,
	}
}

// OAuthGet 发送带OAuth token的GET请求
func (c *OAuthClient) OAuthGet(ctx context.Context, endpoint string, params map[string]string) (*HTTPResponse, error) {
	return c.Request(ctx, http.MethodGet, endpoint, params, nil, map[string]string{
		"Authorization": "Bearer " + c.accessToken,
	})
}

// OAuthPost 发送带OAuth token的POST请求
func (c *OAuthClient) OAuthPost(ctx context.Context, endpoint string, params map[string]string, body interface{}) (*HTTPResponse, error) {
	return c.Request(ctx, http.MethodPost, endpoint, params, body, map[string]string{
		"Authorization": "Bearer " + c.accessToken,
	})
}

// GetJSONWithToken 使用OAuth token获取JSON响应
func (c *OAuthClient) GetJSONWithToken(ctx context.Context, endpoint string, params map[string]string, result interface{}) error {
	resp, err := c.OAuthGet(ctx, endpoint, params)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp.Body, result); err != nil {
		return errors.WrapBizError(err, "Failed to unmarshal JSON response")
	}

	return nil
}

// buildURL 构建完整URL
func (c *HTTPClient) buildURL(endpoint string, params map[string]string) string {
	// 处理baseURL
	baseURL := strings.TrimRight(c.baseURL, "/")
	fullURL := baseURL + "/" + strings.TrimLeft(endpoint, "/")

	// 添加查询参数
	if len(params) > 0 {
		values := url.Values{}
		for key, value := range params {
			values.Set(key, value)
		}
		fullURL = fullURL + "?" + values.Encode()
	}

	return fullURL
}

// setHeaders 设置请求头
func (c *HTTPClient) setHeaders(req *http.Request, headers map[string]string) {
	// 设置默认头部
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// 覆盖或添加自定义头部
	for key, value := range headers {
		if value == "" {
			req.Header.Del(key)
		} else {
			req.Header.Set(key, value)
		}
	}
}

// SetTimeout 设置超时时间
func (c *HTTPClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.client.Timeout = timeout
}

// SetHeader 设置请求头
func (c *HTTPClient) SetHeader(key, value string) {
	c.headers[key] = value
}

// SetUserAgent 设置User-Agent
func (c *HTTPClient) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
	c.headers["User-Agent"] = userAgent
}

// GetBaseURL 获取基础URL
func (c *HTTPClient) GetBaseURL() string {
	return c.baseURL
}
