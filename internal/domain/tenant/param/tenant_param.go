package param

import (
	"time"

	"auth-perm/internal/domain/tenant/constant"
	"auth-perm/internal/domain/tenant/dto"
)

// CreateTenantParams 创建租户参数
type CreateTenantParams struct {
	Name     string              `json:"name"`
	Code     string              `json:"code"`
	Plan     constant.TenantPlan `json:"plan"`
	ExpireAt *time.Time          `json:"expire_at"`
}

// Validate 验证创建参数
func (p *CreateTenantParams) Validate() error {
	if p.Name == "" {
		return errEmpty("租户名称")
	}
	// Code字段改为可选，后端自动生成
	if p.Plan == "" {
		p.Plan = constant.TenantPlanFree
	}
	if !p.Plan.IsValid() {
		return errInvalid("套餐")
	}
	return nil
}

// UpdateTenantParams 更新租户参数（包含基本信息和设置）
type UpdateTenantParams struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Status   *constant.TenantStatus `json:"status"`
	Plan     *constant.TenantPlan   `json:"plan"`
	ExpireAt *time.Time             `json:"expire_at"`
	Settings *dto.TenantSettings    `json:"settings"`
}

// Validate 验证更新参数
func (p *UpdateTenantParams) Validate() error {
	if p.ID == "" {
		return errEmpty("租户ID")
	}
	if p.Status != nil && !p.Status.IsValid() {
		return errInvalid("租户状态")
	}
	if p.Plan != nil && !p.Plan.IsValid() {
		return errInvalid("套餐")
	}
	return nil
}

// DeleteTenantParams 删除租户参数（禁用租户）
type DeleteTenantParams struct {
	ID string `json:"id"`
}

// Validate 验证删除参数
func (p *DeleteTenantParams) Validate() error {
	if p.ID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// EnableTenantParams 启用租户参数
type EnableTenantParams struct {
	ID string `json:"id"`
}

// Validate 验证启用参数
func (p *EnableTenantParams) Validate() error {
	if p.ID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// GetTenantParams 获取租户参数
type GetTenantParams struct {
	ID string `json:"id"`
}

// Validate 验证获取参数
func (p *GetTenantParams) Validate() error {
	if p.ID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// ListTenantsParams 列出租户参数
type ListTenantsParams struct {
	Keyword string `json:"keyword"`
	Status  string `json:"status"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

// Validate 验证列表参数
func (p *ListTenantsParams) Validate() error {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 || p.Size > 100 {
		p.Size = 10
	}
	return nil
}

// UpdateTenantSettingsParams 更新租户设置参数
type UpdateTenantSettingsParams struct {
	ID       string              `json:"id"`
	Settings *dto.TenantSettings `json:"settings"`
}

// Validate 验证设置参数
func (p *UpdateTenantSettingsParams) Validate() error {
	if p.ID == "" {
		return errEmpty("租户ID")
	}
	return nil
}

// ==================== 辅助函数 ====================

func errEmpty(field string) error {
	return &ValidationError{Field: field, Message: field + "不能为空"}
}

func errInvalid(field string) error {
	return &ValidationError{Field: field, Message: field + "无效"}
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
