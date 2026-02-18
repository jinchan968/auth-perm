'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi, User } from '@/lib/api/auth'
import { queryKeys } from '@/lib/query-keys'
import { ErrorClassifier } from '@/lib/auth/error-classifier'

export function useProfile() {
  return useQuery({
    queryKey: queryKeys.auth.profile,
    queryFn: () => authApi.getProfile(),
    retry: 1,
    staleTime: 5 * 60 * 1000, // 5分钟内不重新请求
  })
}

export function useUpdateProfile() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: Partial<User>) => authApi.updateProfile(data),
    onSuccess: (updatedUser) => {
      queryClient.setQueryData(queryKeys.auth.profile, updatedUser)
    },
    onError: (error) => {
      console.error('更新用户资料失败:', ErrorClassifier.getErrorMessage(error))
    },
  })
}
