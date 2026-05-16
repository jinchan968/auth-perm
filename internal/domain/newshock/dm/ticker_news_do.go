package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TickerNews 个股新闻。
// 来源：东方财富个股新闻 API，按 (ticker_id, url) 唯一约束去重。
type TickerNews struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TickerID    string    `gorm:"column:ticker_id;type:varchar(36);not null;uniqueIndex:idx_ticker_news_unique;index" json:"ticker_id"`
	Title       string    `gorm:"column:title;type:varchar(500);not null" json:"title"`
	Content     string    `gorm:"column:content;type:text" json:"content"`
	Source      string    `gorm:"column:source;type:varchar(128)" json:"source"`
	PublishTime time.Time `gorm:"column:publish_time;not null;index" json:"publish_time"`
	URL         string    `gorm:"column:url;type:varchar(500);not null;uniqueIndex:idx_ticker_news_unique" json:"url"`
	TenantID    string    `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

func (TickerNews) TableName() string { return "newshock_ticker_news" }

func (t *TickerNews) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
