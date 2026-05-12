// KlineProvider 日线 K 线数据提供者抽象。
//
// 多个数据源（新浪、腾讯、东方财富）实现此接口，通过 FailoverKlineProvider 组合为容错链。
// 调用方（AStockService.SyncDailyData）只需调用 provider.FetchKline()，无需关心底层数据源。
//
// 数据流：
//
//	DailyDataScheduler → AStockService → KlineProvider.FetchKline → FailoverKlineProvider
//	  → sina.FetchKline / tencent.FetchKline / eastmoney.FetchKline / tdx.FetchKline / tushare.FetchKline
//
// 扩展新数据源：
//  1. 在 internal/infra/<source>/client.go 实现 KlineProvider 接口
//  2. 在 module.go 的 FailoverKlineProvider 构造中追加即可
package dm

import "context"

// KlineBar 日线 K 线数据（所有数据源共享的统一结构）。
// 各 infra 包（sina/tencent/eastmoney/tushare）返回此类型，不再各自定义 KlineBar。
type KlineBar struct {
	Date      string  // 日期 YYYY-MM-DD
	Open      float64 // 开盘价
	Close     float64 // 收盘价
	High      float64 // 最高价
	Low       float64 // 最低价
	Volume    int64   // 成交量（手）
	Amount    float64 // 成交额（元）
	ChangePct float64 // 涨跌幅 %
	Turnover  float64 // 换手率 %
}

// KlineProvider 日线 K 线数据提供者接口。
//
// 实现方：
//   - sina.Client      : 新浪财经 API（主源，HTTP）
//   - tencent.Client   : 腾讯财经 API（备用 1，HTTP）
//   - eastmoney.Client : 东方财富 API（备用 2，HTTP）
//   - tdx.Client       : 通达信协议（备用 3，TCP 直连，无频率限制）
//   - tushare.Client   : Tushare Pro API（备用 4，需 token）
//
// 组合方：
//   - FailoverKlineProvider : 按顺序尝试多个 provider，第一个成功的返回
type KlineProvider interface {
	// Name 返回数据源名称，用于日志标识（如 "sina"、"tencent"、"eastmoney"）
	Name() string
	// FetchKline 获取指定股票的日线 K 线数据。
	// secid: 1.600519（沪市）/ 0.000001（深市）
	// days: 获取最近 N 天的数据
	FetchKline(ctx context.Context, secid string, days int) ([]KlineBar, error)
}
