package dto

import (
	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/auth/constant"
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
