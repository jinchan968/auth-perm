// Tenant API Client

import {
  Tenant,
  TenantListItem,
  TenantListResponse,
  CreateTenantRequest,
  UpdateTenantRequest,
  UpdateTenantSettingsRequest,
  TenantSettings,
} from '@/types/tenant'
import { apiClient } from './client'

const API_BASE = '/tenants'

// List tenants
export async function listTenants(params: {
  keyword?: string
  status?: string
  page?: number
  size?: number
}): Promise<TenantListResponse> {
  const searchParams = new URLSearchParams()
  if (params.keyword) searchParams.set('keyword', params.keyword)
  if (params.status) searchParams.set('status', params.status)
  if (params.page) searchParams.set('page', params.page.toString())
  if (params.size) searchParams.set('size', params.size.toString())

  const data = await apiClient.get<TenantListResponse>(`${API_BASE}?${searchParams.toString()}`)
  return data
}

// Get tenant by ID
export async function getTenant(id: string): Promise<Tenant> {
  const data = await apiClient.get<Tenant>(`${API_BASE}/${id}`)
  return data
}

// Create tenant
export async function createTenant(request: CreateTenantRequest): Promise<Tenant> {
  const data = await apiClient.post<Tenant>(API_BASE, request)
  return data
}

// Update tenant
export async function updateTenant(id: string, request: UpdateTenantRequest): Promise<Tenant> {
  const data = await apiClient.put<Tenant>(`${API_BASE}/${id}`, request)
  return data
}

// Delete tenant (disable tenant)
export async function deleteTenant(id: string): Promise<void> {
  await apiClient.delete(`${API_BASE}/${id}`)
}

// Change tenant status (启用/暂停/删除租户)
export async function changeTenantStatus(id: string, status: 'active' | 'suspended' | 'deleted'): Promise<void> {
  await apiClient.post(`${API_BASE}/${id}/change-status`, { status })
}

// Get tenant settings
export async function getTenantSettings(id: string): Promise<TenantSettings> {
  const data = await apiClient.get<TenantSettings>(`${API_BASE}/${id}/settings`)
  return data
}

// Update tenant settings
export async function updateTenantSettings(
  id: string,
  request: UpdateTenantSettingsRequest
): Promise<Tenant> {
  const data = await apiClient.put<Tenant>(`${API_BASE}/${id}/settings`, request)
  return data
}
