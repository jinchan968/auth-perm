// TickerF10Repo 股票基本面数据仓库。
// UpsertBatch 按 ticker_id 批量去重写入。
package repo

import (
	"context"
	stderrors "errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/newshock/dm"
)

type TickerF10Repo struct {
	db *gorm.DB
}

func NewTickerF10Repo(db *gorm.DB) *TickerF10Repo {
	return &TickerF10Repo{db: db}
}

// UpsertBatch 批量写入 F10 数据，按 ticker_id 冲突时更新全部字段。
func (r *TickerF10Repo) UpsertBatch(ctx context.Context, records []dm.TickerF10) error {
	if len(records) == 0 {
		return nil
	}
	batchSize := 100
	for i := 0; i < len(records); i += batchSize {
		end := min(i+batchSize, len(records))
		batch := records[i:end]
		err := r.db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "ticker_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"pe_ttm", "pe_static", "pb", "total_mcap", "float_mcap",
					"turnover_rate", "volume_ratio", "limit_up", "limit_down",
					"industry", "total_shares", "float_shares", "eps", "bvps", "roe",
					"source", "updated_at",
				}),
			}).
			Create(&batch).Error
		if err != nil {
			return errors.WrapBizError(err, "批量写入F10数据失败")
		}
	}
	return nil
}

// GetByTickerID 获取指定股票的 F10 数据。
// 若无数据返回 NOT_FOUND 类型错误，DB 异常返回 INTERNAL 类型错误。
func (r *TickerF10Repo) GetByTickerID(ctx context.Context, tickerID string) (*dm.TickerF10, error) {
	var record dm.TickerF10
	if err := r.db.WithContext(ctx).
		Where("ticker_id = ?", tickerID).
		First(&record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("F10数据不存在: %s", tickerID)
		}
		return nil, errors.WrapBizError(err, "查询F10数据失败")
	}
	return &record, nil
}
