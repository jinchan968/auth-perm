package repo

import (
	"context"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/journal/dm"
)

// TemplateRepo 模板仓储
type TemplateRepo struct {
	db *gorm.DB
}

// NewTemplateRepo 创建模板仓储
func NewTemplateRepo(db *gorm.DB) *TemplateRepo {
	return &TemplateRepo{db: db}
}

// Create 创建模板
func (r *TemplateRepo) Create(ctx context.Context, template *dm.JournalTemplateDO) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// Update 更新模板
func (r *TemplateRepo) Update(ctx context.Context, template *dm.JournalTemplateDO) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// Delete 删除模板
func (r *TemplateRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&dm.JournalTemplateDO{}).Error
}

// FindByID 根据ID查找
func (r *TemplateRepo) FindByID(ctx context.Context, id string) (*dm.JournalTemplateDO, error) {
	var template dm.JournalTemplateDO
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&template).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundErrorF("模板不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找模板失败")
	}
	return &template, nil
}

// ListByTenant 分页查询租户下的模板
func (r *TemplateRepo) ListByTenant(ctx context.Context, tenantID string, page, pageSize int, nameFilter string, tagFilter string) ([]*dm.JournalTemplateDO, int64, error) {
	var templates []*dm.JournalTemplateDO
	var total int64

	query := r.db.WithContext(ctx).Model(&dm.JournalTemplateDO{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	// 名称搜索
	if nameFilter != "" {
		query = query.Where("name ILIKE ?", "%"+nameFilter+"%")
	}

	// 标签过滤
	if tagFilter != "" {
		query = query.Where("? = ANY(tags)", tagFilter)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计模板数量失败")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "查询模板列表失败")
	}

	return templates, total, nil
}

// FindByAccountID 根据账户ID查找
func (r *TemplateRepo) FindByAccountID(ctx context.Context, tenantID, accountID string) ([]*dm.JournalTemplateDO, error) {
	var templates []*dm.JournalTemplateDO
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND account_id = ? AND deleted_at IS NULL", tenantID, accountID).
		Order("created_at DESC").
		Find(&templates).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查找用户模板失败")
	}
	return templates, nil
}
