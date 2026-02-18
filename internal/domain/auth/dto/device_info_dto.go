package dto

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// DeviceInfoDTO 设备信息结构
type DeviceInfoDTO struct {
	Platform    string `json:"platform"`    // 操作系统
	Browser     string `json:"browser"`     // 浏览器
	Device      string `json:"device"`      // 设备类型
	IPAddress   string `json:"ip_address"`  // IP地址
	UserAgent   string `json:"user_agent"`  // 用户代理
	Fingerprint string `json:"fingerprint"` // 设备指纹
}

// Scan 实现sql.Scanner接口
func (d *DeviceInfoDTO) Scan(value interface{}) error {
	if value == nil {
		*d = DeviceInfoDTO{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for DeviceInfoDTO")
	}

	// 直接反序列化为DeviceInfoDTO结构体
	var deviceInfo DeviceInfoDTO
	if err := json.Unmarshal(bytes, &deviceInfo); err != nil {
		return err
	}

	*d = deviceInfo
	return nil
}

// Value 实现driver.Valuer接口
func (d DeviceInfoDTO) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// MarshalJSON 自定义JSON序列化
func (d DeviceInfoDTO) MarshalJSON() ([]byte, error) {
	type Alias DeviceInfoDTO
	return json.Marshal(Alias(d))
}

// UnmarshalJSON 自定义JSON反序列化
func (d *DeviceInfoDTO) UnmarshalJSON(data []byte) error {
	// 先尝试解析为新的格式
	type Alias DeviceInfoDTO
	aux := &struct {
		DeviceInfo *DeviceInfoDTO `json:"device_info"`
		*Alias
	}{
		DeviceInfo: d,
		Alias:      (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err == nil && aux.DeviceInfo != nil {
		return nil
	}

	// 如果失败，尝试解析为旧的map格式
	var deviceInfoMap map[string]interface{}
	if err := json.Unmarshal(data, &deviceInfoMap); err != nil {
		return err
	}

	// 提取字段
	if platform, ok := deviceInfoMap["platform"].(string); ok {
		d.Platform = platform
	}
	if browser, ok := deviceInfoMap["browser"].(string); ok {
		d.Browser = browser
	}
	if device, ok := deviceInfoMap["device"].(string); ok {
		d.Device = device
	}
	if ipAddress, ok := deviceInfoMap["ip_address"].(string); ok {
		d.IPAddress = ipAddress
	}
	if userAgent, ok := deviceInfoMap["user_agent"].(string); ok {
		d.UserAgent = userAgent
	}
	if fingerprint, ok := deviceInfoMap["fingerprint"].(string); ok {
		d.Fingerprint = fingerprint
	}

	return nil
}
