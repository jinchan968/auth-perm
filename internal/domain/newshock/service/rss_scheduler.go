// Package service RSS 新闻采集调度器，定时执行"拉取 → 去重 → 事件提取"管线。
//
// 执行流程（每个周期）：
//  1. RSSService.FetchAll — 从所有配置的 RSS 源拉取新闻，去重后写入 newshock_news_raw
//  2. NewsProcessor.ProcessUnprocessed — 处理未处理的新闻：提取事件、匹配股票和主题
//
// 调度间隔由 config.yaml 中 rss.fetch_interval（分钟）控制，默认 30 分钟。
// 启动时立即执行一次，无需等待第一个周期。
package service

import (
	"context"
	"log"
	"time"

	"auth-perm/config"
)

// RSSScheduler RSS 采集调度器，实现 container.Scheduler 接口。
type RSSScheduler struct {
	rssService    *RSSService    // RSS 拉取服务
	newsProcessor *NewsProcessor // 新闻 → 事件处理服务
	interval      time.Duration  // 采集间隔（默认 30 分钟）
}

// NewRSSScheduler 创建 RSS 调度器实例，从配置读取采集间隔。
func NewRSSScheduler(rssService *RSSService, newsProcessor *NewsProcessor, cfg *config.Config) *RSSScheduler {
	interval := time.Duration(cfg.RSS.FetchInterval) * time.Minute
	return &RSSScheduler{
		rssService:    rssService,
		newsProcessor: newsProcessor,
		interval:      interval,
	}
}

// Start 实现 container.Scheduler 接口，阻塞运行直到 ctx 取消。
// 启动时立即执行一轮完整管线（拉取 + 处理），之后按 interval 定时重复。
func (s *RSSScheduler) Start(ctx context.Context) {
	log.Printf("[RSSScheduler] starting, interval=%v", s.interval)

	// 启动时立即执行一次：拉取 → 处理
	s.rssService.FetchAll(ctx)
	if ctx.Err() != nil {
		return
	}
	s.newsProcessor.ProcessUnprocessed(ctx)

	// 定时循环
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[RSSScheduler] stopped")
			return
		case <-ticker.C:
			// 每个周期先拉取新新闻，再处理未处理的新闻
			s.rssService.FetchAll(ctx)
			if ctx.Err() != nil {
				return
			}
			s.newsProcessor.ProcessUnprocessed(ctx)
		}
	}
}
