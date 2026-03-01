// Todo System Types

export type TodoStatus = 'pending' | 'in_progress' | 'completed' | 'cancelled'
export type TodoPriority = 'low' | 'medium' | 'high' | 'urgent'

export interface TodoCategory {
  id: string
  tenant_id: string
  account_id: string
  name: string
  color: string
  icon?: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface TodoItem {
  id: string
  tenant_id: string
  account_id: string
  category_id?: string
  category?: TodoCategory
  title: string
  description?: string
  status: TodoStatus
  priority: TodoPriority
  deadline?: string
  completed_at?: string
  sort_order: number
  is_overdue: boolean
  created_at: string
  updated_at: string
}

export interface TodoListResponse {
  data: TodoItem[]
  total: number
  page: number
  size: number
}

export interface CategoryListResponse {
  data: TodoCategory[]
}

export interface CreateTodoRequest {
  title: string
  description?: string
  priority?: TodoPriority
  category_id?: string
  deadline?: string
}

export interface UpdateTodoRequest {
  title?: string
  description?: string
  priority?: TodoPriority
  category_id?: string
  clear_category?: boolean
  deadline?: string
  clear_deadline?: boolean
}

export interface CreateCategoryRequest {
  name: string
  color?: string
  icon?: string
}

export interface UpdateCategoryRequest {
  name?: string
  color?: string
  icon?: string
}
