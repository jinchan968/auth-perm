package http

import (
	"auth-perm/config"
	"auth-perm/internal/controller/middleware"
	authHandler "auth-perm/internal/domain/auth/handler"
	"auth-perm/internal/domain/auth/service"
	journalHandler "auth-perm/internal/domain/journal/handler"
	permHandler "auth-perm/internal/domain/permission/handler"
	permissionService "auth-perm/internal/domain/permission/service"
	tenantHandler "auth-perm/internal/domain/tenant/handler"
	todoHandler "auth-perm/internal/domain/todo/handler"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(
	router *gin.Engine,
	cfg *config.Config,
	authH *authHandler.AuthHandler,
	emailH *authHandler.EmailHandler,
	passwordH *authHandler.PasswordHandler,
	totpH *authHandler.TOTPHandler,
	oauthH *authHandler.OAuthHandler,
	permissionH *permHandler.PermissionHandler,
	permissionResourceH *permHandler.PermissionResourceHandler,
	organizationH *permHandler.OrganizationHandler,
	tenantH *tenantHandler.TenantHandler,
	userH *authHandler.UserHandler,
	resourceH *authHandler.ResourceHandler,
	thTodoHandler *todoHandler.TodoHandler,
	jhJournalHandler *journalHandler.JournalHandler,
	authService *service.AuthService,
	loginService *service.LoginService,
	permService *permissionService.PermissionService,
) {
	// API v1 路由组
	v1 := router.Group("/api/v1")

	// 全局 API 权限中间件（挂载到 v1 路由组，拦截所有 /api/v1/* 请求）
	v1.Use(middleware.APIPermissionMiddleware(cfg, authService, permService))

	{
		// 注册认证相关路由
		RegisterAuthRoutes(v1, authH, emailH, passwordH, totpH, oauthH, resourceH, authService, loginService, permService)

		// 注册权限相关路由
		RegisterPermissionRoutes(v1, permissionH, permissionResourceH, loginService)

		// 注册组织相关路由
		RegisterOrganizationRoutes(v1, organizationH, loginService)

		// 注册租户相关路由
		RegisterTenantRoutes(v1, tenantH, loginService)

		// 注册用户管理路由
		RegisterUserRoutes(v1, userH, loginService)

		// 注册待办路由
		RegisterTodoRoutes(v1, thTodoHandler, loginService)

		// 注册札记路由
		RegisterJournalRoutes(v1, jhJournalHandler, loginService)
	}
}

// RegisterAuthRoutes 注册认证路由
func RegisterAuthRoutes(
	router *gin.RouterGroup,
	authH *authHandler.AuthHandler,
	emailH *authHandler.EmailHandler,
	passwordH *authHandler.PasswordHandler,
	totpH *authHandler.TOTPHandler,
	oauthH *authHandler.OAuthHandler,
	resourceH *authHandler.ResourceHandler,
	_ *service.AuthService,
	loginService *service.LoginService,
	permService *permissionService.PermissionService,
) {
	auth := router.Group("/auth")

	public := auth.Group("/public")
	{
		public.POST("/login", authH.Login)
		public.POST("/register", authH.Register)
		public.POST("/forgot-password", authH.ForgotPassword)
		public.POST("/reset-password", authH.ResetPassword)
	}

	authenticated := auth.Group("/")
	authenticated.Use(middleware.AuthMiddleware(loginService))
	{
		authenticated.GET("/oauth/github/callback", oauthH.GitHubCallback)
		authenticated.GET("/oauth/google/callback", oauthH.GoogleCallback)
		authenticated.GET("/oauth/wechat/callback", oauthH.WeChatCallback)

		authenticated.GET("/validate", authH.ValidateToken)
		authenticated.POST("/logout", authH.Logout)
		authenticated.POST("/refresh", authH.RefreshToken)

		// 资源清单接口（用于前端权限控制，白名单路由，不走权限中间件）
		authenticated.GET("/my-resources", resourceH.GetMyResources)

		admin := auth.Group("/admin")
		admin.Use(middleware.AuthMiddleware(loginService))
		admin.Use(middleware.AdminPermissionMiddleware(permService))
		{
			admin.POST("/logout-all-by-tenant", authH.LogoutAllByTenant)
		}

		authenticated.POST("/send-verification-email", emailH.SendVerificationEmail)
		authenticated.GET("/verify-email", emailH.VerifyEmail)
		authenticated.POST("/resend-verification-email", emailH.ResendVerificationEmail)

		authenticated.POST("/request-password-reset", passwordH.RequestPasswordReset)
		authenticated.POST("/reset-password", passwordH.ResetPassword)

		authenticated.GET("/profile", authH.GetProfile)
		authenticated.PATCH("/profile", authH.UpdateProfile)
		authenticated.POST("/change-password", authH.ChangePassword)
		authenticated.GET("/sessions", authH.GetSessions)

		authenticated.DELETE("/sessions/:sessionId", authH.RevokeSession)
		authenticated.DELETE("/sessions/all", authH.RevokeAllSessions)
		authenticated.GET("/devices", authH.GetDevices)
		authenticated.DELETE("/devices/:deviceId", authH.RevokeDevice)
		authenticated.POST("/devices/:deviceId/trust", authH.TrustDevice)

		authenticated.GET("/security/logs", authH.GetSecurityLogs)

		authenticated.POST("/2fa/setup-init", totpH.TOTPSetupInit)
		authenticated.POST("/2fa/setup-verify", totpH.TOTPSetupVerify)
		authenticated.POST("/2fa/enable", totpH.TOTPEnable)
		authenticated.POST("/2fa/disable", totpH.TOTPDisable)
		authenticated.POST("/2fa/verify", totpH.TOTPVerify)
		authenticated.POST("/2fa/backup-code", totpH.TOTPBackupCode)
		authenticated.GET("/2fa/status", totpH.TOTPStatus)
		authenticated.POST("/2fa/change-secret", totpH.TOTPChangeSecret)
	}
}

// RegisterPermissionRoutes 注册权限路由
func RegisterPermissionRoutes(
	router *gin.RouterGroup,
	permissionH *permHandler.PermissionHandler,
	permissionResourceH *permHandler.PermissionResourceHandler,
	loginService *service.LoginService,
) {
	permissions := router.Group("/permissions")
	permissions.Use(middleware.AuthMiddleware(loginService))
	{
		permissions.POST("/check", permissionH.CheckPermission)
		permissions.POST("/check-any", permissionH.CheckAnyPermission)
		permissions.POST("/check-all", permissionH.CheckAllPermissions)

		permissions.GET("/role-check/:role", permissionH.CheckRole)
		permissions.POST("/check-any-role", permissionH.CheckAnyRole)
		permissions.POST("/check-all-roles", permissionH.CheckAllRoles)

		permissions.GET("", permissionH.GetPermissions)
		permissions.GET("/effective", permissionH.GetEffectivePermissions)
		permissions.GET("/is-super-admin", permissionH.IsSuperAdmin)

		permissions.POST("/org/:org_id/check", permissionH.CheckOrgPermission)
		permissions.GET("/org/:org_id/is-admin", permissionH.IsOrgAdmin)

		permissions.POST("/check-resource", permissionResourceH.CheckResourcePermission)
		permissions.GET("/account-resources", permissionResourceH.GetAccountResources)

		roles := permissions.Group("/roles")
		{
			roles.POST("", permissionH.CreateRole)
			roles.GET("/:id", permissionH.GetRole)
			roles.PUT("/:id", permissionH.UpdateRole)
			roles.DELETE("/:id", permissionH.DeleteRole)
			roles.GET("", permissionH.ListRolesHandler)
			roles.POST("/:id/permissions", permissionH.AssignPermissionToRole)
			roles.DELETE("/:id/permissions/:permissionId", permissionH.RemovePermissionFromRole)
			roles.GET("/:id/permissions", permissionH.GetRolePermissions)
		}

		accounts := permissions.Group("/accounts")
		{
			accounts.GET("/:accountId/roles", permissionH.GetAccountRoles)
			accounts.POST("/:accountId/roles", permissionH.AssignRoleToAccount)
			accounts.DELETE("/:accountId/roles/:roleId", permissionH.RemoveRoleFromAccount)
		}

		permissionItems := permissions.Group("/items")
		{
			permissionItems.POST("", permissionH.CreatePermission)
			permissionItems.GET("/:id", permissionH.GetPermission)
			permissionItems.PUT("/:id", permissionH.UpdatePermission)
			permissionItems.DELETE("/:id", permissionH.DeletePermission)
			permissionItems.GET("", permissionH.ListPermissions)
		}

		permissionResources := permissions.Group("/resources")
		{
			permissionResources.GET("", permissionResourceH.List)
			permissionResources.POST("", permissionResourceH.Create)
			permissionResources.POST("/batch", permissionResourceH.CreateBatch)
			permissionResources.PUT("/:resourceId", permissionResourceH.Update)
			permissionResources.DELETE("/:resourceId", permissionResourceH.Delete)
			permissionResources.POST("/bind", permissionResourceH.Bind)
			permissionResources.POST("/unbind", permissionResourceH.Unbind)
		}
	}
}

// RegisterOrganizationRoutes 注册组织路由
func RegisterOrganizationRoutes(
	router *gin.RouterGroup,
	organizationH *permHandler.OrganizationHandler,
	loginService *service.LoginService,
) {
	organizations := router.Group("/organizations")
	organizations.Use(middleware.AuthMiddleware(loginService))
	{
		organizations.POST("", organizationH.Create)
		organizations.GET("/:id", organizationH.Get)
		organizations.PUT("/:id", organizationH.Update)
		organizations.DELETE("/:id", organizationH.Delete)
		organizations.GET("", organizationH.List)
		organizations.GET("/tree", organizationH.GetTree)
		organizations.POST("/assign-account", organizationH.AssignAccountToOrg)
		organizations.POST("/remove-account", organizationH.RemoveAccountFromOrg)
		organizations.GET("/accounts/:accountId/organizations", organizationH.GetUserOrganizations)
	}
}

// RegisterTenantRoutes 注册租户路由
func RegisterTenantRoutes(
	router *gin.RouterGroup,
	tenantH *tenantHandler.TenantHandler,
	loginService *service.LoginService,
) {
	tenants := router.Group("/tenants")
	tenants.Use(middleware.AuthMiddleware(loginService))
	{
		tenants.POST("", tenantH.Create)
		tenants.GET("/:id", tenantH.Get)
		tenants.PUT("/:id", tenantH.Update)
		tenants.DELETE("/:id", tenantH.Delete)
		tenants.GET("", tenantH.List)
		tenants.GET("/:id/settings", tenantH.GetSettings)
		tenants.PUT("/:id/settings", tenantH.UpdateSettings)
		tenants.POST("/:id/change-status", tenantH.ChangeStatus)
	}
}

// RegisterUserRoutes 注册用户管理路由
func RegisterUserRoutes(
	router *gin.RouterGroup,
	userH *authHandler.UserHandler,
	loginService *service.LoginService,
) {
	users := router.Group("/users")
	users.Use(middleware.AuthMiddleware(loginService))
	{
		users.GET("", userH.ListUsers)
		users.POST("", userH.CreateUser)
		users.GET("/:id", userH.GetUser)
		users.PATCH("/:id/status", userH.UpdateUserStatus)
		users.GET("/:id/accounts", userH.GetUserAccounts)
	}
}

// RegisterJournalRoutes 注册札记路由
func RegisterJournalRoutes(
	router *gin.RouterGroup,
	h *journalHandler.JournalHandler,
	loginService *service.LoginService,
) {
	journal := router.Group("/journal")
	journal.Use(middleware.AuthMiddleware(loginService))
	{
		// 标签
		journal.GET("/tags", h.ListTags)
		journal.POST("/tags", h.CreateTag)
		journal.PUT("/tags/:id", h.UpdateTag)
		journal.DELETE("/tags/:id", h.DeleteTag)

		// 札记条目
		journal.GET("", h.ListEntries)
		journal.POST("", h.CreateEntry)
		journal.GET("/:id", h.GetEntry)
		journal.POST("/:id/corrections", h.AddCorrection)
		journal.PUT("/:id/tags", h.UpdateTags)
		journal.DELETE("/:id", h.DeleteEntry)
	}
}

// RegisterTodoRoutes 注册待办路由
func RegisterTodoRoutes(
	router *gin.RouterGroup,
	h *todoHandler.TodoHandler,
	loginService *service.LoginService,
) {
	todos := router.Group("/todos")
	todos.Use(middleware.AuthMiddleware(loginService))
	{
		todos.GET("/categories", h.ListCategories)
		todos.POST("/categories", h.CreateCategory)
		todos.PUT("/categories/:id", h.UpdateCategory)
		todos.DELETE("/categories/:id", h.DeleteCategory)

		todos.GET("", h.ListTodos)
		todos.POST("", h.CreateTodo)
		todos.GET("/:id", h.GetTodo)
		todos.PUT("/:id", h.UpdateTodo)
		todos.PATCH("/:id/status", h.UpdateTodoStatus)
		todos.PATCH("/:id/priority", h.UpdateTodoPriority)
		todos.DELETE("/:id", h.DeleteTodo)
	}
}
