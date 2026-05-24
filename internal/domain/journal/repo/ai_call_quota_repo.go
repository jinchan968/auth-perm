package repo

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type AICallQuotaRepo struct {
	db *gorm.DB
}

func NewAICallQuotaRepo(db *gorm.DB) *AICallQuotaRepo {
	return &AICallQuotaRepo{db: db}
}

func (r *AICallQuotaRepo) IncrementOrCreate(ctx context.Context, tenantID, accountID, modelID string, date time.Time) (int, error) {
	var count int
	sql := `INSERT INTO ai_call_quotas (id, tenant_id, account_id, model_id, call_date, call_count, created_at, updated_at)
		VALUES (gen_random_uuid(), ?, ?, ?, ?, 1, NOW(), NOW())
		ON CONFLICT ON CONSTRAINT uk_quota_tenant_account_model_date
		DO UPDATE SET call_count = ai_call_quotas.call_count + 1, updated_at = NOW()
		RETURNING call_count`
	if err := r.db.WithContext(ctx).Raw(sql, tenantID, accountID, modelID, date).Scan(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *AICallQuotaRepo) GetQuotas(ctx context.Context, tenantID, accountID string, date time.Time) (map[string]int, error) {
	type row struct {
		ModelID   string `gorm:"column:model_id"`
		CallCount int    `gorm:"column:call_count"`
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Table("ai_call_quotas").
		Where("tenant_id = ? AND account_id = ? AND call_date = ?", tenantID, accountID, date).
		Select("model_id, call_count").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]int, len(rows))
	for _, q := range rows {
		result[q.ModelID] = q.CallCount
	}
	return result, nil
}