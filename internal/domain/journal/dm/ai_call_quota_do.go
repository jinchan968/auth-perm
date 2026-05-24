package dm

import "time"

type AICallQuotaDO struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	TenantID  string    `gorm:"column:tenant_id;type:varchar(64);not null;index"`
	AccountID string    `gorm:"column:account_id;type:varchar(64);not null"`
	ModelID   string    `gorm:"column:model_id;type:varchar(64);not null"`
	CallDate  time.Time `gorm:"column:call_date;type:date;not null"`
	CallCount int       `gorm:"column:call_count;type:integer;not null;default:0"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (*AICallQuotaDO) TableName() string { return "ai_call_quotas" }