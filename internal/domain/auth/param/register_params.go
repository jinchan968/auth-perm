package param

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	commonConstant "auth-perm/internal/common/constant"
	"auth-perm/internal/domain/auth/constant"
)

// RegisterParams 通用注册参数（支持邮箱和手机号）
type RegisterParams struct {
	TenantID       string
	Username       string
	IdentifierType constant.IdentifierType
	Identifier     string
	Password       string
}

// IdentifierMatch 验证标识符是否匹配
func (r *RegisterParams) IdentifierMatch(email, phone string) bool {
	switch r.IdentifierType {
	case constant.IdentifierTypePhone:
		return strings.TrimSpace(r.Identifier) == strings.TrimSpace(phone)
	case constant.IdentifierTypeEmail:
		return strings.TrimSpace(r.Identifier) == strings.TrimSpace(email)
	}
	return false
}

// NewRegisterWithEmailParams FUTURE: 邮箱注册参数创建 - 在实现邮箱注册时使用
func NewRegisterWithEmailParams(tenantID, username, email, password string) *RegisterParams {
	return &RegisterParams{
		TenantID:       tenantID,
		Username:       username,
		IdentifierType: constant.IdentifierTypeEmail,
		Identifier:     email,
		Password:       password,
	}
}

// NewRegisterWithPhoneParams FUTURE: 手机号注册参数创建 - 在实现手机号注册时使用
func NewRegisterWithPhoneParams(tenantID, username, phone, password string) *RegisterParams {
	return &RegisterParams{
		TenantID:       tenantID,
		Username:       username,
		IdentifierType: constant.IdentifierTypePhone,
		Identifier:     phone,
		Password:       password,
	}
}

// NewRegisterParams 创建通用注册参数
func NewRegisterParams(tenantID, username string, identifierType constant.IdentifierType, identifier, password string) *RegisterParams {
	return &RegisterParams{
		TenantID:       tenantID,
		Username:       username,
		IdentifierType: identifierType,
		Identifier:     identifier,
		Password:       password,
	}
}

// Validate 验证注册参数
func (r *RegisterParams) Validate() error {
	if r.TenantID == "" {
		r.TenantID = commonConstant.DefaultTenantID
	}

	// 验证用户名
	if r.Username == "" {
		return errEmpty("用户名")
	}
	if err := r.validateUsername(); err != nil {
		return err
	}

	// 验证标识符
	if r.Identifier == "" {
		return errEmpty("标识符（邮箱或手机号）")
	}

	// 验证密码
	if r.Password == "" {
		return errEmpty("密码")
	}
	if len(r.Password) < 6 {
		return errTooShort("密码", 6)
	}

	// 验证标识符类型
	if !r.IdentifierType.IsValid() {
		return &ValidationError{Field: "标识符类型", Message: "必须是邮箱或手机号"}
	}

	return nil
}

// validateUsername 验证用户名的特殊性和安全性
func (r *RegisterParams) validateUsername() error {
	// 检查用户名是否在保留名单中
	usernameLower := strings.ToLower(r.Username)
	for _, reserved := range commonConstant.ReservedUsernames {
		if usernameLower == reserved {
			return &ValidationError{
				Field:   "用户名",
				Message: fmt.Sprintf("用户名 '%s' 是系统保留用户名，不允许注册", r.Username),
			}
		}
	}

	// 检查用户名格式（使用正则）
	if !regexp.MustCompile(commonConstant.RegexUsername).MatchString(r.Username) {
		return &ValidationError{
			Field:   "用户名",
			Message: "用户名只能包含字母、数字和下划线，长度3-50字符",
		}
	}

	// 检查不能全为数字
	if regexp.MustCompile(`^\d+$`).MatchString(r.Username) {
		return &ValidationError{
			Field:   "用户名",
			Message: "用户名不能全为数字",
		}
	}

	// 检查不能包含连续相同字符（如admin1111）
	// 使用手动检查方式，避免正则表达式反向引用问题
	if hasConsecutiveChars(r.Username, 4) {
		return &ValidationError{
			Field:   "用户名",
			Message: "用户名不能包含连续4个或以上相同字符",
		}
	}

	return nil
}

// hasConsecutiveChars 检查字符串是否包含连续N个相同字符
func hasConsecutiveChars(s string, n int) bool {
	if len(s) < n {
		return false
	}
	for i := 0; i <= len(s)-n; i++ {
		// 检查当前位置开始的n个字符是否都相同
		allSame := true
		for j := 1; j < n; j++ {
			if s[i] != s[i+j] {
				allSame = false
				break
			}
		}
		if allSame {
			return true
		}
	}
	return false
}

// errEmpty FUTURE: 空字段错误 - 在实现参数验证时使用
func errEmpty(field string) error {
	return &ValidationError{Field: field, Message: "不能为空"}
}

// errTooShort FUTURE: 字段过短错误 - 在实现参数验证时使用
func errTooShort(field string, minLen int) error {
	return &ValidationError{Field: field, Message: fmt.Sprintf("长度不能少于 %d 个字符", minLen)}
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// LoginParams 通用登录参数（支持邮箱和手机号）
type LoginParams struct {
	Identifier string
	Password   string
	DeviceID   string
	UserAgent  string
	IPAddress  string
	RememberMe bool
	TenantID   string // 租户ID，支持多租户登录
}

// NewLoginWithEmailParams 创建邮箱登录参数（向后兼容）
func NewLoginWithEmailParams(email, password, deviceID, userAgent, ipAddress, tenantID string, rememberMe bool) *LoginParams {
	return &LoginParams{
		Identifier: email,
		Password:   password,
		DeviceID:   deviceID,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
		RememberMe: rememberMe,
		TenantID:   tenantID,
	}
}

// NewLoginWithPhoneParams 创建手机号登录参数
func NewLoginWithPhoneParams(phone, password, deviceID, userAgent, ipAddress, tenantID string, rememberMe bool) *LoginParams {
	return &LoginParams{
		Identifier: phone,
		Password:   password,
		DeviceID:   deviceID,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
		RememberMe: rememberMe,
		TenantID:   tenantID,
	}
}

// NewLoginParams 创建通用登录参数
func NewLoginParams(identifier, password, deviceID, userAgent, ipAddress, tenantID string, rememberMe bool) *LoginParams {
	return &LoginParams{
		Identifier: identifier,
		Password:   password,
		DeviceID:   deviceID,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
		RememberMe: rememberMe,
		TenantID:   tenantID,
	}
}

// Validate 验证登录参数
func (p *LoginParams) Validate() error {
	if p.Identifier == "" {
		return errEmpty("登录标识（邮箱或手机号）")
	}
	if p.Password == "" {
		return errEmpty("请输入密码")
	}
	return nil
}

// LoginWithOAuthParams OAuth登录参数
type LoginWithOAuthParams struct {
	Provider string
	OAuthID  string
	Username string
}

// NewLoginWithOAuthParams 创建OAuth登录参数
func NewLoginWithOAuthParams(provider, oauthID, username string) *LoginWithOAuthParams {
	return &LoginWithOAuthParams{
		Provider: provider,
		OAuthID:  oauthID,
		Username: username,
	}
}

// Validate 验证OAuth登录参数
func (p *LoginWithOAuthParams) Validate() error {
	if p.Provider == "" {
		return errEmpty("提供商")
	}
	if p.OAuthID == "" {
		return errEmpty("OAuth ID")
	}
	if p.Username == "" {
		return errEmpty("用户名")
	}
	return nil
}

// ChangePasswordParams 修改密码参数
type ChangePasswordParams struct {
	UserID          string
	AccountID       string
	OldPassword     string
	NewPassword     string
	ConfirmPassword string
}

// NewChangePasswordParams 创建修改密码参数
func NewChangePasswordParams(userID, accountID, oldPassword, newPassword, confirmPassword string) *ChangePasswordParams {
	return &ChangePasswordParams{
		UserID:          userID,
		AccountID:       accountID,
		OldPassword:     oldPassword,
		NewPassword:     newPassword,
		ConfirmPassword: confirmPassword,
	}
}

// Validate 验证修改密码参数
func (p *ChangePasswordParams) Validate() error {
	if p.OldPassword == "" {
		return errEmpty("旧密码")
	}
	if p.NewPassword == "" {
		return errEmpty("新密码")
	}
	if p.NewPassword != p.ConfirmPassword {
		return &ValidationError{Field: "确认密码", Message: "与新密码不匹配"}
	}
	if len(p.NewPassword) < 6 {
		return errTooShort("新密码", 6)
	}
	return nil
}

// CreateSessionParams 创建会话参数
type CreateSessionParams struct {
	UserID    string
	AccountID string
	TenantID  string
	ExpiresIn time.Duration
	IPAddress string
	UserAgent string
}

// NewCreateSessionParams 创建会话参数
func NewCreateSessionParams(userID, accountID, tenantID, ipAddress, userAgent string, expiresIn time.Duration) *CreateSessionParams {
	return &CreateSessionParams{
		UserID:    userID,
		AccountID: accountID,
		TenantID:  tenantID,
		ExpiresIn: expiresIn,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

// Validate 验证创建会话参数
func (p *CreateSessionParams) Validate() error {
	if p.UserID == "" {
		return errEmpty("用户ID")
	}
	if p.AccountID == "" {
		return errEmpty("账户ID")
	}
	if p.ExpiresIn <= 0 {
		return &ValidationError{Field: "过期时间", Message: "必须大于0"}
	}
	return nil
}
