// newshock 模块的依赖注入注册文件。
// 使用 dig 容器将 newshock 的所有 repo、service、handler 注册为可注入的依赖。
// 注册顺序：Repository → Service → Scheduler → Handler（按依赖层级从底向上）。
package newshock

import (
	"go.uber.org/dig"
	"gorm.io/gorm"

	"auth-perm/config"
	"auth-perm/internal/domain/newshock/handler"
	"auth-perm/internal/domain/newshock/repo"
	"auth-perm/internal/domain/newshock/service"
	"auth-perm/internal/infra/llm"
)

// RegisterNewshockDomain 将 newshock 模块的所有组件注册到 dig 容器。
// 调用方（main.go 或 container.go）在应用启动时调用此函数完成依赖注入。
//
// 注册层级：
//   - Repository（7 个）：ThemeRepo, TickerRepo, EventRepo, RegimeRepo, NewsRawRepo, RelationRepo, PolymarketRepo
//   - Service（11 个）：ThemeService, TickerService, EventService, StatsService, SearchService,
//     AIService, RSSService, NewsProcessor, ScoringService, PolymarketService 及相关 Scheduler
//   - Handler（1 个）：NewshockHandler
func RegisterNewshockDomain(container *dig.Container) error {
	// Repository layer — 数据仓库层，封装所有数据库操作
	if err := container.Provide(func(db *gorm.DB) *repo.ThemeRepo {
		return repo.NewThemeRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.TickerRepo {
		return repo.NewTickerRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.EventRepo {
		return repo.NewEventRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.RegimeRepo {
		return repo.NewRegimeRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.NewsRawRepo {
		return repo.NewNewsRawRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.RelationRepo {
		return repo.NewRelationRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.PolymarketRepo {
		return repo.NewPolymarketRepo(db)
	}); err != nil {
		return err
	}

	// Service layer — 业务逻辑层
	if err := container.Provide(func(
		themeRepo *repo.ThemeRepo,
		relationRepo *repo.RelationRepo,
		tickerRepo *repo.TickerRepo,
		eventRepo *repo.EventRepo,
		pmRepo *repo.PolymarketRepo,
	) *service.ThemeService {
		return service.NewThemeService(themeRepo, relationRepo, tickerRepo, eventRepo, pmRepo)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		tickerRepo *repo.TickerRepo,
		relationRepo *repo.RelationRepo,
		themeRepo *repo.ThemeRepo,
		eventRepo *repo.EventRepo,
	) *service.TickerService {
		return service.NewTickerService(tickerRepo, relationRepo, themeRepo, eventRepo)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		eventRepo *repo.EventRepo,
		relationRepo *repo.RelationRepo,
		tickerRepo *repo.TickerRepo,
		themeRepo *repo.ThemeRepo,
	) *service.EventService {
		return service.NewEventService(eventRepo, relationRepo, tickerRepo, themeRepo)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		themeRepo *repo.ThemeRepo,
		tickerRepo *repo.TickerRepo,
		eventRepo *repo.EventRepo,
		regimeRepo *repo.RegimeRepo,
		newsRepo *repo.NewsRawRepo,
		pmRepo *repo.PolymarketRepo,
	) *service.StatsService {
		return service.NewStatsService(themeRepo, tickerRepo, eventRepo, regimeRepo, newsRepo, pmRepo)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		themeService *service.ThemeService,
		tickerService *service.TickerService,
		eventService *service.EventService,
	) *service.SearchService {
		return service.NewSearchService(themeService, tickerService, eventService)
	}); err != nil {
		return err
	}

	// AI 分析服务 — 调用 LLM 进行重要度评估、主题描述生成、市场环境判断
	if err := container.Provide(func(
		llmClient *llm.Client,
		themeRepo *repo.ThemeRepo,
		eventRepo *repo.EventRepo,
		regimeRepo *repo.RegimeRepo,
		cfg *config.Config,
	) *service.AIService {
		return service.NewAIService(llmClient, themeRepo, eventRepo, regimeRepo, cfg.RSS.TenantID)
	}); err != nil {
		return err
	}

	// RSS 采集 + 事件提取 — 新闻数据管线
	if err := container.Provide(func(
		newsRawRepo *repo.NewsRawRepo,
		cfg *config.Config,
	) *service.RSSService {
		return service.NewRSSService(newsRawRepo, cfg)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		newsRawRepo *repo.NewsRawRepo,
		eventRepo *repo.EventRepo,
		tickerRepo *repo.TickerRepo,
		themeRepo *repo.ThemeRepo,
		relationRepo *repo.RelationRepo,
		aiService *service.AIService,
		cfg *config.Config,
	) *service.NewsProcessor {
		return service.NewNewsProcessor(newsRawRepo, eventRepo, tickerRepo, themeRepo, relationRepo, aiService, cfg)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		rssService *service.RSSService,
		newsProcessor *service.NewsProcessor,
		cfg *config.Config,
	) *service.RSSScheduler {
		return service.NewRSSScheduler(rssService, newsProcessor, cfg)
	}); err != nil {
		return err
	}

	// 主题评分 + 趋势计算 + 市场环境判断 — 定时评分管线
	if err := container.Provide(func(
		themeRepo *repo.ThemeRepo,
		tickerRepo *repo.TickerRepo,
		eventRepo *repo.EventRepo,
		aiService *service.AIService,
		cfg *config.Config,
	) *service.ScoringService {
		return service.NewScoringService(themeRepo, tickerRepo, eventRepo, aiService, cfg)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		scoringService *service.ScoringService,
		cfg *config.Config,
	) *service.ScoringScheduler {
		return service.NewScoringScheduler(scoringService, cfg)
	}); err != nil {
		return err
	}

	// Polymarket 概率数据 — 预测市场同步管线
	if err := container.Provide(func(
		pmRepo *repo.PolymarketRepo,
		themeRepo *repo.ThemeRepo,
		cfg *config.Config,
	) *service.PolymarketService {
		return service.NewPolymarketService(pmRepo, themeRepo, cfg)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		pmService *service.PolymarketService,
		cfg *config.Config,
	) *service.PolymarketScheduler {
		return service.NewPolymarketScheduler(pmService, cfg)
	}); err != nil {
		return err
	}

	// HTTP handler — API 接口层，依赖所有 service
	if err := container.Provide(func(
		themeService *service.ThemeService,
		tickerService *service.TickerService,
		eventService *service.EventService,
		statsService *service.StatsService,
		searchService *service.SearchService,
		aiService *service.AIService,
	) *handler.NewshockHandler {
		return handler.NewNewshockHandler(themeService, tickerService, eventService, statsService, searchService, aiService)
	}); err != nil {
		return err
	}

	return nil
}
