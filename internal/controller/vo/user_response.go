package vo

import (
	"time"

	"auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dto"
)

// UserResponse 用户响应结构体
type UserResponse struct {
	ID              string                  `json:"id"`
	Username        string                  `json:"username"`
	Nickname        string                  `json:"nickname"`
	Avatar          string                  `json:"avatar"`
	Phone           string                  `json:"phone"`
	IdentifierType  constant.IdentifierType `json:"identifier_type"`
	IdentifierValue string                  `json:"identifier_value"`
	Status          constant.UserStatus     `json:"status"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// FromUserDTO 从UserDTO创建UserResponse
func (r *UserResponse) FromUserDTO(user *dto.UserDTO) {
	r.ID = user.GetID()
	r.Username = user.GetUsername()
	r.Nickname = user.GetNickname()
	r.Avatar = user.GetAvatar()
	r.Phone = user.GetPhone()
	r.IdentifierType = user.GetIdentifierType()
	r.IdentifierValue = user.GetIdentifierValue()
	r.Status = user.GetStatus()
	r.CreatedAt = user.GetCreatedAt()
	r.UpdatedAt = user.GetUpdatedAt()
}

// AccountResponse 账户响应结构体
type AccountResponse struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	AccountType     constant.AccountType   `json:"account_type"`
	Status          constant.AccountStatus `json:"status"`
	EmailVerified   bool                   `json:"email_verified"`
	LastLoginAt     *time.Time             `json:"last_login_at"`
	EmailVerifiedAt *time.Time             `json:"email_verified_at"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// FromAccountDTO 从AccountDTO创建AccountResponse
func (r *AccountResponse) FromAccountDTO(account *dto.AccountDTO) {
	r.ID = account.GetID()
	r.TenantID = account.GetTenantID()
	r.AccountType = account.GetAccountType()
	r.Status = account.GetStatus()
	r.EmailVerified = account.IsEmailVerified()
	r.LastLoginAt = account.GetLastLoginAt()
	r.EmailVerifiedAt = account.GetEmailVerifiedAt()
	r.CreatedAt = account.GetCreatedAt()
	r.UpdatedAt = account.GetUpdatedAt()
}
