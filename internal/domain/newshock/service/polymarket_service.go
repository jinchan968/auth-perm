// PolymarketService Polymarket 预测市场服务，负责：
//  1. 从 Polymarket API 拉取活跃市场数据（SyncMarkets）
//  2. 将市场数据与本地投资主题进行关键词匹配（matchThemes）
//
// 匹配算法（findBestThemeMatch）：
//   - 子串匹配：主题名完整出现在市场标题中 → +5 分，出现在描述中 → +2 分
//   - 字符级双字匹配：主题名的每对相邻字符出现在标题中 → +2 分，描述中 → +1 分
//   - 总分 >= 3 则匹配成功，将市场关联到该主题
//
// 每 4 小时由 PolymarketScheduler 调用一次。
package service

import (
	"context"
	"log"
	"strings"
	"time"

	"auth-perm/config"
	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/repo"
	"auth-perm/internal/infra/polymarket"
)

type PolymarketService struct {
	pmRepo    *repo.PolymarketRepo
	themeRepo *repo.ThemeRepo
	client    *polymarket.Client
	tenantID  string
}

func NewPolymarketService(
	pmRepo *repo.PolymarketRepo,
	themeRepo *repo.ThemeRepo,
	cfg *config.Config,
) *PolymarketService {
	return &PolymarketService{
		pmRepo:    pmRepo,
		themeRepo: themeRepo,
		client:    polymarket.NewClient(),
		tenantID:  cfg.RSS.TenantID,
	}
}

// SyncMarkets 从 Polymarket 拉取活跃市场并存储到数据库。
// 流程：拉取 API → upsert 到数据库 → 尝试匹配未关联的主题
func (s *PolymarketService) SyncMarkets(ctx context.Context) {
	markets, err := s.client.FetchActiveMarkets(ctx, 100)
	if err != nil {
		log.Printf("[PolymarketService] fetch markets error: %v", err)
		return
	}

	now := time.Now()
	saved := 0
	for _, m := range markets {
		if ctx.Err() != nil {
			return
		}
		pm := dm.PolymarketMarket{
			ConditionID:  m.ConditionID,
			Title:        m.Title,
			Description:  m.Description,
			Outcome:      m.Outcome,
			Probability:  m.Probability,
			Volume:       m.Volume,
			TenantID:     s.tenantID,
			LastSyncedAt: &now,
		}
		if err := s.pmRepo.Upsert(ctx, &pm); err != nil {
			log.Printf("[PolymarketService] upsert error: %v", err)
			continue
		}
		saved++
	}

	if saved > 0 {
		log.Printf("[PolymarketService] synced %d markets", saved)
	}

	// 尝试将未关联主题的市场与本地主题匹配
	s.matchThemes(ctx)
}

// matchThemes 将未关联主题的 Polymarket 市场与本地投资主题进行关键词匹配。
// 使用子串 + 双字匹配算法，总分 >= 3 则认为匹配成功。
func (s *PolymarketService) matchThemes(ctx context.Context) {
	unmatched, err := s.pmRepo.GetUnmatched(ctx, s.tenantID)
	if err != nil {
		return
	}
	if len(unmatched) == 0 {
		return
	}

	// 分页加载所有主题
	var themes []dm.Theme
	page := 1
	const batchSize = 200
	for {
		batch, _, err := s.themeRepo.List(ctx, repo.ThemeQueryParams{
			TenantID: s.tenantID,
			PageSize: batchSize,
			Page:     page,
		})
		if err != nil {
			return
		}
		themes = append(themes, batch...)
		if len(batch) < batchSize {
			break
		}
		page++
	}

	matched := 0
	for _, pm := range unmatched {
		titleLower := strings.ToLower(pm.Title)
		descLower := strings.ToLower(pm.Description)
		bestTheme := s.findBestThemeMatch(titleLower, descLower, themes)
		if bestTheme != "" {
			pm.ThemeID = bestTheme
			s.pmRepo.Upsert(ctx, &pm)
			matched++
		}
	}
	if matched > 0 {
		log.Printf("[PolymarketService] matched %d markets to themes", matched)
	}
}

// findBestThemeMatch 通过子串 + 字符级双字匹配找到最佳主题。
// 评分规则：
//   - 主题名完整出现在标题中 → +5 分
//   - 主题名完整出现在描述中 → +2 分
//   - 主题名的每对相邻字符（双字）出现在标题中 → +2 分
//   - 主题名的每对相邻字符（双字）出现在描述中 → +1 分
//
// 总分 >= 3 才返回匹配结果，避免误匹配。
func (s *PolymarketService) findBestThemeMatch(title, desc string, themes []dm.Theme) string {
	bestID := ""
	bestScore := 0

	for _, theme := range themes {
		score := 0
		nameLower := strings.ToLower(theme.Name)

		// 子串匹配：主题名出现在标题/描述中
		if strings.Contains(title, nameLower) {
			score += 5
		}
		if strings.Contains(desc, nameLower) {
			score += 2
		}

		// 双字匹配：主题名的每对相邻字符
		runes := []rune(nameLower)
		for i := 0; i < len(runes)-1; i++ {
			bigram := string(runes[i]) + string(runes[i+1])
			if strings.Contains(title, bigram) {
				score += 2
			}
			if strings.Contains(desc, bigram) {
				score += 1
			}
		}

		if score > bestScore {
			bestScore = score
			bestID = theme.ID
		}
	}

	if bestScore >= 3 {
		return bestID
	}
	return ""
}
