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

// FetchKline 按顺序尝试各 provider，第一个返回充足数据的胜出。
// 如果某个 provider 返回的数据量不足请求天数的 1/3（至少 5 条），视为数据不完整，继续尝试下一个。
func (f *FailoverKlineProvider) FetchKline(ctx context.Context, secid string, days int) ([]dm.KlineBar, error) {
	// 最低数据量阈值：请求天数的 1/3，最少 2 条（避免单条无效数据绕过 failover）
	minBars := max(days/3, 2)

	var bestBars []dm.KlineBar
	var bestProvider string

	for _, p := range f.providers {
		if ctx.Err() != nil {
			break
		}
		bars, err := p.FetchKline(ctx, secid, days)
		if err != nil {
			log.Printf("[KlineProvider] %s FetchKline(%s) error: %v, trying next...", p.Name(), secid, err)
			continue
		}
		if len(bars) == 0 {
			continue
		}
		// 数据充足，直接返回
		if len(bars) >= minBars {
			return bars, nil
		}
		// 数据不足但非空，暂存，继续尝试其他 provider
		if bestBars == nil || len(bars) > len(bestBars) {
			bestBars = bars
			bestProvider = p.Name()
			log.Printf("[KlineProvider] %s FetchKline(%s) returned only %d bars (want >=%d), trying next...", p.Name(), secid, len(bars), minBars)
		}
	}

	// 所有 provider 都不充足，返回数据最多的那个
	if len(bestBars) > 0 {
		log.Printf("[KlineProvider] all providers returned insufficient data for %s, using best from %s (%d bars)", secid, bestProvider, len(bestBars))
		return bestBars, nil
	}
	return nil, fmt.Errorf("all kline providers failed for %s", secid)
}
