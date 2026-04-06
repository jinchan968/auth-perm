import type { AuditLogEntry, AuditLogValues, LoginLogsResponse } from '@/types/security-log'

interface RawAuditLogValues {
  changedFields?: Record<string, unknown>
  changed_fields?: Record<string, unknown>
  context?: Record<string, unknown>
  metadata?: Record<string, unknown>
  tags?: string[]
  ipAddress?: string
  ip_address?: string
  userAgent?: string
  user_agent?: string
  sessionId?: string
  session_id?: string
  requestId?: string
  request_id?: string
  correlationId?: string
  correlation_id?: string
}

interface RawAuditLogEntry {
  id?: string
  tenantId?: string
  tenant_id?: string
  userId?: string
  user_id?: string
  action?: string
  resourceType?: string
  resource_type?: string
  resourceId?: string
  resource_id?: string
  oldValues?: RawAuditLogValues | null
  old_values?: RawAuditLogValues | null
  newValues?: RawAuditLogValues | null
  new_values?: RawAuditLogValues | null
  ipAddress?: string
  ip_address?: string
  userAgent?: string
  user_agent?: string
  success?: boolean
  errorMessage?: string
  error_message?: string
  createdAt?: string
  created_at?: string
}

interface RawLoginLogsResponse {
  logs?: RawAuditLogEntry[]
  total?: number
  page?: number
  pageSize?: number
  page_size?: number
}

function normalizeAuditLogValues(values?: RawAuditLogValues | null): AuditLogValues | null {
  if (!values) {
    return null
  }

  return {
    changedFields: values.changedFields ?? values.changed_fields,
    context: values.context,
    metadata: values.metadata,
    tags: values.tags,
    ipAddress: values.ipAddress ?? values.ip_address,
    userAgent: values.userAgent ?? values.user_agent,
    sessionId: values.sessionId ?? values.session_id,
    requestId: values.requestId ?? values.request_id,
    correlationId: values.correlationId ?? values.correlation_id,
  }
}

export function normalizeAuditLogEntry(raw?: RawAuditLogEntry | null): AuditLogEntry {
  const oldValues = normalizeAuditLogValues(raw?.oldValues ?? raw?.old_values)
  const newValues = normalizeAuditLogValues(raw?.newValues ?? raw?.new_values)

  return {
    id: raw?.id ?? '',
    tenantId: raw?.tenantId ?? raw?.tenant_id ?? '',
    userId: raw?.userId ?? raw?.user_id ?? '',
    action: raw?.action ?? '',
    resourceType: raw?.resourceType ?? raw?.resource_type ?? '',
    resourceId: raw?.resourceId ?? raw?.resource_id ?? '',
    oldValues,
    newValues,
    ipAddress:
      raw?.ipAddress ??
      raw?.ip_address ??
      oldValues?.ipAddress ??
      newValues?.ipAddress ??
      '',
    userAgent:
      raw?.userAgent ??
      raw?.user_agent ??
      oldValues?.userAgent ??
      newValues?.userAgent ??
      '',
    success: raw?.success ?? false,
    errorMessage: raw?.errorMessage ?? raw?.error_message ?? '',
    createdAt: raw?.createdAt ?? raw?.created_at ?? '',
  }
}

export function normalizeLoginLogsResponse(raw?: RawLoginLogsResponse | null): LoginLogsResponse {
  const logs = Array.isArray(raw?.logs) ? raw.logs.map(normalizeAuditLogEntry) : []

  return {
    logs,
    total: raw?.total ?? logs.length,
    page: raw?.page ?? 1,
    pageSize: raw?.pageSize ?? raw?.page_size ?? logs.length,
  }
}

