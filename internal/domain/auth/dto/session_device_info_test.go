package dto

import (
	"testing"
	"time"
)

func TestSessionDTO_GetDeviceInfo(t *testing.T) {
	// 创建测试用的SessionDTO
	session := &SessionDTO{
		ID:        "test-session",
		UserID:    "test-user",
		AccountID: "test-account",
		DeviceInfo: DeviceInfoDTO{
			Platform:    "macOS",
			Browser:     "Chrome",
			Device:      "desktop",
			IPAddress:   "192.168.1.1",
			UserAgent:   "Mozilla/5.0",
			Fingerprint: "abc123",
		},
		ExpiresAt:    time.Now().Add(time.Hour),
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	// 测试获取设备信息
	deviceInfo := session.GetDeviceInfo()

	// 验证字段是否正确映射
	if deviceInfo.Platform != "macOS" {
		t.Errorf("Expected platform 'macOS', got '%s'", deviceInfo.Platform)
	}
	if deviceInfo.Browser != "Chrome" {
		t.Errorf("Expected browser 'Chrome', got '%s'", deviceInfo.Browser)
	}
	if deviceInfo.Device != "desktop" {
		t.Errorf("Expected device 'desktop', got '%s'", deviceInfo.Device)
	}
	if deviceInfo.IPAddress != "192.168.1.1" {
		t.Errorf("Expected ip_address '192.168.1.1', got '%s'", deviceInfo.IPAddress)
	}
	if deviceInfo.UserAgent != "Mozilla/5.0" {
		t.Errorf("Expected user_agent 'Mozilla/5.0', got '%s'", deviceInfo.UserAgent)
	}
	if deviceInfo.Fingerprint != "abc123" {
		t.Errorf("Expected fingerprint 'abc123', got '%s'", deviceInfo.Fingerprint)
	}
}

func TestSessionDTO_GetDeviceInfo_EmptyDeviceInfo(t *testing.T) {
	// 测试空的 DeviceInfo
	session := &SessionDTO{
		ID:           "test-session",
		UserID:       "test-user",
		AccountID:    "test-account",
		DeviceInfo:   DeviceInfoDTO{}, // 空的DeviceInfoDTO
		ExpiresAt:    time.Now().Add(time.Hour),
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	deviceInfo := session.GetDeviceInfo()
	if deviceInfo.Platform != "" {
		t.Errorf("Expected empty platform, got '%s'", deviceInfo.Platform)
	}
}

func TestSessionDTO_GetDeviceInfo_PartialDeviceInfo(t *testing.T) {
	// 测试部分填充的device_info
	session := &SessionDTO{
		ID:        "test-session",
		UserID:    "test-user",
		AccountID: "test-account",
		DeviceInfo: DeviceInfoDTO{
			Platform: "Windows",
			Browser:  "Firefox",
			// 其他字段为空
		},
		ExpiresAt:    time.Now().Add(time.Hour),
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	deviceInfo := session.GetDeviceInfo()
	if deviceInfo.Platform != "Windows" {
		t.Errorf("Expected platform 'Windows', got '%s'", deviceInfo.Platform)
	}
	if deviceInfo.Browser != "Firefox" {
		t.Errorf("Expected browser 'Firefox', got '%s'", deviceInfo.Browser)
	}
	if deviceInfo.Device != "" {
		t.Errorf("Expected empty device, got '%s'", deviceInfo.Device)
	}
}
