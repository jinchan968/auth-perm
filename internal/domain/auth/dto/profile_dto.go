package dto

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ProfileDTO 用户资料结构
type ProfileDTO struct {
	FirstName    string                 `json:"first_name"`    // 名
	LastName     string                 `json:"last_name"`     // 姓
	Bio          string                 `json:"bio"`           // 个人简介
	Location     string                 `json:"location"`      // 所在地
	Website      string                 `json:"website"`       // 个人网站
	Gender       string                 `json:"gender"`        // 性别
	Birthday     *time.Time             `json:"birthday"`      // 生日
	Phone        string                 `json:"phone"`         // 电话号码
	Mobile       string                 `json:"mobile"`        // 手机号码
	Address      string                 `json:"address"`       // 地址
	City         string                 `json:"city"`          // 城市
	State        string                 `json:"state"`         // 州/省
	Country      string                 `json:"country"`       // 国家
	PostalCode   string                 `json:"postal_code"`   // 邮政编码
	TimeZone     string                 `json:"time_zone"`     // 时区
	Language     string                 `json:"language"`      // 语言
	CustomFields map[string]interface{} `json:"custom_fields"` // 自定义字段
}

// Scan 实现sql.Scanner接口
func (p *ProfileDTO) Scan(value interface{}) error {
	if value == nil {
		*p = ProfileDTO{CustomFields: make(map[string]interface{})}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for ProfileDTO")
	}

	// 直接反序列化为ProfileDTO结构体
	var profile ProfileDTO
	if err := json.Unmarshal(bytes, &profile); err != nil {
		return err
	}

	// 处理nil的slice和map
	if profile.CustomFields == nil {
		profile.CustomFields = make(map[string]interface{})
	}

	*p = profile
	return nil
}

// Value 实现driver.Valuer接口
func (p ProfileDTO) Value() (driver.Value, error) {
	return json.Marshal(p)
}
