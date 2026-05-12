// newshock 模块的依赖注入注册文件。
// 使用 dig 容器将 newshock 的所有 repo、service、handler 注册为可注入的依赖。
// 注册顺序：Repository → Provider → Service → Scheduler → Handler（按依赖层级从底向上）。
//
// # Provider 抽象架构
//
// A 股数据采集使用 Provider 模式，将数据源抽象为接口，通过 Failover 实现容错切换：
//
//	dm.StockListProvider  — 股票列表（sina → eastmoney → tdx → tushare）
//	dm.KlineProvider      — 日线 K 线（sina → tencent → eastmoney → tdx → tushare）
//	dm.BoardProvider      — 板块概念（eastmoney → tdx）
//
// 每个 Provider 接口定义在 dm/ 包，infra/ 包实现具体数据源，service/ 包实现 Failover 组合。
// 新增数据源只需：实现接口 → 加入 Failover 链 → 修改本文件对应 Provide 注册。
//
// 调度器只调用 Service 方法，Service 只依赖 Provider 接口，完全解耦：
//
//	StockListScheduler → AStockService.SyncStockList → StockListProvider.FetchAllStocks
//	DailyDataScheduler → AStockService.SyncDailyData → KlineProvider.FetchKline
//	ConceptScheduler   → ConceptService.SyncConcepts → BoardProvider.FetchBoards
package newshock

import (
	"go.uber.org/dig"
	"gorm.io/gorm"

	"auth-perm/config"
	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/handler"
	"auth-perm/internal/domain/newshock/repo"
	"auth-perm/internal/domain/newshock/service"
	"auth-perm/internal/infra/eastmoney"
	"auth-perm/internal/infra/llm"
	"auth-perm/internal/infra/sina"
	"auth-perm/internal/infra/tdx"
	"auth-perm/internal/infra/tencent"
	"auth-perm/internal/infra/tushare"
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
	if err := container.Provide(func(db *gorm.DB) *repo.TickerDailyRepo {
		return repo.NewTickerDailyRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.TickerConceptRepo {
		return repo.NewTickerConceptRepo(db)
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
		dailyRepo *repo.TickerDailyRepo,
		conceptRepo *repo.TickerConceptRepo,
	) *service.TickerService {
		return service.NewTickerService(tickerRepo, relationRepo, themeRepo, eventRepo, dailyRepo, conceptRepo)
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

	// ── Provider 层 ──────────────────────────────────────────────
	// 数据源抽象 + Failover 容错链。各 infra 客户端实现 dm/ 接口，
	// 由 FailoverProvider 按优先级串联，调用方无需关心底层数据源。

	// StockListProvider — 股票列表数据源：新浪 → 东方财富 → 通达信
	if err := container.Provide(func(cfg *config.Config) dm.StockListProvider {
		providers := []dm.StockListProvider{
			sina.NewClient(),      // 主源：腾讯批量行情 API（按代码段遍历）
			eastmoney.NewClient(), // 备用 1：东方财富全量接口（分页拉取）
			tdx.NewClient(),       // 备用 2：通达信协议（TCP 直连，无 HTTP 限制）
		}
		if cfg.Stock.TushareToken != "" {
			providers = append(providers, tushare.NewClient(cfg.Stock.TushareToken)) // 备用 3：Tushare Pro API（需 token）
		}
		return service.NewFailoverStockListProvider(providers...)
	}); err != nil {
		return err
	}
	// KlineProvider — K 线数据源：新浪 → 腾讯 → 东方财富 → 通达信
	if err := container.Provide(func(cfg *config.Config) dm.KlineProvider {
		providers := []dm.KlineProvider{
			sina.NewClient(),      // 主源：新浪 K 线 API
			tencent.NewClient(),   // 备用 1：腾讯前复权日线 API
			eastmoney.NewClient(), // 备用 2：东方财富日线 API
			tdx.NewClient(),       // 备用 3：通达信协议（TCP 直连，前复权日线）
		}
		if cfg.Stock.TushareToken != "" {
			providers = append(providers, tushare.NewClient(cfg.Stock.TushareToken)) // 备用 4：Tushare Pro API（需 token）
		}
		return service.NewFailoverKlineProvider(providers...)
	}); err != nil {
		return err
	}
	// A 股数据采集 — 依赖 StockListProvider + KlineProvider 接口
	if err := container.Provide(func(
		tickerRepo *repo.TickerRepo,
		dailyRepo *repo.TickerDailyRepo,
		stockProvider dm.StockListProvider,
		klineProvider dm.KlineProvider,
		cfg *config.Config,
	) *service.AStockService {
		return service.NewAStockService(tickerRepo, dailyRepo, stockProvider, klineProvider, cfg)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		astockService *service.AStockService,
		cfg *config.Config,
	) *service.StockListScheduler {
		return service.NewStockListScheduler(astockService, cfg)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		astockService *service.AStockService,
		cfg *config.Config,
	) *service.DailyDataScheduler {
		return service.NewDailyDataScheduler(astockService, cfg)
	}); err != nil {
		return err
	}
	// BoardProvider — 板块数据源，东方财富优先，通达信兜底
	if err := container.Provide(func() dm.BoardProvider {
		return service.NewFailoverBoardProvider(
			eastmoney.NewConceptClient(), // 主源：东方财富板块 API（返回 BK0800 等板块代码）
			tdx.NewClient(),              // 备用：通达信板块文件（block_gn.dat 概念、block.dat 行业）
		)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		conceptRepo *repo.TickerConceptRepo,
		tickerRepo *repo.TickerRepo,
		provider dm.BoardProvider,
		cfg *config.Config,
	) *service.ConceptService {
		return service.NewConceptService(conceptRepo, tickerRepo, provider, cfg)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(
		conceptService *service.ConceptService,
		cfg *config.Config,
	) *service.ConceptScheduler {
		return service.NewConceptScheduler(conceptService, cfg)
	}); err != nil {
		return err
	}

	// HTTP handler — API 接口层，依赖所有 service + Failover Provider（健康检查提取子 provider）
	if err := container.Provide(func(
		themeService *service.ThemeService,
		tickerService *service.TickerService,
		eventService *service.EventService,
		statsService *service.StatsService,
		searchService *service.SearchService,
		aiService *service.AIService,
		stockListProvider dm.StockListProvider,
		klineProvider dm.KlineProvider,
		boardProvider dm.BoardProvider,
	) *handler.NewshockHandler {
		return handler.NewNewshockHandler(
			themeService, tickerService, eventService, statsService, searchService, aiService,
			stockListProvider, klineProvider, boardProvider,
		)
	}); err != nil {
		return err
	}

	return nil
}
