package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Event 市场事件，由 RSS 新闻自动提取或手动创建。
// 每个事件关联到一个投资主题(ThemeID)和多个股票标的(通过 event_tickers 中间表)。
// importance 重要度 1-5，由 AI 评估或手动设置：
//
//	1=极低(日常琐事), 2=低(一般新闻), 3=中(值得关注), 4=高(重大事件), 5=极高(市场转折点)
type Event struct {
	ID         string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Title      string         `gorm:"column:title;type:varchar(500);not null" json:"title"`
	Summary    string         `gorm:"column:summary;type:text" json:"summary"`
	Channel    string         `gorm:"column:channel;type:varchar(100)" json:"channel"`        // 渠道：global_macro/industry_news/market_flow
	Importance int            `gorm:"column:importance;not null;default:3" json:"importance"` // 重要度 1-5
	ThemeID    string         `gorm:"column:theme_id;type:varchar(36)" json:"theme_id"`       // 所属主题 ID
	ThemeName  string         `gorm:"column:theme_name;type:varchar(255)" json:"theme_name"`  // 所属主题名称（冗余字段，便于查询）
	EventTime  *time.Time     `gorm:"column:event_time" json:"event_time"`                    // 事件实际发生时间
	TenantID   string         `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	CreatedAt  time.Time      `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Event) TableName() string { return "newshock_events" }

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

// EventTicker 事件-股票关联表（多对多中间表）。
// 表示某个事件影响了某只股票。
type EventTicker struct {
	EventID  string `gorm:"column:event_id;type:varchar(36);primaryKey"`
	TickerID string `gorm:"column:ticker_id;type:varchar(36);primaryKey"`
}

func (EventTicker) TableName() string { return "newshock_event_tickers" }
