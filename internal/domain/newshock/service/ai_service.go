// Package service AI 分析服务，负责调用 LLM 执行三项分析任务：
//  1. 事件重要度评估（EvaluateImportance）：给新闻打 1-5 分
//  2. 主题描述生成（GenerateThemeDescription）：为主题写 2-3 句话的叙事概括
//  3. 市场环境判断（JudgeRegime）：判断 risk_on/risk_off/neutral
//
// 所有 AI 调用都有超时保护和降级策略：LLM 不可用时返回默认值，不会阻塞管线。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"auth-perm/internal/domain/newshock/constant"
	"auth-perm/internal/domain/newshock/dm"
	"auth-perm/internal/domain/newshock/repo"
	"auth-perm/internal/infra/llm"
)

type AIService struct {
	llmClient  *llm.Client
	themeRepo  *repo.ThemeRepo
	eventRepo  *repo.EventRepo
	regimeRepo *repo.RegimeRepo
}

func NewAIService(
	llmClient *llm.Client,
	themeRepo *repo.ThemeRepo,
	eventRepo *repo.EventRepo,
	regimeRepo *repo.RegimeRepo,
	_ string,
) *AIService {
	return &AIService{
		llmClient:  llmClient,
		themeRepo:  themeRepo,
		eventRepo:  eventRepo,
		regimeRepo: regimeRepo,
	}
}

// EvaluateImportance 评估事件重要度（1-5 分）。
// LLM 不可用时默认返回 3（中等重要）。
// 评分标准：
//
//	1=极低（日常琐事）
//	2=低（一般新闻）
//	3=中（值得关注）
//	4=高（重大事件）
//	5=极高（市场转折点）
//
// 超时 15 秒，避免阻塞新闻处理管线。
func (s *AIService) EvaluateImportance(ctx context.Context, title, summary string) int {
	// LLM 不可用时直接返回默认值 3（中等重要），不阻塞管线
	if !s.llmClient.Enabled() {
		return 3
	}

	// 设置 15 秒超时，避免 LLM 响应慢阻塞整个新闻处理管线
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// 构建 prompt：要求 LLM 只返回 JSON 格式的评分结果
	prompt := fmt.Sprintf(`标题: %s
摘要: %s

请评估此金融事件的重要程度，只返回 JSON: {"importance": 1-5}
1=极低(日常琐事), 2=低(一般新闻), 3=中(值得关注), 4=高(重大事件), 5=极高(市场转折点)`, title, summary)

	// 调用 LLM，system prompt 设定角色为金融事件分析师
	resp, err := s.llmClient.Chat(ctx, "你是金融事件分析师，评估事件对市场的影响程度。", prompt)
	if err != nil {
		log.Printf("[AIService] evaluate importance error: %v", err)
		return 3 // 调用失败，降级返回默认值
	}

	// 解析 LLM 返回的 JSON，提取 importance 字段
	var result struct {
		Importance int `json:"importance"`
	}
	if err := json.Unmarshal(extractJSON(resp), &result); err != nil || result.Importance < 1 || result.Importance > 5 {
		return 3 // 解析失败或值越界，降级返回默认值
	}
	return result.Importance
}

// GenerateThemeDescription 为投资主题生成 AI 描述。
// 会读取该主题的近期事件标题，让 AI 概括核心叙事和当前状态。
// LLM 不可用时返回原有描述不变。超时 30 秒。
func (s *AIService) GenerateThemeDescription(ctx context.Context, theme *dm.Theme) string {
	// LLM 不可用时返回原有描述不变
	if !s.llmClient.Enabled() {
		return theme.Description
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 获取该主题的近期事件（最多 10 条），作为 AI 生成描述的素材
	events, err := s.eventRepo.GetByThemeID(ctx, theme.ID, 10)
	if err != nil {
		return theme.Description
	}

	// 将事件标题格式化为列表，作为 prompt 的上下文
	var eventTitles []string
	for _, e := range events {
		eventTitles = append(eventTitles, "- "+e.Title)
	}

	// 构建 prompt：提供主题名称、分类和近期事件，要求 AI 概括核心叙事
	prompt := fmt.Sprintf(`主题: %s
分类: %s
近期事件:
%s

请用 2-3 句话概括这个投资主题的核心叙事和当前状态。`, theme.Name, theme.Category, strings.Join(eventTitles, "\n"))

	resp, err := s.llmClient.Chat(ctx, "你是投资主题分析师，用简洁的语言概括主题叙事。", prompt)
	if err != nil {
		log.Printf("[AIService] generate theme description error: %v", err)
		return theme.Description // 失败时返回原有描述
	}
	return strings.TrimSpace(resp)
}

// JudgeRegime 判断当前市场环境（risk_on/risk_off/neutral）。
// 分析输入：
//   - 当前主题趋势分布（上升/下降/稳定各多少）
//   - 上升和下降主题的名称
//   - 近期 20 条事件标题
//
// 输出：写入 Regime 表，包含环境类型、置信度和一句话总结。
// 超时 60 秒。由 ScoringScheduler 每 60 分钟调用一次。
func (s *AIService) JudgeRegime(ctx context.Context, tenantID string) {
	// 60 秒超时，市场环境判断需要较多上下文，给更多时间
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// 获取该租户的 Top 20 主题，用于分析趋势分布
	themes, _, err := s.themeRepo.List(ctx, repo.ThemeQueryParams{
		TenantID: tenantID,
		PageSize: 20,
	})
	if err != nil {
		log.Printf("[AIService] list themes error: %v", err)
		return
	}

	// 统计上升/下降/稳定主题的数量和名称，作为 AI 判断的输入
	var rising, declining, stable int
	var risingNames, decliningNames []string
	for _, t := range themes {
		switch t.Trend {
		case constant.TrendRising:
			rising++
			risingNames = append(risingNames, t.Name)
		case constant.TrendDeclining:
			declining++
			decliningNames = append(decliningNames, t.Name)
		default:
			stable++
		}
	}

	// 获取最新 20 条事件标题，作为市场情绪的参考
	events, err := s.eventRepo.GetRecentEvents(ctx, tenantID, 20)
	if err != nil {
		log.Printf("[AIService] get recent events error: %v", err)
		return
	}
	var eventLines []string
	for _, e := range events {
		eventLines = append(eventLines, "- "+e.Title)
	}

	// 构建 prompt：提供主题趋势分布和近期事件，要求 AI 判断市场环境
	prompt := fmt.Sprintf(`当前主题趋势:
- 上升主题(%d): %s
- 下降主题(%d): %s
- 稳定主题(%d)

近期事件:
%s

请判断当前市场环境，只返回 JSON: {"regime": "risk_on"|"risk_off"|"neutral", "confidence": 0.0-1.0, "summary": "一句话总结"}`,
		rising, strings.Join(risingNames, ", "),
		declining, strings.Join(decliningNames, ", "),
		stable,
		strings.Join(eventLines, "\n"))

	resp, err := s.llmClient.Chat(ctx, "你是宏观策略分析师，根据主题趋势和事件判断市场环境。", prompt)
	if err != nil {
		log.Printf("[AIService] judge regime error: %v", err)
		return
	}

	// 解析 LLM 返回的 JSON
	var result struct {
		Regime     string  `json:"regime"`
		Confidence float64 `json:"confidence"`
		Summary    string  `json:"summary"`
	}
	if err := json.Unmarshal(extractJSON(resp), &result); err != nil {
		log.Printf("[AIService] parse regime response error: %v", err)
		return
	}

	// 校验 regime 值是否合法（必须是 risk_on/risk_off/neutral 之一）
	if result.Regime != constant.RegimeRiskOn && result.Regime != constant.RegimeRiskOff && result.Regime != constant.RegimeNeutral {
		return
	}
	// 校验置信度范围
	if result.Confidence < 0 || result.Confidence > 1 {
		result.Confidence = 0.5
	}

	// 写入 Regime 表（每次创建新记录，GetLatest 取最新一条展示）
	regime := &dm.Regime{
		RegimeType: result.Regime,
		Confidence: result.Confidence,
		Summary:    result.Summary,
		TenantID:   tenantID,
	}
	if err := s.regimeRepo.Create(ctx, regime); err != nil {
		log.Printf("[AIService] create regime error: %v", err)
	}
}

// extractJSON 从 LLM 回复中提取 JSON 部分。
// LLM 回复可能包含解释文字，此函数找到第一个 { 和最后一个 } 之间的内容。
func extractJSON(s string) []byte {
	if idx := strings.Index(s, "{"); idx >= 0 {
		end := strings.LastIndex(s, "}")
		if end > idx {
			return []byte(s[idx : end+1])
		}
	}
	return []byte(s)
}
