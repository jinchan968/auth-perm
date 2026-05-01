package vo

import (
	"time"

	"auth-perm/internal/domain/tenant/constant"
	"auth-perm/internal/domain/tenant/dto"
)

// ==================== 请求 ====================

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Name     string              `json:"name" binding:"required"`
	Code     string              `json:"code"`
	Plan     constant.TenantPlan `json:"plan"`
	ExpireAt *time.Time          `json:"expire_at"`
}

// UpdateTenantRequest 更新租户请求（包含基本信息和设置）
type UpdateTenantRequest struct {
	Name     string                 `json:"name"`
	Status   *constant.TenantStatus `json:"status"`
	Plan     *constant.TenantPlan   `json:"plan"`
	ExpireAt *time.Time             `json:"expire_at"`
	// Settings 可选，用于同时更新设置
	Settings *dto.TenantSettings `json:"settings"`
}

// DeleteTenantRequest 删除租户请求
type DeleteTenantRequest struct {
	ID string `json:"id" binding:"required"`
}

// GetTenantRequest 获取租户请求
type GetTenantRequest struct {
	ID string `json:"id" binding:"required"`
}

// ListTenantsRequest 列出租户请求
type ListTenantsRequest struct {
	Keyword string `json:"keyword"`
	Status  string `json:"status"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

// UpdateTenantSettingsRequest 更新租户设置请求
type UpdateTenantSettingsRequest struct {
	Settings dto.TenantSettings `json:"settings"`
}

// ChangeStatusRequest 变更租户状态请求
type ChangeStatusRequest struct {
	Status constant.TenantStatus `json:"status" binding:"required"`
}

// ==================== 响应 ====================

// TenantResponse 租户响应
type TenantResponse struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Code      string                `json:"code"`
	Status    constant.TenantStatus `json:"status"`
	Plan      constant.TenantPlan   `json:"plan"`
	ExpireAt  *time.Time            `json:"expire_at,omitempty"`
	Settings  dto.TenantSettings    `json:"settings"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// TenantListItemResponse 租户列表项响应
type TenantListItemResponse struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Code      string                `json:"code"`
	Status    constant.TenantStatus `json:"status"`
	Plan      constant.TenantPlan   `json:"plan"`
	ExpireAt  *time.Time            `json:"expire_at,omitempty"`
	UserCount int64                 `json:"user_count"`
	CreatedAt time.Time             `json:"created_at"`
}

// TenantListResponse 租户列表响应
type TenantListResponse struct {
	Data  []TenantListItemResponse `json:"data"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
}
