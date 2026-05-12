// 数据转换器：将数据库模型（dm）转换为 API 响应结构体（vo）。
// 集中管理转换逻辑，避免在 service 层散落重复的字段映射代码。
package vo

import (
	"auth-perm/internal/domain/newshock/dm"
)

// ToThemeResponse 将主题数据库模型转换为 API 响应结构体
func ToThemeResponse(t dm.Theme) ThemeResponse {
	return ThemeResponse{
		ID:                       t.ID,
		Name:                     t.Name,
		Description:              t.Description,
		Category:                 t.Category,
		Strength:                 t.Strength,
		StrengthNorm:             t.StrengthNorm,
		ClassificationConfidence: t.ClassificationConfidence,
		TickerCount:              t.TickerCount,
		EventCount:               t.EventCount,
		Trend:                    t.Trend,
		CreatedAt:                t.CreatedAt,
		UpdatedAt:                t.UpdatedAt,
	}
}

// ToTickerResponse 将股票数据库模型转换为 API 响应结构体
func ToTickerResponse(t dm.Ticker) TickerResponse {
	return TickerResponse{
		ID:           t.ID,
		Symbol:       t.Symbol,
		Name:         t.Name,
		Market:       t.Market,
		HotScore:     t.HotScore,
		MentionCount: t.MentionCount,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}

// ToEventResponse 将事件数据库模型转换为 API 响应结构体。
// 注意：Tickers 字段不在此处赋值，由 GetByID 等详情方法单独加载。
func ToEventResponse(e dm.Event) EventResponse {
	return EventResponse{
		ID:         e.ID,
		Title:      e.Title,
		Summary:    e.Summary,
		Channel:    e.Channel,
		Importance: e.Importance,
		ThemeID:    e.ThemeID,
		ThemeName:  e.ThemeName,
		EventTime:  e.EventTime,
		CreatedAt:  e.CreatedAt,
	}
}

// ToTickerDailyResponse 将日线行情数据库模型转换为 API 响应结构体
func ToTickerDailyResponse(d dm.TickerDaily) TickerDailyResponse {
	return TickerDailyResponse{
		Date:      d.TradeDate.Format("2006-01-02"),
		Open:      d.Open,
		High:      d.High,
		Low:       d.Low,
		Close:     d.Close,
		Volume:    d.Volume,
		Amount:    d.Amount,
		ChangePct: d.ChangePct,
		Turnover:  d.Turnover,
	}
}
