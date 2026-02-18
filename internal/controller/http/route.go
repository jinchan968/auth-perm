package http

import (
	"auth-perm/internal/controller/middleware"
	"auth-perm/internal/domain/auth/service"
	permissionService "auth-perm/internal/domain/permission/service"
	tenantHandler "auth-perm/internal/domain/tenant/handler"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(
	router *gin.Engine,
	authHandler *AuthHandler,
	emailHandler *EmailHandler,
	passwordHandler *PasswordHandler,
	totpHandler *TOTPHandler,
	oauthHandler *OAuthHandler,
	permissionHandler *PermissionHandler,
	permissionResourceHandler *PermissionResourceHandler,
	organizationHandler *OrganizationHandler,
	tenantHandler *tenantHandler.TenantHandler,
	authService *service.AuthService,
	loginService *service.LoginService,
	permService *permissionService.PermissionService,
) {
	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 注册认证相关路由
		RegisterAuthRoutes(v1, authHandler, emailHandler, passwordHandler, totpHandler, oauthHandler, authService, loginService, permService)

		// 注册权限相关路由
		RegisterPermissionRoutes(v1, permissionHandler, permissionResourceHandler, authService, loginService)

		// 注册组织相关路由
		RegisterOrganizationRoutes(v1, organizationHandler, loginService)

		// 注册租户相关路由
		RegisterTenantRoutes(v1, tenantHandler, loginService)
	}
}

// RegisterAuthRoutes 注册认证路由
func RegisterAuthRoutes(
	router *gin.RouterGroup,
	authHandler *AuthHandler,
	emailHandler *EmailHandler,
	passwordHandler *PasswordHandler,
	totpHandler *TOTPHandler,
	oauthHandler *OAuthHandler,
	authService *service.AuthService,
	loginService *service.LoginService,
	permService *permissionService.PermissionService,
) {
	auth := router.Group("/auth")

	// ========================================================================
	// 公开路由（无需认证）- 仅限login和register
	// ========================================================================
	public := auth.Group("/public")
	{
		public.POST("/login", authHandler.Login)
		public.POST("/register", authHandler.Register)
		// new public routes
		public.POST("/forgot-password", authHandler.ForgotPassword)
		public.POST("/reset-password", authHandler.ResetPassword)
	}

	// ========================================================================
	// 需要认证的路由（除了login、register之外的所有接口）
	// ========================================================================
	authenticated := auth.Group("/")
	authenticated.Use(middleware.AuthMiddleware(loginService))
	{
		// OAuth回调路由（特殊处理：从第三方服务携带授权码回调）
		// 注意：这些路由从外部服务回调，携带授权码而非用户token
		// 认证中间件会验证用户身份，确保只有已登录用户可以发起OAuth流程
		authenticated.GET("/oauth/github/callback", oauthHandler.GitHubCallback)
		authenticated.GET("/oauth/google/callback", oauthHandler.GoogleCallback)
		authenticated.GET("/oauth/wechat/callback", oauthHandler.WeChatCallback)

		// 认证相关
		authenticated.GET("/validate", authHandler.ValidateToken)
		authenticated.POST("/logout", authHandler.Logout)
		authenticated.POST("/refresh", authHandler.RefreshToken)

		// 管理员路由
		admin := auth.Group("/admin")
		admin.Use(middleware.AuthMiddleware(loginService))
		admin.Use(middleware.AdminPermissionMiddleware(permService))
		{
			admin.POST("/logout-all-by-tenant", authHandler.LogoutAllByTenant)
		}

		// 邮箱验证相关（需要认证）
		authenticated.POST("/send-verification-email", emailHandler.SendVerificationEmail)
		authenticated.GET("/verify-email", emailHandler.VerifyEmail)
		authenticated.POST("/resend-verification-email", emailHandler.ResendVerificationEmail)

		// 密码重置相关（需要认证）
		authenticated.POST("/request-password-reset", passwordHandler.RequestPasswordReset)
		authenticated.POST("/reset-password", passwordHandler.ResetPassword)

		// 用户资料相关
		authenticated.GET("/profile", authHandler.GetProfile)
		authenticated.PATCH("/profile", authHandler.UpdateProfile)
		authenticated.POST("/change-password", authHandler.ChangePassword)
		authenticated.GET("/sessions", authHandler.GetSessions)

		// Session & Device Management
		authenticated.DELETE("/sessions/:sessionId", authHandler.RevokeSession)
		authenticated.DELETE("/sessions/all", authHandler.RevokeAllSessions)
		authenticated.GET("/devices", authHandler.GetDevices)
		authenticated.DELETE("/devices/:deviceId", authHandler.RevokeDevice)
		authenticated.POST("/devices/:deviceId/trust", authHandler.TrustDevice)

		// Security Logs
		authenticated.GET("/security/logs", authHandler.GetSecurityLogs)

		// 2FA双因子认证路由
		authenticated.POST("/2fa/setup-init", totpHandler.TOTPSetupInit)
		authenticated.POST("/2fa/setup-verify", totpHandler.TOTPSetupVerify)
		authenticated.POST("/2fa/enable", totpHandler.TOTPEnable)
		authenticated.POST("/2fa/disable", totpHandler.TOTPDisable)
		authenticated.POST("/2fa/verify", totpHandler.TOTPVerify)
		authenticated.POST("/2fa/backup-code", totpHandler.TOTPBackupCode)
		authenticated.GET("/2fa/status", totpHandler.TOTPStatus)
		authenticated.POST("/2fa/change-secret", totpHandler.TOTPChangeSecret)
	}
}

// RegisterPermissionRoutes 注册权限路由
func RegisterPermissionRoutes(router *gin.RouterGroup, permissionHandler *PermissionHandler, permissionResourceHandler *PermissionResourceHandler, authService *service.AuthService, loginService *service.LoginService) {
	permissions := router.Group("/permissions")
	permissions.Use(middleware.AuthMiddleware(loginService))
	{
		// 权限检查
		permissions.POST("/check", permissionHandler.CheckPermission)
		permissions.POST("/check-any", permissionHandler.CheckAnyPermission)
		permissions.POST("/check-all", permissionHandler.CheckAllPermissions)

		// 角色检查
		// 角色检查 - 使用 /role-check/:role 避免与 /roles/:id 冲突
		permissions.GET("/role-check/:role", permissionHandler.CheckRole)
		permissions.POST("/check-any-role", permissionHandler.CheckAnyRole)
		permissions.POST("/check-all-roles", permissionHandler.CheckAllRoles)

		// 获取权限和角色
		permissions.GET("", permissionHandler.GetPermissions)
		permissions.GET("/effective", permissionHandler.GetEffectivePermissions)
		// 注意：/roles 路由在下面的 roles 子组中注册

		// 管理员检查
		permissions.GET("/is-super-admin", permissionHandler.IsSuperAdmin)

		// 组织权限
		permissions.POST("/org/:org_id/check", permissionHandler.CheckOrgPermission)
		permissions.GET("/org/:org_id/is-admin", permissionHandler.IsOrgAdmin)

		// 资源权限检查
		permissions.POST("/check-resource", permissionResourceHandler.CheckResourcePermission)
		permissions.GET("/account-resources", permissionResourceHandler.GetAccountResources)

		// Role CRUD - 需要放在 /permissions/:id 之前
		roles := permissions.Group("/roles")
		{
			roles.POST("", permissionHandler.CreateRole)
			roles.GET("/:id", permissionHandler.GetRole)
			roles.PUT("/:id", permissionHandler.UpdateRole)
			roles.DELETE("/:id", permissionHandler.DeleteRole)
			roles.GET("", permissionHandler.ListRolesHandler)

			// Role-Permission 关联管理
			roles.POST("/:id/permissions", permissionHandler.AssignPermissionToRole)
			roles.DELETE("/:id/permissions/:permissionId", permissionHandler.RemovePermissionFromRole)
			roles.GET("/:id/permissions", permissionHandler.GetRolePermissions)
		}

		// Account-Role 关联管理 - 需要放在 /permissions/:id 之前
		accounts := permissions.Group("/accounts")
		{
			accounts.POST("/:accountId/roles", permissionHandler.AssignRoleToAccount)
			accounts.DELETE("/:accountId/roles/:roleId", permissionHandler.RemoveRoleFromAccount)
		}

		// Permission CRUD
		permissionItems := permissions.Group("/items")
		{
			permissionItems.POST("", permissionHandler.CreatePermission)
			permissionItems.GET("/:id", permissionHandler.GetPermission)
			permissionItems.PUT("/:id", permissionHandler.UpdatePermission)
			permissionItems.DELETE("/:id", permissionHandler.DeletePermission)
			permissionItems.GET("", permissionHandler.ListPermissions)
		}

		// 权限资源管理 - 使用独立路径避免冲突
		permissionResources := permissions.Group("/resources")
		{
			// GET /permissions/resources?permission_id=xxx - 获取权限的所有资源（使用查询参数）
			permissionResources.GET("", permissionResourceHandler.List)
			// POST /permissions/resources - 创建权限资源关联
			permissionResources.POST("", permissionResourceHandler.Create)
			// POST /permissions/resources/batch - 批量创建
			permissionResources.POST("/batch", permissionResourceHandler.CreateBatch)
			// PUT /permissions/resources/:resourceId - 更新
			permissionResources.PUT("/:resourceId", permissionResourceHandler.Update)
			// DELETE /permissions/resources/:resourceId - 删除
			permissionResources.DELETE("/:resourceId", permissionResourceHandler.Delete)
			// POST /permissions/resources/bind - 绑定
			permissionResources.POST("/bind", permissionResourceHandler.Bind)
			// POST /permissions/resources/unbind - 解绑
			permissionResources.POST("/unbind", permissionResourceHandler.Unbind)
		}
	}
}

// RegisterOrganizationRoutes 注册组织路由
func RegisterOrganizationRoutes(router *gin.RouterGroup, organizationHandler *OrganizationHandler, loginService *service.LoginService) {
	organizations := router.Group("/organizations")
	organizations.Use(middleware.AuthMiddleware(loginService))
	{
		// 组织 CRUD
		organizations.POST("", organizationHandler.Create)
		organizations.GET("/:id", organizationHandler.Get)
		organizations.PUT("/:id", organizationHandler.Update)
		organizations.DELETE("/:id", organizationHandler.Delete)

		// 组织列表和树
		organizations.GET("", organizationHandler.List)
		organizations.GET("/tree", organizationHandler.GetTree)

		// 账户-组织关联
		organizations.POST("/assign-account", organizationHandler.AssignAccountToOrg)
		organizations.POST("/remove-account", organizationHandler.RemoveAccountFromOrg)
		organizations.GET("/accounts/:accountId/organizations", organizationHandler.GetUserOrganizations)
	}
}

// RegisterTenantRoutes 注册租户路由
func RegisterTenantRoutes(router *gin.RouterGroup, tenantHandler *tenantHandler.TenantHandler, loginService *service.LoginService) {
	tenants := router.Group("/tenants")
	tenants.Use(middleware.AuthMiddleware(loginService))
	{
		// 租户 CRUD
		tenants.POST("", tenantHandler.Create)
		tenants.GET("/:id", tenantHandler.Get)
		tenants.PUT("/:id", tenantHandler.Update)
		tenants.DELETE("/:id", tenantHandler.Delete)

		// 租户列表
		tenants.GET("", tenantHandler.List)

		// 租户设置
		tenants.GET("/:id/settings", tenantHandler.GetSettings)
		tenants.PUT("/:id/settings", tenantHandler.UpdateSettings)

		// 租户状态管理
		tenants.POST("/:id/change-status", tenantHandler.ChangeStatus)
	}
}
