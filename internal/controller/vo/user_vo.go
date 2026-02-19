package vo

import (
	"time"

	"auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dto"
)

// UserWithAccountResponse 用户和账户组合响应
type UserWithAccountResponse struct {
	// Account 信息
	AccountID     string                 `json:"account_id"`
	TenantID      string                 `json:"tenant_id"`
	AccountType   constant.AccountType   `json:"account_type"`
	AccountStatus constant.AccountStatus `json:"account_status"`
	EmailVerified bool                   `json:"email_verified"`
	LastLoginAt   *time.Time             `json:"last_login_at,omitempty"`

	// User 信息
	UserID     string              `json:"user_id"`
	Username   string              `json:"username"`
	Nickname   string              `json:"nickname"`
	Avatar     string              `json:"avatar"`
	Email      string              `json:"email"`
	Phone      string              `json:"phone"`
	UserStatus constant.UserStatus `json:"user_status"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromUserWithAccountDTO 从DTO转换
func (r *UserWithAccountResponse) FromUserWithAccountDTO(dto *dto.UserWithAccountDTO) {
	r.AccountID = dto.AccountID
	r.TenantID = dto.TenantID
	r.AccountType = dto.AccountType
	r.AccountStatus = dto.AccountStatus
	r.EmailVerified = dto.EmailVerified
	r.LastLoginAt = dto.LastLoginAt
	r.UserID = dto.UserID
	r.Username = dto.Username
	r.Nickname = dto.Nickname
	r.Avatar = dto.Avatar
	r.Email = dto.Email
	r.Phone = dto.Phone
	r.UserStatus = dto.UserStatus
	r.CreatedAt = dto.CreatedAt
	r.UpdatedAt = dto.UpdatedAt
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	TenantID string                 `json:"tenant_id" binding:"required"`
	Status   constant.AccountStatus `json:"status" binding:"required"`
}
