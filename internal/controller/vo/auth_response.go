package vo

import (
	"time"

	"auth-perm/internal/domain/auth/dto"
)

// LoginResponse 登录响应（合并精简版）
type LoginResponse struct {
	Username      string    `json:"username"`       // 用户名
	Nickname      string    `json:"nickname"`       // 昵称
	Avatar        string    `json:"avatar"`         // 头像
	Status        string    `json:"status"`         // 用户状态
	AccountID     string    `json:"account_id"`     // 账户ID
	EmailVerified bool      `json:"email_verified"` // 邮箱是否验证
	Token         string    `json:"token"`          // 访问令牌
	Message       string    `json:"message"`        // 响应消息
	ExpiresAt     time.Time `json:"expires_at"`     // 过期时间
}

// FromUserDTO 从UserDTO创建LoginResponse
func (r *LoginResponse) FromUserDTO(user *dto.UserDTO, account *dto.AccountDTO, token string, message string, expiresAt time.Time) {
	r.Username = user.GetUsername()
	r.Nickname = user.GetNickname()
	r.Avatar = user.GetAvatar()
	r.Status = user.GetDisplayStatus()
	r.AccountID = account.GetID()
	r.EmailVerified = account.IsEmailVerified()
	r.Token = token
	r.Message = message
	r.ExpiresAt = expiresAt
}

// EmailVerificationResponse 邮箱验证响应
type EmailVerificationResponse struct {
	Message string `json:"message"`
	Sent    bool   `json:"sent"`
}

// VerifyEmailResponse 验证邮箱响应
type VerifyEmailResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Verified bool   `json:"verified"`
	Email    string `json:"email"`
}

// PasswordResetResponse 密码重置响应
type PasswordResetResponse struct {
	Message string `json:"message"`
	Sent    bool   `json:"sent"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
