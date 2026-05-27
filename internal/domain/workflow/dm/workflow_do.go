package dm

import "time"

type WorkflowDO struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)"`
	TenantID    string    `gorm:"column:tenant_id;type:varchar(64);not null;default:'default'"`
	AccountID   string    `gorm:"column:account_id;type:varchar(64);not null"`
	Name        string    `gorm:"column:name;type:varchar(128);not null"`
	Description string    `gorm:"column:description;type:text"`
	FlowJSON    string    `gorm:"column:flow_json;type:jsonb;not null"`
	TemplateID  *string   `gorm:"column:template_id;type:varchar(36)"`
	Status      string    `gorm:"column:status;type:varchar(16);not null;default:'draft'"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (WorkflowDO) TableName() string {
	return "workflows"
}
