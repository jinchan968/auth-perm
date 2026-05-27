package repo

import (
	"auth-perm/internal/domain/workflow/constant"
	"auth-perm/internal/domain/workflow/dm"
	"time"

	"gorm.io/gorm"
)

type WorkflowRunRepo struct {
	db *gorm.DB
}

func NewWorkflowRunRepo(db *gorm.DB) *WorkflowRunRepo {
	return &WorkflowRunRepo{db: db}
}

func (r *WorkflowRunRepo) Create(do *dm.WorkflowRunDO) error {
	return r.db.Create(do).Error
}

func (r *WorkflowRunRepo) GetByID(id string) (*dm.WorkflowRunDO, error) {
	var do dm.WorkflowRunDO
	err := r.db.Where("id = ?", id).First(&do).Error
	if err != nil {
		return nil, err
	}
	return &do, nil
}

func (r *WorkflowRunRepo) ListByWorkflow(workflowID string, offset, limit int) ([]*dm.WorkflowRunDO, int64, error) {
	var list []*dm.WorkflowRunDO
	var total int64
	query := r.db.Where("workflow_id = ?", workflowID)
	query.Model(&dm.WorkflowRunDO{}).Count(&total)
	err := query.Order("started_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *WorkflowRunRepo) Update(do *dm.WorkflowRunDO) error {
	return r.db.Save(do).Error
}

func (r *WorkflowRunRepo) CancelIfRunning(id string) (bool, error) {
	result := r.db.Model(&dm.WorkflowRunDO{}).
		Where("id = ? AND status = ?", id, constant.StatusRunning).
		Updates(map[string]interface{}{
			"status":      constant.StatusCancelled,
			"finished_at": time.Now(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
