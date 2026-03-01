package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"auth-perm/internal/common/constant"
	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/domain/auth/service"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(loginService *service.LoginService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			fmt.Println("AuthMiddleware: Token is empty")
			response.Error(c, http.StatusUnauthorized, "未认证", "token不能为空")
			c.Abort()
			return
		}

		// 解析token
		parts := strings.Split(token, ":")
		if len(parts) != 2 {
			fmt.Printf("AuthMiddleware: Invalid token format, expected 2 parts but got %d\n", len(parts))
			response.Error(c, http.StatusUnauthorized, "认证失败", "token格式错误")
			c.Abort()
			return
		}

		tokenHash := parts[0]
		session, err := loginService.ValidateSession(c.Request.Context(), tokenHash)
		if err != nil {
			fmt.Printf("AuthMiddleware: Session validation failed: %v\n", err)
			response.Error(c, http.StatusUnauthorized, "认证失败", err.Error())
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", session.UserID)
		c.Set("account_id", session.AccountID)
		c.Set("session_id", session.ID)
		c.Set("tenant_id", session.GetTenantID())
		c.Set("username", session.Username) // 注入 username 用于超管判断
		c.Set("token", token)

		c.Next()
	}
}

// OptionalAuthMiddleware FUTURE: 可选认证中间件 - 在实现可选认证时使用
func OptionalAuthMiddleware(loginService *service.LoginService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token != "" {
			// 解析token
			parts := strings.Split(token, ":")
			if len(parts) == 2 {
				tokenHash := parts[0]
				session, err := loginService.ValidateSession(c.Request.Context(), tokenHash)
				if err == nil {
					// 认证成功，设置用户信息
					c.Set("user_id", session.UserID)
					c.Set("account_id", session.AccountID)
					c.Set("session_id", session.ID)
					c.Set("tenant_id", session.GetTenantID())
					c.Set("authenticated", true)
					c.Next()
					return
				}
			}
		}

		// 认证失败或不进行认证
		c.Set("authenticated", false)
		c.Next()
	}
}

// extractToken 从请求中提取token
func extractToken(c *gin.Context) string {
	// 优先从x-auth-token header获取 (用于跨端口开发环境)
	authToken := c.GetHeader("x-auth-token")
	if authToken != "" {
		fmt.Printf("ExtractToken: Found x-auth-token: %s\n", authToken)
		return authToken
	}

	// 从Authorization header获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		fmt.Printf("ExtractToken: Found Authorization Bearer token: %s\n", token)
		return token
	}

	// 从Cookie获取
	token, err := c.Cookie(constant.CookieAuthToken)
	if err == nil && token != "" {
		fmt.Printf("ExtractToken: Found cookie token: %s\n", token)
		return token
	}

	// 从查询参数获取（仅用于特殊情况）
	queryToken := c.Query("token")
	if queryToken != "" {
		fmt.Printf("ExtractToken: Found query token: %s\n", queryToken)
		return queryToken
	}

	return ""
}
