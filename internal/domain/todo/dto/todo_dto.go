package dto

import (
	"time"

	"auth-perm/internal/domain/todo/constant"
	"auth-perm/internal/domain/todo/dm"
)

// TodoCategoryDTO 分类数据传输对象
type TodoCategoryDTO struct {
	ID        string  `json:"id"`
	TenantID  string  `json:"tenant_id"`
	AccountID string  `json:"account_id"`
	Name      string  `json:"name"`
	Color     string  `json:"color"`
	Icon      *string `json:"icon,omitempty"`
	SortOrder int     `json:"sort_order"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// TodoDTO 待办数据传输对象
type TodoDTO struct {
	ID          string                `json:"id"`
	TenantID    string                `json:"tenant_id"`
	AccountID   string                `json:"account_id"`
	CategoryID  *string               `json:"category_id,omitempty"`
	Category    *TodoCategoryDTO      `json:"category,omitempty"`
	Title       string                `json:"title"`
	Description *string               `json:"description,omitempty"`
	Status      constant.TodoStatus   `json:"status"`
	Priority    constant.TodoPriority `json:"priority"`
	Deadline    *time.Time            `json:"deadline,omitempty"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
	SortOrder   int                   `json:"sort_order"`
	IsOverdue   bool                  `json:"is_overdue"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// TodoListResult 待办列表结果
type TodoListResult struct {
	Data  []*TodoDTO `json:"data"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

// CategoryListResult 分类列表结果
type CategoryListResult struct {
	Data []*TodoCategoryDTO `json:"data"`
}

// FromCategoryDO 从领域对象转换分类 DTO
func FromCategoryDO(d *dm.TodoCategoryDO) *TodoCategoryDTO {
	if d == nil {
		return nil
	}
	return &TodoCategoryDTO{
		ID:        d.ID,
		TenantID:  d.TenantID,
		AccountID: d.AccountID,
		Name:      d.Name,
		Color:     d.Color,
		Icon:      d.Icon,
		SortOrder: d.SortOrder,
		CreatedAt: d.CreatedAt.Format(time.RFC3339),
		UpdatedAt: d.UpdatedAt.Format(time.RFC3339),
	}
}

// FromTodoDO 从领域对象转换待办 DTO
func FromTodoDO(d *dm.TodoDO) *TodoDTO {
	if d == nil {
		return nil
	}

	now := time.Now()
	isOverdue := d.Deadline != nil && d.Deadline.Before(now) && d.IsActive()

	t := &TodoDTO{
		ID:          d.ID,
		TenantID:    d.TenantID,
		AccountID:   d.AccountID,
		CategoryID:  d.CategoryID,
		Title:       d.Title,
		Description: d.Description,
		Status:      d.Status,
		Priority:    d.Priority,
		Deadline:    d.Deadline,
		CompletedAt: d.CompletedAt,
		SortOrder:   d.SortOrder,
		IsOverdue:   isOverdue,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}

	if d.Category != nil {
		t.Category = FromCategoryDO(d.Category)
	}

	return t
}
