// Package service A股概念板块同步调度器。
package service

import (
	"context"
	"log"
	"time"

	"auth-perm/config"
)

// ConceptScheduler 概念板块同步调度器，实现 container.Scheduler 接口。
type ConceptScheduler struct {
	service      *ConceptService
	interval     time.Duration
	startupDelay time.Duration
}

// NewConceptScheduler 创建概念板块调度器实例。
// 启动延迟默认为 startup_delay_min * 2，确保股票列表先同步完成。
func NewConceptScheduler(service *ConceptService, cfg *config.Config) *ConceptScheduler {
	interval := time.Duration(cfg.Stock.SyncInterval) * time.Hour
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	startupDelay := time.Duration(cfg.Stock.StartupDelayMin) * 2 * time.Minute
	if startupDelay <= 0 {
		startupDelay = 10 * time.Minute
	}
	return &ConceptScheduler{
		service:      service,
		interval:     interval,
		startupDelay: startupDelay,
	}
}

// Start 实现 container.Scheduler 接口，阻塞运行直到 ctx 取消。
func (s *ConceptScheduler) Start(ctx context.Context) {
	log.Printf("[ConceptScheduler] starting, interval=%v, startup_delay=%v", s.interval, s.startupDelay)

	// 启动延迟等待股票列表先同步完成
	select {
	case <-ctx.Done():
		return
	case <-time.After(s.startupDelay):
	}

	// 首次执行
	s.service.SyncConcepts(ctx)

	// 定时循环
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[ConceptScheduler] stopped")
			return
		case <-ticker.C:
			s.service.SyncConcepts(ctx)
		}
	}
}
