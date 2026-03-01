import { apiClient } from './client'
import {
  TodoItem,
  TodoCategory,
  TodoListResponse,
  CreateTodoRequest,
  UpdateTodoRequest,
  CreateCategoryRequest,
  UpdateCategoryRequest,
  TodoStatus,
  TodoPriority,
} from '@/types/todo'

const BASE = '/todos'

// -------- Categories --------

export async function listCategories(tenantId: string): Promise<TodoCategory[]> {
  const data = await apiClient.get<{ data: TodoCategory[] }>(
    `${BASE}/categories?tenant_id=${tenantId}`
  )
  return data.data || []
}

export async function createCategory(
  req: CreateCategoryRequest & { tenant_id: string }
): Promise<TodoCategory> {
  return apiClient.post<TodoCategory>(
    `${BASE}/categories?tenant_id=${req.tenant_id}`,
    req
  )
}

export async function updateCategory(
  id: string,
  req: UpdateCategoryRequest & { tenant_id: string }
): Promise<TodoCategory> {
  return apiClient.put<TodoCategory>(
    `${BASE}/categories/${id}?tenant_id=${req.tenant_id}`,
    req
  )
}

export async function deleteCategory(id: string, tenantId: string): Promise<void> {
  await apiClient.delete(`${BASE}/categories/${id}?tenant_id=${tenantId}`)
}

// -------- Todos --------

export async function listTodos(params: {
  tenant_id: string
  status?: TodoStatus
  priority?: TodoPriority
  category_id?: string
  keyword?: string
  page?: number
  page_size?: number
}): Promise<TodoListResponse> {
  const q = new URLSearchParams()
  q.set('tenant_id', params.tenant_id)
  if (params.status)      q.set('status', params.status)
  if (params.priority)    q.set('priority', params.priority)
  if (params.category_id) q.set('category_id', params.category_id)
  if (params.keyword)     q.set('keyword', params.keyword)
  if (params.page)        q.set('page', String(params.page))
  if (params.page_size)   q.set('page_size', String(params.page_size))

  return apiClient.get<TodoListResponse>(`${BASE}?${q.toString()}`)
}

export async function getTodo(id: string, tenantId: string): Promise<TodoItem> {
  return apiClient.get<TodoItem>(`${BASE}/${id}?tenant_id=${tenantId}`)
}

export async function createTodo(
  req: CreateTodoRequest & { tenant_id: string }
): Promise<TodoItem> {
  return apiClient.post<TodoItem>(`${BASE}?tenant_id=${req.tenant_id}`, req)
}

export async function updateTodo(
  id: string,
  req: UpdateTodoRequest & { tenant_id: string }
): Promise<TodoItem> {
  return apiClient.put<TodoItem>(`${BASE}/${id}?tenant_id=${req.tenant_id}`, req)
}

export async function updateTodoStatus(
  id: string,
  status: TodoStatus,
  tenantId: string
): Promise<TodoItem> {
  return apiClient.patch<TodoItem>(`${BASE}/${id}/status?tenant_id=${tenantId}`, { status })
}

export async function updateTodoPriority(
  id: string,
  priority: TodoPriority,
  tenantId: string
): Promise<TodoItem> {
  return apiClient.patch<TodoItem>(`${BASE}/${id}/priority?tenant_id=${tenantId}`, { priority })
}

export async function deleteTodo(id: string, tenantId: string): Promise<void> {
  await apiClient.delete(`${BASE}/${id}?tenant_id=${tenantId}`)
}
