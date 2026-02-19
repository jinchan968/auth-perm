package dto

import (
	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/auth/constant"
	"time"
)

// AccountSearchQueryDTO 账户搜索查询
type AccountSearchQueryDTO struct {
	TenantID    string                  `json:"tenant_id"`
	Keyword     string                  `json:"keyword"`
	Status      *constant.AccountStatus `json:"status,omitempty"`
	AccountType *constant.AccountType   `json:"account_type,omitempty"`
	CreatedAt   *model.TimeRange        `json:"created_at,omitempty"`
	UpdatedAt   *model.TimeRange        `json:"updated_at,omitempty"`
	Pagination  *model.Pagination       `json:"pagination"`
}

// AccountStatsDTO 账户统计信息
type AccountStatsDTO struct {
	TotalAccounts    int64            `json:"total_accounts"`
	ActiveAccounts   int64            `json:"active_accounts"`
	EmailAccounts    int64            `json:"email_accounts"`
	OAuthAccounts    int64            `json:"oauth_accounts"`
	VerifiedAccounts int64            `json:"verified_accounts"`
	AccountTypes     map[string]int64 `json:"account_types"`
	ByProvider       map[string]int64 `json:"by_provider"`
	RecentLogins     int64            `json:"recent_logins"`
}

// UserWithAccountDTO 用户和账户组合信息
type UserWithAccountDTO struct {
	// Account 信息
	AccountID     string                 `json:"account_id"`
	TenantID      string                 `json:"tenant_id"`
	AccountType   constant.AccountType   `json:"account_type"`
	AccountStatus constant.AccountStatus `json:"account_status"`
	EmailVerified bool                   `json:"email_verified"`
	LastLoginAt   *time.Time             `json:"last_login_at"`

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
