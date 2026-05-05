package repo

import (
	"context"
	stdErr "errors"
	"time"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/journal/dm"
)

// JournalQueryParams 札记查询参数
type JournalQueryParams struct {
	TenantID  string
	AccountID string
	StartDate time.Time // 查询起始日期
	EndDate   time.Time // 查询结束日期
	Page      int
	PageSize  int
}

// JournalRepo 札记仓储
type JournalRepo struct {
	db *gorm.DB
}

func NewJournalRepo(db *gorm.DB) *JournalRepo {
	return &JournalRepo{db: db}
}

// Create 创建札记
func (r *JournalRepo) Create(ctx context.Context, entry *dm.JournalEntryDO) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return errors.WrapBizError(err, "创建札记失败")
	}
	return nil
}

// FindByIDAndAccount 按 ID + 账户查找（防越权），含关联
func (r *JournalRepo) FindByIDAndAccount(ctx context.Context, id, accountID, tenantID string) (*dm.JournalEntryDO, error) {
	var entry dm.JournalEntryDO
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("Corrections").
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		Where("parent_id IS NULL"). // 排除修正条目
		First(&entry).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("札记不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找札记失败")
	}
	return &entry, nil
}

// FindCorrectionByID 查找修正条目
func (r *JournalRepo) FindCorrectionByID(ctx context.Context, id, accountID, tenantID string) (*dm.JournalEntryDO, error) {
	var entry dm.JournalEntryDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ? AND parent_id IS NOT NULL", id, accountID, tenantID).
		First(&entry).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("修正条目不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找修正条目失败")
	}
	return &entry, nil
}

// ListByDateRange 按日期范围分页查询主条目列表
func (r *JournalRepo) ListByDateRange(ctx context.Context, p *JournalQueryParams) ([]*dm.JournalEntryDO, int64, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&dm.JournalEntryDO{}).
		Preload("Tags").
		Preload("Corrections").
		Where("account_id = ? AND tenant_id = ?", p.AccountID, p.TenantID).
		Where("parent_id IS NULL") // 只查主条目

	if !p.StartDate.IsZero() {
		q = q.Where("entry_date >= ?", p.StartDate)
	}
	if !p.EndDate.IsZero() {
		q = q.Where("entry_date <= ?", p.EndDate)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计札记失败")
	}

	var entries []*dm.JournalEntryDO
	offset := (p.Page - 1) * p.PageSize
	err := q.Order("entry_date DESC, created_at DESC").
		Offset(offset).Limit(p.PageSize).
		Find(&entries).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询札记列表失败")
	}

	return entries, total, nil
}

// SoftDelete 软删除札记主条目（级联软删除关联的修正条目和标签关联）
func (r *JournalRepo) SoftDelete(ctx context.Context, id, accountID, tenantID string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除修正条目（加 account_id 约束，防止越权）
		if err := tx.Where("parent_id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).Delete(&dm.JournalEntryDO{}).Error; err != nil {
			return errors.WrapBizError(err, "删除修正条目失败")
		}
		// 删除标签关联
		if err := tx.Where("diary_id = ?", id).Delete(&dm.DiaryTagDO{}).Error; err != nil {
			return errors.WrapBizError(err, "删除标签关联失败")
		}
		// 删除主条目
		result := tx.Where("id = ? AND account_id = ? AND tenant_id = ? AND parent_id IS NULL", id, accountID, tenantID).
			Delete(&dm.JournalEntryDO{})
		if result.Error != nil {
			return errors.WrapBizError(result.Error, "删除札记失败")
		}
		if result.RowsAffected == 0 {
			return errors.NewNotFoundErrorF("札记不存在: %s", id)
		}
		return nil
	})
	return err
}

// SoftDeleteCorrection 软删除修正条目
func (r *JournalRepo) SoftDeleteCorrection(ctx context.Context, id, accountID, tenantID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ? AND parent_id IS NOT NULL", id, accountID, tenantID).
		Delete(&dm.JournalEntryDO{})
	if result.Error != nil {
		return errors.WrapBizError(result.Error, "删除修正条目失败")
	}
	if result.RowsAffected == 0 {
		return errors.NewNotFoundErrorF("修正条目不存在: %s", id)
	}
	return nil
}

// ReplaceTags 替换札记的标签关联
func (r *JournalRepo) ReplaceTags(ctx context.Context, entryID string, tagIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除旧关联
		if err := tx.Where("diary_id = ?", entryID).Delete(&dm.DiaryTagDO{}).Error; err != nil {
			return errors.WrapBizError(err, "清除旧标签关联失败")
		}
		// 添加新关联
		for _, tagID := range tagIDs {
			dt := &dm.DiaryTagDO{
				DiaryID:   entryID,
				TagID:     tagID,
				CreatedAt: time.Now(),
			}
			if err := tx.Create(dt).Error; err != nil {
				return errors.WrapBizError(err, "创建标签关联失败")
			}
		}
		return nil
	})
}
