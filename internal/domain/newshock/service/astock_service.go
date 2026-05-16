// Package service A 股数据采集服务，负责从数据源同步股票列表和日线行情。
//
// SyncStockList：拉取所有 A 股，自动创建不存在的 ticker（market=cn）。
// SyncDailyData：为所有 CN market 的 ticker 拉取最近 N 天 K 线，批量写入 ticker_daily 表。
package service

import (
	"context"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"auth-perm/config"
	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/repo"
)

type AStockService struct {
	tickerRepo    *repo.TickerRepo
	dailyRepo     *repo.TickerDailyRepo
	f10Repo       *repo.TickerF10Repo
	newsRepo      *repo.TickerNewsRepo
	stockProvider dm.StockListProvider
	klineProvider dm.KlineProvider
	f10Provider   dm.F10Provider
	newsProvider  dm.NewsProvider
	tenantID      string
	newsTopN      int
}

func NewAStockService(
	tickerRepo *repo.TickerRepo,
	dailyRepo *repo.TickerDailyRepo,
	f10Repo *repo.TickerF10Repo,
	newsRepo *repo.TickerNewsRepo,
	stockProvider dm.StockListProvider,
	klineProvider dm.KlineProvider,
	f10Provider dm.F10Provider,
	newsProvider dm.NewsProvider,
	cfg *config.Config,
) *AStockService {
	newsTopN := cfg.Stock.NewsTopN
	if newsTopN <= 0 {
		newsTopN = 200
	}
	return &AStockService{
		tickerRepo:    tickerRepo,
		dailyRepo:     dailyRepo,
		f10Repo:       f10Repo,
		newsRepo:      newsRepo,
		stockProvider: stockProvider,
		klineProvider: klineProvider,
		f10Provider:   f10Provider,
		newsProvider:  newsProvider,
		tenantID:      cfg.Stock.TenantID,
		newsTopN:      newsTopN,
	}
}

// SyncStockList 从数据源拉取所有证券（股票/ETF/可转债/指数），自动创建不存在的 ticker。
// symbol 使用 secid 格式：1.600519（沪市）/ 0.000001（深市）。
func (s *AStockService) SyncStockList(ctx context.Context) {
	log.Printf("[AStockService] SyncStockList started, tenantID=%s", s.tenantID)

	stocks, err := s.stockProvider.FetchAllStocks(ctx)
	if err != nil {
		log.Printf("[AStockService] fetch stock list error: %v", err)
		return
	}
	log.Printf("[AStockService] fetched %d securities from %s", len(stocks), s.stockProvider.Name())

	// 过滤特殊股票（仅对 stock 品种生效），构建待插入列表
	filtered := 0
	tickers := make([]dm.Ticker, 0, len(stocks))
	for _, stock := range stocks {
		secType := stock.SecurityType
		if secType == "" {
			secType = "stock"
		}
		// 只对股票过滤 ST/PT/B 股，ETF/指数/可转债不过滤
		if secType == "stock" && isSpecialStock(stock.Code, stock.Name) {
			filtered++
			continue
		}
		tickers = append(tickers, dm.Ticker{
			Symbol:       stock.Symbol,
			Name:         stock.Name,
			Market:       "cn",
			SecurityType: secType,
			TenantID:     s.tenantID,
		})
	}
	log.Printf("[AStockService] filtered %d special stocks, %d to insert", filtered, len(tickers))

	// 批量插入，冲突跳过（已存在的 symbol 不会重复插入）
	affected, err := s.tickerRepo.CreateBatchIfNotExists(ctx, tickers)
	if err != nil {
		log.Printf("[AStockService] batch create tickers error: %v", err)
		return
	}

	log.Printf("[AStockService] batch insert result: %d new rows (total candidates=%d)", affected, len(tickers))

	// 查询当前总数
	total, _ := s.tickerRepo.CountByTenant(ctx, s.tenantID)
	log.Printf("[AStockService] SyncStockList done: total tickers in DB=%d", total)
}

// SyncDailyData 为所有 CN market 的 ticker 拉取 K 线日线数据。
// 首次运行回溯 90 天，后续只拉取最近 2 天（当日 + 前一日）。
// 5 并发 goroutine + FailoverKlineProvider + 连续失败熔断。
func (s *AStockService) SyncDailyData(ctx context.Context) {
	// 获取所有 CN market 的 ticker
	tickers, err := s.tickerRepo.GetAllByTenant(ctx, s.tenantID)
	if err != nil {
		log.Printf("[AStockService] list tickers error: %v", err)
		return
	}

	var cnTickers []dm.Ticker
	for _, t := range tickers {
		if t.Market == "cn" {
			cnTickers = append(cnTickers, t)
		}
	}
	if len(cnTickers) == 0 {
		log.Println("[AStockService] no CN tickers found, run SyncStockList first")
		return
	}

	// 前置探测：先测 DB 再测 kline API，避免 DB 不可用时做无意义的 API 调用
	const minHealthyRecords = 20 // 低于此数量视为数据不足，触发 90 天回溯
	countMap, err := s.dailyRepo.CountAllByTenant(ctx, s.tenantID)
	if err != nil {
		log.Printf("[AStockService] DB unavailable (%v), skip daily sync", err)
		return
	}
	if _, err := s.klineProvider.FetchKline(ctx, cnTickers[0].Symbol, 2); err != nil {
		log.Printf("[AStockService] kline API unavailable (%v), skip daily sync", err)
		return
	}

	var saved atomic.Int64
	var failCount atomic.Int64
	var consecFail atomic.Int64
	var circuitMu sync.Mutex      // 熔断门：同一时间只允许一个 goroutine 执行暂停等待
	sem := make(chan struct{}, 5) // 并发度 5
	var wg sync.WaitGroup

	for _, ticker := range cnTickers {
		if ctx.Err() != nil {
			break
		}

		// 每只票单独判断回溯天数：无数据或数据不足拉 90 天，数据充足的拉 2 天
		days := 2
		if countMap[ticker.ID] < minHealthyRecords {
			days = 90
		}

		wg.Add(1)
		go func(t dm.Ticker, d int) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			// 间隔 200ms，避免触发限流（并发 5 时等效 5×200ms = 1s/批）
			time.Sleep(200 * time.Millisecond)

			// 熔断检测：连续失败 ≥20 时暂停 5 分钟，且只让一个 goroutine 进入等待
			if consecFail.Load() >= 20 {
				circuitMu.Lock()
				if consecFail.Load() >= 20 {
					c := consecFail.Load()
					log.Printf("[AStockService] circuit break: %d consecutive failures, pausing 5min", c)
					select {
					case <-ctx.Done():
						circuitMu.Unlock()
						return
					case <-time.After(5 * time.Minute):
					}
					consecFail.Store(0)
					log.Printf("[AStockService] circuit break reset, resuming")
				}
				circuitMu.Unlock()
			}

			bars, err := s.klineProvider.FetchKline(ctx, t.Symbol, d)
			if err != nil {
				fc := failCount.Add(1)
				cc := consecFail.Add(1)
				if fc <= 10 || fc%100 == 0 {
					log.Printf("[AStockService] fetch kline %s error: %v (fail=%d, consec=%d)", t.Symbol, err, fc, cc)
				}
				return
			}

			// 成功，重置连续失败计数
			consecFail.Store(0)

			records := make([]dm.TickerDaily, 0, len(bars))
			for _, bar := range bars {
				tradeDate, err := time.Parse("2006-01-02", bar.Date)
				if err != nil {
					continue
				}
				records = append(records, dm.TickerDaily{
					TickerID:  t.ID,
					TradeDate: tradeDate,
					Open:      bar.Open,
					High:      bar.High,
					Low:       bar.Low,
					Close:     bar.Close,
					Volume:    bar.Volume,
					Amount:    bar.Amount,
					ChangePct: bar.ChangePct,
					Turnover:  bar.Turnover,
					TenantID:  s.tenantID,
				})
			}

			if len(records) > 0 {
				if err := s.dailyRepo.UpsertBatch(ctx, records); err != nil {
					log.Printf("[AStockService] upsert daily %s error: %v", t.Symbol, err)
					return
				}
				saved.Add(int64(len(records)))
			}
		}(ticker, days)
	}
	wg.Wait()

	fc := failCount.Load()
	sv := saved.Load()
	if sv > 0 || fc > 0 {
		log.Printf("[AStockService] synced %d daily records for %d CN tickers (failed=%d)", sv, len(cnTickers), fc)
	}
}

// SyncF10Data 批量拉取 CN 股票的 F10 基本面数据（ETF/指数/可转债无 F10）。
// 分批处理，每批 50 只，避免 API 过载。连续失败 3 次提前退出。
func (s *AStockService) SyncF10Data(ctx context.Context) {
	cnTickers, err := s.tickerRepo.GetAllCNTickersByType(ctx, s.tenantID, "stock")
	if err != nil {
		log.Printf("[AStockService] list tickers for F10 error: %v", err)
		return
	}
	if len(cnTickers) == 0 {
		return
	}

	log.Printf("[AStockService] SyncF10Data started, %d CN tickers", len(cnTickers))

	// 构建 symbol → tickerID 映射
	symbolMap := make(map[string]string, len(cnTickers))
	symbols := make([]string, 0, len(cnTickers))
	for _, t := range cnTickers {
		symbolMap[t.Symbol] = t.ID
		symbols = append(symbols, t.Symbol)
	}

	// 分批获取，连续失败 3 次提前退出
	const batchSize = 50
	const maxConsecutiveFails = 3
	var saved int64
	consecutiveFails := 0
	for i := 0; i < len(symbols); i += batchSize {
		if ctx.Err() != nil {
			break
		}
		end := min(i+batchSize, len(symbols))
		batch := symbols[i:end]

		results, err := s.f10Provider.FetchF10(ctx, batch)
		if err != nil {
			consecutiveFails++
			log.Printf("[AStockService] fetch F10 batch %d error: %v (consecutive=%d)", i/batchSize, err, consecutiveFails)
			if consecutiveFails >= maxConsecutiveFails {
				log.Printf("[AStockService] SyncF10Data aborting after %d consecutive failures", maxConsecutiveFails)
				break
			}
			continue
		}
		consecutiveFails = 0

		// 将 TickerID 从 symbol 替换为 DB ticker ID
		records := make([]dm.TickerF10, 0, len(results))
		for _, r := range results {
			dbID, ok := symbolMap[r.TickerID]
			if !ok {
				continue
			}
			r.TickerID = dbID
			r.TenantID = s.tenantID
			records = append(records, r)
		}

		if len(records) > 0 {
			if err := s.f10Repo.UpsertBatch(ctx, records); err != nil {
				consecutiveFails++
				log.Printf("[AStockService] upsert F10 batch error: %v (consecutive=%d)", err, consecutiveFails)
				if consecutiveFails >= maxConsecutiveFails {
					break
				}
				continue
			}
			consecutiveFails = 0
			saved += int64(len(records))
		}
	}

	log.Printf("[AStockService] SyncF10Data done: saved %d records", saved)
}

// SyncStockNews 拉取热门 CN 股票的个股新闻（ETF/指数/可转债无个股新闻 API）。
// 按热度排序取前 N 只（默认 200），每只拉取最近 20 条新闻。
// 5 并发 goroutine，与 SyncDailyData 保持一致的并发策略。
func (s *AStockService) SyncStockNews(ctx context.Context) {
	targetTickers, err := s.tickerRepo.GetTopCNTickersByType(ctx, s.tenantID, "stock", s.newsTopN)
	if err != nil {
		log.Printf("[AStockService] list tickers for news error: %v", err)
		return
	}
	if len(targetTickers) == 0 {
		return
	}

	log.Printf("[AStockService] SyncStockNews started, top %d tickers", len(targetTickers))

	var saved atomic.Int64
	var failCount atomic.Int64
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup

	for _, ticker := range targetTickers {
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		go func(t dm.Ticker) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			news, err := s.newsProvider.FetchNews(ctx, t.Symbol, 20)
			if err != nil {
				fc := failCount.Add(1)
				if fc <= 5 {
					log.Printf("[AStockService] fetch news %s error: %v", t.Symbol, err)
				}
				return
			}

			records := make([]dm.TickerNews, 0, len(news))
			for _, n := range news {
				n.TickerID = t.ID
				n.TenantID = s.tenantID
				records = append(records, n)
			}

			if len(records) > 0 {
				if err := s.newsRepo.UpsertBatch(ctx, records); err != nil {
					log.Printf("[AStockService] upsert news %s error: %v", t.Symbol, err)
					failCount.Add(1)
					return
				}
				saved.Add(int64(len(records)))
			}
		}(ticker)
	}
	wg.Wait()

	sv := saved.Load()
	fc := failCount.Load()
	log.Printf("[AStockService] SyncStockNews done: saved %d records (failed=%d)", sv, fc)
}

// isSpecialStock 过滤 ST、退市、B 股等特殊股票。
// B 股用代码前缀判断（沪市 900xxx、深市 200xxx），不依赖名称后缀避免误杀。
func isSpecialStock(code, name string) bool {
	upper := strings.ToUpper(name)
	if strings.HasPrefix(upper, "ST") ||
		strings.HasPrefix(upper, "PT") ||
		strings.HasPrefix(upper, "*ST") ||
		strings.Contains(upper, "退") {
		return true
	}
	// B 股：沪市 900xxx / 深市 200xxx
	if strings.HasPrefix(code, "900") || strings.HasPrefix(code, "200") {
		return true
	}
	return false
}
