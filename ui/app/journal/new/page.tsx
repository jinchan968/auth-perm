'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import { useRouter } from 'next/navigation'
import {
  Loader2, ArrowLeft, Tag as TagIcon, FileText, ChevronDown,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { ShellLayout } from '@/components/layout/shell-layout'
import { useTenant } from '@/lib/tenant-context'
import { showError, showSuccess } from '@/lib/toast'
import * as journalApi from '@/lib/api/journal'
import type {
  Tag, Period, Weather, CreateEntryRequest, Template, TemplateListResponse,
} from '@/types/journal'
import { PERIODS, WEATHERS, WEEKDAYS, charLen, inferPeriod, formatDate, formatErrMsg } from '@/components/journal/constants'
import { FormTagPill } from '@/components/journal/journal-entry-card'

export default function JournalNewPage() {
  const router = useRouter()
  const { tenantId, selectedTenantId } = useTenant()
  const tenantReady = !!selectedTenantId

  const [tags, setTags] = useState<Tag[]>([])

  const [formTitle, setFormTitle] = useState('')
  const [formContent, setFormContent] = useState('')
  const [formPeriod, setFormPeriod] = useState<Period>(inferPeriod())
  const [formWeather, setFormWeather] = useState<Weather | undefined>()
  const [formLocation, setFormLocation] = useState('')
  const [formTagIds, setFormTagIds] = useState<string[]>([])
  const [formDate, setFormDate] = useState(formatDate(new Date()))
  const [saving, setSaving] = useState(false)

  const fetchTags = useCallback(async () => {
    if (!tenantId) return
    try {
      const res = await journalApi.listTags(tenantId)
      setTags(res.data || [])
    } catch (_e) {
      showError('标签加载失败')
    }
  }, [tenantId])

  useEffect(() => {
    if (!tenantId) return
    fetchTags()
  }, [tenantId, fetchTags])

  // Template state
  const [templates, setTemplates] = useState<Template[]>([])
  const [templateLoading, setTemplateLoading] = useState(false)
  const [templateDropdownOpen, setTemplateDropdownOpen] = useState(false)
  const [templateSearch, setTemplateSearch] = useState('')
  const dropdownRef = useRef<HTMLDivElement>(null)
  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const searchReqIdRef = useRef(0)

  const fetchTemplates = useCallback(async (search?: string) => {
    if (!tenantId) return
    const reqId = ++searchReqIdRef.current
    setTemplateLoading(true)
    try {
      const res = await journalApi.listTemplates({
        tenant_id: tenantId,
        page_size: 50,
        name: search || undefined,
      })
      if (reqId === searchReqIdRef.current) {
        setTemplates(res.data || [])
      }
    } catch (_e) {
      // Silently fail - template is optional
    } finally {
      if (reqId === searchReqIdRef.current) {
        setTemplateLoading(false)
      }
    }
  }, [tenantId])

  useEffect(() => {
    fetchTemplates()
  }, [fetchTemplates])

  useEffect(() => {
    return () => {
      if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
    }
  }, [])

  // Debounced search
  const handleTemplateSearch = (value: string) => {
    setTemplateSearch(value)
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
    searchTimerRef.current = setTimeout(() => {
      fetchTemplates(value)
    }, 300)
  }

  // Click outside to close dropdown
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setTemplateDropdownOpen(false)
      }
    }
    if (templateDropdownOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [templateDropdownOpen])

  const applyTemplate = (template: Template) => {
    if (template.content) {
      setFormContent(prev => (prev ? prev + '\n\n' + template.content : template.content || ''))
    }
    if (template.name) {
      setFormTitle(prev => prev || template.name || '')
    }
    if (template.tags && template.tags.length > 0) {
      // Find matching tags
      const matchingTags = tags.filter(t => template.tags?.includes(t.id)).map(t => t.id)
      if (matchingTags.length > 0) {
        setFormTagIds(prev => {
          const combined = [...prev, ...matchingTags]
          return Array.from(new Set(combined))
        })
      }
    }
    setTemplateDropdownOpen(false)
    setTemplateSearch('')
    showSuccess(`已应用模板: ${template.name}`)
  }

  const weatherIcon = WEATHERS.find(w => w.value === formWeather)?.icon

  const handleSubmit = async () => {
    if (!formContent.trim()) {
      showError('札记内容不能为空')
      return
    }
    if (!tenantId) return
    setSaving(true)
    try {
      const req: CreateEntryRequest & { tenant_id: string } = {
        tenant_id: tenantId,
        content: formContent,
        period: formPeriod,
        entry_date: formDate,
        title: formTitle.trim() || undefined,
        weather: formWeather,
        location: formLocation.trim() || undefined,
        tag_ids: formTagIds.length > 0 ? formTagIds : undefined,
      }
      await journalApi.createEntry(req)
      showSuccess('札记已创建')
      router.push('/journal')
    } catch (e: unknown) {
      showError(formatErrMsg(e, '创建失败'))
    } finally {
      setSaving(false)
    }
  }

  const toggleFormTag = (tagId: string) => {
    setFormTagIds(prev =>
      prev.includes(tagId) ? prev.filter(id => id !== tagId) : [...prev, tagId]
    )
  }

  const today = formatDate(new Date())
  const dateDisplay = formDate.replace(/-/g, '/')
  const [year, month, day] = formDate.split('-')
  const WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六']
  const dateObj = new Date(formDate + 'T00:00:00')
  const weekday = WEEKDAYS[dateObj.getDay()]

  return (
    <ShellLayout pathname="/journal">
          <Breadcrumb
            items={[
              { label: '首页', href: '/home' },
              { label: '札记', href: '/journal' },
              { label: '新建', href: '/journal/new' },
            ]}
          />

          {/* Page Header */}
          <div className="mt-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mb-8">
            <div className="flex items-center gap-4">
              <button
                type="button"
                className="h-9 w-9 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 hover:border-slate-300 flex items-center justify-center transition-colors"
                onClick={() => router.push('/journal')}
              >
                <ArrowLeft className="h-4 w-4 text-slate-500" />
              </button>
              <div>
                <h1 className="text-2xl font-bold text-slate-800">写札记</h1>
                <p className="text-sm text-slate-400 mt-0.5">
                  记录此刻的想法与感悟
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Button
                variant="outline"
                onClick={() => router.push('/journal')}
              >
                取消
              </Button>
              <Button
                onClick={handleSubmit}
                disabled={saving || !tenantReady}
                className="bg-gradient-to-r from-primary to-accent text-white shadow-lg shadow-primary/25 hover:shadow-xl hover:shadow-primary/30 hover:-translate-y-0.5 transition-all"
              >
                {saving && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
                保存
              </Button>
            </div>
          </div>

          {/* Form Card */}
          <div className="max-w-3xl mx-auto">
            <div className="bg-white/90 backdrop-blur-sm rounded-xl border border-slate-200/60 shadow-sm p-6 lg:p-8 space-y-6">

              {/* Date & Period row */}
              <div className="grid grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-slate-600 mb-2">
                    日期
                  </label>
                  <Input
                    type="date"
                    value={formDate}
                    onChange={e => setFormDate(e.target.value)}
                  />
                  <p className="text-xs text-slate-400 mt-1.5">
                    {year}年{month?.replace(/^0/, '')}月{day?.replace(/^0/, '')}日 星期{weekday}
                    {formDate === today && <span className="ml-1 text-primary font-medium">今天</span>}
                  </p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-600 mb-2">
                    时段 <span className="text-red-400">*</span>
                  </label>
                  <div className="flex flex-wrap gap-2">
                    {PERIODS.map(p => (
                      <button
                        key={p.value}
                        type="button"
                        className={`
                          px-3.5 py-2 rounded-lg text-sm font-medium border transition-all duration-200
                          ${formPeriod === p.value
                            ? p.bgClass + ' shadow-sm'
                            : 'bg-white text-slate-500 border-slate-200 hover:border-slate-300'
                          }
                        `}
                        onClick={() => setFormPeriod(p.value)}
                      >
                        {p.emoji} {p.label}
                      </button>
                    ))}
                  </div>
                </div>
              </div>

              {/* Title */}
              <div>
                <label className="block text-sm font-medium text-slate-600 mb-2">
                  标题 <span className="text-slate-300 font-normal">(可选)</span>
                </label>
                <Input
                  value={formTitle}
                  onChange={e => setFormTitle(e.target.value)}
                  placeholder="给这篇札记起个标题"
                  maxLength={100}
                />
              </div>

              {/* Weather & Location row */}
              <div className="grid grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-slate-600 mb-2">
                    天气 <span className="text-slate-300 font-normal">(可选)</span>
                    {formWeather && weatherIcon && (
                      <span className="inline-flex items-center ml-2 text-primary">
                        {weatherIcon}
                      </span>
                    )}
                  </label>
                  <div className="flex flex-wrap gap-2">
                    {WEATHERS.map(w => (
                      <button
                        key={w.value}
                        type="button"
                        className={`
                          px-3 py-1.5 rounded-lg text-sm font-medium border transition-all duration-200
                          flex items-center gap-1.5
                          ${formWeather === w.value
                            ? 'bg-primary/10 text-primary border-primary/30 shadow-sm'
                            : 'bg-white text-slate-500 border-slate-200 hover:border-slate-300'
                          }
                        `}
                        onClick={() => setFormWeather(formWeather === w.value ? undefined : w.value)}
                      >
                        {w.icon}
                        {w.label}
                      </button>
                    ))}
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-600 mb-2">
                    位置 <span className="text-slate-300 font-normal">(可选)</span>
                  </label>
                  <Input
                    value={formLocation}
                    onChange={e => setFormLocation(e.target.value)}
                    placeholder="输入位置描述"
                    maxLength={200}
                  />
                </div>
              </div>

              {/* Template Selector */}
              <div className="relative" ref={dropdownRef}>
                <label className="block text-sm font-medium text-slate-600 mb-2">
                  应用模板 <span className="text-slate-300 font-normal">(可选)</span>
                </label>
                <div className="relative">
                  <button
                    type="button"
                    className="w-full flex items-center justify-between px-3 py-2 rounded-lg border border-slate-200 bg-white hover:border-slate-300 transition-colors text-left"
                    onClick={() => setTemplateDropdownOpen(!templateDropdownOpen)}
                  >
                    <span className="text-sm text-slate-500">
                      点击选择模板...
                    </span>
                    <ChevronDown className="h-4 w-4 text-slate-400" />
                  </button>
                  {templateDropdownOpen && (
                    <div className="absolute z-20 mt-1 w-full bg-white rounded-lg border border-slate-200 shadow-lg max-h-[300px] overflow-hidden">
                      <div className="p-2 border-b border-slate-100">
                        <Input
                          placeholder="搜索模板..."
                          value={templateSearch}
                          onChange={e => handleTemplateSearch(e.target.value)}
                          className="h-8 text-sm"
                          onClick={e => e.stopPropagation()}
                        />
                      </div>
                      <div className="overflow-y-auto max-h-[240px]">
                        {templateLoading ? (
                          <div className="flex justify-center py-4">
                            <Loader2 className="h-5 w-5 animate-spin text-slate-400" />
                          </div>
                        ) : templates.length === 0 ? (
                          <div className="py-4 text-center text-sm text-slate-400">
                            暂无可用模板
                          </div>
                        ) : (
                          templates.map(t => (
                            <button
                              key={t.id}
                              type="button"
                              className="w-full px-3 py-2 text-left hover:bg-slate-50 flex items-center gap-2"
                              onClick={() => applyTemplate(t)}
                            >
                              <FileText className="h-4 w-4 text-slate-400" />
                              <div className="flex-1 min-w-0">
                                <div className="text-sm font-medium text-slate-700 truncate">{t.name}</div>
                                {t.content && (
                                  <div className="text-xs text-slate-400 truncate">{t.content.slice(0, 50)}</div>
                                )}
                              </div>
                            </button>
                          ))
                        )}
                      </div>
                    </div>
                  )}
                </div>
              </div>

              {/* Content - main textarea */}
              <div>
                <label className="block text-sm font-medium text-slate-600 mb-2">
                  正文 <span className="text-red-400">*</span>
                  <span className="text-xs text-slate-300 ml-2 tabular-nums">
                    {charLen(formContent)}/800
                  </span>
                </label>
                <textarea
                  className="w-full min-h-[280px] rounded-lg border-2 border-slate-200 bg-white px-4 py-3 text-[15px] leading-relaxed placeholder:text-slate-300 focus:outline-none focus:border-primary focus:ring-4 focus:ring-primary/10 transition-all duration-200 resize-y"
                  value={formContent}
                  onChange={e => {
                    const v = e.target.value
                    if (charLen(v) <= 800) setFormContent(v)
                  }}
                  placeholder="写下今天的所思所想..."
                  autoFocus
                />
              </div>

              {/* Tags */}
              {tags.length > 0 && (
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
                        selected={formTagIds.includes(t.id)}
                        onToggle={() => toggleFormTag(t.id)}
                      />
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
    </ShellLayout>
  )
}
