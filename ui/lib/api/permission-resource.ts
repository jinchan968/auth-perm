// Permission Resource API Client

import { apiClient } from './client'

export interface PermissionResource {
  id: string
  permission_id: string
  resource_id: string
  resource_type: string
  resource_name: string
  tenant_id: string
  created_at: string
  updated_at: string
}

export interface PermissionResourceListResponse {
  data: PermissionResource[]
  total: number
  page: number
  size: number
}

export interface CreatePermissionResourceRequest {
  permission_id: string
  resource_id: string
  resource_type: string
  resource_name: string
  tenant_id: string
}

// List permission resources
export async function listPermissionResources(
  permissionId: string,
  params: {
    resource_type?: string
    page?: number
    size?: number
    tenant_id?: string
  }
): Promise<PermissionResourceListResponse> {
  const searchParams = new URLSearchParams()
  searchParams.set('permission_id', permissionId)
  if (params.resource_type) searchParams.set('resource_type', params.resource_type)
  if (params.page) searchParams.set('page', params.page.toString())
  if (params.size) searchParams.set('page_size', params.size.toString())
  if (params.tenant_id) searchParams.set('tenant_id', params.tenant_id)

  const data = await apiClient.get<PermissionResourceListResponse>(
    `/permissions/resources?${searchParams.toString()}`
  )
  return data
}

// Create permission resource
export async function createPermissionResource(
  permissionId: string,
  request: CreatePermissionResourceRequest
): Promise<PermissionResource> {
  const data = await apiClient.post<PermissionResource>(
    `/permissions/resources`,
    request
  )
  return data
}

// Delete permission resource
export async function deletePermissionResource(
  permissionId: string,
  resourceId: string,
  tenantId?: string
): Promise<void> {
  const url = tenantId
    ? `/permissions/resources/${resourceId}?tenant_id=${tenantId}`
    : `/permissions/resources/${resourceId}`
  await apiClient.delete(url)
}
