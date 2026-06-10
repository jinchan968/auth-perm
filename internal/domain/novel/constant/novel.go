package constant

type NovelStatus string
type ChapterStatus string
type CodexKind string
type RuleConflictLevel string
type RuleConflictStatus string
type ImportTaskStatus string

const (
	NovelStatusDraft     NovelStatus = "draft"
	NovelStatusSerial    NovelStatus = "serial"
	NovelStatusCompleted NovelStatus = "completed"
	NovelStatusArchived  NovelStatus = "archived"
)

const (
	ChapterStatusDraft     ChapterStatus = "draft"
	ChapterStatusReview    ChapterStatus = "review"
	ChapterStatusPublished ChapterStatus = "published"
	ChapterStatusLocked    ChapterStatus = "locked"
)

const (
	CodexKindCharacter    CodexKind = "character"
	CodexKindEncyclopedia CodexKind = "encyclopedia"
	CodexKindGeography    CodexKind = "geography"
	CodexKindWorldview    CodexKind = "worldview"
)

const (
	RuleConflictLevelBlocking RuleConflictLevel = "blocking"
	RuleConflictLevelWarning  RuleConflictLevel = "warning"
	RuleConflictLevelHint     RuleConflictLevel = "hint"
)

const (
	RuleConflictStatusOpen     RuleConflictStatus = "open"
	RuleConflictStatusResolved RuleConflictStatus = "resolved"
	RuleConflictStatusWaived   RuleConflictStatus = "waived"
)

const (
	ImportTaskStatusPending    ImportTaskStatus = "pending"
	ImportTaskStatusProcessing ImportTaskStatus = "processing"
	ImportTaskStatusSuccess    ImportTaskStatus = "success"
	ImportTaskStatusFailed     ImportTaskStatus = "failed"
)
