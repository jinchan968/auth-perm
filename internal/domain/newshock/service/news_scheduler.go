package service

import (
	"context"
	"log"
	"time"

	"auth-perm/config"
)

// StockNewsScheduler 个股新闻同步调度器。
type StockNewsScheduler struct {
	service      *AStockService
	interval     time.Duration
	startupDelay time.Duration
	enabled      bool
}

// NewStockNewsScheduler 创建新闻调度器。
// 启动延迟为 startup_delay_min * 3，确保股票列表先同步完成。
func NewStockNewsScheduler(service *AStockService, cfg *config.Config) *StockNewsScheduler {
	interval := time.Duration(cfg.Stock.NewsSyncInterval) * time.Hour
	if interval <= 0 {
		interval = 1 * time.Hour
	}
	startupDelay := time.Duration(cfg.Stock.StartupDelayMin) * 3 * time.Minute
	if startupDelay <= 0 {
		startupDelay = 15 * time.Minute
	}
	return &StockNewsScheduler{
		service:      service,
		interval:     interval,
		startupDelay: startupDelay,
		enabled:      cfg.Stock.Enabled,
	}
}

func (s *StockNewsScheduler) Start(ctx context.Context) {
	if !s.enabled {
		log.Println("[StockNewsScheduler] disabled by config (STOCK_SCHEDULER_ENABLED=false)")
		return
	}
	log.Printf("[StockNewsScheduler] starting, interval=%v, startup_delay=%v", s.interval, s.startupDelay)

	select {
	case <-ctx.Done():
		return
	case <-time.After(s.startupDelay):
	}

	s.service.SyncStockNews(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[StockNewsScheduler] stopped")
			return
		case <-ticker.C:
			s.service.SyncStockNews(ctx)
		}
	}
}
