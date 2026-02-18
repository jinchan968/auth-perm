// Permission Types

export interface Permission {
  id: string
  tenant_id: string
  code: string
  name: string
  description: string
  resource: string
  is_system: boolean
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface PermissionListItem {
  id: string
  code: string
  name: string
  description: string
  resource: string
  is_system: boolean
  is_active: boolean
}

export interface PermissionListResponse {
  data: PermissionListItem[]
  total: number
  page: number
  size: number
}

export interface CreatePermissionRequest {
  tenant_id: string
  code?: string
  name: string
  description?: string
  is_system?: boolean
}

export interface UpdatePermissionRequest {
  id: string
  name?: string
  description?: string
  is_active?: boolean
}

export interface Role {
  id: string
  tenant_id: string
  org_id: string
  code: string
  name: string
  description: string
  priority: number
  is_system: boolean
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface RoleListItem {
  id: string
  code: string
  name: string
  description: string
  priority: number
  is_system: boolean
  is_active: boolean
  permission_count?: number
}

export interface RoleListResponse {
  data: RoleListItem[]
  total: number
  page: number
  size: number
}

export interface CreateRoleRequest {
  tenant_id: string
  org_id?: string
  code?: string
  name: string
  description?: string
  priority?: number
  is_system?: boolean
}

export interface UpdateRoleRequest {
  id: string
  name?: string
  description?: string
  priority?: number
  is_active?: boolean
}

export interface AssignPermissionToRoleRequest {
  role_id: string
  permission_ids: string[]
  tenant_id: string
}

export interface AssignRoleToAccountRequest {
  account_id: string
  role_ids: string[]
  tenant_id: string
}

export interface UserWithRoles {
  id: string
  username: string
  email: string
  roles: RoleListItem[]
}
