package dto

import (
	"fmt"
	"time"

	"auth-perm/internal/common/constant"
)

// EmailVerificationRequest 邮箱验证请求
type EmailVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// EmailVerificationResponse 邮箱验证响应
type EmailVerificationResponse struct {
	Message string `json:"message"`
	Sent    bool   `json:"sent"`
}

// VerifyEmailRequest 验证邮箱请求
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// VerifyEmailResponse 验证邮箱响应
type VerifyEmailResponse struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message"`
	Verified   bool       `json:"verified"`
	Email      string     `json:"email"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}

// EmailVerificationDTO 邮箱验证数据传输对象
type EmailVerificationDTO struct {
	ID         string     `json:"id"`
	AccountID  string     `json:"account_id"`
	Email      string     `json:"email"`
	Token      string     `json:"token"`
	ExpiresAt  time.Time  `json:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// NewEmailVerificationDTO FUTURE: 邮箱验证DTO创建 - 在实现邮箱验证时使用
func NewEmailVerificationDTO(accountID, email, token string, expiresAt time.Time) *EmailVerificationDTO {
	return &EmailVerificationDTO{
		AccountID: accountID,
		Email:     email,
		Token:     token,
		ExpiresAt: expiresAt,
	}
}

// IsExpired 检查验证链接是否过期
func (v *EmailVerificationDTO) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

// IsVerified 检查邮箱是否已验证
func (v *EmailVerificationDTO) IsVerified() bool {
	return v.VerifiedAt != nil
}

// MarkAsVerified 标记为已验证
func (v *EmailVerificationDTO) MarkAsVerified() {
	now := time.Now()
	v.VerifiedAt = &now
	v.UpdatedAt = now
}

// ToVerificationStatus 转换为验证状态
func (v *EmailVerificationDTO) ToVerificationStatus() VerificationStatus {
	if v.IsVerified() {
		return VerificationStatusVerified
	}
	if v.IsExpired() {
		return VerificationStatusExpired
	}
	return VerificationStatusPending
}

// VerificationStatus 验证状态
type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusVerified VerificationStatus = "verified"
	VerificationStatusExpired  VerificationStatus = "expired"
)

// IsValid 验证状态是否有效
func (s VerificationStatus) IsValid() bool {
	switch s {
	case VerificationStatusPending, VerificationStatusVerified, VerificationStatusExpired:
		return true
	default:
		return false
	}
}

// String 转换为字符串
func (s VerificationStatus) String() string {
	return string(s)
}

// PasswordResetRequest 密码重置请求
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetResponse 密码重置响应
type PasswordResetResponse struct {
	Message string `json:"message"`
	Sent    bool   `json:"sent"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PasswordResetDTO 密码重置数据传输对象
type PasswordResetDTO struct {
	ID        string     `json:"id"`
	AccountID string     `json:"account_id"`
	Email     string     `json:"email"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// NewPasswordResetDTO FUTURE: 密码重置DTO创建 - 在实现密码重置时使用
func NewPasswordResetDTO(accountID, email, token string, expiresAt time.Time) *PasswordResetDTO {
	return &PasswordResetDTO{
		AccountID: accountID,
		Email:     email,
		Token:     token,
		ExpiresAt: expiresAt,
	}
}

// IsExpired 检查重置链接是否过期
func (p *PasswordResetDTO) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// IsUsed 检查是否已使用
func (p *PasswordResetDTO) IsUsed() bool {
	return p.UsedAt != nil
}

// MarkAsUsed 标记为已使用
func (p *PasswordResetDTO) MarkAsUsed() {
	now := time.Now()
	p.UsedAt = &now
	p.UpdatedAt = now
}

// IsValid 检查重置令牌是否有效
func (p *PasswordResetDTO) IsValid() bool {
	return !p.IsExpired() && !p.IsUsed()
}

// EmailVerificationSettings 邮箱验证设置
type EmailVerificationSettings struct {
	Enabled        bool          `json:"enabled"`
	RequiredFor    []string      `json:"required_for"`
	ExpireDuration time.Duration `json:"expire_duration"`
	ResendDelay    time.Duration `json:"resend_delay"`
}

// DefaultEmailVerificationSettings FUTURE: 默认邮箱验证设置 - 在实现邮箱验证时使用
func DefaultEmailVerificationSettings() *EmailVerificationSettings {
	return &EmailVerificationSettings{
		Enabled:        true,
		RequiredFor:    []string{"register", "password_reset"},
		ExpireDuration: constant.TokenExpiryDefault,
		ResendDelay:    60 * time.Second,
	}
}

// EmailTemplate 邮件模板
type EmailTemplate struct {
	Name    string
	Subject string
	Body    string
	IsHTML  bool
}

// BuildVerificationEmail FUTURE: 验证邮件构建 - 在实现邮件发送时使用
func BuildVerificationEmail(username, verificationURL string) *EmailTemplate {
	subject := "验证您的邮箱地址"
	body := fmt.Sprintf(`
		亲爱的 %s，

		感谢您注册我们的服务！

		请点击以下链接验证您的邮箱地址：
		%s

		此链接将在24小时后过期。

		如果您没有注册此账户，请忽略此邮件。

		谢谢！
	`, username, verificationURL)

	return &EmailTemplate{
		Name:    "verification",
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	}
}

// BuildPasswordResetEmail FUTURE: 密码重置邮件构建 - 在实现邮件发送时使用
func BuildPasswordResetEmail(username, resetURL string) *EmailTemplate {
	subject := "重置您的密码"
	body := fmt.Sprintf(`
		亲爱的 %s，

		我们收到了重置您账户密码的请求。

		请点击以下链接重置您的密码：
		%s

		此链接将在1小时后过期。

		如果您没有请求重置密码，请忽略此邮件。

		谢谢！
	`, username, resetURL)

	return &EmailTemplate{
		Name:    "password_reset",
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	}
}

// BuildWelcomeEmail FUTURE: 欢迎邮件构建 - 在实现邮件发送时使用
func BuildWelcomeEmail(username string) *EmailTemplate {
	subject := "欢迎加入我们！"
	body := fmt.Sprintf(`
		亲爱的 %s，

		欢迎加入我们的平台！

		您的账户已经成功创建。

		请点击以下链接验证您的邮箱地址以获得完整体验：
		`+constant.FullURL(constant.SendVerificationEmailPath)+`

		谢谢！
	`, username)

	return &EmailTemplate{
		Name:    "welcome",
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	}
}
