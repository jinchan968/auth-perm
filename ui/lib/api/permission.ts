// Permission API Client

import {
  Permission,
  PermissionListItem,
  PermissionListResponse,
  CreatePermissionRequest,
  UpdatePermissionRequest,
} from '@/types/permission'
import { apiClient } from './client'

const API_BASE = '/permissions/items'

// List permissions
export async function listPermissions(params: {
  tenant_id: string
  keyword?: string
  page?: number
  size?: number
}): Promise<PermissionListResponse> {
  const searchParams = new URLSearchParams()
  searchParams.set('tenant_id', params.tenant_id)
  if (params.keyword) searchParams.set('keyword', params.keyword)
  if (params.page) searchParams.set('page', params.page.toString())
  if (params.size) searchParams.set('page_size', params.size.toString())

  const data = await apiClient.get<PermissionListResponse>(`${API_BASE}?${searchParams.toString()}`)
  return data
}

export async function listAllPermissions(params: {
  tenant_id: string
  keyword?: string
  size?: number
}): Promise<PermissionListItem[]> {
  const pageSize = params.size || 100
  const firstPage = await listPermissions({ ...params, page: 1, size: pageSize })
  const permissions = [...(firstPage.data || [])]
  const totalPages = Math.ceil((firstPage.total || permissions.length) / (firstPage.size || pageSize))

  if (totalPages <= 1) {
    return permissions
  }

  const restPages = await Promise.all(
    Array.from({ length: totalPages - 1 }, (_, index) =>
      listPermissions({ ...params, page: index + 2, size: pageSize })
    )
  )

  restPages.forEach((page) => {
    permissions.push(...(page.data || []))
  })

  return permissions
}

// Get permission by ID
export async function getPermission(id: string, tenantId?: string): Promise<Permission> {
  const url = tenantId ? `${API_BASE}/${id}?tenant_id=${tenantId}` : `${API_BASE}/${id}`
  const data = await apiClient.get<Permission>(url)
  return data
}

// Create permission
export async function createPermission(request: CreatePermissionRequest): Promise<Permission> {
  const data = await apiClient.post<Permission>(API_BASE, request)
  return data
}

// Update permission
export async function updatePermission(id: string, tenantId: string, request: UpdatePermissionRequest): Promise<Permission> {
  const data = await apiClient.put<Permission>(`${API_BASE}/${id}`, {
    ...request,
    tenant_id: tenantId,
  })
  return data
}

// Delete permission
export async function deletePermission(id: string, tenantId: string): Promise<void> {
  await apiClient.delete(`${API_BASE}/${id}?tenant_id=${tenantId}`)
}
