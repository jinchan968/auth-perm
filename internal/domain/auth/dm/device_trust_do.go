package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeviceTrustDO 设备信任领域对象
type DeviceTrustDO struct {
	ID         string         `gorm:"primaryKey;type:uuid"`
	TenantID   string         `gorm:"column:tenant_id;type:uuid;not null;index"`
	UserID     string         `gorm:"column:user_id;type:uuid;not null;index"`
	DeviceID   string         `gorm:"column:device_id;type:varchar(255);not null;index"`
	DeviceInfo string         `gorm:"column:device_info;type:jsonb;default:'{}'"`
	Trusted    bool           `gorm:"column:trusted;default:true;not null;index"`
	Reason     string         `gorm:"column:reason;type:text"`
	CreatedAt  time.Time      `gorm:"column:created_at;not null;index"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	// 关联关系
	User UserDO `gorm:"foreignKey:UserID"`
}

// NewDeviceTrust 创建设备信任记录
func NewDeviceTrust(tenantID, userID, deviceID, deviceInfo string) *DeviceTrustDO {
	now := time.Now()
	return &DeviceTrustDO{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		UserID:     userID,
		DeviceID:   deviceID,
		DeviceInfo: deviceInfo,
		Trusted:    true,
		Reason:     "用户主动信任",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// SetTrustReason 设置信任原因
func (d *DeviceTrustDO) SetTrustReason(reason string) {
	d.Reason = reason
	d.UpdatedAt = time.Now()
}

// SetTrusted 设置信任状态
func (d *DeviceTrustDO) SetTrusted(trusted bool) {
	d.Trusted = trusted
	d.UpdatedAt = time.Now()
}

// BeforeCreate GORM钩子
func (d *DeviceTrustDO) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate GORM钩子
func (d *DeviceTrustDO) BeforeUpdate(tx *gorm.DB) error {
	d.UpdatedAt = time.Now()
	return nil
}

// TableName 指定表名
func (d *DeviceTrustDO) TableName() string {
	return "device_trusts"
}
