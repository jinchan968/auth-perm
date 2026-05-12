// PolymarketRepo Polymarket 预测市场数据仓库。
// Upsert 方法按 condition_id 去重：已存在则更新概率/交易量，不存在则新建。
// GetUnmatched 返回尚未匹配到投资主题的市场（theme_id 为空）。
package repo

import (
	"context"
	stdErr "errors"
	"time"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/newshock/dm"
)

type PolymarketRepo struct {
	db *gorm.DB
}

func NewPolymarketRepo(db *gorm.DB) *PolymarketRepo {
	return &PolymarketRepo{db: db}
}

// Upsert 按 condition_id 去重插入或更新 Polymarket 市场。
// 已存在：更新 probability、volume、outcome、last_synced_at；
// 不存在：新建记录。
func (r *PolymarketRepo) Upsert(ctx context.Context, market *dm.PolymarketMarket) error {
	var existing dm.PolymarketMarket
	err := r.db.WithContext(ctx).
		Where("condition_id = ?", market.ConditionID).
		First(&existing).Error

	if stdErr.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(market).Error
	}
	if err != nil {
		return errors.WrapBizError(err, "查询 Polymarket 市场失败")
	}

	existing.Probability = market.Probability
	existing.Volume = market.Volume
	existing.Outcome = market.Outcome
	existing.LastSyncedAt = market.LastSyncedAt
	if market.ThemeID != "" {
		existing.ThemeID = market.ThemeID
	}
	return r.db.WithContext(ctx).Save(&existing).Error
}

// GetByThemeID 获取某主题关联的所有 Polymarket 市场，按概率降序
func (r *PolymarketRepo) GetByThemeID(ctx context.Context, themeID string) ([]dm.PolymarketMarket, error) {
	markets := make([]dm.PolymarketMarket, 0)
	if err := r.db.WithContext(ctx).
		Where("theme_id = ?", themeID).
		Order("probability DESC").
		Find(&markets).Error; err != nil {
		return nil, errors.WrapBizError(err, "查询主题关联市场失败")
	}
	return markets, nil
}

func (r *PolymarketRepo) GetUnmatched(ctx context.Context, tenantID string) ([]dm.PolymarketMarket, error) {
	markets := make([]dm.PolymarketMarket, 0)
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND theme_id = ''", tenantID).
		Order("volume DESC").
		Find(&markets).Error; err != nil {
		return nil, errors.WrapBizError(err, "查询未匹配市场失败")
	}
	return markets, nil
}

// SyncBatch 批量同步 Polymarket 市场（逐条 upsert），设置 last_synced_at 为当前时间
func (r *PolymarketRepo) SyncBatch(ctx context.Context, markets []dm.PolymarketMarket) error {
	now := time.Now()
	for i := range markets {
		markets[i].LastSyncedAt = &now
		if err := r.Upsert(ctx, &markets[i]); err != nil {
			return err
		}
	}
	return nil
}

// ListByTenant returns all Polymarket markets for a tenant, ordered by volume
func (r *PolymarketRepo) ListByTenant(ctx context.Context, tenantID string, limit int) ([]dm.PolymarketMarket, error) {
	markets := make([]dm.PolymarketMarket, 0)
	q := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("volume DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&markets).Error; err != nil {
		return nil, errors.WrapBizError(err, "查询 Polymarket 市场列表失败")
	}
	return markets, nil
}

func (r *PolymarketRepo) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&dm.PolymarketMarket{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}
