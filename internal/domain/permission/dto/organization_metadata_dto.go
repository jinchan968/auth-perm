package dto

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// OrganizationMetadataDTO 组织元数据结构（用于数据库存储）
type OrganizationMetadataDTO struct {
	DisplayName    string                 `json:"display_name"`    // 显示名称
	Description    string                 `json:"description"`     // 组织描述
	Industry       string                 `json:"industry"`        // 行业
	Website        string                 `json:"website"`         // 组织网站
	Logo           string                 `json:"logo"`            // 组织Logo
	ContactEmail   string                 `json:"contact_email"`   // 联系邮箱
	ContactPhone   string                 `json:"contact_phone"`   // 联系电话
	Address        string                 `json:"address"`         // 地址
	City           string                 `json:"city"`            // 城市
	State          string                 `json:"state"`           // 州/省
	Country        string                 `json:"country"`         // 国家
	PostalCode     string                 `json:"postal_code"`     // 邮政编码
	FiscalYear     string                 `json:"fiscal_year"`     // 财年
	TimeZone       string                 `json:"time_zone"`       // 时区
	Language       string                 `json:"language"`        // 语言
	Currency       string                 `json:"currency"`        // 货币
	TaxID          string                 `json:"tax_id"`          // 税号
	RegistrationNo string                 `json:"registration_no"` // 注册号
	Settings       map[string]interface{} `json:"settings"`        // 组织设置
	Features       map[string]bool        `json:"features"`        // 功能特性
	CustomFields   map[string]interface{} `json:"custom_fields"`   // 自定义字段
}

// Scan 实现sql.Scanner接口
func (o *OrganizationMetadataDTO) Scan(value interface{}) error {
	if value == nil {
		*o = OrganizationMetadataDTO{
			Settings:     make(map[string]interface{}),
			Features:     make(map[string]bool),
			CustomFields: make(map[string]interface{}),
		}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for OrganizationMetadataDTO")
	}

	// 直接反序列化为OrganizationMetadataDTO结构体
	var metadata OrganizationMetadataDTO
	if err := json.Unmarshal(bytes, &metadata); err != nil {
		return err
	}

	// 处理nil的slice和map
	if metadata.Settings == nil {
		metadata.Settings = make(map[string]interface{})
	}
	if metadata.Features == nil {
		metadata.Features = make(map[string]bool)
	}
	if metadata.CustomFields == nil {
		metadata.CustomFields = make(map[string]interface{})
	}

	*o = metadata
	return nil
}

// Value 实现driver.Valuer接口
func (o OrganizationMetadataDTO) Value() (driver.Value, error) {
	return json.Marshal(o)
}
