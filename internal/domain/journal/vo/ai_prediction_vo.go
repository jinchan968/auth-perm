package vo

import "auth-perm/internal/common/errors"

var validReasoningModes = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
}

type CreateAIPredictionRequest struct {
	Question      string   `json:"question" binding:"required,max=4000"`
	SystemPrompt  string   `json:"system_prompt"`
	Models        []string `json:"models" binding:"max=5"`
	ReasoningMode string   `json:"reasoning_mode"`
}

func (r *CreateAIPredictionRequest) ValidateReasoningMode() error {
	if r.ReasoningMode == "" {
		r.ReasoningMode = "low"
		return nil
	}
	if !validReasoningModes[r.ReasoningMode] {
		return errors.NewValidationError("不支持的推理模式: " + r.ReasoningMode + "，可选值: low, medium, high")
	}
	return nil
}

type AIPredictionResultVO struct {
	ModelID    string `json:"model_id"`
	ModelName  string `json:"model_name"`
	Content    string `json:"content"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

type AIPredictionResponse struct {
	Results      []*AIPredictionResultVO `json:"results"`
	PredictionID string                  `json:"prediction_id"`
}

type AIModelVO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ListModelsResponse struct {
	Defaults           []string     `json:"defaults"`
	Replaceable        []AIModelVO  `json:"replaceable"`
	All                []AIModelVO  `json:"all"`
	DefaultSystemPrompt string      `json:"default_system_prompt"`
	DailyLimit         int          `json:"daily_limit"`
}


