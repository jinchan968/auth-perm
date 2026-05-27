package repo

import (
	"auth-perm/internal/domain/workflow/constant"
	"auth-perm/internal/domain/workflow/dm"
	"gorm.io/gorm"
)

type WorkflowRepo struct {
	db *gorm.DB
}

func NewWorkflowRepo(db *gorm.DB) *WorkflowRepo {
	return &WorkflowRepo{db: db}
}

func (r *WorkflowRepo) Create(do *dm.WorkflowDO) error {
	return r.db.Create(do).Error
}

func (r *WorkflowRepo) GetByID(id string) (*dm.WorkflowDO, error) {
	var do dm.WorkflowDO
	err := r.db.Where("id = ?", id).First(&do).Error
	if err != nil {
		return nil, err
	}
	return &do, nil
}

func (r *WorkflowRepo) List(tenantID, accountID string, offset, limit int) ([]*dm.WorkflowDO, int64, error) {
	var list []*dm.WorkflowDO
	var total int64
	query := r.db.Where("tenant_id = ? AND account_id = ?", tenantID, accountID)
	query.Model(&dm.WorkflowDO{}).Count(&total)
	err := query.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *WorkflowRepo) Update(do *dm.WorkflowDO) error {
	return r.db.Save(do).Error
}

func (r *WorkflowRepo) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&dm.WorkflowDO{}).Error
}

func (r *WorkflowRepo) ListTemplates(tenantID string) ([]*dm.WorkflowDO, error) {
	var list []*dm.WorkflowDO
	err := r.db.Where("tenant_id = ? AND status = ? AND template_id IS NULL", tenantID, constant.StatusTemplate).
		Order("created_at DESC").Find(&list).Error
	return list, err
}
