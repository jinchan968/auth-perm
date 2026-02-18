package middleware

import (
	"net/http"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/domain/auth/service"
	permissionParam "auth-perm/internal/domain/permission/param"
	permissionService "auth-perm/internal/domain/permission/service"

	"github.com/gin-gonic/gin"
)

// AdminPermissionMiddleware 管理员权限检查中间件
// 使用方法：admin.Use(middleware.AdminPermissionMiddleware(permissionService))
func AdminPermissionMiddleware(ps *permissionService.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取账户ID
		accountID, exists := c.Get("account_id")
		if !exists || accountID == "" {
			response.Error(c, http.StatusUnauthorized, "未认证", "无法获取账户信息")
			c.Abort()
			return
		}

		// 检查是否为系统管理员
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

		// 设置管理员标志到上下文
		c.Set("is_admin", true)

		c.Next()
	}
}

// SuperAdminPermissionMiddleware 超级管理员权限检查中间件
func SuperAdminPermissionMiddleware(ps *permissionService.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取账户ID
		accountID, exists := c.Get("account_id")
		if !exists || accountID == "" {
			response.Error(c, http.StatusUnauthorized, "未认证", "无法获取账户信息")
			c.Abort()
			return
		}

		// 检查是否为超级管理员
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

		// 设置超级管理员标志到上下文
		c.Set("is_super_admin", true)

		c.Next()
	}
}

// AdminMiddlewareWithLoginService FUTURE: 使用LoginService的管理员中间件
// 注意：当前版本推荐使用 AdminPermissionMiddleware，因为它直接使用 PermissionService
// 此函数保留用于需要使用 LoginService 的场景
func AdminMiddlewareWithLoginService(loginService *service.LoginService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这个函数是一个占位符，当前版本不需要实现
		// 如有需要，可以通过 loginService 查询用户角色来判断是否为管理员
		c.Next()
	}
}
