package dm

import "context"

// F10Provider 股票基本面数据提供者接口。
// 不同实现各取所长：腾讯取估值，东财取财务指标。
type F10Provider interface {
	// Name 返回数据源名称
	Name() string
	// FetchF10 批量获取股票基本面数据。
	// codes 为 secid 格式（1.600519 / 0.000001），返回结果的 TickerID 字段为传入的 code。
	FetchF10(ctx context.Context, codes []string) ([]TickerF10, error)
}
