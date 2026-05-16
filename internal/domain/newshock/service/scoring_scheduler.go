// Package service 评分调度器，定时执行主题评分 + 趋势计算 + 市场环境判断。
//
// 调用 ScoringService.ScoreAll，依次完成：
//  1. 遍历所有租户，对每个租户的主题和股票重新评分
//  2. 根据近期事件数量和重要度计算主题趋势（rising/stable/declining）
//  3. 调用 AI 判断当前市场环境（risk_on/risk_off/neutral）
//
// 调度间隔由 config.yaml 中 rss.score_interval（分钟）控制，默认 60 分钟。
// 启动时延迟 30 秒执行，等待 RSS 调度器先拉取一轮数据。
package service

import (
	"context"
	"log"
	"time"

	"auth-perm/config"
)

// ScoringScheduler 评分调度器，实现 container.Scheduler 接口。
type ScoringScheduler struct {
	scoringService *ScoringService // 评分服务
	interval       time.Duration   // 评分间隔（默认 60 分钟）
	enabled        bool            // 是否启用
}

// NewScoringScheduler 创建评分调度器实例，从配置读取评分间隔。
func NewScoringScheduler(scoringService *ScoringService, cfg *config.Config) *ScoringScheduler {
	interval := time.Duration(cfg.RSS.ScoreInterval) * time.Minute
	return &ScoringScheduler{
		scoringService: scoringService,
		interval:       interval,
		enabled:        cfg.Stock.Enabled,
	}
}

// Start 实现 container.Scheduler 接口，阻塞运行直到 ctx 取消。
// 启动时延迟 30 秒，确保 RSS 调度器已完成第一轮数据拉取。
func (s *ScoringScheduler) Start(ctx context.Context) {
	if !s.enabled {
		log.Println("[ScoringScheduler] disabled by config (STOCK_SCHEDULER_ENABLED=false)")
		return
	}
	log.Printf("[ScoringScheduler] starting, interval=%v", s.interval)

	// 启动时延迟 30 秒执行，等 RSS 先拉取一轮数据，确保有新闻可评分
	select {
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
	}
	s.scoringService.ScoreAll(ctx)

	// 定时循环
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[ScoringScheduler] stopped")
			return
		case <-ticker.C:
			s.scoringService.ScoreAll(ctx)
		}
	}
}
