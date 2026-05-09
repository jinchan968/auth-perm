package repo

import (
	"context"
	stdErr "errors"
	"time"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/model"
	"auth-perm/internal/common/utils"
	"auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
)

// UserRepo 用户仓储
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓储
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

// FindByID 根据ID查找用户
func (r *UserRepo) FindByID(ctx context.Context, id string) (*dm.UserDO, error) {
	var user dm.UserDO
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error

	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("用户不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找用户失败")
	}

	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*dm.UserDO, error) {
	var user dm.UserDO
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("邮箱用户不存在: %s", email)
		}
		return nil, errors.WrapBizError(err, "通过邮箱查找用户失败")
	}

	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*dm.UserDO, error) {
	var user dm.UserDO
	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error

	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("用户不存在: %s", username)
		}
		return nil, errors.WrapBizError(err, "通过用户名查找用户失败")
	}

	return &user, nil
}

// Save 保存用户
func (r *UserRepo) Save(ctx context.Context, user *dm.UserDO) error {
	if user.ID == "" {
		return r.db.WithContext(ctx).Create(user).Error
	}
	return r.db.WithContext(ctx).Save(user).Error
}

// SaveWithTx 使用事务保存用户
func (r *UserRepo) SaveWithTx(ctx context.Context, tx *gorm.DB, user *dm.UserDO) error {
	if user.ID == "" {
		return tx.WithContext(ctx).Create(user).Error
	}
	return tx.WithContext(ctx).Save(user).Error
}

// GetDB 获取数据库连接（用于事务）
func (r *UserRepo) GetDB() *gorm.DB {
	return r.db
}

// Delete 删除用户
func (r *UserRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&dm.UserDO{}, "id = ?", id).Error
}

// FindByIDs 根据ID列表批量查找用户
func (r *UserRepo) FindByIDs(ctx context.Context, ids []string) ([]*dm.UserDO, error) {
	if len(ids) == 0 {
		return []*dm.UserDO{}, nil
	}

	var users []*dm.UserDO
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&users).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "批量查找用户失败")
	}

	return users, nil
}

// FindByTenantID 根据租户ID查找用户
func (r *UserRepo) FindByTenantID(ctx context.Context, tenantID string, pagination *model.Pagination) ([]*dm.UserDO, error) {
	var users []*dm.UserDO

	query := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID)

	query = r.applyPagination(query, pagination)

	err := query.Find(&users).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "通过租户查找用户列表失败")
	}

	return users, nil
}

// FindByAccountID 根据账户ID查找用户
func (r *UserRepo) FindByAccountID(ctx context.Context, accountID string) ([]*dm.UserDO, error) {
	var users []*dm.UserDO
	err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Find(&users).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "通过账户查找用户列表失败")
	}

	return users, nil
}

// CountByTenantID 统计租户的用户数量
func (r *UserRepo) CountByTenantID(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&dm.UserDO{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error

	if err != nil {
		return 0, errors.WrapBizError(err, "统计用户数量失败")
	}

	return count, nil
}

// FindActiveUsers 查找活跃用户
func (r *UserRepo) FindActiveUsers(ctx context.Context, tenantID string) ([]*dm.UserDO, error) {
	var users []*dm.UserDO
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ?", tenantID, constant.UserStatusActive).
		Find(&users).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "查找活跃用户失败")
	}

	return users, nil
}

// FindByStatus 根据状态查找用户
func (r *UserRepo) FindByStatus(ctx context.Context, status constant.UserStatus) ([]*dm.UserDO, error) {
	var users []*dm.UserDO
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Find(&users).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "通过状态查找用户列表失败")
	}

	return users, nil
}

// SearchUsers 搜索用户
func (r *UserRepo) SearchUsers(ctx context.Context, query *dto.UserSearchQueryDTO) ([]*dm.UserDO, error) {
	var users []*dm.UserDO

	db := r.db.WithContext(ctx)

	if query.TenantID != "" {
		db = db.Where("users.tenant_id = ?", query.TenantID)
	}

	if query.Keyword != "" {
		pat := utils.ILIKEPattern(query.Keyword)
		db = db.Where("username LIKE ? OR nickname LIKE ?", pat, pat)
	}

	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	if query.RoleCode != "" {
		db = db.Joins("JOIN user_roles ON users.id = user_roles.user_id").
			Joins("JOIN roles ON user_roles.role_id = roles.id").
			Where("roles.code = ?", query.RoleCode)
	}

	if query.OrgID != "" {
		db = db.Joins("JOIN user_org ON users.id = user_org.user_id").
			Where("user_org.organization_id = ?", query.OrgID)
	}

	if query.CreatedAt != nil {
		db = db.Where("users.created_at BETWEEN ? AND ?", query.CreatedAt.Start, query.CreatedAt.End)
	}

	if query.UpdatedAt != nil {
		db = db.Where("users.updated_at BETWEEN ? AND ?", query.UpdatedAt.Start, query.UpdatedAt.End)
	}

	if query.Pagination != nil {
		db = r.applyPagination(db, query.Pagination)
	}

	err := db.Find(&users).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "搜索用户失败")
	}

	return users, nil
}

// BatchCreate 批量创建用户
func (r *UserRepo) BatchCreate(ctx context.Context, users []*dm.UserDO) error {
	if len(users) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(users).Error
}

// BatchUpdate 批量更新用户
func (r *UserRepo) BatchUpdate(ctx context.Context, users []*dm.UserDO) error {
	if len(users) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, user := range users {
			if err := tx.Save(user).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchDelete 批量删除用户
func (r *UserRepo) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Delete(&dm.UserDO{}, "id IN ?", ids).Error
}

// GetUserStats 获取用户统计
func (r *UserRepo) GetUserStats(ctx context.Context, tenantID string) (*dto.UserStatsDTO, error) {
	stats := &dto.UserStatsDTO{
		UsersByStatus: make(map[string]int64),
		AccountTypes:  make(map[string]int64),
	}

	if err := r.db.WithContext(ctx).
		Model(&dm.UserDO{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalUsers).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Model(&dm.UserDO{}).
		Where("tenant_id = ? AND status = ?", tenantID, constant.UserStatusActive).
		Count(&stats.ActiveUsers).Error; err != nil {
		return nil, err
	}

	stats.InactiveUsers = stats.TotalUsers - stats.ActiveUsers

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if err := r.db.WithContext(ctx).
		Model(&dm.UserDO{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, thirtyDaysAgo).
		Count(&stats.NewUsers).Error; err != nil {
		return nil, err
	}

	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := r.db.WithContext(ctx).
		Model(&dm.UserDO{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}

	for _, sc := range statusCounts {
		stats.UsersByStatus[sc.Status] = sc.Count
	}

	var accountTypeCounts []struct {
		AccountType string
		Count       int64
	}
	if err := r.db.WithContext(ctx).
		Table("users").
		Select("accounts.account_type, COUNT(*) as count").
		Joins("JOIN accounts ON users.account_id = accounts.id").
		Where("users.tenant_id = ?", tenantID).
		Group("accounts.account_type").
		Scan(&accountTypeCounts).Error; err != nil {
		return nil, err
	}

	for _, atc := range accountTypeCounts {
		stats.AccountTypes[atc.AccountType] = atc.Count
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := r.db.WithContext(ctx).
		Table("users").
		Joins("JOIN accounts ON users.account_id = accounts.id").
		Where("users.tenant_id = ? AND accounts.last_login_at >= ?", tenantID, sevenDaysAgo).
		Count(&stats.RecentLogins).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetRecentUserCount 获取最近注册的用户数量
func (r *UserRepo) GetRecentUserCount(ctx context.Context, tenantID string, duration time.Duration) (int64, error) {
	var count int64
	since := time.Now().Add(-duration)

	err := r.db.WithContext(ctx).
		Model(&dm.UserDO{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Count(&count).Error

	if err != nil {
		return 0, errors.WrapBizError(err, "统计最近用户数量失败")
	}

	return count, nil
}

// ==================== 辅助方法 ====================

// applyPagination 应用分页和排序
func (r *UserRepo) applyPagination(query *gorm.DB, pagination *model.Pagination) *gorm.DB {
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
