// Package service A股概念板块同步服务，负责从东方财富 API 同步个股题材概念、行业板块、地域板块。
package service

import (
	"context"
	"log"
	"time"

	"auth-perm/config"
	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/repo"
)

type ConceptService struct {
	conceptRepo *repo.TickerConceptRepo
	tickerRepo  *repo.TickerRepo
	provider    dm.BoardProvider
	tenantID    string
}

func NewConceptService(
	conceptRepo *repo.TickerConceptRepo,
	tickerRepo *repo.TickerRepo,
	provider dm.BoardProvider,
	cfg *config.Config,
) *ConceptService {
	return &ConceptService{
		conceptRepo: conceptRepo,
		tickerRepo:  tickerRepo,
		provider:    provider,
		tenantID:    cfg.Stock.TenantID,
	}
}

// SyncConcepts 从东方财富同步概念板块和行业板块数据。
// 流程：拉取板块列表 → 遍历板块获取股票 → 匹配本地 ticker → 批量写入概念关系。
func (s *ConceptService) SyncConcepts(ctx context.Context) {
	start := time.Now()

	// 1. 获取所有本地 CN ticker，建立 symbol -> tickerID 映射
	allTickers, err := s.tickerRepo.GetAllByTenant(ctx, s.tenantID)
	if err != nil {
		log.Printf("[ConceptService] list tickers error: %v", err)
		return
	}
	symbolToID := make(map[string]string, len(allTickers))
	for _, t := range allTickers {
		if t.Market == "cn" {
			symbolToID[t.Symbol] = t.ID
		}
	}
	if len(symbolToID) == 0 {
		log.Println("[ConceptService] no CN tickers found, skip concept sync")
		return
	}

	totalWritten := 0

	// 2. 同步概念板块 (type=3)
	written, err := s.syncBoardType(ctx, 3, "concept", symbolToID)
	if err != nil {
		log.Printf("[ConceptService] sync concept boards error: %v", err)
	} else {
		totalWritten += written
	}

	// 3. 同步行业板块 (type=2)
	written, err = s.syncBoardType(ctx, 2, "industry", symbolToID)
	if err != nil {
		log.Printf("[ConceptService] sync industry boards error: %v", err)
	} else {
		totalWritten += written
	}

	log.Printf("[ConceptService] synced %d concept records in %v", totalWritten, time.Since(start).Round(time.Second))
}

// syncBoardType 同步指定类型的板块数据
func (s *ConceptService) syncBoardType(ctx context.Context, boardType int, conceptType string, symbolToID map[string]string) (int, error) {
	boards, err := s.provider.FetchBoards(ctx, boardType)
	if err != nil {
		return 0, err
	}
	log.Printf("[ConceptService] fetched %d %s boards", len(boards), conceptType)

	written := 0
	for i, board := range boards {
		if ctx.Err() != nil {
			return written, ctx.Err()
		}

		// 获取板块内股票，每个请求独立超时 30s
		boardCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		stocks, err := s.provider.FetchBoardStocks(boardCtx, board.Code)
		cancel()
		if err != nil {
			log.Printf("[ConceptService] fetch board %s(%s) stocks error: %v", board.Name, board.Code, err)
			continue
		}

		// 匹配本地 ticker，构建概念记录
		var concepts []dm.TickerConcept
		for _, stock := range stocks {
			tickerID, ok := symbolToID[stock.Symbol]
			if !ok {
				continue
			}
			concepts = append(concepts, dm.TickerConcept{
				TickerID:   tickerID,
				Name:       board.Name,
				Type:       conceptType,
				SourceCode: board.Code,
				TenantID:   s.tenantID,
			})
		}

		if len(concepts) > 0 {
			affected, err := s.conceptRepo.UpsertBatch(ctx, concepts)
			if err != nil {
				log.Printf("[ConceptService] upsert concepts for board %s error: %v", board.Name, err)
				continue
			}
			written += int(affected)
		}

		// 每处理 50 个板块打印进度
		if (i+1)%50 == 0 {
			log.Printf("[ConceptService] progress: %d/%d %s boards processed, %d records written", i+1, len(boards), conceptType, written)
		}

		// 限流：每个板块间隔 200ms
		time.Sleep(200 * time.Millisecond)
	}

	return written, nil
}
