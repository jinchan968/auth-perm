package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PolymarketMarket Polymarket 预测市场数据。
// 从 Polymarket API 拉取，通过关键词自动匹配到投资主题(ThemeID)。
// 匹配算法：子串匹配(主题名出现在标题中+5分) + 字符级双字匹配(每个双字+2分)，总分>=3则匹配成功。
type PolymarketMarket struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ConditionID  string         `gorm:"column:condition_id;type:varchar(100);uniqueIndex" json:"condition_id"` // Polymarket 条件 ID（唯一）
	Title        string         `gorm:"column:title;type:varchar(500);not null" json:"title"`
	Description  string         `gorm:"column:description;type:text" json:"description"`
	Outcome      string         `gorm:"column:outcome;type:varchar(50)" json:"outcome"`         // 结果选项，如 "Yes"/"No"
	Probability  float64        `gorm:"column:probability;default:0" json:"probability"`        // 当前概率 0.0-1.0
	Volume       float64        `gorm:"column:volume;default:0" json:"volume"`                  // 交易量（美元）
	ThemeID      string         `gorm:"column:theme_id;type:varchar(36);index" json:"theme_id"` // 匹配到的主题 ID
	TenantID     string         `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	LastSyncedAt *time.Time     `gorm:"column:last_synced_at" json:"last_synced_at"` // 最后同步时间
	CreatedAt    time.Time      `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (PolymarketMarket) TableName() string { return "newshock_polymarket" }

func (p *PolymarketMarket) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
