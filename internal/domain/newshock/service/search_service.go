// SearchService 全局搜索服务，在主题、股票、事件三个维度并行搜索。
// 使用 errgroup 实现并发查询，三个搜索任务同时执行，任一失败则整体失败。
package service

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"auth-perm/internal/domain/newshock/vo"
)

type SearchService struct {
	themeService  *ThemeService
	tickerService *TickerService
	eventService  *EventService
}

func NewSearchService(
	themeService *ThemeService,
	tickerService *TickerService,
	eventService *EventService,
) *SearchService {
	return &SearchService{
		themeService:  themeService,
		tickerService: tickerService,
		eventService:  eventService,
	}
}

// Search 并行搜索主题、股票、事件，返回合并结果。
// 三个搜索任务通过 errgroup 并发执行，任一出错则整体返回错误。
func (s *SearchService) Search(ctx context.Context, tenantID, keyword string, limit int) (*vo.SearchResponse, error) {
	var themes []vo.ThemeResponse
	var tickers []vo.TickerResponse
	var events []vo.EventResponse

	g, ctx := errgroup.WithContext(ctx)

	// 并行搜索三个维度
	g.Go(func() error {
		var err error
		themes, err = s.themeService.Search(ctx, tenantID, keyword, limit)
		return err
	})
	g.Go(func() error {
		var err error
		tickers, err = s.tickerService.Search(ctx, tenantID, keyword, limit)
		return err
	})
	g.Go(func() error {
		var err error
		events, err = s.eventService.Search(ctx, tenantID, keyword, limit)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	return &vo.SearchResponse{
		Themes:  themes,
		Tickers: tickers,
		Events:  events,
	}, nil
}
