package dm

import (
	"strings"
	"time"

	"auth-perm/internal/domain/permission/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionResourceDO 权限资源关联领域对象
type PermissionResourceDO struct {
	ID           string `gorm:"primaryKey;type:uuid"`
	PermissionID string `gorm:"column:permission_id;type:uuid;not null;index"`
	ResourceID   string `gorm:"column:resource_id;type:varchar(255);not null;index"`
	ResourceType string `gorm:"column:resource_type;type:varchar(50);not null;index"` // api_path, menu, button
	ResourceName string `gorm:"column:resource_name;type:varchar(255)"`
	TenantID     string `gorm:"column:tenant_id;type:uuid;not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// NewPermissionResource 创建新权限资源关联
func NewPermissionResource(permissionID, resourceID, resourceType, resourceName, tenantID string) *PermissionResourceDO {
	now := time.Now()

	return &PermissionResourceDO{
		ID:           uuid.New().String(),
		PermissionID: permissionID,
		ResourceID:   strings.TrimSpace(resourceID),
		ResourceType: strings.ToLower(strings.TrimSpace(resourceType)),
		ResourceName: strings.TrimSpace(resourceName),
		TenantID:     tenantID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// BeforeCreate GORM钩子
func (pr *PermissionResourceDO) BeforeCreate(tx *gorm.DB) error {
	// 如果ID为空，自动生成UUID
	if pr.ID == "" {
		pr.ID = uuid.New().String()
	}

	// 确保资源类型为小写
	pr.ResourceType = strings.ToLower(pr.ResourceType)

	return nil
}

// ToDTO 转换为DTO
func (pr *PermissionResourceDO) ToDTO() *dto.PermissionResourceDTO {
	if pr == nil {
		return nil
	}
	return &dto.PermissionResourceDTO{
		ID:           pr.ID,
		PermissionID: pr.PermissionID,
		ResourceID:   pr.ResourceID,
		ResourceType: pr.ResourceType,
		ResourceName: pr.ResourceName,
		TenantID:     pr.TenantID,
		CreatedAt:    pr.CreatedAt,
		UpdatedAt:    pr.UpdatedAt,
	}
}

// TableName 指定表名
func (pr *PermissionResourceDO) TableName() string {
	return "permission_resources"
}
