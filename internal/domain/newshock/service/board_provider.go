// FailoverBoardProvider 板块数据容错提供者。
//
// 按顺序尝试多个 BoardProvider，第一个返回非空结果的胜出。
// 典型链路：eastmoney → tdx
//
// 容错策略：
//   - 某个 provider 返回 error 或空数据时，自动切换到下一个
//   - 全部失败后返回 error
//
// 日志：每次 failover 都会打印 [BoardProvider] 日志。
package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"auth-perm/internal/domain/newshock/dm"
)

// FailoverBoardProvider 按顺序尝试多个板块数据源，第一个返回非空结果的胜出。
type FailoverBoardProvider struct {
	providers []dm.BoardProvider
}

// NewFailoverBoardProvider 创建 failover 板块数据提供者。
// providers 按优先级排列，第一个为主源，后续为备用源。
//
// 用法（module.go）：
//
//	dm.BoardProvider → NewFailoverBoardProvider(eastmoney.NewConceptClient(), tdx.NewClient())
func NewFailoverBoardProvider(providers ...dm.BoardProvider) *FailoverBoardProvider {
	return &FailoverBoardProvider{providers: providers}
}

func (f *FailoverBoardProvider) Name() string { return "failover-board" }

// Providers 返回内部的 provider 列表，用于健康检查逐个探测。
func (f *FailoverBoardProvider) Providers() []dm.BoardProvider { return f.providers }

// FetchBoards 按顺序尝试各 provider，第一个返回非空结果的胜出。
func (f *FailoverBoardProvider) FetchBoards(ctx context.Context, boardType int) ([]dm.BoardInfo, error) {
	for _, p := range f.providers {
		boards, err := p.FetchBoards(ctx, boardType)
		if err == nil && len(boards) > 0 {
			return boards, nil
		}
		if err != nil {
			log.Printf("[BoardProvider] %s FetchBoards error: %v, trying next...", p.Name(), err)
		} else {
			log.Printf("[BoardProvider] %s returned 0 boards, trying next...", p.Name())
		}
	}
	return nil, fmt.Errorf("all board providers failed for type %d", boardType)
}

// FetchBoardStocks 按顺序尝试各 provider，第一个返回非空结果的胜出。
func (f *FailoverBoardProvider) FetchBoardStocks(ctx context.Context, boardCode string) ([]dm.BoardStockInfo, error) {
	for _, p := range f.providers {
		stocks, err := p.FetchBoardStocks(ctx, boardCode)
		if err == nil && len(stocks) > 0 {
			return stocks, nil
		}
		if err != nil {
			// 404/not-found 表示板块无成分股数据，属于正常情况，不打日志
			if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found") {
				log.Printf("[BoardProvider] %s FetchBoardStocks(%s) error: %v, trying next...", p.Name(), boardCode, err)
			}
		}
	}
	return nil, fmt.Errorf("all board providers failed for board %s", boardCode)
}
