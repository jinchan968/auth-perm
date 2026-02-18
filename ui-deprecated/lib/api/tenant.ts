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

const API_BASE = '/api/tenants'

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

  const res = await fetch(`${API_BASE}?${searchParams.toString()}`, {
    method: 'GET',
    credentials: 'include',
  })

  if (!res.ok) {
    const error = await res.json()
    throw new Error(error.message || 'Failed to fetch tenants')
  }

  const data = await res.json()
  return data.data
}

// Get tenant by ID
export async function getTenant(id: string): Promise<Tenant> {
  const res = await fetch(`${API_BASE}/${id}`, {
    method: 'GET',
    credentials: 'include',
  })

  if (!res.ok) {
    const error = await res.json()
    throw new Error(error.message || 'Failed to fetch tenant')
  }

  const data = await res.json()
  return data.data
}

// Create tenant
export async function createTenant(request: CreateTenantRequest): Promise<Tenant> {
  const res = await fetch(API_BASE, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
    credentials: 'include',
  })

  if (!res.ok) {
    const error = await res.json()
    throw new Error(error.message || 'Failed to create tenant')
  }

  const data = await res.json()
  return data.data
}

// Update tenant
export async function updateTenant(id: string, request: UpdateTenantRequest): Promise<Tenant> {
  const res = await fetch(`${API_BASE}/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
    credentials: 'include',
  })

  if (!res.ok) {
    const error = await res.json()
    throw new Error(error.message || 'Failed to update tenant')
  }

  const data = await res.json()
  return data.data
}

// Delete tenant
export async function deleteTenant(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/${id}`, {
    method: 'DELETE',
    credentials: 'include',
  })

  if (!res.ok) {
    const error = await res.json()
    throw new Error(error.message || 'Failed to delete tenant')
  }
}

// Get tenant settings
export async function getTenantSettings(id: string): Promise<TenantSettings> {
  const res = await fetch(`${API_BASE}/${id}/settings`, {
    method: 'GET',
    credentials: 'include',
  })

  if (!res.ok) {
    const error = await res.json()
    throw new Error(error.message || 'Failed to fetch tenant settings')
  }

  const data = await res.json()
  return data.data
}

// Update tenant settings
export async function updateTenantSettings(
  id: string,
  request: UpdateTenantSettingsRequest
): Promise<Tenant> {
  const res = await fetch(`${API_BASE}/${id}/settings`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
    credentials: 'include',
  })

  if (!res.ok) {
    const error = await res.json()
    throw new Error(error.message || 'Failed to update tenant settings')
  }

  const data = await res.json()
  return data.data
}
