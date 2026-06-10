import { ApiError, apiClient } from './client'
import type {
  CreateNovelRequest,
  MarkdownBundleImportResult,
  MarkdownBundleInspectResult,
  ImportTaskStatus,
  NovelDetail,
  NovelListItem,
  NovelListResponse,
  NovelStatus,
  UpdateNovelRequest,
  Volume,
  Unit,
  Chapter,
  ChapterStatus,
  BatchChapterStatusUpdateResult,
} from '@/types/novel'

const BASE = '/novel-admin'
const DIRECT_API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || '/api'

function buildDirectApiUrl(endpoint: string): string {
  const cleanBase = DIRECT_API_BASE_URL.replace(/\/+$/, '')
  const cleanEndpoint = endpoint.replace(/^\/+/, '')
  return `${cleanBase}/${cleanEndpoint}`
}

function getBrowserAuthToken(): string | null {
  if (typeof window === 'undefined') {
    return null
  }
  return localStorage.getItem('auth_token')
}

async function handleDirectUploadResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    let message = `请求失败: ${response.status} ${response.statusText}`
    let details: unknown

    try {
      details = await response.json()
      if (details && typeof details === 'object') {
        const payload = details as { msg?: string; message?: string; error?: string; code?: string }
        const summary = payload.msg || payload.message || ''
        const detail = payload.error || ''
        message = summary && detail && summary !== detail ? `${summary}: ${detail}` : summary || detail || message
        throw new ApiError(message, response.status, payload.code, details)
      }
    } catch (error) {
      if (error instanceof ApiError) {
        throw error
      }
    }

    throw new ApiError(message, response.status, 'UPLOAD_ERROR', details)
  }

  const payload = await response.json()
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return payload.data as T
  }
  return payload as T
}

async function postFormDirect<T>(endpoint: string, formData: FormData): Promise<T> {
  const headers = new Headers({ Accept: 'application/json' })
  const token = getBrowserAuthToken()
  if (token) {
    headers.set('x-auth-token', token)
  }

  const response = await fetch(buildDirectApiUrl(endpoint), {
    method: 'POST',
    headers,
    body: formData,
  })

  return handleDirectUploadResponse<T>(response)
}

export function listMyNovels(tenantId: string): Promise<NovelListResponse> {
  return apiClient.get<NovelListResponse>(`${BASE}/mine?tenant_id=${tenantId}&page=1&page_size=100`)
}

export function createNovel(
  tenantId: string,
  data: CreateNovelRequest,
): Promise<NovelListItem> {
  return apiClient.post<NovelListItem>(`${BASE}?tenant_id=${tenantId}`, data)
}

export function getManagedNovel(tenantId: string, novelId: string): Promise<NovelDetail> {
  return apiClient.get<NovelDetail>(`${BASE}/${novelId}/manage?tenant_id=${tenantId}`)
}

export function updateNovel(
  tenantId: string,
  novelId: string,
  data: UpdateNovelRequest,
): Promise<NovelListItem> {
  return apiClient.put<NovelListItem>(`${BASE}/${novelId}?tenant_id=${tenantId}`, data)
}

export function updateNovelStatus(
  tenantId: string,
  novelId: string,
  status: NovelStatus,
): Promise<NovelListItem> {
  return updateNovel(tenantId, novelId, { status })
}

export function listVolumes(tenantId: string, novelId: string): Promise<Volume[]> {
  return apiClient.get<Volume[]>(`${BASE}/${novelId}/volumes?tenant_id=${tenantId}`)
}

export function listUnits(tenantId: string, novelId: string): Promise<Unit[]> {
  return apiClient.get<Unit[]>(`${BASE}/${novelId}/units?tenant_id=${tenantId}`)
}

export function inspectMarkdownBundle(
  tenantId: string,
  file: File,
): Promise<MarkdownBundleInspectResult> {
  const formData = new FormData()
  formData.append('file', file)
  return postFormDirect<MarkdownBundleInspectResult>(
    `${BASE}/import-md-bundle/inspect?tenant_id=${tenantId}`,
    formData,
  )
}

export function importMarkdownBundle(
  tenantId: string,
  novelId: string,
  file: File,
): Promise<{ task_id: string; status: string }> {
  const formData = new FormData()
  formData.append('file', file)
  return postFormDirect<{ task_id: string; status: string }>(
    `${BASE}/${novelId}/import-md-bundle?tenant_id=${tenantId}`,
    formData,
  )
}

export function getImportTask(taskId: string): Promise<ImportTaskStatus> {
  return apiClient.get<ImportTaskStatus>(`${BASE}/import-tasks/${taskId}`)
}

export function deleteNovel(tenantId: string, novelId: string): Promise<void> {
  return apiClient.delete(`${BASE}/${novelId}?tenant_id=${tenantId}`)
}

export interface ImportMarkdownChapterPayload {
  volumeId: string
  unitId?: string
  file: File
  slug?: string
  number?: string
  title?: string
  summary?: string
  status?: ChapterStatus | ''
  sortOrder?: number
}

export function importMarkdownChapter(
  tenantId: string,
  novelId: string,
  payload: ImportMarkdownChapterPayload,
): Promise<Chapter> {
  const formData = new FormData()
  formData.append('volume_id', payload.volumeId)
  if (payload.unitId) formData.append('unit_id', payload.unitId)
  if (payload.slug) formData.append('slug', payload.slug)
  if (payload.number) formData.append('number', payload.number)
  if (payload.title) formData.append('title', payload.title)
  if (payload.summary) formData.append('summary', payload.summary)
  if (payload.status) formData.append('status', payload.status)
  if (payload.sortOrder) formData.append('sort_order', String(payload.sortOrder))
  formData.append('file', payload.file)

  return postFormDirect<Chapter>(
    `${BASE}/${novelId}/chapters/import-md?tenant_id=${tenantId}`,
    formData,
  )
}

export function updateChapterStatus(
  tenantId: string,
  chapterId: string,
  status: ChapterStatus,
  note?: string,
): Promise<Chapter> {
  return apiClient.patch<Chapter>(
    `${BASE}/chapters/${chapterId}/status?tenant_id=${tenantId}`,
    { status, note },
  )
}

export function batchUpdateChapterStatus(
  tenantId: string,
  novelId: string,
  ids: string[],
  status: ChapterStatus,
  note?: string,
): Promise<BatchChapterStatusUpdateResult> {
  return apiClient.patch<BatchChapterStatusUpdateResult>(
    `${BASE}/${novelId}/chapters/status?tenant_id=${tenantId}`,
    { ids, status, note },
  )
}
