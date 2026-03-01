'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import {
  Plus, Search, Trash2, CheckSquare, Square, AlertCircle,
  ChevronDown, Calendar, X, Pencil
} from 'lucide-react'
import {
  listTodos, createTodo, updateTodo, updateTodoStatus, updateTodoPriority,
  deleteTodo, listCategories, createCategory, deleteCategory,
} from '@/lib/api/todo'
import { TodoItem, TodoCategory, TodoStatus, TodoPriority } from '@/types/todo'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { Label } from '@/components/ui/label'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from '@/components/ui/dialog'
import { useTenant } from '@/lib/tenant-context'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'

// ─── constants ────────────────────────────────────────────────────────────────

const PRIORITY_LABEL: Record<TodoPriority, string> = {
  low: '低', medium: '中', high: '高', urgent: '紧急',
}
const PRIORITY_COLOR: Record<TodoPriority, string> = {
  low: 'bg-slate-100 text-slate-600',
  medium: 'bg-blue-100 text-blue-700',
  high: 'bg-orange-100 text-orange-700',
  urgent: 'bg-red-100 text-red-700',
}
const STATUS_LABEL: Record<TodoStatus, string> = {
  pending: '待处理', in_progress: '进行中', completed: '已完成', cancelled: '已取消',
}
const CATEGORY_COLORS = [
  '#6366f1', '#ec4899', '#f59e0b', '#10b981', '#3b82f6',
  '#8b5cf6', '#ef4444', '#14b8a6', '#f97316', '#84cc16',
]

// ─── helper ───────────────────────────────────────────────────────────────────

function fmtDate(iso?: string) {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

function isToday(iso?: string) {
  if (!iso) return false
  const d = new Date(iso)
  const now = new Date()
  return d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate()
}

// ─── main component ───────────────────────────────────────────────────────────

export default function TodosPage() {
  const { user } = useAuthStore()
  const { selectedTenantId } = useTenant()
  const tenantReady = !!selectedTenantId

  // list state
  const [todos, setTodos] = useState<TodoItem[]>([])
  const [categories, setCategories] = useState<TodoCategory[]>([])
  const [loading, setLoading] = useState(true)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const PAGE_SIZE = 20

  // filter state
  const [keyword, setKeyword] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('')
  const [filterPriority, setFilterPriority] = useState<string>('')
  const [activeCategoryId, setActiveCategoryId] = useState<string>('') // '' = all, 'none' = uncategorised

  // error
  const [error, setError] = useState('')

  const ensureTenant = () => {
    if (!selectedTenantId) {
      setError('请选择租户')
      return false
    }
    return true
  }

  // ── Create/Edit Todo dialog ──
  const [todoDialogOpen, setTodoDialogOpen] = useState(false)
  const [editingTodo, setEditingTodo] = useState<TodoItem | null>(null)
  const [todoSaving, setTodoSaving] = useState(false)
  const [todoForm, setTodoForm] = useState({
    title: '',
    description: '',
    priority: 'medium' as TodoPriority,
    category_id: '',
    deadline: '',
  })

  // ── Create Category dialog ──
  const [catDialogOpen, setCatDialogOpen] = useState(false)
  const [catSaving, setCatSaving] = useState(false)
  const [catForm, setCatForm] = useState({ name: '', color: CATEGORY_COLORS[0] })

  // ── Delete confirm ──
  const [deletingId, setDeletingId] = useState<string | null>(null)

  // ─── fetch ──────────────────────────────────────────────────────────────────

  const fetchCategories = useCallback(async () => {
    if (!selectedTenantId) return
    try {
      const cats = await listCategories(selectedTenantId)
      setCategories(cats)
    } catch { /* silent */ }
  }, [selectedTenantId])

  const fetchTodos = useCallback(async (p = page) => {
    if (!selectedTenantId) return
    setLoading(true)
    setError('')
    try {
      const res = await listTodos({
        tenant_id: selectedTenantId,
        keyword: keyword || undefined,
        status: (filterStatus || undefined) as TodoStatus | undefined,
        priority: (filterPriority || undefined) as TodoPriority | undefined,
        category_id: activeCategoryId || undefined,
        page: p,
        page_size: PAGE_SIZE,
      })
      setTodos(res.data || [])
      setTotal(res.total)
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }, [selectedTenantId, keyword, filterStatus, filterPriority, activeCategoryId, page])

  useEffect(() => {
    fetchCategories()
  }, [fetchCategories])

  useEffect(() => {
    setPage(1)
    fetchTodos(1)
  }, [selectedTenantId, filterStatus, filterPriority, activeCategoryId])

  // ─── handlers ───────────────────────────────────────────────────────────────

  const handleSearch = () => { setPage(1); fetchTodos(1) }

  const openCreate = () => {
    setEditingTodo(null)
    setTodoForm({ title: '', description: '', priority: 'medium', category_id: '', deadline: '' })
    setTodoDialogOpen(true)
  }

  const openEdit = (t: TodoItem) => {
    setEditingTodo(t)
    setTodoForm({
      title: t.title,
      description: t.description || '',
      priority: t.priority,
      category_id: t.category_id || '',
      deadline: t.deadline ? new Date(t.deadline).toISOString().slice(0, 16) : '',
    })
    setTodoDialogOpen(true)
  }

  const handleSaveTodo = async () => {
    if (!ensureTenant()) return
    if (!todoForm.title.trim()) { setError('标题不能为空'); return }
    setTodoSaving(true)
    setError('')
    try {
      const payload = {
        tenant_id: selectedTenantId,
        title: todoForm.title.trim(),
        description: todoForm.description || undefined,
        priority: todoForm.priority,
        category_id: todoForm.category_id || undefined,
        deadline: todoForm.deadline ? new Date(todoForm.deadline).toISOString() : undefined,
      }
      if (editingTodo) {
        const updated = await updateTodo(editingTodo.id, {
          ...payload,
          clear_category: !todoForm.category_id,
          clear_deadline: !todoForm.deadline,
        })
        setTodos(prev => prev.map(t => t.id === updated.id ? updated : t))
      } else {
        const created = await createTodo(payload)
        setTodos(prev => [created, ...prev])
        setTotal(t => t + 1)
      }
      setTodoDialogOpen(false)
    } catch (e) {
      setError(e instanceof Error ? e.message : '保存失败')
    } finally {
      setTodoSaving(false)
    }
  }

  const handleToggleComplete = async (t: TodoItem) => {
    if (!ensureTenant()) return
    const newStatus: TodoStatus = t.status === 'completed' ? 'pending' : 'completed'
    try {
      const updated = await updateTodoStatus(t.id, newStatus, selectedTenantId)
      setTodos(prev => prev.map(x => x.id === updated.id ? updated : x))
    } catch (e) {
      setError(e instanceof Error ? e.message : '更新失败')
    }
  }

  const handlePriorityClick = async (t: TodoItem, p: TodoPriority) => {
    if (!ensureTenant()) return
    if (t.priority === p) return
    try {
      const updated = await updateTodoPriority(t.id, p, selectedTenantId)
      setTodos(prev => prev.map(x => x.id === updated.id ? updated : x))
    } catch (e) {
      setError(e instanceof Error ? e.message : '更新失败')
    }
  }

  const handleDelete = async (id: string) => {
    if (!ensureTenant()) return
    try {
      await deleteTodo(id, selectedTenantId)
      setTodos(prev => prev.filter(t => t.id !== id))
      setTotal(t => t - 1)
    } catch (e) {
      setError(e instanceof Error ? e.message : '删除失败')
    } finally {
      setDeletingId(null)
    }
  }

  const handleCreateCategory = async () => {
    if (!ensureTenant()) return
    if (!catForm.name.trim()) return
    setCatSaving(true)
    try {
      await createCategory({ name: catForm.name.trim(), color: catForm.color, tenant_id: selectedTenantId })
      await fetchCategories()
      setCatDialogOpen(false)
      setCatForm({ name: '', color: CATEGORY_COLORS[0] })
    } catch (e) {
      setError(e instanceof Error ? e.message : '创建分类失败')
    } finally {
      setCatSaving(false)
    }
  }

  const handleDeleteCategory = async (id: string) => {
    if (!ensureTenant()) return
    try {
      await deleteCategory(id, selectedTenantId)
      setCategories(prev => prev.filter(c => c.id !== id))
      if (activeCategoryId === id) {
        setActiveCategoryId('')
        setPage(1)
        fetchTodos(1)
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : '删除分类失败')
    }
  }

  const totalPages = Math.ceil(total / PAGE_SIZE)

  // ─── render ──────────────────────────────────────────────────────────────────

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '待办事项' },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-slate-50">
      {/* Header */}
      <header className="bg-white/95 backdrop-blur-xl border-b border-slate-200/20 shadow-sm sticky top-0 z-10">
        <div className="px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
              Auth-Perm
            </h1>
            <AvatarDropdown user={user ?? null} />
          </div>
        </div>
      </header>

      <div className="flex">
        <DashboardSidebar pathname="/todos" />

        <main className="flex-1 p-8">
          <Breadcrumb items={breadcrumbItems} />

          {error && (
            <div className="bg-red-50 text-red-600 p-3 rounded mb-4 flex items-center gap-2">
              <AlertCircle className="h-4 w-4 shrink-0" />
              {error}
              <button className="ml-auto" onClick={() => setError('')}><X className="h-4 w-4" /></button>
            </div>
          )}

          <div className="flex gap-6 mt-4">
            {/* ── Left: Category panel ── */}
            <aside className="w-52 shrink-0">
              <div className="bg-white rounded-xl border border-slate-200/60 shadow-sm overflow-hidden">
                <div className="px-4 py-3 border-b border-slate-100 flex items-center justify-between">
                  <span className="text-sm font-semibold text-slate-700">分类</span>
                  <button
                    className={`text-blue-600 hover:text-blue-700 transition-colors ${!tenantReady ? 'opacity-50 cursor-not-allowed' : ''}`}
                    onClick={() => tenantReady && setCatDialogOpen(true)}
                    title="新建分类"
                    disabled={!tenantReady}
                  >
                    <Plus className="h-4 w-4" />
                  </button>
                </div>

                <nav className="p-2 space-y-0.5">
                  {/* All */}
                  <CategoryNavItem
                    active={activeCategoryId === ''}
                    onClick={() => setActiveCategoryId('')}
                    color="#94a3b8"
                    label="全部"
                  />
                  {/* Uncategorised */}
                  <CategoryNavItem
                    active={activeCategoryId === 'none'}
                    onClick={() => setActiveCategoryId('none')}
                    color="#cbd5e1"
                    label="未分类"
                  />
                  {/* User categories */}
                  {categories.map(cat => (
                    <div key={cat.id} className="group relative">
                      <CategoryNavItem
                        active={activeCategoryId === cat.id}
                        onClick={() => setActiveCategoryId(cat.id)}
                        color={cat.color}
                        label={cat.name}
                      />
                      <button
                        className={`absolute right-2 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-opacity text-slate-400 hover:text-red-500 ${!tenantReady ? 'cursor-not-allowed' : ''}`}
                        onClick={() => tenantReady && handleDeleteCategory(cat.id)}
                        disabled={!tenantReady}
                      >
                        <X className="h-3 w-3" />
                      </button>
                    </div>
                  ))}
                </nav>
              </div>
            </aside>

            {/* ── Right: Todo list ── */}
            <div className="flex-1 min-w-0">
              {/* Toolbar */}
              <div className="flex flex-wrap gap-2 mb-4 items-center">
                <div className="flex gap-2 flex-1 min-w-0">
                  <Input
                    placeholder="搜索待办..."
                    value={keyword}
                    onChange={e => setKeyword(e.target.value)}
                    onKeyDown={e => e.key === 'Enter' && handleSearch()}
                    className="max-w-xs"
                  />
                  <Button variant="outline" onClick={handleSearch} disabled={!tenantReady}>
                    <Search className="h-4 w-4" />
                  </Button>
                </div>

                <Select value={filterStatus || '_all'} onValueChange={v => setFilterStatus(v === '_all' ? '' : v)}>
                  <SelectTrigger className="w-[110px]"><SelectValue placeholder="状态" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="_all">全部状态</SelectItem>
                    <SelectItem value="pending">待处理</SelectItem>
                    <SelectItem value="in_progress">进行中</SelectItem>
                    <SelectItem value="completed">已完成</SelectItem>
                    <SelectItem value="cancelled">已取消</SelectItem>
                  </SelectContent>
                </Select>

                <Select value={filterPriority || '_all'} onValueChange={v => setFilterPriority(v === '_all' ? '' : v)}>
                  <SelectTrigger className="w-[110px]"><SelectValue placeholder="优先级" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="_all">全部优先级</SelectItem>
                    <SelectItem value="urgent">紧急</SelectItem>
                    <SelectItem value="high">高</SelectItem>
                    <SelectItem value="medium">中</SelectItem>
                    <SelectItem value="low">低</SelectItem>
                  </SelectContent>
                </Select>

                <Button onClick={openCreate} disabled={!tenantReady}>
                  <Plus className="h-4 w-4 mr-1" />
                  新建待办
                </Button>
              </div>

              {/* List */}
              <div className="space-y-2">
                {loading ? (
                  <div className="text-center py-12 text-slate-400">加载中...</div>
                ) : todos.length === 0 ? (
                  <div className="text-center py-12 text-slate-400">
                    <CheckSquare className="h-12 w-12 mx-auto mb-3 opacity-30" />
                    暂无待办
                  </div>
                ) : (
                  todos.map(todo => (
                    <TodoRow
                      key={todo.id}
                      todo={todo}
                      categories={categories}
                      onToggle={handleToggleComplete}
                      onEdit={openEdit}
                      onPriority={handlePriorityClick}
                      onDelete={id => setDeletingId(id)}
                      disabled={!tenantReady}
                    />
                  ))
                )}
              </div>

              {/* Pagination */}
              {total > PAGE_SIZE && (
                <div className="flex justify-between items-center mt-4">
                  <span className="text-sm text-slate-500">共 {total} 条，第 {page}/{totalPages} 页</span>
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm" disabled={page <= 1}
                      onClick={() => { const p = page - 1; setPage(p); fetchTodos(p) }}>上一页</Button>
                    <Button variant="outline" size="sm" disabled={page >= totalPages}
                      onClick={() => { const p = page + 1; setPage(p); fetchTodos(p) }}>下一页</Button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </main>
      </div>

      {/* ── Todo Dialog (Create / Edit) ── */}
      <Dialog open={todoDialogOpen} onOpenChange={setTodoDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingTodo ? '编辑待办' : '新建待办'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 pt-1">
            <div>
              <Label>标题 *</Label>
              <Input
                value={todoForm.title}
                onChange={e => setTodoForm(f => ({ ...f, title: e.target.value }))}
                placeholder="待办标题"
                autoFocus
              />
            </div>
            <div>
              <Label>描述</Label>
              <textarea
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm min-h-[80px] resize-none focus:outline-none focus:ring-1 focus:ring-ring"
                value={todoForm.description}
                onChange={e => setTodoForm(f => ({ ...f, description: e.target.value }))}
                placeholder="可选描述..."
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <Label>优先级</Label>
                <Select value={todoForm.priority} onValueChange={v => setTodoForm(f => ({ ...f, priority: v as TodoPriority }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="low">低</SelectItem>
                    <SelectItem value="medium">中</SelectItem>
                    <SelectItem value="high">高</SelectItem>
                    <SelectItem value="urgent">紧急</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>分类</Label>
                <Select value={todoForm.category_id || '_none'} onValueChange={v => setTodoForm(f => ({ ...f, category_id: v === '_none' ? '' : v }))}>
                  <SelectTrigger><SelectValue placeholder="未分类" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="_none">未分类</SelectItem>
                    {categories.map(c => (
                      <SelectItem key={c.id} value={c.id}>
                        <span className="flex items-center gap-1.5">
                          <span className="inline-block w-2 h-2 rounded-full" style={{ background: c.color }} />
                          {c.name}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div>
              <Label>截止时间（可选）</Label>
              <Input
                type="datetime-local"
                value={todoForm.deadline}
                onChange={e => setTodoForm(f => ({ ...f, deadline: e.target.value }))}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setTodoDialogOpen(false)}>取消</Button>
            <Button onClick={handleSaveTodo} disabled={todoSaving}>
              {todoSaving ? '保存中...' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ── Category Dialog ── */}
      <Dialog open={catDialogOpen} onOpenChange={setCatDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader><DialogTitle>新建分类</DialogTitle></DialogHeader>
          <div className="space-y-4 pt-1">
            <div>
              <Label>名称 *</Label>
              <Input
                value={catForm.name}
                onChange={e => setCatForm(f => ({ ...f, name: e.target.value }))}
                placeholder="分类名称"
                autoFocus
              />
            </div>
            <div>
              <Label>颜色</Label>
              <div className="flex flex-wrap gap-2 mt-1.5">
                {CATEGORY_COLORS.map(c => (
                  <button
                    key={c}
                    className={`w-7 h-7 rounded-full transition-all ${catForm.color === c ? 'ring-2 ring-offset-2 ring-slate-400 scale-110' : 'hover:scale-110'}`}
                    style={{ background: c }}
                    onClick={() => setCatForm(f => ({ ...f, color: c }))}
                  />
                ))}
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCatDialogOpen(false)}>取消</Button>
            <Button onClick={handleCreateCategory} disabled={catSaving || !catForm.name.trim()}>
              {catSaving ? '创建中...' : '创建'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ── Delete Confirm Dialog ── */}
      <Dialog open={!!deletingId} onOpenChange={open => !open && setDeletingId(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader><DialogTitle>确认删除</DialogTitle></DialogHeader>
          <p className="text-sm text-slate-600 py-2">确定要删除这条待办吗？此操作不可撤销。</p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeletingId(null)}>取消</Button>
            <Button variant="destructive" onClick={() => deletingId && handleDelete(deletingId)}>删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

// ─── sub-components ────────────────────────────────────────────────────────────

function CategoryNavItem({
  active, onClick, color, label,
}: { active: boolean; onClick: () => void; color: string; label: string }) {
  return (
    <button
      onClick={onClick}
      className={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-all ${
        active
          ? 'bg-blue-50 text-blue-700 font-medium'
          : 'text-slate-600 hover:bg-slate-50'
      }`}
    >
      <span className="inline-block w-2.5 h-2.5 rounded-full shrink-0" style={{ background: color }} />
      <span className="truncate">{label}</span>
    </button>
  )
}

function TodoRow({
  todo, categories, onToggle, onEdit, onPriority, onDelete, disabled,
}: {
  todo: TodoItem
  categories: TodoCategory[]
  onToggle: (t: TodoItem) => void
  onEdit: (t: TodoItem) => void
  onPriority: (t: TodoItem, p: TodoPriority) => void
  onDelete: (id: string) => void
  disabled?: boolean
}) {
  const [priorityOpen, setPriorityOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!priorityOpen) return
    const onClick = (evt: MouseEvent) => {
      const target = evt.target as Node
      if (menuRef.current && !menuRef.current.contains(target)) {
        setPriorityOpen(false)
      }
    }
    document.addEventListener('mousedown', onClick)
    return () => document.removeEventListener('mousedown', onClick)
  }, [priorityOpen])

  const done = todo.status === 'completed' || todo.status === 'cancelled'
  const cat = categories.find(c => c.id === todo.category_id)

  return (
    <div ref={menuRef} className={`group bg-white rounded-xl border shadow-sm px-4 py-3 flex items-start gap-3 transition-all hover:shadow-md ${
      done ? 'opacity-60' : ''
    } ${todo.is_overdue && !done ? 'border-red-200' : 'border-slate-200/60'}`}>

      {/* Checkbox */}
      <button
        className={`mt-0.5 shrink-0 text-slate-400 hover:text-blue-600 transition-colors ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
        onClick={() => !disabled && onToggle(todo)}
        disabled={disabled}
      >
        {todo.status === 'completed'
          ? <CheckSquare className="h-5 w-5 text-blue-600" />
          : <Square className="h-5 w-5" />
        }
      </button>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <span className={`text-sm font-medium ${done ? 'line-through text-slate-400' : 'text-slate-800'}`}>
            {todo.title}
          </span>

          {/* Priority badge — clickable to cycle */}
          <div className="relative">
            <button
              className={`inline-flex items-center gap-1 text-xs px-1.5 py-0.5 rounded font-medium ${PRIORITY_COLOR[todo.priority]} transition-opacity hover:opacity-80 ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
              onClick={() => !disabled && setPriorityOpen(p => !p)}
              title="点击更改优先级"
              disabled={disabled}
            >
              {todo.priority === 'urgent' && <AlertCircle className="h-3 w-3" />}
              {PRIORITY_LABEL[todo.priority]}
              <ChevronDown className="h-3 w-3" />
            </button>
            {priorityOpen && (
              <div className="absolute z-20 top-full left-0 mt-1 bg-white border border-slate-200 rounded-lg shadow-lg py-1 min-w-[90px]">
                {(['urgent', 'high', 'medium', 'low'] as TodoPriority[]).map(p => (
                  <button
                    key={p}
                    className={`w-full text-left px-3 py-1.5 text-xs hover:bg-slate-50 ${todo.priority === p ? 'font-semibold' : ''}`}
                    onClick={() => { onPriority(todo, p); setPriorityOpen(false) }}
                  >
                    {PRIORITY_LABEL[p]}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Status badge */}
          {todo.status !== 'pending' && (
            <span className="text-xs px-1.5 py-0.5 rounded bg-slate-100 text-slate-500">
              {STATUS_LABEL[todo.status]}
            </span>
          )}
        </div>

        {/* Meta row */}
        <div className="flex items-center gap-3 mt-1 flex-wrap">
          {cat && (
            <span className="flex items-center gap-1 text-xs text-slate-500">
              <span className="inline-block w-2 h-2 rounded-full" style={{ background: cat.color }} />
              {cat.name}
            </span>
          )}
          {todo.deadline && (
            <span className={`flex items-center gap-1 text-xs ${
              todo.is_overdue && !done ? 'text-red-500 font-medium' : isToday(todo.deadline) ? 'text-orange-500' : 'text-slate-400'
            }`}>
              <Calendar className="h-3 w-3" />
              {todo.is_overdue && !done ? '已过期 ' : ''}
              {fmtDate(todo.deadline)}
            </span>
          )}
          {todo.description && (
            <span className="text-xs text-slate-400 truncate max-w-[200px]">{todo.description}</span>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
        <button className="p-1 text-slate-400 hover:text-blue-600 transition-colors" onClick={() => !disabled && onEdit(todo)} disabled={disabled}>
          <Pencil className="h-4 w-4" />
        </button>
        <button className="p-1 text-slate-400 hover:text-red-500 transition-colors" onClick={() => !disabled && onDelete(todo.id)} disabled={disabled}>
          <Trash2 className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}
