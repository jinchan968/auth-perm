package vo

import (
	"time"

	"auth-perm/internal/domain/journal/constant"
)

// CreateEntryRequest 创建札记请求
type CreateEntryRequest struct {
	Title    *string           `json:"title"`
	Content  string            `json:"content"  binding:"required"`
	Weather  *constant.Weather `json:"weather"`
	Location *string           `json:"location"`
	Period   constant.Period   `json:"period"   binding:"required"`
	EntryDate string           `json:"entry_date" binding:"required"` // 格式: 2006-01-02
	TagIDs   []string          `json:"tag_ids"`
}

// AddCorrectionRequest 追加修正请求
type AddCorrectionRequest struct {
	Content string `json:"content" binding:"required"`
}

// UpdateTagsRequest 更新标签请求
type UpdateTagsRequest struct {
	TagIDs []string `json:"tag_ids" binding:"required"`
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name  string `json:"name"  binding:"required,min=1"`
	Color string `json:"color"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// CreateEntryParams 创建札记参数（内部使用）
type CreateEntryParams struct {
	TenantID  string
	AccountID string
	Title     *string
	Content   string
	Weather   *constant.Weather
	Location  *string
	Period    constant.Period
	EntryDate time.Time
	TagIDs    []string
}
