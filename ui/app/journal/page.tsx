'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import {
  Plus, Tag as TagIcon, Trash2, Loader2, MessageSquarePlus, Pencil, FileText,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription,
} from '@/components/ui/dialog'
import { ShellLayout } from '@/components/layout/shell-layout'
import { PermGuard } from '@/components/ui/perm-guard'
import { useTenant } from '@/lib/tenant-context'
import { showError, showSuccess } from '@/lib/toast'
import * as journalApi from '@/lib/api/journal'
import type {
  Entry, Tag,
  AddCorrectionRequest, UpdateTagsRequest, CreateTagRequest, UpdateTagRequest,
  Template, TemplateListResponse,
} from '@/types/journal'
import { charLen, formatDate, formatErrMsg, PAGE_SIZE, TAG_COLORS } from '@/components/journal/constants'
import { JournalEntryCard, FormTagPill } from '@/components/journal/journal-entry-card'
import { JournalDateNav } from '@/components/journal/journal-date-nav'
import { JournalEmptyState } from '@/components/journal/journal-empty-state'
import AIPredictionPanel from '@/components/journal/ai-prediction-panel'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuthStore } from '@/store/auth-store'

type TabType = 'entries' | 'templates' | 'ai_matrix'

export default function JournalPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { user } = useAuthStore()
  const { tenantId, selectedTenantId } = useTenant()
  const tenantReady = !!selectedTenantId

  // Tab state - read from URL query param
  const urlTab = searchParams.get('tab')
  const initialTab: TabType = urlTab === 'templates' || urlTab === 'ai_matrix' ? urlTab : 'entries'
  const [activeTab, setActiveTab] = useState<TabType>(initialTab)

  const [currentDate, setCurrentDate] = useState<Date>(() => new Date())
  const [entries, setEntries] = useState<Entry[]>([])
  const [totalEntries, setTotalEntries] = useState(0)
  const [page, setPage] = useState(1)
  const [tags, setTags] = useState<Tag[]>([])
  const [loading, setLoading] = useState(false)

  // Template state
  const [templates, setTemplates] = useState<Template[]>([])
  const [templateTotal, setTemplateTotal] = useState(0)
  const [templatePage, setTemplatePage] = useState(1)
  const [templateLoading, setTemplateLoading] = useState(false)
  const [templateNameFilter, setTemplateNameFilter] = useState('')
  const [templateTagFilter, setTemplateTagFilter] = useState('')
  const [deleteTemplateOpen, setDeleteTemplateOpen] = useState(false)
  const [deleteTemplateTarget, setDeleteTemplateTarget] = useState<Template | null>(null)
  const [deleteTemplateSaving, setDeleteTemplateSaving] = useState(false)

  const [correctionOpen, setCorrectionOpen] = useState(false)
  const [correctionEntryId, setCorrectionEntryId] = useState('')
  const [correctionContent, setCorrectionContent] = useState('')
  const [correctionSaving, setCorrectionSaving] = useState(false)

  const [editTagsOpen, setEditTagsOpen] = useState(false)
  const [editTagsEntryId, setEditTagsEntryId] = useState('')
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([])
  const [editTagsSaving, setEditTagsSaving] = useState(false)

  const [tagManagerOpen, setTagManagerOpen] = useState(false)
  const [newTagName, setNewTagName] = useState('')
  const [newTagColor, setNewTagColor] = useState('#6366f1')
  const [tagCreating, setTagCreating] = useState(false)

  const [editingTagId, setEditingTagId] = useState<string | null>(null)
  const [editingTagName, setEditingTagName] = useState('')
  const [editingTagColor, setEditingTagColor] = useState('')
  const [tagSaving, setTagSaving] = useState(false)

  const [deleteTagConfirmOpen, setDeleteTagConfirmOpen] = useState(false)
  const [deleteTagTargetId, setDeleteTagTargetId] = useState('')
  const [deleteTagTargetName, setDeleteTagTargetName] = useState('')
  const [deleteTagSaving, setDeleteTagSaving] = useState(false)

  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [deleteTargetId, setDeleteTargetId] = useState('')
  const [deleteSaving, setDeleteSaving] = useState(false)
  const templateSearchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const handleTemplateSearch = (value: string) => {
    if (templateSearchTimerRef.current) clearTimeout(templateSearchTimerRef.current)
    templateSearchTimerRef.current = setTimeout(() => {
      setTemplateNameFilter(value)
      setTemplatePage(1)
    }, 300)
  }

  useEffect(() => {
    return () => {
      if (templateSearchTimerRef.current) clearTimeout(templateSearchTimerRef.current)
    }
  }, [])

  const handleTabChange = (tab: TabType) => {
    setActiveTab(tab)
    setPage(1)
    setTemplatePage(1)
    setTemplateNameFilter('')
    setTemplateTagFilter('')
  }

  // Template fetch
  const fetchTemplates = useCallback(async () => {
    if (!tenantId) return
    setTemplateLoading(true)
    try {
      const res = await journalApi.listTemplates({
        tenant_id: tenantId,
        page: templatePage,
        page_size: PAGE_SIZE,
        name: templateNameFilter || undefined,
        tag: templateTagFilter || undefined,
      })
      setTemplates(res.data || [])
      setTemplateTotal(res.total || 0)
    } catch (e: unknown) {
      showError(formatErrMsg(e, '加载模板失败'))
    } finally {
      setTemplateLoading(false)
    }
  }, [tenantId, templatePage, templateNameFilter, templateTagFilter])

  useEffect(() => {
    if (tenantReady && activeTab === 'templates') {
      fetchTemplates()
    }
  }, [tenantReady, activeTab, fetchTemplates])

  const fetchEntries = useCallback(async () => {
    if (!tenantId) return
    setLoading(true)
    try {
      const d = formatDate(currentDate)
      const res = await journalApi.listEntries({
        tenant_id: tenantId,
        start_date: d,
        end_date: d,
        page,
        page_size: PAGE_SIZE,
      })
      setEntries(res.data || [])
      setTotalEntries(res.total || 0)
    } catch (e: unknown) {
      showError(formatErrMsg(e, '加载札记失败'))
    } finally {
      setLoading(false)
    }
  }, [tenantId, currentDate, page])

  const fetchTags = useCallback(async () => {
    if (!tenantId) return
    try {
      const res = await journalApi.listTags(tenantId)
      setTags(res.data || [])
    } catch (_e) {
      showError('标签加载失败，部分功能可能不可用')
      setTags([])
    }
  }, [tenantId])

  useEffect(() => {
    if (!tenantId) return
    fetchEntries()
  }, [tenantId, fetchEntries])

  useEffect(() => {
    if (!tenantId) return
    fetchTags()
  }, [tenantId, fetchTags])

  // ---- date navigation ----

  const handleDateChange = (date: Date) => {
    setPage(1)
    setCurrentDate(date)
  }

  const totalPages = Math.max(1, Math.ceil(totalEntries / PAGE_SIZE))

  // ---- create entry ----

  // ---- correction ----

  const openCorrection = (entryId: string) => {
    setCorrectionEntryId(entryId)
    setCorrectionContent('')
    setCorrectionOpen(true)
  }

  const handleSaveCorrection = async () => {
    if (!correctionContent.trim()) {
      showError('修正内容不能为空')
      return
    }
    if (!tenantId || !correctionEntryId) return
    setCorrectionSaving(true)
    try {
      const req: AddCorrectionRequest & { tenant_id: string } = {
        tenant_id: tenantId,
        content: correctionContent,
      }
      await journalApi.addCorrection(correctionEntryId, req)
      showSuccess('修正已追加')
      setCorrectionOpen(false)
      fetchEntries()
    } catch (e: unknown) {
      showError(formatErrMsg(e, '追加修正失败'))
    } finally {
      setCorrectionSaving(false)
    }
  }

  // ---- edit tags ----

  const openEditTags = (entry: Entry) => {
    setEditTagsEntryId(entry.id)
    setSelectedTagIds((entry.tags || []).map(t => t.id))
    setEditTagsOpen(true)
  }

  const handleSaveTags = async () => {
    if (!tenantId || !editTagsEntryId) return
    setEditTagsSaving(true)
    try {
      const req: UpdateTagsRequest & { tenant_id: string } = {
        tenant_id: tenantId,
        tag_ids: selectedTagIds,
      }
      await journalApi.updateTags(editTagsEntryId, req)
      showSuccess('标签已更新')
      setEditTagsOpen(false)
      fetchEntries()
    } catch (e: unknown) {
      showError(formatErrMsg(e, '更新标签失败'))
    } finally {
      setEditTagsSaving(false)
    }
  }

  const toggleTagSelection = (tagId: string) => {
    setSelectedTagIds(prev =>
      prev.includes(tagId) ? prev.filter(id => id !== tagId) : [...prev, tagId]
    )
  }

  // ---- delete entry ----

  const confirmDelete = (id: string) => {
    setDeleteTargetId(id)
    setDeleteConfirmOpen(true)
  }

  const handleDelete = async () => {
    if (!tenantId || !deleteTargetId) return
    setDeleteSaving(true)
    try {
      await journalApi.deleteEntry(deleteTargetId, tenantId)
      showSuccess('札记已删除')
      setDeleteConfirmOpen(false)
      setPage(1)
      fetchEntries()
    } catch (e: unknown) {
      showError(formatErrMsg(e, '删除失败'))
    } finally {
      setDeleteSaving(false)
    }
  }

  // ---- tag management ----

  const handleCreateTag = async () => {
    if (!newTagName.trim()) {
      showError('标签名称不能为空')
      return
    }
    if (!tenantId) return
    setTagCreating(true)
    try {
      const req: CreateTagRequest & { tenant_id: string } = {
        tenant_id: tenantId,
        name: newTagName.trim(),
        color: newTagColor,
      }
      await journalApi.createTag(req)
      showSuccess('标签已创建')
      setNewTagName('')
      fetchTags()
    } catch (e: unknown) {
      showError(formatErrMsg(e, '创建标签失败'))
    } finally {
      setTagCreating(false)
    }
  }

  const startEditTag = (tag: Tag) => {
    setEditingTagId(tag.id)
    setEditingTagName(tag.name)
    setEditingTagColor(tag.color)
  }

  const cancelEditTag = () => {
    setEditingTagId(null)
    setEditingTagName('')
    setEditingTagColor('')
  }

  const handleSaveEditTag = async () => {
    if (!tenantId || !editingTagId) return
    if (!editingTagName.trim()) {
      showError('标签名称不能为空')
      return
    }
    setTagSaving(true)
    try {
      const req: UpdateTagRequest & { tenant_id: string } = {
        tenant_id: tenantId,
        name: editingTagName.trim(),
        color: editingTagColor || undefined,
      }
      await journalApi.updateTag(editingTagId, req)
      showSuccess('标签已更新')
      setEditingTagId(null)
      fetchTags()
    } catch (e: unknown) {
      showError(formatErrMsg(e, '更新标签失败'))
    } finally {
      setTagSaving(false)
    }
  }

  const confirmDeleteTag = (tag: Tag) => {
    setDeleteTagTargetId(tag.id)
    setDeleteTagTargetName(tag.name)
    setDeleteTagConfirmOpen(true)
  }

  const handleDeleteTag = async () => {
    if (!tenantId || !deleteTagTargetId) return
    setDeleteTagSaving(true)
    try {
      await journalApi.deleteTag(deleteTagTargetId, tenantId)
      showSuccess('标签已删除')
      setDeleteTagConfirmOpen(false)
      fetchTags()
      fetchEntries()
    } catch (e: unknown) {
      showError(formatErrMsg(e, '删除标签失败'))
    } finally {
      setDeleteTagSaving(false)
    }
  }

  // ---- render ----

  return (
    <ShellLayout pathname="/journal">
          <Breadcrumb
            items={[
              { label: '首页', href: '/home' },
              { label: '札记', href: '/journal' },
            ]}
          />

          {/* Tab 切换 */}
          <div className="flex flex-wrap gap-2 mb-6 mt-4">
            <PermGuard button="journal.tab.entries">
              <Button
                variant={activeTab === 'entries' ? 'default' : 'outline'}
                onClick={() => handleTabChange('entries')}
              >
                当前札记
              </Button>
            </PermGuard>
            <PermGuard button="journal.tab.templates">
              <Button
                variant={activeTab === 'templates' ? 'default' : 'outline'}
                onClick={() => handleTabChange('templates')}
              >
                模板管理
              </Button>
            </PermGuard>
            <PermGuard button="journal.tab.ai_predictions">
              <Button
                variant={activeTab === 'ai_matrix' ? 'default' : 'outline'}
                onClick={() => handleTabChange('ai_matrix')}
              >
                AI矩阵推理
              </Button>
            </PermGuard>
          </div>

          {/* Tab Content */}
          {activeTab === 'entries' ? (
            <>
          {/* Toolbar: Date nav + Actions */}
          <div className="mt-4">
            <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-2">
              <JournalDateNav
                currentDate={currentDate}
                onDateChange={handleDateChange}
              />
              <div className="flex items-center gap-2 pt-1">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setTagManagerOpen(true)}
                  disabled={!tenantReady}
                >
                  <TagIcon className="h-4 w-4 mr-1" />
                  标签管理
                </Button>
                <Button
                  size="sm"
                  onClick={() => router.push('/journal/new')}
                  disabled={!tenantReady}
                  className="bg-gradient-to-r from-primary to-accent text-white shadow-lg shadow-primary/25 hover:shadow-xl hover:shadow-primary/30 hover:-translate-y-0.5 transition-all"
                >
                  <Plus className="h-4 w-4 mr-1" />
                  写札记
                </Button>
              </div>
            </div>
          </div>

          {/* Content Area */}
          {loading ? (
            <div className="flex flex-col items-center justify-center py-20 animate-fade-in">
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-primary/10 to-accent/10 flex items-center justify-center mb-3">
                <Loader2 className="h-6 w-6 animate-spin text-primary" />
              </div>
              <p className="text-sm text-slate-400">加载中...</p>
            </div>
          ) : entries.length === 0 ? (
            <JournalEmptyState onCreate={() => router.push('/journal/new')} />
          ) : (
            <div className="space-y-4">
              {entries.map((entry, idx) => (
                <div
                  key={entry.id}
                  className="animate-slide-up"
                  style={{ animationDelay: `${idx * 60}ms`, animationFillMode: 'both' }}
                >
                  <JournalEntryCard
                    entry={entry}
                    onCorrection={openCorrection}
                    onEditTags={openEditTags}
                    onDelete={confirmDelete}
                  />
                </div>
              ))}

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-center gap-4 mt-8 py-4">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page <= 1}
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    className="gap-1"
                  >
                    上一页
                  </Button>
                  <span className="text-sm text-slate-500 tabular-nums">
                    {page} / {totalPages}
                    <span className="mx-1.5 text-slate-300">|</span>
                    共 {totalEntries} 条
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page >= totalPages}
                    onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                    className="gap-1"
                  >
                    下一页
                  </Button>
                </div>
              )}
            </div>
          )}
          </>
        ) : activeTab === 'templates' ? (
          /* ---- Templates Tab ---- */
          <div className="mt-4 space-y-4">
            {/* Template Toolbar */}
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
              <div className="flex-1 max-w-md">
                <Input
                  placeholder="搜索模板名称..."
                  value={templateNameFilter}
                  onChange={e => handleTemplateSearch(e.target.value)}
                  className="h-9"
                />
              </div>
              <div className="flex items-center gap-2">
                {tags.length > 0 && (
                  <select
                    className="h-9 rounded-lg border border-slate-200 bg-white px-3 text-sm"
                    value={templateTagFilter}
                    onChange={e => { setTemplateTagFilter(e.target.value); setTemplatePage(1) }}
                  >
                    <option value="">全部标签</option>
                    {tags.map(tag => (
                      <option key={tag.id} value={tag.id}>{tag.name}</option>
                    ))}
                  </select>
                )}
                <Button
                  size="sm"
                  onClick={() => router.push('/journal/templates/new')}
                  disabled={!tenantReady}
                  className="bg-gradient-to-r from-primary to-accent text-white"
                >
                  <Plus className="h-4 w-4 mr-1" />
                  新建模板
                </Button>
              </div>
            </div>

            {/* Template List */}
            {templateLoading ? (
              <div className="flex justify-center py-12">
                <Loader2 className="h-8 w-8 animate-spin text-slate-400" />
              </div>
            ) : templates.length === 0 ? (
              <div className="text-center py-12 text-slate-400">
                <FileText className="h-12 w-12 mx-auto mb-3 opacity-50" />
                <p>暂无模板，点击上方按钮创建</p>
              </div>
            ) : (
              <div className="grid gap-3">
                {templates.map(t => (
                  <Card key={t.id} className="hover:shadow-md transition-shadow">
                    <CardContent className="py-3">
                      <div className="flex items-start justify-between">
                        <div className="flex-1 min-w-0">
                          <h3 className="font-medium text-slate-900 truncate">{t.name}</h3>
                          {t.content && (
                            <p className="text-sm text-slate-500 mt-1 line-clamp-2">{t.content}</p>
                          )}
                          {t.tags && t.tags.length > 0 && (
                            <div className="flex gap-1 mt-2 flex-wrap">
                              {t.tags.map(tagId => {
                                const tag = tags.find(tag => tag.id === tagId)
                                return tag ? (
                                  <span key={tagId} className="px-2 py-0.5 rounded text-xs" style={{ backgroundColor: tag.color + '20', color: tag.color }}>
                                    {tag.name}
                                  </span>
                                ) : null
                              })}
                            </div>
                          )}
                        </div>
                        <div className="flex items-center gap-1 ml-3">
                          {t.account_id === user?.id && (
                            <>
                              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => router.push(`/journal/templates/new?id=${t.id}`)}>
                                <Pencil className="h-4 w-4" />
                              </Button>
                              <Button variant="ghost" size="icon" className="h-8 w-8 text-red-500" onClick={() => { setDeleteTemplateTarget(t); setDeleteTemplateOpen(true) }}>
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            </>
                          )}
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}

            {/* Template Pagination */}
            {templateTotal > PAGE_SIZE && (
              <div className="flex justify-center gap-2 pt-2">
                <Button variant="outline" size="sm" disabled={templatePage <= 1} onClick={() => setTemplatePage(p => Math.max(1, p - 1))}>
                  上一页
                </Button>
                <span className="text-sm text-slate-500 py-1">
                  {templatePage} / {Math.ceil(templateTotal / PAGE_SIZE)}
                </span>
                <Button variant="outline" size="sm" disabled={templatePage >= Math.ceil(templateTotal / PAGE_SIZE)} onClick={() => setTemplatePage(p => p + 1)}>
                  下一页
                </Button>
              </div>
            )}
          </div>
        ) : activeTab === 'ai_matrix' ? (
          <div className="mt-4">
            <AIPredictionPanel tenantId={tenantId} tenantReady={tenantReady} />
          </div>
        ) : null}

      {/* ---- Correction Dialog ---- */}
      <Dialog open={correctionOpen} onOpenChange={setCorrectionOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="text-xl">追加修正</DialogTitle>
            <DialogDescription>原始内容不可修改，在此追加修正说明</DialogDescription>
          </DialogHeader>
          <div className="space-y-3 mt-2">
            <textarea
              className="w-full min-h-[140px] rounded-lg border-2 border-slate-200 bg-white px-4 py-3 text-[15px] leading-relaxed placeholder:text-slate-300 focus:outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all duration-200 resize-y"
              value={correctionContent}
              onChange={e => {
                const v = e.target.value
                if (charLen(v) <= 800) setCorrectionContent(v)
              }}
              placeholder="写下修正或补充内容..."
            />
            <div className="text-xs text-slate-300 text-right tabular-nums">
              {charLen(correctionContent)}/800
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCorrectionOpen(false)}>取消</Button>
            <Button
              onClick={handleSaveCorrection}
              disabled={correctionSaving}
              className="bg-gradient-to-r from-primary to-accent text-white shadow-lg shadow-primary/25"
            >
              {correctionSaving && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
              追加
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ---- Edit Tags Dialog ---- */}
      <Dialog open={editTagsOpen} onOpenChange={setEditTagsOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="text-xl">编辑标签</DialogTitle>
            <DialogDescription>选择要关联的标签，可随时增减</DialogDescription>
          </DialogHeader>
          <div className="flex gap-2 flex-wrap mt-3">
            {tags.map(t => (
              <FormTagPill
                key={t.id}
                tag={t}
                selected={selectedTagIds.includes(t.id)}
                onToggle={() => toggleTagSelection(t.id)}
              />
            ))}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditTagsOpen(false)}>取消</Button>
            <Button
              onClick={handleSaveTags}
              disabled={editTagsSaving}
              className="bg-gradient-to-r from-primary to-accent text-white"
            >
              {editTagsSaving && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
              保存
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ---- Tag Manager Dialog ---- */}
      <Dialog open={tagManagerOpen} onOpenChange={setTagManagerOpen}>
        <DialogContent className="max-w-sm max-h-[70vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>标签管理</DialogTitle>
            <DialogDescription>创建、编辑或删除标签</DialogDescription>
          </DialogHeader>

          {/* Create row */}
          <div className="flex items-center gap-2 pt-2">
            <div className="flex items-center gap-1">
              {TAG_COLORS.map(c => (
                <button
                  key={c}
                  type="button"
                  className={`w-4 h-4 rounded-full transition-all duration-150 ${newTagColor === c ? 'ring-2 ring-offset-1 ring-slate-700 scale-125' : 'hover:scale-110'}`}
                  style={{ backgroundColor: c }}
                  onClick={() => setNewTagColor(c)}
                />
              ))}
            </div>
            <Input
              value={newTagName}
              onChange={e => setNewTagName(e.target.value)}
              placeholder="新标签"
              maxLength={30}
              className="flex-1 h-8 text-sm"
              onKeyDown={e => { if (e.key === 'Enter' && newTagName.trim()) handleCreateTag() }}
            />
            <Button size="sm" onClick={handleCreateTag} disabled={tagCreating || !newTagName.trim()}
              className="h-8 px-2.5 bg-gradient-to-r from-primary to-accent text-white shrink-0">
              {tagCreating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
            </Button>
          </div>

          {/* Tag list */}
          <div className="mt-3">
            {tags.length === 0 ? (
              <p className="text-xs text-slate-400 text-center py-4">暂无标签，创建第一个吧</p>
            ) : (
              <div className="space-y-0.5">
                {tags.map(t => (
                  <div key={t.id}>
                    {editingTagId === t.id ? (
                      <div className="flex items-center gap-2 py-1.5 px-2 rounded-lg bg-primary/5 border border-primary/20">
                        <div className="flex items-center gap-0.5">
                          {TAG_COLORS.map(c => (
                            <button
                              key={c}
                              type="button"
                              className={`w-3.5 h-3.5 rounded-full transition-all duration-150 ${editingTagColor === c ? 'ring-1.5 ring-offset-0.5 ring-slate-700 scale-125' : 'hover:scale-110 opacity-60 hover:opacity-100'}`}
                              style={{ backgroundColor: c }}
                              onClick={() => setEditingTagColor(c)}
                            />
                          ))}
                        </div>
                        <Input
                          value={editingTagName}
                          onChange={e => setEditingTagName(e.target.value)}
                          className="flex-1 h-7 text-sm"
                          maxLength={30}
                          onKeyDown={e => { if (e.key === 'Enter') handleSaveEditTag() }}
                        />
                        <Button size="sm" variant="ghost" className="h-7 w-7 p-0 text-primary" onClick={handleSaveEditTag} disabled={tagSaving}>
                          {tagSaving ? <Loader2 className="h-3 w-3 animate-spin" /> : '✓'}
                        </Button>
                        <Button size="sm" variant="ghost" className="h-7 w-7 p-0 text-slate-400" onClick={cancelEditTag}>
                          ✕
                        </Button>
                      </div>
                    ) : (
                      <div className="flex items-center justify-between py-1.5 px-2 rounded-lg hover:bg-slate-50 group transition-colors">
                        <div className="flex items-center gap-2">
                          <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: t.color }} />
                          <span className="text-sm text-slate-700">{t.name}</span>
                        </div>
                        <div className="flex items-center gap-0 opacity-0 group-hover:opacity-100 transition-opacity">
                          <Button variant="ghost" size="sm" className="h-6 w-6 p-0 text-slate-400 hover:text-primary" onClick={() => startEditTag(t)}>
                            <Pencil className="h-3 w-3" />
                          </Button>
                          <Button variant="ghost" size="sm" className="h-6 w-6 p-0 text-slate-400 hover:text-red-500" onClick={() => confirmDeleteTag(t)}>
                            <Trash2 className="h-3 w-3" />
                          </Button>
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* ---- Delete Tag Confirm ---- */}
      <Dialog open={deleteTagConfirmOpen} onOpenChange={setDeleteTagConfirmOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>确认删除标签</DialogTitle>
            <DialogDescription>
              确定要删除标签「{deleteTagTargetName}」吗？删除后将从所有相关札记中移除该标签。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTagConfirmOpen(false)}>取消</Button>
            <Button variant="destructive" onClick={handleDeleteTag} disabled={deleteTagSaving}>
              {deleteTagSaving && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
              确认删除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ---- Delete Entry Confirm ---- */}
      <Dialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>
              删除后不可恢复，关联的修正条目也会一并删除
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteConfirmOpen(false)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteSaving}>
              {deleteSaving && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
              确认删除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ---- Delete Template Confirm ---- */}
      <Dialog open={deleteTemplateOpen} onOpenChange={setDeleteTemplateOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>确认删除模板</DialogTitle>
            <DialogDescription>
              确定要删除 &ldquo;{deleteTemplateTarget?.name}&rdquo; 吗？此操作不可恢复。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTemplateOpen(false)}>取消</Button>
            <Button variant="destructive" onClick={async () => {
              if (!deleteTemplateTarget) return
              setDeleteTemplateSaving(true)
              try {
                await journalApi.deleteTemplate(deleteTemplateTarget.id, tenantId)
                showSuccess('模板已删除')
                setDeleteTemplateOpen(false)
                fetchTemplates()
              } catch (e) { showError(formatErrMsg(e, '删除失败')) }
              finally { setDeleteTemplateSaving(false) }
            }} disabled={deleteTemplateSaving}>
              {deleteTemplateSaving && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
              确认删除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </ShellLayout>
  )
}