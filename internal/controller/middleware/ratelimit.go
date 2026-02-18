package middleware

import (
	"auth-perm/internal/common/constant"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	// 每个时间窗口的请求数
	Requests int
	// 时间窗口大小（秒）
	Window int
	// 键的前缀
	KeyPrefix string
	// Redis客户端
	RedisClient *redis.Client
}

// DefaultRateLimitConfig 默认速率限制配置
func DefaultRateLimitConfig(redisClient *redis.Client) RateLimitConfig {
	return RateLimitConfig{
		Requests:    constant.RateLimitDefaultRequests,
		Window:      constant.RateLimitDefaultWindow,
		KeyPrefix:   "rate_limit:",
		RedisClient: redisClient,
	}
}

// CacheKeyPrefix FUTURE: 缓存键前缀 - 在实现限流缓存时使用
func CacheKeyPrefix() string {
	return constant.CacheKeyUser
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端标识符
		key := getClientKey(c, config.KeyPrefix)

		// 检查速率限制
		allowed, remaining, resetTime, err := checkRateLimit(c.Request.Context(), config, key)
		if err != nil {
			// 如果Redis出错，允许请求通过（fail open）
			c.Next()
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			// 请求被限制
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        "RATE_LIMIT_EXCEEDED",
				"message":     fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", resetTime-time.Now().Unix()),
				"retry_after": resetTime - time.Now().Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientKey 获取客户端标识符
func getClientKey(c *gin.Context, prefix string) string {
	// 优先使用用户ID（如果已认证）
	if userID := c.GetString("user_id"); userID != "" {
		return fmt.Sprintf("%s%s", prefix, userID)
	}

	// 使用租户ID + IP地址
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		tenantID = constant.DefaultTenantID
	}

	clientIP := c.ClientIP()
	return fmt.Sprintf("%s:%s:ip:%s", prefix, tenantID, clientIP)
}

// checkRateLimit 检查速率限制
func checkRateLimit(ctx context.Context, config RateLimitConfig, key string) (bool, int, int64, error) {
	now := time.Now().Unix()
	windowStart := now - int64(config.Window)

	// 使用Redis滑动窗口算法
	pipe := config.RedisClient.Pipeline()

	// 移除过期的请求记录
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

	// 添加当前请求
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d-%d", now, time.Now().Nanosecond()),
	})

	// 设置过期时间
	pipe.Expire(ctx, key, time.Duration(config.Window)*time.Second)

	// 获取当前窗口内的请求数
	countCmd := pipe.ZCard(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	currentCount := int(countCmd.Val())
	remaining := config.Requests - currentCount
	if remaining < 0 {
		remaining = 0
	}

	// 计算重置时间
	resetTime := now + int64(config.Window)

	allowed := currentCount <= config.Requests
	return allowed, remaining, resetTime, nil
}

// AdvancedRateLimitMiddleware FUTURE: 高级限流中间件 - 在实现复杂限流策略时使用
func AdvancedRateLimitMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 根据不同的端点和用户类型设置不同的限制
		config := getRateLimitConfig(c, redisClient)

		// 检查速率限制
		allowed, remaining, resetTime, err := checkRateLimit(c.Request.Context(), config, getClientKey(c, config.KeyPrefix))
		if err != nil {
			// 如果Redis出错，允许请求通过（fail open）
			c.Next()
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        "RATE_LIMIT_EXCEEDED",
				"message":     fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", resetTime-time.Now().Unix()),
				"retry_after": resetTime - time.Now().Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getRateLimitConfig FUTURE: 限流配置获取 - 在实现动态限流时使用
func getRateLimitConfig(c *gin.Context, redisClient *redis.Client) RateLimitConfig {
	path := c.Request.URL.Path
	method := c.Request.Method
	userID := c.GetString("user_id")

	// 默认配置
	config := DefaultRateLimitConfig(redisClient)

	// 认证相关的端点限制更严格
	if strings.HasPrefix(path, "/api/v1/auth") {
		config.Requests = constant.RateLimitAuthRequests
		config.Window = constant.RateLimitAuthWindow
	}

	// 登录端点特别限制
	if path == "/api/v1/auth/login" && method == string(constant.MethodPOST) {
		config.Requests = constant.RateLimitLoginRequests
		config.Window = constant.RateLimitLoginWindow
	}

	// 注册端点特别限制
	if path == "/api/v1/auth/register" && method == string(constant.MethodPOST) {
		config.Requests = constant.RateLimitRegisterRequests
		config.Window = constant.RateLimitRegisterWindow
	}

	// 已认证用户的限制更宽松
	if userID != "" {
		config.Requests *= 2 // 翻倍限制
	}

	// 管理员用户的限制更宽松
	if isAdminUser(c) {
		config.Requests *= 3 // 三倍限制
	}

	return config
}

// isAdminUser FUTURE: 管理员用户检查 - 在实现基于角色的限流时使用
func isAdminUser(c *gin.Context) bool {
	// 这里可以根据实际的用户角色系统来实现
	// 暂时返回false
	return false
}

// IPBasedRateLimitMiddleware FUTURE: 基于IP的限流中间件 - 在实现IP限流时使用
func IPBasedRateLimitMiddleware(redisClient *redis.Client, requests int, window int) gin.HandlerFunc {
	config := RateLimitConfig{
		Requests:    requests,
		Window:      window,
		KeyPrefix:   "ip_rate_limit:",
		RedisClient: redisClient,
	}

	return func(c *gin.Context) {
		// 只使用IP作为键
		key := fmt.Sprintf("%s%s", config.KeyPrefix, c.ClientIP())

		allowed, remaining, resetTime, err := checkRateLimit(c.Request.Context(), config, key)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "IP rate limit exceeded",
				"code":        "IP_RATE_LIMIT_EXCEEDED",
				"message":     "Too many requests from your IP address",
				"retry_after": resetTime - time.Now().Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserBasedRateLimitMiddleware FUTURE: 基于用户的限流中间件 - 在实现用户限流时使用
func UserBasedRateLimitMiddleware(redisClient *redis.Client, requests int, window int) gin.HandlerFunc {
	config := RateLimitConfig{
		Requests:    requests,
		Window:      window,
		KeyPrefix:   "user_rate_limit:",
		RedisClient: redisClient,
	}

	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			// 如果用户未认证，跳过用户级限制
			c.Next()
			return
		}

		key := fmt.Sprintf("%s%s", config.KeyPrefix, userID)

		allowed, remaining, resetTime, err := checkRateLimit(c.Request.Context(), config, key)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "User rate limit exceeded",
				"code":        "USER_RATE_LIMIT_EXCEEDED",
				"message":     "Too many requests from your account",
				"retry_after": resetTime - time.Now().Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
