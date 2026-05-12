package vo

type ListThemesRequest struct {
	Category string `form:"category"`
	Trend    string `form:"trend"`
	Keyword  string `form:"keyword"`
	OrderBy  string `form:"order_by"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

type ListTickersRequest struct {
	Market   string `form:"market"`
	Keyword  string `form:"keyword"`
	OrderBy  string `form:"order_by"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

type ListEventsRequest struct {
	ThemeID    string `form:"theme_id"`
	Channel    string `form:"channel"`
	Importance int    `form:"importance"`
	Keyword    string `form:"keyword"`
	OrderBy    string `form:"order_by"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

type SearchRequest struct {
	Keyword string `form:"keyword" binding:"required"`
	Limit   int    `form:"limit,default=10"`
}

// 管理端请求

type CreateThemeRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type UpdateThemeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Trend       string `json:"trend"`
}

type CreateTickerRequest struct {
	Symbol string `json:"symbol" binding:"required"`
	Name   string `json:"name"`
	Market string `json:"market"`
}

type UpdateTickerRequest struct {
	Name   string `json:"name"`
	Market string `json:"market"`
}

type CreateEventRequest struct {
	Title      string   `json:"title" binding:"required"`
	Summary    string   `json:"summary"`
	Channel    string   `json:"channel"`
	Importance int      `json:"importance"`
	ThemeID    string   `json:"theme_id"`
	TickerIDs  []string `json:"ticker_ids"`
}

type UpdateEventRequest struct {
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	Channel    string `json:"channel"`
	Importance int    `json:"importance"`
	ThemeID    string `json:"theme_id"`
}

type TickerDailyRequest struct {
	Days int `form:"days,default=90"` // 查询最近 N 天的日线数据
}
