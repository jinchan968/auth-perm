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

// Ticker 股票标的，如 AAPL、TSLA 等。
// hot_score 由定时任务根据 mention_count（被新闻提及次数）计算：
// hot_score = log2(mention_count + 1) × 10，然后归一化到 0-100
type Ticker struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Symbol       string         `gorm:"column:symbol;type:varchar(20);not null" json:"symbol"`            // 股票代码，如 AAPL
	Name         string         `gorm:"column:name;type:varchar(255)" json:"name"`                        // 公司名称
	Market       string         `gorm:"column:market;type:varchar(10);not null;default:us" json:"market"` // 市场：us/cn/hk/kr
	HotScore     float64        `gorm:"column:hot_score;default:0" json:"hot_score"`                      // 热度评分（0-100）
	MentionCount int            `gorm:"column:mention_count;default:0" json:"mention_count"`              // 被新闻提及次数
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

// ThemeTicker 主题-股票关联表（多对多中间表）。
// 表示某只股票属于某个投资主题。
type ThemeTicker struct {
	ThemeID  string `gorm:"column:theme_id;type:varchar(36);primaryKey"`
	TickerID string `gorm:"column:ticker_id;type:varchar(36);primaryKey"`
}

func (ThemeTicker) TableName() string { return "newshock_theme_tickers" }

// Regime 市场环境判断，由 AI 定期分析主题趋势和事件后生成。
// regime_type: risk_on(风险偏好) / risk_off(风险回避) / neutral(中性)
// confidence: AI 对判断的置信度 0.0-1.0
type Regime struct {
	ID         string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	RegimeType string         `gorm:"column:regime_type;type:varchar(20);not null;default:neutral" json:"regime_type"`
	Confidence float64        `gorm:"column:confidence;default:0.5" json:"confidence"`
	Summary    string         `gorm:"column:summary;type:text" json:"summary"`
	TenantID   string         `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	CreatedAt  time.Time      `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Regime) TableName() string { return "newshock_regime" }

func (r *Regime) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

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
