package dto

import (
	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/auth/constant"
)

// UserSearchQueryDTO 用户搜索查询
type UserSearchQueryDTO struct {
	TenantID   string               `json:"tenant_id"`
	Keyword    string               `json:"keyword"`
	Status     *constant.UserStatus `json:"status,omitempty"`
	RoleCode   string               `json:"role_code,omitempty"`
	OrgID      string               `json:"org_id,omitempty"`
	CreatedAt  *model.TimeRange     `json:"created_at,omitempty"`
	UpdatedAt  *model.TimeRange     `json:"updated_at,omitempty"`
	Pagination *model.Pagination    `json:"pagination"`
}

// UserStatsDTO 用户统计信息
type UserStatsDTO struct {
	TotalUsers    int64            `json:"total_users"`
	ActiveUsers   int64            `json:"active_users"`
	InactiveUsers int64            `json:"inactive_users"`
	NewUsers      int64            `json:"new_users"`
	UsersByStatus map[string]int64 `json:"users_by_status"`
	RecentLogins  int64            `json:"recent_logins"`
	AccountTypes  map[string]int64 `json:"account_types"`
}
