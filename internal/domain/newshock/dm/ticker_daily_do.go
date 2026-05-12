package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TickerDaily 股票日线行情数据。
// 从东方财富 API 获取，按 (ticker_id, trade_date) 唯一约束去重。
// 用于绘制价格走势图和量能分析。
type TickerDaily struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TickerID  string    `gorm:"column:ticker_id;type:varchar(36);not null;index" json:"ticker_id"`
	TradeDate time.Time `gorm:"column:trade_date;type:date;not null" json:"trade_date"`
	Open      float64   `gorm:"column:open;not null;default:0" json:"open"`
	High      float64   `gorm:"column:high;not null;default:0" json:"high"`
	Low       float64   `gorm:"column:low;not null;default:0" json:"low"`
	Close     float64   `gorm:"column:close;not null;default:0" json:"close"`
	Volume    int64     `gorm:"column:volume;not null;default:0" json:"volume"`
	Amount    float64   `gorm:"column:amount;not null;default:0" json:"amount"`
	ChangePct float64   `gorm:"column:change_pct;not null;default:0" json:"change_pct"`
	Turnover  float64   `gorm:"column:turnover;not null;default:0" json:"turnover"`
	TenantID  string    `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

func (TickerDaily) TableName() string { return "newshock_ticker_daily" }

func (t *TickerDaily) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
