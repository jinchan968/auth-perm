package constant

// 审计日志动作常量
const (
	ActionLogin                   = "login"
	ActionLogout                  = "logout"
	ActionCreateSession           = "create_session"
	ActionRefreshToken            = "refresh_token"
	ActionChangePassword          = "change_password"
	ActionResetPassword           = "reset_password"
	ActionResetPasswordWithToken  = "reset_password_with_token"
	ActionInvalidateSessionsError = "invalidate_sessions_error"
	ActionLogoutAllByTenant       = "logout_all_by_tenant"
	ActionLogoutAllByUser         = "logout_all_by_user"
)
