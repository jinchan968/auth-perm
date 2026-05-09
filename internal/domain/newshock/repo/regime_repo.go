// RegimeRepo 市场环境数据仓库，由 AI 定期写入，前端读取最新一条展示。
package repo

import (
	"context"
	stdErr "errors"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/newshock/dm"
)

type RegimeRepo struct {
	db *gorm.DB
}

func NewRegimeRepo(db *gorm.DB) *RegimeRepo {
	return &RegimeRepo{db: db}
}

// Create 创建新的市场环境记录（由 AI 生成）
func (r *RegimeRepo) Create(ctx context.Context, regime *dm.Regime) error {
	if err := r.db.WithContext(ctx).Create(regime).Error; err != nil {
		return errors.WrapBizError(err, "创建市场环境失败")
	}
	return nil
}

// FindByID 按 ID 查找市场环境记录
func (r *RegimeRepo) FindByID(ctx context.Context, id string) (*dm.Regime, error) {
	var regime dm.Regime
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&regime).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("市场环境不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找市场环境失败")
	}
	return &regime, nil
}

// GetLatest 获取租户最新的市场环境记录，供 Dashboard 展示。无记录返回 nil（非错误）。
func (r *RegimeRepo) GetLatest(ctx context.Context, tenantID string) (*dm.Regime, error) {
	var regime dm.Regime
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		First(&regime).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.WrapBizError(err, "查询最新市场环境失败")
	}
	return &regime, nil
}

// Update 更新市场环境记录（全量更新）
func (r *RegimeRepo) Update(ctx context.Context, regime *dm.Regime) error {
	if err := r.db.WithContext(ctx).Save(regime).Error; err != nil {
		return errors.WrapBizError(err, "更新市场环境失败")
	}
	return nil
}
