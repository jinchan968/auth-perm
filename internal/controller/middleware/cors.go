package middleware

import (
	"auth-perm/internal/common/constant"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Config CORS配置
type Config struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultConfig 默认CORS配置
func DefaultConfig() Config {
	return Config{
		AllowOrigins: []string{
			constant.Localhost3000,
			constant.Localhost8080,
			constant.SecureLocalhost3000,
			constant.SecureLocalhost8080,
		},
		AllowMethods: []string{
			string(constant.MethodGET), string(constant.MethodPOST), string(constant.MethodPUT),
			string(constant.MethodDELETE), string(constant.MethodOPTIONS), string(constant.MethodPATCH),
		},
		AllowHeaders: []string{
			constant.HeaderOrigin,
			constant.HeaderContentType,
			constant.HeaderAccept,
			constant.HeaderAuthorization,
			constant.HeaderAuthToken,
			constant.HeaderTenantID,
			constant.HeaderRequestID,
		},
		ExposeHeaders: []string{
			constant.ExposeTotalCount,
			constant.ExposeRequestID,
		},
		AllowCredentials: true,
		MaxAge:           int(constant.CORSMaxAge.Seconds()),
	}
}

// DevelopmentConfig FUTURE: 开发环境CORS配置 - 在实现开发环境配置时使用
func DevelopmentConfig() Config {
	config := DefaultConfig()
	config.AllowOrigins = append(config.AllowOrigins, "*")
	return config
}

// ProductionConfig FUTURE: 生产环境CORS配置 - 在实现生产环境配置时使用
func ProductionConfig(allowedOrigins []string) Config {
	config := DefaultConfig()
	config.AllowOrigins = allowedOrigins
	return config
}

// CORSMiddleware CORS中间件
func CORSMiddleware(config Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否允许该源
		if isAllowedOrigin(origin, config.AllowOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(config.AllowOrigins) > 0 && config.AllowOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// 设置允许的方法
		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		}

		// 设置允许的头部
		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}

		// 设置暴露的头部
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		// 设置是否允许凭证
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 设置预检请求的缓存时间
		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// isAllowedOrigin FUTURE: 允许的源检查 - 在实现CORS验证时使用
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// 支持通配符域名，如 *.example.com
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}
