package dm

import "time"

type WorkflowRunNodeDO struct {
	ID         string     `gorm:"primaryKey;type:varchar(36)"`
	RunID      string     `gorm:"column:run_id;type:varchar(36);not null"`
	NodeID     string     `gorm:"column:node_id;type:varchar(128);not null"`
	NodeType   string     `gorm:"column:node_type;type:varchar(32);not null"`
	NodeLabel  string     `gorm:"column:node_label;type:varchar(128)"`
	Status     string     `gorm:"column:status;type:varchar(16);not null;default:'pending'"`
	InputJSON  string     `gorm:"column:input_json;type:jsonb"`
	OutputJSON string     `gorm:"column:output_json;type:jsonb"`
	Error      string     `gorm:"column:error;type:text"`
	StartedAt  *time.Time `gorm:"column:started_at"`
	FinishedAt *time.Time `gorm:"column:finished_at"`
	DurationMs int        `gorm:"column:duration_ms"`
	RetryCount int        `gorm:"column:retry_count;default:0"`
}

func (WorkflowRunNodeDO) TableName() string {
	return "workflow_run_nodes"
}
