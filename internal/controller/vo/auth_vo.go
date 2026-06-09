package vo

import (
	"regexp"

	"auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	authConstant "auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/param"
)

// LoginRequest 登录请求（支持邮箱或手机号）
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // 支持邮箱或手机号
	Password   string `json:"password" binding:"required,min=6"`
	DeviceID   string `json:"device_id,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	IPAddress  string `json:"ip_address,omitempty"`
	RememberMe bool   `json:"remember_me,omitempty"`
	TenantID   string `json:"tenant_id,omitempty"` // 租户ID，支持多租户登录
}

// ToLoginParams 转换为登录参数
func (r *LoginRequest) ToLoginParams() *param.LoginParams {
	return param.NewLoginParams(
		r.Identifier,
		r.Password,
		r.DeviceID,
		r.UserAgent,
		r.IPAddress,
		r.TenantID,
		r.RememberMe,
	)
}

// RegisterRequest 注册请求（支持邮箱或手机号）
type RegisterRequest struct {
	IdentifierType  string `json:"identifier_type" binding:"required,oneof=email phone"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Username        string `json:"username" binding:"required,min=3,max=50"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	InviteCode      string `json:"invite_code"`
	Nickname        string `json:"nickname,omitempty"`
	TenantID        string `json:"tenant_id,omitempty"`
}

// ValidateIdentifier 验证标识符参数
func (r *RegisterRequest) ValidateIdentifier() error {
	switch r.IdentifierType {
	case "email":
		if r.Email == "" {
			return errors.NewValidationError("邮箱不能为空")
		}
	case "phone":
		if r.Phone == "" {
			return errors.NewValidationError("手机号不能为空")
		}
		// 验证手机号格式
		if !regexp.MustCompile(constant.RegexPhone).MatchString(r.Phone) {
			return errors.NewValidationError("手机号格式不正确")
		}
	default:
		return errors.NewValidationError("标识符类型必须是 email 或 phone")
	}
	return nil
}

// GetIdentifier 获取标识符类型和值
func (r *RegisterRequest) GetIdentifier() (string, authConstant.IdentifierType) {
	switch r.IdentifierType {
	case "email":
		return r.Email, authConstant.IdentifierTypeEmail
	case "phone":
		return r.Phone, authConstant.IdentifierTypePhone
	default:
		return "", ""
	}
}

// ToRegisterParams 转换为注册参数
func (r *RegisterRequest) ToRegisterParams() (*param.RegisterParams, error) {
	// 验证标识符
	if err := r.ValidateIdentifier(); err != nil {
		return nil, err
	}

	// 获取标识符
	identifier, identifierType := r.GetIdentifier()

	// 验证密码确认
	if r.Password != r.ConfirmPassword {
		return nil, errors.NewValidationError("两次输入的密码不一致")
	}

	// 设置默认租户ID
	tenantID := r.TenantID
	if tenantID == "" {
		// todo 是否生成默认租户UUID并存在对应表中
		tenantID = constant.DefaultTenantID
	}

	// 创建注册参数
	params := param.NewRegisterParams(
		tenantID,       // TenantID
		r.Username,     // Username
		identifierType, // IdentifierType
		identifier,     // Identifier
		r.Password,     // Password
	)

	return params, nil
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	AccountID       string `json:"account_id"`
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// UpdateProfileRequest 更新个人信息请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

// EmailVerificationRequest 邮箱验证请求
type EmailVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyEmailRequest 验证邮箱请求
type VerifyEmailRequest struct {
	Token string `form:"token" binding:"required"`
}

// PasswordResetRequest 密码重置请求
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=6"`
}

// AdminResetPasswordRequest 管理员重置用户密码请求
type AdminResetPasswordRequest struct {
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// OAuthCallbackRequest OAuth回调请求
type OAuthCallbackRequest struct {
	Code  string `form:"code" binding:"required"`
	State string `form:"state"`
}

// TOTPSetupInitRequest 初始化2FA设置请求
type TOTPSetupInitRequest struct {
	AccountID string `json:"account_id" binding:"required"`
}

// TOTPSetupVerifyRequest 验证2FA设置请求
type TOTPSetupVerifyRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Secret    string `json:"secret" binding:"required"`
	Token     string `json:"token" binding:"required"`
}

// TOTPEnableRequest 启用2FA请求
type TOTPEnableRequest struct {
	AccountID   string   `json:"account_id" binding:"required"`
	Secret      string   `json:"secret" binding:"required"`
	Token       string   `json:"token" binding:"required"`
	BackupCodes []string `json:"backup_codes" binding:"required"`
}

// TOTPDisableRequest 禁用2FA请求
type TOTPDisableRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Token     string `json:"token" binding:"required"`
}

// TOTPVerifyRequest 验证2FA请求
type TOTPVerifyRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Token     string `json:"token" binding:"required"`
	UseBackup bool   `json:"use_backup"`
}

// TOTPBackupCodeRequest 使用备份码请求
type TOTPBackupCodeRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

// TOTPStatusRequest 获取2FA状态请求
type TOTPStatusRequest struct {
	AccountID string `json:"account_id" binding:"required"`
}

// TOTPChangeSecretRequest 更换2FA密钥请求
type TOTPChangeSecretRequest struct {
	AccountID   string   `json:"account_id" binding:"required"`
	Token       string   `json:"token" binding:"required"`
	NewSecret   string   `json:"new_secret" binding:"required"`
	BackupCodes []string `json:"backup_codes" binding:"required"`
}

// LogoutAllByTenantRequest 管理员按租户登出所有会话请求
type LogoutAllByTenantRequest struct {
	TenantID string `json:"tenant_id" binding:"required"`
	Reason   string `json:"reason,omitempty"`
}

// LogoutRequest 登出请求（支持登出所有租户选项）
type LogoutRequest struct {
	LogoutAllTenants bool   `json:"logout_all_tenants,omitempty"`
	Reason           string `json:"reason,omitempty"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Identifier string `json:"identifier" binding:"required"` // 支持邮箱或手机号
}

// RevokeSessionRequest 撤销会话请求
type RevokeSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// GetDevicesRequest 获取设备列表请求
type GetDevicesRequest struct {
	Page     int `form:"page,omitempty"`      // 页码，默认1
	PageSize int `form:"page_size,omitempty"` // 每页数量，默认20
}

// RevokeDeviceRequest 撤销设备请求
type RevokeDeviceRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

// TrustDeviceRequest 信任设备请求
type TrustDeviceRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

// GetSecurityLogsRequest 获取安全日志请求
type GetSecurityLogsRequest struct {
	StartDate string `form:"start_date,omitempty"` // 开始日期，格式：2023-01-01
	EndDate   string `form:"end_date,omitempty"`   // 结束日期，格式：2023-01-31
	Action    string `form:"action,omitempty"`     // 操作类型过滤
	Search    string `form:"search,omitempty"`     // 搜索关键字（IP、设备信息）
	Page      int    `form:"page,omitempty"`       // 页码，默认1
	PageSize  int    `form:"page_size,omitempty"`  // 每页数量，默认20
}
