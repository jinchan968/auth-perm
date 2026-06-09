'use client'

import { useCallback, useEffect, useMemo, useState, type PointerEvent } from 'react'
import { ArrowLeft, ChevronDown, ChevronRight, FileArchive, Layers3, RefreshCw, Search } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { showError, showSuccess } from '@/lib/toast'
import { useTenant } from '@/lib/tenant-context'
import { batchUpdateChapterStatus, getManagedNovel, updateNovelStatus } from '@/lib/api/novel'
import type { Chapter, ChapterStatus, NovelDetail, NovelStatus, Unit, Volume } from '@/types/novel'

const novelStatusLabels: Record<NovelStatus, string> = {
  draft: '草稿',
  serial: '连载中',
  completed: '已完结',
  archived: '已归档',
}

const chapterStatusLabels: Record<ChapterStatus, string> = {
  draft: '草稿',
  review: '审核中',
  published: '已发布',
  locked: '已锁定',
}

const chapterStatusVariants: Record<ChapterStatus, 'outline' | 'warning' | 'success' | 'secondary'> = {
  draft: 'outline',
  review: 'warning',
  published: 'success',
  locked: 'secondary',
}

function sortByOrder<T extends { sort_order: number; title: string }>(items: T[]) {
  return [...items].sort((left, right) => {
    if (left.sort_order !== right.sort_order) return left.sort_order - right.sort_order
    return left.title.localeCompare(right.title, 'zh-CN')
  })
}

function formatDate(value: string) {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function volumeGroupKey(group: VolumeGroup) {
  return group.volume?.id || 'no-volume'
}

function directChapterGroupKey(groupKey: string) {
  return `${groupKey}:direct`
}

function unitGroupKey(unit: Unit | null, groupKey: string) {
  return unit?.id || `${groupKey}:unitless`
}

function isSelectableChapter(chapter: Chapter) {
  return chapter.status !== 'published'
}

interface UnitGroup {
  unit: Unit | null
  chapters: Chapter[]
}

interface VolumeGroup {
  volume: Volume | null
  unitGroups: UnitGroup[]
  directChapters: Chapter[]
  chapterCount: number
  wordCount: number
}

export function NovelDetailWorkspace({ novelId }: { novelId: string }) {
  const router = useRouter()
  const { tenantId } = useTenant()
  const [detail, setDetail] = useState<NovelDetail | null>(null)
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [selectedChapterIds, setSelectedChapterIds] = useState<Set<string>>(() => new Set())
  const [novelStatusUpdating, setNovelStatusUpdating] = useState(false)
  const [statusUpdating, setStatusUpdating] = useState(false)
  const [rangeSelectEnabled, setRangeSelectEnabled] = useState(false)
  const [rangeSelecting, setRangeSelecting] = useState(false)
  const [rangeStartChapterId, setRangeStartChapterId] = useState<string | null>(null)
  const [rangeSelectAction, setRangeSelectAction] = useState<'select' | 'deselect' | null>(null)
  const [collapsedVolumeKeys, setCollapsedVolumeKeys] = useState<Set<string>>(() => new Set())
  const [collapsedUnitKeys, setCollapsedUnitKeys] = useState<Set<string>>(() => new Set())

  const refreshDetail = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getManagedNovel(tenantId, novelId)
      setDetail(result)
    } catch (err) {
      showError(err instanceof Error ? err.message : '加载小说明细失败')
    } finally {
      setLoading(false)
    }
  }, [novelId, tenantId])

  useEffect(() => {
    if (!tenantId || !novelId) return
    refreshDetail()
  }, [novelId, refreshDetail, tenantId])

  const filteredChapters = useMemo(() => {
    const chapters = detail?.chapters || []
    const query = keyword.trim().toLowerCase()
    if (!query) return chapters
    return chapters.filter((chapter) => {
      const source = [
        chapter.title,
        chapter.slug,
        chapter.number,
        chapter.summary,
        chapter.status,
      ].join(' ').toLowerCase()
      return source.includes(query)
    })
  }, [detail?.chapters, keyword])

  const structure = useMemo<VolumeGroup[]>(() => {
    if (!detail) return []

    const chaptersByVolume = new Map<string, Chapter[]>()
    const chaptersWithoutVolume: Chapter[] = []
    filteredChapters.forEach((chapter) => {
      if (!chapter.volume_id) {
        chaptersWithoutVolume.push(chapter)
        return
      }
      const list = chaptersByVolume.get(chapter.volume_id) || []
      list.push(chapter)
      chaptersByVolume.set(chapter.volume_id, list)
    })

    const unitsByVolume = new Map<string, Unit[]>()
    detail.units.forEach((unit) => {
      const list = unitsByVolume.get(unit.volume_id) || []
      list.push(unit)
      unitsByVolume.set(unit.volume_id, list)
    })

    const groups: VolumeGroup[] = sortByOrder(detail.volumes).map((volume) => {
      const volumeChapters = sortByOrder(chaptersByVolume.get(volume.id) || [])
      const unitGroups = sortByOrder(unitsByVolume.get(volume.id) || [])
        .map((unit) => ({
          unit,
          chapters: volumeChapters.filter((chapter) => chapter.unit_id === unit.id),
        }))
        .filter((group) => group.chapters.length > 0 || !keyword.trim())

      const directChapters = volumeChapters.filter((chapter) => !chapter.unit_id)
      return {
        volume,
        unitGroups,
        directChapters,
        chapterCount: volumeChapters.length,
        wordCount: volumeChapters.reduce((sum, chapter) => sum + chapter.word_count, 0),
      }
    }).filter((group) => group.chapterCount > 0 || !keyword.trim())

    if (chaptersWithoutVolume.length > 0) {
      groups.push({
        volume: null,
        unitGroups: [],
        directChapters: sortByOrder(chaptersWithoutVolume),
        chapterCount: chaptersWithoutVolume.length,
        wordCount: chaptersWithoutVolume.reduce((sum, chapter) => sum + chapter.word_count, 0),
      })
    }

    return groups
  }, [detail, filteredChapters, keyword])

  const stats = useMemo(() => {
    const chapters = detail?.chapters || []
    return {
      volumes: detail?.volumes.length || 0,
      units: detail?.units.length || 0,
      chapters: chapters.length,
      words: chapters.reduce((sum, chapter) => sum + chapter.word_count, 0),
    }
  }, [detail])

  const selectedChapters = useMemo(() => {
    const chapters = detail?.chapters || []
    return chapters.filter((chapter) => selectedChapterIds.has(chapter.id))
  }, [detail?.chapters, selectedChapterIds])

  const selectableChapterIds = useMemo(() => {
    const chapters = detail?.chapters || []
    return new Set(chapters.filter(isSelectableChapter).map((chapter) => chapter.id))
  }, [detail?.chapters])

  const visibleChapterIds = useMemo(() => {
    return structure.flatMap((volumeGroup) => {
      const groupKey = volumeGroupKey(volumeGroup)
      if (collapsedVolumeKeys.has(groupKey)) return []

      const directIds = collapsedUnitKeys.has(directChapterGroupKey(groupKey))
        ? []
        : volumeGroup.directChapters.filter(isSelectableChapter).map((chapter) => chapter.id)
      const unitIds = volumeGroup.unitGroups.flatMap((unitGroup) => (
        collapsedUnitKeys.has(unitGroupKey(unitGroup.unit, groupKey))
          ? []
          : unitGroup.chapters.filter(isSelectableChapter).map((chapter) => chapter.id)
      ))
      return [...directIds, ...unitIds]
    })
  }, [collapsedUnitKeys, collapsedVolumeKeys, structure])

  useEffect(() => {
    setSelectedChapterIds((current) => {
      const next = new Set<string>()
      current.forEach((chapterId) => {
        if (selectableChapterIds.has(chapterId)) {
          next.add(chapterId)
        }
      })
      return next.size === current.size ? current : next
    })
  }, [selectableChapterIds])

  useEffect(() => {
    if (!rangeSelecting) return

    function stopRangeSelect() {
      setRangeSelecting(false)
      setRangeStartChapterId(null)
      setRangeSelectAction(null)
    }

    window.addEventListener('pointerup', stopRangeSelect)
    window.addEventListener('pointercancel', stopRangeSelect)
    return () => {
      window.removeEventListener('pointerup', stopRangeSelect)
      window.removeEventListener('pointercancel', stopRangeSelect)
    }
  }, [rangeSelecting])

  function toggleChapter(chapterId: string, checked: boolean) {
    if (checked && !selectableChapterIds.has(chapterId)) return
    setSelectedChapterIds((current) => {
      const next = new Set(current)
      if (checked) {
        next.add(chapterId)
      } else {
        next.delete(chapterId)
      }
      return next
    })
  }

  function toggleChapters(chapterIds: string[], checked: boolean) {
    const selectableIds = chapterIds.filter((chapterId) => selectableChapterIds.has(chapterId))
    setSelectedChapterIds((current) => {
      const next = new Set(current)
      selectableIds.forEach((chapterId) => {
        if (checked) {
          next.add(chapterId)
        } else {
          next.delete(chapterId)
        }
      })
      return next
    })
  }

  function applyChapterRange(startChapterId: string, endChapterId: string, action: 'select' | 'deselect') {
    const startIndex = visibleChapterIds.indexOf(startChapterId)
    const endIndex = visibleChapterIds.indexOf(endChapterId)
    if (startIndex < 0 || endIndex < 0) return

    const [from, to] = startIndex <= endIndex ? [startIndex, endIndex] : [endIndex, startIndex]
    const rangeIds = visibleChapterIds.slice(from, to + 1)
    setSelectedChapterIds((current) => {
      const next = new Set(current)
      rangeIds.forEach((chapterId) => {
        if (action === 'select') {
          next.add(chapterId)
        } else {
          next.delete(chapterId)
        }
      })
      return next
    })
  }

  function beginRangeSelect(chapterId: string) {
    if (!rangeSelectEnabled || statusUpdating) return
    if (!selectableChapterIds.has(chapterId)) return
    const action = selectedChapterIds.has(chapterId) ? 'deselect' : 'select'
    setRangeStartChapterId(chapterId)
    setRangeSelectAction(action)
    setRangeSelecting(true)
    applyChapterRange(chapterId, chapterId, action)
  }

  function extendRangeSelect(chapterId: string) {
    if (!rangeSelectEnabled || !rangeSelecting || !rangeStartChapterId || !rangeSelectAction || statusUpdating) return
    if (!selectableChapterIds.has(chapterId)) return
    applyChapterRange(rangeStartChapterId, chapterId, rangeSelectAction)
  }

  function toggleRangeSelectMode() {
    setRangeSelectEnabled((enabled) => !enabled)
    setRangeSelecting(false)
    setRangeStartChapterId(null)
    setRangeSelectAction(null)
  }

  function toggleVolumeCollapse(groupKey: string) {
    setCollapsedVolumeKeys((current) => {
      const next = new Set(current)
      if (next.has(groupKey)) {
        next.delete(groupKey)
      } else {
        next.add(groupKey)
      }
      return next
    })
  }

  function toggleUnitCollapse(groupKey: string) {
    setCollapsedUnitKeys((current) => {
      const next = new Set(current)
      if (next.has(groupKey)) {
        next.delete(groupKey)
      } else {
        next.add(groupKey)
      }
      return next
    })
  }

  async function handleBulkStatus(nextStatus: ChapterStatus) {
    if (selectedChapters.length === 0) {
      showError('请先选择章节')
      return
    }

    const targets = selectedChapters.filter((chapter) => chapter.status !== nextStatus)
    if (targets.length === 0) {
      showError(nextStatus === 'published' ? '选中章节已经是已发布状态' : '选中章节已经是草稿状态')
      return
    }

    setStatusUpdating(true)
    try {
      const result = await batchUpdateChapterStatus(
        tenantId,
        novelId,
        targets.map((chapter) => chapter.id),
        nextStatus,
        nextStatus === 'published' ? '管理端批量发布' : '管理端批量退回草稿',
      )
      showSuccess(nextStatus === 'published' ? `已发布 ${result.updated} 章` : `已退回草稿 ${result.updated} 章`)
      setSelectedChapterIds(new Set())
      await refreshDetail()
    } catch (err) {
      showError(err instanceof Error ? err.message : '批量更新章节状态失败')
    } finally {
      setStatusUpdating(false)
    }
  }

  function goImport() {
    router.push(`/novels/import?novel_id=${encodeURIComponent(novelId)}`)
  }

  async function handleNovelStatus(nextStatus: NovelStatus) {
    if (!detail) return
    setNovelStatusUpdating(true)
    try {
      const updated = await updateNovelStatus(tenantId, novelId, nextStatus)
      setDetail((current) => current ? { ...current, novel: updated } : current)
      showSuccess(nextStatus === 'completed' ? '小说已设为完结' : '小说已恢复连载')
    } catch (err) {
      showError(err instanceof Error ? err.message : '更新小说状态失败')
    } finally {
      setNovelStatusUpdating(false)
    }
  }

  if (loading && !detail) {
    return (
      <div className="rounded-lg border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
        正在加载小说明细...
      </div>
    )
  }

  if (!detail) {
    return (
      <div className="rounded-lg border border-dashed border-border p-8 text-center">
        <div className="text-sm font-medium text-foreground">未找到小说明细</div>
        <Button className="mt-4" variant="outline" onClick={() => router.push('/novels')}>
          <ArrowLeft className="h-4 w-4" />
          返回列表
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => router.push('/novels')}>
              <ArrowLeft className="h-4 w-4" />
              返回
            </Button>
            <Badge variant="outline">{novelStatusLabels[detail.novel.status]}</Badge>
          </div>
          <h1 className="mt-4 text-3xl font-semibold tracking-tight text-foreground">{detail.novel.title}</h1>
          <p className="mt-2 max-w-3xl text-sm text-muted-foreground">
            {detail.novel.subtitle || detail.novel.description || '暂无简介'}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button variant="outline" onClick={refreshDetail} disabled={loading}>
            <RefreshCw className="h-4 w-4" />
            刷新
          </Button>
          {detail.novel.status === 'completed' ? (
            <Button
              variant="outline"
              onClick={() => handleNovelStatus('serial')}
              disabled={novelStatusUpdating}
            >
              {novelStatusUpdating ? '更新中...' : '恢复连载'}
            </Button>
          ) : (
            <Button
              variant="outline"
              onClick={() => handleNovelStatus('completed')}
              disabled={novelStatusUpdating}
            >
              {novelStatusUpdating ? '更新中...' : '设为完结'}
            </Button>
          )}
          <Button onClick={goImport}>
            <FileArchive className="h-4 w-4" />
            继续导入
          </Button>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="分卷" value={stats.volumes} />
        <StatCard label="单元" value={stats.units} />
        <StatCard label="章节" value={stats.chapters} />
        <StatCard label="字数" value={stats.words} />
      </div>

      {detail.novel.status !== 'serial' && detail.novel.status !== 'completed' ? (
        <div className="rounded-lg border border-border bg-muted/40 px-4 py-3 text-sm text-muted-foreground">
          当前小说状态是“{novelStatusLabels[detail.novel.status]}”。章节发布后仍需要小说状态为“连载中”或“已完结”，才会出现在阅读端公开页面。
        </div>
      ) : null}

      <Card>
        <CardHeader className="gap-4 sm:flex-row sm:items-center sm:justify-between">
          <CardTitle className="flex items-center gap-2">
            <Layers3 className="h-5 w-5 text-primary" />
            章节结构核对
          </CardTitle>
          <div className="relative w-full sm:w-96">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={keyword}
              onChange={(event) => setKeyword(event.target.value)}
              placeholder="搜索章节标题、编号、slug"
              className="pl-9"
            />
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col gap-3 rounded-lg border border-border bg-muted/30 p-3 lg:flex-row lg:items-center lg:justify-between">
            <div className="text-sm text-muted-foreground">
              已选择 <span className="font-medium text-foreground">{selectedChapters.length}</span> 章
            </div>
            <div className="flex flex-wrap gap-2">
              <Button
                size="sm"
                onClick={() => handleBulkStatus('published')}
                disabled={statusUpdating || selectedChapters.length === 0}
              >
                {statusUpdating ? '处理中...' : '发布选中'}
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleBulkStatus('draft')}
                disabled={statusUpdating || selectedChapters.length === 0}
              >
                退回草稿
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={toggleRangeSelectMode}
                disabled={statusUpdating || visibleChapterIds.length === 0}
              >
                {rangeSelectEnabled ? '关闭范围选择' : '范围选择'}
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => setSelectedChapterIds(new Set())}
                disabled={statusUpdating || selectedChapters.length === 0}
              >
                清空选择
              </Button>
            </div>
          </div>
          {structure.length === 0 ? (
            <div className="rounded-lg border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
              {detail.chapters.length === 0 ? '暂无章节，导入 zip 后可在这里核对识别结果。' : '没有匹配的章节。'}
            </div>
          ) : (
            structure.map((volumeGroup) => (
              <VolumeSection
                key={volumeGroup.volume?.id || 'no-volume'}
                group={volumeGroup}
                selectedChapterIds={selectedChapterIds}
                statusUpdating={statusUpdating}
                rangeSelectEnabled={rangeSelectEnabled}
                rangeSelecting={rangeSelecting}
                onToggleChapter={toggleChapter}
                onToggleChapters={toggleChapters}
                onRangeStart={beginRangeSelect}
                onRangeEnter={extendRangeSelect}
                collapsedVolumeKeys={collapsedVolumeKeys}
                collapsedUnitKeys={collapsedUnitKeys}
                onToggleVolumeCollapse={toggleVolumeCollapse}
                onToggleUnitCollapse={toggleUnitCollapse}
              />
            ))
          )}
        </CardContent>
      </Card>

      <div className="text-xs text-muted-foreground">
        最后更新：{formatDate(detail.novel.updated_at)}
      </div>
    </div>
  )
}

interface ChapterSelectionProps {
  selectedChapterIds: Set<string>
  statusUpdating: boolean
  rangeSelectEnabled: boolean
  rangeSelecting: boolean
  onToggleChapter: (chapterId: string, checked: boolean) => void
  onToggleChapters: (chapterIds: string[], checked: boolean) => void
  onRangeStart: (chapterId: string) => void
  onRangeEnter: (chapterId: string) => void
  collapsedVolumeKeys: Set<string>
  collapsedUnitKeys: Set<string>
  onToggleVolumeCollapse: (groupKey: string) => void
  onToggleUnitCollapse: (groupKey: string) => void
}

function VolumeSection({
  group,
  selectedChapterIds,
  statusUpdating,
  rangeSelectEnabled,
  rangeSelecting,
  onToggleChapter,
  onToggleChapters,
  onRangeStart,
  onRangeEnter,
  collapsedVolumeKeys,
  collapsedUnitKeys,
  onToggleVolumeCollapse,
  onToggleUnitCollapse,
}: { group: VolumeGroup } & ChapterSelectionProps) {
  const groupKey = volumeGroupKey(group)
  const collapsed = collapsedVolumeKeys.has(groupKey)

  return (
    <section className="rounded-lg border border-border">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border bg-muted/40 px-4 py-3">
        <div className="flex min-w-0 items-start gap-2">
          <Button
            size="sm"
            variant="outline"
            className="h-7 w-7 shrink-0 p-0"
            onClick={() => onToggleVolumeCollapse(groupKey)}
            aria-label={collapsed ? '展开分卷' : '折叠分卷'}
          >
            {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
          </Button>
          <div className="min-w-0">
            <h2 className="truncate text-base font-semibold text-foreground">
              {group.volume?.title || '未绑定分卷'}
            </h2>
            {group.volume?.subtitle ? (
              <p className="mt-1 truncate text-xs text-muted-foreground">{group.volume.subtitle}</p>
            ) : null}
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Badge variant="outline">{group.chapterCount} 章</Badge>
          <Badge variant="outline">{group.wordCount} 字</Badge>
        </div>
      </div>

      {!collapsed ? (
        <div className="divide-y divide-border">
        {group.directChapters.length > 0 ? (
          <ChapterTable
            title={group.unitGroups.length > 0 ? '未绑定单元' : ''}
            chapters={group.directChapters}
            groupKey={directChapterGroupKey(groupKey)}
            selectedChapterIds={selectedChapterIds}
            statusUpdating={statusUpdating}
            rangeSelectEnabled={rangeSelectEnabled}
            rangeSelecting={rangeSelecting}
            onToggleChapter={onToggleChapter}
            onToggleChapters={onToggleChapters}
            onRangeStart={onRangeStart}
            onRangeEnter={onRangeEnter}
            collapsedVolumeKeys={collapsedVolumeKeys}
            collapsedUnitKeys={collapsedUnitKeys}
            onToggleVolumeCollapse={onToggleVolumeCollapse}
            onToggleUnitCollapse={onToggleUnitCollapse}
          />
        ) : null}
        {group.unitGroups.map((unitGroup) => (
          <ChapterTable
            key={unitGroup.unit?.id || 'direct'}
            title={unitGroup.unit?.title || '未绑定单元'}
            chapters={unitGroup.chapters}
            groupKey={unitGroupKey(unitGroup.unit, groupKey)}
            selectedChapterIds={selectedChapterIds}
            statusUpdating={statusUpdating}
            rangeSelectEnabled={rangeSelectEnabled}
            rangeSelecting={rangeSelecting}
            onToggleChapter={onToggleChapter}
            onToggleChapters={onToggleChapters}
            onRangeStart={onRangeStart}
            onRangeEnter={onRangeEnter}
            collapsedVolumeKeys={collapsedVolumeKeys}
            collapsedUnitKeys={collapsedUnitKeys}
            onToggleVolumeCollapse={onToggleVolumeCollapse}
            onToggleUnitCollapse={onToggleUnitCollapse}
          />
        ))}
        </div>
      ) : null}
    </section>
  )
}

function ChapterTable({
  title,
  groupKey,
  chapters,
  selectedChapterIds,
  statusUpdating,
  rangeSelectEnabled,
  rangeSelecting,
  onToggleChapter,
  onToggleChapters,
  onRangeStart,
  onRangeEnter,
  collapsedUnitKeys,
  onToggleUnitCollapse,
}: { title: string; groupKey: string; chapters: Chapter[] } & ChapterSelectionProps) {
  const sortedChapters = sortByOrder(chapters)
  const chapterIds = sortedChapters.filter(isSelectableChapter).map((chapter) => chapter.id)
  const allSelected = chapterIds.length > 0 && chapterIds.every((chapterId) => selectedChapterIds.has(chapterId))
  const collapsed = collapsedUnitKeys.has(groupKey)

  function handleRangePointerMove(event: PointerEvent<HTMLDivElement>) {
    if (!rangeSelectEnabled || !rangeSelecting || statusUpdating) return
    const target = document.elementFromPoint(event.clientX, event.clientY)
    const row = target?.closest('[data-chapter-row-id]') as HTMLElement | null
    const chapterId = row?.dataset.chapterRowId
    if (chapterId) {
      onRangeEnter(chapterId)
    }
  }

  return (
    <div className="p-4">
      {title ? (
        <div className="mb-3 flex items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-2">
            <Button
              size="sm"
              variant="outline"
              className="h-7 w-7 shrink-0 p-0"
              onClick={() => onToggleUnitCollapse(groupKey)}
              aria-label={collapsed ? '展开单元' : '折叠单元'}
            >
              {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
            </Button>
            <h3 className="truncate text-sm font-medium text-foreground">{title}</h3>
          </div>
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-2 text-xs text-muted-foreground">
              <input
                type="checkbox"
                className="h-4 w-4 accent-primary"
                checked={allSelected}
                disabled={statusUpdating || collapsed || chapterIds.length === 0}
                onChange={(event) => onToggleChapters(chapterIds, event.target.checked)}
              />
              全选本组
            </label>
            <Badge variant="secondary">{chapters.length} 章</Badge>
          </div>
        </div>
      ) : null}
      {collapsed ? null : (
      <div className="overflow-x-auto" onPointerMove={handleRangePointerMove}>
        <table className="w-full min-w-[840px] text-left text-sm">
          <thead className="text-xs uppercase text-muted-foreground">
            <tr>
              <th className="px-3 py-2 font-medium">
                {!title ? (
                  <input
                    type="checkbox"
                    className="h-4 w-4 accent-primary"
                    checked={allSelected}
                    disabled={statusUpdating || chapterIds.length === 0}
                    onChange={(event) => onToggleChapters(chapterIds, event.target.checked)}
                    aria-label="全选章节"
                  />
                ) : null}
              </th>
              <th className="px-3 py-2 font-medium">顺序</th>
              <th className="px-3 py-2 font-medium">编号</th>
              <th className="px-3 py-2 font-medium">章节</th>
              <th className="px-3 py-2 font-medium">Slug</th>
              <th className="px-3 py-2 font-medium">状态</th>
              <th className="px-3 py-2 text-right font-medium">字数</th>
            </tr>
          </thead>
          <tbody>
            {sortedChapters.map((chapter) => {
              const selectable = isSelectableChapter(chapter)
              return (
                <tr
                  key={chapter.id}
                  data-chapter-row-id={chapter.id}
                  className={`border-t border-border/70 ${selectedChapterIds.has(chapter.id) ? 'bg-primary/5' : ''} ${!selectable ? 'bg-muted/30 text-muted-foreground' : ''} ${rangeSelectEnabled && selectable ? 'cursor-crosshair select-none touch-none hover:bg-primary/10' : ''}`}
                  onPointerDown={(event) => {
                    if (!rangeSelectEnabled || statusUpdating || !selectable) return
                    event.preventDefault()
                    onRangeStart(chapter.id)
                  }}
                  onPointerEnter={() => {
                    if (!selectable) return
                    onRangeEnter(chapter.id)
                  }}
                >
                <td className="px-3 py-3">
                  <input
                    type="checkbox"
                    className="h-4 w-4 accent-primary"
                    checked={selectedChapterIds.has(chapter.id)}
                    disabled={statusUpdating || !selectable}
                    onChange={(event) => {
                      if (rangeSelectEnabled) return
                      onToggleChapter(chapter.id, event.target.checked)
                    }}
                    aria-label={selectable ? `选择章节 ${chapter.title}` : `章节 ${chapter.title} 已发布，不可选择`}
                  />
                </td>
                <td className="px-3 py-3 text-muted-foreground">{chapter.sort_order}</td>
                <td className="px-3 py-3 font-mono text-xs text-muted-foreground">{chapter.number || '-'}</td>
                <td className="max-w-[24rem] px-3 py-3">
                  <div className="truncate font-medium text-foreground">{chapter.title}</div>
                  {chapter.summary ? (
                    <div className="mt-1 truncate text-xs text-muted-foreground">{chapter.summary}</div>
                  ) : null}
                </td>
                <td className="px-3 py-3 font-mono text-xs text-muted-foreground">{chapter.slug}</td>
                <td className="px-3 py-3">
                  <Badge variant={chapterStatusVariants[chapter.status]}>
                    {chapterStatusLabels[chapter.status]}
                  </Badge>
                </td>
                <td className="px-3 py-3 text-right text-muted-foreground">{chapter.word_count}</td>
              </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      )}
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <Card>
      <CardContent className="p-5">
        <div className="text-sm text-muted-foreground">{label}</div>
        <div className="mt-2 text-3xl font-semibold tracking-tight text-foreground">{value}</div>
      </CardContent>
    </Card>
  )
}
