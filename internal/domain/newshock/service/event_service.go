// Package service 事件服务，负责市场事件的 CRUD 和搜索。
// 事件是 Newshock 的核心数据单元，来源有两种：
//  1. 自动提取：RSS 新闻经 NewsProcessor 处理后自动生成
//  2. 手动创建：通过 API 接口手动添加
//
// 每个事件关联到一个投资主题(ThemeID)和多个股票标的(通过 event_tickers)。
// importance 重要度 1-5，自动创建时由 AI 评估，手动创建时由用户指定。
package service

import (
	"context"

	"auth-perm/internal/domain/newshock/constant"
	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/repo"
	"auth-perm/internal/domain/newshock/vo"
)

type EventService struct {
	eventRepo    *repo.EventRepo
	relationRepo *repo.RelationRepo
	tickerRepo   *repo.TickerRepo
	themeRepo    *repo.ThemeRepo
}

func NewEventService(
	eventRepo *repo.EventRepo,
	relationRepo *repo.RelationRepo,
	tickerRepo *repo.TickerRepo,
	themeRepo *repo.ThemeRepo,
) *EventService {
	return &EventService{
		eventRepo:    eventRepo,
		relationRepo: relationRepo,
		tickerRepo:   tickerRepo,
		themeRepo:    themeRepo,
	}
}

// List 事件列表查询，支持按主题、渠道、重要度、关键词筛选，分页返回
func (s *EventService) List(ctx context.Context, tenantID string, req vo.ListEventsRequest) (*vo.PagedResponse, error) {
	params := repo.EventQueryParams{
		TenantID:   tenantID,
		ThemeID:    req.ThemeID,
		Channel:    req.Channel,
		Importance: req.Importance,
		Keyword:    req.Keyword,
		OrderBy:    req.OrderBy,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	events, total, err := s.eventRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]vo.EventResponse, 0, len(events))
	for _, e := range events {
		items = append(items, vo.ToEventResponse(e))
	}
	return &vo.PagedResponse{Items: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// GetByID 事件详情，返回事件信息及关联的股票列表
func (s *EventService) GetByID(ctx context.Context, id, tenantID string) (*vo.EventResponse, error) {
	event, err := s.eventRepo.FindByIDAndTenantID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	resp := vo.ToEventResponse(*event)

	// 加载关联股票
	relations, _ := s.relationRepo.GetTickersByEventID(ctx, id)
	if len(relations) > 0 {
		tickerIDs := make([]string, 0, len(relations))
		for _, r := range relations {
			tickerIDs = append(tickerIDs, r.TickerID)
		}
		tickers, _ := s.tickerRepo.FindByIDs(ctx, tickerIDs)
		for _, t := range tickers {
			resp.Tickers = append(resp.Tickers, vo.ToTickerResponse(t))
		}
	}

	return &resp, nil
}

// GetRecentEvents 获取最新 N 条事件
func (s *EventService) GetRecentEvents(ctx context.Context, tenantID string, limit int) ([]vo.EventResponse, error) {
	events, err := s.eventRepo.GetRecentEvents(ctx, tenantID, limit)
	if err != nil {
		return nil, err
	}
	items := make([]vo.EventResponse, 0, len(events))
	for _, e := range events {
		items = append(items, vo.ToEventResponse(e))
	}
	return items, nil
}

// Search 按关键词搜索事件（ILIKE 模糊匹配标题和摘要）
func (s *EventService) Search(ctx context.Context, tenantID, keyword string, limit int) ([]vo.EventResponse, error) {
	events, err := s.eventRepo.Search(ctx, tenantID, keyword, limit)
	if err != nil {
		return nil, err
	}
	items := make([]vo.EventResponse, 0, len(events))
	for _, e := range events {
		items = append(items, vo.ToEventResponse(e))
	}
	return items, nil
}

// Create 创建事件，同时：
//   - 建立事件-股票关联（event_tickers）
//   - 更新关联主题的事件计数
//
// importance 不在 1-5 范围内时默认为 3
func (s *EventService) Create(ctx context.Context, tenantID string, req vo.CreateEventRequest) (*vo.EventResponse, error) {
	// 校验重要度范围，无效值默认为 3（中等）
	if req.Importance < 1 || req.Importance > 5 {
		req.Importance = 3
	}
	// 校验渠道值，无效值默认为 global_macro
	channel := req.Channel
	if channel == "" || !vo.IsValidChannel(channel) {
		channel = constant.ChannelGlobalMacro
	}
	// 创建事件记录
	event := &dm.Event{
		Title:      req.Title,
		Summary:    req.Summary,
		Channel:    channel,
		Importance: req.Importance,
		ThemeID:    req.ThemeID,
		TenantID:   tenantID,
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}
	// 建立事件-股票关联（event_tickers 中间表）
	for _, tickerID := range req.TickerIDs {
		s.relationRepo.AddEventTicker(ctx, event.ID, tickerID)
	}
	// 如果关联了主题，更新该主题的事件计数
	if event.ThemeID != "" {
		s.themeRepo.UpdateEventCount(ctx, event.ThemeID)
	}
	resp := vo.ToEventResponse(*event)
	return &resp, nil
}

// Update 更新事件，只更新非空字段
func (s *EventService) Update(ctx context.Context, id, tenantID string, req vo.UpdateEventRequest) (*vo.EventResponse, error) {
	event, err := s.eventRepo.FindByIDAndTenantID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	if req.Title != "" {
		event.Title = req.Title
	}
	if req.Summary != "" {
		event.Summary = req.Summary
	}
	if req.Channel != "" && vo.IsValidChannel(req.Channel) {
		event.Channel = req.Channel
	}
	if req.Importance >= 1 && req.Importance <= 5 {
		event.Importance = req.Importance
	}
	if req.ThemeID != "" {
		event.ThemeID = req.ThemeID
	}
	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, err
	}
	resp := vo.ToEventResponse(*event)
	return &resp, nil
}

// Delete 删除事件，同时清理 event_tickers 关联并更新主题事件计数
func (s *EventService) Delete(ctx context.Context, id, tenantID string) error {
	// 先查出事件（需要 themeID 用于后续更新计数）
	event, err := s.eventRepo.FindByIDAndTenantID(ctx, id, tenantID)
	if err != nil {
		return err
	}
	// 清理事件-股票关联
	s.relationRepo.ClearEventTickers(ctx, id)
	// 删除事件本身
	if err := s.eventRepo.Delete(ctx, id); err != nil {
		return err
	}
	// 更新关联主题的事件计数
	if event.ThemeID != "" {
		s.themeRepo.UpdateEventCount(ctx, event.ThemeID)
	}
	return nil
}
