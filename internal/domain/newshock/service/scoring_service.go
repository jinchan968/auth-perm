// ScoringService 评分服务，负责计算主题强度、趋势和股票热度。
//
// 主题评分算法：
//
//	strength = 近7天事件数×2 + 近7天重要度总和 + 关联股票数×0.5
//	trend = 近7天事件数 / 前7天事件数的比值：
//	  >1.2 → "rising"（上升趋势）
//	  <0.8 → "declining"（下降趋势）
//	  其余 → "stable"（稳定）
//	strength_norm = 自身强度 / 最强主题强度 × 100（归一化到 0-100）
//
// 股票热度算法：
//
//	hot_score = log2(mention_count + 1) × 10
//	归一化 = 自身hot_score / 最高hot_score × 100
//
// 每个租户独立计算，遍历所有有主题数据的租户逐一评分。
package service

import (
	"context"
	"log"
	"math"
	"time"

	"auth-perm/config"
	"auth-perm/internal/domain/newshock/constant"
	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/repo"
)

type ScoringService struct {
	themeRepo  *repo.ThemeRepo
	tickerRepo *repo.TickerRepo
	eventRepo  *repo.EventRepo
	aiService  *AIService
}

func NewScoringService(
	themeRepo *repo.ThemeRepo,
	tickerRepo *repo.TickerRepo,
	eventRepo *repo.EventRepo,
	aiService *AIService,
	_ *config.Config,
) *ScoringService {
	return &ScoringService{
		themeRepo:  themeRepo,
		tickerRepo: tickerRepo,
		eventRepo:  eventRepo,
		aiService:  aiService,
	}
}

// ScoreAll 对所有租户执行评分：主题评分 → 股票评分 → AI 市场环境判断。
// 由 ScoringScheduler 定时调用，默认每 60 分钟执行一次。
func (s *ScoringService) ScoreAll(ctx context.Context) {
	tenantIDs, err := s.themeRepo.DistinctTenantIDs(ctx)
	if err != nil || len(tenantIDs) == 0 {
		log.Printf("[ScoringService] no tenants found, skip scoring")
		return
	}
	for _, tenantID := range tenantIDs {
		if ctx.Err() != nil {
			return
		}
		s.scoreThemesForTenant(ctx, tenantID)
		s.scoreTickersForTenant(ctx, tenantID)
		if s.aiService != nil {
			s.aiService.JudgeRegime(ctx, tenantID)
		}
	}
}

// scoreThemesForTenant 为单个租户计算所有主题的强度和趋势。
// 流程：
//  1. 分页加载该租户所有主题
//  2. 对每个主题：统计近7天事件数和重要度、前7天事件数
//  3. 计算 strength 和 trend
//  4. 找出最大 strength，用于归一化
//  5. 回写 strength_norm 并更新数据库
func (s *ScoringService) scoreThemesForTenant(ctx context.Context, tenantID string) {
	var allThemes []dm.Theme
	page := 1
	const batchSize = 200
	for {
		themes, _, err := s.themeRepo.List(ctx, repo.ThemeQueryParams{
			TenantID: tenantID,
			PageSize: batchSize,
			Page:     page,
		})
		if err != nil {
			log.Printf("[ScoringService] list themes page %d error: %v", page, err)
			return
		}
		allThemes = append(allThemes, themes...)
		if len(themes) < batchSize {
			break
		}
		page++
	}

	if len(allThemes) == 0 {
		return
	}

	now := time.Now()
	d7 := now.Add(-7 * 24 * time.Hour)
	d14 := now.Add(-14 * 24 * time.Hour)

	var maxStrength float64

	for i := range allThemes {
		theme := &allThemes[i]

		// 近7天数据：事件数量和重要度总和
		recent7Count, _ := s.eventRepo.CountByThemeSince(ctx, theme.ID, d7)
		recent7Imp, _ := s.eventRepo.SumImportanceByThemeSince(ctx, theme.ID, d7)

		// 前7天数据（8-14天前）：用于计算趋势
		prev7Count, _ := s.eventRepo.CountByThemeSince(ctx, theme.ID, d14)
		prev7Count = prev7Count - recent7Count // 减去近7天的，得到前7天的
		if prev7Count < 0 {
			prev7Count = 0
		}

		// 强度公式：近7天事件数×2 + 近7天重要度总和 + 关联股票数×0.5
		strength := float64(recent7Count)*2 + recent7Imp + float64(theme.TickerCount)*0.5
		theme.Strength = math.Round(strength*10) / 10

		// 趋势判定：近7天 vs 前7天
		theme.Trend = calcTrend(recent7Count, prev7Count)

		if theme.Strength > maxStrength {
			maxStrength = theme.Strength
		}
	}

	// 归一化并回写数据库
	for i := range allThemes {
		theme := &allThemes[i]
		if maxStrength > 0 {
			theme.StrengthNorm = math.Round(theme.Strength/maxStrength*100) / 10 * 10
		}
		if err := s.themeRepo.Update(ctx, &allThemes[i]); err != nil {
			log.Printf("[ScoringService] update theme %s error: %v", theme.ID, err)
		}
	}

	log.Printf("[ScoringService] scored %d themes for tenant %s, max_strength=%.1f", len(allThemes), tenantID, maxStrength)
}

// scoreTickersForTenant 为单个租户计算所有股票的热度评分。
// 公式：hot_score = log2(mention_count + 1) × 10
// 然后归一化到 0-100。
func (s *ScoringService) scoreTickersForTenant(ctx context.Context, tenantID string) {
	tickers, err := s.tickerRepo.GetAllByTenant(ctx, tenantID)
	if err != nil {
		log.Printf("[ScoringService] list tickers for tenant %s error: %v", tenantID, err)
		return
	}

	var maxHot float64
	for i := range tickers {
		t := &tickers[i]
		hot := math.Log2(float64(t.MentionCount)+1) * 10
		t.HotScore = math.Round(hot*10) / 10
		if t.HotScore > maxHot {
			maxHot = t.HotScore
		}
	}

	for i := range tickers {
		t := &tickers[i]
		if maxHot > 0 {
			t.HotScore = math.Round(t.HotScore / maxHot * 100)
		}
		if err := s.tickerRepo.Update(ctx, &tickers[i]); err != nil {
			log.Printf("[ScoringService] update ticker %s error: %v", t.ID, err)
		}
	}
}

// calcTrend 根据近7天和前7天的事件数比值判断趋势。
// 比值 >1.2 表示上升，<0.8 表示下降，其余为稳定。
// 前7天为0时：有事件则为上升，无事件则为稳定。
func calcTrend(recent7, prev7 int64) string {
	if prev7 == 0 {
		if recent7 > 0 {
			return constant.TrendRising
		}
		return constant.TrendStable
	}
	ratio := float64(recent7) / float64(prev7)
	if ratio > 1.2 {
		return constant.TrendRising
	}
	if ratio < 0.8 {
		return constant.TrendDeclining
	}
	return constant.TrendStable
}
