package service

import (
	"context"
	"time"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/todo/dm"
	"auth-perm/internal/domain/todo/dto"
	"auth-perm/internal/domain/todo/repo"
)

// TodoService 待办业务服务
type TodoService struct {
	todoRepo     *repo.TodoRepo
	categoryRepo *repo.TodoCategoryRepo
}

func NewTodoService(todoRepo *repo.TodoRepo, categoryRepo *repo.TodoCategoryRepo) *TodoService {
	return &TodoService{todoRepo: todoRepo, categoryRepo: categoryRepo}
}

// --------- Category ---------

// ListCategories 获取账户分类列表
func (s *TodoService) ListCategories(ctx context.Context, accountID, tenantID string) (*dto.CategoryListResult, error) {
	cats, err := s.categoryRepo.ListByAccount(ctx, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.TodoCategoryDTO, 0, len(cats))
	for _, c := range cats {
		result = append(result, dto.FromCategoryDO(c))
	}
	return &dto.CategoryListResult{Data: result}, nil
}

// CreateCategory 创建分类
func (s *TodoService) CreateCategory(ctx context.Context, accountID, tenantID, name, color string, icon *string) (*dto.TodoCategoryDTO, error) {
	if name == "" {
		return nil, errors.NewValidationError("分类名称不能为空")
	}
	if color == "" {
		color = "#6366f1"
	}
	cat := dm.NewTodoCategory(tenantID, accountID, name, color, icon)
	if err := s.categoryRepo.Create(ctx, cat); err != nil {
		return nil, err
	}
	return dto.FromCategoryDO(cat), nil
}

// UpdateCategory 更新分类
func (s *TodoService) UpdateCategory(ctx context.Context, id, accountID, tenantID, name, color string, icon *string) (*dto.TodoCategoryDTO, error) {
	cat, err := s.categoryRepo.FindByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if name != "" {
		cat.Name = name
	}
	if color != "" {
		cat.Color = color
	}
	cat.Icon = icon
	cat.UpdatedAt = time.Now()
	if err := s.categoryRepo.Update(ctx, cat); err != nil {
		return nil, err
	}
	return dto.FromCategoryDO(cat), nil
}

// DeleteCategory 软删除分类
func (s *TodoService) DeleteCategory(ctx context.Context, id, accountID, tenantID string) error {
	count, err := s.todoRepo.CountByCategoryID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.NewValidationError("该分类下存在待办事项，无法删除")
	}
	return s.categoryRepo.SoftDelete(ctx, id, accountID, tenantID)
}

// --------- Todo ---------

// CreateTodoParams 创建待办参数
type CreateTodoParams struct {
	TenantID    string
	AccountID   string
	CategoryID  *string
	Title       string
	Description *string
	Priority    dm.TodoPriority
	Deadline    *time.Time
}

// UpdateTodoParams 更新待办参数
type UpdateTodoParams struct {
	ID            string
	TenantID      string
	AccountID     string
	CategoryID    *string // nil = 不变；空字符串 ptr = 清空
	Title         *string
	Description   *string
	Priority      *dm.TodoPriority
	Deadline      *time.Time
	ClearDeadline bool // 明确清空 deadline
}

// CreateTodo 创建待办
func (s *TodoService) CreateTodo(ctx context.Context, p *CreateTodoParams) (*dto.TodoDTO, error) {
	if p.Title == "" {
		return nil, errors.NewValidationError("待办标题不能为空")
	}
	if p.Priority == "" {
		p.Priority = dm.TodoPriorityMedium
	}

	// 如果指定了分类，验证分类属于该账户
	if p.CategoryID != nil && *p.CategoryID != "" {
		if _, err := s.categoryRepo.FindByIDAndAccount(ctx, *p.CategoryID, p.AccountID, p.TenantID); err != nil {
			return nil, errors.NewValidationError("分类不存在或无权访问")
		}
	}

	todo := dm.NewTodo(p.TenantID, p.AccountID, p.Title, p.Priority)
	todo.CategoryID = p.CategoryID
	todo.Description = p.Description
	todo.Deadline = p.Deadline

	if err := s.todoRepo.Create(ctx, todo); err != nil {
		return nil, err
	}

	// 预加载分类
	if todo.CategoryID != nil && *todo.CategoryID != "" {
		cat, _ := s.categoryRepo.FindByID(ctx, *todo.CategoryID)
		todo.Category = cat
	}

	return dto.FromTodoDO(todo), nil
}

// GetTodo 获取待办详情
func (s *TodoService) GetTodo(ctx context.Context, id, accountID, tenantID string) (*dto.TodoDTO, error) {
	todo, err := s.todoRepo.FindByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	return dto.FromTodoDO(todo), nil
}

// ListTodos 查询待办列表
func (s *TodoService) ListTodos(ctx context.Context, p *repo.TodoQueryParams) (*dto.TodoListResult, error) {
	todos, total, err := s.todoRepo.List(ctx, p)
	if err != nil {
		return nil, err
	}
	data := make([]*dto.TodoDTO, 0, len(todos))
	for _, t := range todos {
		data = append(data, dto.FromTodoDO(t))
	}
	return &dto.TodoListResult{
		Data:  data,
		Total: total,
		Page:  p.Page,
		Size:  p.PageSize,
	}, nil
}

// UpdateTodo 更新待办
func (s *TodoService) UpdateTodo(ctx context.Context, p *UpdateTodoParams) (*dto.TodoDTO, error) {
	todo, err := s.todoRepo.FindByIDAndAccount(ctx, p.ID, p.AccountID, p.TenantID)
	if err != nil {
		return nil, err
	}

	if p.Title != nil && *p.Title != "" {
		todo.Title = *p.Title
	}
	if p.Description != nil {
		todo.Description = p.Description
	}
	if p.Priority != nil {
		todo.Priority = *p.Priority
	}
	if p.CategoryID != nil {
		if *p.CategoryID == "" {
			todo.CategoryID = nil
		} else {
			if _, err := s.categoryRepo.FindByIDAndAccount(ctx, *p.CategoryID, p.AccountID, p.TenantID); err != nil {
				return nil, errors.NewValidationError("分类不存在或无权访问")
			}
			todo.CategoryID = p.CategoryID
		}
	}
	if p.ClearDeadline {
		todo.Deadline = nil
	} else if p.Deadline != nil {
		todo.Deadline = p.Deadline
	}
	todo.UpdatedAt = time.Now()

	if err := s.todoRepo.Update(ctx, todo); err != nil {
		return nil, err
	}
	return dto.FromTodoDO(todo), nil
}

// UpdateTodoStatus 更新待办状态
func (s *TodoService) UpdateTodoStatus(ctx context.Context, id, accountID, tenantID string, status dm.TodoStatus) (*dto.TodoDTO, error) {
	todo, err := s.todoRepo.FindByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}

	// 完成时记录完成时间；反完成时清空
	now := time.Now()
	if status == dm.TodoStatusCompleted {
		todo.CompletedAt = &now
	} else {
		todo.CompletedAt = nil
	}

	todo.Status = status
	todo.UpdatedAt = now

	if err := s.todoRepo.Update(ctx, todo); err != nil {
		return nil, err
	}
	return dto.FromTodoDO(todo), nil
}

// UpdateTodoPriority 更新待办优先级
func (s *TodoService) UpdateTodoPriority(ctx context.Context, id, accountID, tenantID string, priority dm.TodoPriority) (*dto.TodoDTO, error) {
	todo, err := s.todoRepo.FindByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	todo.Priority = priority
	todo.UpdatedAt = time.Now()
	if err := s.todoRepo.Update(ctx, todo); err != nil {
		return nil, err
	}
	return dto.FromTodoDO(todo), nil
}

// DeleteTodo 软删除待办
func (s *TodoService) DeleteTodo(ctx context.Context, id, accountID, tenantID string) error {
	return s.todoRepo.SoftDelete(ctx, id, accountID, tenantID)
}
