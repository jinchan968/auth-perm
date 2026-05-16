// MergeF10Provider 合并多个 F10 数据源，各取所长。
//
// 腾讯取估值（PE/PB/市值/换手率/涨跌停价/量比），
// 东财取财务指标（行业/股本/EPS/ROE），
// 合并后返回完整 F10 数据。
package service

import (
	"context"
	"fmt"
	"log"

	"auth-perm/internal/domain/newshock/dm"
)

// MergeF10Provider 合并多个 F10 数据源。
type MergeF10Provider struct {
	providers []dm.F10Provider
}

// NewMergeF10Provider 创建合并 F10 提供者。
func NewMergeF10Provider(providers ...dm.F10Provider) *MergeF10Provider {
	return &MergeF10Provider{providers: providers}
}

func (m *MergeF10Provider) Name() string { return "merge-f10" }

// FetchF10 从所有 provider 获取数据，按 TickerID 合并字段。
// 非空字段优先（后者的非空值覆盖前者的零值）。
// 所有 provider 都失败时返回 error。
func (m *MergeF10Provider) FetchF10(ctx context.Context, codes []string) ([]dm.TickerF10, error) {
	merged := make(map[string]*dm.TickerF10)
	var lastErr error
	successCount := 0

	for _, p := range m.providers {
		if ctx.Err() != nil {
			break
		}
		results, err := p.FetchF10(ctx, codes)
		if err != nil {
			log.Printf("[MergeF10Provider] %s FetchF10 error: %v", p.Name(), err)
			lastErr = err
			continue
		}
		successCount++

		for i := range results {
			r := &results[i]
			existing, ok := merged[r.TickerID]
			if !ok {
				cp := *r
				merged[r.TickerID] = &cp
				continue
			}
			// 合并：后者的非空字段覆盖前者的零值
			mergeF10(existing, r)
		}
	}

	if successCount == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("all F10 providers failed: %w", lastErr)
		}
		return nil, fmt.Errorf("all F10 providers failed")
	}

	out := make([]dm.TickerF10, 0, len(merged))
	for _, v := range merged {
		out = append(out, *v)
	}
	return out, nil
}

// mergeF10 将 src 的非零字段合入 dst。
func mergeF10(dst, src *dm.TickerF10) {
	if src.PeTtm != 0 {
		dst.PeTtm = src.PeTtm
	}
	if src.PeStatic != 0 {
		dst.PeStatic = src.PeStatic
	}
	if src.Pb != 0 {
		dst.Pb = src.Pb
	}
	if src.TotalMcap != 0 {
		dst.TotalMcap = src.TotalMcap
	}
	if src.FloatMcap != 0 {
		dst.FloatMcap = src.FloatMcap
	}
	if src.TurnoverRate != 0 {
		dst.TurnoverRate = src.TurnoverRate
	}
	if src.VolumeRatio != 0 {
		dst.VolumeRatio = src.VolumeRatio
	}
	if src.LimitUp != 0 {
		dst.LimitUp = src.LimitUp
	}
	if src.LimitDown != 0 {
		dst.LimitDown = src.LimitDown
	}
	if src.Industry != "" {
		dst.Industry = src.Industry
	}
	if src.TotalShares != 0 {
		dst.TotalShares = src.TotalShares
	}
	if src.FloatShares != 0 {
		dst.FloatShares = src.FloatShares
	}
	if src.Eps != 0 {
		dst.Eps = src.Eps
	}
	if src.Bvps != 0 {
		dst.Bvps = src.Bvps
	}
	if src.Roe != 0 {
		dst.Roe = src.Roe
	}
	if src.Source != "" {
		if dst.Source != "" {
			dst.Source = dst.Source + "+" + src.Source
		} else {
			dst.Source = src.Source
		}
	}
}
