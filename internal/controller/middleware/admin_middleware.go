package middleware

import (
	"net/http"

	"auth-perm/config"
	"auth-perm/internal/common/dto/response"
	permissionParam "auth-perm/internal/domain/permission/param"
	permissionService "auth-perm/internal/domain/permission/service"
	"auth-perm/internal/domain/auth/service"

	"github.com/gin-gonic/gin"
)

// AdminPermissionMiddleware 管理员权限检查中间件
// 检查顺序：配置超管 → 角色体系（super_admin / admin）
// 使用方法：admin.Use(middleware.AdminPermissionMiddleware(cfg, permService))
func AdminPermissionMiddleware(cfg *config.Config, ps *permissionService.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取账户ID
		accountID, exists := c.Get("account_id")
		if !exists || accountID == "" {
			response.Error(c, http.StatusUnauthorized, "未认证", "无法获取账户信息")
			c.Abort()
			return
		}

		// 优先检查配置超管（与 APIPermissionMiddleware 超管逻辑保持一致）
		if cfg.Server.SuperAdmin != "" {
			if username, ok := c.Get("username"); ok {
				if usernameStr, ok2 := username.(string); ok2 && usernameStr == cfg.Server.SuperAdmin {
					c.Set("is_admin", true)
					c.Next()
					return
				}
			}
		}

		// 检查角色体系（super_admin / admin）
		params := permissionParam.NewIsSystemAdminParams(accountID.(string))
		isAdmin, err := ps.IsSystemAdmin(c.Request.Context(), params)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "权限检查失败", err.Error())
			c.Abort()
			return
		}

		if !isAdmin {
			response.Error(c, http.StatusForbidden, "权限不足", "需要管理员权限")
			c.Abort()
			return
		}

		c.Set("is_admin", true)
		c.Next()
	}
}

// SuperAdminPermissionMiddleware 超级管理员权限检查中间件
func SuperAdminPermissionMiddleware(ps *permissionService.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, exists := c.Get("account_id")
		if !exists || accountID == "" {
			response.Error(c, http.StatusUnauthorized, "未认证", "无法获取账户信息")
			c.Abort()
			return
		}

		params := permissionParam.NewIsSystemAdminParams(accountID.(string))
		isSuperAdmin, err := ps.IsSystemAdmin(c.Request.Context(), params)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "权限检查失败", err.Error())
			c.Abort()
			return
		}

		if !isSuperAdmin {
			response.Error(c, http.StatusForbidden, "权限不足", "需要超级管理员权限")
			c.Abort()
			return
		}

		c.Set("is_super_admin", true)
		c.Next()
	}
}

// AdminMiddlewareWithLoginService FUTURE: 使用LoginService的管理员中间件
func AdminMiddlewareWithLoginService(loginService *service.LoginService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
