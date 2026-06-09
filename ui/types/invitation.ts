export interface InvitationItem {
  id: string
  code_preview: string
  tenant_id: string
  status: 'active' | 'used' | 'invalidated' | 'expired'
  expires_at: string
  used_at?: string
  used_by_account_id?: string
  created_by_account_id: string
  invalidated_at?: string
  invalidated_by_account_id?: string
  created_at: string
  updated_at: string
}

export interface InvitationListResponse {
  data: InvitationItem[]
  total: number
  page: number
  size: number
}

export interface CreateInvitationRequest {
  tenant_id?: string
  expires_at?: string
}

export interface CreateInvitationResponse extends InvitationItem {
  invite_code: string
  invite_url: string
}
