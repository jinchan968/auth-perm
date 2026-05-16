package dm

import "context"

// NewsProvider 个股新闻提供者接口。
type NewsProvider interface {
	// Name 返回数据源名称
	Name() string
	// FetchNews 获取指定股票的新闻列表。
	// secid 为 secid 格式（1.600519 / 0.000001），limit 为返回条数。
	FetchNews(ctx context.Context, secid string, limit int) ([]TickerNews, error)
}
