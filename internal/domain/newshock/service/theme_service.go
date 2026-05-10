// Package service 主题服务，负责投资主题的 CRUD 和搜索。
//
// 主题是 Newshock 的核心实体，代表一个投资叙事，如"AI半导体"、"地缘政治"。
// 每个主题关联多只股票(ThemeTicker)和多个事件(Event)。
// 主题的 strength 和 trend 由 ScoringService 定时计算。
package service

import (
	"context"

	"auth-perm/internal/domain/newshock/constant"
	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/repo"
	"auth-perm/internal/domain/newshock/vo"
)

type ThemeService struct {
	themeRepo    *repo.ThemeRepo
	relationRepo *repo.RelationRepo
	tickerRepo   *repo.TickerRepo
	eventRepo    *repo.EventRepo
	pmRepo       *repo.PolymarketRepo
}

func NewThemeService(
	themeRepo *repo.ThemeRepo,
	relationRepo *repo.RelationRepo,
	tickerRepo *repo.TickerRepo,
	eventRepo *repo.EventRepo,
	pmRepo *repo.PolymarketRepo,
) *ThemeService {
	return &ThemeService{
		themeRepo:    themeRepo,
		relationRepo: relationRepo,
		tickerRepo:   tickerRepo,
		eventRepo:    eventRepo,
		pmRepo:       pmRepo,
	}
}

// List 主题列表查询，支持按分类、趋势、关键词筛选，分页返回
func (s *ThemeService) List(ctx context.Context, tenantID string, req vo.ListThemesRequest) (*vo.PagedResponse, error) {
	params := repo.ThemeQueryParams{
		TenantID: tenantID,
		Category: req.Category,
		Trend:    req.Trend,
		Keyword:  req.Keyword,
		OrderBy:  req.OrderBy,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	themes, total, err := s.themeRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	items := make([]vo.ThemeResponse, 0, len(themes))
	for _, t := range themes {
		items = append(items, vo.ToThemeResponse(t))
	}
	return &vo.PagedResponse{Items: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// GetByID 主题详情，返回主题信息及其关联的股票、事件、Polymarket 数据。
// 关联数据加载顺序：股票(通过 theme_tickers) → 事件(通过 theme_id) → Polymarket(通过 theme_id)
func (s *ThemeService) GetByID(ctx context.Context, id, tenantID string) (*vo.ThemeDetailResponse, error) {
	// 先查询主题基本信息（带租户隔离校验）
	theme, err := s.themeRepo.FindByIDAndTenantID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	resp := &vo.ThemeDetailResponse{
		ThemeResponse: vo.ToThemeResponse(*theme),
	}

	// 加载关联股票：先查 theme_tickers 中间表获取 tickerID 列表，再批量查 tickers
	relations, _ := s.relationRepo.GetTickersByThemeID(ctx, id)
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

	// 加载关联事件（最近 20 条，按创建时间降序）
	events, _ := s.eventRepo.GetByThemeID(ctx, id, 20)
	for _, e := range events {
		resp.Events = append(resp.Events, vo.ToEventResponse(e))
	}

	// 加载关联的 Polymarket 预测市场（通过 theme_id 关联）
	if s.pmRepo != nil {
		pmMarkets, _ := s.pmRepo.GetByThemeID(ctx, id)
		for _, pm := range pmMarkets {
			resp.Polymarket = append(resp.Polymarket, vo.PolymarketResponse{
				ConditionID: pm.ConditionID,
				Title:       pm.Title,
				Description: pm.Description,
				Outcome:     pm.Outcome,
				Probability: pm.Probability,
				Volume:      pm.Volume,
				UpdatedAt:   pm.UpdatedAt,
			})
		}
	}

	return resp, nil
}

// Search 按关键词搜索主题（ILIKE 模糊匹配名称和描述）
func (s *ThemeService) Search(ctx context.Context, tenantID, keyword string, limit int) ([]vo.ThemeResponse, error) {
	themes, err := s.themeRepo.Search(ctx, tenantID, keyword, limit)
	if err != nil {
		return nil, err
	}
	items := make([]vo.ThemeResponse, 0, len(themes))
	for _, t := range themes {
		items = append(items, vo.ToThemeResponse(t))
	}
	return items, nil
}

// GetTopThemes 获取强度最高的 TOP N 主题
func (s *ThemeService) GetTopThemes(ctx context.Context, tenantID string, limit int) ([]vo.ThemeResponse, error) {
	themes, err := s.themeRepo.GetTopThemes(ctx, tenantID, limit)
	if err != nil {
		return nil, err
	}
	items := make([]vo.ThemeResponse, 0, len(themes))
	for _, t := range themes {
		items = append(items, vo.ToThemeResponse(t))
	}
	return items, nil
}

// Create 创建主题，分类为空时默认为 exploratory（探索中）
func (s *ThemeService) Create(ctx context.Context, tenantID string, req vo.CreateThemeRequest) (*vo.ThemeResponse, error) {
	theme := &dm.Theme{
		Name:     req.Name,
		Category: req.Category,
		TenantID: tenantID,
	}
	if theme.Category == "" {
		theme.Category = constant.CategoryExploratory
	} else if !vo.IsValidCategory(theme.Category) {
		theme.Category = constant.CategoryExploratory
	}
	if err := s.themeRepo.Create(ctx, theme); err != nil {
		return nil, err
	}
	resp := vo.ToThemeResponse(*theme)
	return &resp, nil
}

// Update 更新主题，只更新非空字段
func (s *ThemeService) Update(ctx context.Context, id, tenantID string, req vo.UpdateThemeRequest) (*vo.ThemeResponse, error) {
	theme, err := s.themeRepo.FindByIDAndTenantID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		theme.Name = req.Name
	}
	if req.Description != "" {
		theme.Description = req.Description
	}
	if req.Category != "" && vo.IsValidCategory(req.Category) {
		theme.Category = req.Category
	}
	if req.Trend != "" && vo.IsValidTrend(req.Trend) {
		theme.Trend = req.Trend
	}
	if err := s.themeRepo.Update(ctx, theme); err != nil {
		return nil, err
	}
	resp := vo.ToThemeResponse(*theme)
	return &resp, nil
}

// GetRawByID 返回原始主题模型（用于 AI 描述生成等需要完整模型的场景）
func (s *ThemeService) GetRawByID(ctx context.Context, id, tenantID string) (*dm.Theme, error) {
	return s.themeRepo.FindByIDAndTenantID(ctx, id, tenantID)
}

// UpdateDescription 更新主题描述（由 AI 生成后调用）
func (s *ThemeService) UpdateDescription(ctx context.Context, id, tenantID, description string) error {
	theme, err := s.themeRepo.FindByIDAndTenantID(ctx, id, tenantID)
	if err != nil {
		return err
	}
	theme.Description = description
	return s.themeRepo.Update(ctx, theme)
}

// Delete 删除主题，同时清理关联数据：theme_tickers 关联和事件的主题绑定
func (s *ThemeService) Delete(ctx context.Context, id, tenantID string) error {
	// 先校验主题存在且属于该租户
	_, err := s.themeRepo.FindByIDAndTenantID(ctx, id, tenantID)
	if err != nil {
		return err
	}
	// 清理关联数据：删除 theme_tickers 中间表记录，清除事件的 theme_id 绑定
	s.relationRepo.ClearThemeTickers(ctx, id)
	s.eventRepo.ClearByThemeID(ctx, id)
	// 删除主题本身
	return s.themeRepo.Delete(ctx, id)
}
