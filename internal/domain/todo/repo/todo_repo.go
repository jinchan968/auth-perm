package repo

import (
	"context"
	stdErr "errors"
	"time"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/todo/dm"
)

// TodoQueryParams 待办查询参数
type TodoQueryParams struct {
	TenantID   string
	AccountID  string
	CategoryID string // "" 表示不过滤；"none" 表示仅未分类
	Status     string // "" 表示不过滤
	Priority   string // "" 表示不过滤
	Keyword    string
	Page       int
	PageSize   int
}

// TodoRepo 待办仓储
type TodoRepo struct {
	db *gorm.DB
}

func NewTodoRepo(db *gorm.DB) *TodoRepo {
	return &TodoRepo{db: db}
}

func (r *TodoRepo) GetDB() *gorm.DB { return r.db }

// Create 创建待办
func (r *TodoRepo) Create(ctx context.Context, todo *dm.TodoDO) error {
	if err := r.db.WithContext(ctx).Create(todo).Error; err != nil {
		return errors.WrapBizError(err, "创建待办失败")
	}
	return nil
}

// Update 更新待办
func (r *TodoRepo) Update(ctx context.Context, todo *dm.TodoDO) error {
	if err := r.db.WithContext(ctx).Save(todo).Error; err != nil {
		return errors.WrapBizError(err, "更新待办失败")
	}
	return nil
}

// FindByID 按 ID 查找（含软删除过滤）
func (r *TodoRepo) FindByID(ctx context.Context, id string) (*dm.TodoDO, error) {
	var todo dm.TodoDO
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("id = ?", id).
		First(&todo).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("待办不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找待办失败")
	}
	return &todo, nil
}

// FindByIDAndAccount 按 ID + 账户查找（防越权）
func (r *TodoRepo) FindByIDAndAccount(ctx context.Context, id, accountID, tenantID string) (*dm.TodoDO, error) {
	var todo dm.TodoDO
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		First(&todo).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("待办不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找待办失败")
	}
	return &todo, nil
}

// List 分页查询待办列表
func (r *TodoRepo) List(ctx context.Context, p *TodoQueryParams) ([]*dm.TodoDO, int64, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&dm.TodoDO{}).
		Preload("Category").
		Where("account_id = ? AND tenant_id = ?", p.AccountID, p.TenantID)

	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}
	if p.Priority != "" {
		q = q.Where("priority = ?", p.Priority)
	}
	if p.CategoryID == "none" {
		q = q.Where("category_id IS NULL")
	} else if p.CategoryID != "" {
		q = q.Where("category_id = ?", p.CategoryID)
	}
	if p.Keyword != "" {
		like := "%" + p.Keyword + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计待办失败")
	}

	var todos []*dm.TodoDO
	offset := (p.Page - 1) * p.PageSize
	err := q.Order("sort_order ASC, created_at DESC").
		Offset(offset).Limit(p.PageSize).
		Find(&todos).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询待办列表失败")
	}

	return todos, total, nil
}

// SoftDelete 软删除
func (r *TodoRepo) SoftDelete(ctx context.Context, id, accountID, tenantID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		Delete(&dm.TodoDO{})
	if result.Error != nil {
		return errors.WrapBizError(result.Error, "删除待办失败")
	}
	if result.RowsAffected == 0 {
		return errors.NewNotFoundErrorF("待办不存在: %s", id)
	}
	return nil
}

// CountByCategoryID 统计指定分类下的待办数量
func (r *TodoRepo) CountByCategoryID(ctx context.Context, categoryID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&dm.TodoDO{}).
		Where("category_id = ?", categoryID).
		Count(&count).Error
	if err != nil {
		return 0, errors.WrapBizError(err, "统计分类待办数量失败")
	}
	return count, nil
}

// EscalateOverdue 将过期且未完成的待办优先级提升为 urgent
// 返回受影响的行数，供调度器日志使用
func (r *TodoRepo) EscalateOverdue(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Model(&dm.TodoDO{}).
		Where("deadline <= ? AND status IN ? AND priority != ? AND deleted_at IS NULL",
			time.Now(),
			[]dm.TodoStatus{dm.TodoStatusPending, dm.TodoStatusInProgress},
			dm.TodoPriorityUrgent,
		).
		Updates(map[string]interface{}{
			"priority":   dm.TodoPriorityUrgent,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return 0, errors.WrapBizError(result.Error, "升级过期待办优先级失败")
	}
	return result.RowsAffected, nil
}

// -------- Category --------

// TodoCategoryRepo 待办分类仓储
type TodoCategoryRepo struct {
	db *gorm.DB
}

func NewTodoCategoryRepo(db *gorm.DB) *TodoCategoryRepo {
	return &TodoCategoryRepo{db: db}
}

// Create 创建分类
func (r *TodoCategoryRepo) Create(ctx context.Context, cat *dm.TodoCategoryDO) error {
	if err := r.db.WithContext(ctx).Create(cat).Error; err != nil {
		return errors.WrapBizError(err, "创建分类失败")
	}
	return nil
}

// Update 更新分类
func (r *TodoCategoryRepo) Update(ctx context.Context, cat *dm.TodoCategoryDO) error {
	if err := r.db.WithContext(ctx).Save(cat).Error; err != nil {
		return errors.WrapBizError(err, "更新分类失败")
	}
	return nil
}

// FindByID 按 ID 查找分类
func (r *TodoCategoryRepo) FindByID(ctx context.Context, id string) (*dm.TodoCategoryDO, error) {
	var cat dm.TodoCategoryDO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&cat).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("分类不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找分类失败")
	}
	return &cat, nil
}

// FindByIDAndAccount 按 ID + 账户查找（防越权）
func (r *TodoCategoryRepo) FindByIDAndAccount(ctx context.Context, id, accountID, tenantID string) (*dm.TodoCategoryDO, error) {
	var cat dm.TodoCategoryDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		First(&cat).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("分类不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找分类失败")
	}
	return &cat, nil
}

// ListByAccount 查询账户下所有分类
func (r *TodoCategoryRepo) ListByAccount(ctx context.Context, accountID, tenantID string) ([]*dm.TodoCategoryDO, error) {
	var cats []*dm.TodoCategoryDO
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND tenant_id = ?", accountID, tenantID).
		Order("sort_order ASC, created_at ASC").
		Find(&cats).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询分类列表失败")
	}
	return cats, nil
}

// SoftDelete 软删除分类
func (r *TodoCategoryRepo) SoftDelete(ctx context.Context, id, accountID, tenantID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		Delete(&dm.TodoCategoryDO{})
	if result.Error != nil {
		return errors.WrapBizError(result.Error, "删除分类失败")
	}
	if result.RowsAffected == 0 {
		return errors.NewNotFoundErrorF("分类不存在: %s", id)
	}
	return nil
}
