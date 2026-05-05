package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"auth-perm/internal/domain/journal/constant"
)

// JournalEntryDO 札记条目领域对象
type JournalEntryDO struct {
	ID        string            `gorm:"primaryKey;type:varchar(36)"`
	TenantID  string            `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	AccountID string            `gorm:"column:account_id;type:varchar(36);not null;index"`
	ParentID  *string           `gorm:"column:parent_id;type:varchar(36);index"` // nil=主条目，非nil=修正条目
	Title     *string           `gorm:"column:title;type:text"`
	Content   string            `gorm:"column:content;type:text;not null"`
	Weather   *constant.Weather `gorm:"column:weather;type:varchar(20)"`
	Location  *string           `gorm:"column:location;type:varchar(200)"`
	Period    constant.Period   `gorm:"column:period;type:varchar(10);not null"`
	EntryDate time.Time         `gorm:"column:entry_date;type:date;not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// 关联（可选预加载）
	Tags        []*TagDO          `gorm:"many2many:diary_tags;foreignKey:ID;joinForeignKey:DiaryID;References:ID;joinReferences:TagID"`
	Corrections []*JournalEntryDO `gorm:"foreignKey:ParentID"`
	Parent      *JournalEntryDO   `gorm:"foreignKey:ParentID"`
}

func (*JournalEntryDO) TableName() string { return "journal_entries" }

// NewJournalEntry 创建新札记主条目
func NewJournalEntry(tenantID, accountID string, title *string, content string, weather *constant.Weather, location *string, period constant.Period, entryDate time.Time) *JournalEntryDO {
	now := time.Now()
	return &JournalEntryDO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		AccountID: accountID,
		Title:     title,
		Content:   content,
		Weather:   weather,
		Location:  location,
		Period:    period,
		EntryDate: entryDate,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewCorrection 创建修正条目
func NewCorrection(tenantID, accountID, parentID, content string, parentEntryDate time.Time) *JournalEntryDO {
	now := time.Now()
	return &JournalEntryDO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		AccountID: accountID,
		ParentID:  &parentID,
		Content:   content,
		Period:    constant.PeriodMorning, // 修正条目不关心时段，填默认值
		EntryDate: parentEntryDate,        // 继承主条目的日期
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsCorrection 是否为修正条目
func (e *JournalEntryDO) IsCorrection() bool {
	return e.ParentID != nil
}

// TagDO 标签领域对象
type TagDO struct {
	ID        string `gorm:"primaryKey;type:varchar(36)"`
	TenantID  string `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	AccountID string `gorm:"column:account_id;type:varchar(36);not null;index"`
	Name      string `gorm:"column:name;type:varchar(30);not null"`
	Color     string `gorm:"column:color;type:varchar(7);not null;default:'#6366f1'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (*TagDO) TableName() string { return "journal_tags" }

// NewTag 创建新标签
func NewTag(tenantID, accountID, name, color string) *TagDO {
	now := time.Now()
	if color == "" {
		color = "#6366f1"
	}
	return &TagDO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		AccountID: accountID,
		Name:      name,
		Color:     color,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// DiaryTagDO 札记-标签关联表（用于手动管理 GORM many2many）
type DiaryTagDO struct {
	DiaryID   string `gorm:"primaryKey;type:varchar(36)"`
	TagID     string `gorm:"primaryKey;type:varchar(36)"`
	CreatedAt time.Time
}

func (*DiaryTagDO) TableName() string { return "diary_tags" }
