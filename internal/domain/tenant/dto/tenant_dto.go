package dto

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// TenantSettings 租户设置
type TenantSettings struct {
	// 功能开关
	Features FeaturesConfig `json:"features"`
	// 配额限制
	Quota QuotaConfig `json:"quota"`
	// 其他设置
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// FeaturesConfig 功能开关配置
type FeaturesConfig struct {
	EmailVerification  bool `json:"email_verification"`  // 邮箱验证
	OAuthLogin         bool `json:"oauth_login"`         // OAuth登录
	TOTPEnabled        bool `json:"totp_enabled"`        // TOTP双因子
	SessionLimit       bool `json:"session_limit"`       // 会话限制
	PasswordComplexity bool `json:"password_complexity"` // 密码复杂度要求
}

// QuotaConfig 配额配置
type QuotaConfig struct {
	MaxUsers           int `json:"max_users"`             // 最大用户数 (-1表示无限制)
	MaxRoles           int `json:"max_roles"`             // 最大角色数
	MaxOrganizations   int `json:"max_organizations"`     // 最大组织数
	MaxSessionsPerUser int `json:"max_sessions_per_user"` // 每用户最大会话数
	APIRateLimit       int `json:"api_rate_limit"`        // API速率限制(次/分钟)
}

// Scan 实现 sql.Scanner 接口
func (t *TenantSettings) Scan(value interface{}) error {
	if value == nil {
		*t = TenantSettings{
			Features: FeaturesConfig{},
			Quota:    QuotaConfig{},
			Custom:   make(map[string]interface{}),
		}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, t)
}

// Value 实现 driver.Valuer 接口
func (t TenantSettings) Value() (driver.Value, error) {
	if t.Custom == nil {
		t.Custom = make(map[string]interface{})
	}
	return json.Marshal(t)
}

// DefaultTenantSettings 返回默认租户设置
func DefaultTenantSettings() TenantSettings {
	return TenantSettings{
		Features: FeaturesConfig{
			EmailVerification:  true,
			OAuthLogin:         true,
			TOTPEnabled:        true,
			SessionLimit:       true,
			PasswordComplexity: true,
		},
		Quota: QuotaConfig{
			MaxUsers:           -1,
			MaxRoles:           100,
			MaxOrganizations:   50,
			MaxSessionsPerUser: 5,
			APIRateLimit:       1000,
		},
		Custom: make(map[string]interface{}),
	}
}

// ==================== 租户套餐 ====================

// TenantPlan 租户套餐
type TenantPlan string

const (
	TenantPlanFree       TenantPlan = "free"
	TenantPlanBasic      TenantPlan = "basic"
	TenantPlanPro        TenantPlan = "pro"
	TenantPlanEnterprise TenantPlan = "enterprise"
)

// IsValid 检查套餐是否有效
func (p TenantPlan) IsValid() bool {
	switch p {
	case TenantPlanFree, TenantPlanBasic, TenantPlanPro, TenantPlanEnterprise:
		return true
	}
	return false
}

// ==================== 租户状态 ====================

// TenantStatus 租户状态
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// IsValid 检查状态是否有效
func (s TenantStatus) IsValid() bool {
	switch s {
	case TenantStatusActive, TenantStatusSuspended, TenantStatusDeleted:
		return true
	}
	return false
}

// IsActive 检查租户是否活跃
func (s TenantStatus) IsActive() bool {
	return s == TenantStatusActive
}

// ==================== 租户 DTO ====================

// TenantDTO 租户数据传输对象
type TenantDTO struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Code      string         `json:"code"`
	Status    TenantStatus   `json:"status"`
	Plan      TenantPlan     `json:"plan"`
	ExpireAt  *time.Time     `json:"expire_at,omitempty"`
	Settings  TenantSettings `json:"settings"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// TenantListItemDTO 租户列表项DTO
type TenantListItemDTO struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Code      string       `json:"code"`
	Status    TenantStatus `json:"status"`
	Plan      TenantPlan   `json:"plan"`
	ExpireAt  *time.Time   `json:"expire_at,omitempty"`
	UserCount int64        `json:"user_count"`
	CreatedAt time.Time    `json:"created_at"`
}
