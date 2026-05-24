package dto

import (
	"time"

	"auth-perm/internal/domain/journal/dm"
)

type AIPredictionResultDTO struct {
	ModelID    string `json:"model_id"`
	ModelName  string `json:"model_name"`
	Content    string `json:"content"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

type AIPredictionDTO struct {
	ID            string                  `json:"id"`
	Question      string                  `json:"question"`
	SystemPrompt  string                  `json:"system_prompt"`
	ReasoningMode string                  `json:"reasoning_mode"`
	Results       []*AIPredictionResultDTO `json:"results"`
	ModelSnapshot  []string               `json:"model_snapshot"`
	CreatedAt     string                  `json:"created_at"`
}

type AIPredictionHistoryItem struct {
	ID        string `json:"id"`
	Question  string `json:"question"`
	CreatedAt string `json:"created_at"`
}

type AIPredictionHistoryListResult struct {
	Data  []*AIPredictionHistoryItem `json:"data"`
	Total int64                      `json:"total"`
	Page  int                        `json:"page"`
	PageSize int                     `json:"page_size"`
}

func FromAIPredictionResultDO(r dm.PredictionResult) *AIPredictionResultDTO {
	return &AIPredictionResultDTO{
		ModelID:    r.ModelID,
		ModelName:  r.ModelName,
		Content:    r.Content,
		DurationMs: r.DurationMs,
		Error:      r.Error,
	}
}

func FromAIPredictionDO(d *dm.AIPredictionDO, withResults bool) *AIPredictionDTO {
	aid := &AIPredictionDTO{
		ID:            d.ID,
		Question:      d.Question,
		SystemPrompt:  d.SystemPrompt,
		ReasoningMode: d.ReasoningMode,
		ModelSnapshot: []string(d.ModelSnapshot),
		CreatedAt:     d.CreatedAt.Format(time.RFC3339),
	}
	if withResults {
		results := make([]*AIPredictionResultDTO, 0, len(d.Results))
		for _, r := range d.Results {
			results = append(results, FromAIPredictionResultDO(r))
		}
		aid.Results = results
	}
	return aid
}

func FromAIPredictionHistoryDOList(list []*dm.AIPredictionDO, page, pageSize int, total int64) *AIPredictionHistoryListResult {
	items := make([]*AIPredictionHistoryItem, 0, len(list))
	for _, d := range list {
		items = append(items, &AIPredictionHistoryItem{
			ID:        d.ID,
			Question:  d.Question,
			CreatedAt: d.CreatedAt.Format(time.RFC3339),
		})
	}
	return &AIPredictionHistoryListResult{
		Data:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}
