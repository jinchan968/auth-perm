// Package service RSS 采集服务，负责从配置的 RSS 源拉取新闻。
//
// 采集流程：
//  1. 遍历配置文件中所有 RSS Feed（reuters、bloomberg 等）
//  2. 使用 gofeed 库解析 RSS XML
//  3. 对每条新闻计算 SHA256 content_hash（title|link），查重去重
//  4. 新闻写入 news_raw 表，标记 processed=false
//  5. 后续由 NewsProcessor 提取事件
//
// 配置示例（config/app.yaml）：
//
//	rss:
//	  fetch_interval: 30        # 拉取间隔（分钟）
//	  tenant_id: "xxx"          # 默认租户
//	  user_agent: "NewshockBot/1.0"
//	  feeds:
//	    - url: "https://feeds.reuters.com/reuters/businessNews"
//	      source: "reuters"
//	      channel: "global_macro"
package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/mmcdole/gofeed"

	"auth-perm/config"
	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/repo"
)

type RSSService struct {
	newsRawRepo *repo.NewsRawRepo
	cfg         *config.RSSConfig
	parser      *gofeed.Parser
}

func NewRSSService(newsRawRepo *repo.NewsRawRepo, cfg *config.Config) *RSSService {
	parser := gofeed.NewParser()
	parser.UserAgent = cfg.RSS.UserAgent
	return &RSSService{
		newsRawRepo: newsRawRepo,
		cfg:         &cfg.RSS,
		parser:      parser,
	}
}

// FetchAll 拉取所有配置的 RSS feeds，逐个执行。
// 某个 feed 拉取失败不影响其他 feed。
func (s *RSSService) FetchAll(ctx context.Context) {
	for _, feed := range s.cfg.Feeds {
		if err := s.fetchFeed(ctx, feed); err != nil {
			log.Printf("[RSSService] fetch %s error: %v", feed.Source, err)
		}
	}
}

// fetchFeed 拉取单个 RSS feed，流程：
//  1. 解析 RSS XML 获取所有条目
//  2. 对每条目计算 content_hash（SHA256(title|link)）
//  3. 查 news_raw 表是否已有相同 hash，有则跳过（去重）
//  4. 新条目写入 news_raw 表
func (s *RSSService) fetchFeed(ctx context.Context, feedCfg config.FeedConfig) error {
	// 使用 gofeed 库解析 RSS XML，获取所有条目
	feed, err := s.parser.ParseURL(feedCfg.URL)
	if err != nil {
		return fmt.Errorf("parse %s: %w", feedCfg.URL, err)
	}

	newCount := 0
	for _, item := range feed.Items {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 计算内容哈希（SHA256(title|link)），用于去重
		hash := makeContentHash(item.Title, item.Link)
		existing, err := s.newsRawRepo.FindByContentHash(ctx, hash)
		if err != nil {
			log.Printf("[RSSService] hash check error: %v", err)
			continue
		}
		if existing != nil {
			continue // 已存在相同哈希的新闻，跳过
		}

		// 解析发布时间：优先用 PublishedParsed，其次用 UpdatedParsed
		var publishedAt *time.Time
		if item.PublishedParsed != nil {
			publishedAt = item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			publishedAt = item.UpdatedParsed
		}

		// 取新闻正文：优先用 Description，其次用 Content
		content := item.Description
		if content == "" && item.Content != "" {
			content = item.Content
		}

		// 构建 NewsRaw 记录，标题截断到 500 字符避免数据库溢出
		news := &dm.NewsRaw{
			Title:       truncate(item.Title, 500),
			Content:     content,
			Source:      feedCfg.Source,  // 来源标识（如 "reuters"）
			Channel:     feedCfg.Channel, // 渠道分类（如 "global_macro"）
			URL:         item.Link,
			PublishedAt: publishedAt,
			ContentHash: hash,
			Processed:   false, // 标记为未处理，等待 NewsProcessor 消费
			TenantID:    s.cfg.TenantID,
		}

		if err := s.newsRawRepo.Create(ctx, news); err != nil {
			log.Printf("[RSSService] create news error: %v", err)
			continue
		}
		newCount++
	}

	if newCount > 0 {
		log.Printf("[RSSService] %s: %d new items", feedCfg.Source, newCount)
	}
	return nil
}

// makeContentHash 计算新闻去重哈希：SHA256(title + "|" + link)
func makeContentHash(title, link string) string {
	h := sha256.Sum256([]byte(title + "|" + link))
	return fmt.Sprintf("%x", h)
}

// truncate 截断字符串到指定 rune 长度，避免超出数据库字段限制
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
