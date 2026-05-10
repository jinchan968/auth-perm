package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NewsRaw RSS 采集的原始新闻，是数据管线的入口。
// 处理流程：RSS拉取 → 按 content_hash 去重 → 存入此表 → NewsProcessor 提取事件 → 标记 processed=true
// 每条新闻可能匹配到多个股票(Ticker)，匹配成功后会创建 Event 并建立关联。
type NewsRaw struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Title       string         `gorm:"column:title;type:varchar(500);not null" json:"title"`
	Content     string         `gorm:"column:content;type:text" json:"content"`
	Source      string         `gorm:"column:source;type:varchar(100)" json:"source"`            // 来源标识，如 "reuters", "bloomberg"
	Channel     string         `gorm:"column:channel;type:varchar(50)" json:"channel"`           // 渠道分类，如 "global_macro"
	URL         string         `gorm:"column:url;type:varchar(1000)" json:"url"`                 // 原文链接
	PublishedAt *time.Time     `gorm:"column:published_at" json:"published_at"`                  // 原文发布时间
	ContentHash string         `gorm:"column:content_hash;type:varchar(64)" json:"content_hash"` // SHA256(title|link)，用于去重
	Processed   bool           `gorm:"column:processed;default:false" json:"processed"`          // 是否已被 NewsProcessor 处理
	TenantID    string         `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	CreatedAt   time.Time      `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (NewsRaw) TableName() string { return "newshock_news_raw" }

func (n *NewsRaw) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}
