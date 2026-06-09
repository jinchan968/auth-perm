import { apiClient } from './client'
import type { InvitationListResponse, CreateInvitationRequest, CreateInvitationResponse } from '@/types/invitation'

const BASE = '/auth/invitations'

export async function listInvitations(params: {
  tenant_id?: string
  status?: string
  page?: number
  size?: number
}) {
  const query = new URLSearchParams()
  if (params.tenant_id) query.set('tenant_id', params.tenant_id)
  if (params.status) query.set('status', params.status)
  if (params.page !== undefined) query.set('page', String(params.page))
  if (params.size !== undefined) query.set('page_size', String(params.size))
  const suffix = query.toString()
  return apiClient.get<InvitationListResponse>(suffix ? `${BASE}?${suffix}` : BASE)
}

export async function createInvitation(data: CreateInvitationRequest) {
  return apiClient.post<CreateInvitationResponse>(BASE, data)
}

export async function invalidateInvitation(id: string) {
  return apiClient.post<{ message: string }>(`${BASE}/${id}/invalidate`, {})
}
