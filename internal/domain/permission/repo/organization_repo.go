package repo

import (
	"context"
	"time"

	errStd "errors"
	"gorm.io/gorm"

	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/permission/dm"
)

// OrganizationRepo 组织仓储
type OrganizationRepo struct {
	db *gorm.DB
}

// NewOrganizationRepo 创建组织仓储
func NewOrganizationRepo(db *gorm.DB) *OrganizationRepo {
	return &OrganizationRepo{db: db}
}

// Create 创建组织
func (r *OrganizationRepo) Create(ctx context.Context, org *dm.OrganizationDO) error {
	return r.db.WithContext(ctx).Create(org).Error
}

// Update 更新组织
func (r *OrganizationRepo) Update(ctx context.Context, org *dm.OrganizationDO) error {
	return r.db.WithContext(ctx).Save(org).Error
}

// Delete 删除组织（软删除）
func (r *OrganizationRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&dm.OrganizationDO{}, "id = ?", id).Error
}

// FindByID 根据ID查找组织
func (r *OrganizationRepo) FindByID(ctx context.Context, id string) (*dm.OrganizationDO, error) {
	var org dm.OrganizationDO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&org).Error
	if errStd.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByCode 根据租户ID和编码查找组织
func (r *OrganizationRepo) FindByCode(ctx context.Context, tenantID, code string) (*dm.OrganizationDO, error) {
	var org dm.OrganizationDO
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND code = ?", tenantID, code).
		First(&org).Error
	if errStd.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByTenantID 根据租户ID查找组织列表
func (r *OrganizationRepo) FindByTenantID(ctx context.Context, tenantID string, pagination *model.Pagination) ([]*dm.OrganizationDO, int64, error) {
	var orgs []*dm.OrganizationDO
	var total int64

	query := r.db.WithContext(ctx).Model(&dm.OrganizationDO{}).Where("tenant_id = ?", tenantID)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if pagination != nil && pagination.Page > 0 {
		offset := pagination.GetOffset()
		query = query.Offset(offset).Limit(pagination.GetLimit())
	}

	// 排序
	orderBy := "sort_order ASC, created_at DESC"
	if pagination != nil && pagination.SortBy != "" {
		if pagination.SortDesc {
			orderBy = pagination.SortBy + " DESC"
		} else {
			orderBy = pagination.SortBy + " ASC"
		}
	}
	query = query.Order(orderBy)

	if err := query.Find(&orgs).Error; err != nil {
		return nil, 0, err
	}

	return orgs, total, nil
}

// FindByParentID 根据父组织ID查找子组织
func (r *OrganizationRepo) FindByParentID(ctx context.Context, parentID string) ([]*dm.OrganizationDO, error) {
	var orgs []*dm.OrganizationDO
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("sort_order ASC, name ASC").
		Find(&orgs).Error
	return orgs, err
}

// FindRootOrgs 查找顶级组织（无父组织）
func (r *OrganizationRepo) FindRootOrgs(ctx context.Context, tenantID string) ([]*dm.OrganizationDO, error) {
	var orgs []*dm.OrganizationDO
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND (parent_id IS NULL OR parent_id = '')", tenantID).
		Order("sort_order ASC, name ASC").
		Find(&orgs).Error
	return orgs, err
}

// FindByIDs 根据ID列表查找组织
func (r *OrganizationRepo) FindByIDs(ctx context.Context, ids []string) ([]*dm.OrganizationDO, error) {
	var orgs []*dm.OrganizationDO
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&orgs).Error
	return orgs, err
}

// CountUsersByOrgID 统计组织用户数量
func (r *OrganizationRepo) CountUsersByOrgID(ctx context.Context, orgID string) (int64, error) {
	var count int64
	err := r.db.Model(&dm.AccountOrgDO{}).
		Where("organization_id = ?", orgID).
		Count(&count).Error
	return count, err
}

// CountActiveUsersByOrgID 统计组织活跃用户数量
func (r *OrganizationRepo) CountActiveUsersByOrgID(ctx context.Context, orgID string, days int) (int64, error) {
	if days <= 0 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)

	var count int64
	err := r.db.Model(&dm.AccountOrgDO{}).
		Where("organization_id = ?", orgID).
		Where("account_id IN (SELECT DISTINCT account_id FROM sessions WHERE created_at >= ?)", since).
		Count(&count).Error
	return count, err
}

// ExistsAccountOrg 检查账户组织关联是否存在
func (r *OrganizationRepo) ExistsAccountOrg(ctx context.Context, accountID, orgID string) (bool, error) {
	var count int64
	err := r.db.Model(&dm.AccountOrgDO{}).
		Where("account_id = ? AND organization_id = ?", accountID, orgID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AssignAccountToOrg 将账户分配到组织
func (r *OrganizationRepo) AssignAccountToOrg(ctx context.Context, accountID, orgID, tenantID string) error {
	accountOrg := dm.NewAccountOrg(accountID, orgID, tenantID)
	return r.db.WithContext(ctx).Create(accountOrg).Error
}

// RemoveAccountFromOrg 从组织移除账户
func (r *OrganizationRepo) RemoveAccountFromOrg(ctx context.Context, accountID, orgID string) error {
	return r.db.WithContext(ctx).
		Where("account_id = ? AND organization_id = ?", accountID, orgID).
		Delete(&dm.AccountOrgDO{}).Error
}

// FindAccountOrgs 查找账户所属的组织列表
func (r *OrganizationRepo) FindAccountOrgs(ctx context.Context, accountID string) ([]*dm.OrganizationDO, error) {
	var orgs []*dm.OrganizationDO
	err := r.db.WithContext(ctx).
		Table("organizations").
		Joins("JOIN account_org ON organizations.id = account_org.organization_id").
		Where("account_org.account_id = ?", accountID).
		Find(&orgs).Error
	return orgs, err
}

// HasChildren 检查组织是否有子组织
func (r *OrganizationRepo) HasChildren(ctx context.Context, orgID string) (bool, error) {
	var count int64
	err := r.db.Model(&dm.OrganizationDO{}).
		Where("parent_id = ?", orgID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdatePath 更新组织路径（内部方法）
func (r *OrganizationRepo) UpdatePath(ctx context.Context, orgID, path string) error {
	return r.db.WithContext(ctx).
		Model(&dm.OrganizationDO{}).
		Where("id = ?", orgID).
		Update("path", path).Error
}
