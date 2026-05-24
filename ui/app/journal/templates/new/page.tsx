'use client'

import { useState, useEffect, useCallback } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { ArrowLeft, Tag as TagIcon, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { ShellLayout } from '@/components/layout/shell-layout'
import { useTenant } from '@/lib/tenant-context'
import { showError, showSuccess } from '@/lib/toast'
import * as journalApi from '@/lib/api/journal'
import type { Template, Tag } from '@/types/journal'
import { FormTagPill } from '@/components/journal/journal-entry-card'

export default function TemplateFormPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const editId = searchParams.get('id')
  const isEdit = !!editId

  const { tenantId, selectedTenantId } = useTenant()
  const tenantReady = !!selectedTenantId

  const [tags, setTags] = useState<Tag[]>([])
  const [name, setName] = useState('')
  const [content, setContent] = useState('')
  const [tagIds, setTagIds] = useState<string[]>([])
  const [saving, setSaving] = useState(false)
  const [loading, setLoading] = useState(isEdit)
  const [loadFailed, setLoadFailed] = useState(false)

  const fetchTags = useCallback(async () => {
    if (!tenantId) return
    try {
      const res = await journalApi.listTags(tenantId)
      setTags(res.data || [])
    } catch {
      showError('标签加载失败')
    }
  }, [tenantId])

  useEffect(() => {
    if (!tenantId) return
    fetchTags()
  }, [tenantId, fetchTags])

  useEffect(() => {
    if (!editId || !tenantId) return
    const loadTemplate = async () => {
      setLoading(true)
      try {
        const t = await journalApi.getTemplate(editId, tenantId)
        setName(t.name)
        setContent(t.content || '')
        setTagIds(t.tags || [])
      } catch {
        showError('模板加载失败')
        setLoadFailed(true)
      } finally {
        setLoading(false)
      }
    }
    loadTemplate()
  }, [editId, tenantId])

  const handleSave = async () => {
    if (!name.trim()) {
      showError('请输入模板名称')
      return
    }
    if (!tenantId) return
    setSaving(true)
    try {
      if (isEdit) {
        await journalApi.updateTemplate(editId, {
          tenant_id: tenantId,
          name: name.trim() || undefined,
          content: content,
          tags: tagIds,
        })
      } else {
        await journalApi.createTemplate({
          tenant_id: tenantId,
          name: name.trim(),
          content: content || undefined,
          tags: tagIds.length > 0 ? tagIds : undefined,
        })
      }
      showSuccess(isEdit ? '模板已更新' : '模板已创建')
      router.push('/journal?tab=templates')
    } catch (e) {
      showError(e instanceof Error ? e.message : '操作失败')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <ShellLayout pathname="/journal/templates">
        <div className="flex justify-center py-20">
          <Loader2 className="h-8 w-8 animate-spin text-slate-400" />
        </div>
      </ShellLayout>
    )
  }

  if (loadFailed) {
    return (
      <ShellLayout pathname="/journal/templates">
        <div className="text-center py-20">
          <p className="text-slate-500 mb-4">模板加载失败</p>
          <Button variant="outline" onClick={() => router.push('/journal?tab=templates')}>返回</Button>
        </div>
      </ShellLayout>
    )
  }

  return (
    <ShellLayout pathname="/journal/templates">
          <Breadcrumb
            items={[
              { label: '首页', href: '/home' },
              { label: '札记', href: '/journal' },
              { label: isEdit ? '编辑模板' : '新建模板' },
            ]}
          />

          {/* Page Header */}
          <div className="mt-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mb-8">
            <div className="flex items-center gap-4">
              <button
                type="button"
                className="h-9 w-9 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 hover:border-slate-300 flex items-center justify-center transition-colors"
                onClick={() => router.push('/journal?tab=templates')}
              >
                <ArrowLeft className="h-4 w-4 text-slate-500" />
              </button>
              <div>
                <h1 className="text-2xl font-bold text-slate-800">
                  {isEdit ? '编辑模板' : '新建模板'}
                </h1>
                <p className="text-sm text-slate-400 mt-0.5">
                  {isEdit ? '修改模板名称、内容和标签' : '创建一个可复用的札记模板'}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Button variant="outline" onClick={() => router.push('/journal?tab=templates')}>
                取消
              </Button>
              <Button
                onClick={handleSave}
                disabled={saving || !tenantReady}
                className="bg-gradient-to-r from-primary to-accent text-white shadow-lg shadow-primary/25 hover:shadow-xl hover:shadow-primary/30 hover:-translate-y-0.5 transition-all"
              >
                {saving && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
                保存
              </Button>
            </div>
          </div>

          {/* Form Card */}
          <div className="max-w-2xl mx-auto">
            <div className="bg-white/90 backdrop-blur-sm rounded-xl border border-slate-200/60 shadow-sm p-6 lg:p-8 space-y-6">
              {/* Name */}
              <div>
                <label className="block text-sm font-medium text-slate-600 mb-2">
                  模板名称 <span className="text-red-400">*</span>
                </label>
                <Input
                  value={name}
                  onChange={e => setName(e.target.value)}
                  placeholder="例如：每日复盘"
                  maxLength={255}
                />
              </div>

              {/* Content */}
              <div>
                <label className="block text-sm font-medium text-slate-600 mb-2">
                  模板内容 <span className="text-slate-300 font-normal">(可选)</span>
                </label>
                <textarea
                  className="w-full min-h-[280px] rounded-lg border-2 border-slate-200 bg-white px-4 py-3 text-[15px] leading-relaxed placeholder:text-slate-300 focus:outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all duration-200 resize-y"
                  value={content}
                  onChange={e => setContent(e.target.value)}
                  placeholder="输入模板内容，新建札记时将自动填入..."
                />
              </div>

              {/* Tags */}
              <div>
                <label className="block text-sm font-medium text-slate-600 mb-2">
                  <TagIcon className="h-3.5 w-3.5 inline mr-1 text-slate-400" />
                  标签 <span className="text-slate-300 font-normal">(可选)</span>
                </label>
                <div className="flex gap-2 flex-wrap">
                  {tags.map(t => (
                    <FormTagPill
                      key={t.id}
                      tag={t}
                      selected={tagIds.includes(t.id)}
                      onToggle={() => {
                        setTagIds(prev =>
                          prev.includes(t.id)
                            ? prev.filter(id => id !== t.id)
                            : [...prev, t.id]
                        )
                      }}
                    />
                  ))}
                  {tags.length === 0 && (
                    <p className="text-xs text-slate-400 py-1">暂无可用标签</p>
                  )}
                </div>
              </div>
            </div>
          </div>
    </ShellLayout>
  )
}
