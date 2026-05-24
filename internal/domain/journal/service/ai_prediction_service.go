package service

import (
	"context"
	"sync"
	"time"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/journal/constant"
	"auth-perm/internal/domain/journal/dm"
	"auth-perm/internal/domain/journal/dto"
	"auth-perm/internal/domain/journal/repo"
	"auth-perm/internal/domain/journal/vo"
	"auth-perm/internal/infra/opencode"
)

type AIPredictionService struct {
	openCodeClient *opencode.Client
	predictionRepo *repo.AIPredictionRepo
	quotaRepo     *repo.AICallQuotaRepo
}

func NewAIPredictionService(oc *opencode.Client, pr *repo.AIPredictionRepo, qr *repo.AICallQuotaRepo) *AIPredictionService {
	return &AIPredictionService{openCodeClient: oc, predictionRepo: pr, quotaRepo: qr}
}

func (s *AIPredictionService) CreatePrediction(ctx context.Context, tenantID, accountID string, req *vo.CreateAIPredictionRequest) (*vo.AIPredictionResponse, error) {
	if !s.openCodeClient.Enabled() {
		return nil, errors.NewBusinessErrorF("AI 预测服务未配置，请设置 OPENCODE_API_KEY")
	}

	models := req.Models
	if len(models) == 0 {
		models = constant.DefaultModelIDs
	}
	if len(models) == 0 {
		return nil, errors.NewValidationError("至少需要指定一个模型")
	}
	if len(models) > 5 {
		return nil, errors.NewValidationError("单次请求最多支持 5 个模型")
	}
	for _, m := range models {
		if !constant.IsValidModel(m) {
			return nil, errors.NewValidationError("不支持的模型: " + m)
		}
	}

	today := time.Now().Truncate(24 * time.Hour)
	quotas, err := s.quotaRepo.GetQuotas(ctx, tenantID, accountID, today)
	if err != nil {
		return nil, errors.NewInternalErrorWithDetails("查询配额失败", err.Error(), err)
	}

	var exceeded []string
	for _, m := range models {
		if count, ok := quotas[m]; ok && count >= constant.DailyCallLimitPerModel {
			exceeded = append(exceeded, constant.GetModelName(m))
		}
	}
	if len(exceeded) > 0 {
		return nil, errors.NewBusinessErrorF("以下模型今日调用次数已达上限: %v（每日限 %d 次）", exceeded, constant.DailyCallLimitPerModel)
	}

	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = constant.DefaultSystemPrompt
	}

	if err := req.ValidateReasoningMode(); err != nil {
		return nil, err
	}
	reasoningMode := req.ReasoningMode

	type modelResult struct {
		modelID    string
		modelName  string
		content    string
		durationMs int64
		err        error
	}

	resultCh := make(chan modelResult, len(models))
	var wg sync.WaitGroup

	for _, modelID := range models {
		wg.Add(1)
		mID := modelID
		mName := constant.GetModelName(mID)
		go func() {
			defer wg.Done()
			start := time.Now()
			useMessages := constant.IsMiniMax(mID)
			content, callErr := s.openCodeClient.Chat(ctx, mID, systemPrompt, req.Question, useMessages, reasoningMode)
			elapsed := time.Since(start).Milliseconds()
			if callErr != nil {
				resultCh <- modelResult{modelID: mID, modelName: mName, err: callErr, durationMs: elapsed}
				return
			}
			resultCh <- modelResult{modelID: mID, modelName: mName, content: content, durationMs: elapsed}
		}()
	}

	wg.Wait()
	close(resultCh)

	resultMap := make(map[string]dm.PredictionResult, len(models))
	for r := range resultCh {
		pr := dm.PredictionResult{
			ModelID:    r.modelID,
			ModelName:  r.modelName,
			Content:    r.content,
			DurationMs: r.durationMs,
		}
		if r.err != nil {
			pr.Error = r.err.Error()
		}
		resultMap[r.modelID] = pr
	}

	results := make(dm.PredictionResults, 0, len(models))
	for _, mID := range models {
		if pr, ok := resultMap[mID]; ok {
			results = append(results, pr)
		}
	}

	// 配额软限制说明：并发请求可能短暂超出限制，但在可接受范围内。
	// 如需硬限制，需将外部 API 调用也纳入数据库事务，代价过高。
	for _, m := range models {
		if pr, ok := resultMap[m]; ok && pr.Error == "" {
			if _, quotaErr := s.quotaRepo.IncrementOrCreate(ctx, tenantID, accountID, m, today); quotaErr != nil {
				// log but don't fail the prediction
				_ = quotaErr
			}
		}
	}

	modelSnapshot := make(dm.ModelSnapshot, len(models))
	copy(modelSnapshot, models)

	prediction := dm.NewAIPrediction(tenantID, accountID, req.Question, systemPrompt, reasoningMode, results, modelSnapshot)
	if err := s.predictionRepo.Create(ctx, prediction); err != nil {
		return nil, err
	}

	voResults := make([]*vo.AIPredictionResultVO, 0, len(results))
	for _, r := range results {
		voResults = append(voResults, &vo.AIPredictionResultVO{
			ModelID:    r.ModelID,
			ModelName:  r.ModelName,
			Content:    r.Content,
			DurationMs: r.DurationMs,
			Error:      r.Error,
		})
	}

	return &vo.AIPredictionResponse{
		Results:      voResults,
		PredictionID: prediction.ID,
	}, nil
}

func (s *AIPredictionService) ListPredictions(ctx context.Context, tenantID, accountID string, page, pageSize int) (*dto.AIPredictionHistoryListResult, error) {
	list, total, err := s.predictionRepo.List(ctx, tenantID, accountID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return dto.FromAIPredictionHistoryDOList(list, page, pageSize, total), nil
}

func (s *AIPredictionService) GetPrediction(ctx context.Context, id, accountID, tenantID string) (*dto.AIPredictionDTO, error) {
	p, err := s.predictionRepo.FindByID(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	return dto.FromAIPredictionDO(p, true), nil
}

func (s *AIPredictionService) GetQuotas(ctx context.Context, tenantID, accountID string) (map[string]int, error) {
	today := time.Now().Truncate(24 * time.Hour)
	quotas, err := s.quotaRepo.GetQuotas(ctx, tenantID, accountID, today)
	if err != nil {
		return nil, err
	}
	result := make(map[string]int, len(constant.Models))
	for _, m := range constant.Models {
		remaining := constant.DailyCallLimitPerModel - quotas[m.ID]
		if remaining < 0 {
			remaining = 0
		}
		result[m.ID] = remaining
	}
	return result, nil
}