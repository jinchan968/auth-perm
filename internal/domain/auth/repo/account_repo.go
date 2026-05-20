package repo

import (
	"context"
	stdErr "errors"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/model"
	"auth-perm/internal/common/utils"
	"auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
)

// 匹配纯数字（允许 + - 空格 括号），用于区分手机号和用户名
var phoneRe = regexp.MustCompile(`^[\d\s\+\-\(\)]+$`)

// AccountRepo 账户仓储
type AccountRepo struct {
	db *gorm.DB
}

// NewAccountRepo 创建账户仓储
func NewAccountRepo(db *gorm.DB) *AccountRepo {
	return &AccountRepo{
		db: db,
	}
}

// FindByID 根据ID查找账户
func (r *AccountRepo) FindByID(ctx context.Context, id string) (*dm.AccountDO, error) {
	var account dm.AccountDO
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&account).Error

	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("账户不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找账户失败")
	}

	return &account, nil
}

// FindByEmail 根据邮箱查找账户
func (r *AccountRepo) FindByEmail(ctx context.Context, email string) (*dm.AccountDO, error) {
	var account dm.AccountDO
	err := r.db.WithContext(ctx).
		Joins("JOIN users ON accounts.user_id = users.id").
		Where("users.email = ?", email).
		First(&account).Error

	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("邮箱账户不存在: %s", email)
		}
		return nil, errors.WrapBizError(err, "通过邮箱查找账户失败")
	}

	return &account, nil
}

// FindByPhone 根据手机号查找账户
func (r *AccountRepo) FindByPhone(ctx context.Context, phone string) (*dm.AccountDO, error) {
	var account dm.AccountDO
	err := r.db.WithContext(ctx).
		Joins("JOIN users ON accounts.user_id = users.id").
		Where("users.phone = ?", phone).
		First(&account).Error

	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("手机号账户不存在: %s", phone)
		}
		return nil, errors.WrapBizError(err, "通过手机号查找账户失败")
	}

	return &account, nil
}

// FindByIdentifier 根据标识符（邮箱、手机号或用户名）查找账户
func (r *AccountRepo) FindByIdentifier(ctx context.Context, identifier string) (*dm.AccountDO, error) {
	var account dm.AccountDO

	baseQuery := r.db.WithContext(ctx).Joins("JOIN users ON accounts.user_id = users.id")

	// 自动判断是邮箱、手机号还是用户名
	if strings.Contains(identifier, "@") {
		// 邮箱登录
		err := baseQuery.Where("users.email = ?", identifier).First(&account).Error
		if err != nil {
			if stdErr.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.NewNotFoundErrorF("账户不存在: %s", identifier)
			}
			return nil, errors.WrapBizError(err, "查找账户失败")
		}
	} else if phoneRe.MatchString(identifier) {
		// 手机号登录
		err := baseQuery.Where("users.phone = ?", identifier).First(&account).Error
		if err != nil {
			if stdErr.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.NewNotFoundErrorF("账户不存在: %s", identifier)
			}
			return nil, errors.WrapBizError(err, "查找账户失败")
		}
	} else {
		// 用户名登录
		err := baseQuery.Where("users.username = ?", identifier).First(&account).Error
		if err != nil {
			if stdErr.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.NewNotFoundErrorF("账户不存在: %s", identifier)
			}
			return nil, errors.WrapBizError(err, "查找账户失败")
		}
	}

	return &account, nil
}

// FindByUserID 根据用户ID查找账户列表
func (r *AccountRepo) FindByUserID(ctx context.Context, userID string) ([]*dm.AccountDO, error) {
	var accounts []*dm.AccountDO
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&accounts).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "通过用户ID查找账户列表失败")
	}

	return accounts, nil
}

// FetchWithTenantUser 根据用户ID查找账户列表
func (r *AccountRepo) FetchWithTenantUser(ctx context.Context, userID, tenantID string) (*dm.AccountDO, error) {
	var accounts dm.AccountDO
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("tenant_id = ?", tenantID).
		Take(&accounts).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("账户不存在")
		}
		return nil, errors.WrapBizError(err, "通过用户ID查找账户列表失败")
	}

	return &accounts, nil
}

// FindByOAuth 根据OAuth信息查找账户
func (r *AccountRepo) FindByOAuth(ctx context.Context, provider, oauthID string) (*dm.AccountDO, error) {
	var account dm.AccountDO
	err := r.db.WithContext(ctx).
		Where("oauth_provider = ? AND oauth_id = ?", provider, oauthID).
		First(&account).Error

	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("OAuth账户不存在: %s/%s", provider, oauthID)
		}
		return nil, errors.WrapBizError(err, "通过OAuth查找账户失败")
	}

	return &account, nil
}

// Save 保存账户
func (r *AccountRepo) Save(ctx context.Context, account *dm.AccountDO) error {
	if account.ID == "" {
		// 创建新账户 - 必须通过NewAccount工厂方法创建
		if account.CreatedAt.IsZero() {
			return errors.NewInternalError("账户必须通过NewAccount工厂方法创建")
		}
		return r.db.WithContext(ctx).Create(account).Error
	}

	// 更新现有账户
	return r.db.WithContext(ctx).Save(account).Error
}

// SaveWithTx 使用事务保存账户
func (r *AccountRepo) SaveWithTx(ctx context.Context, tx *gorm.DB, account *dm.AccountDO) error {
	if account.ID == "" {
		// 创建新账户 - 必须通过NewAccount工厂方法创建
		if account.CreatedAt.IsZero() {
			return errors.NewInternalError("账户必须通过NewAccount工厂方法创建")
		}
		return tx.WithContext(ctx).Create(account).Error
	}

	// 更新现有账户
	return tx.WithContext(ctx).Save(account).Error
}

// GetDB 获取数据库连接（用于事务）
func (r *AccountRepo) GetDB() *gorm.DB {
	return r.db
}

// Delete 删除账户
func (r *AccountRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&dm.AccountDO{}, "id = ?", id).Error
}

// FindByTenantID 根据租户ID查找账户
func (r *AccountRepo) FindByTenantID(ctx context.Context, tenantID string, pagination *model.Pagination) ([]*dm.AccountDO, error) {
	var accounts []*dm.AccountDO

	query := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID)

	query = r.applyPagination(query, pagination)

	err := query.Find(&accounts).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "通过租户查找账户列表失败")
	}

	return accounts, nil
}

// FindByStatus 根据状态查找账户
func (r *AccountRepo) FindByStatus(ctx context.Context, status constant.AccountStatus) ([]*dm.AccountDO, error) {
	var accounts []*dm.AccountDO
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Find(&accounts).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "通过状态查找账户列表失败")
	}

	return accounts, nil
}

// UpdateStatus 更新账户状态
func (r *AccountRepo) UpdateStatus(ctx context.Context, id string, status constant.AccountStatus) error {
	return r.db.WithContext(ctx).
		Model(&dm.AccountDO{}).
		Where("id = ?", id).
		Update("status", status).
		Update("updated_at", time.Now()).
		Error
}

// VerifyEmail 验证邮箱
func (r *AccountRepo) VerifyEmail(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&dm.AccountDO{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"email_verified":    true,
			"email_verified_at": now,
			"updated_at":        now,
		}).Error
}

// UpdateLastLogin 更新最后登录时间
func (r *AccountRepo) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&dm.AccountDO{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"updated_at":    now,
		}).Error
}

// SearchAccounts 搜索账户
func (r *AccountRepo) SearchAccounts(ctx context.Context, query *dto.AccountSearchQueryDTO) ([]*dm.AccountDO, error) {
	var accounts []*dm.AccountDO

	db := r.db.WithContext(ctx).Joins("JOIN users ON accounts.user_id = users.id")

	// 应用租户过滤
	if query.TenantID != "" {
		db = db.Where("accounts.tenant_id = ?", query.TenantID)
	}

	// 应用关键词搜索（搜索users表的email和phone）
	if query.Keyword != "" {
		pat := utils.ILIKEPattern(query.Keyword)
		db = db.Where("users.email LIKE ? OR users.phone LIKE ?", pat, pat)
	}

	// 应用状态过滤
	if query.Status != nil {
		db = db.Where("accounts.status = ?", *query.Status)
	}

	// 应用账户类型过滤
	if query.AccountType != nil {
		db = db.Where("accounts.account_type = ?", *query.AccountType)
	}

	// 应用时间范围过滤
	if query.CreatedAt != nil {
		db = db.Where("accounts.created_at BETWEEN ? AND ?", query.CreatedAt.Start, query.CreatedAt.End)
	}

	if query.UpdatedAt != nil {
		db = db.Where("accounts.updated_at BETWEEN ? AND ?", query.UpdatedAt.Start, query.UpdatedAt.End)
	}

	// 应用分页
	if query.Pagination != nil {
		db = db.Offset(query.Pagination.GetOffset()).Limit(query.Pagination.GetLimit())

		if query.Pagination.SortBy != "" {
			order := query.Pagination.SortBy
			if query.Pagination.SortDesc {
				order += " DESC"
			} else {
				order += " ASC"
			}
			db = db.Order(order)
		}
	}

	err := db.Find(&accounts).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "搜索账户失败")
	}

	return accounts, nil
}

// SearchAccountsWithCount 搜索账户（带总数）
func (r *AccountRepo) SearchAccountsWithCount(ctx context.Context, query *dto.AccountSearchQueryDTO) ([]*dm.AccountDO, int64, error) {
	var accounts []*dm.AccountDO
	var total int64

	// 构建基础查询
	baseQuery := r.db.WithContext(ctx).Model(&dm.AccountDO{}).Joins("JOIN users ON accounts.user_id = users.id")

	// 应用租户过滤
	if query.TenantID != "" {
		baseQuery = baseQuery.Where("accounts.tenant_id = ?", query.TenantID)
	}

	// 应用关键词搜索（搜索users表的email、phone和username）
	if query.Keyword != "" {
		pat := utils.ILIKEPattern(query.Keyword)
		baseQuery = baseQuery.Where("users.email LIKE ? OR users.phone LIKE ? OR users.username LIKE ?", pat, pat, pat)
	}

	// 应用状态过滤
	if query.Status != nil {
		baseQuery = baseQuery.Where("accounts.status = ?", *query.Status)
	}

	// 应用账户类型过滤
	if query.AccountType != nil {
		baseQuery = baseQuery.Where("accounts.account_type = ?", *query.AccountType)
	}

	// 应用时间范围过滤
	if query.CreatedAt != nil {
		baseQuery = baseQuery.Where("accounts.created_at BETWEEN ? AND ?", query.CreatedAt.Start, query.CreatedAt.End)
	}

	if query.UpdatedAt != nil {
		baseQuery = baseQuery.Where("accounts.updated_at BETWEEN ? AND ?", query.UpdatedAt.Start, query.UpdatedAt.End)
	}

	// 先获取总数
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计账户数量失败")
	}

	// 应用分页和排序
	db := baseQuery
	if query.Pagination != nil {
		db = db.Offset(query.Pagination.GetOffset()).Limit(query.Pagination.GetLimit())

		sortBy := query.Pagination.SortBy
		if sortBy == "" {
			sortBy = "accounts.created_at"
		}
		order := sortBy
		if query.Pagination.SortDesc {
			order += " DESC"
		} else {
			order += " ASC"
		}
		db = db.Order(order)
	}

	// 获取账户列表
	err := db.Find(&accounts).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "搜索账户失败")
	}

	return accounts, total, nil
}

// GetAccountStats 获取账户统计
func (r *AccountRepo) GetAccountStats(ctx context.Context, tenantID string) (*dto.AccountStatsDTO, error) {
	stats := &dto.AccountStatsDTO{
		AccountTypes: make(map[string]int64),
		ByProvider:   make(map[string]int64),
	}

	// 统计总账户数
	if err := r.db.WithContext(ctx).
		Model(&dm.AccountDO{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalAccounts).Error; err != nil {
		return nil, err
	}

	// 统计活跃账户数
	if err := r.db.WithContext(ctx).
		Model(&dm.AccountDO{}).
		Where("tenant_id = ? AND status = ?", tenantID, constant.AccountStatusActive).
		Count(&stats.ActiveAccounts).Error; err != nil {
		return nil, err
	}

	// 统计邮箱账户数（通过JOIN users表查询）
	if err := r.db.WithContext(ctx).
		Table("accounts").
		Joins("JOIN users ON accounts.user_id = users.id").
		Where("accounts.tenant_id = ? AND accounts.account_type = ? AND users.email IS NOT NULL", tenantID, constant.AccountTypeEmail).
		Count(&stats.EmailAccounts).Error; err != nil {
		return nil, err
	}

	// 统计OAuth账户数
	stats.OAuthAccounts = stats.TotalAccounts - stats.EmailAccounts

	// 统计已验证账户数（通过JOIN users表查询邮箱验证状态）
	if err := r.db.WithContext(ctx).
		Table("accounts").
		Joins("JOIN users ON accounts.user_id = users.id").
		Where("accounts.tenant_id = ? AND users.email IS NOT NULL AND users.email != ''", tenantID).
		Count(&stats.VerifiedAccounts).Error; err != nil {
		return nil, err
	}

	// 按账户类型统计
	var typeCounts []struct {
		AccountType string
		Count       int64
	}
	if err := r.db.WithContext(ctx).
		Model(&dm.AccountDO{}).
		Select("account_type, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("account_type").
		Scan(&typeCounts).Error; err != nil {
		return nil, err
	}

	for _, tc := range typeCounts {
		stats.AccountTypes[tc.AccountType] = tc.Count
	}

	// 按OAuth提供商统计
	var providerCounts []struct {
		Provider string
		Count    int64
	}
	if err := r.db.WithContext(ctx).
		Model(&dm.AccountDO{}).
		Select("oauth_provider, COUNT(*) as count").
		Where("tenant_id = ? AND oauth_provider IS NOT NULL AND oauth_provider != ''", tenantID).
		Group("oauth_provider").
		Scan(&providerCounts).Error; err != nil {
		return nil, err
	}

	for _, pc := range providerCounts {
		stats.ByProvider[pc.Provider] = pc.Count
	}

	// 统计最近登录数（最近7天）
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := r.db.WithContext(ctx).
		Model(&dm.AccountDO{}).
		Where("tenant_id = ? AND last_login_at >= ?", tenantID, sevenDaysAgo).
		Count(&stats.RecentLogins).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// ==================== 辅助方法 ====================

// applyPagination 应用分页和排序
func (r *AccountRepo) applyPagination(query *gorm.DB, pagination *model.Pagination) *gorm.DB {
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

// UpdateEmailVerificationStatus 更新邮箱验证状态
func (r *AccountRepo) UpdateEmailVerificationStatus(ctx context.Context, accountID string, verified bool) error {
	now := time.Now()
	updates := map[string]interface{}{
		"email_verified":    verified,
		"email_verified_at": &now,
		"updated_at":        now,
	}

	return r.db.WithContext(ctx).Model(&dm.AccountDO{}).Where("id = ?", accountID).Updates(updates).Error
}

// FindByResetToken 根据密码重置令牌查找账户
func (r *AccountRepo) FindByResetToken(ctx context.Context, tokenHash string) (*dm.AccountDO, error) {
	var account dm.AccountDO
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("reset_password_token = ? AND reset_password_expires > ?", tokenHash, now).
		First(&account).Error

	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundError("重置令牌无效或已过期")
		}
		return nil, errors.WrapBizError(err, "通过重置令牌查找账户失败")
	}

	return &account, nil
}

// UpdateResetPasswordToken 更新密码重置令牌
func (r *AccountRepo) UpdateResetPasswordToken(ctx context.Context, accountID string, tokenHash string, expiresAt time.Time) error {
	updates := map[string]interface{}{
		"reset_password_token":   tokenHash,
		"reset_password_expires": expiresAt,
		"updated_at":             time.Now(),
	}

	return r.db.WithContext(ctx).Model(&dm.AccountDO{}).Where("id = ?", accountID).Updates(updates).Error
}

// ClearResetPasswordToken 清除密码重置令牌
func (r *AccountRepo) ClearResetPasswordToken(ctx context.Context, accountID string) error {
	updates := map[string]interface{}{
		"reset_password_token":   nil,
		"reset_password_expires": nil,
		"updated_at":             time.Now(),
	}

	return r.db.WithContext(ctx).Model(&dm.AccountDO{}).Where("id = ?", accountID).Updates(updates).Error
}
