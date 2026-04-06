'use client'

import { useQuery } from '@tanstack/react-query'
import { authApi } from '@/lib/api/auth'
import { normalizeAuditLogEntry, normalizeLoginLogsResponse } from '@/lib/api/security-log'
import { queryKeys } from '@/lib/query-keys'
import type { AuditLogEntry, LoginLogFilters, LoginLogsResponse } from '@/types/security-log'

/**
 * Get user's login logs with filters
 * 获取用户的登录日志，支持过滤条件
 */
export function useLoginLogs(filters: LoginLogFilters = {}) {
  return useQuery<LoginLogsResponse>({
    queryKey: [...queryKeys.auth.securityLogs, filters],
    queryFn: async () => {
      const params = new URLSearchParams()
      
      if (filters.action) {
        params.append('action', filters.action)
      }
      if (filters.search) {
        params.append('search', filters.search)
      }
      if (filters.startDate) {
        params.append('start_date', filters.startDate)
      }
      if (filters.endDate) {
        params.append('end_date', filters.endDate)
      }
      if (filters.page) {
        params.append('page', String(filters.page))
      }
      if (filters.pageSize) {
        params.append('page_size', String(filters.pageSize))
      }

      const queryString = params.toString()
      const endpoint = queryString 
        ? `/auth/security/logs?${queryString}` 
        : '/auth/security/logs'

      const response = await authApi.get<LoginLogsResponse>(endpoint)
      return normalizeLoginLogsResponse(response)
    },
    staleTime: 30 * 1000, // 30 seconds
    retry: 1,
  })
}

/**
 * Get single security log by ID
 * 根据ID获取单个安全日志
 */
export function useSecurityLogById(logId: string) {
  return useQuery<AuditLogEntry>({
    queryKey: [...queryKeys.auth.securityLogs, logId],
    queryFn: async () => {
      const response = await authApi.get<AuditLogEntry>(`/auth/security/logs/${logId}`)
      return normalizeAuditLogEntry(response)
    },
    enabled: !!logId,
    staleTime: 60 * 1000, // 1 minute
  })
}

/**
 * Get login log statistics
 * 获取登录日志统计信息
 */
export function useLoginLogStats(startDate?: string, endDate?: string) {
  return useQuery({
    queryKey: [...queryKeys.auth.securityLogs, 'stats', startDate, endDate],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (startDate) params.append('start_date', startDate)
      if (endDate) params.append('end_date', endDate)

      const queryString = params.toString()
      const endpoint = queryString 
        ? `/auth/security/logs/stats?${queryString}` 
        : '/auth/security/logs/stats'

      return authApi.get(endpoint)
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 1,
  })
}

/**
 * Hook to get recent login activity (last 10 logs)
 * 获取最近的登录活动（最近10条日志）
 */
export function useRecentLoginActivity() {
  return useQuery<AuditLogEntry[]>({
    queryKey: [...queryKeys.auth.securityLogs, 'recent'],
    queryFn: async () => {
      const response = await authApi.get<LoginLogsResponse>('/auth/security/logs?page=1&page_size=10')
      return normalizeLoginLogsResponse(response).logs
    },
    staleTime: 1 * 60 * 1000, // 1 minute
    retry: 1,
  })
}
