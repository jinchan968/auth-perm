package dto

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// PreferencesDTO 用户偏好结构
type PreferencesDTO struct {
	Theme             string                 `json:"theme"`              // 主题（light/dark）
	Language          string                 `json:"language"`           // 语言
	TimeZone          string                 `json:"time_zone"`          // 时区
	DateFormat        string                 `json:"date_format"`        // 日期格式
	TimeFormat        string                 `json:"time_format"`        // 时间格式（12h/24h）
	Notifications     map[string]bool        `json:"notifications"`      // 通知偏好
	Privacy           map[string]interface{} `json:"privacy"`            // 隐私设置
	Accessibility     map[string]interface{} `json:"accessibility"`      // 无障碍设置
	DashboardLayout   map[string]interface{} `json:"dashboard_layout"`   // 仪表板布局
	KeyboardShortcuts map[string]string      `json:"keyboard_shortcuts"` // 键盘快捷键
	CustomSettings    map[string]interface{} `json:"custom_settings"`    // 自定义设置
}

// Scan 实现sql.Scanner接口
func (p *PreferencesDTO) Scan(value interface{}) error {
	if value == nil {
		*p = PreferencesDTO{
			Theme:             "light",
			Language:          "en",
			TimeFormat:        "24h",
			Notifications:     make(map[string]bool),
			Privacy:           make(map[string]interface{}),
			Accessibility:     make(map[string]interface{}),
			DashboardLayout:   make(map[string]interface{}),
			KeyboardShortcuts: make(map[string]string),
			CustomSettings:    make(map[string]interface{}),
		}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for PreferencesDTO")
	}

	// 直接反序列化为PreferencesDTO结构体
	var preferences PreferencesDTO
	if err := json.Unmarshal(bytes, &preferences); err != nil {
		return err
	}

	// 处理nil的slice和map
	if preferences.Notifications == nil {
		preferences.Notifications = make(map[string]bool)
	}
	if preferences.Privacy == nil {
		preferences.Privacy = make(map[string]interface{})
	}
	if preferences.Accessibility == nil {
		preferences.Accessibility = make(map[string]interface{})
	}
	if preferences.DashboardLayout == nil {
		preferences.DashboardLayout = make(map[string]interface{})
	}
	if preferences.KeyboardShortcuts == nil {
		preferences.KeyboardShortcuts = make(map[string]string)
	}
	if preferences.CustomSettings == nil {
		preferences.CustomSettings = make(map[string]interface{})
	}

	*p = preferences
	return nil
}

// Value 实现driver.Valuer接口
func (p PreferencesDTO) Value() (driver.Value, error) {
	return json.Marshal(p)
}
