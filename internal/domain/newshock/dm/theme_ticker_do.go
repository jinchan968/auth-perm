package dm

// ThemeTicker 主题-股票关联表（多对多中间表）。
// 表示某只股票属于某个投资主题。
type ThemeTicker struct {
	ThemeID  string `gorm:"column:theme_id;type:varchar(36);primaryKey"`
	TickerID string `gorm:"column:ticker_id;type:varchar(36);primaryKey"`
}

func (ThemeTicker) TableName() string { return "newshock_theme_tickers" }
