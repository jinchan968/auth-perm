'use client'

import { useCallback, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/store/auth-store'
import { tokenStorage } from '@/lib/services/token-storage'
import { logger } from '@/lib/services/logger'
import { ErrorHandler } from '@/lib/services/error-handler'

interface RegisterForm {
  identifier_type: 'email' | 'phone'
  email?: string
  phone?: string
  username: string
  password: string
  confirm_password: string
}

interface RegisterResponse {
  token: string
  expires_at: string
  account_id: string
  username: string
  nickname?: string
  avatar?: string
}

export function useRegisterLogic() {
  const router = useRouter()
  const { setAuth } = useAuthStore()
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = useCallback(async (data: RegisterForm) => {
    setError(null)

    try {
      logger.info('Register: 开始注册流程', { identifier_type: data.identifier_type })

      const registerData = {
        identifier_type: data.identifier_type,
        email: data.email,
        phone: data.phone,
        username: data.username,
        password: data.password,
        confirm_password: data.confirm_password,
      }

      const response = await fetch('/api/auth/public/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(registerData),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || '注册失败')
      }

      const responseData = await response.json()
      const result = responseData.data as RegisterResponse

      // 检查 token 是否有效（非空字符串）
      if (!result?.token || typeof result.token !== 'string' || result.token.trim() === '') {
        logger.error('Register: 服务器返回的 token 无效', { result })
        throw new Error('注册失败：服务器返回的认证令牌无效')
      }

      if (result.token) {
        tokenStorage.setToken(result.token, result.expires_at ? new Date(result.expires_at).getTime() : undefined)

        const identifier = data.identifier_type === 'email' ? data.email : data.phone
        const displayName = result.nickname && result.nickname !== result.username
          ? result.nickname
          : result.username

        const user = {
          id: result.account_id,
          email: data.identifier_type === 'email' ? identifier || '' : '',
          username: result.username,
          name: displayName,
          roles: [],
          avatar: result.avatar,
          profile: {
            phone: data.identifier_type === 'phone' ? identifier || '' : '',
          },
        }

        setAuth(user, true, result.expires_at ? new Date(result.expires_at).getTime() : undefined)
        
        logger.info('Register: 注册成功', { username: user.username })
        
        setTimeout(() => {
          router.push('/home')
        }, 0)
      } else {
        router.push('/login')
      }
    } catch (error) {
      let message: string
      try {
        message = ErrorHandler.getUserMessage(error)
      } catch {
        message = error instanceof Error ? error.message : '注册失败，请重试'
      }
      setError(message)
      logger.error('Register: 注册失败', { error })
    }
  }, [router, setAuth])

  const handleClearError = useCallback(() => {
    setError(null)
  }, [])

  return {
    error,
    onSubmit: handleSubmit,
    onClearError: handleClearError,
  }
}
