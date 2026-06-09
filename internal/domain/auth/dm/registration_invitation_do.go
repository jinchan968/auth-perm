package dm

import (
	"time"

	authConstant "auth-perm/internal/domain/auth/constant"

	"github.com/google/uuid"
)

// RegistrationInvitationDO 注册邀请码领域对象
type RegistrationInvitationDO struct {
	ID                     string     `gorm:"primaryKey;type:varchar(36)"`
	CodeHash               string     `gorm:"column:code_hash;type:varchar(128);not null;uniqueIndex"`
	CodePreview            string     `gorm:"column:code_preview;type:varchar(32);not null"`
	TenantID               string     `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	Status                 string     `gorm:"column:status;type:varchar(32);not null;default:active;index"`
	ExpiresAt              time.Time  `gorm:"column:expires_at;not null;index"`
	UsedAt                 *time.Time `gorm:"column:used_at"`
	UsedByAccountID        *string    `gorm:"column:used_by_account_id;type:varchar(36);index"`
	CreatedByAccountID     string     `gorm:"column:created_by_account_id;type:varchar(36);not null;index"`
	InvalidatedAt          *time.Time `gorm:"column:invalidated_at"`
	InvalidatedByAccountID *string    `gorm:"column:invalidated_by_account_id;type:varchar(36)"`
	CreatedAt              time.Time  `gorm:"column:created_at"`
	UpdatedAt              time.Time  `gorm:"column:updated_at"`
}

func NewRegistrationInvitation(codeHash, codePreview, tenantID, createdByAccountID string, expiresAt time.Time) *RegistrationInvitationDO {
	now := time.Now()
	return &RegistrationInvitationDO{
		ID:                 uuid.New().String(),
		CodeHash:           codeHash,
		CodePreview:        codePreview,
		TenantID:           tenantID,
			Status:             authConstant.InvitationStatusActive,
		ExpiresAt:          expiresAt,
		CreatedByAccountID: createdByAccountID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func (i *RegistrationInvitationDO) TableName() string {
	return "registration_invitations"
}
