package dto

import (
	"time"
)

// DeviceTrustDTO 设备信任数据传输对象
type DeviceTrustDTO struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	DeviceID   string    `json:"device_id"`
	DeviceInfo string    `json:"device_info"`
	Trusted    bool      `json:"trusted"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NewDeviceTrustDTO 创建设备信任DTO
func NewDeviceTrustDTO(userID, deviceID, deviceInfo string) *DeviceTrustDTO {
	return &DeviceTrustDTO{
		UserID:     userID,
		DeviceID:   deviceID,
		DeviceInfo: deviceInfo,
		Trusted:    true,
		Reason:     "用户主动信任",
	}
}

// ToResponse 转换为响应格式
func (d *DeviceTrustDTO) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":          d.ID,
		"user_id":     d.UserID,
		"device_id":   d.DeviceID,
		"device_info": d.DeviceInfo,
		"trusted":     d.Trusted,
		"reason":      d.Reason,
		"created_at":  d.CreatedAt,
		"updated_at":  d.UpdatedAt,
	}
}
