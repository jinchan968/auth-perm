package dm

import (
	"auth-perm/internal/domain/tenant/dto"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantDO 租户领域对象
type TenantDO struct {
	ID        string             `gorm:"primaryKey;type:uuid"`
	Name      string             `gorm:"column:name;not null"`
	Code      string             `gorm:"column:code;uniqueIndex;not null"`
	Status    dto.TenantStatus   `gorm:"column:status;not null;default:active"`
	Plan      dto.TenantPlan     `gorm:"column:plan;not null;default:free"`
	ExpireAt  *time.Time         `gorm:"column:expire_at"`
	Settings  dto.TenantSettings `gorm:"column:settings;type:jsonb;default:'{}'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 指定表名
func (*TenantDO) TableName() string {
	return "tenants"
}

// BeforeCreate GORM钩子函数：创建前
func (t *TenantDO) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.Status == "" {
		t.Status = dto.TenantStatusActive
	}
	if t.Plan == "" {
		t.Plan = dto.TenantPlanFree
	}
	// 设置默认设置（检查 Features 和 Quota 是否都是零值）
	if t.Settings.Features == (dto.FeaturesConfig{}) && t.Settings.Quota == (dto.QuotaConfig{}) {
		t.Settings = dto.DefaultTenantSettings()
	}
	return nil
}

// BeforeUpdate GORM钩子函数：更新前
func (t *TenantDO) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("updated_at", time.Now())
	return nil
}

// ToDTO 转换为DTO
func (t *TenantDO) ToDTO() *dto.TenantDTO {
	if t == nil {
		return nil
	}
	return &dto.TenantDTO{
		ID:        t.ID,
		Name:      t.Name,
		Code:      t.Code,
		Status:    t.Status,
		Plan:      t.Plan,
		ExpireAt:  t.ExpireAt,
		Settings:  t.Settings,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
