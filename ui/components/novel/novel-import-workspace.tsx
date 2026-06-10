'use client'

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { AlertCircle, BookOpen, FileArchive, FileText, RefreshCw, Upload } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { showError, showSuccess } from '@/lib/toast'
import { useTenant } from '@/lib/tenant-context'
import {
  createNovel,
  getImportTask,
  importMarkdownBundle,
  importMarkdownChapter,
  inspectMarkdownBundle,
  listMyNovels,
  listUnits,
  listVolumes,
} from '@/lib/api/novel'
import type {
  ChapterStatus,
  MarkdownBundleImportResult,
  MarkdownBundleInspectResult,
  NovelListItem,
  Unit,
  Volume,
} from '@/types/novel'
import { IMPORT_TASK_STATUS } from '@/types/novel'

const statusLabels: Record<string, string> = {
  draft: '草稿',
  serial: '连载中',
  completed: '已完结',
  archived: '已归档',
  review: '审核中',
  published: '已发布',
  locked: '已锁定',
}

export function NovelImportWorkspace() {
  const { tenantId } = useTenant()
  const [novels, setNovels] = useState<NovelListItem[]>([])
  const [novelsLoading, setNovelsLoading] = useState(false)
  const [selectedNovelId, setSelectedNovelId] = useState('')
  const [createForm, setCreateForm] = useState({
    title: '',
    subtitle: '',
    description: '',
    status: 'serial' as 'draft' | 'serial',
  })
  const [creating, setCreating] = useState(false)

  const [zipFile, setZipFile] = useState<File | null>(null)
  const [inspecting, setInspecting] = useState(false)
  const [inspectResult, setInspectResult] = useState<MarkdownBundleInspectResult | null>(null)
  const [importing, setImporting] = useState(false)
  const [bundleImported, setBundleImported] = useState(false)
  const [importResult, setImportResult] = useState<MarkdownBundleImportResult | null>(null)
  const [requestedNovelId, setRequestedNovelId] = useState('')
  const pollRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const [volumes, setVolumes] = useState<Volume[]>([])
  const [units, setUnits] = useState<Unit[]>([])
  const [singleFile, setSingleFile] = useState<File | null>(null)
  const [singleForm, setSingleForm] = useState({
    volumeId: '',
    unitId: 'none',
    status: 'draft' as ChapterStatus,
  })
  const [singleImporting, setSingleImporting] = useState(false)

  const selectedNovel = useMemo(
    () => novels.find((novel) => novel.id === selectedNovelId),
    [novels, selectedNovelId],
  )

  const importZipDisabledReason = useMemo(() => {
    if (!zipFile) return '请先选择 zip 文件'
    if (!selectedNovelId) return '请先在左侧选择或创建目标小说'
    if (inspectResult && inspectResult.valid === 0) return '识别结果中没有可导入章节'
    if (bundleImported) return '本次 zip 已导入完成，重新选择文件后可再次导入'
    return ''
  }, [bundleImported, inspectResult, selectedNovelId, zipFile])

  useEffect(() => {
    const query = new URLSearchParams(window.location.search)
    setRequestedNovelId(query.get('novel_id') || '')
  }, [])

  const refreshNovels = useCallback(async (nextSelectedId?: string) => {
    setNovelsLoading(true)
    try {
      const data = await listMyNovels(tenantId)
      setNovels(data.data)
      setSelectedNovelId((current) => {
        const preferredId = nextSelectedId || requestedNovelId || current
        if (preferredId && data.data.some((novel) => novel.id === preferredId)) {
          return preferredId
        }
        return data.data[0]?.id || ''
      })
    } catch (err) {
      showError(err instanceof Error ? err.message : '加载小说列表失败')
    } finally {
      setNovelsLoading(false)
    }
  }, [requestedNovelId, tenantId])

  useEffect(() => {
    return () => {
      if (pollRef.current) clearTimeout(pollRef.current)
    }
  }, [])

  useEffect(() => {
    if (!tenantId) return
    refreshNovels()
  }, [tenantId, refreshNovels])

  useEffect(() => {
    if (!tenantId || !selectedNovelId) {
      setVolumes([])
      setUnits([])
      return
    }

    let cancelled = false
    async function loadStructure() {
      try {
        const [nextVolumes, nextUnits] = await Promise.all([
          listVolumes(tenantId, selectedNovelId),
          listUnits(tenantId, selectedNovelId),
        ])
        if (cancelled) return
        setVolumes(nextVolumes)
        setUnits(nextUnits)
        setSingleForm((current) => ({
          ...current,
          volumeId: current.volumeId || nextVolumes[0]?.id || '',
        }))
      } catch (err) {
        if (!cancelled) {
          showError(err instanceof Error ? err.message : '加载分卷失败')
        }
      }
    }
    loadStructure()

    return () => {
      cancelled = true
    }
  }, [tenantId, selectedNovelId])

  async function handleCreateNovel() {
    const title = createForm.title.trim()
    if (!title) {
      showError('请填写小说标题')
      return
    }

    setCreating(true)
    try {
      const created = await createNovel(tenantId, {
        title,
        subtitle: createForm.subtitle.trim(),
        description: createForm.description.trim(),
        status: createForm.status,
      })
      showSuccess('小说已创建')
      setCreateForm({ title: '', subtitle: '', description: '', status: 'serial' })
      await refreshNovels(created.id)
    } catch (err) {
      showError(err instanceof Error ? err.message : '创建小说失败')
    } finally {
      setCreating(false)
    }
  }

  async function handleInspectZip() {
    if (!zipFile) {
      showError('请先选择 zip 文件')
      return
    }

    setInspecting(true)
    setInspectResult(null)
    setImportResult(null)
    setBundleImported(false)
    try {
      const result = await inspectMarkdownBundle(tenantId, zipFile)
      setInspectResult(result)
      showSuccess(`已识别 ${result.valid} 个 Markdown 文件`)
    } catch (err) {
      showError(err instanceof Error ? err.message : '识别 zip 失败')
    } finally {
      setInspecting(false)
    }
  }

  async function handleImportZip() {
    if (!selectedNovelId) {
      showError('请先选择小说')
      return
    }
    if (!zipFile) {
      showError('请先选择 zip 文件')
      return
    }
    if (importZipDisabledReason) {
      showError(importZipDisabledReason)
      return
    }

    setImporting(true)
    try {
      const { task_id } = await importMarkdownBundle(tenantId, selectedNovelId, zipFile)

      const poll = async () => {
        try {
          const task = await getImportTask(task_id)
          if (task.status === IMPORT_TASK_STATUS.SUCCESS && task.result) {
            setImportResult(task.result)
            setBundleImported(true)
            showSuccess(`导入完成：新增 ${task.result.created}，更新 ${task.result.updated}，跳过 ${task.result.skipped}`)
            setImporting(false)
            await refreshNovels(selectedNovelId)
          } else if (task.status === IMPORT_TASK_STATUS.FAILED) {
            showError(task.error || '导入失败')
            setImporting(false)
          } else {
            pollRef.current = setTimeout(poll, 2000)
          }
        } catch {
          showError('获取导入状态失败')
          setImporting(false)
        }
      }
      poll()
    } catch (err) {
      showError(err instanceof Error ? err.message : '导入 zip 失败')
      setImporting(false)
    }
  }

  async function handleImportSingleMarkdown() {
    if (!selectedNovelId) {
      showError('请先选择小说')
      return
    }
    if (!singleFile) {
      showError('请先选择 Markdown 文件')
      return
    }
    if (!singleForm.volumeId) {
      showError('单章导入需要已有分卷；批量 zip 导入可自动创建分卷')
      return
    }

    setSingleImporting(true)
    try {
      await importMarkdownChapter(tenantId, selectedNovelId, {
        file: singleFile,
        volumeId: singleForm.volumeId,
        unitId: singleForm.unitId === 'none' ? undefined : singleForm.unitId,
        status: singleForm.status,
      })
      showSuccess('单章 Markdown 已导入')
      setSingleFile(null)
      await refreshNovels(selectedNovelId)
    } catch (err) {
      showError(err instanceof Error ? err.message : '导入 Markdown 失败')
    } finally {
      setSingleImporting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-slate-950">小说导入</h1>
          <p className="mt-2 text-sm text-slate-500">
            上传 Markdown 或 zip 文件树，系统会按目录名识别卷、单元和章节并写入数据库。
          </p>
        </div>
        <Button variant="outline" onClick={() => refreshNovels()} disabled={novelsLoading}>
          <RefreshCw className="h-4 w-4" />
          刷新
        </Button>
      </div>

      <div className="grid gap-6 xl:grid-cols-[22rem_minmax(0,1fr)]">
        <section className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <BookOpen className="h-5 w-5 text-primary" />
                选择小说
              </CardTitle>
              <CardDescription>导入内容会写入选中的小说。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <Select value={selectedNovelId} onValueChange={setSelectedNovelId}>
                <SelectTrigger>
                  <SelectValue placeholder={novelsLoading ? '加载中...' : '选择小说'} />
                </SelectTrigger>
                <SelectContent>
                  {novels.map((novel) => (
                    <SelectItem key={novel.id} value={novel.id}>
                      {novel.title}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              {selectedNovel ? (
                <div className="rounded-lg border border-slate-200 bg-slate-50 p-4 text-sm">
                  <div className="flex items-center justify-between gap-3">
                    <div className="min-w-0">
                      <div className="truncate font-medium text-slate-900">{selectedNovel.title}</div>
                      <div className="mt-1 truncate text-slate-500">{selectedNovel.subtitle || '暂无副标题'}</div>
                    </div>
                    <Badge variant="outline">{statusLabels[selectedNovel.status] || selectedNovel.status}</Badge>
                  </div>
                  <p className="mt-3 line-clamp-3 text-slate-500">
                    {selectedNovel.description || '暂无简介'}
                  </p>
                </div>
              ) : (
                <div className="rounded-lg border border-dashed border-slate-300 p-4 text-sm text-slate-500">
                  当前租户下还没有可导入的小说。
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>新建小说</CardTitle>
              <CardDescription>还没有目标小说时，可先创建后导入。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="novel-title">标题</Label>
                <Input
                  id="novel-title"
                  value={createForm.title}
                  onChange={(event) => setCreateForm({ ...createForm, title: event.target.value })}
                  placeholder="例如：旧城回声"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="novel-subtitle">副标题</Label>
                <Input
                  id="novel-subtitle"
                  value={createForm.subtitle}
                  onChange={(event) => setCreateForm({ ...createForm, subtitle: event.target.value })}
                  placeholder="可选"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="novel-description">简介</Label>
                <Textarea
                  id="novel-description"
                  value={createForm.description}
                  onChange={(event) => setCreateForm({ ...createForm, description: event.target.value })}
                  rows={4}
                  placeholder="写一点给读者看的说明"
                />
              </div>
              <Button className="w-full" onClick={handleCreateNovel} disabled={creating}>
                {creating ? '创建中...' : '创建小说'}
              </Button>
            </CardContent>
          </Card>
        </section>

        <section className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileArchive className="h-5 w-5 text-primary" />
                zip 文件树导入
              </CardTitle>
              <CardDescription>
                支持“卷/单元/章.md”、“卷/章.md”、“章.md”。上传内容由后端内存读取，不会持久保存到磁盘。
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              <div className="rounded-xl border-2 border-dashed border-slate-300 bg-white p-6">
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <div className="font-medium text-slate-900">
                      {zipFile ? zipFile.name : '选择 Markdown zip 包'}
                    </div>
                    <p className="mt-1 text-sm text-slate-500">
                      最大 30MB，最多 1000 个 Markdown 文件，解压后最大 50MB。
                    </p>
                  </div>
                  <Input
                    type="file"
                    accept=".zip,application/zip,application/x-zip-compressed"
                    className="max-w-sm"
                    onChange={(event) => {
                      setZipFile(event.target.files?.[0] || null)
                      setInspectResult(null)
                      setImportResult(null)
                      setBundleImported(false)
                    }}
                  />
                </div>
              </div>

              <div className="flex flex-wrap gap-3">
                <Button variant="outline" onClick={handleInspectZip} disabled={inspecting || !zipFile}>
                  <FileText className="h-4 w-4" />
                  {inspecting ? '识别中...' : '识别目录'}
                </Button>
                <Button onClick={handleImportZip} disabled={importing || Boolean(importZipDisabledReason)}>
                  <Upload className="h-4 w-4" />
                  {importing ? '导入中...' : '确认导入'}
                </Button>
              </div>

              {importZipDisabledReason ? (
                <p className="text-sm text-slate-500">{importZipDisabledReason}</p>
              ) : null}

              {inspectResult ? <InspectResultPanel result={inspectResult} /> : null}
              {importResult ? <ImportResultPanel result={importResult} /> : null}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>单章 Markdown 导入</CardTitle>
              <CardDescription>适合补录单章；需要目标小说已存在分卷。</CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4 lg:grid-cols-[1fr_1fr_auto] lg:items-end">
              <div className="space-y-2">
                <Label>分卷</Label>
                <Select
                  value={singleForm.volumeId}
                  onValueChange={(volumeId) => setSingleForm({ ...singleForm, volumeId, unitId: 'none' })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择分卷" />
                  </SelectTrigger>
                  <SelectContent>
                    {volumes.map((volume) => (
                      <SelectItem key={volume.id} value={volume.id}>
                        {volume.title}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>单元</Label>
                <Select
                  value={singleForm.unitId}
                  onValueChange={(unitId) => setSingleForm({ ...singleForm, unitId })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="可选" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">不绑定单元</SelectItem>
                    {units
                      .filter((unit) => unit.volume_id === singleForm.volumeId)
                      .map((unit) => (
                        <SelectItem key={unit.id} value={unit.id}>
                          {unit.title}
                        </SelectItem>
                      ))}
                  </SelectContent>
                </Select>
              </div>
              <Input
                type="file"
                accept=".md,text/markdown,text/plain"
                onChange={(event) => setSingleFile(event.target.files?.[0] || null)}
              />
              <div className="lg:col-span-3">
                <Button
                  variant="outline"
                  onClick={handleImportSingleMarkdown}
                  disabled={singleImporting || !singleFile || !selectedNovelId}
                >
                  {singleImporting ? '导入中...' : '导入单章'}
                </Button>
              </div>
            </CardContent>
          </Card>
        </section>
      </div>
    </div>
  )
}

function InspectResultPanel({ result }: { result: MarkdownBundleInspectResult }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
      <div className="flex flex-wrap items-center gap-2 text-sm">
        <Badge>有效 {result.valid}</Badge>
        <Badge variant="outline">跳过 {result.skipped}</Badge>
        <Badge variant="outline">卷 {result.volumes.length}</Badge>
        <Badge variant="outline">单元 {result.units.length}</Badge>
        <span className="text-slate-500">识别策略：目录名规则</span>
      </div>
      <div className="mt-4 overflow-x-auto">
        <table className="min-w-[760px] w-full text-left text-sm">
          <thead className="text-xs uppercase text-slate-500">
            <tr>
              <th className="px-3 py-2">路径</th>
              <th className="px-3 py-2">卷</th>
              <th className="px-3 py-2">单元</th>
              <th className="px-3 py-2">章节</th>
              <th className="px-3 py-2">Slug</th>
              <th className="px-3 py-2">字数</th>
            </tr>
          </thead>
          <tbody>
            {result.items.slice(0, 80).map((item) => (
              <tr key={item.path} className="border-t border-slate-200">
                <td className="px-3 py-2 font-mono text-xs text-slate-500">{item.path}</td>
                <td className="px-3 py-2">{item.volume_title || '-'}</td>
                <td className="px-3 py-2">{item.unit_title || '-'}</td>
                <td className="px-3 py-2">
                  {item.skipped ? (
                    <span className="inline-flex items-center gap-1 text-amber-600">
                      <AlertCircle className="h-3.5 w-3.5" />
                      {item.reason}
                    </span>
                  ) : (
                    item.title
                  )}
                </td>
                <td className="px-3 py-2 font-mono text-xs">{item.slug || '-'}</td>
                <td className="px-3 py-2">{item.word_count || '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {result.items.length > 80 ? (
        <p className="mt-3 text-xs text-slate-500">仅预览前 80 个文件。</p>
      ) : null}
    </div>
  )
}

function ImportResultPanel({ result }: { result: MarkdownBundleImportResult }) {
  return (
    <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-4">
      <div className="flex flex-wrap items-center gap-2 text-sm">
        <Badge className="bg-emerald-600">导入 {result.imported}</Badge>
        <Badge variant="outline">新增 {result.created}</Badge>
        <Badge variant="outline">更新 {result.updated}</Badge>
        <Badge variant="outline">跳过 {result.skipped}</Badge>
      </div>
      <div className="mt-4 grid gap-2">
        {result.items.slice(0, 20).map((item) => (
          <div key={item.path} className="flex flex-wrap items-center justify-between gap-3 rounded-lg bg-white px-3 py-2 text-sm">
            <span className="min-w-0 truncate font-mono text-xs text-slate-500">{item.path}</span>
            <Badge variant={item.skipped ? 'outline' : 'default'}>
              {item.skipped ? item.reason || '跳过' : item.action}
            </Badge>
          </div>
        ))}
      </div>
    </div>
  )
}
