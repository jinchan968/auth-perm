package repo

import (
	"context"
	stdErr "errors"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/journal/dm"
)

// TagRepo 标签仓储
type TagRepo struct {
	db *gorm.DB
}

func NewTagRepo(db *gorm.DB) *TagRepo {
	return &TagRepo{db: db}
}

// Create 创建标签
func (r *TagRepo) Create(ctx context.Context, tag *dm.TagDO) error {
	if err := r.db.WithContext(ctx).Create(tag).Error; err != nil {
		return errors.WrapBizError(err, "创建标签失败")
	}
	return nil
}

// Update 更新标签
func (r *TagRepo) Update(ctx context.Context, tag *dm.TagDO) error {
	if err := r.db.WithContext(ctx).Save(tag).Error; err != nil {
		return errors.WrapBizError(err, "更新标签失败")
	}
	return nil
}

// FindByIDAndAccount 按 ID + 账户查找（防越权）
func (r *TagRepo) FindByIDAndAccount(ctx context.Context, id, accountID, tenantID string) (*dm.TagDO, error) {
	var tag dm.TagDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		First(&tag).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("标签不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找标签失败")
	}
	return &tag, nil
}

// ListByAccount 查询账户下所有标签
func (r *TagRepo) ListByAccount(ctx context.Context, accountID, tenantID string) ([]*dm.TagDO, error) {
	var tags []*dm.TagDO
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND tenant_id = ?", accountID, tenantID).
		Order("created_at ASC").
		Find(&tags).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询标签列表失败")
	}
	return tags, nil
}

// SoftDelete 软删除标签
func (r *TagRepo) SoftDelete(ctx context.Context, id, accountID, tenantID string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除关联
		if err := tx.Where("tag_id = ?", id).Delete(&dm.DiaryTagDO{}).Error; err != nil {
			return errors.WrapBizError(err, "删除标签关联失败")
		}
		// 删除标签
		result := tx.Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
			Delete(&dm.TagDO{})
		if result.Error != nil {
			return errors.WrapBizError(result.Error, "删除标签失败")
		}
		if result.RowsAffected == 0 {
			return errors.NewNotFoundErrorF("标签不存在: %s", id)
		}
		return nil
	})
	return err
}

// FindByNameAndAccount 按名称 + 账户查找（用于去重）
func (r *TagRepo) FindByNameAndAccount(ctx context.Context, name, accountID, tenantID string) (*dm.TagDO, error) {
	var tag dm.TagDO
	err := r.db.WithContext(ctx).
		Where("name = ? AND account_id = ? AND tenant_id = ?", name, accountID, tenantID).
		First(&tag).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.WrapBizError(err, "查找标签失败")
	}
	return &tag, nil
}

// FindByIDs 按 ID 列表批量查询
func (r *TagRepo) FindByIDs(ctx context.Context, ids []string, accountID, tenantID string) ([]*dm.TagDO, error) {
	if len(ids) == 0 {
		return []*dm.TagDO{}, nil
	}
	var tags []*dm.TagDO
	err := r.db.WithContext(ctx).
		Where("id IN ? AND account_id = ? AND tenant_id = ?", ids, accountID, tenantID).
		Find(&tags).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "批量查询标签失败")
	}
	return tags, nil
}
