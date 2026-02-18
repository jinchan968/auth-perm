// Tenant Types

export type TenantStatus = 'active' | 'suspended' | 'deleted'

export type TenantPlan = 'free' | 'basic' | 'pro' | 'enterprise'

export interface TenantSettings {
  features: FeaturesConfig
  quota: QuotaConfig
  custom?: Record<string, unknown>
}

export interface FeaturesConfig {
  email_verification: boolean
  oauth_login: boolean
  totp_enabled: boolean
  session_limit: boolean
  password_complexity: boolean
}

export interface QuotaConfig {
  max_users: number
  max_roles: number
  max_organizations: number
  max_sessions_per_user: number
  api_rate_limit: number
}

export interface Tenant {
  id: string
  name: string
  code: string
  status: TenantStatus
  plan: TenantPlan
  expire_at?: string | null
  settings: TenantSettings
  created_at: string
  updated_at: string
}

export interface TenantListItem {
  id: string
  name: string
  code: string
  status: TenantStatus
  plan: TenantPlan
  expire_at?: string | null
  user_count: number
  created_at: string
}

export interface TenantListResponse {
  data: TenantListItem[]
  total: number
  page: number
  size: number
}

export interface CreateTenantRequest {
  name: string
  code?: string
  plan?: TenantPlan
  expire_at?: string
}

export interface UpdateTenantRequest {
  name?: string
  status?: TenantStatus
  plan?: TenantPlan
  expire_at?: string
  settings?: TenantSettings
}

export interface UpdateTenantSettingsRequest {
  settings: TenantSettings
}

export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
}
