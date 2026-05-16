// TickerNewsRepo 个股新闻数据仓库。
// UpsertBatch 按 (ticker_id, url) 批量去重写入。
package repo

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/newshock/dm"
)

type TickerNewsRepo struct {
	db *gorm.DB
}

func NewTickerNewsRepo(db *gorm.DB) *TickerNewsRepo {
	return &TickerNewsRepo{db: db}
}

// UpsertBatch 批量写入新闻数据，按 (ticker_id, url) 冲突时跳过。
func (r *TickerNewsRepo) UpsertBatch(ctx context.Context, records []dm.TickerNews) error {
	if len(records) == 0 {
		return nil
	}
	batchSize := 100
	for i := 0; i < len(records); i += batchSize {
		end := min(i+batchSize, len(records))
		batch := records[i:end]
		err := r.db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "ticker_id"}, {Name: "url"}},
				DoNothing: true,
			}).
			Create(&batch).Error
		if err != nil {
			return errors.WrapBizError(err, "批量写入新闻数据失败")
		}
	}
	return nil
}

// GetByTickerID 获取指定股票的新闻列表，按发布时间降序。
func (r *TickerNewsRepo) GetByTickerID(ctx context.Context, tickerID string, limit int) ([]dm.TickerNews, error) {
	if limit <= 0 {
		limit = 20
	}
	var records []dm.TickerNews
	if err := r.db.WithContext(ctx).
		Where("ticker_id = ?", tickerID).
		Order("publish_time DESC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, errors.WrapBizError(err, "查询新闻数据失败")
	}
	return records, nil
}
