package repo

import (
	"context"
	stdErr "errors"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/journal/dm"
)

type AIPredictionRepo struct {
	db *gorm.DB
}

func NewAIPredictionRepo(db *gorm.DB) *AIPredictionRepo {
	return &AIPredictionRepo{db: db}
}

func (r *AIPredictionRepo) Create(ctx context.Context, prediction *dm.AIPredictionDO) error {
	if err := r.db.WithContext(ctx).Create(prediction).Error; err != nil {
		return errors.WrapBizError(err, "保存AI预测失败")
	}
	return nil
}

func (r *AIPredictionRepo) FindByID(ctx context.Context, id, accountID, tenantID string) (*dm.AIPredictionDO, error) {
	var p dm.AIPredictionDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		First(&p).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("AI预测记录不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查询AI预测失败")
	}
	return &p, nil
}

func (r *AIPredictionRepo) List(ctx context.Context, tenantID, accountID string, page, pageSize int) ([]*dm.AIPredictionDO, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&dm.AIPredictionDO{}).
		Where("account_id = ? AND tenant_id = ?", accountID, tenantID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计AI预测失败")
	}

	var list []*dm.AIPredictionDO
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询AI预测列表失败")
	}

	return list, total, nil
}
