// User Management Types

export type AccountStatus = 'active' | 'inactive' | 'suspended'
export type AccountType = 'email' | 'phone' | 'github' | 'google' | 'wechat'
export type UserStatus = 'active' | 'inactive'

export interface AccountListItem {
  account_id: string
  tenant_id: string
  account_type: AccountType
  account_status: AccountStatus
  email_verified: boolean
  last_login_at?: string

  user_id: string
  username: string
  nickname: string
  avatar: string
  email: string
  phone: string
  user_status: UserStatus

  created_at: string
  updated_at: string
}

export interface AccountListResponse {
  data: AccountListItem[]
  total: number
  page: number
  size: number
}

export interface CreateUserRequest {
  identifier_type: 'email' | 'phone'
  email?: string
  phone?: string
  username: string
  password: string
  confirm_password: string
  tenant_id: string
  nickname?: string
}

export interface UpdateUserStatusRequest {
  tenant_id: string
  status: AccountStatus
}
