// FailoverStockListProvider 股票列表数据容错提供者。
//
// 按顺序尝试多个 StockListProvider，第一个返回非空结果的胜出。
// 典型链路：sina → eastmoney → tdx → tushare
//
// 容错策略：
//   - 某个 provider 返回 error 或 0 只股票时，自动切换到下一个
//   - 全部失败后返回 error
//
// 日志：每次 failover 都会打印 [StockListProvider] 日志。
package service

import (
	"context"
	"fmt"
	"log"

	"auth-perm/internal/domain/newshock/dm"
)

// FailoverStockListProvider 按顺序尝试多个股票列表数据源，第一个成功的返回结果。
type FailoverStockListProvider struct {
	providers []dm.StockListProvider
}

// NewFailoverStockListProvider 创建 failover 股票列表提供者。
// providers 按优先级排列，第一个为主源，后续为备用源。
//
// 用法（module.go）：
//
//	dm.StockListProvider → NewFailoverStockListProvider(sina.NewClient(), eastmoney.NewClient(), tdx.NewClient())
func NewFailoverStockListProvider(providers ...dm.StockListProvider) *FailoverStockListProvider {
	return &FailoverStockListProvider{providers: providers}
}

func (f *FailoverStockListProvider) Name() string { return "failover-stocklist" }

// Providers 返回内部的 provider 列表，用于健康检查逐个探测。
func (f *FailoverStockListProvider) Providers() []dm.StockListProvider { return f.providers }

// Probe 探测第一个可用的 provider。
func (f *FailoverStockListProvider) Probe(ctx context.Context) error {
	for _, p := range f.providers {
		if err := p.Probe(ctx); err == nil {
			return nil
		}
	}
	return fmt.Errorf("all stocklist providers probe failed")
}

// FetchAllStocks 按顺序尝试各 provider，第一个返回非空结果的胜出。
func (f *FailoverStockListProvider) FetchAllStocks(ctx context.Context) ([]dm.StockInfo, error) {
	for _, p := range f.providers {
		stocks, err := p.FetchAllStocks(ctx)
		if err == nil && len(stocks) > 0 {
			log.Printf("[StockListProvider] %s returned %d stocks", p.Name(), len(stocks))
			return stocks, nil
		}
		if err != nil {
			log.Printf("[StockListProvider] %s FetchAllStocks error: %v, trying next...", p.Name(), err)
		} else {
			log.Printf("[StockListProvider] %s returned 0 stocks, trying next...", p.Name())
		}
	}
	return nil, fmt.Errorf("all stock list providers failed")
}
