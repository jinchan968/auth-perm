// User Management API Client

import {
  AccountListItem,
  AccountListResponse,
  CreateUserRequest,
  UpdateUserStatusRequest,
  AccountStatus,
  AccountType,
} from '@/types/user'
import { apiClient } from './client'

const API_BASE = '/users'

// List users (accounts with user info)
export async function listUsers(params: {
  tenant_id: string
  keyword?: string
  status?: AccountStatus
  account_type?: AccountType
  page?: number
  page_size?: number
}): Promise<AccountListResponse> {
  const searchParams = new URLSearchParams()
  searchParams.set('tenant_id', params.tenant_id)
  if (params.keyword) searchParams.set('keyword', params.keyword)
  if (params.status) searchParams.set('status', params.status)
  if (params.account_type) searchParams.set('account_type', params.account_type)
  if (params.page) searchParams.set('page', params.page.toString())
  if (params.page_size) searchParams.set('page_size', params.page_size.toString())

  const data = await apiClient.get<AccountListResponse>(`${API_BASE}?${searchParams.toString()}`)
  return data
}

// Get user by account ID
export async function getUser(id: string, tenantId: string): Promise<AccountListItem> {
  const data = await apiClient.get<AccountListItem>(`${API_BASE}/${id}?tenant_id=${tenantId}`)
  return data
}

// Update user status
export async function updateUserStatus(
  id: string,
  request: UpdateUserStatusRequest
): Promise<void> {
  await apiClient.patch(`${API_BASE}/${id}/status`, request)
}

// Create user (admin function, equivalent to registration)
export async function createUser(request: CreateUserRequest): Promise<AccountListItem> {
  const data = await apiClient.post<{ data: AccountListItem }>(`${API_BASE}`, request)
  return data.data
}

// Get user accounts (all accounts for a user across tenants)
export async function getUserAccounts(userId: string): Promise<AccountListItem[]> {
  const data = await apiClient.get<{ data: AccountListItem[] }>(`${API_BASE}/${userId}/accounts`)
  return data.data || []
}
