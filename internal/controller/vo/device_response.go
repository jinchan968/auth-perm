package vo

import (
	"time"
)

// DeviceResponse 设备响应结构体
type DeviceResponse struct {
	DeviceID   string    `json:"device_id"`
	Platform   string    `json:"platform"`
	Browser    string    `json:"browser"`
	Device     string    `json:"device"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	SessionID  string    `json:"session_id"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
	ExpiresAt  time.Time `json:"expires_at"`
	Trusted    bool      `json:"trusted"`
}
