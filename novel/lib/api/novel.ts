import { editorialChapters, worldbuilding } from "@/lib/novel-data"
import type {
  AdjacentChapter,
  Chapter,
  ChapterStatus,
  EditorialChapter,
  Novel,
  NovelStatus,
  NovelSummary,
  Volume,
  WorldbuildingSnapshot,
} from "@/types/novel"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"
const TENANT_ID = process.env.NEXT_PUBLIC_TENANT_ID

type ApiResponse<T> = {
  code: number
  msg?: string
  data?: T
  error?: string
}

type ListResult<T> = {
  data: T[]
  total: number
  page: number
  size: number
}

type NovelVO = {
  id: string
  title: string
  subtitle: string
  description: string
  cover_url: string
  status: NovelStatus
  tags: string[] | null
  updated_at: string
}

type VolumeVO = {
  id: string
  title: string
  subtitle: string
  description: string
  sort_order: number
}

type ChapterVO = {
  id: string
  novel_id: string
  volume_id: string
  slug: string
  number: string
  title: string
  summary: string
  body?: string
  status: ChapterStatus
  word_count: number
  reading_minutes: number
  sort_order: number
}

type NovelDetailVO = {
  novel: NovelVO
  volumes: VolumeVO[]
  chapters: ChapterVO[]
}

export class NovelApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = "NovelApiError"
    this.status = status
  }
}

function buildApiUrl(path: string, query?: Record<string, string | number | undefined>) {
  const url = new URL(`${API_BASE_URL.replace(/\/$/, "")}/${path.replace(/^\//, "")}`)

  if (TENANT_ID) {
    url.searchParams.set("tenant_id", TENANT_ID)
  }

  Object.entries(query ?? {}).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      url.searchParams.set(key, String(value))
    }
  })

  return url.toString()
}

async function requestApi<T>(path: string, query?: Record<string, string | number | undefined>) {
  const response = await fetch(buildApiUrl(path, query), {
    cache: "no-store",
    headers: {
      Accept: "application/json",
    },
  })

  let payload: ApiResponse<T> | undefined
  try {
    payload = (await response.json()) as ApiResponse<T>
  } catch {
    payload = undefined
  }

  if (!response.ok || (payload && payload.code !== 0)) {
    const message = payload?.error || payload?.msg || `请求失败: ${response.status}`
    throw new NovelApiError(message, response.status)
  }

  if (!payload || payload.data === undefined) {
    throw new NovelApiError("后端响应缺少 data 字段", response.status)
  }

  return payload.data
}

export async function listNovels(): Promise<NovelSummary[]> {
  const result = await requestApi<ListResult<NovelVO>>("/novels", { page: 1, page_size: 24 })
  return result.data.map(toNovelSummary)
}

export async function getNovel(id?: string): Promise<Novel | undefined> {
  const novelId = id ?? (await listNovels())[0]?.id
  if (!novelId) {
    return undefined
  }

  try {
    const detail = await requestApi<NovelDetailVO>(`/novels/${novelId}`)
    return toNovel(detail)
  } catch (error) {
    if (error instanceof NovelApiError && error.status === 404) {
      return undefined
    }
    throw error
  }
}

export async function getChapterBySlug(
  novelId: string,
  slug: string,
): Promise<Chapter | undefined> {
  try {
    const chapter = await requestApi<ChapterVO>(`/novels/${novelId}/chapters/${slug}`)
    return toChapter(chapter)
  } catch (error) {
    if (error instanceof NovelApiError && error.status === 404) {
      return undefined
    }
    throw error
  }
}

export async function getAdjacentChapter(novelId: string, slug: string): Promise<AdjacentChapter> {
  const novel = await getNovel(novelId)
  const chapterIndex = novel?.chapters.findIndex((chapter) => chapter.slug === slug) ?? -1

  if (!novel || chapterIndex < 0) {
    return {}
  }

  const previous = novel.chapters[chapterIndex - 1]
  const next = novel.chapters[chapterIndex + 1]

  return {
    previous: previous
      ? { slug: previous.slug, title: previous.title, number: previous.number }
      : undefined,
    next: next ? { slug: next.slug, title: next.title, number: next.number } : undefined,
  }
}

export async function getEditorialChapters(): Promise<EditorialChapter[]> {
  return editorialChapters
}

export async function getWorldbuildingSnapshot(): Promise<WorldbuildingSnapshot> {
  return worldbuilding
}

function toNovelSummary(item: NovelVO): NovelSummary {
  return {
    id: item.id,
    title: item.title,
    subtitle: item.subtitle,
    description: item.description,
    status: item.status,
    tags: item.tags ?? [],
    updatedAt: item.updated_at,
  }
}

function toNovel(detail: NovelDetailVO): Novel {
  const sourceVolumes = detail.volumes.map(toVolume).sort(sortBySortOrder)
  const sourceChapters = detail.chapters.map(toChapter)
  const fallbackVolumes = buildFallbackVolumes(sourceVolumes, sourceChapters)
  const volumeOrder = new Map(fallbackVolumes.map((volume, index) => [volume.id, index]))
  const chapters = sourceChapters.sort((left, right) => sortChaptersByCatalog(left, right, volumeOrder))

  return {
    id: detail.novel.id,
    title: detail.novel.title,
    subtitle: detail.novel.subtitle || "Novel Workspace",
    issue: formatIssue(detail.novel),
    heroTitle: toHeroTitle(detail.novel.title),
    description: detail.novel.description || "这部小说还没有简介。",
    stats: [
      { label: "章节", value: `${fallbackVolumes.length} 卷 · ${chapters.length} 章` },
      { label: "状态", value: formatNovelStatus(detail.novel.status) },
      { label: "最近更新", value: formatDate(detail.novel.updated_at) },
    ],
    volumes: fallbackVolumes,
    chapters,
  }
}

function toVolume(item: VolumeVO): Volume {
  return {
    id: item.id,
    title: item.title,
    subtitle: item.subtitle,
    description: item.description,
    sortOrder: item.sort_order,
  }
}

function buildFallbackVolumes(volumes: Volume[], chapters: Chapter[]) {
  if (volumes.length > 0) {
    return volumes
  }

  const volumeIds = Array.from(new Set(chapters.map((chapter) => chapter.volumeId))).filter(Boolean)
  return volumeIds.map((id, index) => ({
    id,
    title: `Vol. ${index + 1}`,
    subtitle: "未命名分卷",
    description: "这个分卷还没有简介。",
    sortOrder: index + 1,
  }))
}

function toChapter(item: ChapterVO): Chapter {
  return {
    slug: item.slug,
    volumeId: item.volume_id,
    number: item.number || String(item.sort_order).padStart(2, "0"),
    eyebrow: item.number ? `Chapter ${item.number}` : "Chapter",
    title: item.title,
    summary: item.summary,
    wordCount: item.word_count,
    readingMinutes: item.reading_minutes,
    sortOrder: item.sort_order,
    status: item.status,
    paragraphs: splitParagraphs(item.body ?? ""),
  }
}

function sortBySortOrder<T extends { sortOrder: number; title: string }>(left: T, right: T) {
  if (left.sortOrder !== right.sortOrder) {
    return left.sortOrder - right.sortOrder
  }
  return left.title.localeCompare(right.title, "zh-CN")
}

function sortChaptersByCatalog(left: Chapter, right: Chapter, volumeOrder: Map<string, number>) {
  const leftVolumeOrder = volumeOrder.get(left.volumeId) ?? Number.MAX_SAFE_INTEGER
  const rightVolumeOrder = volumeOrder.get(right.volumeId) ?? Number.MAX_SAFE_INTEGER
  if (leftVolumeOrder !== rightVolumeOrder) {
    return leftVolumeOrder - rightVolumeOrder
  }
  return sortBySortOrder(left, right)
}

function splitParagraphs(body: string) {
  return body
    .split(/\r?\n\s*\r?\n/)
    .map((paragraph) => paragraph.trim())
    .filter(Boolean)
}

function toHeroTitle(title: string) {
  const trimmed = title.trim()
  if (!trimmed) {
    return ["未命名小说"]
  }
  if (trimmed.length <= 8) {
    return [trimmed]
  }
  return trimmed.match(/.{1,8}/g) ?? [trimmed]
}

function formatIssue(item: NovelVO) {
  return `${formatNovelStatus(item.status)} · ${formatDate(item.updated_at)}`
}

function formatNovelStatus(status: NovelStatus) {
  const labels: Record<NovelStatus, string> = {
    draft: "草稿",
    serial: "连载中",
    completed: "已完结",
    archived: "已归档",
  }
  return labels[status] ?? status
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return "未记录更新时间"
  }
  return date.toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  })
}
