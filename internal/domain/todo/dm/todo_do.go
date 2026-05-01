package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"auth-perm/internal/domain/todo/constant"
)

// TodoCategoryDO 待办分类领域对象
type TodoCategoryDO struct {
	ID        string  `gorm:"primaryKey;type:uuid"`
	TenantID  string  `gorm:"column:tenant_id;type:uuid;not null;index"`
	AccountID string  `gorm:"column:account_id;type:uuid;not null;index"`
	Name      string  `gorm:"column:name;not null"`
	Color     string  `gorm:"column:color;not null;default:'#6366f1'"`
	Icon      *string `gorm:"column:icon"`
	SortOrder int     `gorm:"column:sort_order;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (*TodoCategoryDO) TableName() string { return "todo_categories" }

// NewTodoCategory 创建新分类
func NewTodoCategory(tenantID, accountID, name, color string, icon *string) *TodoCategoryDO {
	now := time.Now()
	return &TodoCategoryDO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		AccountID: accountID,
		Name:      name,
		Color:     color,
		Icon:      icon,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// TodoDO 待办领域对象
type TodoDO struct {
	ID          string                `gorm:"primaryKey;type:uuid"`
	TenantID    string                `gorm:"column:tenant_id;type:uuid;not null;index"`
	AccountID   string                `gorm:"column:account_id;type:uuid;not null;index"`
	CategoryID  *string               `gorm:"column:category_id;type:uuid"`
	Title       string                `gorm:"column:title;not null"`
	Description *string               `gorm:"column:description;type:text"`
	Status      constant.TodoStatus   `gorm:"column:status;not null;default:pending"`
	Priority    constant.TodoPriority `gorm:"column:priority;not null;default:medium"`
	Deadline    *time.Time            `gorm:"column:deadline"`
	CompletedAt *time.Time            `gorm:"column:completed_at"`
	SortOrder   int                   `gorm:"column:sort_order;default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// 关联（可选预加载）
	Category *TodoCategoryDO `gorm:"foreignKey:CategoryID"`
}

func (*TodoDO) TableName() string { return "todos" }

// NewTodo 创建新待办
func NewTodo(tenantID, accountID, title string, priority constant.TodoPriority) *TodoDO {
	now := time.Now()
	return &TodoDO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		AccountID: accountID,
		Title:     title,
		Status:    constant.TodoStatusPending,
		Priority:  priority,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsActive 是否活跃（未完成未取消）
func (t *TodoDO) IsActive() bool {
	return t.Status == constant.TodoStatusPending || t.Status == constant.TodoStatusInProgress
}
