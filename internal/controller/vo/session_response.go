package vo

import (
	"time"
)

// SessionResponse 会话响应结构体
type SessionResponse struct {
	SessionID  string            `json:"session_id"`
	DeviceInfo SessionDeviceInfo `json:"device_info"`
	IsActive   bool              `json:"is_active"`
	CreatedAt  time.Time         `json:"created_at"`
	LastActive time.Time         `json:"last_active"`
	ExpiresAt  time.Time         `json:"expires_at"`
}

// SessionDeviceInfo 会话设备信息
type SessionDeviceInfo struct {
	Platform  string `json:"platform"`
	Browser   string `json:"browser"`
	Device    string `json:"device"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}
