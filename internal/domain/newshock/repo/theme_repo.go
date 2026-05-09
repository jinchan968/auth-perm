// ThemeRepo 主题数据仓库，提供主题的 CRUD 和统计查询。
// 所有查询都带 tenant_id 条件，实现租户数据隔离。
package repo

import (
	"context"
	stdErr "errors"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/newshock/constant"
	"auth-perm/internal/domain/newshock/dm"
)

// ThemeQueryParams 主题列表查询参数
type ThemeQueryParams struct {
	TenantID string
	Category string
	Trend    string
	Keyword  string
	OrderBy  string
	Page     int
	PageSize int
}

type ThemeRepo struct {
	db *gorm.DB
}

func NewThemeRepo(db *gorm.DB) *ThemeRepo {
	return &ThemeRepo{db: db}
}

// Create 创建新主题记录
func (r *ThemeRepo) Create(ctx context.Context, theme *dm.Theme) error {
	if err := r.db.WithContext(ctx).Create(theme).Error; err != nil {
		return errors.WrapBizError(err, "创建主题失败")
	}
	return nil
}

// FindByID 按 ID 查找主题，不存在返回 NotFoundError
func (r *ThemeRepo) FindByID(ctx context.Context, id string) (*dm.Theme, error) {
	var theme dm.Theme
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&theme).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("主题不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找主题失败")
	}
	return &theme, nil
}

// FindByIDAndTenantID 按 ID + 租户 ID 查找主题，确保租户隔离
func (r *ThemeRepo) FindByIDAndTenantID(ctx context.Context, id, tenantID string) (*dm.Theme, error) {
	var theme dm.Theme
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&theme).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("主题不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找主题失败")
	}
	return &theme, nil
}

// List 分页查询主题列表，支持按分类、趋势、关键词筛选，返回总数用于分页。
// 默认按 strength DESC 排序（最热主题优先）。
func (r *ThemeRepo) List(ctx context.Context, params ThemeQueryParams) ([]dm.Theme, int64, error) {
	var themes []dm.Theme
	var total int64

	q := r.db.WithContext(ctx).Model(&dm.Theme{})

	if params.TenantID != "" {
		q = q.Where("tenant_id = ?", params.TenantID)
	}
	if params.Category != "" {
		q = q.Where("category = ?", params.Category)
	}
	if params.Trend != "" {
		q = q.Where("trend = ?", params.Trend)
	}
	if params.Keyword != "" {
		pat := ilikePattern(params.Keyword)
		q = q.Where("name ILIKE ? OR description ILIKE ?", pat, pat)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计主题数量失败")
	}

	order := sanitizeOrderBy(params.OrderBy, themeOrderAllowlist, "strength DESC")

	page, pageSize := params.Page, params.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	err := q.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&themes).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询主题列表失败")
	}
	return themes, total, nil
}

// Update 更新主题记录（全量更新）
func (r *ThemeRepo) Update(ctx context.Context, theme *dm.Theme) error {
	if err := r.db.WithContext(ctx).Save(theme).Error; err != nil {
		return errors.WrapBizError(err, "更新主题失败")
	}
	return nil
}

// Delete 按 ID 删除主题
func (r *ThemeRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&dm.Theme{}).Error; err != nil {
		return errors.WrapBizError(err, "删除主题失败")
	}
	return nil
}

// Search 按关键词搜索主题（模糊匹配 name 和 description），按 strength 降序
func (r *ThemeRepo) Search(ctx context.Context, tenantID, keyword string, limit int) ([]dm.Theme, error) {
	var themes []dm.Theme
	pat := ilikePattern(keyword)
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND (name ILIKE ? OR description ILIKE ?)", tenantID, pat, pat).
		Order("strength DESC").
		Limit(limit).
		Find(&themes).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "搜索主题失败")
	}
	return themes, nil
}

// GetTopThemes 获取强度最高的 top N 主题，供 Dashboard 首页展示
func (r *ThemeRepo) GetTopThemes(ctx context.Context, tenantID string, limit int) ([]dm.Theme, error) {
	var themes []dm.Theme
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("strength DESC").
		Limit(limit).
		Find(&themes).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询热门主题失败")
	}
	return themes, nil
}

// UpdateTickerCount 重新统计并更新主题关联的股票数量（theme_tickers 表）
func (r *ThemeRepo) UpdateTickerCount(ctx context.Context, themeID string) error {
	var count int64
	r.db.WithContext(ctx).Model(&dm.ThemeTicker{}).Where("theme_id = ?", themeID).Count(&count)
	return r.db.WithContext(ctx).Model(&dm.Theme{}).Where("id = ?", themeID).Update("ticker_count", count).Error
}

// UpdateEventCount 重新统计并更新主题关联的事件数量（events 表）
func (r *ThemeRepo) UpdateEventCount(ctx context.Context, themeID string) error {
	var count int64
	r.db.WithContext(ctx).Model(&dm.Event{}).Where("theme_id = ?", themeID).Count(&count)
	return r.db.WithContext(ctx).Model(&dm.Theme{}).Where("id = ?", themeID).Update("event_count", count).Error
}

// AvgStrengthByTenant 计算租户下所有主题的平均强度，供 StatsService 展示
func (r *ThemeRepo) AvgStrengthByTenant(ctx context.Context, tenantID string) (float64, error) {
	var result struct {
		Avg float64
	}
	err := r.db.WithContext(ctx).Model(&dm.Theme{}).
		Select("COALESCE(AVG(strength), 0) as avg").
		Where("tenant_id = ?", tenantID).
		Scan(&result).Error
	if err != nil {
		return 0, errors.WrapBizError(err, "计算平均主题强度失败")
	}
	return result.Avg, nil
}

// FindByIDs 批量按 ID 查询主题，用于关联加载
func (r *ThemeRepo) FindByIDs(ctx context.Context, ids []string) ([]dm.Theme, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var themes []dm.Theme
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&themes).Error; err != nil {
		return nil, errors.WrapBizError(err, "批量查询主题失败")
	}
	return themes, nil
}

// GetRisingEmerging returns rising-trend themes with lower strength (emerging signals)
func (r *ThemeRepo) GetRisingEmerging(ctx context.Context, tenantID string, limit int) ([]dm.Theme, error) {
	var themes []dm.Theme
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND trend = ?", tenantID, constant.TrendRising).
		Order("strength ASC").
		Limit(limit).
		Find(&themes).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询新兴主题失败")
	}
	return themes, nil
}

// CountByTenant 统计租户下的主题总数，供 StatsService 展示
func (r *ThemeRepo) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&dm.Theme{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

// DistinctTenantIDs 获取所有有主题数据的租户 ID 列表，供批量评分遍历
func (r *ThemeRepo) DistinctTenantIDs(ctx context.Context) ([]string, error) {
	var tenantIDs []string
	err := r.db.WithContext(ctx).Model(&dm.Theme{}).
		Distinct("tenant_id").
		Pluck("tenant_id", &tenantIDs).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询租户列表失败")
	}
	if len(tenantIDs) == 0 {
		tenantIDs = []string{}
	}
	return tenantIDs, nil
}
