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
	stockProvider dm.StockListProvider
	klineProvider dm.KlineProvider
	tenantID      string
}

func NewAStockService(
	tickerRepo *repo.TickerRepo,
	dailyRepo *repo.TickerDailyRepo,
	stockProvider dm.StockListProvider,
	klineProvider dm.KlineProvider,
	cfg *config.Config,
) *AStockService {
	return &AStockService{
		tickerRepo:    tickerRepo,
		dailyRepo:     dailyRepo,
		stockProvider: stockProvider,
		klineProvider: klineProvider,
		tenantID:      cfg.Stock.TenantID,
	}
}

// SyncStockList 从数据源拉取所有 A 股，自动创建不存在的 ticker。
// symbol 使用 secid 格式：1.600519（沪市）/ 0.000001（深市）。
func (s *AStockService) SyncStockList(ctx context.Context) {
	log.Printf("[AStockService] SyncStockList started, tenantID=%s", s.tenantID)

	stocks, err := s.stockProvider.FetchAllStocks(ctx)
	if err != nil {
		log.Printf("[AStockService] fetch stock list error: %v", err)
		return
	}
	log.Printf("[AStockService] fetched %d stocks from %s", len(stocks), s.stockProvider.Name())

	// 过滤特殊股票，构建待插入列表
	filtered := 0
	tickers := make([]dm.Ticker, 0, len(stocks))
	for _, stock := range stocks {
		if isSpecialStock(stock.Code, stock.Name) {
			filtered++
			continue
		}
		tickers = append(tickers, dm.Ticker{
			Symbol:   stock.Symbol,
			Name:     stock.Name,
			Market:   "cn",
			TenantID: s.tenantID,
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

	// 前置探测：用第一条数据测试 kline API 是否可用
	if _, err := s.klineProvider.FetchKline(ctx, cnTickers[0].Symbol, 2); err != nil {
		log.Printf("[AStockService] kline API unavailable (%v), skip daily sync", err)
		return
	}

	// 获取每只票的最新数据日期，用于判断回溯天数。
	// 先用第一条探测 DB 是否可用，避免 DB 故障时所有 ticker 被误判为"无数据"触发大量 90 天回溯。
	latestMap := make(map[string]bool) // tickerID -> 是否有数据
	if _, err := s.dailyRepo.GetLatestByTickerID(ctx, cnTickers[0].ID); err != nil {
		log.Printf("[AStockService] DB unavailable (%v), skip daily sync", err)
		return
	}
	for _, t := range cnTickers {
		latest, _ := s.dailyRepo.GetLatestByTickerID(ctx, t.ID)
		latestMap[t.ID] = latest != nil
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

		// 每只票单独判断回溯天数：新票拉 90 天，已有数据的拉 2 天
		days := 2
		if !latestMap[ticker.ID] {
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
