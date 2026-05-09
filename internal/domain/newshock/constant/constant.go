// constant 包定义 newshock 模块使用的所有枚举常量。
// 包括：主题分类、趋势类型、市场类型、新闻渠道、市场环境类型。
// 这些常量用于数据库存储和 API 交互，避免魔法字符串。
package constant

// 主题分类（Theme Category）—— 标识投资主题所属的宏观领域。
// 存储在 themes.category 字段，用于前端筛选和分组展示。
const (
	CategoryGeopolitical  = "geopolitical"   // 地缘政治：战争、制裁、外交冲突
	CategoryAISemi        = "ai_semi"        // AI 与半导体：芯片、大模型、算力
	CategoryMacroMonetary = "macro_monetary" // 宏观货币政策：利率、通胀、央行决议
	CategorySupplyChain   = "supply_chain"   // 供应链：物流、港口、原材料
	CategoryDefense       = "defense"        // 国防军工：军费、武器订单、防务合作
	CategoryEnergy        = "energy"         // 能源：石油、天然气、新能源
	CategoryEarningsEvent = "earnings_event" // 财报事件：季度财报、业绩预警
	CategoryExploratory   = "exploratory"    // 探索性主题：尚未确认的新兴方向
	CategoryPharma        = "pharma"         // 医药：新药审批、临床试验、FDA
	CategoryRegulatory    = "regulatory"     // 监管政策：反垄断、数据安全、行业监管
)

// 趋势类型（Trend）—— 标识主题的热度变化方向。
// 由 ScoringService 根据近期事件数量和重要度动态计算。
const (
	TrendRising    = "rising"    // 上升：近期事件增多或重要度提升
	TrendStable    = "stable"    // 稳定：事件频率和重要度无明显变化
	TrendDeclining = "declining" // 下降：近期事件减少或重要度降低
)

// 市场类型（Market）—— 标识股票所属的交易市场。
// 存储在 tickers.market 字段，用于按市场筛选。
const (
	MarketUS = "us" // 美股
	MarketCN = "cn" // A 股
	MarketHK = "hk" // 港股
	MarketKR = "kr" // 韩股
)

// 新闻渠道（Channel）—— 标识 RSS 新闻的来源分类。
// 对应 config.yaml 中 RSS feed 的 channel 配置字段。
const (
	ChannelGlobalMacro  = "global_macro"  // 全球宏观：Reuters、Bloomberg 等综合财经
	ChannelIndustryNews = "industry_news" // 行业新闻：特定行业垂直媒体
	ChannelMarketFlow   = "market_flow"   // 资金流向：期权异动、资金流数据
)

// 市场环境类型（Regime）—— AI 判断的当前市场整体状态。
// 由 AIService.JudgeRegime 生成，展示在 Dashboard 顶部。
const (
	RegimeRiskOn  = "risk_on"  // 风险偏好：市场乐观，资金流向高风险资产
	RegimeRiskOff = "risk_off" // 风险规避：市场悲观，资金流向避险资产
	RegimeNeutral = "neutral"  // 中性：市场无明显方向
)
