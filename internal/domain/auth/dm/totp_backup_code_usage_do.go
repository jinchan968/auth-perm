package dm

import (
	"time"

	"github.com/google/uuid"
)

// TOTPBackupCodeUsageDO 备份码使用记录
type TOTPBackupCodeUsageDO struct {
	ID        string    `gorm:"primaryKey"`
	AccountID string    `gorm:"column:account_id;index;not null"`
	Code      string    `gorm:"column:code;size:36;not null"`
	UsedAt    time.Time `gorm:"column:used_at"`
	IPAddress string    `gorm:"column:ip_address;size:45"`
	UserAgent string    `gorm:"column:user_agent;type:text"`
	Success   bool      `gorm:"column:success;default:true"`
	CreatedAt time.Time
}

// TableName 指定表名
func (u *TOTPBackupCodeUsageDO) TableName() string {
	return "totp_backup_code_usage"
}

// NewTOTPBackupCodeUsage 创建新的备份码使用记录
func NewTOTPBackupCodeUsage(accountID, code, ipAddress, userAgent string, success bool) *TOTPBackupCodeUsageDO {
	return &TOTPBackupCodeUsageDO{
		ID:        uuid.New().String(),
		AccountID: accountID,
		Code:      code,
		UsedAt:    time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
		CreatedAt: time.Now(),
	}
}
