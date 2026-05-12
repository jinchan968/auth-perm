// BoardProvider A 股板块/概念数据提供者抽象。
//
// 多个数据源（东方财富、通达信）实现此接口，通过 FailoverBoardProvider 组合为容错链。
// 调用方（ConceptService.SyncConcepts）只需调用 provider.FetchBoards() + provider.FetchBoardStocks()。
//
// 数据流：
//
//	ConceptScheduler → ConceptService → BoardProvider.FetchBoards/FetchBoardStocks
//	  → FailoverBoardProvider → eastmoney.FetchBoards / tdx.FetchBoards
//
// 扩展新数据源：
//  1. 在 internal/infra/<source>/client.go 实现 BoardProvider 接口
//  2. 在 module.go 的 FailoverBoardProvider 构造中追加即可
package dm

import "context"

// BoardInfo 板块信息（统一数据结构）
type BoardInfo struct {
	Code string // 板块代码，如 BK0800（东方财富）或板块名称（通达信）
	Name string // 板块名称，如"人工智能"
}

// BoardStockInfo 板块内股票信息
type BoardStockInfo struct {
	Symbol string // secid 格式：1.600519 / 0.000001
	Code   string // 纯股票代码
	Name   string // 股票名称
	Market int    // 0=深市, 1=沪市
}

// BoardProvider 板块数据提供者接口。
//
// 实现方：
//   - eastmoney.ConceptClient : 东方财富板块 API（主源，返回板块代码如 BK0800）
//   - tdx.Client              : 通达信板块文件（备用源，通过 TDX 协议下载 block_gn.dat）
//
// 组合方：
//   - FailoverBoardProvider : 按顺序尝试多个 provider，第一个成功的返回
type BoardProvider interface {
	// Name 返回数据源名称，用于日志标识（如 "eastmoney"、"tdx"）
	Name() string
	// FetchBoards 获取指定类型的板块列表。
	// boardType: 3=概念板块, 2=行业板块, 1=地域板块
	FetchBoards(ctx context.Context, boardType int) ([]BoardInfo, error)
	// FetchBoardStocks 获取板块内所有股票。
	// boardCode: 板块代码（东方财富如 BK0800）或板块名称（通达信）
	FetchBoardStocks(ctx context.Context, boardCode string) ([]BoardStockInfo, error)
}
