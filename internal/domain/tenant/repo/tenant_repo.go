package repo

import (
	"context"

	errStd "errors"

	"auth-perm/internal/common/model"
	"auth-perm/internal/common/utils"
	authDm "auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/tenant/constant"
	"auth-perm/internal/domain/tenant/dm"

	"gorm.io/gorm"
)

// TenantRepo 租户仓储
type TenantRepo struct {
	db *gorm.DB
}

// NewTenantRepo 创建租户仓储
func NewTenantRepo(db *gorm.DB) *TenantRepo {
	return &TenantRepo{db: db}
}

// Create 创建租户
func (r *TenantRepo) Create(ctx context.Context, tenant *dm.TenantDO) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

// Update 更新租户
func (r *TenantRepo) Update(ctx context.Context, tenant *dm.TenantDO) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

// Delete 删除租户（软删除）
func (r *TenantRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&dm.TenantDO{}, "id = ?", id).Error
}

// UpdateStatus 更新租户状态
func (r *TenantRepo) UpdateStatus(ctx context.Context, id string, status constant.TenantStatus) error {
	return r.db.WithContext(ctx).Model(&dm.TenantDO{}).Where("id = ?", id).Update("status", status).Error
}

// FindByID 根据ID查找租户
func (r *TenantRepo) FindByID(ctx context.Context, id string) (*dm.TenantDO, error) {
	var tenant dm.TenantDO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error
	if errStd.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// FindByCode 根据租户代码查找租户
func (r *TenantRepo) FindByCode(ctx context.Context, code string) (*dm.TenantDO, error) {
	var tenant dm.TenantDO
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&tenant).Error
	if errStd.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// FindMaxCodeByPrefix 根据前缀查找最大的code
// 例如 prefix="T" 返回 "T000025"
func (r *TenantRepo) FindMaxCodeByPrefix(ctx context.Context, prefix string) (string, error) {
	var tenant dm.TenantDO
	// 使用正则表达式匹配以 prefix 开头的 code，并按 code 降序排序
	err := r.db.WithContext(ctx).
		Where("code LIKE ?", prefix+"%").
		Order("code DESC").
		First(&tenant).Error
	if errStd.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return tenant.Code, nil
}

// FindByTenantID 查找租户（兼容旧API）
func (r *TenantRepo) FindByTenantID(ctx context.Context, tenantID string) (*dm.TenantDO, error) {
	return r.FindByID(ctx, tenantID)
}

// FindByTenantID 查找租户列表
func (r *TenantRepo) FindByTenantIDList(ctx context.Context, tenantID []string) ([]*dm.TenantDO, error) {
	var tenants []*dm.TenantDO
	err := r.db.WithContext(ctx).Where("id IN ?", tenantID).Find(&tenants).Error
	return tenants, err
}

// ListParams 列表查询参数
type ListParams struct {
	Keyword string
	Status  string
	*model.Pagination
}

// List 分页查询租户列表
func (r *TenantRepo) List(ctx context.Context, params *ListParams) ([]*dm.TenantDO, int64, error) {
	var tenants []*dm.TenantDO
	var total int64

	query := r.db.WithContext(ctx).Model(&dm.TenantDO{})

	// 关键词搜索
	if params.Keyword != "" {
		pat := utils.ILIKEPattern(params.Keyword)
		query = query.Where("name ILIKE ? OR code ILIKE ?", pat, pat)
	}

	// 状态过滤
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	query = query.Offset(params.GetOffset()).Limit(params.GetLimit())

	// 排序
	orderBy := "created_at DESC"
	if params.SortBy != "" {
		if params.SortDesc {
			orderBy = params.SortBy + " DESC"
		} else {
			orderBy = params.SortBy + " ASC"
		}
	}
	query = query.Order(orderBy)

	if err := query.Find(&tenants).Error; err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}

// CountAccountsByTenantID 统计租户下的账户数量
func (r *TenantRepo) CountAccountsByTenantID(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&authDm.AccountDO{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return count, err
}

// ExistsByID 检查租户是否存在
func (r *TenantRepo) ExistsByID(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&dm.TenantDO{}).
		Where("id = ?", id).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByCode 检查租户代码是否存在
func (r *TenantRepo) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&dm.TenantDO{}).
		Where("code = ?", code).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateSettings 更新租户设置
func (r *TenantRepo) UpdateSettings(ctx context.Context, tenantID string, settings interface{}) error {
	return r.db.WithContext(ctx).Model(&dm.TenantDO{}).
		Where("id = ?", tenantID).
		Update("settings", settings).Error
}
