package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"auth-perm/internal/domain/newshock/dm"
)

// ProviderStatus 单个 provider 的探测结果。
type ProviderStatus struct {
	Name      string `json:"name"`            // 数据源名称
	Interface string `json:"interface"`       // 接口类型（stocklist/kline/board）
	OK        bool   `json:"ok"`              // 是否健康
	Latency   string `json:"latency"`         // 响应耗时
	Count     int    `json:"count"`           // 返回数据条数
	Error     string `json:"error,omitempty"` // 错误信息
}

// ProviderHealthResult 全部 provider 的探测结果。
type ProviderHealthResult struct {
	Results []ProviderStatus `json:"results"`
	Summary string           `json:"summary"`
}

// ProbeEvent 每次探测前的回调事件，用于 CLI 输出进度。
type ProbeEvent struct {
	Name      string // provider 名称
	Interface string // 接口类型
}

// CheckProviderHealth 探测所有已注册 provider 的健康状态。
// onProbe 回调在每次探测前调用（可为 nil），用于 CLI 输出进度。
// 每个探测使用独立的短超时（stocklist 15s，kline 10s，board 10s），不全量拉取。
func CheckProviderHealth(
	ctx context.Context,
	stockProviders []dm.StockListProvider,
	klineProviders []dm.KlineProvider,
	boardProviders []dm.BoardProvider,
	onProbe func(ProbeEvent),
) ProviderHealthResult {
	var results []ProviderStatus
	healthy, total := 0, 0

	// 探测 StockListProvider — Probe() 轻量验证，不做全量拉取
	for _, p := range stockProviders {
		total++
		if onProbe != nil {
			onProbe(ProbeEvent{Name: p.Name(), Interface: "stocklist"})
		}
		probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		status := probeProviderLight(probeCtx, p.Name(), "stocklist", p.Probe)
		cancel()
		if status.OK {
			healthy++
		}
		results = append(results, status)
	}

	// 探测 KlineProvider — 10s 超时，茅台 3 天 K 线
	for _, p := range klineProviders {
		total++
		if onProbe != nil {
			onProbe(ProbeEvent{Name: p.Name(), Interface: "kline"})
		}
		probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		status := func() ProviderStatus {
			defer cancel()
			bars, err := p.FetchKline(probeCtx, "1.600519", 3)
			return probeProvider(p.Name(), "kline", func() (int, error) {
				return len(bars), err
			})
		}()
		if status.OK {
			healthy++
		}
		results = append(results, status)
	}

	// 探测 BoardProvider — 10s 超时，概念板块
	for _, p := range boardProviders {
		total++
		if onProbe != nil {
			onProbe(ProbeEvent{Name: p.Name(), Interface: "board"})
		}
		probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		status := func() ProviderStatus {
			defer cancel()
			boards, err := p.FetchBoards(probeCtx, 3)
			return probeProvider(p.Name(), "board", func() (int, error) {
				return len(boards), err
			})
		}()
		if status.OK {
			healthy++
		}
		results = append(results, status)
	}

	return ProviderHealthResult{
		Results: results,
		Summary: fmt.Sprintf("%d/%d healthy", healthy, total),
	}
}

// probeProviderLight 轻量探测（Probe 模式），只验证 API 能响应。
func probeProviderLight(ctx context.Context, name, iface string, probeFn func(context.Context) error) ProviderStatus {
	start := time.Now()
	err := probeFn(ctx)
	elapsed := time.Since(start)

	status := ProviderStatus{
		Name:      name,
		Interface: iface,
		Latency:   formatDuration(elapsed),
	}

	if err != nil {
		status.Error = err.Error()
	} else {
		status.OK = true
		status.Count = 1 // Probe 模式下标记为 1 表示可达
	}

	return status
}

// probeProvider 探测单个 provider，记录延迟和结果。
func probeProvider(name, iface string, fn func() (int, error)) ProviderStatus {
	start := time.Now()
	count, err := fn()
	elapsed := time.Since(start)

	status := ProviderStatus{
		Name:      name,
		Interface: iface,
		Latency:   formatDuration(elapsed),
		Count:     count,
	}

	if err != nil {
		status.Error = err.Error()
		log.Printf("[ProviderHealth] %s/%s FAIL (%v): %v", name, iface, elapsed, err)
	} else if count == 0 {
		status.Error = "returned 0 records"
		log.Printf("[ProviderHealth] %s/%s EMPTY (%v)", name, iface, elapsed)
	} else {
		status.OK = true
		log.Printf("[ProviderHealth] %s/%s OK (%v, %d records)", name, iface, elapsed, count)
	}

	return status
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
