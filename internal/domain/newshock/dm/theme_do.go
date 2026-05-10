// Package dm 定义 Newshock 模块的数据模型（Domain Model）。
// Newshock 是一个事件驱动的主题投资雷达系统，核心流程为：
//
//	RSS采集 → 新闻去重 → 提取事件 → 评分计算 → 前端展示
//
// 主要实体关系：
//
//	Theme(投资主题) --N:N-- Ticker(股票标的)
//	Theme              --1:N-- Event(市场事件)
//	Event              --N:N-- Ticker
//	Regime(市场环境)   独立实体，由 AI 定期判断
//	NewsRaw(RSS原始新闻) 临时表，处理后标记为已处理
//	PolymarketMarket(预测市场) 通过关键词自动匹配到 Theme
package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Theme 投资主题，如"AI半导体"、"地缘政治"、"国防军工"等。
// 每个主题有强度评分(strength)和趋势(trend)，由定时任务自动计算。
// strength 计算公式：近7天事件数×2 + 近7天重要度总和 + 关联股票数×0.5
// trend 判定：近7天事件数 / 前7天事件数，>1.2 为上升，<0.8 为下降，其余为稳定
type Theme struct {
	ID                       string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name                     string         `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description              string         `gorm:"column:description;type:text" json:"description"`
	Category                 string         `gorm:"column:category;type:varchar(50);not null;default:exploratory" json:"category"`  // 分类：geopolitical, ai_semi, macro_monetary 等
	Strength                 float64        `gorm:"column:strength;default:0" json:"strength"`                                      // 绝对强度评分
	StrengthNorm             float64        `gorm:"column:strength_norm;default:0" json:"strength_norm"`                            // 归一化强度（0-100），相对于最强主题
	ClassificationConfidence float64        `gorm:"column:classification_confidence;default:0.85" json:"classification_confidence"` // AI 分类置信度
	TickerCount              int            `gorm:"column:ticker_count;default:0" json:"ticker_count"`                              // 关联股票数量
	EventCount               int            `gorm:"column:event_count;default:0" json:"event_count"`                                // 关联事件数量
	Trend                    string         `gorm:"column:trend;type:varchar(20);not null;default:stable" json:"trend"`             // 趋势：rising/stable/declining
	TenantID                 string         `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	CreatedAt                time.Time      `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt                time.Time      `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Theme) TableName() string { return "newshock_themes" }

func (t *Theme) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
