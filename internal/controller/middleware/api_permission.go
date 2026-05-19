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
// 须挂载在 AuthMiddleware 之后，通过白名单 + 超管绕过 + 资源匹配三层控制
func APIPermissionMiddleware(
	cfg *config.Config,
	authService *service.AuthService,
	permService *permissionService.PermissionService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// ========== 白名单：直接放行 ==========
		if isWhitelisted(path) {
			c.Next()
			return
		}

		// ========== 超管：直接放行 ==========
		isSuperAdmin := false
		if cfg.Server.SuperAdmin != "" {
			usernameStr := resolveUsernameFromContext(c, authService)
			if usernameStr != "" && usernameStr == cfg.Server.SuperAdmin {
				isSuperAdmin = true
			}
		}
		// 如果不是用户名超管，检查是否为系统管理员（数据库中的超管标记）
		if !isSuperAdmin {
			accountID, exists := c.Get("account_id")
			if exists && accountID != nil && accountID != "" {
				if accountIDStr, ok := accountID.(string); ok {
					params := permissionParam.NewIsSystemAdminParams(accountIDStr)
					isAdmin, err := permService.IsSystemAdmin(c.Request.Context(), params)
					if err == nil && isAdmin {
						isSuperAdmin = true
					}
				}
			}
		}
		if isSuperAdmin {
			c.Set("is_super_admin", true)
			c.Next()
			return
		}

		// ========== 普通用户：检查权限 ==========
		// 本中间件须挂载在 AuthMiddleware 之后，account_id 必然已注入
		accountID, exists := c.Get("account_id")
		if !exists || accountID == nil || accountID == "" {
			response.Error(c, http.StatusUnauthorized, "未认证", "缺少认证信息")
			c.Abort()
			return
		}

		accountIDStr, ok := accountID.(string)
		if !ok {
			response.Error(c, http.StatusInternalServerError, "系统错误", "账户ID类型错误")
			c.Abort()
			return
		}

		resources, err := permService.GetAccountResources(c.Request.Context(), permissionParam.NewGetAccountResourcesParams(accountIDStr, constant.ResourceTypeAPIPath))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "权限检查失败", err.Error())
			c.Abort()
			return
		}

		if matchAPIPath(path, method, resources) {
			c.Next()
			return
		}

		response.Error(c, http.StatusForbidden, "权限不足", "您没有访问此资源的权限")
		c.Abort()
	}
}

// 白名单数据 — 包级变量，仅初始化一次，避免每次请求分配 slice
var (
	whitelistExact = map[string]bool{
		"/health":                   true,
		"/api/v1/auth/validate":     true,
		"/api/v1/auth/refresh":      true,
		"/api/v1/auth/logout":       true,
		"/api/v1/auth/my-resources": true,
	}
	// 去掉公共前缀 /api/v1/auth/ 后的后缀，用于前缀匹配
	whitelistSuffixes = []string{
		"public/",
		"admin/",
		"profile",
		"sessions",
		"devices",
		"security/logs",
		"2fa/",
		"oauth/",
		"send-verification-email",
		"verify-email",
		"resend-verification-email",
		"request-password-reset",
		"reset-password",
		"change-password",
	}
)

// isWhitelisted 检查路径是否在白名单中（免权限校验）
func isWhitelisted(path string) bool {
	// 1. 精确匹配 O(1)
	if whitelistExact[path] {
		return true
	}

	// 2. 快速拒绝：绝大多数非白名单路径共享此公共前缀
	const authPrefix = "/api/v1/auth/"
	if !strings.HasPrefix(path, authPrefix) {
		return false
	}

	// 3. 前缀匹配已确认的 /api/v1/auth/* 路径
	suffix := path[len(authPrefix):]
	for _, allowed := range whitelistSuffixes {
		if strings.HasPrefix(suffix, allowed) {
			return true
		}
	}
	return false
}

// matchAPIPath 匹配 API 路径（支持 HTTP 方法区分）
// allowedPaths 支持两种格式：
//  1. 纯路径："/api/v1/journal" — 匹配任意 HTTP 方法（向后兼容）
//  2. 方法+路径："DELETE /api/v1/journal/:id" — 仅匹配指定方法
//
// requestPath 和 requestMethod 分别为当前请求的路径
// requestPath 和 requestMethod 分别为当前请求的路径和方法
func matchAPIPath(requestPath string, requestMethod string, allowedPaths []string) bool {
	for _, allowed := range allowedPaths {
		// 检查是否包含 HTTP 方法前缀（格式："METHOD /path"）
		var allowedMethod, allowedPath string
		if idx := strings.Index(allowed, " "); idx > 0 {
			allowedMethod = strings.ToUpper(allowed[:idx])
			allowedPath = allowed[idx+1:]
		} else {
			// 纯路径格式，匹配任意方法
			allowedMethod = ""
			allowedPath = allowed
		}

		// 若 allowed 指定了方法，必须匹配
		if allowedMethod != "" && allowedMethod != requestMethod {
			continue
		}

		// 1. 精确匹配
		if requestPath == allowedPath {
			return true
		}

		// 2. 通配符匹配（/*）
		if strings.HasSuffix(allowedPath, "/*") {
			prefix := strings.TrimSuffix(allowedPath, "/*")
			if strings.HasPrefix(requestPath, prefix+"/") || requestPath == prefix {
				return true
			}
		}

		// 3. 路径参数匹配（忽略 UUID/ID 段，做前缀匹配）
		if strings.HasPrefix(requestPath, allowedPath+"/") {
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
