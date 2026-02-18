package dto

import (
	"strings"

	"auth-perm/internal/common/errors"
	"time"
)

// OrganizationDTO 组织数据传输对象
type OrganizationDTO struct {
	ID          string                  `json:"id"`
	TenantID    string                  `json:"tenant_id"`
	ParentID    *string                 `json:"parent_id,omitempty"`
	Name        string                  `json:"name"`
	Code        string                  `json:"code"`
	Description string                  `json:"description"`
	Level       int                     `json:"level"`
	Path        string                  `json:"path"`
	IsActive    bool                    `json:"is_active"`
	SortOrder   int                     `json:"sort_order"`
	Metadata    OrganizationMetadataDTO `json:"metadata"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`

	// 统计信息
	IsRoot_    bool `json:"is_root"`     // 临时字段，避免与方法名冲突
	IsLeaf_    bool `json:"is_leaf"`     // 临时字段，避免与方法名冲突
	UserCount  int  `json:"user_count"`  // 用户数量
	RoleCount  int  `json:"role_count"`  // 角色数量
	ChildCount int  `json:"child_count"` // 子组织数量

	Parent *ParentOrgDTO `json:"parent,omitempty"`
}

// IsRoot 获取是否为根组织（从临时字段）
func (o *OrganizationDTO) IsRoot() bool {
	return o.IsRoot_
}

// IsLeaf 获取是否为叶子节点（从临时字段）
func (o *OrganizationDTO) IsLeaf() bool {
	return o.IsLeaf_
}

// GetUserCount 获取用户数量
func (o *OrganizationDTO) GetUserCount() int {
	return o.UserCount
}

// GetRoleCount 获取角色数量
func (o *OrganizationDTO) GetRoleCount() int {
	return o.RoleCount
}

// GetChildCount 获取子组织数量
func (o *OrganizationDTO) GetChildCount() int {
	return o.ChildCount
}

// ParentOrgDTO 父组织信息
type ParentOrgDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// OrganizationTreeNodeDTO 组织树节点
type OrganizationTreeNodeDTO struct {
	ID          string                    `json:"id"`
	TenantID    string                    `json:"tenant_id"`
	ParentID    *string                   `json:"parent_id,omitempty"`
	Name        string                    `json:"name"`
	Code        string                    `json:"code"`
	Description string                    `json:"description"`
	Level       int                       `json:"level"`
	Path        string                    `json:"path"`
	IsActive    bool                      `json:"is_active"`
	SortOrder   int                       `json:"sort_order"`
	Metadata    OrganizationMetadataDTO   `json:"metadata"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	IsRoot      bool                      `json:"is_root"`
	IsLeaf      bool                      `json:"is_leaf"`
	UserCount   int                       `json:"user_count"`
	RoleCount   int                       `json:"role_count"`
	ChildCount  int                       `json:"child_count"`
	Parent      *ParentOrgDTO             `json:"parent,omitempty"`
	Children    []OrganizationTreeNodeDTO `json:"children,omitempty"`
}

// GetIsRoot 获取是否为根组织
func (n *OrganizationTreeNodeDTO) GetIsRoot() bool {
	return n.IsRoot
}

// GetIsLeaf 获取是否为叶子节点
func (n *OrganizationTreeNodeDTO) GetIsLeaf() bool {
	return n.IsLeaf
}

// OrganizationSummaryDTO 组织摘要信息
type OrganizationSummaryDTO struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Code       string    `json:"code"`
	Level      int       `json:"level"`
	IsActive   bool      `json:"is_active"`
	UserCount  int       `json:"user_count"`
	RoleCount  int       `json:"role_count"`
	ChildCount int       `json:"child_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// 转换函数已移动到 transformer 包以避免循环导入问题
/*
func FromOrganizationDO(org *dm.OrganizationDO) *OrganizationDTO {
	if org == nil {
		return nil
	}

	dto := &OrganizationDTO{
		ID:          org.ID,
		TenantID:    org.TenantID,
		ParentID:    org.ParentID,
		Name:        org.Name,
		Code:        org.Code,
		Description: org.Description,
		Level:       org.Level,
		Path:        org.Path,
		IsActive:    org.IsActive,
		SortOrder:   org.SortOrder,
		Metadata:    org.Metadata,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
		IsRoot_:     org.ParentID == nil || *org.ParentID == "",
		IsLeaf_:     false,
		UserCount:   0,
		RoleCount:   0,
		ChildCount:  0,
	}

	if org.Parent != nil {
		dto.Parent = &ParentOrgDTO{
			ID:   org.Parent.ID,
			Name: org.Parent.Name,
			Code: org.Parent.Code,
		}
	}

	return dto
}
*/

// FromOrganizationDOToTreeNode FUTURE: 组织树节点转换 - 在实现组织树时使用
/*
func FromOrganizationDOToTreeNode(org *dm.OrganizationDO) *OrganizationTreeNodeDTO {
	if org == nil {
		return nil
	}

	node := &OrganizationTreeNodeDTO{
		ID:          org.ID,
		TenantID:    org.TenantID,
		ParentID:    org.ParentID,
		Name:        org.Name,
		Code:        org.Code,
		Description: org.Description,
		Level:       org.Level,
		Path:        org.Path,
		IsActive:    org.IsActive,
		SortOrder:   org.SortOrder,
		Metadata:    org.Metadata,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
		IsRoot:      org.ParentID == nil || *org.ParentID == "", // 直接计算IsRoot
		IsLeaf:      false,                                      // 需要外部数据才能判断，临时设为false
		UserCount:   0,                                          // 需要外部查询，临时设为0
		RoleCount:   0,                                          // 需要外部查询，临时设为0
		ChildCount:  0,                                          // 需要外部查询，临时设为0
	}

	// 添加父组织信息
	if org.Parent != nil {
		node.Parent = &ParentOrgDTO{
			ID:   org.Parent.ID,
			Name: org.Parent.Name,
			Code: org.Parent.Code,
		}
	}

	// 递归添加子组织
	if len(org.Children) > 0 {
		children := make([]OrganizationTreeNodeDTO, 0)
		for _, child := range org.Children {
			if child.IsActive {
				children = append(children, *FromOrganizationDOToTreeNode(&child))
			}
		}
		node.Children = children
		node.IsLeaf = len(children) == 0 // 根据Children判断是否为叶子节点
	}

	return node
}
*/

// ==================== 业务方法（从DO迁移） ====================

// UpdateInfo 更新组织信息
func (o *OrganizationDTO) UpdateInfo(name, description string) error {
	if name != "" {
		if len(name) > 200 {
			return errors.NewValidationError("组织名称过长，最大200个字符")
		}
		o.Name = strings.TrimSpace(name)
	}

	if description != "" {
		if len(description) > 1000 {
			return errors.NewValidationError("描述过长，最大1000个字符")
		}
		o.Description = strings.TrimSpace(description)
	}

	return nil
}

// SetSortOrder 设置排序顺序
func (o *OrganizationDTO) SetSortOrder(order int) {
	o.SortOrder = order
}

// SetMetadata 设置元数据
func (o *OrganizationDTO) SetMetadata(key string, value interface{}) error {
	if key == "" {
		return errors.NewValidationError("元数据键不能为空")
	}

	if o.Metadata.CustomFields == nil {
		o.Metadata.CustomFields = make(map[string]interface{})
	}

	o.Metadata.CustomFields[key] = value
	return nil
}

// GetMetadata 获取元数据
func (o *OrganizationDTO) GetMetadata(key string) (interface{}, bool) {
	if o.Metadata.CustomFields == nil {
		return nil, false
	}

	val, exists := o.Metadata.CustomFields[key]
	return val, exists
}

// Activate 激活组织
func (o *OrganizationDTO) Activate() error {
	o.IsActive = true
	return nil
}

// Deactivate 禁用组织
func (o *OrganizationDTO) Deactivate() error {
	o.IsActive = false
	return nil
}

// GetDepth 获取组织深度
func (o *OrganizationDTO) GetDepth() int {
	return o.Level
}

// GetPathComponents 获取路径组件
func (o *OrganizationDTO) GetPathComponents() []string {
	if o.Path == "" {
		return []string{}
	}

	// 移除开头的斜杠并分割
	path := strings.TrimPrefix(o.Path, "/")
	if path == "" {
		return []string{}
	}

	return strings.Split(path, "/")
}

// IsAncestorOf 是否为指定组织的祖先
func (o *OrganizationDTO) IsAncestorOf(other *OrganizationDTO) bool {
	if other == nil || other.Path == "" || o.Path == "" {
		return false
	}

	// 检查other的路径是否以o的路径开头
	return strings.HasPrefix(other.Path, o.Path+"/")
}

// IsDescendantOf 是否为指定组织的后代
func (o *OrganizationDTO) IsDescendantOf(other *OrganizationDTO) bool {
	if other == nil || other.Path == "" || o.Path == "" {
		return false
	}

	// 检查o的路径是否以other的路径开头
	return strings.HasPrefix(o.Path, other.Path+"/")
}

// GetActiveUserCount 获取活跃用户数
func (o *OrganizationDTO) GetActiveUserCount() int {
	// 这里应该从repository查询实际数据
	// 临时返回UserCount，实际实现中应该根据IsActive字段过滤
	return o.UserCount
}

// HasUser 检查是否包含指定用户
func (o *OrganizationDTO) HasUser(userID string) bool {
	// 这里需要外部提供用户数据
	// 临时返回false，实际实现中应该查询数据库
	return false
}

// CanDelete 是否可以删除
func (o *OrganizationDTO) CanDelete() bool {
	// 如果有子组织或用户，则不能删除
	return o.ChildCount == 0 && o.UserCount == 0
}
