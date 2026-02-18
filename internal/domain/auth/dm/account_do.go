package dm

import (
	"auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dto"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountDO 账户领域对象
type AccountDO struct {
	ID                   string                 `gorm:"primaryKey;type:uuid"`
	UserID               string                 `gorm:"column:user_id;type:uuid;not null;index"`
	TenantID             string                 `gorm:"column:tenant_id;type:uuid;not null;index"`
	AccountType          constant.AccountType   `gorm:"column:account_type;not null"`
	OAuthID              string                 `gorm:"column:oauth_id;index"`
	OAuthProvider        string                 `gorm:"column:oauth_provider;index"`
	Status               constant.AccountStatus `gorm:"not null;default:active"`
	LastLoginAt          *time.Time             `gorm:"column:last_login_at"`
	UserAgent            string                 `gorm:"column:user_agent"`
	IPAddress            *string                `gorm:"column:ip_address;type:inet"`
	DeviceInfo           dto.DeviceInfoDTO      `gorm:"column:device_info;type:jsonb;default:'{}'"`
	EmailVerified        bool                   `gorm:"column:email_verified;default:false"`
	EmailVerifiedAt      *time.Time             `gorm:"column:email_verified_at"`
	ResetPasswordToken   *string                `gorm:"column:password_reset_token;index"`      // 密码重置令牌（哈希）
	ResetPasswordExpires *time.Time             `gorm:"column:password_reset_expires_at;index"` // 密码重置令牌过期时间
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            gorm.DeletedAt `gorm:"index"`

	// 关联关系
	User *UserDO `gorm:"foreignKey:UserID"`
}

func (a *AccountDO) TableName() string {
	return "accounts"
}

func (a *AccountDO) Exist() bool {
	return a != nil && a.ID != ""
}

// NewAccount 创建新账户（数据库操作）
func NewAccount(userID, tenantID string, accountType constant.AccountType) *AccountDO {
	now := time.Now()
	return &AccountDO{
		ID:          uuid.New().String(),
		UserID:      userID,
		TenantID:    tenantID,
		AccountType: accountType,
		Status:      constant.AccountStatusActive,
		DeviceInfo:  dto.DeviceInfoDTO{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ToDTO 转换为DTO
func (a *AccountDO) ToDTO() *dto.AccountDTO {
	if a == nil {
		return nil
	}
	return &dto.AccountDTO{
		ID:            a.ID,
		UserID:        a.UserID,
		TenantID:      a.TenantID,
		AccountType:   a.AccountType,
		OAuthID:       a.OAuthID,
		OAuthProvider: a.OAuthProvider,
		Status:        a.Status,
		UserAgent:     a.UserAgent,
		DeviceInfo:    *a.GetDeviceInfo(),
		EmailVerified: a.EmailVerified,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}

// AccountFromDTO 从DTO创建AccountDO
func AccountFromDTO(account *dto.AccountDTO) *AccountDO {
	if account == nil {
		return nil
	}

	return &AccountDO{
		ID:            account.ID,
		UserID:        account.UserID,
		TenantID:      account.TenantID,
		AccountType:   account.AccountType,
		OAuthID:       account.OAuthID,
		OAuthProvider: account.OAuthProvider,
		Status:        account.Status,
		UserAgent:     account.UserAgent,
		DeviceInfo:    account.DeviceInfo,
		EmailVerified: account.EmailVerified,
		CreatedAt:     account.CreatedAt,
		UpdatedAt:     account.UpdatedAt,
	}
}

// GetDeviceInfo 获取设备信息
func (a *AccountDO) GetDeviceInfo() *dto.DeviceInfoDTO {
	return &a.DeviceInfo
}

// SetDeviceInfo 设置设备信息
func (a *AccountDO) SetDeviceInfo(deviceInfo *dto.DeviceInfoDTO) {
	a.DeviceInfo = *deviceInfo
}
