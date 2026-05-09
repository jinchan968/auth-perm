// PolymarketScheduler Polymarket 预测市场数据同步调度器。
//
// 定时从 Polymarket API 拉取活跃市场数据，存入数据库，并尝试与本地投资主题进行关键词匹配。
// 匹配算法详见 PolymarketService.findBestThemeMatch。
//
// Polymarket 数据更新频率较低，调度间隔为 4 小时（硬编码）。
// 启动时延迟 2 分钟，等待 RSS 和评分调度器先完成初始数据准备。
package service

import (
	"context"
	"log"
	"time"

	"auth-perm/config"
)

// PolymarketScheduler Polymarket 同步调度器，实现 container.Scheduler 接口。
type PolymarketScheduler struct {
	pmService *PolymarketService // Polymarket 服务
	interval  time.Duration      // 同步间隔（固定 4 小时）
}

// NewPolymarketScheduler 创建 Polymarket 调度器实例。
func NewPolymarketScheduler(pmService *PolymarketService, cfg *config.Config) *PolymarketScheduler {
	// Polymarket 数据更新频率较低，默认 4 小时
	interval := 4 * time.Hour
	return &PolymarketScheduler{
		pmService: pmService,
		interval:  interval,
	}
}

// Start 实现 container.Scheduler 接口，阻塞运行直到 ctx 取消。
// 启动时延迟 2 分钟，确保 RSS 和评分管线先完成初始数据准备。
func (s *PolymarketScheduler) Start(ctx context.Context) {
	log.Printf("[PolymarketScheduler] starting, interval=%v", s.interval)

	// 启动延迟 2 分钟，等其他数据先就绪
	select {
	case <-ctx.Done():
		return
	case <-time.After(2 * time.Minute):
	}
	s.pmService.SyncMarkets(ctx)

	// 定时循环
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[PolymarketScheduler] stopped")
			return
		case <-ticker.C:
			s.pmService.SyncMarkets(ctx)
		}
	}
}
