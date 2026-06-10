package vo

import (
	"time"

	"auth-perm/internal/domain/novel/constant"
	"auth-perm/internal/domain/novel/dm"
)

type NovelVO struct {
	ID          string               `json:"id"`
	TenantID    string               `json:"tenant_id"`
	AccountID   string               `json:"account_id"`
	Title       string               `json:"title"`
	Subtitle    string               `json:"subtitle"`
	Description string               `json:"description"`
	CoverURL    string               `json:"cover_url"`
	Status      constant.NovelStatus `json:"status"`
	Tags        []string             `json:"tags"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type VolumeVO struct {
	ID          string    `json:"id"`
	NovelID     string    `json:"novel_id"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UnitVO struct {
	ID          string    `json:"id"`
	NovelID     string    `json:"novel_id"`
	VolumeID    string    `json:"volume_id"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ChapterVO struct {
	ID             string                 `json:"id"`
	NovelID        string                 `json:"novel_id"`
	VolumeID       string                 `json:"volume_id"`
	UnitID         *string                `json:"unit_id,omitempty"`
	Slug           string                 `json:"slug"`
	Number         string                 `json:"number"`
	Title          string                 `json:"title"`
	Summary        string                 `json:"summary"`
	Body           string                 `json:"body,omitempty"`
	Status         constant.ChapterStatus `json:"status"`
	WordCount      int                    `json:"word_count"`
	ReadingMinutes int                    `json:"reading_minutes"`
	SortOrder      int                    `json:"sort_order"`
	PublishedAt    *time.Time             `json:"published_at"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type ChapterVersionVO struct {
	ID        string    `json:"id"`
	ChapterID string    `json:"chapter_id"`
	Label     string    `json:"label"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Body      string    `json:"body,omitempty"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type CodexEntryVO struct {
	ID         string             `json:"id"`
	NovelID    string             `json:"novel_id"`
	Kind       constant.CodexKind `json:"kind"`
	Title      string             `json:"title"`
	Summary    string             `json:"summary"`
	Aliases    []string           `json:"aliases"`
	Properties map[string]string  `json:"properties"`
	Relations  []string           `json:"relations"`
	Evidence   string             `json:"evidence"`
	SortOrder  int                `json:"sort_order"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type RuleConflictVO struct {
	ID         string                      `json:"id"`
	NovelID    string                      `json:"novel_id"`
	TargetID   string                      `json:"target_id"`
	TargetType string                      `json:"target_type"`
	Level      constant.RuleConflictLevel  `json:"level"`
	Status     constant.RuleConflictStatus `json:"status"`
	Title      string                      `json:"title"`
	Detail     string                      `json:"detail"`
	Resolution string                      `json:"resolution"`
	CreatedAt  time.Time                   `json:"created_at"`
	UpdatedAt  time.Time                   `json:"updated_at"`
}

type NovelDetailVO struct {
	Novel    *NovelVO     `json:"novel"`
	Volumes  []*VolumeVO  `json:"volumes"`
	Units    []*UnitVO    `json:"units"`
	Chapters []*ChapterVO `json:"chapters"`
}

type MarkdownBundleImportItemVO struct {
	Path      string     `json:"path"`
	VolumeID  string     `json:"volume_id"`
	UnitID    *string    `json:"unit_id,omitempty"`
	ChapterID string     `json:"chapter_id,omitempty"`
	Slug      string     `json:"slug,omitempty"`
	Action    string     `json:"action"`
	Skipped   bool       `json:"skipped"`
	Reason    string     `json:"reason,omitempty"`
	Chapter   *ChapterVO `json:"chapter,omitempty"`
}

type MarkdownBundleImportResultVO struct {
	Imported int                           `json:"imported"`
	Created  int                           `json:"created"`
	Updated  int                           `json:"updated"`
	Skipped  int                           `json:"skipped"`
	Items    []*MarkdownBundleImportItemVO `json:"items"`
}

type ImportTaskVO struct {
	TaskID   string                          `json:"task_id"`
	NovelID  string                          `json:"novel_id"`
	Status   constant.ImportTaskStatus       `json:"status"`
	Progress *ImportTaskProgressVO           `json:"progress,omitempty"`
	Result   *MarkdownBundleImportResultVO   `json:"result,omitempty"`
	Error    string                          `json:"error,omitempty"`
}

type ImportTaskProgressVO struct {
	Total     int `json:"total"`
	Processed int `json:"processed"`
}

type MarkdownBundleInspectItemVO struct {
	Path         string                 `json:"path"`
	VolumeTitle  string                 `json:"volume_title"`
	UnitTitle    string                 `json:"unit_title"`
	ChapterTitle string                 `json:"chapter_title"`
	Title        string                 `json:"title"`
	Slug         string                 `json:"slug"`
	Number       string                 `json:"number"`
	Summary      string                 `json:"summary"`
	Status       constant.ChapterStatus `json:"status"`
	SortOrder    int                    `json:"sort_order"`
	WordCount    int                    `json:"word_count"`
	Skipped      bool                   `json:"skipped"`
	Reason       string                 `json:"reason,omitempty"`
}

type MarkdownBundleInspectResultVO struct {
	Total    int                            `json:"total"`
	Valid    int                            `json:"valid"`
	Skipped  int                            `json:"skipped"`
	Volumes  []string                       `json:"volumes"`
	Units    []string                       `json:"units"`
	Items    []*MarkdownBundleInspectItemVO `json:"items"`
	Strategy string                         `json:"strategy"`
}

type BatchChapterStatusUpdateResultVO struct {
	Updated  int          `json:"updated"`
	Skipped  int          `json:"skipped"`
	Chapters []*ChapterVO `json:"chapters"`
}

type ListResult[T any] struct {
	Data  []*T  `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

func FromNovelDO(n *dm.NovelDO) *NovelVO {
	if n == nil {
		return nil
	}
	return &NovelVO{
		ID:          n.ID,
		TenantID:    n.TenantID,
		AccountID:   n.AccountID,
		Title:       n.Title,
		Subtitle:    n.Subtitle,
		Description: n.Description,
		CoverURL:    n.CoverURL,
		Status:      n.Status,
		Tags:        n.Tags,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
}

func FromVolumeDO(v *dm.NovelVolumeDO) *VolumeVO {
	if v == nil {
		return nil
	}
	return &VolumeVO{
		ID:          v.ID,
		NovelID:     v.NovelID,
		Title:       v.Title,
		Subtitle:    v.Subtitle,
		Description: v.Description,
		SortOrder:   v.SortOrder,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

func FromUnitDO(u *dm.NovelUnitDO) *UnitVO {
	if u == nil {
		return nil
	}
	return &UnitVO{
		ID:          u.ID,
		NovelID:     u.NovelID,
		VolumeID:    u.VolumeID,
		Title:       u.Title,
		Subtitle:    u.Subtitle,
		Description: u.Description,
		SortOrder:   u.SortOrder,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

func FromChapterDO(c *dm.NovelChapterDO, includeBody bool) *ChapterVO {
	if c == nil {
		return nil
	}
	body := ""
	if includeBody {
		body = c.Body
	}
	return &ChapterVO{
		ID:             c.ID,
		NovelID:        c.NovelID,
		VolumeID:       c.VolumeID,
		UnitID:         c.UnitID,
		Slug:           c.Slug,
		Number:         c.Number,
		Title:          c.Title,
		Summary:        c.Summary,
		Body:           body,
		Status:         c.Status,
		WordCount:      c.WordCount,
		ReadingMinutes: c.ReadingMinutes,
		SortOrder:      c.SortOrder,
		PublishedAt:    c.PublishedAt,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func FromChapterVersionDO(v *dm.NovelChapterVersionDO, includeBody bool) *ChapterVersionVO {
	if v == nil {
		return nil
	}
	body := ""
	if includeBody {
		body = v.Body
	}
	return &ChapterVersionVO{
		ID:        v.ID,
		ChapterID: v.ChapterID,
		Label:     v.Label,
		Title:     v.Title,
		Summary:   v.Summary,
		Body:      body,
		Note:      v.Note,
		CreatedAt: v.CreatedAt,
	}
}

func FromCodexEntryDO(e *dm.NovelCodexEntryDO) *CodexEntryVO {
	if e == nil {
		return nil
	}
	return &CodexEntryVO{
		ID:         e.ID,
		NovelID:    e.NovelID,
		Kind:       e.Kind,
		Title:      e.Title,
		Summary:    e.Summary,
		Aliases:    e.Aliases,
		Properties: e.Properties,
		Relations:  e.Relations,
		Evidence:   e.Evidence,
		SortOrder:  e.SortOrder,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}
}

func FromRuleConflictDO(c *dm.NovelRuleConflictDO) *RuleConflictVO {
	if c == nil {
		return nil
	}
	return &RuleConflictVO{
		ID:         c.ID,
		NovelID:    c.NovelID,
		TargetID:   c.TargetID,
		TargetType: c.TargetType,
		Level:      c.Level,
		Status:     c.Status,
		Title:      c.Title,
		Detail:     c.Detail,
		Resolution: c.Resolution,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

func IsValidNovelStatus(status constant.NovelStatus) bool {
	switch status {
	case "", constant.NovelStatusDraft, constant.NovelStatusSerial, constant.NovelStatusCompleted, constant.NovelStatusArchived:
		return true
	default:
		return false
	}
}

func IsValidChapterStatus(status constant.ChapterStatus) bool {
	switch status {
	case "", constant.ChapterStatusDraft, constant.ChapterStatusReview, constant.ChapterStatusPublished, constant.ChapterStatusLocked:
		return true
	default:
		return false
	}
}

func IsValidCodexKind(kind constant.CodexKind) bool {
	switch kind {
	case constant.CodexKindCharacter, constant.CodexKindEncyclopedia, constant.CodexKindGeography, constant.CodexKindWorldview:
		return true
	default:
		return false
	}
}

func IsValidRuleConflictLevel(level constant.RuleConflictLevel) bool {
	switch level {
	case constant.RuleConflictLevelBlocking, constant.RuleConflictLevelWarning, constant.RuleConflictLevelHint:
		return true
	default:
		return false
	}
}

func IsValidRuleConflictStatus(status constant.RuleConflictStatus) bool {
	switch status {
	case constant.RuleConflictStatusOpen, constant.RuleConflictStatusResolved, constant.RuleConflictStatusWaived:
		return true
	default:
		return false
	}
}
