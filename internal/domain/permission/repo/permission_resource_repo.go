package repo

import (
	"context"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/permission/constant"
	"auth-perm/internal/domain/permission/dm"
)

// PermissionResourceRepo 权限资源仓储
type PermissionResourceRepo struct {
	db *gorm.DB
}

// NewPermissionResourceRepo 创建权限资源仓储
func NewPermissionResourceRepo(db *gorm.DB) *PermissionResourceRepo {
	return &PermissionResourceRepo{db: db}
}

// FindByPermissionID 根据权限ID查找资源
func (r *PermissionResourceRepo) FindByPermissionID(ctx context.Context, permissionID string) ([]*dm.PermissionResourceDO, error) {
	var resources []*dm.PermissionResourceDO
	err := r.db.WithContext(ctx).Where("permission_id = ?", permissionID).
		Where("deleted_at IS NULL").
		Find(&resources).Error
	return resources, err
}

// FindByResourceType 根据资源类型查找
func (r *PermissionResourceRepo) FindByResourceType(ctx context.Context, resourceType string) ([]*dm.PermissionResourceDO, error) {
	var resources []*dm.PermissionResourceDO
	err := r.db.WithContext(ctx).Where("resource_type = ?", resourceType).
		Where("deleted_at IS NULL").
		Find(&resources).Error
	return resources, err
}

// FindByResourceIDAndType 根据资源ID和类型查找
func (r *PermissionResourceRepo) FindByResourceIDAndType(ctx context.Context, resourceID, resourceType string) (*dm.PermissionResourceDO, error) {
	var resource dm.PermissionResourceDO
	err := r.db.WithContext(ctx).Where("resource_id = ? AND resource_type = ?", resourceID, resourceType).
		Where("deleted_at IS NULL").
		First(&resource).Error
	return &resource, err
}

// FindResources 统一的资源查找方法
// resourceType: 资源类型（必填），如 "api_path", "menu", "button", "field"
// resourceID: 资源ID（可选）
// useWildcard: 是否使用 LIKE 模糊匹配（true 时将 resourceID 中的 % 作为 SQL 的 %）
func (r *PermissionResourceRepo) FindResources(ctx context.Context, resourceType, resourceID string, useWildcard bool) ([]*dm.PermissionResourceDO, error) {
	var resources []*dm.PermissionResourceDO

	query := r.db.WithContext(ctx).Where("resource_type = ?", resourceType).
		Where("deleted_at IS NULL")

	if resourceID != "" {
		if useWildcard {
			query = query.Where("resource_id LIKE ?", resourceID)
		} else {
			query = query.Where("resource_id = ?", resourceID)
		}
	}

	err := query.Find(&resources).Error
	return resources, err
}

// FindByPermissionAndResource 根据权限和资源查找
func (r *PermissionResourceRepo) FindByPermissionAndResource(ctx context.Context, permissionID, resourceID, resourceType string) (*dm.PermissionResourceDO, error) {
	var resource dm.PermissionResourceDO
	err := r.db.WithContext(ctx).Where("permission_id = ? AND resource_id = ? AND resource_type = ?", permissionID, resourceID, resourceType).
		Where("deleted_at IS NULL").
		First(&resource).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &resource, nil
}

// Create 创建权限资源关联
func (r *PermissionResourceRepo) Create(ctx context.Context, pr *dm.PermissionResourceDO) error {
	return r.db.WithContext(ctx).Create(pr).Error
}

// CreateBatch 批量创建
func (r *PermissionResourceRepo) CreateBatch(ctx context.Context, resources []*dm.PermissionResourceDO) error {
	if len(resources) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(resources, 100).Error
}

// Update 更新权限资源关联
func (r *PermissionResourceRepo) Update(ctx context.Context, pr *dm.PermissionResourceDO) error {
	return r.db.WithContext(ctx).Save(pr).Error
}

// DeleteByPermissionID 根据权限ID删除（软删除）
func (r *PermissionResourceRepo) DeleteByPermissionID(ctx context.Context, permissionID string) error {
	return r.db.WithContext(ctx).Where("permission_id = ?", permissionID).Delete(&dm.PermissionResourceDO{}).Error
}

// DeleteByResource 根据资源ID和类型删除（软删除）
func (r *PermissionResourceRepo) DeleteByResource(ctx context.Context, resourceID, resourceType string) error {
	return r.db.WithContext(ctx).Where("resource_id = ? AND resource_type = ?", resourceID, resourceType).Delete(&dm.PermissionResourceDO{}).Error
}

// GetPermissionResourceMap 获取权限资源映射
func (r *PermissionResourceRepo) GetPermissionResourceMap(ctx context.Context, permissionIDs []string) (map[string][]*dm.PermissionResourceDO, error) {
	var resources []*dm.PermissionResourceDO
	if len(permissionIDs) == 0 {
		return map[string][]*dm.PermissionResourceDO{}, nil
	}

	err := r.db.WithContext(ctx).Where("permission_id IN ?", permissionIDs).
		Where("deleted_at IS NULL").
		Find(&resources).Error
	if err != nil {
		return nil, err
	}

	// 按权限ID分组
	result := make(map[string][]*dm.PermissionResourceDO)
	for _, resource := range resources {
		result[resource.PermissionID] = append(result[resource.PermissionID], resource)
	}

	return result, nil
}

// GetResourcesByAPIPath 根据API路径查找权限资源（支持模糊匹配）
// 使用 useWildcard 参数控制是否使用 LIKE 模糊匹配
func (r *PermissionResourceRepo) GetResourcesByAPIPath(ctx context.Context, apiPath string, useWildcard bool) ([]*dm.PermissionResourceDO, error) {
	return r.FindResources(ctx, constant.ResourceTypeAPIPath, apiPath, useWildcard)
}

// GetResourcesByMenu 根据菜单查找权限资源
func (r *PermissionResourceRepo) GetResourcesByMenu(ctx context.Context, menuID string) ([]*dm.PermissionResourceDO, error) {
	return r.FindResources(ctx, constant.ResourceTypeMenu, menuID, false)
}

// DeleteByID 根据ID删除（软删除）
func (r *PermissionResourceRepo) DeleteByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&dm.PermissionResourceDO{}).Error
}

// FindByID 根据ID查找
func (r *PermissionResourceRepo) FindByID(ctx context.Context, id string) (*dm.PermissionResourceDO, error) {
	var resource dm.PermissionResourceDO
	err := r.db.WithContext(ctx).Where("id = ?", id).
		Where("deleted_at IS NULL").
		First(&resource).Error
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// CountByPermissionID 统计权限关联的资源数量
func (r *PermissionResourceRepo) CountByPermissionID(ctx context.Context, permissionID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&dm.PermissionResourceDO{}).
		Where("permission_id = ?", permissionID).
		Where("deleted_at IS NULL").
		Count(&count).Error
	return count, err
}

// FindAllResources 获取所有权限资源（用于超管）
func (r *PermissionResourceRepo) FindAllResources(ctx context.Context) ([]*dm.PermissionResourceDO, error) {
	var resources []*dm.PermissionResourceDO
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Find(&resources).Error
	return resources, err
}
