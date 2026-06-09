package dto

import "auth-perm/internal/domain/novel/constant"

type CreateNovelRequest struct {
	Title       string               `json:"title" binding:"required,min=1,max=255"`
	Subtitle    string               `json:"subtitle"`
	Description string               `json:"description"`
	CoverURL    string               `json:"cover_url"`
	Status      constant.NovelStatus `json:"status"`
	Tags        []string             `json:"tags"`
}

type UpdateNovelRequest struct {
	Title       *string               `json:"title"`
	Subtitle    *string               `json:"subtitle"`
	Description *string               `json:"description"`
	CoverURL    *string               `json:"cover_url"`
	Status      *constant.NovelStatus `json:"status"`
	Tags        []string              `json:"tags"`
}

type CreateVolumeRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Subtitle    string `json:"subtitle"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateVolumeRequest struct {
	Title       *string `json:"title"`
	Subtitle    *string `json:"subtitle"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}

type CreateUnitRequest struct {
	VolumeID    string `json:"volume_id" binding:"required"`
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Subtitle    string `json:"subtitle"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateUnitRequest struct {
	VolumeID    *string `json:"volume_id"`
	Title       *string `json:"title"`
	Subtitle    *string `json:"subtitle"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}

type CreateChapterRequest struct {
	VolumeID  string                 `json:"volume_id" binding:"required"`
	UnitID    *string                `json:"unit_id"`
	Slug      string                 `json:"slug"`
	Number    string                 `json:"number"`
	Title     string                 `json:"title" binding:"required,min=1,max=255"`
	Summary   string                 `json:"summary"`
	Body      string                 `json:"body"`
	Status    constant.ChapterStatus `json:"status"`
	SortOrder int                    `json:"sort_order"`
}

type ImportMarkdownChapterRequest struct {
	VolumeID  string                 `json:"volume_id" binding:"required"`
	UnitID    *string                `json:"unit_id"`
	Markdown  string                 `json:"markdown" binding:"required"`
	Slug      string                 `json:"slug"`
	Number    string                 `json:"number"`
	Title     string                 `json:"title"`
	Summary   string                 `json:"summary"`
	Status    constant.ChapterStatus `json:"status"`
	SortOrder int                    `json:"sort_order"`
}

type UpdateChapterRequest struct {
	VolumeID  *string                 `json:"volume_id"`
	UnitID    *string                 `json:"unit_id"`
	Slug      *string                 `json:"slug"`
	Number    *string                 `json:"number"`
	Title     *string                 `json:"title"`
	Summary   *string                 `json:"summary"`
	Body      *string                 `json:"body"`
	Status    *constant.ChapterStatus `json:"status"`
	SortOrder *int                    `json:"sort_order"`
	Note      string                  `json:"note"`
}

type UpdateChapterStatusRequest struct {
	Status constant.ChapterStatus `json:"status" binding:"required"`
	Note   string                 `json:"note"`
}

type BatchUpdateChapterStatusRequest struct {
	IDs    []string               `json:"ids" binding:"required,min=1"`
	Status constant.ChapterStatus `json:"status" binding:"required"`
	Note   string                 `json:"note"`
}

type MarkdownBundleFile struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type ImportMarkdownBundleRequest struct {
	Files []MarkdownBundleFile `json:"files" binding:"required,min=1"`
}

type InspectMarkdownBundleRequest struct {
	Files []MarkdownBundleFile `json:"files" binding:"required,min=1"`
}

type UpsertCodexEntryRequest struct {
	Kind       constant.CodexKind `json:"kind" binding:"required"`
	Title      string             `json:"title" binding:"required,min=1,max=255"`
	Summary    string             `json:"summary"`
	Aliases    []string           `json:"aliases"`
	Properties map[string]string  `json:"properties"`
	Relations  []string           `json:"relations"`
	Evidence   string             `json:"evidence"`
	SortOrder  int                `json:"sort_order"`
}

type CreateRuleConflictRequest struct {
	TargetID   string                     `json:"target_id" binding:"required"`
	TargetType string                     `json:"target_type" binding:"required"`
	Level      constant.RuleConflictLevel `json:"level" binding:"required"`
	Title      string                     `json:"title" binding:"required,min=1,max=255"`
	Detail     string                     `json:"detail"`
}

type ResolveRuleConflictRequest struct {
	Status     constant.RuleConflictStatus `json:"status" binding:"required"`
	Resolution string                      `json:"resolution"`
}
