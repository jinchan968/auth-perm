package vo

import (
	"time"

	"auth-perm/internal/domain/todo/constant"
)

// CreateTodoRequest 创建待办请求
type CreateTodoRequest struct {
	CategoryID  *string               `json:"category_id"`
	Title       string                `json:"title"       binding:"required,min=1,max=500"`
	Description *string               `json:"description"`
	Priority    constant.TodoPriority `json:"priority"`
	Deadline    *time.Time            `json:"deadline"`
}

// UpdateTodoRequest 更新待办请求
type UpdateTodoRequest struct {
	CategoryID    *string                `json:"category_id"`
	ClearCategory bool                   `json:"clear_category"`
	Title         *string                `json:"title"`
	Description   *string                `json:"description"`
	Priority      *constant.TodoPriority `json:"priority"`
	Deadline      *time.Time             `json:"deadline"`
	ClearDeadline bool                   `json:"clear_deadline"`
}

// UpdateStatusRequest 更新状态请求
type UpdateStatusRequest struct {
	Status constant.TodoStatus `json:"status" binding:"required"`
}

// UpdatePriorityRequest 更新优先级请求
type UpdatePriorityRequest struct {
	Priority constant.TodoPriority `json:"priority" binding:"required"`
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name  string  `json:"name"  binding:"required,min=1,max=100"`
	Color string  `json:"color"`
	Icon  *string `json:"icon"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name  string  `json:"name"`
	Color string  `json:"color"`
	Icon  *string `json:"icon"`
}
