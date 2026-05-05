package dto

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	commonConstant "auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	authConstant "auth-perm/internal/domain/auth/constant"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AccountDTO 账户数据传输对象（包含业务逻辑）
type AccountDTO struct {
	// 基本信息
	ID          string                   `json:"id"`
	UserID      string                   `json:"user_id"`
	TenantID    string                   `json:"tenant_id"`
	AccountType authConstant.AccountType `json:"account_type"`

	// OAuth信息
	OAuthID       string `json:"oauth_id"`
	OAuthProvider string `json:"oauth_provider"`

	// 状态信息
	Status authConstant.AccountStatus `json:"status"`

	// 登录信息
	lastLoginAt *time.Time `json:"-"`
	UserAgent   string     `json:"user_agent"`
	ipAddress   *string    `json:"-"`

	// 设备信息
	DeviceInfo DeviceInfoDTO `json:"device_info"`

	// 验证状态
	EmailVerified   bool       `json:"email_verified"`
	emailVerifiedAt *time.Time `json:"-"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewAccountDTO 创建账户DTO
func NewAccountDTO(userID, tenantID string, accountType authConstant.AccountType) *AccountDTO {
	now := time.Now()
	return &AccountDTO{
		ID:          uuid.New().String(),
		UserID:      userID,
		TenantID:    tenantID,
		AccountType: accountType,
		Status:      authConstant.AccountStatusActive,
		DeviceInfo:  DeviceInfoDTO{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ==================== Getter 方法 ====================

// GetID 获取账户ID
func (a *AccountDTO) GetID() string {
	return a.ID
}

// GetUserID 获取用户ID
func (a *AccountDTO) GetUserID() string {
	return a.UserID
}

// GetTenantID 获取租户ID
func (a *AccountDTO) GetTenantID() string {
	return a.TenantID
}

// GetAccountType 获取账户类型
func (a *AccountDTO) GetAccountType() authConstant.AccountType {
	return a.AccountType
}

// GetOAuthID 获取OAuth ID
func (a *AccountDTO) GetOAuthID() string {
	return a.OAuthID
}

// GetOAuthProvider 获取OAuth提供商
func (a *AccountDTO) GetOAuthProvider() string {
	return a.OAuthProvider
}

// GetStatus 获取账户状态
func (a *AccountDTO) GetStatus() authConstant.AccountStatus {
	return a.Status
}

// GetLastLoginAt 获取最后登录时间
func (a *AccountDTO) GetLastLoginAt() *time.Time {
	return a.lastLoginAt
}

// GetUserAgent 获取用户代理
func (a *AccountDTO) GetUserAgent() string {
	return a.UserAgent
}

// GetIPAddress 获取IP地址
func (a *AccountDTO) GetIPAddress() *string {
	return a.ipAddress
}

// GetDeviceInfo 获取设备信息
func (a *AccountDTO) GetDeviceInfo() *DeviceInfoDTO {
	return &a.DeviceInfo
}

// IsEmailVerified 检查邮箱是否已验证
func (a *AccountDTO) IsEmailVerified() bool {
	return a.EmailVerified
}

// GetEmailVerifiedAt 获取邮箱验证时间
func (a *AccountDTO) GetEmailVerifiedAt() *time.Time {
	return a.emailVerifiedAt
}

// GetCreatedAt 获取创建时间
func (a *AccountDTO) GetCreatedAt() time.Time {
	return a.CreatedAt
}

// GetUpdatedAt 获取更新时间
func (a *AccountDTO) GetUpdatedAt() time.Time {
	return a.UpdatedAt
}

// ==================== Setter 方法 ====================

// SetUserID 设置用户ID
func (a *AccountDTO) SetUserID(userID string) error {
	a.UserID = userID
	a.UpdatedAt = time.Now()
	return nil
}

// SetLastLoginAt 设置最后登录时间
func (a *AccountDTO) SetLastLoginAt(t time.Time) {
	a.lastLoginAt = &t
	a.UpdatedAt = time.Now()
}

// SetIPAddress 设置IP地址
func (a *AccountDTO) SetIPAddress(ip string) {
	a.ipAddress = &ip
	a.UpdatedAt = time.Now()
}

// SetEmailVerifiedAt 设置邮箱验证时间
func (a *AccountDTO) SetEmailVerifiedAt(t time.Time) {
	a.emailVerifiedAt = &t
	a.UpdatedAt = time.Now()
}

// ==================== 业务方法 ====================

// IsActive 检查账户是否活跃
func (a *AccountDTO) IsActive() bool {
	return a.Status.IsActive()
}

// CanLogin 检查账户是否可以登录
func (a *AccountDTO) CanLogin() bool {
	if !a.IsActive() {
		return false
	}

	// OAuth账户需要有效的OAuth信息
	if a.AccountType.IsOAuth() {
		return a.OAuthID != "" && a.OAuthProvider != ""
	}

	// Email/Phone账户总是可以登录（密码验证在User层面处理）
	return true
}

// UpdateLastLogin 更新最后登录时间
func (a *AccountDTO) UpdateLastLogin() {
	now := time.Now()
	a.lastLoginAt = &now
	a.UpdatedAt = now
}

// Deactivate 停用账户
func (a *AccountDTO) Deactivate(reason string) error {
	if !a.IsActive() {
		return errors.NewBusinessError("账户已经是停用状态")
	}

	a.Status = authConstant.AccountStatusInactive
	a.UpdatedAt = time.Now()
	return nil
}

// Suspend 暂停账户
func (a *AccountDTO) Suspend(reason string) error {
	if a.Status.IsSuspended() {
		return errors.NewBusinessError("账户已经是暂停状态")
	}

	a.Status = authConstant.AccountStatusSuspended
	a.UpdatedAt = time.Now()
	return nil
}

// Reactivate 重新激活账户
func (a *AccountDTO) Reactivate() error {
	if a.Status.IsActive() {
		return errors.NewBusinessError("账户已经是激活状态")
	}

	a.Status = authConstant.AccountStatusActive
	a.UpdatedAt = time.Now()
	return nil
}

// VerifyEmail 验证邮箱
func (a *AccountDTO) VerifyEmail() {
	if a.EmailVerified {
		return // 已经验证过
	}

	now := time.Now()
	a.EmailVerified = true
	a.emailVerifiedAt = &now
	a.UpdatedAt = now
}

// LinkOAuth 链接OAuth账户
func (a *AccountDTO) LinkOAuth(oauthID, oauthProvider string) error {
	if oauthID == "" || oauthProvider == "" {
		return errors.NewValidationError("OAuth ID和提供商不能为空")
	}

	// 检查OAuth提供商是否有效
	if !isValidOAuthProvider(oauthProvider) {
		return errors.NewValidationError("无效的OAuth提供商")
	}

	a.OAuthID = oauthID
	a.OAuthProvider = oauthProvider
	a.UpdatedAt = time.Now()
	return nil
}

// UnlinkOAuth 取消链接OAuth账户
func (a *AccountDTO) UnlinkOAuth() {
	a.OAuthID = ""
	a.OAuthProvider = ""
	a.UpdatedAt = time.Now()
}

// HasOAuthLinked 检查是否已链接OAuth
func (a *AccountDTO) HasOAuthLinked() bool {
	return a.OAuthID != "" && a.OAuthProvider != ""
}

// GetLoginMethods 获取可用的登录方式
func (a *AccountDTO) GetLoginMethods() []string {
	methods := make([]string, 0)

	// Email/Phone账户总是支持密码登录
	if !a.AccountType.IsOAuth() {
		methods = append(methods, authConstant.LoginMethodPassword)
	}

	if a.HasOAuthLinked() {
		methods = append(methods, strings.ToLower(a.OAuthProvider))
	}

	return methods
}

// IsOAuthAccount 检查是否为OAuth账户
func (a *AccountDTO) IsOAuthAccount() bool {
	return a.AccountType.IsOAuth()
}

// GetSecurityScore 获取账户安全评分
func (a *AccountDTO) GetSecurityScore() int {
	score := 0

	// 账户活跃
	if a.IsActive() {
		score += 20
	}

	// 邮箱验证
	if a.EmailVerified {
		score += 25
	}

	// OAuth链接（增加安全性）
	if a.HasOAuthLinked() {
		score += 30
	}

	// 最近有登录活动
	if a.lastLoginAt != nil && time.Since(*a.lastLoginAt) < commonConstant.SessionExpiryLong {
		score += 15
	}

	// 账户类型（非OAuth账户有密码保护）
	if !a.AccountType.IsOAuth() {
		score += 10
	}

	return score
}

// IsSecure 检查账户是否安全
func (a *AccountDTO) IsSecure() bool {
	return a.GetSecurityScore() >= 70
}

// GetSecurityRecommendations 获取安全建议
func (a *AccountDTO) GetSecurityRecommendations() []string {
	recommendations := make([]string, 0)

	if !a.EmailVerified {
		recommendations = append(recommendations, "验证您的邮箱地址以保护账户安全")
	}

	if !a.HasOAuthLinked() && !a.AccountType.IsOAuth() {
		recommendations = append(recommendations, "链接第三方账户以增强登录安全性")
	}

	if a.GetSecurityScore() < 50 {
		recommendations = append(recommendations, "您的账户安全性较低，建议采取更多安全措施")
	}

	return recommendations
}

// GenerateEmailVerificationToken 生成邮箱验证令牌
func (a *AccountDTO) GenerateEmailVerificationToken() (string, error) {
	return GenerateSecureToken(commonConstant.EmailVerificationTokenLength)
}

// GeneratePasswordResetToken 生成密码重置令牌
func (a *AccountDTO) GeneratePasswordResetToken() (string, error) {
	return GenerateSecureToken(commonConstant.PasswordResetTokenLength)
}

// 私有方法

// isStrongPassword 检查密码强度
func (a *AccountDTO) isStrongPassword(password string) bool {
	hasLetter := false
	hasNumber := false

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z':
			hasLetter = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}
	}

	return hasLetter && hasNumber
}

// isValidEmail 验证邮箱格式
func (a *AccountDTO) isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// isValidOAuthProvider FUTURE: OAuth提供商验证 - 在实现OAuth验证时使用
func isValidOAuthProvider(provider string) bool {
	validProviders := []string{string(authConstant.AccountTypeGitHub), string(authConstant.AccountTypeGoogle), string(authConstant.AccountTypeWeChat)}
	for _, valid := range validProviders {
		if provider == valid {
			return true
		}
	}
	return false
}

// GenerateSecureToken 生成安全令牌
func GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		length = commonConstant.DefaultTokenLength
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", errors.WrapBizError(err, "生成安全令牌失败")
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashPassword 哈希密码
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.NewValidationError("密码不能为空")
	}

	// 使用bcrypt哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.WrapBizError(err, "哈希密码失败")
	}

	return string(hashedPassword), nil
}

// VerifyPassword FUTURE: 密码验证 - 在实现密码验证时使用
func VerifyPassword(hashedPassword, password string) error {
	if hashedPassword == "" || password == "" {
		return errors.NewValidationError("密码不能为空")
	}

	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return errors.NewAuthError("密码错误")
	}

	return nil
}

// SetOAuthInfo 设置OAuth信息
func (a *AccountDTO) SetOAuthInfo(provider, oauthID string) {
	a.OAuthProvider = provider
	a.OAuthID = oauthID
	a.UpdatedAt = time.Now()
}
