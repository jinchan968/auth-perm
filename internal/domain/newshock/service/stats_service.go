// StatsService 聚合统计服务，为首页、边缘信号、管线状态等接口提供数据。
// 本身不包含复杂业务逻辑，主要从各 Repo 查询数据并组装成响应结构体。
package service

import (
	"context"
	"time"

	"auth-perm/internal/domain/newshock/constant"
	"auth-perm/internal/domain/newshock/repo"
	"auth-perm/internal/domain/newshock/vo"
)

type StatsService struct {
	themeRepo  *repo.ThemeRepo
	tickerRepo *repo.TickerRepo
	eventRepo  *repo.EventRepo
	regimeRepo *repo.RegimeRepo
	newsRepo   *repo.NewsRawRepo
	pmRepo     *repo.PolymarketRepo
}

func NewStatsService(
	themeRepo *repo.ThemeRepo,
	tickerRepo *repo.TickerRepo,
	eventRepo *repo.EventRepo,
	regimeRepo *repo.RegimeRepo,
	newsRepo *repo.NewsRawRepo,
	pmRepo *repo.PolymarketRepo,
) *StatsService {
	return &StatsService{
		themeRepo:  themeRepo,
		tickerRepo: tickerRepo,
		eventRepo:  eventRepo,
		regimeRepo: regimeRepo,
		newsRepo:   newsRepo,
		pmRepo:     pmRepo,
	}
}

// GetStats 获取统计数据：主题/股票/事件总数 + 平均主题强度
func (s *StatsService) GetStats(ctx context.Context, tenantID string) (*vo.StatsResponse, error) {
	_, themeTotal, err := s.themeRepo.List(ctx, repo.ThemeQueryParams{TenantID: tenantID, PageSize: 1})
	if err != nil {
		return nil, err
	}
	_, tickerTotal, err := s.tickerRepo.List(ctx, repo.TickerQueryParams{TenantID: tenantID, PageSize: 1})
	if err != nil {
		return nil, err
	}
	eventCount, err := s.eventRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	avgStrength, err := s.themeRepo.AvgStrengthByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &vo.StatsResponse{
		ThemeCount:  themeTotal,
		TickerCount: tickerTotal,
		EventCount:  eventCount,
		AvgStrength: avgStrength,
	}, nil
}

// GetRegime 获取最新市场环境。无数据时返回默认中性环境。
func (s *StatsService) GetRegime(ctx context.Context, tenantID string) (*vo.RegimeResponse, error) {
	regime, err := s.regimeRepo.GetLatest(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if regime == nil {
		return &vo.RegimeResponse{
			RegimeType: constant.RegimeNeutral,
			Confidence: 0.5,
			Summary:    "暂无市场环境数据",
		}, nil
	}
	return &vo.RegimeResponse{
		ID:         regime.ID,
		RegimeType: regime.RegimeType,
		Confidence: regime.Confidence,
		Summary:    regime.Summary,
		CreatedAt:  regime.CreatedAt,
	}, nil
}

// GetFreshness 获取数据新鲜度，返回主题/事件/股票各自的最后更新时间。
func (s *StatsService) GetFreshness(ctx context.Context, tenantID string) (*vo.FreshnessResponse, error) {
	events, _ := s.eventRepo.GetRecentEvents(ctx, tenantID, 1)
	themes, _ := s.themeRepo.GetTopThemes(ctx, tenantID, 1)
	tickers, _ := s.tickerRepo.GetTopTickers(ctx, tenantID, 1)

	now := time.Now()
	resp := &vo.FreshnessResponse{
		ThemesUpdated:  now,
		EventsUpdated:  now,
		TickersUpdated: now,
	}
	if len(events) > 0 {
		resp.EventsUpdated = events[0].UpdatedAt
	}
	if len(themes) > 0 {
		resp.ThemesUpdated = themes[0].UpdatedAt
	}
	if len(tickers) > 0 {
		resp.TickersUpdated = tickers[0].UpdatedAt
	}
	return resp, nil
}

// GetHomeData 首页聚合数据，一次查询返回首页需要的所有数据：
// 统计数字、市场环境、热门主题TOP5、热门股票TOP5、最新事件10条、数据新鲜度
func (s *StatsService) GetHomeData(ctx context.Context, tenantID string) (*vo.HomeDataResponse, error) {
	stats, err := s.GetStats(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	regime, err := s.GetRegime(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	topThemes, err := s.themeRepo.GetTopThemes(ctx, tenantID, 5)
	if err != nil {
		return nil, err
	}
	topTickers, err := s.tickerRepo.GetTopTickers(ctx, tenantID, 5)
	if err != nil {
		return nil, err
	}
	recentEvents, err := s.eventRepo.GetRecentEvents(ctx, tenantID, 10)
	if err != nil {
		return nil, err
	}
	freshness, err := s.GetFreshness(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	themeResponses := make([]vo.ThemeResponse, 0, len(topThemes))
	for _, t := range topThemes {
		themeResponses = append(themeResponses, toThemeResponse(t))
	}

	tickerResponses := make([]vo.TickerResponse, 0, len(topTickers))
	for _, t := range topTickers {
		tickerResponses = append(tickerResponses, toTickerResponse(t))
	}

	eventResponses := make([]vo.EventResponse, 0, len(recentEvents))
	for _, e := range recentEvents {
		eventResponses = append(eventResponses, toEventResponse(e))
	}

	return &vo.HomeDataResponse{
		Stats:        *stats,
		Regime:       regime,
		TopThemes:    themeResponses,
		TopTickers:   tickerResponses,
		RecentEvents: eventResponses,
		Freshness:    *freshness,
	}, nil
}

// GetEdgeSignals 边缘信号，用于发现新兴投资机会：
// - 上升趋势中强度较低的新兴主题（GetRisingEmerging）
// - 近期被频繁提及的热门股票（GetRecentHot）
// - 重要度>=4 的最新事件（GetRecentHighImpact）
func (s *StatsService) GetEdgeSignals(ctx context.Context, tenantID string) (*vo.EdgeResponse, error) {
	risingThemes, err := s.themeRepo.GetRisingEmerging(ctx, tenantID, 10)
	if err != nil {
		return nil, err
	}
	hotTickers, err := s.tickerRepo.GetRecentHot(ctx, tenantID, 10)
	if err != nil {
		return nil, err
	}
	recentEvents, err := s.eventRepo.GetRecentHighImpact(ctx, tenantID, 4, 10)
	if err != nil {
		return nil, err
	}

	themeResponses := make([]vo.ThemeResponse, 0, len(risingThemes))
	for _, t := range risingThemes {
		themeResponses = append(themeResponses, toThemeResponse(t))
	}

	tickerResponses := make([]vo.TickerResponse, 0, len(hotTickers))
	for _, t := range hotTickers {
		tickerResponses = append(tickerResponses, toTickerResponse(t))
	}

	eventResponses := make([]vo.EventResponse, 0, len(recentEvents))
	for _, e := range recentEvents {
		eventResponses = append(eventResponses, toEventResponse(e))
	}

	return &vo.EdgeResponse{
		RisingThemes: themeResponses,
		HotTickers:   tickerResponses,
		RecentEvents: eventResponses,
	}, nil
}

// GetPolymarketMarkets 获取 Polymarket 预测市场列表（最多 50 条）
func (s *StatsService) GetPolymarketMarkets(ctx context.Context, tenantID string) ([]vo.PolymarketResponse, error) {
	if s.pmRepo == nil {
		return nil, nil
	}
	markets, err := s.pmRepo.ListByTenant(ctx, tenantID, 50)
	if err != nil {
		return nil, err
	}
	result := make([]vo.PolymarketResponse, 0, len(markets))
	for _, m := range markets {
		result = append(result, vo.PolymarketResponse{
			ConditionID: m.ConditionID,
			Title:       m.Title,
			Description: m.Description,
			Outcome:     m.Outcome,
			Probability: m.Probability,
			Volume:      m.Volume,
			UpdatedAt:   m.UpdatedAt,
		})
	}
	return result, nil
}

// GetPipelineStatus 数据管线状态，展示各环节的数据量和最新运行时间：
// - news_total / news_unprocessed：RSS 新闻总数和待处理数
// - theme_count / ticker_count / event_count：各实体总数
// - polymarket_count：Polymarket 市场数
// - latest_news_time / latest_event_time：最新新闻和事件时间
func (s *StatsService) GetPipelineStatus(ctx context.Context, tenantID string) (*vo.PipelineStatus, error) {
	status := &vo.PipelineStatus{}

	newsTotal, err := s.newsRepo.CountTotal(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	status.NewsTotal = newsTotal

	newsUnprocessed, err := s.newsRepo.CountUnprocessed(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	status.NewsUnprocessed = newsUnprocessed

	status.LatestNewsTime, err = s.newsRepo.GetLatestTime(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	themeCount, err := s.themeRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	status.ThemeCount = themeCount

	tickerCount, err := s.tickerRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	status.TickerCount = tickerCount

	eventCount, err := s.eventRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	status.EventCount = eventCount

	events, err := s.eventRepo.GetRecentEvents(ctx, tenantID, 1)
	if err != nil {
		return nil, err
	}
	if len(events) > 0 {
		status.LatestEventTime = events[0].EventTime
	}

	if s.pmRepo != nil {
		pmCount, err := s.pmRepo.CountByTenant(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		status.PolymarketCount = pmCount
	}

	return status, nil
}
