package param

import (
	"time"

	"auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
)

// SendVerificationEmailParams 发送验证邮件参数
type SendVerificationEmailParams struct {
	Email string `json:"email"`
}

// NewSendVerificationEmailParams FUTURE: 发送验证邮件参数创建 - 在实现邮箱验证时使用
func NewSendVerificationEmailParams(email string) *SendVerificationEmailParams {
	return &SendVerificationEmailParams{
		Email: email,
	}
}

// Validate 验证参数
func (p *SendVerificationEmailParams) Validate() error {
	if p.Email == "" {
		return errors.NewValidationError("邮箱不能为空")
	}
	return nil
}

// VerifyEmailParams 验证邮箱参数
type VerifyEmailParams struct {
	Token string `json:"token"`
}

// NewVerifyEmailParams FUTURE: 验证邮箱参数创建 - 在实现邮箱验证时使用
func NewVerifyEmailParams(token string) *VerifyEmailParams {
	return &VerifyEmailParams{
		Token: token,
	}
}

// Validate 验证参数
func (p *VerifyEmailParams) Validate() error {
	if p.Token == "" {
		return errors.NewValidationError("令牌不能为空")
	}
	return nil
}

// ResendVerificationEmailParams 重新发送验证邮件参数
type ResendVerificationEmailParams struct {
	Email string `json:"email"`
}

// NewResendVerificationEmailParams FUTURE: 重新发送验证邮件参数创建 - 在实现邮箱验证时使用
func NewResendVerificationEmailParams(email string) *ResendVerificationEmailParams {
	return &ResendVerificationEmailParams{
		Email: email,
	}
}

// Validate 验证参数
func (p *ResendVerificationEmailParams) Validate() error {
	if p.Email == "" {
		return errors.NewValidationError("邮箱不能为空")
	}
	return nil
}

// RequestPasswordResetParams 请求密码重置参数
type RequestPasswordResetParams struct {
	Email string `json:"email"`
}

// NewRequestPasswordResetParams FUTURE: 请求密码重置参数创建 - 在实现密码重置时使用
func NewRequestPasswordResetParams(email string) *RequestPasswordResetParams {
	return &RequestPasswordResetParams{
		Email: email,
	}
}

// Validate 验证参数
func (p *RequestPasswordResetParams) Validate() error {
	if p.Email == "" {
		return errors.NewValidationError("邮箱不能为空")
	}
	return nil
}

// ResetPasswordParams 重置密码参数
type ResetPasswordParams struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// NewResetPasswordParams FUTURE: 重置密码参数创建 - 在实现密码重置时使用
func NewResetPasswordParams(token, newPassword string) *ResetPasswordParams {
	return &ResetPasswordParams{
		Token:       token,
		NewPassword: newPassword,
	}
}

// Validate 验证参数
func (p *ResetPasswordParams) Validate() error {
	if p.Token == "" {
		return errors.NewValidationError("令牌不能为空")
	}
	if len(p.NewPassword) < 6 {
		return errors.NewValidationError("新密码长度不能少于6位")
	}
	return nil
}

// EmailVerificationConfig 邮箱验证配置参数
type EmailVerificationConfig struct {
	Enabled           bool          `json:"enabled"`
	Required          bool          `json:"required"`
	ExpireDuration    time.Duration `json:"expire_duration"`
	ResendDelay       time.Duration `json:"resend_delay"`
	MaxResendAttempts int           `json:"max_resend_attempts"`
	Template          string        `json:"template"`
}

// DefaultEmailVerificationConfig FUTURE: 默认邮箱验证配置 - 在实现邮箱验证时使用
func DefaultEmailVerificationConfig() *EmailVerificationConfig {
	return &EmailVerificationConfig{
		Enabled:           true,
		Required:          true,
		ExpireDuration:    constant.TokenExpiryDefault,
		ResendDelay:       60 * time.Second,
		MaxResendAttempts: 3,
		Template:          constant.DefaultNickname,
	}
}

// IsValid 验证配置是否有效
func (c *EmailVerificationConfig) IsValid() error {
	if c.ExpireDuration <= 0 {
		return errors.NewValidationError("过期时间必须大于0")
	}
	if c.ResendDelay < 0 {
		return errors.NewValidationError("重发延迟不能为负数")
	}
	if c.MaxResendAttempts < 0 {
		return errors.NewValidationError("最大重发次数不能为负数")
	}
	return nil
}

// EmailSettings 邮箱设置参数
type EmailSettings struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
	UseTLS       bool   `json:"use_tls"`
	UseSSL       bool   `json:"use_ssl"`
}

// Validate 验证邮箱设置
func (e *EmailSettings) Validate() error {
	if e.SMTPHost == "" {
		return errors.NewValidationError("SMTP服务器地址不能为空")
	}
	if e.SMTPPort <= 0 || e.SMTPPort > 65535 {
		return errors.NewValidationError("SMTP端口无效")
	}
	if e.SMTPUsername == "" {
		return errors.NewValidationError("SMTP用户名不能为空")
	}
	if e.FromEmail == "" {
		return errors.NewValidationError("发件人邮箱不能为空")
	}
	return nil
}

// GetSMTPConfig 获取SMTP配置
func (e *EmailSettings) GetSMTPConfig() map[string]interface{} {
	return map[string]interface{}{
		"host":      e.SMTPHost,
		"port":      e.SMTPPort,
		"username":  e.SMTPUsername,
		"password":  e.SMTPPassword,
		"from":      e.FromEmail,
		"from_name": e.FromName,
		"use_tls":   e.UseTLS,
		"use_ssl":   e.UseSSL,
	}
}
