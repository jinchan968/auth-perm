package middleware

import (
	"auth-perm/config"
	"auth-perm/internal/common/constant"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggingMiddleware 日志中间件
func LoggingMiddleware(cfg *config.Config) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 获取请求ID
		requestID := param.Keys["request_id"]
		if requestID == nil {
			requestID = constant.UnknownRequestID
		}

		// 获取租户ID
		tenantID := param.Keys["tenant_id"]
		if tenantID == nil {
			tenantID = constant.DefaultTenantID
		}

		// 获取用户ID
		userID := param.Keys["user_id"]
		if userID == nil {
			userID = "anonymous"
		}

		// 构建日志条目
		logEntry := map[string]interface{}{
			"timestamp":    param.TimeStamp.Format(time.RFC3339),
			"level":        string(constant.LogLevelInfo),
			"message":      "HTTP Request",
			"request_id":   requestID,
			"tenant_id":    tenantID,
			"user_id":      userID,
			"method":       param.Method,
			"path":         param.Path,
			"status_code":  param.StatusCode,
			"latency_ms":   param.Latency.Milliseconds(),
			"client_ip":    param.ClientIP,
			"user_agent":   param.Request.UserAgent(),
			"referer":      param.Request.Referer(),
			"request_size": param.Request.ContentLength,
		}

		// 添加错误信息（如果有）
		if param.ErrorMessage != "" {
			logEntry["error"] = param.ErrorMessage
			logEntry["level"] = string(constant.LogLevelError)
		}

		// 如果有响应体且状态码不是2xx，记录响应内容
		if param.StatusCode >= constant.StatusCode4xx && param.BodySize > 0 && param.BodySize < constant.MaxLogBodySize {
			if body := param.Keys["response_body"]; body != nil {
				logEntry["response_body"] = body
			}
		}

		// 转换为JSON格式
		jsonData, _ := json.Marshal(logEntry)
		return string(jsonData) + "\n"
	})
}

// RequestIDMiddleware FUTURE: 请求ID中间件 - 在实现请求跟踪时使用
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取或生成新的请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置到上下文和响应头
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// TenantMiddleware FUTURE: 租户中间件 - 在实现多租户时使用
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取租户ID
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			// 如果没有传递租户ID，使用默认租户
			tenantID = "default"
		}

		// 验证租户ID格式
		if !isValidTenantID(tenantID) {
			c.JSON(400, gin.H{
				"error": "Invalid tenant ID format",
				"code":  "INVALID_TENANT_ID",
			})
			c.Abort()
			return
		}

		// 设置到上下文
		c.Set("tenant_id", tenantID)

		c.Next()
	}
}

// isValidTenantID FUTURE: 租户ID有效性检查 - 在实现租户验证时使用
func isValidTenantID(tenantID string) bool {
	// 这里可以添加更严格的验证逻辑
	// 例如：UUID格式、特定长度、特定字符等
	return len(tenantID) > 0 && len(tenantID) <= 100
}

// RequestLoggingMiddleware FUTURE: 请求日志中间件 - 在实现请求日志时使用
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 获取客户端IP
		clientIP := c.ClientIP()

		// 获取请求方法
		method := c.Request.Method

		// 获取状态码
		statusCode := c.Writer.Status()

		// 构建完整路径
		if raw != "" {
			path = path + "?" + raw
		}

		// 记录请求日志
		message := fmt.Sprintf("%s %s %d %v %s", method, path, statusCode, latency, clientIP)

		// 根据状态码决定日志级别
		if statusCode >= 500 {
			// 服务器错误
			c.Error(fmt.Errorf("%s", message))
		} else if statusCode >= 400 {
			// 客户端错误
			c.Writer.WriteString(message + "\n")
		} else {
			// 成功请求
			c.Writer.WriteString(message + "\n")
		}
	}
}

// responseBodyWriter 响应体写入器，用于捕获响应内容
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// ResponseLoggingMiddleware FUTURE: 响应日志中间件 - 在实现响应日志时使用
func ResponseLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 包装响应写入器以捕获响应体
		w := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		// 处理请求
		c.Next()

		// 只记录错误响应的内容
		if c.Writer.Status() >= 400 && w.body.Len() < 1024 {
			// 设置响应体到上下文，供日志中间件使用
			c.Set("response_body", w.body.String())
		}
	}
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	EnableRequestBody  bool
	EnableResponseBody bool
	MaxBodySize        int64
	SkipPaths          []string
}

// RequestBodyLoggingMiddleware FUTURE: 请求体日志中间件 - 在实现请求体日志时使用
func RequestBodyLoggingMiddleware(config LoggingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过该路径
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// 只记录特定方法的请求体
		if !shouldLogRequestBody(c.Request.Method) {
			c.Next()
			return
		}

		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}

		// 检查请求体大小
		if int64(len(body)) > config.MaxBodySize {
			c.Next()
			return
		}

		// 重置请求体
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// 记录请求体（敏感信息需要过滤）
		filteredBody := filterSensitiveData(body)
		c.Set("request_body", string(filteredBody))

		c.Next()
	}
}

// shouldLogRequestBody FUTURE: 请求体日志判断 - 在实现条件日志时使用
func shouldLogRequestBody(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH"
}

// filterSensitiveData FUTURE: 敏感数据过滤 - 在实现数据脱敏时使用
func filterSensitiveData(body []byte) []byte {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body // 如果不是JSON，返回原内容
	}

	// 过滤敏感字段
	sensitiveFields := []string{"password", "token", "secret", "key", "authorization"}
	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			data[field] = "[FILTERED]"
		}
	}

	// 过滤嵌套对象中的敏感字段
	filterNestedSensitiveData(data)

	filtered, _ := json.Marshal(data)
	return filtered
}

// filterNestedSensitiveData FUTURE: 嵌套敏感数据过滤 - 在实现嵌套数据脱敏时使用
func filterNestedSensitiveData(data map[string]interface{}) {
	sensitiveFields := []string{"password", "token", "secret", "key", "authorization"}

	for key, value := range data {
		// 检查当前键是否为敏感字段
		for _, field := range sensitiveFields {
			if key == field {
				data[key] = "[FILTERED]"
				break
			}
		}

		// 递归处理嵌套对象
		if nestedObj, ok := value.(map[string]interface{}); ok {
			filterNestedSensitiveData(nestedObj)
		}
	}
}

// RouteLoggingMiddleware 路由日志中间件 - 打印每次请求的路由信息
func RouteLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 打印路由信息
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		fmt.Printf("[ROUTE] %s %s - Client: %s - Agent: %s\n", method, path, clientIP, userAgent)

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 根据状态码打印不同颜色的状态（可选）
		if statusCode >= 500 {
			fmt.Printf("[ROUTE] %s %s - Status: %d - Latency: %v - ❌ ERROR\n", method, path, statusCode, latency)
		} else if statusCode >= 400 {
			fmt.Printf("[ROUTE] %s %s - Status: %d - Latency: %v - ⚠️  CLIENT ERROR\n", method, path, statusCode, latency)
		} else {
			fmt.Printf("[ROUTE] %s %s - Status: %d - Latency: %v - ✅ OK\n", method, path, statusCode, latency)
		}
	}
}
