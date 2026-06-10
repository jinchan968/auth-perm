'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { BookOpen, Eye, FileArchive, RefreshCw, Search, Trash2 } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { showError, showSuccess } from '@/lib/toast'
import { useTenant } from '@/lib/tenant-context'
import { deleteNovel, listMyNovels } from '@/lib/api/novel'
import type { NovelListItem, NovelStatus } from '@/types/novel'

const statusLabels: Record<NovelStatus, string> = {
  draft: '草稿',
  serial: '连载中',
  completed: '已完结',
  archived: '已归档',
}

const statusBadgeVariants: Record<NovelStatus, 'outline' | 'info' | 'success' | 'secondary'> = {
  draft: 'outline',
  serial: 'info',
  completed: 'success',
  archived: 'secondary',
}

function formatDate(value: string) {
  if (!value) return '-'
  return new Date(value).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

export function NovelListWorkspace() {
  const router = useRouter()
  const { tenantId } = useTenant()
  const [novels, setNovels] = useState<NovelListItem[]>([])
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')

  const refreshNovels = useCallback(async () => {
    setLoading(true)
    try {
      const result = await listMyNovels(tenantId)
      setNovels(result.data)
    } catch (err) {
      showError(err instanceof Error ? err.message : '加载小说列表失败')
    } finally {
      setLoading(false)
    }
  }, [tenantId])

  useEffect(() => {
    if (!tenantId) return
    refreshNovels()
  }, [tenantId, refreshNovels])

  const handleDelete = async (id: string, title: string) => {
    if (!confirm(`确定删除「${title}」？将一并删除所有章节、版本、元信息，不可恢复。`)) return
    try {
      await deleteNovel(tenantId, id)
      showSuccess('已删除')
      refreshNovels()
    } catch (err) {
      showError(err instanceof Error ? err.message : '删除失败')
    }
  }

  const filteredNovels = useMemo(() => {
    const query = keyword.trim().toLowerCase()
    if (!query) return novels
    return novels.filter((novel) => {
      const source = [
        novel.title,
        novel.subtitle,
        novel.description,
        novel.status,
        ...(novel.tags || []),
      ].join(' ').toLowerCase()
      return source.includes(query)
    })
  }, [keyword, novels])

  const stats = useMemo(() => ({
    total: novels.length,
    serial: novels.filter((novel) => novel.status === 'serial').length,
    draft: novels.filter((novel) => novel.status === 'draft').length,
    completed: novels.filter((novel) => novel.status === 'completed').length,
  }), [novels])

  function goImport(novelId?: string) {
    const query = novelId ? `?novel_id=${encodeURIComponent(novelId)}` : ''
    router.push(`/novels/import${query}`)
  }

  function goDetail(novelId: string) {
    router.push(`/novels/${encodeURIComponent(novelId)}`)
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-foreground">小说列表</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            查看当前租户下的小说，并进入导入流程维护卷、单元和章节内容。
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button variant="outline" onClick={refreshNovels} disabled={loading}>
            <RefreshCw className="h-4 w-4" />
            刷新
          </Button>
          <Button onClick={() => goImport()}>
            <FileArchive className="h-4 w-4" />
            导入小说
          </Button>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="全部小说" value={stats.total} />
        <StatCard label="连载中" value={stats.serial} />
        <StatCard label="草稿" value={stats.draft} />
        <StatCard label="已完结" value={stats.completed} />
      </div>

      <Card>
        <CardHeader className="gap-4 sm:flex-row sm:items-center sm:justify-between">
          <CardTitle className="flex items-center gap-2">
            <BookOpen className="h-5 w-5 text-primary" />
            作品库
          </CardTitle>
          <div className="relative w-full sm:w-80">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={keyword}
              onChange={(event) => setKeyword(event.target.value)}
              placeholder="搜索标题、标签、状态"
              className="pl-9"
            />
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="rounded-lg border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
              正在加载小说列表...
            </div>
          ) : filteredNovels.length === 0 ? (
            <div className="rounded-lg border border-dashed border-border p-8 text-center">
              <div className="text-sm font-medium text-foreground">
                {novels.length === 0 ? '当前还没有小说' : '没有匹配的小说'}
              </div>
              <p className="mt-2 text-sm text-muted-foreground">
                {novels.length === 0 ? '可以先导入 zip 文件树，或在导入页创建小说。' : '换一个关键词再试试。'}
              </p>
              {novels.length === 0 ? (
                <Button className="mt-4" onClick={() => goImport()}>
                  <FileArchive className="h-4 w-4" />
                  去导入
                </Button>
              ) : null}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[840px] text-left text-sm">
                <thead className="border-b border-border text-xs uppercase text-muted-foreground">
                  <tr>
                    <th className="px-3 py-3 font-medium">小说</th>
                    <th className="px-3 py-3 font-medium">状态</th>
                    <th className="px-3 py-3 font-medium">标签</th>
                    <th className="px-3 py-3 font-medium">更新时间</th>
                    <th className="px-3 py-3 text-right font-medium">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredNovels.map((novel) => (
                    <tr key={novel.id} className="border-b border-border/70 last:border-0">
                      <td className="max-w-[28rem] px-3 py-4">
                        <div className="truncate font-medium text-foreground">{novel.title}</div>
                        <div className="mt-1 truncate text-xs text-muted-foreground">
                          {novel.subtitle || novel.description || '暂无副标题'}
                        </div>
                      </td>
                      <td className="px-3 py-4">
                        <Badge variant={statusBadgeVariants[novel.status]}>
                          {statusLabels[novel.status]}
                        </Badge>
                      </td>
                      <td className="px-3 py-4">
                        {novel.tags?.length ? (
                          <div className="flex max-w-[18rem] flex-wrap gap-1">
                            {novel.tags.slice(0, 4).map((tag) => (
                              <Badge key={tag} variant="outline">{tag}</Badge>
                            ))}
                            {novel.tags.length > 4 ? <Badge variant="secondary">+{novel.tags.length - 4}</Badge> : null}
                          </div>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </td>
                      <td className="px-3 py-4 text-muted-foreground">{formatDate(novel.updated_at)}</td>
                      <td className="px-3 py-4">
                        <div className="flex justify-end gap-2">
                          <Button variant="outline" size="sm" onClick={() => goDetail(novel.id)}>
                            <Eye className="h-4 w-4" />
                            明细
                          </Button>
                          <Button variant="outline" size="sm" onClick={() => goImport(novel.id)}>
                            <FileArchive className="h-4 w-4" />
                            导入
                          </Button>
                          <Button variant="outline" size="sm" onClick={() => handleDelete(novel.id, novel.title)}>
                            <Trash2 className="h-4 w-4 text-red-500" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
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
