package repo

import (
	"context"
	errStd "errors"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/permission/dm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PermissionRepo 权限仓储
type PermissionRepo struct {
	db *gorm.DB
}

// NewPermissionRepo 创建权限仓储
func NewPermissionRepo(db *gorm.DB) *PermissionRepo {
	return &PermissionRepo{db: db}
}

// ==================== 角色相关方法 ====================

// FindRolesByAccountID 根据账户ID查找角色
func (r *PermissionRepo) FindRolesByAccountID(ctx context.Context, accountID string) ([]*dm.RoleDO, error) {
	var roles []*dm.RoleDO
	err := r.db.WithContext(ctx).Model(&dm.RoleDO{}).
		Joins("JOIN account_roles ON account_roles.role_id = roles.id").
		Where("account_roles.account_id = ?", accountID).
		Where("roles.deleted_at IS NULL").
		Find(&roles).Error
	return roles, err
}

// FindRoleByID 根据ID查找角色
func (r *PermissionRepo) FindRoleByID(ctx context.Context, id string) (*dm.RoleDO, error) {
	var role dm.RoleDO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if err != nil {
		if errStd.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("角色不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找角色失败")
	}
	return &role, nil
}

// FindRoleByCode 根据代码查找角色
func (r *PermissionRepo) FindRoleByCode(ctx context.Context, tenantID, code string) (*dm.RoleDO, error) {
	var role dm.RoleDO
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND code = ?", tenantID, code).First(&role).Error
	if err != nil {
		if errStd.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 角色不存在，可用于创建
		}
		return nil, errors.WrapBizError(err, "通过代码查找角色失败")
	}
	return &role, nil
}

// FindRolesByTenantID 根据租户ID查找角色
func (r *PermissionRepo) FindRolesByTenantID(ctx context.Context, tenantID string, pagination *model.Pagination) ([]*dm.RoleDO, error) {
	var roles []*dm.RoleDO

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	query = r.applyPagination(query, pagination)

	err := query.Find(&roles).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "通过租户查找角色失败")
	}

	return roles, nil
}

// FindRolesByOrgID 根据组织ID查找角色
func (r *PermissionRepo) FindRolesByOrgID(ctx context.Context, orgID string, pagination *model.Pagination) ([]*dm.RoleDO, error) {
	var roles []*dm.RoleDO

	query := r.db.WithContext(ctx).Where("org_id = ?", orgID)

	if pagination != nil {
		query = r.applyPagination(query, pagination)
	}

	err := query.Find(&roles).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "通过组织查找角色失败")
	}

	return roles, nil
}

// FindActiveRoles 查找活跃角色
func (r *PermissionRepo) FindActiveRoles(ctx context.Context, tenantID string) ([]*dm.RoleDO, error) {
	var roles []*dm.RoleDO
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Order("priority DESC, name ASC").
		Find(&roles).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "查找活跃角色失败")
	}

	return roles, nil
}

// SaveRole 保存角色
func (r *PermissionRepo) SaveRole(ctx context.Context, role *dm.RoleDO) error {
	if role.ID == "" {
		return r.db.WithContext(ctx).Create(role).Error
	}
	return r.db.WithContext(ctx).Save(role).Error
}

// DeleteRole 删除角色
func (r *PermissionRepo) DeleteRole(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&dm.RoleDO{}, "id = ?", id).Error
}

// AssignRoleToAccount 为账户分配角色
func (r *PermissionRepo) AssignRoleToAccount(ctx context.Context, accountID, roleID, tenantID string) error {
	ar := dm.NewAccountRole(accountID, roleID, tenantID)
	return r.createAccountRolesIgnoreDuplicates(r.db.WithContext(ctx), ar)
}

// RemoveRoleFromAccount 从账户移除角色
func (r *PermissionRepo) RemoveRoleFromAccount(ctx context.Context, accountID, roleID string) error {
	return r.db.WithContext(ctx).Where("account_id = ? AND role_id = ?", accountID, roleID).Delete(&dm.AccountRoleDO{}).Error
}

// GetAccountRoles 获取账户的角色ID列表
func (r *PermissionRepo) GetAccountRoles(ctx context.Context, accountID string) ([]string, error) {
	var roleIDs []string
	err := r.db.WithContext(ctx).
		Table("account_roles").
		Where("account_id = ?", accountID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

// SyncAccountRoles 同步账户角色（差量更新：对比已有和期望的角色列表，仅增删差异部分）
func (r *PermissionRepo) SyncAccountRoles(ctx context.Context, accountID string, roleIDs []string, tenantID string) error {
	existingIDs, err := r.GetAccountRoles(ctx, accountID)
	if err != nil {
		return errors.WrapBizError(err, "查询账户现有角色失败")
	}

	existingSet := make(map[string]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		existingSet[id] = struct{}{}
	}

	desiredSet := make(map[string]struct{}, len(roleIDs))
	for _, id := range roleIDs {
		desiredSet[id] = struct{}{}
	}

	var toAdd []string
	for _, id := range roleIDs {
		if _, exists := existingSet[id]; !exists {
			toAdd = append(toAdd, id)
		}
	}

	var toRemove []string
	for _, id := range existingIDs {
		if _, exists := desiredSet[id]; !exists {
			toRemove = append(toRemove, id)
		}
	}

	if len(toAdd) == 0 && len(toRemove) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(toRemove) > 0 {
			if err := tx.Where("account_id = ? AND role_id IN ?", accountID, toRemove).
				Delete(&dm.AccountRoleDO{}).Error; err != nil {
				return errors.WrapBizError(err, "删除账户角色关联失败")
			}
		}

		if len(toAdd) > 0 {
			accountRoles := make([]*dm.AccountRoleDO, 0, len(toAdd))
			for _, roleID := range toAdd {
				accountRoles = append(accountRoles, dm.NewAccountRole(accountID, roleID, tenantID))
			}

			if err := r.createAccountRolesIgnoreDuplicates(tx, accountRoles); err != nil {
				return errors.WrapBizError(err, "新增账户角色关联失败")
			}
		}

		return nil
	})
}

// FindAccountIDsByRoleIDs 根据角色ID列表查找关联的账户ID列表
func (r *PermissionRepo) FindAccountIDsByRoleIDs(ctx context.Context, roleIDs []string) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}
	var accountIDs []string
	err := r.db.WithContext(ctx).
		Table("account_roles").
		Where("role_id IN ?", roleIDs).
		Distinct("account_id").
		Pluck("account_id", &accountIDs).Error
	return accountIDs, err
}

// ==================== 权限相关方法 ====================

// FindPermissionsByRoleIDs 根据角色ID列表查找权限
func (r *PermissionRepo) FindPermissionsByRoleIDs(ctx context.Context, roleIDs []string) ([]*dm.PermissionDO, error) {
	var permissions []*dm.PermissionDO
	err := r.db.WithContext(ctx).Model(&dm.PermissionDO{}).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id IN ?", roleIDs).
		Where("permissions.deleted_at IS NULL").
		Find(&permissions).Error
	return permissions, err
}

// FindPermissionByID 根据ID查找权限
func (r *PermissionRepo) FindPermissionByID(ctx context.Context, id string) (*dm.PermissionDO, error) {
	var permission dm.PermissionDO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("权限不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找权限失败")
	}
	return &permission, nil
}

// FindPermissionByCode 根据代码查找权限
// 如果不存在返回 nil, nil（用于创建前检查）
func (r *PermissionRepo) FindPermissionByCode(ctx context.Context, tenantID, code string) (*dm.PermissionDO, error) {
	var permission dm.PermissionDO
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND code = ?", tenantID, code).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在时返回 nil，用于创建新权限前的检查
			return nil, nil
		}
		return nil, errors.WrapBizError(err, "通过代码查找权限失败")
	}
	return &permission, nil
}

// FindByID 根据ID查找权限（返回nil而非错误）
func (r *PermissionRepo) FindByID(ctx context.Context, id string) (*dm.PermissionDO, error) {
	var permission dm.PermissionDO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// FindByCodes 根据代码列表查找权限
func (r *PermissionRepo) FindByCodes(ctx context.Context, codes []string) ([]*dm.PermissionDO, error) {
	if len(codes) == 0 {
		return []*dm.PermissionDO{}, nil
	}
	var permissions []*dm.PermissionDO
	err := r.db.WithContext(ctx).Where("code IN ?", codes).Find(&permissions).Error
	return permissions, err
}

// FindByIDs 根据ID列表查找权限
func (r *PermissionRepo) FindByIDs(ctx context.Context, ids []string) ([]*dm.PermissionDO, error) {
	if len(ids) == 0 {
		return []*dm.PermissionDO{}, nil
	}
	var permissions []*dm.PermissionDO
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&permissions).Error
	return permissions, err
}

// FindPermissionsByTenantID 根据租户ID查找权限
func (r *PermissionRepo) FindPermissionsByTenantID(ctx context.Context, tenantID string, pagination *model.Pagination) ([]*dm.PermissionDO, error) {
	var permissions []*dm.PermissionDO

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if pagination != nil {
		query = r.applyPagination(query, pagination)
	}

	err := query.Find(&permissions).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "通过租户查找权限失败")
	}

	return permissions, nil
}

// FindActivePermissions 查找活跃权限
func (r *PermissionRepo) FindActivePermissions(ctx context.Context, tenantID string) ([]*dm.PermissionDO, error) {
	var permissions []*dm.PermissionDO
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Order("resource, action").
		Find(&permissions).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "查找活跃权限失败")
	}

	return permissions, nil
}

// FindPermissionsByResource 根据资源类型查找权限
func (r *PermissionRepo) FindPermissionsByResource(ctx context.Context, tenantID, resource string) ([]*dm.PermissionDO, error) {
	var permissions []*dm.PermissionDO
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND resource = ?", tenantID, resource).
		Find(&permissions).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "通过资源查找权限失败")
	}

	return permissions, nil
}

// SavePermission 保存权限
func (r *PermissionRepo) SavePermission(ctx context.Context, permission *dm.PermissionDO) error {
	if permission.ID == "" {
		return r.db.WithContext(ctx).Create(permission).Error
	}
	return r.db.WithContext(ctx).Save(permission).Error
}

// DeletePermission 删除权限
func (r *PermissionRepo) DeletePermission(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&dm.PermissionDO{}, "id = ?", id).Error
}

// AssignPermissionToRole 为角色分配权限
func (r *PermissionRepo) AssignPermissionToRole(ctx context.Context, roleID, permissionID, tenantID string) error {
	rp := dm.NewRolePermission(roleID, permissionID, tenantID)
	return r.createRolePermissionsIgnoreDuplicates(r.db.WithContext(ctx), rp)
}

// RemovePermissionFromRole 从角色移除权限
func (r *PermissionRepo) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	return r.db.WithContext(ctx).Where("role_id = ? AND permission_id = ?", roleID, permissionID).Delete(&dm.RolePermissionDO{}).Error
}

// SyncRolePermissions 同步角色权限（差量更新：对比已有和期望的权限列表，仅增删差异部分）
func (r *PermissionRepo) SyncRolePermissions(ctx context.Context, roleID string, permissionIDs []string, tenantID string) error {
	// 1. 查询当前角色已关联的权限ID
	existingIDs, err := r.GetRolePermissions(ctx, roleID)
	if err != nil {
		return errors.WrapBizError(err, "查询角色现有权限失败")
	}

	// 2. 计算差量
	existingSet := make(map[string]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		existingSet[id] = struct{}{}
	}

	desiredSet := make(map[string]struct{}, len(permissionIDs))
	for _, id := range permissionIDs {
		desiredSet[id] = struct{}{}
	}

	// 需要新增的：在期望列表中但不在现有列表中
	var toAdd []string
	for _, id := range permissionIDs {
		if _, exists := existingSet[id]; !exists {
			toAdd = append(toAdd, id)
		}
	}

	// 需要删除的：在现有列表中但不在期望列表中
	var toRemove []string
	for _, id := range existingIDs {
		if _, exists := desiredSet[id]; !exists {
			toRemove = append(toRemove, id)
		}
	}

	// 3. 无变化则直接返回
	if len(toAdd) == 0 && len(toRemove) == 0 {
		return nil
	}

	// 4. 在事务中执行增删
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除需要移除的关联
		if len(toRemove) > 0 {
			if err := tx.Where("role_id = ? AND permission_id IN ?", roleID, toRemove).
				Delete(&dm.RolePermissionDO{}).Error; err != nil {
				return errors.WrapBizError(err, "删除角色权限关联失败")
			}
		}

		// 批量插入需要新增的关联
		if len(toAdd) > 0 {
			rolePermissions := make([]*dm.RolePermissionDO, 0, len(toAdd))
			for _, permissionID := range toAdd {
				rolePermissions = append(rolePermissions, dm.NewRolePermission(roleID, permissionID, tenantID))
			}
			if err := r.createRolePermissionsIgnoreDuplicates(tx, rolePermissions); err != nil {
				return errors.WrapBizError(err, "新增角色权限关联失败")
			}
		}

		return nil
	})
}

// GetRolePermissions 获取角色的权限ID列表
func (r *PermissionRepo) GetRolePermissions(ctx context.Context, roleID string) ([]string, error) {
	var permissionIDs []string
	err := r.db.WithContext(ctx).
		Table("role_permissions").
		Where("role_id = ?", roleID).
		Pluck("permission_id", &permissionIDs).Error
	return permissionIDs, err
}

// ==================== 统计相关方法 ====================

// GetRoleStats 获取角色统计
func (r *PermissionRepo) GetRoleStats(ctx context.Context, tenantID string) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 总角色数
	var totalRoles int64
	err := r.db.WithContext(ctx).Model(&dm.RoleDO{}).Where("tenant_id = ?", tenantID).Count(&totalRoles).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "统计角色总数失败")
	}
	stats["total_roles"] = totalRoles

	// 活跃角色数
	var activeRoles int64
	err = r.db.WithContext(ctx).Model(&dm.RoleDO{}).Where("tenant_id = ? AND is_active = ?", tenantID, true).Count(&activeRoles).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "统计活跃角色失败")
	}
	stats["active_roles"] = activeRoles

	// 系统角色数
	var systemRoles int64
	err = r.db.WithContext(ctx).Model(&dm.RoleDO{}).Where("tenant_id = ? AND is_system = ?", tenantID, true).Count(&systemRoles).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "统计系统角色失败")
	}
	stats["system_roles"] = systemRoles

	return stats, nil
}

// GetPermissionStats 获取权限统计
func (r *PermissionRepo) GetPermissionStats(ctx context.Context, tenantID string) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 总权限数
	var totalPermissions int64
	err := r.db.WithContext(ctx).Model(&dm.PermissionDO{}).Where("tenant_id = ?", tenantID).Count(&totalPermissions).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "统计权限总数失败")
	}
	stats["total_permissions"] = totalPermissions

	// 活跃权限数
	var activePermissions int64
	err = r.db.WithContext(ctx).Model(&dm.PermissionDO{}).Where("tenant_id = ? AND is_active = ?", tenantID, true).Count(&activePermissions).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "统计活跃权限失败")
	}
	stats["active_permissions"] = activePermissions

	// 系统权限数
	var systemPermissions int64
	err = r.db.WithContext(ctx).Model(&dm.PermissionDO{}).Where("tenant_id = ? AND is_system = ?", tenantID, true).Count(&systemPermissions).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "统计系统权限失败")
	}
	stats["system_permissions"] = systemPermissions

	// 按资源统计
	var resourceStats []struct {
		Resource string
		Count    int64
	}
	err = r.db.WithContext(ctx).Model(&dm.PermissionDO{}).Select("resource, COUNT(*) as count").Where("tenant_id = ?", tenantID).Group("resource").Scan(&resourceStats).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "按资源统计权限失败")
	}
	for _, stat := range resourceStats {
		stats["resource_"+stat.Resource] = stat.Count
	}

	return stats, nil
}

// ==================== Code生成相关方法 ====================

// FindMaxPermissionCodeByPrefix 根据租户ID和前缀查找最大的权限code
// 例如 tenantID="xxx", prefix="P" 返回 "P000025"
func (r *PermissionRepo) FindMaxPermissionCodeByPrefix(ctx context.Context, tenantID, prefix string) (string, error) {
	var permission dm.PermissionDO
	// 使用 tenant_id 条件和 code 前缀匹配，利用联合索引
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND code LIKE ?", tenantID, prefix+"%").
		Order("code DESC").
		First(&permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return permission.Code, nil
}

// FindMaxRoleCodeByPrefix 根据租户ID和前缀查找最大的角色code
// 例如 tenantID="xxx", prefix="R" 返回 "R000025"
func (r *PermissionRepo) FindMaxRoleCodeByPrefix(ctx context.Context, tenantID, prefix string) (string, error) {
	var role dm.RoleDO
	// 使用 tenant_id 条件和 code 前缀匹配，利用联合索引
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND code LIKE ?", tenantID, prefix+"%").
		Order("code DESC").
		First(&role).Error
	if errStd.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return role.Code, nil
}

// ==================== 辅助方法 ====================

// applyPagination 应用分页和排序
func (r *PermissionRepo) applyPagination(query *gorm.DB, pagination *model.Pagination) *gorm.DB {
	if pagination == nil {
		return query
	}

	query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())

	if pagination.SortBy != "" {
		order := pagination.SortBy
		if pagination.SortDesc {
			order += " DESC"
		} else {
			order += " ASC"
		}
		query = query.Order(order)
	}

	return query
}

func (r *PermissionRepo) createAccountRolesIgnoreDuplicates(db *gorm.DB, value interface{}) error {
	return db.Clauses(accountRoleOnConflictClause()).Create(value).Error
}

func (r *PermissionRepo) createRolePermissionsIgnoreDuplicates(db *gorm.DB, value interface{}) error {
	return db.Clauses(rolePermissionOnConflictClause()).Create(value).Error
}

func accountRoleOnConflictClause() clause.OnConflict {
	return clause.OnConflict{
		Columns:   []clause.Column{{Name: "account_id"}, {Name: "role_id"}},
		DoNothing: true,
	}
}

func rolePermissionOnConflictClause() clause.OnConflict {
	return clause.OnConflict{
		Columns:   []clause.Column{{Name: "role_id"}, {Name: "permission_id"}},
		DoNothing: true,
	}
}
