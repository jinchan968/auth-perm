package middleware

import (
	"net/http"
	"strings"

	"auth-perm/config"
	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/domain/auth/service"
	"auth-perm/internal/domain/permission/constant"
	permissionParam "auth-perm/internal/domain/permission/param"
	permissionService "auth-perm/internal/domain/permission/service"

	"github.com/gin-gonic/gin"
)

// APIPermissionMiddleware API 权限拦截中间件
// 全局挂载，通过白名单 + 超管绕过 + 资源匹配三层控制
func APIPermissionMiddleware(
	cfg *config.Config,
	authService *service.AuthService,
	permService *permissionService.PermissionService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// ========== 白名单：直接放行 ==========
		if isWhitelisted(path) {
			c.Next()
			return
		}

		// ========== 超管：直接放行 ==========
		if cfg.Server.SuperAdmin != "" {
			usernameStr := resolveUsernameFromContext(c, authService)
			if usernameStr != "" && usernameStr == cfg.Server.SuperAdmin {
				c.Set("is_super_admin", true)
				c.Next()
				return
			}
		}

		// ========== 普通用户：检查权限 ==========
		// 如果 account_id 不存在，说明请求未经过 AuthMiddleware 认证
		// 直接放行，让后续的 AuthMiddleware 处理认证逻辑
		accountID, exists := c.Get("account_id")
		if !exists || accountID == nil || accountID == "" {
			c.Next()
			return
		}

		accountIDStr, ok := accountID.(string)
		if !ok {
			response.Error(c, http.StatusInternalServerError, "系统错误", "账户ID类型错误")
			c.Abort()
			return
		}

		// 获取用户拥有的所有 API 资源
		resources, err := permService.GetAccountResources(c.Request.Context(), permissionParam.NewGetAccountResourcesParams(accountIDStr, constant.ResourceTypeAPIPath))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "权限检查失败", err.Error())
			c.Abort()
			return
		}

		// 匹配当前请求路径
		if matchAPIPath(path, resources) {
			c.Next()
			return
		}

		// 无权限，返回 403
		response.Error(c, http.StatusForbidden, "权限不足", "您没有访问此资源的权限")
		c.Abort()
	}
}

// isWhitelisted 检查路径是否在白名单中（免权限校验）
func isWhitelisted(path string) bool {
	// 精确匹配白名单
	exactWhitelist := []string{
		"/health",
		"/api/v1/auth/validate",
		"/api/v1/auth/refresh",
		"/api/v1/auth/logout",
		"/api/v1/auth/my-resources",
	}
	for _, exact := range exactWhitelist {
		if path == exact {
			return true
		}
	}

	// 前缀匹配白名单（以 / 结尾表示匹配该路径下所有子路径）
	prefixWhitelist := []string{
		"/api/v1/auth/public/",       // 登录、注册、忘记密码等
		"/api/v1/auth/profile",       // 个人资料（含 GET/PATCH）
		"/api/v1/auth/sessions",      // 会话管理
		"/api/v1/auth/devices",       // 设备管理
		"/api/v1/auth/security/logs", // 安全日志（个人）
		"/api/v1/auth/2fa/",          // 2FA 相关
		"/api/v1/auth/oauth/",        // OAuth 回调
		"/api/v1/auth/send-verification-email",
		"/api/v1/auth/verify-email",
		"/api/v1/auth/resend-verification-email",
		"/api/v1/auth/request-password-reset",
		"/api/v1/auth/reset-password",
		"/api/v1/auth/change-password",
	}
	for _, prefix := range prefixWhitelist {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// matchAPIPath 匹配 API 路径
// 支持三种匹配模式：
// 1. 精确匹配：/api/v1/tenants
// 2. 通配符匹配：/api/v1/tenants/* 匹配 /api/v1/tenants/xxx/settings
// 3. 路径参数匹配：忽略 UUID/ID 段，做前缀匹配
func matchAPIPath(requestPath string, allowedPaths []string) bool {
	for _, allowed := range allowedPaths {
		// 1. 精确匹配
		if requestPath == allowed {
			return true
		}

		// 2. 通配符匹配（/*）
		if strings.HasSuffix(allowed, "/*") {
			prefix := strings.TrimSuffix(allowed, "/*")
			if strings.HasPrefix(requestPath, prefix+"/") || requestPath == prefix {
				return true
			}
		}

		// 3. 路径参数匹配（忽略 UUID/ID 段）
		// 例如：allowed = "/api/v1/permissions/roles"
		//      request = "/api/v1/permissions/roles/123e4567-e89b-12d3-a456-426614174000"
		//      应该匹配成功
		if strings.HasPrefix(requestPath, allowed+"/") {
			return true
		}
	}

	return false
}

// resolveUsernameFromContext 从 gin context 获取 username
// 优先从 context（Redis 缓存），兜底通过 user_id 查 DB
func resolveUsernameFromContext(c *gin.Context, authService *service.AuthService) string {
	// 1. 优先从 context 获取（Redis 缓存，零 DB 查询）
	if username, exists := c.Get("username"); exists {
		if s, ok := username.(string); ok && s != "" {
			return s
		}
	}

	// 2. 兜底：通过 user_id 查 DB（旧 session 没有 username 的情况）
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		return ""
	}

	user, err := authService.FindUserByID(c.Request.Context(), userIDStr)
	if err != nil || user == nil {
		return ""
	}

	// 写回 context，后续中间件/handler 不需要再查
	c.Set("username", user.Username)
	return user.Username
}
