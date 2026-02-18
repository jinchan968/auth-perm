package dm

import (
	"auth-perm/internal/domain/permission/dto"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationDO 组织领域对象（纯数据模型）
type OrganizationDO struct {
	ID          string                      `gorm:"primaryKey;type:uuid"`
	TenantID    string                      `gorm:"column:tenant_id;type:uuid;not null;index"`
	ParentID    *string                     `gorm:"column:parent_id;type:uuid;index"` // 父组织ID，支持层级结构
	Name        string                      `gorm:"column:name;not null"`
	Code        string                      `gorm:"column:code;uniqueIndex:idx_tenant_code;not null"`
	Description string                      `gorm:"column:description;type:text"`
	Level       int                         `gorm:"column:level;default:1"` // 组织层级，1为顶层
	Path        string                      `gorm:"column:path;index"`      // 组织路径，如：/org1/org2/org3
	IsActive    bool                        `gorm:"column:is_active;default:true"`
	SortOrder   int                         `gorm:"column:sort_order;default:0"`             // 排序顺序
	Metadata    dto.OrganizationMetadataDTO `gorm:"column:metadata;type:jsonb;default:'{}'"` // 额外的元数据
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Parent   *OrganizationDO  `gorm:"foreignKey:ParentID"`
	Children []OrganizationDO `gorm:"foreignKey:ParentID"`
	Roles    []RoleDO         `gorm:"foreignKey:OrgID"`
}

// TableName 指定表名
func (*OrganizationDO) TableName() string {
	return "organizations"
}

// BeforeCreate GORM钩子函数：创建前
func (o *OrganizationDO) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	if o.Path == "" && o.Code != "" {
		o.Path = "/" + o.Code
	}
	if o.Level == 0 {
		o.Level = 1
	}
	if o.IsActive {
		// 默认激活
	}
	return nil
}

// BeforeUpdate GORM钩子函数：更新前
func (o *OrganizationDO) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("updated_at", time.Now())
	return nil
}

// SetMetadata 设置组织元数据
func (o *OrganizationDO) SetMetadata(metadata *dto.OrganizationMetadataDTO) {
	o.Metadata = *metadata
}

// GetMetadata 获取组织元数据
func (o *OrganizationDO) GetMetadata() *dto.OrganizationMetadataDTO {
	return &o.Metadata
}

// ToDTO 转换为DTO（避免循环导入，在dm层定义）
func (o *OrganizationDO) ToDTO() *dto.OrganizationDTO {
	if o == nil {
		return nil
	}
	return &dto.OrganizationDTO{
		ID:          o.ID,
		TenantID:    o.TenantID,
		ParentID:    o.ParentID,
		Name:        o.Name,
		Code:        o.Code,
		Description: o.Description,
		Level:       o.Level,
		Path:        o.Path,
		IsActive:    o.IsActive,
		SortOrder:   o.SortOrder,
		Metadata:    o.Metadata,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
		IsRoot_:     o.ParentID == nil || *o.ParentID == "",
		IsLeaf_:     false,
		UserCount:   0,
		RoleCount:   0,
		ChildCount:  0,
	}
}
