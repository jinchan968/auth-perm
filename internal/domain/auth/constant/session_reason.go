package constant

// 会话失效原因常量
const (
	ReasonUserLogout             = "user_logout"
	ReasonUserRequestAllTenants  = "user_request_all_tenants"
	ReasonUserLogoutAll          = "user_logout_all"
	ReasonPasswordChange         = "password_change"
	ReasonPasswordReset          = "password_reset"
	ReasonPasswordResetWithToken = "password_reset_with_token"
	ReasonAdminForceLogout       = "admin_force_logout"
	ReasonUserRevokeSession      = "用户主动撤销会话"
	ReasonUserRevokeAllSessions  = "用户主动撤销所有会话"
	ReasonUserRevokeDevice       = "用户撤销设备"
	ReasonUserTrustDevice        = "用户主动信任"
	ReasonUserUntrustDevice      = "用户主动取消信任"
)
