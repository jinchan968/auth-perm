// Role API Client

import {
  AssignPermissionToRoleRequest,
  CreateRoleRequest,
  PermissionListItem,
  Role,
  RoleListItem,
  RoleListResponse,
  UpdateRoleRequest,
} from '@/types/permission'
import {apiClient} from './client'

const API_BASE = '/permissions/roles'

// List roles
export async function listRoles(params: {
  tenant_id: string
  keyword?: string
  page?: number
  size?: number
}): Promise<RoleListResponse> {
  const searchParams = new URLSearchParams()
  searchParams.set('tenant_id', params.tenant_id)
  if (params.keyword) searchParams.set('keyword', params.keyword)
  if (params.page) searchParams.set('page', params.page.toString())
  if (params.size) searchParams.set('size', params.size.toString())

  const data = await apiClient.get<RoleListResponse>(`${API_BASE}?${searchParams.toString()}`)
  return data
}

// Get role by ID
export async function getRole(id: string, tenantId: string): Promise<Role> {
  const data = await apiClient.get<Role>(`${API_BASE}/${id}?tenant_id=${tenantId}`)
  return data
}

// Create role
export async function createRole(request: CreateRoleRequest): Promise<Role> {
  const data = await apiClient.post<Role>(API_BASE, request)
  return data
}

// Update role
export async function updateRole(id: string, tenantId: string, request: UpdateRoleRequest): Promise<Role> {
  const data = await apiClient.put<Role>(`${API_BASE}/${id}`, {
    ...request,
    tenant_id: tenantId,
  })
  return data
}

// Delete role
export async function deleteRole(id: string, tenantId: string): Promise<void> {
  await apiClient.delete(`${API_BASE}/${id}?tenant_id=${tenantId}`)
}

// Get role permissions
export async function getRolePermissions(roleId: string, tenantId: string): Promise<PermissionListItem[]> {
  const data = await apiClient.get<{ data: PermissionListItem[] }>(`${API_BASE}/${roleId}/permissions?tenant_id=${tenantId}`)
  return data.data || []
}

// Assign permissions to role
export async function assignPermissionsToRole(request: AssignPermissionToRoleRequest): Promise<void> {
  await apiClient.post(`${API_BASE}/${request.role_id}/permissions`, {
    role_id: request.role_id,
    permission_ids: request.permission_ids,
    tenant_id: request.tenant_id,
  })
}

// Remove permission from role
export async function removePermissionFromRole(roleId: string, permissionId: string, tenantId: string): Promise<void> {
  await apiClient.delete(`${API_BASE}/${roleId}/permissions/${permissionId}?tenant_id=${tenantId}`)
}

// ==================== Account-Role APIs ====================

const ACCOUNTS_BASE = '/permissions/accounts'

// Assign role to account
export async function assignRoleToAccount(request: {
  account_id: string
  role_ids: string[]
  tenant_id: string
}): Promise<void> {
  await apiClient.post(`${ACCOUNTS_BASE}/${request.account_id}/roles`, {
    role_ids: request.role_ids,
    tenant_id: request.tenant_id,
  })
}

// Remove role from account
export async function removeRoleFromAccount(accountId: string, roleId: string, tenantId: string): Promise<void> {
  await apiClient.delete(`${ACCOUNTS_BASE}/${accountId}/roles/${roleId}?tenant_id=${tenantId}`)
}

// Get user roles (for account)
export async function getUserRoles(accountId: string, tenantId?: string): Promise<RoleListItem[]> {
  const query = tenantId ? `?tenant_id=${tenantId}` : ''
  return await apiClient.get<RoleListItem[]>(`${ACCOUNTS_BASE}/${accountId}/roles${query}`)
}
