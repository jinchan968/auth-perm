package middleware

import (
	"auth-perm/internal/common/constant"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware 恢复中间件，处理panic
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := make([]byte, constant.DefaultStackSize)
				length := runtime.Stack(stack, false)
				stackStr := string(stack[:length])

				// 记录panic信息
				logPanic(c, err, stackStr)

				// 返回错误响应
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal Server Error",
					"code":       "INTERNAL_SERVER_ERROR",
					"message":    "An unexpected error occurred",
					"request_id": c.GetString("request_id"),
				})
			}
		}()

		c.Next()
	}
}

// logPanic FUTURE: panic日志记录 - 在实现panic处理时使用
func logPanic(c *gin.Context, err interface{}, stack string) {
	// 构建panic日志
	panicLog := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"level":      "error",
		"message":    "Panic recovered",
		"panic":      fmt.Sprintf("%v", err),
		"stack":      stack,
		"request_id": c.GetString("request_id"),
		"tenant_id":  c.GetString("tenant_id"),
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"query":      c.Request.URL.RawQuery,
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"headers":    extractHeaders(c),
	}

	// 如果有请求体，也记录下来
	if body := c.GetString("request_body"); body != "" {
		panicLog["request_body"] = body
	}

	// 转换为JSON并输出
	jsonData, _ := json.Marshal(panicLog)
	log.Printf("PANIC: %s\n", string(jsonData))
}

// extractHeaders FUTURE: 头部提取 - 在实现请求分析时使用
func extractHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)

	// 定义要排除的敏感头
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"set-cookie":    true,
	}

	for name, values := range c.Request.Header {
		lowerName := strings.ToLower(name)
		if !sensitiveHeaders[lowerName] && len(values) > 0 {
			headers[name] = values[0]
		}
	}

	return headers
}

// CustomRecovery FUTURE: 自定义恢复中间件 - 在实现自定义错误处理时使用
func CustomRecovery(onPanic func(*gin.Context, interface{}, string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := make([]byte, constant.DefaultStackSize)
				length := runtime.Stack(stack, false)
				stackStr := string(stack[:length])

				// 调用自定义panic处理函数
				onPanic(c, err, stackStr)

				// 返回错误响应
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal Server Error",
					"code":       "INTERNAL_SERVER_ERROR",
					"message":    "An unexpected error occurred",
					"request_id": c.GetString("request_id"),
				})
			}
		}()

		c.Next()
	}
}

// Writer 状态码和响应体捕获器
type RecoveryWriter struct {
	gin.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (w *RecoveryWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *RecoveryWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// DetailedRecoveryMiddleware FUTURE: 详细恢复中间件 - 在实现详细错误报告时使用
func DetailedRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 包装响应写入器
		writer := &RecoveryWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}
		c.Writer = writer

		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := make([]byte, constant.DetailedStackSize)
				length := runtime.Stack(stack, false)
				stackStr := string(stack[:length])

				// 记录详细的panic信息
				logDetailedPanic(c, err, stackStr, writer)

				// 如果已经开始写入响应，尝试写入错误信息
				if writer.statusCode >= 200 && writer.statusCode < 300 {
					// 成功响应中出现panic，记录但不修改响应
					c.Next()
					return
				}

				// 返回错误响应
				response := gin.H{
					"error":      "Internal Server Error",
					"code":       "INTERNAL_SERVER_ERROR",
					"message":    "An unexpected error occurred",
					"request_id": c.GetString("request_id"),
					"timestamp":  time.Now().Unix(),
				}

				// 在开发环境中，返回更多信息
				if gin.Mode() == gin.DebugMode {
					response["debug"] = gin.H{
						"panic": fmt.Sprintf("%v", err),
						"stack": strings.Split(stackStr, "\n"),
					}
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			}
		}()

		c.Next()
	}
}

// logDetailedPanic FUTURE: 详细panic日志记录 - 在实现详细错误日志时使用
func logDetailedPanic(c *gin.Context, err interface{}, stack string, writer *RecoveryWriter) {
	panicLog := map[string]interface{}{
		"timestamp":     time.Now().Format(time.RFC3339),
		"level":         "error",
		"message":       "Panic recovered",
		"panic":         fmt.Sprintf("%v", err),
		"stack":         stack,
		"request_id":    c.GetString("request_id"),
		"tenant_id":     c.GetString("tenant_id"),
		"method":        c.Request.Method,
		"path":          c.Request.URL.Path,
		"query":         c.Request.URL.RawQuery,
		"client_ip":     c.ClientIP(),
		"user_agent":    c.Request.UserAgent(),
		"response_code": writer.statusCode,
		"response_body": "",
		"headers":       extractHeaders(c),
		"env":           gin.Mode(),
	}

	// 添加请求体
	if body := c.GetString("request_body"); body != "" {
		panicLog["request_body"] = body
	}

	// 添加响应体（如果是错误响应）
	if writer.statusCode >= 400 && writer.body.Len() < 1024 {
		panicLog["response_body"] = writer.body.String()
	}

	// 添加运行时信息
	panicLog["runtime"] = map[string]interface{}{
		"go_version":    runtime.Version(),
		"go_os":         runtime.GOOS,
		"go_arch":       runtime.GOARCH,
		"num_goroutine": runtime.NumGoroutine(),
		"num_cpu":       runtime.NumCPU(),
	}

	// 内存信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	panicLog["memory"] = map[string]interface{}{
		"alloc_mb":       m.Alloc / 1024 / 1024,
		"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
		"sys_mb":         m.Sys / 1024 / 1024,
		"num_gc":         m.NumGC,
	}

	// 转换为JSON并输出
	jsonData, _ := json.Marshal(panicLog)
	log.Printf("DETAILED PANIC: %s\n", string(jsonData))
}
