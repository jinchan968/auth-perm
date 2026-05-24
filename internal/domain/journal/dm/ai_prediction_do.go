package dm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PredictionResult struct {
	ModelID    string `json:"model_id"`
	ModelName  string `json:"model_name"`
	Content    string `json:"content"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

type PredictionResults []PredictionResult

func (r PredictionResults) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *PredictionResults) Scan(value interface{}) error {
	if value == nil {
		*r = PredictionResults{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("expected []byte or string for PredictionResults, got %T", value)
	}
	return json.Unmarshal(bytes, r)
}

type ModelSnapshot []string

func (m ModelSnapshot) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *ModelSnapshot) Scan(value interface{}) error {
	if value == nil {
		*m = ModelSnapshot{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("expected []byte or string for ModelSnapshot, got %T", value)
	}
	return json.Unmarshal(bytes, m)
}

type AIPredictionDO struct {
	ID            string           `gorm:"primaryKey;type:varchar(36)"`
	TenantID      string           `gorm:"column:tenant_id;type:varchar(64);not null;index"`
	AccountID     string           `gorm:"column:account_id;type:varchar(64);not null;index"`
	Question      string           `gorm:"column:question;type:text;not null"`
	SystemPrompt  string           `gorm:"column:system_prompt;type:text"`
	ReasoningMode string           `gorm:"column:reasoning_mode;type:varchar(16);not null;default:'low'"`
	Results       PredictionResults `gorm:"column:results;type:jsonb;not null;default:'[]'"`
	ModelSnapshot ModelSnapshot     `gorm:"column:model_snapshot;type:jsonb;not null;default:'[]'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (*AIPredictionDO) TableName() string { return "ai_predictions" }

func NewAIPrediction(tenantID, accountID, question, systemPrompt, reasoningMode string, results PredictionResults, modelSnapshot ModelSnapshot) *AIPredictionDO {
	now := time.Now()
	return &AIPredictionDO{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		AccountID:     accountID,
		Question:      question,
		SystemPrompt:  systemPrompt,
		ReasoningMode: reasoningMode,
		Results:       results,
		ModelSnapshot: modelSnapshot,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
