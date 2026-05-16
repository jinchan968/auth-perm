package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Ticker 证券标的，包括股票、ETF、指数、可转债等。
// hot_score 由定时任务根据 mention_count（被新闻提及次数）计算：
// hot_score = log2(mention_count + 1) × 10，然后归一化到 0-100
type Ticker struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Symbol       string         `gorm:"column:symbol;type:varchar(20);not null" json:"symbol"`                             // 证券代码，secid 格式
	Name         string         `gorm:"column:name;type:varchar(255)" json:"name"`                                         // 证券名称
	Market       string         `gorm:"column:market;type:varchar(10);not null;default:us" json:"market"`                  // 市场：us/cn/hk/kr
	SecurityType string         `gorm:"column:security_type;type:varchar(20);not null;default:stock" json:"security_type"` // 品种：stock/etf/index/bond/warrant/preferred
	HotScore     float64        `gorm:"column:hot_score;default:0" json:"hot_score"`                                       // 热度评分（0-100）
	MentionCount int            `gorm:"column:mention_count;default:0" json:"mention_count"`                               // 被新闻提及次数
	TenantID     string         `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	CreatedAt    time.Time      `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Ticker) TableName() string { return "newshock_tickers" }

func (t *Ticker) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
