// TickerRepo 股票数据仓库，提供股票标的的 CRUD、搜索和统计查询。
// 搜索支持按股票代码(symbol)和名称(name)模糊匹配。
package repo

import (
	"context"
	stdErr "errors"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/newshock/dm"
)

// TickerQueryParams 股票列表查询参数
type TickerQueryParams struct {
	TenantID string
	Market   string
	Keyword  string
	OrderBy  string
	Page     int
	PageSize int
}

type TickerRepo struct {
	db *gorm.DB
}

func NewTickerRepo(db *gorm.DB) *TickerRepo {
	return &TickerRepo{db: db}
}

// Create 创建新股票记录
func (r *TickerRepo) Create(ctx context.Context, ticker *dm.Ticker) error {
	if err := r.db.WithContext(ctx).Create(ticker).Error; err != nil {
		return errors.WrapBizError(err, "创建股票失败")
	}
	return nil
}

// FindByID 按 ID 查找股票，不存在返回 NotFoundError
func (r *TickerRepo) FindByID(ctx context.Context, id string) (*dm.Ticker, error) {
	var ticker dm.Ticker
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&ticker).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("股票不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找股票失败")
	}
	return &ticker, nil
}

// FindByIDAndTenantID 按 ID + 租户 ID 查找股票，确保租户隔离
func (r *TickerRepo) FindByIDAndTenantID(ctx context.Context, id, tenantID string) (*dm.Ticker, error) {
	var ticker dm.Ticker
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&ticker).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("股票不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找股票失败")
	}
	return &ticker, nil
}

// FindBySymbol 按股票代码（如 AAPL、600519）查找股票
func (r *TickerRepo) FindBySymbol(ctx context.Context, symbol, tenantID string) (*dm.Ticker, error) {
	var ticker dm.Ticker
	err := r.db.WithContext(ctx).Where("symbol = ? AND tenant_id = ?", symbol, tenantID).First(&ticker).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("股票不存在: %s", symbol)
		}
		return nil, errors.WrapBizError(err, "查找股票失败")
	}
	return &ticker, nil
}

// List 分页查询股票列表，支持按市场、关键词筛选，默认按 hot_score 降序
func (r *TickerRepo) List(ctx context.Context, params TickerQueryParams) ([]dm.Ticker, int64, error) {
	var tickers []dm.Ticker
	var total int64

	q := r.db.WithContext(ctx).Model(&dm.Ticker{})

	if params.TenantID != "" {
		q = q.Where("tenant_id = ?", params.TenantID)
	}
	if params.Market != "" {
		q = q.Where("market = ?", params.Market)
	}
	if params.Keyword != "" {
		pat := ilikePattern(params.Keyword)
		q = q.Where("symbol ILIKE ? OR name ILIKE ?", pat, pat)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计股票数量失败")
	}

	order := sanitizeOrderBy(params.OrderBy, tickerOrderAllowlist, "hot_score DESC")

	page, pageSize := params.Page, params.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	err := q.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&tickers).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询股票列表失败")
	}
	return tickers, total, nil
}

// Update 更新股票记录（全量更新）
func (r *TickerRepo) Update(ctx context.Context, ticker *dm.Ticker) error {
	if err := r.db.WithContext(ctx).Save(ticker).Error; err != nil {
		return errors.WrapBizError(err, "更新股票失败")
	}
	return nil
}

// Delete 按 ID 删除股票
func (r *TickerRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&dm.Ticker{}).Error; err != nil {
		return errors.WrapBizError(err, "删除股票失败")
	}
	return nil
}

// Search 按关键词搜索股票（模糊匹配 symbol 和 name），按 hot_score 降序
func (r *TickerRepo) Search(ctx context.Context, tenantID, keyword string, limit int) ([]dm.Ticker, error) {
	var tickers []dm.Ticker
	pat := ilikePattern(keyword)
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND (symbol ILIKE ? OR name ILIKE ?)", tenantID, pat, pat).
		Order("hot_score DESC").
		Limit(limit).
		Find(&tickers).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "搜索股票失败")
	}
	return tickers, nil
}

// GetTopTickers 获取热度最高的 top N 股票，供 Dashboard 首页展示
func (r *TickerRepo) GetTopTickers(ctx context.Context, tenantID string, limit int) ([]dm.Ticker, error) {
	var tickers []dm.Ticker
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("hot_score DESC").
		Limit(limit).
		Find(&tickers).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询热门股票失败")
	}
	return tickers, nil
}

// FindByIDs 批量按 ID 查询股票，用于关联加载
func (r *TickerRepo) FindByIDs(ctx context.Context, ids []string) ([]dm.Ticker, error) {
	var tickers []dm.Ticker
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&tickers).Error; err != nil {
		return nil, errors.WrapBizError(err, "批量查询股票失败")
	}
	return tickers, nil
}

// GetAllByTenant 获取租户下所有股票（不分页），用于批量评分和关键词匹配
func (r *TickerRepo) GetAllByTenant(ctx context.Context, tenantID string) ([]dm.Ticker, error) {
	var tickers []dm.Ticker
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&tickers).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询全部股票失败")
	}
	return tickers, nil
}

// IncrementMentionCount 原子递增股票的提及次数（mention_count），由 NewsProcessor 在事件提取时调用
func (r *TickerRepo) IncrementMentionCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&dm.Ticker{}).
		Where("id = ?", id).
		UpdateColumn("mention_count", gorm.Expr("mention_count + 1")).Error
}

// GetRecentHot returns tickers with recent activity, ordered by mention_count
func (r *TickerRepo) GetRecentHot(ctx context.Context, tenantID string, limit int) ([]dm.Ticker, error) {
	var tickers []dm.Ticker
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND mention_count > 0", tenantID).
		Order("mention_count DESC, updated_at DESC").
		Limit(limit).
		Find(&tickers).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询近期热门股票失败")
	}
	return tickers, nil
}

// CountByTenant 统计租户下的股票总数，供 StatsService 展示
func (r *TickerRepo) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&dm.Ticker{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}
