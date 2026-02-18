'use client'

import { useCallback, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/store/auth-store'
import { tokenStorage } from '@/lib/services/token-storage'
import { logger } from '@/lib/services/logger'
import { getUserMessage, ErrorHandler } from '@/lib/services/error-handler'

interface LoginForm {
  identifier: string
  password: string
}

export function useLoginLogic() {
  const router = useRouter()
  const { setAuth, setLoading, setError, clearError, isLoading } = useAuthStore()
  const [error, setLocalError] = useState<string | null>(null)

  const handleSubmit = useCallback(async (data: LoginForm) => {
    clearError()
    setLoading(true)
    setLocalError(null)

    try {
      logger.info('Login: 开始登录流程', { identifier: data.identifier })

      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      })

      if (!res.ok) {
        const errorData = await res.json()
        throw new Error(errorData.message || '登录失败')
      }

      const userResponse = await res.json()
      const responseData = userResponse.data

      // 检查 token 是否有效（非空字符串）
      if (!responseData?.token || typeof responseData.token !== 'string' || responseData.token.trim() === '') {
        logger.error('Login: 服务器返回的 token 无效', { responseData })
        throw new Error('登录失败：服务器返回的认证令牌无效')
      }

      // 存储 Token
      const expiresAt = responseData.expires_at 
        ? new Date(responseData.expires_at).getTime() 
        : undefined
      tokenStorage.setToken(responseData.token, expiresAt)

      // 构建用户对象
      const displayName = responseData.nickname && responseData.nickname !== responseData.username
        ? responseData.nickname
        : responseData.username

      const user = {
        id: responseData.account_id,
        email: data.identifier.includes('@') ? data.identifier : '',
        username: responseData.username,
        name: displayName,
        roles: [],
        avatar: responseData.avatar,
        profile: {
          phone: data.identifier.includes('@') ? '' : data.identifier,
        },
      }

      // 设置认证状态
      setAuth(user, true, expiresAt)
      
      logger.info('Login: 登录成功', { username: user.username })
      
      // 使用 setTimeout 避免在渲染过程中更新路由
      setTimeout(() => {
        router.push('/home')
      }, 0)

    } catch (error) {
      let message: string
      try {
        message = ErrorHandler.getUserMessage(error)
      } catch {
        message = error instanceof Error ? error.message : '登录失败，请重试'
      }
      setLocalError(message)
      setError(message)
      logger.error('Login: 登录失败', { error })
    } finally {
      setLoading(false)
    }
  }, [router, setAuth, setLoading, setError, clearError])

  const handleClearError = useCallback(() => {
    clearError()
    setLocalError(null)
  }, [clearError])

  return {
    isLoading,
    error,
    onSubmit: handleSubmit,
    onClearError: handleClearError,
  }
}
