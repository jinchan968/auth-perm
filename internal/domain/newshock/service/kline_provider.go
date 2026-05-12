// FailoverKlineProvider K 线数据容错提供者。
//
// 按顺序尝试多个 KlineProvider，第一个返回非空结果的胜出。
// 典型链路：sina → tencent → eastmoney → tdx → tushare
//
// 容错策略：
//   - 每个 provider 独立调用一次，不做 provider 内部重试（重试由各 provider 自行处理）
//   - 某个 provider 返回 error 或空数据时，自动切换到下一个
//   - 全部失败后返回 error
//
// 日志：每次 failover 都会打印 [KlineProvider] 日志，便于排查数据源问题。
package service

import (
	"context"
	"fmt"
	"log"

	"auth-perm/internal/domain/newshock/dm"
)

// FailoverKlineProvider 按顺序尝试多个 K 线数据源，第一个成功的返回结果。
type FailoverKlineProvider struct {
	providers []dm.KlineProvider
}

// NewFailoverKlineProvider 创建 failover K 线提供者。
// providers 按优先级排列，第一个为主源，后续为备用源。
//
// 用法（module.go）：
//
//	dm.KlineProvider → NewFailoverKlineProvider(sina.NewClient(), tencent.NewClient(), eastmoney.NewClient(), tdx.NewClient())
func NewFailoverKlineProvider(providers ...dm.KlineProvider) *FailoverKlineProvider {
	return &FailoverKlineProvider{providers: providers}
}

func (f *FailoverKlineProvider) Name() string { return "failover-kline" }

// Providers 返回内部的 provider 列表，用于健康检查逐个探测。
func (f *FailoverKlineProvider) Providers() []dm.KlineProvider { return f.providers }

// FetchKline 按顺序尝试各 provider，第一个返回非空结果的胜出。
func (f *FailoverKlineProvider) FetchKline(ctx context.Context, secid string, days int) ([]dm.KlineBar, error) {
	for _, p := range f.providers {
		bars, err := p.FetchKline(ctx, secid, days)
		if err == nil && len(bars) > 0 {
			return bars, nil
		}
		if err != nil {
			log.Printf("[KlineProvider] %s FetchKline(%s) error: %v, trying next...", p.Name(), secid, err)
		}
	}
	return nil, fmt.Errorf("all kline providers failed for %s", secid)
}
