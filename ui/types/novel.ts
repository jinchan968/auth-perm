export type NovelStatus = 'draft' | 'serial' | 'completed' | 'archived'
export type ChapterStatus = 'draft' | 'review' | 'published' | 'locked'
export type ImportTaskStatusValue = 'pending' | 'processing' | 'success' | 'failed'

export const IMPORT_TASK_STATUS = {
  PENDING: 'pending',
  PROCESSING: 'processing',
  SUCCESS: 'success',
  FAILED: 'failed',
} as const satisfies Record<string, ImportTaskStatusValue>

export interface NovelListItem {
  id: string
  tenant_id: string
  account_id: string
  title: string
  subtitle: string
  description: string
  cover_url: string
  status: NovelStatus
  tags: string[]
  created_at: string
  updated_at: string
}

export interface NovelListResponse {
  data: NovelListItem[]
  total: number
  page: number
  size: number
}

export interface CreateNovelRequest {
  title: string
  subtitle?: string
  description?: string
  cover_url?: string
  status?: NovelStatus
  tags?: string[]
}

export interface UpdateNovelRequest {
  title?: string
  subtitle?: string
  description?: string
  cover_url?: string
  status?: NovelStatus
  tags?: string[]
}

export interface Volume {
  id: string
  novel_id: string
  title: string
  subtitle: string
  description: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface Unit {
  id: string
  novel_id: string
  volume_id: string
  title: string
  subtitle: string
  description: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface Chapter {
  id: string
  novel_id: string
  volume_id: string
  unit_id?: string
  slug: string
  number: string
  title: string
  summary: string
  status: ChapterStatus
  word_count: number
  reading_minutes: number
  sort_order: number
  created_at: string
  updated_at: string
}

export interface NovelDetail {
  novel: NovelListItem
  volumes: Volume[]
  units: Unit[]
  chapters: Chapter[]
}

export interface MarkdownBundleInspectItem {
  path: string
  volume_title: string
  unit_title: string
  chapter_title: string
  title: string
  slug: string
  number: string
  summary: string
  status: ChapterStatus | ''
  sort_order: number
  word_count: number
  skipped: boolean
  reason?: string
}

export interface MarkdownBundleInspectResult {
  total: number
  valid: number
  skipped: number
  volumes: string[]
  units: string[]
  items: MarkdownBundleInspectItem[]
  strategy: string
}

export interface MarkdownBundleImportItem {
  path: string
  volume_id: string
  unit_id?: string
  chapter_id?: string
  slug?: string
  action: 'created' | 'updated' | 'skipped'
  skipped: boolean
  reason?: string
  chapter?: Chapter
}

export interface MarkdownBundleImportResult {
  imported: number
  created: number
  updated: number
  skipped: number
  items: MarkdownBundleImportItem[]
}

export interface ImportTaskProgress {
  total: number
  processed: number
}

export interface ImportTaskStatus {
  task_id: string
  novel_id: string
  status: ImportTaskStatusValue
  progress?: ImportTaskProgress
  result?: MarkdownBundleImportResult
  error?: string
}

export interface BatchChapterStatusUpdateResult {
  updated: number
  skipped: number
  chapters: Chapter[]
}
