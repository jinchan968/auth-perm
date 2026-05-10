// 枚举校验器：提供各字段合法值的白名单校验。
// 集中管理校验逻辑，避免在 service 层散落魔法字符串和重复的校验代码。
package vo

import "auth-perm/internal/domain/newshock/constant"

// 主题分类白名单
var validCategories = map[string]bool{
	constant.CategoryGeopolitical:  true,
	constant.CategoryAISemi:        true,
	constant.CategoryMacroMonetary: true,
	constant.CategorySupplyChain:   true,
	constant.CategoryDefense:       true,
	constant.CategoryEnergy:        true,
	constant.CategoryEarningsEvent: true,
	constant.CategoryExploratory:   true,
	constant.CategoryPharma:        true,
	constant.CategoryRegulatory:    true,
}

// IsValidCategory 校验主题分类是否合法
func IsValidCategory(c string) bool {
	return validCategories[c]
}

// 趋势类型白名单
var validTrends = map[string]bool{
	constant.TrendRising:    true,
	constant.TrendStable:    true,
	constant.TrendDeclining: true,
}

// IsValidTrend 校验趋势类型是否合法
func IsValidTrend(t string) bool {
	return validTrends[t]
}

// 新闻渠道白名单
var validChannels = map[string]bool{
	constant.ChannelGlobalMacro:  true,
	constant.ChannelIndustryNews: true,
	constant.ChannelMarketFlow:   true,
}

// IsValidChannel 校验新闻渠道是否合法
func IsValidChannel(c string) bool {
	return validChannels[c]
}

// 市场类型白名单
var validMarkets = map[string]bool{
	constant.MarketUS: true,
	constant.MarketCN: true,
	constant.MarketHK: true,
	constant.MarketKR: true,
}

// IsValidMarket 校验市场类型是否合法
func IsValidMarket(m string) bool {
	return validMarkets[m]
}
