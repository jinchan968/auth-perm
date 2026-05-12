package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TickerConcept A股个股题材概念/行业板块/地域板块。
// type 取值：concept（概念板块）、industry（行业板块）、region（地域板块）。
type TickerConcept struct {
	ID         string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TickerID   string    `gorm:"column:ticker_id;type:varchar(36);not null;uniqueIndex:idx_newshock_tc_ticker_name" json:"ticker_id"`
	Name       string    `gorm:"column:name;type:varchar(255);not null;uniqueIndex:idx_newshock_tc_ticker_name" json:"name"`
	Type       string    `gorm:"column:type;type:varchar(20);not null;default:concept" json:"type"`
	SourceCode string    `gorm:"column:source_code;type:varchar(50);default:''" json:"source_code"`
	TenantID   string    `gorm:"column:tenant_id;type:varchar(36);not null;uniqueIndex:idx_newshock_tc_ticker_name" json:"tenant_id"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

func (TickerConcept) TableName() string { return "newshock_ticker_concepts" }

func (t *TickerConcept) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
