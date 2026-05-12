// StockListProvider A 股股票列表数据提供者抽象。
//
// 多个数据源（新浪/腾讯批量行情、东方财富全量接口）实现此接口，
// 通过 FailoverStockListProvider 组合为容错链。
// 调用方（AStockService.SyncStockList）只需调用 provider.FetchAllStocks()。
//
// 数据流：
//
//	StockListScheduler → AStockService → StockListProvider.FetchAllStocks
//	  → FailoverStockListProvider → sina / eastmoney / tdx / tushare
//
// 扩展新数据源：
//  1. 在 internal/infra/<source>/client.go 实现 StockListProvider 接口
//  2. 在 module.go 的 FailoverStockListProvider 构造中追加即可
package dm

import "context"

// StockInfo A 股股票基本信息（所有数据源共享的统一结构）。
// 各 infra 包返回此类型，不再各自定义 StockInfo。
type StockInfo struct {
	Symbol string // secid 格式：1.600519 / 0.000001（沪市前缀 1，深市前缀 0）
	Code   string // 纯股票代码：600519
	Name   string // 股票名称
	Market int    // 市场：0=深市, 1=沪市
}

// StockListProvider 股票列表数据提供者接口。
//
// 实现方：
//   - sina.Client      : 腾讯财经批量行情 API（主源，HTTP 按代码段遍历）
//   - eastmoney.Client : 东方财富全量接口（备用 1，HTTP 分页拉取）
//   - tdx.Client       : 通达信协议（备用 2，TCP 直连，无频率限制）
//   - tushare.Client   : Tushare Pro API（备用 3，需 token）
//
// 组合方：
//   - FailoverStockListProvider : 按顺序尝试多个 provider，第一个成功的返回
type StockListProvider interface {
	// Name 返回数据源名称，用于日志标识
	Name() string
	// FetchAllStocks 获取所有 A 股股票列表。
	// 返回的 Symbol 统一为 secid 格式：1.600519（沪市）/ 0.000001（深市）。
	FetchAllStocks(ctx context.Context) ([]StockInfo, error)
	// Probe 轻量健康检查，只验证 API 能响应（单次请求），不做全量拉取。
	// 返回 nil 表示健康，error 表示不可达。
	Probe(ctx context.Context) error
}
