package service

import (
	"context"
	"log"
	"time"

	"auth-perm/config"
)

// F10DataScheduler F10 基本面数据同步调度器。
type F10DataScheduler struct {
	service      *AStockService
	interval     time.Duration
	startupDelay time.Duration
	enabled      bool
}

// NewF10DataScheduler 创建 F10 调度器。
// 启动延迟为 startup_delay_min * 3，确保股票列表和日线先同步完成。
func NewF10DataScheduler(service *AStockService, cfg *config.Config) *F10DataScheduler {
	interval := time.Duration(cfg.Stock.F10SyncInterval) * time.Hour
	if interval <= 0 {
		interval = 1 * time.Hour
	}
	startupDelay := time.Duration(cfg.Stock.StartupDelayMin) * 3 * time.Minute
	if startupDelay <= 0 {
		startupDelay = 15 * time.Minute
	}
	return &F10DataScheduler{
		service:      service,
		interval:     interval,
		startupDelay: startupDelay,
		enabled:      cfg.Stock.Enabled,
	}
}

func (s *F10DataScheduler) Start(ctx context.Context) {
	if !s.enabled {
		log.Println("[F10DataScheduler] disabled by config (STOCK_SCHEDULER_ENABLED=false)")
		return
	}
	log.Printf("[F10DataScheduler] starting, interval=%v, startup_delay=%v", s.interval, s.startupDelay)

	select {
	case <-ctx.Done():
		return
	case <-time.After(s.startupDelay):
	}

	s.service.SyncF10Data(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[F10DataScheduler] stopped")
			return
		case <-ticker.C:
			s.service.SyncF10Data(ctx)
		}
	}
}
