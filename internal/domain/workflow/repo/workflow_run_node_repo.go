package repo

import (
	"auth-perm/internal/domain/workflow/dm"
	"gorm.io/gorm"
)

type WorkflowRunNodeRepo struct {
	db *gorm.DB
}

func NewWorkflowRunNodeRepo(db *gorm.DB) *WorkflowRunNodeRepo {
	return &WorkflowRunNodeRepo{db: db}
}

func (r *WorkflowRunNodeRepo) Create(do *dm.WorkflowRunNodeDO) error {
	return r.db.Create(do).Error
}

func (r *WorkflowRunNodeRepo) Update(do *dm.WorkflowRunNodeDO) error {
	return r.db.Model(&dm.WorkflowRunNodeDO{}).
		Where("run_id = ? AND node_id = ?", do.RunID, do.NodeID).
		Select("status", "output_json", "error", "finished_at", "duration_ms").
		Updates(do).Error
}

func (r *WorkflowRunNodeRepo) ListByRun(runID string) ([]*dm.WorkflowRunNodeDO, error) {
	var list []*dm.WorkflowRunNodeDO
	err := r.db.Where("run_id = ?", runID).Order("started_at ASC").Find(&list).Error
	return list, err
}
