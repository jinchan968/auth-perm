package dm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"auth-perm/internal/domain/novel/constant"
)

type NovelDO struct {
	ID           string               `gorm:"primaryKey;type:varchar(36)"`
	TenantID     string               `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	AccountID    string               `gorm:"column:account_id;type:varchar(36);not null;index"`
	Title        string               `gorm:"column:title;type:varchar(255);not null"`
	Subtitle     string               `gorm:"column:subtitle;type:varchar(255)"`
	Description  string               `gorm:"column:description;type:text"`
	CoverURL     string               `gorm:"column:cover_url;type:text"`
	Status       constant.NovelStatus `gorm:"column:status;type:varchar(32);not null;default:serial;index"`
	Tags         []string             `gorm:"column:tags;type:jsonb;serializer:json"`
	CreatedAt    time.Time            `gorm:"column:created_at"`
	UpdatedAt    time.Time            `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt       `gorm:"index"`
	Volumes      []*NovelVolumeDO     `gorm:"foreignKey:NovelID"`
	Units        []*NovelUnitDO       `gorm:"foreignKey:NovelID"`
	Chapters     []*NovelChapterDO    `gorm:"foreignKey:NovelID"`
	CodexEntries []*NovelCodexEntryDO `gorm:"foreignKey:NovelID"`
}

func (*NovelDO) TableName() string { return "novels" }

func NewNovel(tenantID, accountID, title string) *NovelDO {
	now := time.Now()
	return &NovelDO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		AccountID: accountID,
		Title:     title,
		Status:    constant.NovelStatusSerial,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type NovelVolumeDO struct {
	ID          string            `gorm:"primaryKey;type:varchar(36)"`
	TenantID    string            `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	AccountID   string            `gorm:"column:account_id;type:varchar(36);not null;index"`
	NovelID     string            `gorm:"column:novel_id;type:varchar(36);not null;index"`
	Title       string            `gorm:"column:title;type:varchar(255);not null"`
	Subtitle    string            `gorm:"column:subtitle;type:varchar(255)"`
	Description string            `gorm:"column:description;type:text"`
	SortOrder   int               `gorm:"column:sort_order;not null;default:0"`
	CreatedAt   time.Time         `gorm:"column:created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt    `gorm:"index"`
	Novel       *NovelDO          `gorm:"foreignKey:NovelID"`
	Units       []*NovelUnitDO    `gorm:"foreignKey:VolumeID"`
	Chapters    []*NovelChapterDO `gorm:"foreignKey:VolumeID"`
}

func (*NovelVolumeDO) TableName() string { return "novel_volumes" }

func NewVolume(tenantID, accountID, novelID, title string, sortOrder int) *NovelVolumeDO {
	now := time.Now()
	return &NovelVolumeDO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		AccountID: accountID,
		NovelID:   novelID,
		Title:     title,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type NovelUnitDO struct {
	ID          string            `gorm:"primaryKey;type:varchar(36)"`
	TenantID    string            `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	AccountID   string            `gorm:"column:account_id;type:varchar(36);not null;index"`
	NovelID     string            `gorm:"column:novel_id;type:varchar(36);not null;index"`
	VolumeID    string            `gorm:"column:volume_id;type:varchar(36);not null;index"`
	Title       string            `gorm:"column:title;type:varchar(255);not null"`
	Subtitle    string            `gorm:"column:subtitle;type:varchar(255)"`
	Description string            `gorm:"column:description;type:text"`
	SortOrder   int               `gorm:"column:sort_order;not null;default:0"`
	CreatedAt   time.Time         `gorm:"column:created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt    `gorm:"index"`
	Novel       *NovelDO          `gorm:"foreignKey:NovelID"`
	Volume      *NovelVolumeDO    `gorm:"foreignKey:VolumeID"`
	Chapters    []*NovelChapterDO `gorm:"foreignKey:UnitID"`
}

func (*NovelUnitDO) TableName() string { return "novel_units" }

func NewUnit(tenantID, accountID, novelID, volumeID, title string, sortOrder int) *NovelUnitDO {
	now := time.Now()
	return &NovelUnitDO{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		AccountID: accountID,
		NovelID:   novelID,
		VolumeID:  volumeID,
		Title:     title,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type NovelChapterDO struct {
	ID             string                   `gorm:"primaryKey;type:varchar(36)"`
	TenantID       string                   `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	AccountID      string                   `gorm:"column:account_id;type:varchar(36);not null;index"`
	NovelID        string                   `gorm:"column:novel_id;type:varchar(36);not null;index"`
	VolumeID       string                   `gorm:"column:volume_id;type:varchar(36);not null;index"`
	UnitID         *string                  `gorm:"column:unit_id;type:varchar(36);index"`
	Slug           string                   `gorm:"column:slug;type:varchar(128);not null;index"`
	Number         string                   `gorm:"column:number;type:varchar(32)"`
	Title          string                   `gorm:"column:title;type:varchar(255);not null"`
	Summary        string                   `gorm:"column:summary;type:text"`
	Body           string                   `gorm:"column:body;type:text"`
	Status         constant.ChapterStatus   `gorm:"column:status;type:varchar(32);not null;default:draft;index"`
	WordCount      int                      `gorm:"column:word_count;not null;default:0"`
	ReadingMinutes int                      `gorm:"column:reading_minutes;not null;default:1"`
	SortOrder      int                      `gorm:"column:sort_order;not null;default:0"`
	PublishedAt    *time.Time               `gorm:"column:published_at"`
	CreatedAt      time.Time                `gorm:"column:created_at"`
	UpdatedAt      time.Time                `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt           `gorm:"index"`
	Novel          *NovelDO                 `gorm:"foreignKey:NovelID"`
	Volume         *NovelVolumeDO           `gorm:"foreignKey:VolumeID"`
	Unit           *NovelUnitDO             `gorm:"foreignKey:UnitID"`
	Versions       []*NovelChapterVersionDO `gorm:"foreignKey:ChapterID"`
}

func (*NovelChapterDO) TableName() string { return "novel_chapters" }

func NewChapter(tenantID, accountID, novelID, volumeID, title string, sortOrder int) *NovelChapterDO {
	now := time.Now()
	return &NovelChapterDO{
		ID:             uuid.New().String(),
		TenantID:       tenantID,
		AccountID:      accountID,
		NovelID:        novelID,
		VolumeID:       volumeID,
		Title:          title,
		Status:         constant.ChapterStatusDraft,
		ReadingMinutes: 1,
		SortOrder:      sortOrder,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

type NovelChapterVersionDO struct {
	ID        string          `gorm:"primaryKey;type:varchar(36)"`
	TenantID  string          `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	AccountID string          `gorm:"column:account_id;type:varchar(36);not null;index"`
	ChapterID string          `gorm:"column:chapter_id;type:varchar(36);not null;index"`
	Label     string          `gorm:"column:label;type:varchar(64);not null"`
	Title     string          `gorm:"column:title;type:varchar(255);not null"`
	Summary   string          `gorm:"column:summary;type:text"`
	Body      string          `gorm:"column:body;type:text"`
	Note      string          `gorm:"column:note;type:text"`
	CreatedAt time.Time       `gorm:"column:created_at"`
	Chapter   *NovelChapterDO `gorm:"foreignKey:ChapterID"`
}

func (*NovelChapterVersionDO) TableName() string { return "novel_chapter_versions" }

func NewChapterVersion(chapter *NovelChapterDO, label, note string) *NovelChapterVersionDO {
	return &NovelChapterVersionDO{
		ID:        uuid.New().String(),
		TenantID:  chapter.TenantID,
		AccountID: chapter.AccountID,
		ChapterID: chapter.ID,
		Label:     label,
		Title:     chapter.Title,
		Summary:   chapter.Summary,
		Body:      chapter.Body,
		Note:      note,
		CreatedAt: time.Now(),
	}
}

type NovelCodexEntryDO struct {
	ID         string                 `gorm:"primaryKey;type:varchar(36)"`
	TenantID   string                 `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	AccountID  string                 `gorm:"column:account_id;type:varchar(36);not null;index"`
	NovelID    string                 `gorm:"column:novel_id;type:varchar(36);not null;index"`
	Kind       constant.CodexKind     `gorm:"column:kind;type:varchar(32);not null;index"`
	Title      string                 `gorm:"column:title;type:varchar(255);not null"`
	Summary    string                 `gorm:"column:summary;type:text"`
	Aliases    []string               `gorm:"column:aliases;type:jsonb;serializer:json"`
	Properties map[string]string      `gorm:"column:properties;type:jsonb;serializer:json"`
	Relations  []string               `gorm:"column:relations;type:jsonb;serializer:json"`
	Evidence   string                 `gorm:"column:evidence;type:text"`
	SortOrder  int                    `gorm:"column:sort_order;not null;default:0"`
	CreatedAt  time.Time              `gorm:"column:created_at"`
	UpdatedAt  time.Time              `gorm:"column:updated_at"`
	DeletedAt  gorm.DeletedAt         `gorm:"index"`
	Novel      *NovelDO               `gorm:"foreignKey:NovelID"`
	Conflicts  []*NovelRuleConflictDO `gorm:"foreignKey:TargetID;references:ID"`
}

func (*NovelCodexEntryDO) TableName() string { return "novel_codex_entries" }

func NewCodexEntry(tenantID, accountID, novelID string, kind constant.CodexKind, title string) *NovelCodexEntryDO {
	now := time.Now()
	return &NovelCodexEntryDO{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		AccountID:  accountID,
		NovelID:    novelID,
		Kind:       kind,
		Title:      title,
		Aliases:    []string{},
		Properties: map[string]string{},
		Relations:  []string{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

type NovelRuleConflictDO struct {
	ID         string                      `gorm:"primaryKey;type:varchar(36)"`
	TenantID   string                      `gorm:"column:tenant_id;type:varchar(36);not null;index"`
	AccountID  string                      `gorm:"column:account_id;type:varchar(36);not null;index"`
	NovelID    string                      `gorm:"column:novel_id;type:varchar(36);not null;index"`
	TargetID   string                      `gorm:"column:target_id;type:varchar(36);not null;index"`
	TargetType string                      `gorm:"column:target_type;type:varchar(64);not null;index"`
	Level      constant.RuleConflictLevel  `gorm:"column:level;type:varchar(32);not null;index"`
	Status     constant.RuleConflictStatus `gorm:"column:status;type:varchar(32);not null;default:open;index"`
	Title      string                      `gorm:"column:title;type:varchar(255);not null"`
	Detail     string                      `gorm:"column:detail;type:text"`
	Resolution string                      `gorm:"column:resolution;type:text"`
	CreatedAt  time.Time                   `gorm:"column:created_at"`
	UpdatedAt  time.Time                   `gorm:"column:updated_at"`
	DeletedAt  gorm.DeletedAt              `gorm:"index"`
}

func (*NovelRuleConflictDO) TableName() string { return "novel_rule_conflicts" }

func NewRuleConflict(tenantID, accountID, novelID, targetID, targetType string, level constant.RuleConflictLevel, title string) *NovelRuleConflictDO {
	now := time.Now()
	return &NovelRuleConflictDO{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		AccountID:  accountID,
		NovelID:    novelID,
		TargetID:   targetID,
		TargetType: targetType,
		Level:      level,
		Status:     constant.RuleConflictStatusOpen,
		Title:      title,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
