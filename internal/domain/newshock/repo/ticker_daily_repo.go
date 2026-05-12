// TickerDailyRepo 股票日线行情数据仓库。
// UpsertBatch 按 (ticker_id, trade_date) 批量去重写入。
// GetByTickerID 返回指定股票最近 N 天的日线数据，用于前端画图。
package repo

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/newshock/dm"
)

type TickerDailyRepo struct {
	db *gorm.DB
}

func NewTickerDailyRepo(db *gorm.DB) *TickerDailyRepo {
	return &TickerDailyRepo{db: db}
}

// UpsertBatch 批量写入日线数据，按 (ticker_id, trade_date) 冲突时更新全部字段。
func (r *TickerDailyRepo) UpsertBatch(ctx context.Context, records []dm.TickerDaily) error {
	if len(records) == 0 {
		return nil
	}
	// 分批 upsert，每批 100 条
	batchSize := 100
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]
		err := r.db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "ticker_id"}, {Name: "trade_date"}},
				DoUpdates: clause.AssignmentColumns([]string{"open", "high", "low", "close", "volume", "amount", "change_pct", "turnover", "updated_at"}),
			}).
			Create(&batch).Error
		if err != nil {
			return errors.WrapBizError(err, "批量写入日线数据失败")
		}
	}
	return nil
}

// GetByTickerID 获取指定股票最近 N 天的日线数据，按日期升序。
func (r *TickerDailyRepo) GetByTickerID(ctx context.Context, tickerID string, days int) ([]dm.TickerDaily, error) {
	if days <= 0 {
		days = 90
	}
	var records []dm.TickerDaily
	since := time.Now().AddDate(0, 0, -days)
	if err := r.db.WithContext(ctx).
		Where("ticker_id = ? AND trade_date >= ?", tickerID, since).
		Order("trade_date ASC").
		Find(&records).Error; err != nil {
		return nil, errors.WrapBizError(err, "查询日线数据失败")
	}
	return records, nil
}

// GetLatestByTickerID 获取指定股票最新的日线记录。
func (r *TickerDailyRepo) GetLatestByTickerID(ctx context.Context, tickerID string) (*dm.TickerDaily, error) {
	var record dm.TickerDaily
	if err := r.db.WithContext(ctx).
		Where("ticker_id = ?", tickerID).
		Order("trade_date DESC").
		First(&record).Error; err != nil {
		return nil, errors.WrapBizError(err, "查询最新日线数据失败")
	}
	return &record, nil
}
