package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JournalTemplateDO 札记模板领域对象
type JournalTemplateDO struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	TenantID  string    `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	AccountID string    `gorm:"column:account_id;type:varchar(36);not null;index"`
	Name      string    `gorm:"column:name;type:varchar(255);not null"`
	Content   *string   `gorm:"column:content;type:text"`
	Tags      []string  `gorm:"column:tags;type:jsonb;serializer:json"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (*JournalTemplateDO) TableName() string { return "journal_templates" }

// NewJournalTemplate 创建新模板
func NewJournalTemplate(tenantID, accountID, name string, content *string, tags []string) *JournalTemplateDO {
	return &JournalTemplateDO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		AccountID: accountID,
		Name:      name,
		Content:   content,
		Tags:      tags,
	}
}