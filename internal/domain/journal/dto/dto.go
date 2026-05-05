package dto

import (
	"time"

	"auth-perm/internal/domain/journal/constant"
	"auth-perm/internal/domain/journal/dm"
)

// TagDTO 标签数据传输对象
type TagDTO struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CorrectionDTO 修正条目数据传输对象
type CorrectionDTO struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// EntryDTO 札记条目数据传输对象
type EntryDTO struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	AccountID   string            `json:"account_id"`
	ParentID    *string           `json:"parent_id,omitempty"`
	Title       *string           `json:"title,omitempty"`
	Content     string            `json:"content"`
	Weather     *constant.Weather `json:"weather,omitempty"`
	Location    *string           `json:"location,omitempty"`
	Period      constant.Period   `json:"period"`
	EntryDate   string            `json:"entry_date"`
	Tags        []*TagDTO         `json:"tags,omitempty"`
	Corrections []*CorrectionDTO  `json:"corrections,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// EntryListResult 札记列表结果
type EntryListResult struct {
	Data     []*EntryDTO `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// TagListResult 标签列表结果
type TagListResult struct {
	Data []*TagDTO `json:"data"`
}

// FromTagDO 从领域对象转换标签 DTO
func FromTagDO(d *dm.TagDO) *TagDTO {
	if d == nil {
		return nil
	}
	return &TagDTO{
		ID:        d.ID,
		TenantID:  d.TenantID,
		AccountID: d.AccountID,
		Name:      d.Name,
		Color:     d.Color,
		CreatedAt: d.CreatedAt.Format(time.RFC3339),
		UpdatedAt: d.UpdatedAt.Format(time.RFC3339),
	}
}

// FromEntryDO 从领域对象转换札记 DTO
func FromEntryDO(d *dm.JournalEntryDO) *EntryDTO {
	if d == nil {
		return nil
	}

	e := &EntryDTO{
		ID:        d.ID,
		TenantID:  d.TenantID,
		AccountID: d.AccountID,
		ParentID:  d.ParentID,
		Title:     d.Title,
		Content:   d.Content,
		Weather:   d.Weather,
		Location:  d.Location,
		Period:    d.Period,
		EntryDate: d.EntryDate.Format("2006-01-02"),
		CreatedAt: d.CreatedAt.Format(time.RFC3339),
		UpdatedAt: d.UpdatedAt.Format(time.RFC3339),
	}

	if len(d.Tags) > 0 {
		tags := make([]*TagDTO, 0, len(d.Tags))
		for _, t := range d.Tags {
			tags = append(tags, FromTagDO(t))
		}
		e.Tags = tags
	}

	if len(d.Corrections) > 0 {
		corrections := make([]*CorrectionDTO, 0, len(d.Corrections))
		for _, c := range d.Corrections {
			corrections = append(corrections, &CorrectionDTO{
				ID:        c.ID,
				Content:   c.Content,
				CreatedAt: c.CreatedAt.Format(time.RFC3339),
			})
		}
		e.Corrections = corrections
	}

	return e
}

// FromEntryDOList 批量转换
func FromEntryDOList(entries []*dm.JournalEntryDO) []*EntryDTO {
	result := make([]*EntryDTO, 0, len(entries))
	for _, e := range entries {
		result = append(result, FromEntryDO(e))
	}
	return result
}
