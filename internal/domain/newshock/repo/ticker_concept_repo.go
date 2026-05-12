package repo

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/newshock/dm"
)

type TickerConceptRepo struct {
	db *gorm.DB
}

func NewTickerConceptRepo(db *gorm.DB) *TickerConceptRepo {
	return &TickerConceptRepo{db: db}
}

// UpsertBatch 批量插入概念，ticker_id+name+tenant_id 冲突时更新 updated_at。
func (r *TickerConceptRepo) UpsertBatch(ctx context.Context, concepts []dm.TickerConcept) (int64, error) {
	if len(concepts) == 0 {
		return 0, nil
	}
	const batchSize = 500
	var totalAffected int64
	for i := 0; i < len(concepts); i += batchSize {
		end := i + batchSize
		if end > len(concepts) {
			end = len(concepts)
		}
		batch := concepts[i:end]
		result := r.db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "ticker_id"}, {Name: "name"}, {Name: "tenant_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
			}).
			Create(&batch)
		if result.Error != nil {
			return totalAffected, errors.WrapBizError(result.Error, "批量写入概念失败")
		}
		totalAffected += result.RowsAffected
	}
	return totalAffected, nil
}

// GetByTickerID 获取指定股票的所有概念
func (r *TickerConceptRepo) GetByTickerID(ctx context.Context, tickerID string) ([]dm.TickerConcept, error) {
	var concepts []dm.TickerConcept
	err := r.db.WithContext(ctx).Where("ticker_id = ?", tickerID).Order("type, name").Find(&concepts).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询股票概念失败")
	}
	return concepts, nil
}

// GetByTickerIDGrouped 按类型分组获取股票概念
func (r *TickerConceptRepo) GetByTickerIDGrouped(ctx context.Context, tickerID string) (map[string][]dm.TickerConcept, error) {
	concepts, err := r.GetByTickerID(ctx, tickerID)
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]dm.TickerConcept)
	for _, c := range concepts {
		grouped[c.Type] = append(grouped[c.Type], c)
	}
	return grouped, nil
}

// DeleteByTickerID 删除指定股票的所有概念（用于全量刷新）
func (r *TickerConceptRepo) DeleteByTickerID(ctx context.Context, tickerID string) error {
	return r.db.WithContext(ctx).Where("ticker_id = ?", tickerID).Delete(&dm.TickerConcept{}).Error
}

// CountByTickerID 统计指定股票的概念数量
func (r *TickerConceptRepo) CountByTickerID(ctx context.Context, tickerID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&dm.TickerConcept{}).Where("ticker_id = ?", tickerID).Count(&count).Error
	return count, err
}
