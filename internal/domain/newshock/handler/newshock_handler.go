// Package handler 实现 Newshock 模块的 HTTP 接口层。
// 所有接口挂在 /api/v1/newshock 路由组下，需要认证和 newshock.read/newshock.write 权限。
//
// 接口分为三类：
//  1. 只读接口（GET）：首页、详情、列表、搜索、边缘信号等
//  2. 管理接口（POST/PUT/DELETE）：创建、更新、删除主题/股票/事件
//  3. AI 接口（POST）：生成主题描述
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/controller/util"
	"auth-perm/internal/domain/newshock/service"
	"auth-perm/internal/domain/newshock/vo"
)

// NewshockHandler 聚合了所有 Newshock 子服务，负责接收 HTTP 请求并调用对应 Service。
type NewshockHandler struct {
	themeService  *service.ThemeService  // 主题 CRUD + 搜索
	tickerService *service.TickerService // 股票 CRUD + 搜索
	eventService  *service.EventService  // 事件 CRUD + 搜索
	statsService  *service.StatsService  // 聚合统计（首页、边缘信号、管线状态）
	searchService *service.SearchService // 全局搜索（并行搜索主题/股票/事件）
	aiService     *service.AIService     // AI 分析（事件重要度评估、主题描述生成、市场环境判断）
}

func NewNewshockHandler(
	themeService *service.ThemeService,
	tickerService *service.TickerService,
	eventService *service.EventService,
	statsService *service.StatsService,
	searchService *service.SearchService,
	aiService *service.AIService,
) *NewshockHandler {
	return &NewshockHandler{
		themeService:  themeService,
		tickerService: tickerService,
		eventService:  eventService,
		statsService:  statsService,
		searchService: searchService,
		aiService:     aiService,
	}
}

// getTenantID 从请求上下文中获取租户 ID，所有 newshock 接口都需要租户隔离。
func (h *NewshockHandler) getTenantID(c *gin.Context) string {
	tenantID, _ := util.GetTenantID(c)
	return tenantID
}

// clampPagination 限制分页参数在合理范围内：page>=1, 1<=pageSize<=100
func clampPagination(page, pageSize *int) {
	if *page < 1 {
		*page = 1
	}
	if *pageSize < 1 || *pageSize > 100 {
		*pageSize = 20
	}
}

// ==================== 首页聚合接口 ====================

// Home 首页雷达数据，一次返回所有首页需要的数据：
// 统计数字、市场环境、热门主题TOP5、热门股票TOP5、最新事件、数据新鲜度
// GET /api/v1/newshock/home
func (h *NewshockHandler) Home(c *gin.Context) {
	tenantID := h.getTenantID(c)
	data, err := h.statsService.GetHomeData(c.Request.Context(), tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取首页数据失败", err.Error())
		return
	}
	response.Success(c, data)
}

// Stats 统计面板数据，返回主题/股票/事件总数和平均强度
// GET /api/v1/newshock/stats
func (h *NewshockHandler) Stats(c *gin.Context) {
	tenantID := h.getTenantID(c)
	data, err := h.statsService.GetStats(c.Request.Context(), tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取统计数据失败", err.Error())
		return
	}
	response.Success(c, data)
}

// Regime 当前市场环境（risk_on/risk_off/neutral），由 AI 定期判断
// GET /api/v1/newshock/regime
func (h *NewshockHandler) Regime(c *gin.Context) {
	tenantID := h.getTenantID(c)
	data, err := h.statsService.GetRegime(c.Request.Context(), tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取市场环境失败", err.Error())
		return
	}
	response.Success(c, data)
}

// Edge 边缘信号，用于发现新兴投资机会：
// - 新兴主题（趋势上升但强度还较低）
// - 近期热门股票（被新闻频繁提及）
// - 高重要度最新事件
// GET /api/v1/newshock/edge
func (h *NewshockHandler) Edge(c *gin.Context) {
	tenantID := h.getTenantID(c)
	data, err := h.statsService.GetEdgeSignals(c.Request.Context(), tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取边缘信号失败", err.Error())
		return
	}
	response.Success(c, data)
}

// Pipeline 数据管线状态，展示 RSS 采集、事件提取、评分等各环节的运行状况
// GET /api/v1/newshock/pipeline
func (h *NewshockHandler) Pipeline(c *gin.Context) {
	tenantID := h.getTenantID(c)
	data, err := h.statsService.GetPipelineStatus(c.Request.Context(), tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取管线状态失败", err.Error())
		return
	}
	response.Success(c, data)
}

// ==================== 主题接口 ====================

// ListThemes 主题列表，支持按分类、趋势、关键词筛选，按强度排序
// GET /api/v1/newshock/themes?category=&trend=&keyword=&page=&page_size=&order_by=
func (h *NewshockHandler) ListThemes(c *gin.Context) {
	tenantID := h.getTenantID(c)
	var req vo.ListThemesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	clampPagination(&req.Page, &req.PageSize)
	data, err := h.themeService.List(c.Request.Context(), tenantID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询主题列表失败", err.Error())
		return
	}
	response.Success(c, data)
}

// GetTheme 主题详情，返回主题信息及其关联的股票、事件、Polymarket 数据
// GET /api/v1/newshock/themes/:id
func (h *NewshockHandler) GetTheme(c *gin.Context) {
	id := c.Param("id")
	tenantID := h.getTenantID(c)
	data, err := h.themeService.GetByID(c.Request.Context(), id, tenantID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "主题不存在", err.Error())
		return
	}
	response.Success(c, data)
}

// CreateTheme 创建主题，分类默认为 exploratory，可选值见 constant 包
// POST /api/v1/newshock/themes
func (h *NewshockHandler) CreateTheme(c *gin.Context) {
	tenantID := h.getTenantID(c)
	var req vo.CreateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	data, err := h.themeService.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建主题失败", err.Error())
		return
	}
	response.Success(c, data)
}

// UpdateTheme 更新主题，支持修改名称、描述、分类、趋势
// PUT /api/v1/newshock/themes/:id
func (h *NewshockHandler) UpdateTheme(c *gin.Context) {
	id := c.Param("id")
	tenantID := h.getTenantID(c)
	var req vo.UpdateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	data, err := h.themeService.Update(c.Request.Context(), id, tenantID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新主题失败", err.Error())
		return
	}
	response.Success(c, data)
}

// DeleteTheme 删除主题，同时清理关联的 theme_tickers 和事件的主题绑定
// DELETE /api/v1/newshock/themes/:id
func (h *NewshockHandler) DeleteTheme(c *gin.Context) {
	id := c.Param("id")
	tenantID := h.getTenantID(c)
	if err := h.themeService.Delete(c.Request.Context(), id, tenantID); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除主题失败", err.Error())
		return
	}
	response.Success(c, nil)
}

// GenerateThemeDescription 调用 AI 为指定主题生成描述文字。
// 会读取该主题的近期事件，让 AI 概括投资叙事和当前状态。
// POST /api/v1/newshock/themes/:id/generate-description
func (h *NewshockHandler) GenerateThemeDescription(c *gin.Context) {
	id := c.Param("id")
	tenantID := h.getTenantID(c)

	themeModel, err := h.themeService.GetRawByID(c.Request.Context(), id, tenantID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "主题不存在", err.Error())
		return
	}

	description := h.aiService.GenerateThemeDescription(c.Request.Context(), themeModel)
	if description != "" && description != themeModel.Description {
		_ = h.themeService.UpdateDescription(c.Request.Context(), id, tenantID, description)
	}

	response.Success(c, map[string]string{"description": description})
}

// ==================== 股票接口 ====================

// ListTickers 股票列表，支持按市场、关键词筛选
// GET /api/v1/newshock/tickers?market=&keyword=&page=&page_size=&order_by=
func (h *NewshockHandler) ListTickers(c *gin.Context) {
	tenantID := h.getTenantID(c)
	var req vo.ListTickersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	clampPagination(&req.Page, &req.PageSize)
	data, err := h.tickerService.List(c.Request.Context(), tenantID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询股票列表失败", err.Error())
		return
	}
	response.Success(c, data)
}

// GetTicker 股票详情，通过股票代码(如 AAPL)查询，返回股票信息及关联的主题和事件
// GET /api/v1/newshock/tickers/:symbol
func (h *NewshockHandler) GetTicker(c *gin.Context) {
	symbol := c.Param("symbol")
	tenantID := h.getTenantID(c)
	data, err := h.tickerService.GetBySymbol(c.Request.Context(), symbol, tenantID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "股票不存在", err.Error())
		return
	}
	response.Success(c, data)
}

// CreateTicker 创建股票标的
// POST /api/v1/newshock/tickers
func (h *NewshockHandler) CreateTicker(c *gin.Context) {
	tenantID := h.getTenantID(c)
	var req vo.CreateTickerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	data, err := h.tickerService.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建股票失败", err.Error())
		return
	}
	response.Success(c, data)
}

// UpdateTicker 更新股票信息
// PUT /api/v1/newshock/tickers/:id
func (h *NewshockHandler) UpdateTicker(c *gin.Context) {
	id := c.Param("id")
	tenantID := h.getTenantID(c)
	var req vo.UpdateTickerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	data, err := h.tickerService.Update(c.Request.Context(), id, tenantID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新股票失败", err.Error())
		return
	}
	response.Success(c, data)
}

// DeleteTicker 删除股票，同时清理所有 theme_tickers 和 event_tickers 关联
// DELETE /api/v1/newshock/tickers/:id
func (h *NewshockHandler) DeleteTicker(c *gin.Context) {
	id := c.Param("id")
	tenantID := h.getTenantID(c)
	if err := h.tickerService.Delete(c.Request.Context(), id, tenantID); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除股票失败", err.Error())
		return
	}
	response.Success(c, nil)
}

// ==================== 事件接口 ====================

// ListEvents 事件列表，支持按主题、渠道、重要度、关键词筛选
// GET /api/v1/newshock/events?theme_id=&channel=&importance=&keyword=&page=&page_size=&order_by=
func (h *NewshockHandler) ListEvents(c *gin.Context) {
	tenantID := h.getTenantID(c)
	var req vo.ListEventsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	clampPagination(&req.Page, &req.PageSize)
	data, err := h.eventService.List(c.Request.Context(), tenantID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询事件列表失败", err.Error())
		return
	}
	response.Success(c, data)
}

// GetEvent 事件详情，返回事件信息及关联的股票列表
// GET /api/v1/newshock/events/:id
func (h *NewshockHandler) GetEvent(c *gin.Context) {
	id := c.Param("id")
	tenantID := h.getTenantID(c)
	data, err := h.eventService.GetByID(c.Request.Context(), id, tenantID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "事件不存在", err.Error())
		return
	}
	response.Success(c, data)
}

// CreateEvent 创建事件，同时建立事件-股票关联，并更新主题的事件计数
// POST /api/v1/newshock/events
func (h *NewshockHandler) CreateEvent(c *gin.Context) {
	tenantID := h.getTenantID(c)
	var req vo.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	data, err := h.eventService.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建事件失败", err.Error())
		return
	}
	response.Success(c, data)
}

// UpdateEvent 更新事件
// PUT /api/v1/newshock/events/:id
func (h *NewshockHandler) UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	tenantID := h.getTenantID(c)
	var req vo.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	data, err := h.eventService.Update(c.Request.Context(), id, tenantID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新事件失败", err.Error())
		return
	}
	response.Success(c, data)
}

// DeleteEvent 删除事件，同时清理 event_tickers 关联并更新主题事件计数
// DELETE /api/v1/newshock/events/:id
func (h *NewshockHandler) DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	tenantID := h.getTenantID(c)
	if err := h.eventService.Delete(c.Request.Context(), id, tenantID); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除事件失败", err.Error())
		return
	}
	response.Success(c, nil)
}

// ==================== 搜索接口 ====================

// Search 全局搜索，同时在主题、股票、事件三个维度并行搜索关键词。
// 使用 errgroup 实现并行查询，提高响应速度。
// GET /api/v1/newshock/search?keyword=&limit=
func (h *NewshockHandler) Search(c *gin.Context) {
	tenantID := h.getTenantID(c)
	var req vo.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	if req.Limit < 1 || req.Limit > 50 {
		req.Limit = 10
	}
	if len(req.Keyword) > 200 {
		req.Keyword = req.Keyword[:200]
	}

	data, err := h.searchService.Search(c.Request.Context(), tenantID, req.Keyword, req.Limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "搜索失败", err.Error())
		return
	}
	response.Success(c, data)
}

// ==================== Polymarket 接口 ====================

// ListPolymarket 获取 Polymarket 预测市场列表
// GET /api/v1/newshock/polymarket
func (h *NewshockHandler) ListPolymarket(c *gin.Context) {
	tenantID := h.getTenantID(c)
	data, err := h.statsService.GetPolymarketMarkets(c.Request.Context(), tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取预测市场失败", err.Error())
		return
	}
	response.Success(c, data)
}
