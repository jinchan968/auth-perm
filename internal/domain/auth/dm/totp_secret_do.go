package dm

import (
	"time"

	"github.com/google/uuid"
)

// TOTPSecretDO TOTP密钥数据模型
type TOTPSecretDO struct {
	ID          string   `gorm:"primaryKey"`
	AccountID   string   `gorm:"column:account_id;index;not null"`
	Secret      string   `gorm:"column:secret;size:255;not null"`
	Algorithm   string   `gorm:"column:algorithm;size:20;default:'SHA1'"`
	Digits      int      `gorm:"column:digits;default:6"`
	Period      int      `gorm:"column:period;default:30"`
	IsEnabled   bool     `gorm:"column:is_enabled;default:false"`
	BackupCodes []string `gorm:"column:backup_codes;type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// TableName 指定表名
func (t *TOTPSecretDO) TableName() string {
	return "totp_secrets"
}

// NewTOTPSecret 创建新的TOTP密钥
func NewTOTPSecret(accountID, secret string) *TOTPSecretDO {
	now := time.Now()
	return &TOTPSecretDO{
		ID:        uuid.New().String(),
		AccountID: accountID,
		Secret:    secret,
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
		IsEnabled: false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
