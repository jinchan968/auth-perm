package dm

import (
	"auth-perm/internal/common/utils"
	"auth-perm/internal/domain/auth/dto"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
)

// SessionDO 会话领域对象
type SessionDO struct {
	ID           string            `gorm:"primaryKey;type:uuid"`
	UserID       string            `gorm:"column:user_id;type:uuid;not null;index"`
	AccountID    string            `gorm:"column:account_id;type:uuid;not null;index"`
	TenantID     string            `gorm:"column:tenant_id;type:varchar(255);not null;index;default:'default'"` // 租户ID，支持多租户
	TokenHash    string            `gorm:"column:token_hash;type:varchar(255);not null;index"`
	DeviceInfo   dto.DeviceInfoDTO `gorm:"column:device_info;type:jsonb;default:'{}'"`
	IPAddress    *string           `gorm:"column:ip_address;type:inet"`
	UserAgent    *string           `gorm:"column:user_agent;type:text"`
	ExpiresAt    time.Time         `gorm:"column:expires_at;type:timestamptz;not null;index"`
	LastActiveAt time.Time         `gorm:"column:last_active_at;type:timestamptz;not null;index"`
	IsActive     bool              `gorm:"column:is_active;default:true;index"`
	CreatedAt    time.Time         `gorm:"column:created_at;type:timestamptz;not null;default:now()"`
	UpdatedAt    time.Time         `gorm:"column:updated_at;type:timestamptz;not null;default:now()"`
}

// NewSession 创建新会话（数据库操作）
func NewSession(userID, accountID, tenantID string, expiresAt time.Time) *SessionDO {
	now := time.Now()
	token := generateSessionToken()

	return &SessionDO{
		ID:           uuid.New().String(),
		UserID:       userID,
		AccountID:    accountID,
		TenantID:     tenantID,
		TokenHash:    utils.HashToken(token),
		DeviceInfo:   dto.DeviceInfoDTO{},
		ExpiresAt:    expiresAt,
		LastActiveAt: now,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// SetDeviceInfo 设置设备信息
func (s *SessionDO) SetDeviceInfo(deviceInfo *dto.DeviceInfoDTO) {
	s.DeviceInfo = *deviceInfo
}

// GetDeviceInfo 获取设备信息
func (s *SessionDO) GetDeviceInfo() *dto.DeviceInfoDTO {
	return &s.DeviceInfo
}

// ToDTO 转换为DTO
func (s *SessionDO) ToDTO() *dto.SessionDTO {
	if s == nil {
		return nil
	}
	return &dto.SessionDTO{
		ID:           s.ID,
		UserID:       s.UserID,
		AccountID:    s.AccountID,
		TokenHash:    s.TokenHash,
		DeviceInfo:   *s.GetDeviceInfo(),
		IPAddress:    s.IPAddress,
		UserAgent:    s.UserAgent,
		ExpiresAt:    s.ExpiresAt,
		LastActiveAt: s.LastActiveAt,
		IsActive:     s.IsActive,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

// SessionFromDTO 从DTO创建SessionDO
func SessionFromDTO(d *dto.SessionDTO) *SessionDO {
	if d == nil {
		return nil
	}

	return &SessionDO{
		ID:           d.ID,
		UserID:       d.UserID,
		AccountID:    d.AccountID,
		TokenHash:    d.TokenHash,
		DeviceInfo:   d.DeviceInfo,
		IPAddress:    d.IPAddress,
		UserAgent:    d.UserAgent,
		ExpiresAt:    d.ExpiresAt,
		LastActiveAt: d.LastActiveAt,
		IsActive:     d.IsActive,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

// TableName 指定表名
func (s *SessionDO) TableName() string {
	return "sessions"
}

// 私有工具方法

// generateSessionToken FUTURE: 会话令牌生成 - 在实现会话管理时使用
func generateSessionToken() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
