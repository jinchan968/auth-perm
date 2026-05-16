package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TickerF10 股票基本面数据（F10）。
// 来源：腾讯财经（估值）+ 东财（财务指标），按 ticker_id 唯一约束更新。
type TickerF10 struct {
	ID           string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TickerID     string    `gorm:"column:ticker_id;type:varchar(36);not null;uniqueIndex" json:"ticker_id"`
	PeTtm        float64   `gorm:"column:pe_ttm;not null;default:0" json:"pe_ttm"`
	PeStatic     float64   `gorm:"column:pe_static;not null;default:0" json:"pe_static"`
	Pb           float64   `gorm:"column:pb;not null;default:0" json:"pb"`
	TotalMcap    float64   `gorm:"column:total_mcap;not null;default:0" json:"total_mcap"`       // 总市值(亿)
	FloatMcap    float64   `gorm:"column:float_mcap;not null;default:0" json:"float_mcap"`       // 流通市值(亿)
	TurnoverRate float64   `gorm:"column:turnover_rate;not null;default:0" json:"turnover_rate"` // 换手率(%)
	VolumeRatio  float64   `gorm:"column:volume_ratio;not null;default:0" json:"volume_ratio"`   // 量比
	LimitUp      float64   `gorm:"column:limit_up;not null;default:0" json:"limit_up"`           // 涨停价
	LimitDown    float64   `gorm:"column:limit_down;not null;default:0" json:"limit_down"`       // 跌停价
	Industry     string    `gorm:"column:industry;type:varchar(64)" json:"industry"`             // 行业
	TotalShares  float64   `gorm:"column:total_shares;not null;default:0" json:"total_shares"`   // 总股本(万)
	FloatShares  float64   `gorm:"column:float_shares;not null;default:0" json:"float_shares"`   // 流通股本(万)
	Eps          float64   `gorm:"column:eps;not null;default:0" json:"eps"`                     // 每股收益
	Bvps         float64   `gorm:"column:bvps;not null;default:0" json:"bvps"`                   // 每股净资产
	Roe          float64   `gorm:"column:roe;not null;default:0" json:"roe"`                     // ROE(%)
	Source       string    `gorm:"column:source;type:varchar(32)" json:"source"`                 // 数据源
	TenantID     string    `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

func (TickerF10) TableName() string { return "newshock_ticker_f10" }

func (t *TickerF10) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
