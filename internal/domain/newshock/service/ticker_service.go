// Package service 股票服务，负责股票标的的 CRUD 和搜索。
// 股票标的代表一个可投资的金融标的（如 AAPL、TSLA），属于某个市场（us/cn/hk/kr）。
// 每只股票通过 theme_tickers 关联到投资主题，通过 event_tickers 关联到市场事件。
// hot_score（热度评分）由 ScoringService 根据 mention_count（被提及次数）定时计算。
package service

import (
	"context"

	"auth-perm/internal/domain/newshock/constant"
	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/repo"
	"auth-perm/internal/domain/newshock/vo"
)

type TickerService struct {
	tickerRepo   *repo.TickerRepo
	relationRepo *repo.RelationRepo
	themeRepo    *repo.ThemeRepo
	eventRepo    *repo.EventRepo
	dailyRepo    *repo.TickerDailyRepo
	conceptRepo  *repo.TickerConceptRepo
}

func NewTickerService(
	tickerRepo *repo.TickerRepo,
	relationRepo *repo.RelationRepo,
	themeRepo *repo.ThemeRepo,
	eventRepo *repo.EventRepo,
	dailyRepo *repo.TickerDailyRepo,
	conceptRepo *repo.TickerConceptRepo,
) *TickerService {
	return &TickerService{
		tickerRepo:   tickerRepo,
		relationRepo: relationRepo,
		themeRepo:    themeRepo,
		eventRepo:    eventRepo,
		dailyRepo:    dailyRepo,
		conceptRepo:  conceptRepo,
	}
}

// List 股票列表查询，支持按市场、关键词筛选，分页返回
func (s *TickerService) List(ctx context.Context, tenantID string, req vo.ListTickersRequest) (*vo.PagedResponse, error) {
	params := repo.TickerQueryParams{
		TenantID: tenantID,
		Market:   req.Market,
		Keyword:  req.Keyword,
		OrderBy:  req.OrderBy,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	tickers, total, err := s.tickerRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]vo.TickerResponse, 0, len(tickers))
	for _, t := range tickers {
		items = append(items, vo.ToTickerResponse(t))
	}
	return &vo.PagedResponse{Items: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// GetBySymbol 通过股票代码查询详情（如 AAPL），返回股票信息及关联的主题和事件。
// 关联数据加载：主题(通过 theme_tickers) → 事件(通过 event_tickers，最近20条)
func (s *TickerService) GetBySymbol(ctx context.Context, symbol, tenantID string) (*vo.TickerDetailResponse, error) {
	// 通过股票代码（如 AAPL）查询，带租户隔离
	ticker, err := s.tickerRepo.FindBySymbol(ctx, symbol, tenantID)
	if err != nil {
		return nil, err
	}

	resp := &vo.TickerDetailResponse{
		TickerResponse: vo.ToTickerResponse(*ticker),
	}

	// 加载关联主题：先查 theme_tickers 中间表，再批量查主题
	themeRelations, _ := s.relationRepo.GetThemesByTickerID(ctx, ticker.ID)
	if len(themeRelations) > 0 {
		themeIDs := make([]string, 0, len(themeRelations))
		for _, r := range themeRelations {
			themeIDs = append(themeIDs, r.ThemeID)
		}
		themes, _ := s.themeRepo.FindByIDs(ctx, themeIDs)
		for _, t := range themes {
			resp.Themes = append(resp.Themes, vo.ToThemeResponse(t))
		}
	}

	// 加载关联事件：先查 event_tickers 中间表，再批量查事件（最近 20 条）
	eventRelations, _ := s.relationRepo.GetEventsByTickerID(ctx, ticker.ID, 20)
	if len(eventRelations) > 0 {
		eventIDs := make([]string, 0, len(eventRelations))
		for _, r := range eventRelations {
			eventIDs = append(eventIDs, r.EventID)
		}
		events, _ := s.eventRepo.FindByIDs(ctx, eventIDs)
		for _, e := range events {
			resp.Events = append(resp.Events, vo.ToEventResponse(e))
		}
	}

	// 加载日线行情：最近 90 天
	dailyRecords, _ := s.dailyRepo.GetByTickerID(ctx, ticker.ID, 90)
	for _, d := range dailyRecords {
		resp.Daily = append(resp.Daily, vo.ToTickerDailyResponse(d))
	}

	// 加载概念板块（A股）
	concepts, _ := s.conceptRepo.GetByTickerID(ctx, ticker.ID)
	for _, c := range concepts {
		resp.Concepts = append(resp.Concepts, vo.TickerConceptResponse{Name: c.Name, Type: c.Type})
	}

	return resp, nil
}

// GetDailyBySymbol 获取指定股票的日线行情数据，支持自定义天数
func (s *TickerService) GetDailyBySymbol(ctx context.Context, symbol, tenantID string, days int) ([]vo.TickerDailyResponse, error) {
	ticker, err := s.tickerRepo.FindBySymbol(ctx, symbol, tenantID)
	if err != nil {
		return nil, err
	}
	records, err := s.dailyRepo.GetByTickerID(ctx, ticker.ID, days)
	if err != nil {
		return nil, err
	}
	items := make([]vo.TickerDailyResponse, 0, len(records))
	for _, d := range records {
		items = append(items, vo.ToTickerDailyResponse(d))
	}
	return items, nil
}

// Search 按关键词搜索股票（ILIKE 模糊匹配代码和名称）
func (s *TickerService) Search(ctx context.Context, tenantID, keyword string, limit int) ([]vo.TickerResponse, error) {
	tickers, err := s.tickerRepo.Search(ctx, tenantID, keyword, limit)
	if err != nil {
		return nil, err
	}
	items := make([]vo.TickerResponse, 0, len(tickers))
	for _, t := range tickers {
		items = append(items, vo.ToTickerResponse(t))
	}
	return items, nil
}

// GetTopTickers 获取热度最高的 TOP N 股票
func (s *TickerService) GetTopTickers(ctx context.Context, tenantID string, limit int) ([]vo.TickerResponse, error) {
	tickers, err := s.tickerRepo.GetTopTickers(ctx, tenantID, limit)
	if err != nil {
		return nil, err
	}
	items := make([]vo.TickerResponse, 0, len(tickers))
	for _, t := range tickers {
		items = append(items, vo.ToTickerResponse(t))
	}
	return items, nil
}

// Create 创建股票，市场值无效时默认为 us（美股）
func (s *TickerService) Create(ctx context.Context, tenantID string, req vo.CreateTickerRequest) (*vo.TickerResponse, error) {
	ticker := &dm.Ticker{
		Symbol:   req.Symbol,
		Name:     req.Name,
		Market:   req.Market,
		TenantID: tenantID,
	}
	if ticker.Market == "" || !vo.IsValidMarket(ticker.Market) {
		ticker.Market = constant.MarketUS
	}
	if err := s.tickerRepo.Create(ctx, ticker); err != nil {
		return nil, err
	}
	resp := vo.ToTickerResponse(*ticker)
	return &resp, nil
}

// Update 更新股票信息
func (s *TickerService) Update(ctx context.Context, id, tenantID string, req vo.UpdateTickerRequest) (*vo.TickerResponse, error) {
	ticker, err := s.tickerRepo.FindByIDAndTenantID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		ticker.Name = req.Name
	}
	if req.Market != "" && vo.IsValidMarket(req.Market) {
		ticker.Market = req.Market
	}
	if err := s.tickerRepo.Update(ctx, ticker); err != nil {
		return nil, err
	}
	resp := vo.ToTickerResponse(*ticker)
	return &resp, nil
}

// Delete 删除股票，同时清理所有 theme_tickers 和 event_tickers 关联
func (s *TickerService) Delete(ctx context.Context, id, tenantID string) error {
	// 先校验股票存在且属于该租户
	_, err := s.tickerRepo.FindByIDAndTenantID(ctx, id, tenantID)
	if err != nil {
		return err
	}
	// 清理关联数据：删除该股票在 theme_tickers 和 event_tickers 中的所有记录
	err = s.relationRepo.ClearThemeTickersByTicker(ctx, id)
	if err != nil {
		return err
	}
	err = s.relationRepo.ClearEventTickersByTicker(ctx, id)
	if err != nil {
		return err
	}
	// 删除股票本身
	return s.tickerRepo.Delete(ctx, id)
}
