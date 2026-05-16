// FailoverNewsProvider 新闻容错提供者。
// 按顺序尝试多个 NewsProvider，第一个成功的返回结果。
package service

import (
	"context"
	"fmt"
	"log"

	"auth-perm/internal/domain/newshock/dm"
)

// FailoverNewsProvider 按顺序尝试多个新闻数据源。
type FailoverNewsProvider struct {
	providers []dm.NewsProvider
}

// NewFailoverNewsProvider 创建 failover 新闻提供者。
func NewFailoverNewsProvider(providers ...dm.NewsProvider) *FailoverNewsProvider {
	return &FailoverNewsProvider{providers: providers}
}

func (f *FailoverNewsProvider) Name() string { return "failover-news" }

// FetchNews 按顺序尝试各 provider，第一个返回非空结果的胜出。
func (f *FailoverNewsProvider) FetchNews(ctx context.Context, secid string, limit int) ([]dm.TickerNews, error) {
	for _, p := range f.providers {
		if ctx.Err() != nil {
			break
		}
		news, err := p.FetchNews(ctx, secid, limit)
		if err != nil {
			log.Printf("[NewsProvider] %s FetchNews(%s) error: %v, trying next...", p.Name(), secid, err)
			continue
		}
		if len(news) > 0 {
			return news, nil
		}
	}
	return nil, fmt.Errorf("all news providers failed for %s", secid)
}
