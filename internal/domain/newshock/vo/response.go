// Package vo 定义 Newshock 模块的视图对象（View Object），即 API 响应结构体。
// 所有响应体的 JSON 字段使用 snake_case 命名。
package vo

import "time"

// ThemeResponse 主题响应，对应 Theme 模型的公开字段
type ThemeResponse struct {
	ID                       string    `json:"id"`
	Name                     string    `json:"name"`
	Description              string    `json:"description"`
	Category                 string    `json:"category"`                  // 分类：geopolitical, ai_semi, macro_monetary 等
	Strength                 float64   `json:"strength"`                  // 绝对强度评分
	StrengthNorm             float64   `json:"strength_norm"`             // 归一化强度（0-100）
	ClassificationConfidence float64   `json:"classification_confidence"` // AI 分类置信度
	TickerCount              int       `json:"ticker_count"`              // 关联股票数
	EventCount               int       `json:"event_count"`               // 关联事件数
	Trend                    string    `json:"trend"`                     // 趋势：rising/stable/declining
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// ThemeDetailResponse 主题详情，包含关联的股票、事件和 Polymarket 数据
type ThemeDetailResponse struct {
	ThemeResponse
	Tickers    []TickerResponse     `json:"tickers,omitempty"`    // 关联的股票列表
	Events     []EventResponse      `json:"events,omitempty"`     // 关联的事件列表（最近20条）
	Polymarket []PolymarketResponse `json:"polymarket,omitempty"` // 关联的 Polymarket 预测市场
}

// PolymarketResponse Polymarket 预测市场响应
type PolymarketResponse struct {
	ConditionID string    `json:"condition_id"` // Polymarket 条件 ID
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Outcome     string    `json:"outcome"`     // 结果选项，如 "Yes"/"No"
	Probability float64   `json:"probability"` // 当前概率 0.0-1.0
	Volume      float64   `json:"volume"`      // 交易量（美元）
	UpdatedAt   time.Time `json:"updated_at"`
}

// TickerResponse 股票响应
type TickerResponse struct {
	ID           string    `json:"id"`
	Symbol       string    `json:"symbol"`        // 股票代码，如 AAPL
	Name         string    `json:"name"`          // 公司名称
	Market       string    `json:"market"`        // 市场：us/cn/hk/kr
	HotScore     float64   `json:"hot_score"`     // 热度评分（0-100）
	MentionCount int       `json:"mention_count"` // 被提及次数
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TickerDailyResponse 日线行情响应
type TickerDailyResponse struct {
	Date      string  `json:"date"`       // 日期 YYYY-MM-DD
	Open      float64 `json:"open"`       // 开盘价
	High      float64 `json:"high"`       // 最高价
	Low       float64 `json:"low"`        // 最低价
	Close     float64 `json:"close"`      // 收盘价
	Volume    int64   `json:"volume"`     // 成交量（手）
	Amount    float64 `json:"amount"`     // 成交额（元）
	ChangePct float64 `json:"change_pct"` // 涨跌幅 %
	Turnover  float64 `json:"turnover"`   // 换手率 %
}

// TickerConceptResponse 股票概念/板块响应
type TickerConceptResponse struct {
	Name string `json:"name"` // 板块名称
	Type string `json:"type"` // concept/industry/region
}

// TickerDetailResponse 股票详情，包含关联的主题、事件和日线行情
type TickerDetailResponse struct {
	TickerResponse
	Themes   []ThemeResponse         `json:"themes,omitempty"`   // 所属的投资主题
	Events   []EventResponse         `json:"events,omitempty"`   // 相关事件（最近20条）
	Daily    []TickerDailyResponse   `json:"daily,omitempty"`    // 日线行情（最近90天）
	Concepts []TickerConceptResponse `json:"concepts,omitempty"` // 概念板块（A股）
}

// EventResponse 事件响应
type EventResponse struct {
	ID         string           `json:"id"`
	Title      string           `json:"title"`
	Summary    string           `json:"summary"`
	Channel    string           `json:"channel"`           // 渠道：global_macro/industry_news/market_flow
	Importance int              `json:"importance"`        // 重要度 1-5
	ThemeID    string           `json:"theme_id"`          // 所属主题 ID
	ThemeName  string           `json:"theme_name"`        // 所属主题名称
	EventTime  *time.Time       `json:"event_time"`        // 事件发生时间
	Tickers    []TickerResponse `json:"tickers,omitempty"` // 关联的股票
	CreatedAt  time.Time        `json:"created_at"`
}

// RegimeResponse 市场环境响应
type RegimeResponse struct {
	ID         string    `json:"id"`
	RegimeType string    `json:"regime_type"` // risk_on / risk_off / neutral
	Confidence float64   `json:"confidence"`  // AI 置信度 0.0-1.0
	Summary    string    `json:"summary"`     // 一句话总结
	CreatedAt  time.Time `json:"created_at"`
}

// StatsResponse 首页统计数据
type StatsResponse struct {
	ThemeCount  int64   `json:"theme_count"`  // 主题总数
	TickerCount int64   `json:"ticker_count"` // 股票总数
	EventCount  int64   `json:"event_count"`  // 事件总数
	AvgStrength float64 `json:"avg_strength"` // 平均主题强度
}

// FreshnessResponse 数据新鲜度，记录各类数据的最后更新时间
type FreshnessResponse struct {
	ThemesUpdated  time.Time `json:"themes_updated"`  // 主题数据最后更新时间
	EventsUpdated  time.Time `json:"events_updated"`  // 事件数据最后更新时间
	TickersUpdated time.Time `json:"tickers_updated"` // 股票数据最后更新时间
}

// HomeDataResponse 首页聚合数据，包含统计、环境、热门主题/股票、最新事件、数据新鲜度
type HomeDataResponse struct {
	Stats        StatsResponse     `json:"stats"`         // 统计数据
	Regime       *RegimeResponse   `json:"regime"`        // 当前市场环境
	TopThemes    []ThemeResponse   `json:"top_themes"`    // 热门主题 TOP 5
	TopTickers   []TickerResponse  `json:"top_tickers"`   // 热门股票 TOP 5
	RecentEvents []EventResponse   `json:"recent_events"` // 最新事件（最近10条）
	Freshness    FreshnessResponse `json:"freshness"`     // 数据新鲜度
}

// SearchResponse 全局搜索结果，同时返回匹配的主题、股票和事件
type SearchResponse struct {
	Themes  []ThemeResponse  `json:"themes"`
	Tickers []TickerResponse `json:"tickers"`
	Events  []EventResponse  `json:"events"`
}

// PagedResponse 通用分页响应，所有列表接口共用
type PagedResponse struct {
	Items    interface{} `json:"items"`     // 数据列表
	Total    int64       `json:"total"`     // 总记录数
	Page     int         `json:"page"`      // 当前页码
	PageSize int         `json:"page_size"` // 每页条数
}

// EdgeResponse 边缘信号（Edge Signals），用于发现新兴机会：
// - 上升趋势中强度较低的新兴主题
// - 近期被频繁提及的热门股票
// - 高重要度的最新事件
type EdgeResponse struct {
	RisingThemes []ThemeResponse  `json:"rising_themes"` // 新兴主题（上升趋势 + 强度低）
	HotTickers   []TickerResponse `json:"hot_tickers"`   // 热门股票（近期被提及多）
	RecentEvents []EventResponse  `json:"recent_events"` // 高重要度事件
}

// PipelineStatus 数据管线运行状态，展示各环节的数据量和最新运行时间
type PipelineStatus struct {
	NewsTotal       int64      `json:"news_total"`        // RSS 采集的新闻总数
	NewsUnprocessed int64      `json:"news_unprocessed"`  // 待处理的新闻数
	ThemeCount      int64      `json:"theme_count"`       // 主题总数
	TickerCount     int64      `json:"ticker_count"`      // 股票总数
	EventCount      int64      `json:"event_count"`       // 事件总数
	PolymarketCount int64      `json:"polymarket_count"`  // Polymarket 市场数
	LatestNewsTime  *time.Time `json:"latest_news_time"`  // 最新新闻时间
	LatestEventTime *time.Time `json:"latest_event_time"` // 最新事件时间
}
