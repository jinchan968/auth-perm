// Package service A 股数据同步调度器。
//
// 拆分为两个独立调度器并行运行：
//   - StockListScheduler：定时同步股票列表（默认 24h）
//   - DailyDataScheduler：定时同步日线行情（默认 4h）
//
// 两者共享同一个 AStockService 实例，互不阻塞。
package service

import (
	"context"
	"log"
	"time"

	"auth-perm/config"
)

// StockListScheduler 股票列表同步调度器。
// 定时从数据源拉取全部 A 股股票，写入 newshock_tickers 表。
type StockListScheduler struct {
	service      *AStockService
	interval     time.Duration
	startupDelay time.Duration
	enabled      bool
}

// NewStockListScheduler 创建股票列表调度器。
func NewStockListScheduler(service *AStockService, cfg *config.Config) *StockListScheduler {
	interval := time.Duration(cfg.Stock.SyncInterval) * time.Hour
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	startupDelay := time.Duration(cfg.Stock.StartupDelayMin) * time.Minute
	if startupDelay <= 0 {
		startupDelay = 5 * time.Minute
	}
	return &StockListScheduler{
		service:      service,
		interval:     interval,
		startupDelay: startupDelay,
		enabled:      cfg.Stock.Enabled,
	}
}

// Start 阻塞运行直到 ctx 取消。
func (s *StockListScheduler) Start(ctx context.Context) {
	if !s.enabled {
		log.Println("[StockListScheduler] disabled by config (STOCK_SCHEDULER_ENABLED=false)")
		return
	}
	log.Printf("[StockListScheduler] starting, interval=%v, startup_delay=%v", s.interval, s.startupDelay)

	select {
	case <-ctx.Done():
		return
	case <-time.After(s.startupDelay):
	}

	s.service.SyncStockList(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[StockListScheduler] stopped")
			return
		case <-ticker.C:
			s.service.SyncStockList(ctx)
		}
	}
}

// DailyDataScheduler 日线数据同步调度器。
// 定时为所有 CN ticker 拉取最近 N 天 K 线，写入 ticker_daily 表。
type DailyDataScheduler struct {
	service      *AStockService
	interval     time.Duration
	startupDelay time.Duration
	enabled      bool
}

// NewDailyDataScheduler 创建日线数据调度器。
// 启动延迟默认为 startup_delay_min * 2，确保股票列表先同步完成。
func NewDailyDataScheduler(service *AStockService, cfg *config.Config) *DailyDataScheduler {
	interval := time.Duration(cfg.Stock.DailySyncInterval) * time.Hour
	if interval <= 0 {
		interval = 4 * time.Hour
	}
	startupDelay := time.Duration(cfg.Stock.StartupDelayMin) * 2 * time.Minute
	if startupDelay <= 0 {
		startupDelay = 10 * time.Minute
	}
	return &DailyDataScheduler{
		service:      service,
		interval:     interval,
		startupDelay: startupDelay,
		enabled:      cfg.Stock.Enabled,
	}
}

// Start 阻塞运行直到 ctx 取消。
func (s *DailyDataScheduler) Start(ctx context.Context) {
	if !s.enabled {
		log.Println("[DailyDataScheduler] disabled by config (STOCK_SCHEDULER_ENABLED=false)")
		return
	}
	log.Printf("[DailyDataScheduler] starting, interval=%v, startup_delay=%v", s.interval, s.startupDelay)

	select {
	case <-ctx.Done():
		return
	case <-time.After(s.startupDelay):
	}

	s.service.SyncDailyData(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[DailyDataScheduler] stopped")
			return
		case <-ticker.C:
			s.service.SyncDailyData(ctx)
		}
	}
}
