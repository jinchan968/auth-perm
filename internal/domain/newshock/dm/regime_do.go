package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Regime 市场环境判断，由 AI 定期分析主题趋势和事件后生成。
// regime_type: risk_on(风险偏好) / risk_off(风险回避) / neutral(中性)
// confidence: AI 对判断的置信度 0.0-1.0
type Regime struct {
	ID         string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	RegimeType string         `gorm:"column:regime_type;type:varchar(20);not null;default:neutral" json:"regime_type"`
	Confidence float64        `gorm:"column:confidence;default:0.5" json:"confidence"`
	Summary    string         `gorm:"column:summary;type:text" json:"summary"`
	TenantID   string         `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	CreatedAt  time.Time      `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Regime) TableName() string { return "newshock_regime" }

func (r *Regime) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}
