'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi, SessionInfo } from '@/lib/api/auth'
import { queryKeys } from '@/lib/query-keys'
import { ErrorClassifier } from '@/lib/auth/error-classifier'

export function useSessions() {
  return useQuery({
    queryKey: queryKeys.auth.sessions,
    queryFn: () => authApi.getSessions(),
    retry: 1,
  })
}

export function useRevokeSession() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (sessionId: string) => authApi.revokeSession(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.sessions })
    },
    onError: (error) => {
      console.error('撤销会话失败:', ErrorClassifier.getErrorMessage(error))
    },
  })
}

export function useRevokeAllSessions() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => authApi.revokeAllSessions(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.sessions })
    },
    onError: (error) => {
      console.error('撤销所有会话失败:', ErrorClassifier.getErrorMessage(error))
    },
  })
}
