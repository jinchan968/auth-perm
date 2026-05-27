package dm

import "time"

type WorkflowRunDO struct {
	ID            string     `gorm:"primaryKey;type:varchar(36)"`
	WorkflowID    string     `gorm:"column:workflow_id;type:varchar(36);not null"`
	TenantID      string     `gorm:"column:tenant_id;type:varchar(64);not null;default:'default'"`
	AccountID     string     `gorm:"column:account_id;type:varchar(64);not null"`
	ExecutionMode string     `gorm:"column:execution_mode;type:varchar(8);not null;default:'sync'"`
	InputText     string     `gorm:"column:input_text;type:text"`
	InputJSON     string     `gorm:"column:input_json;type:jsonb"`
	ResultJSON    string     `gorm:"column:result_json;type:jsonb"`
	Status        string     `gorm:"column:status;type:varchar(16);not null;default:'pending'"`
	StartedAt     *time.Time `gorm:"column:started_at"`
	FinishedAt    *time.Time `gorm:"column:finished_at"`
	DurationMs    int        `gorm:"column:duration_ms"`
	Error         string     `gorm:"column:error;type:text"`
}

func (WorkflowRunDO) TableName() string {
	return "workflow_runs"
}
