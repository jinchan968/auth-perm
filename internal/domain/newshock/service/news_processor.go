// Package service 新闻处理器，负责将 news_raw 中的原始新闻转化为 Event（市场事件）。
//
// 处理流程：
//  1. 按租户遍历，加载该租户所有股票标的（Ticker）
//  2. 获取未处理的 news_raw（processed=false），每次最多 100 条
//  3. 对每条新闻：
//     a. 分词后在标题+正文中匹配股票代码（如 AAPL、TSLA）
//     b. 匹配成功 → 调用 AI 评估重要度 → 创建 Event
//     c. 建立 Event-Ticker 关联，增加 Ticker 的 mention_count
//     d. 通过关联的 Ticker 反查最匹配的 Theme，将事件归入该主题
//     e. 标记 news_raw 为 processed=true
//  4. 匹配不到任何股票的新闻直接标记为已处理（跳过）
package service

import (
	"context"
	"log"
	"strings"

	"auth-perm/config"
	"auth-perm/internal/domain/newshock/constant"
	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/repo"
)

type NewsProcessor struct {
	newsRawRepo  *repo.NewsRawRepo
	eventRepo    *repo.EventRepo
	tickerRepo   *repo.TickerRepo
	themeRepo    *repo.ThemeRepo
	relationRepo *repo.RelationRepo
	aiService    *AIService
	tenantID     string
}

func NewNewsProcessor(
	newsRawRepo *repo.NewsRawRepo,
	eventRepo *repo.EventRepo,
	tickerRepo *repo.TickerRepo,
	themeRepo *repo.ThemeRepo,
	relationRepo *repo.RelationRepo,
	aiService *AIService,
	cfg *config.Config,
) *NewsProcessor {
	return &NewsProcessor{
		newsRawRepo:  newsRawRepo,
		eventRepo:    eventRepo,
		tickerRepo:   tickerRepo,
		themeRepo:    themeRepo,
		relationRepo: relationRepo,
		aiService:    aiService,
		tenantID:     cfg.RSS.TenantID,
	}
}

// ProcessUnprocessed 处理未处理新闻。
// 由 RSSScheduler 在每次 FetchAll 之后调用。
func (p *NewsProcessor) ProcessUnprocessed(ctx context.Context) {
	if p.tenantID == "" {
		log.Println("[NewsProcessor] tenantID is empty, skip processing")
		return
	}
	if ctx.Err() != nil {
		return
	}
	p.processForTenant(ctx, p.tenantID)
}

// processForTenant 处理单个租户的未处理新闻。
// 先加载该租户所有股票构建 symbolMap（代码→Ticker），然后逐条处理。
func (p *NewsProcessor) processForTenant(ctx context.Context, tenantID string) {
	// 加载该租户所有股票，用于后续在新闻正文中匹配股票代码
	tickers, err := p.tickerRepo.GetAllByTenant(ctx, tenantID)
	if err != nil {
		log.Printf("[NewsProcessor] load tickers error: %v", err)
		return
	}
	if len(tickers) == 0 {
		return // 没有股票数据，无法匹配，跳过
	}

	// 构建股票代码索引：AAPL → Ticker, TSLA → Ticker（key 统一转大写）
	symbolMap := make(map[string]dm.Ticker, len(tickers))
	for _, t := range tickers {
		symbolMap[strings.ToUpper(t.Symbol)] = t
	}

	// 获取未处理的 news_raw（processed=false），每次最多 100 条
	newsList, err := p.newsRawRepo.GetUnprocessed(ctx, tenantID, 100)
	if err != nil {
		log.Printf("[NewsProcessor] get unprocessed error: %v", err)
		return
	}

	// 逐条处理新闻
	processed := 0
	for _, news := range newsList {
		if ctx.Err() != nil {
			return
		}
		if p.processOne(ctx, news, symbolMap, tenantID) {
			processed++
		}
	}

	if processed > 0 {
		log.Printf("[NewsProcessor] processed %d/%d news items", processed, len(newsList))
	}
}

// processOne 处理单条新闻，核心逻辑：
//  1. 在标题+正文中查找股票代码匹配
//  2. 匹配成功 → AI 评估重要度 → 创建事件
//  3. 建立事件-股票关联，增加股票提及次数
//  4. 通过关联股票反查最匹配的主题，将事件归入该主题
//  5. 标记为已处理
func (p *NewsProcessor) processOne(ctx context.Context, news dm.NewsRaw, symbolMap map[string]dm.Ticker, tenantID string) bool {
	// 第一步：在新闻标题+正文中查找匹配的股票代码
	matched := findTickerMatches(news.Title, news.Content, symbolMap)

	if len(matched) > 0 {
		// 第二步：调用 AI 评估事件重要度（1-5 分），AI 不可用时默认 3
		importance := 3
		if p.aiService != nil {
			importance = p.aiService.EvaluateImportance(ctx, news.Title, news.Content)
		}

		// 第三步：创建 Event 记录
		channel := news.Channel
		if channel == "" {
			channel = constant.ChannelGlobalMacro
		}
		event := &dm.Event{
			Title:      news.Title,
			Summary:    truncate(news.Content, 1000), // 截断到 1000 字符，避免摘要过长
			Channel:    channel,
			Importance: importance,
			EventTime:  news.PublishedAt,
			TenantID:   tenantID,
		}

		if err := p.eventRepo.Create(ctx, event); err != nil {
			log.Printf("[NewsProcessor] create event error: %v", err)
			return false
		}

		// 第四步：建立事件-股票关联，并递增股票的提及次数（用于热度评分）
		for _, ticker := range matched {
			if err := p.relationRepo.AddEventTicker(ctx, event.ID, ticker.ID); err != nil {
				log.Printf("[NewsProcessor] add event_ticker error: %v", err)
			}
			if err := p.tickerRepo.IncrementMentionCount(ctx, ticker.ID); err != nil {
				log.Printf("[NewsProcessor] increment mention count for ticker %s error: %v", ticker.ID, err)
			}
		}

		// 第五步：通过关联的股票反查主题，将事件归入匹配度最高的主题
		if theme := p.findBestTheme(ctx, matched); theme != nil {
			event.ThemeID = theme.ID
			event.ThemeName = theme.Name
			p.eventRepo.Update(ctx, event)              // 回写 theme_id
			p.themeRepo.UpdateEventCount(ctx, theme.ID) // 更新主题的事件计数
		}
	}

	// 无论是否匹配到股票，都标记为已处理（避免下次重复处理）
	if err := p.newsRawRepo.MarkProcessed(ctx, news.ID); err != nil {
		log.Printf("[NewsProcessor] mark processed error: %v", err)
		return false
	}
	return true
}

// findBestTheme 通过关联的股票反查最匹配的投资主题。
// 算法：遍历每只匹配的股票，找到它所属的所有主题，统计每只股票关联了同一主题的数量。
// 选择关联股票数最多的主题作为最佳匹配。
func (p *NewsProcessor) findBestTheme(ctx context.Context, matched []dm.Ticker) *dm.Theme {
	// 统计每只匹配的股票关联了哪些主题，以及每个主题被多少只股票关联
	themeCounts := make(map[string]int) // themeID → 关联的股票数量

	for _, ticker := range matched {
		// 查询该股票关联的所有主题（通过 theme_tickers 中间表）
		relations, err := p.relationRepo.GetThemesByTickerID(ctx, ticker.ID)
		if err != nil {
			continue
		}
		for _, rel := range relations {
			themeCounts[rel.ThemeID]++
		}
	}

	if len(themeCounts) == 0 {
		return nil // 没有股票关联到任何主题
	}

	// 选择关联股票数最多的主题（最多股票指向的主题最可能是正确归类）
	var bestID string
	var bestCount int
	for id, count := range themeCounts {
		if count > bestCount {
			bestCount = count
			bestID = id
		}
	}

	theme, err := p.themeRepo.FindByID(ctx, bestID)
	if err != nil {
		return nil
	}
	return theme
}

// findTickerMatches 在新闻标题+正文中查找匹配的股票代码。
// 使用分词器将文本拆分为 token，与 symbolMap 中的股票代码逐一比对。
// 返回去重后的匹配 Ticker 列表。
func findTickerMatches(title, content string, symbolMap map[string]dm.Ticker) []dm.Ticker {
	seen := make(map[string]bool)
	var result []dm.Ticker

	text := title + " " + content
	words := tokenize(text)

	for _, w := range words {
		upper := strings.ToUpper(w)
		if seen[upper] {
			continue
		}
		if ticker, ok := symbolMap[upper]; ok {
			seen[upper] = true
			result = append(result, ticker)
		}
	}
	return result
}

// tokenize 将文本拆分为可能的 token（字母数字组合）。
// 用于在新闻正文中匹配股票代码。
func tokenize(text string) []string {
	var tokens []string
	var buf strings.Builder
	for _, r := range text {
		if isTokenChar(r) {
			buf.WriteRune(r)
		} else {
			if buf.Len() > 0 {
				tokens = append(tokens, buf.String())
				buf.Reset()
			}
		}
	}
	if buf.Len() > 0 {
		tokens = append(tokens, buf.String())
	}
	return tokens
}

// isTokenChar 判断字符是否为 token 的合法组成部分（字母、数字、点、连字符）
func isTokenChar(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '-'
}
