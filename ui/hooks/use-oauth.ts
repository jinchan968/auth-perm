'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { authApi, parseOAuthError, isNetworkError } from '@/lib/api/auth'
import { useAuthStore } from '@/store/auth-store'
import { ApiError } from '@/lib/api/client'

export function useOAuth() {
  const router = useRouter()
  const setAuth = useAuthStore((state) => state.setAuth)
  const setError = useAuthStore((state) => state.setError)
  const clearError = useAuthStore((state) => state.clearError)
  const [loading, setLoading] = useState(false)

  const loginWithOAuth = async (provider: string) => {
    setLoading(true)
    clearError()

    try {
      // 获取 OAuth URL
      const { url } = await authApi.getOAuthUrl(provider)

      // 打开 OAuth 授权页面（仅在客户端）
      if (typeof window !== 'undefined') {
        window.location.href = url
      }
    } catch (err) {
      console.error('OAuth login error:', err)

      let errorMessage = 'OAuth 登录失败'

      if (err instanceof ApiError) {
        if (isNetworkError(err)) {
          errorMessage = '网络连接失败，请检查网络后重试'
        } else {
          errorMessage = err.message
        }
      } else if (err instanceof Error) {
        errorMessage = err.message
      }

      setError(errorMessage)
      setLoading(false)
      throw err
    }
  }

  const handleCallback = async (provider: string, code: string, state?: string) => {
    setLoading(true)
    clearError()

    try {
      // 处理 OAuth 回调
      const response = await authApi.handleOAuthCallback(provider, code, state)

      // 设置认证状态
      setAuth(response.user, !!response.token)

      // 如果是新用户，可能需要完善信息
      if (response.isNewUser) {
        router.push('/profile/complete')
      } else {
        router.push('/dashboard')
      }
    } catch (err) {
      console.error('OAuth callback error:', err)

      let errorMessage = 'OAuth 授权失败'

      if (err instanceof ApiError) {
        if (isNetworkError(err)) {
          errorMessage = '网络连接失败，请重试'
        } else {
          errorMessage = err.message
        }
      } else if (err instanceof Error) {
        errorMessage = err.message
      }

      setError(errorMessage)
      setLoading(false)
      throw err
    }
  }

  return {
    loading,
    error: null,
    loginWithOAuth,
    handleCallback,
    clearError,
  }
}

// OAuth 回调页面 Hook
export function useOAuthCallback() {
  const router = useRouter()
  const setAuth = useAuthStore((state) => state.setAuth)
  const setError = useAuthStore((state) => state.setError)
  const clearError = useAuthStore((state) => state.clearError)
  const [loading, setLoading] = useState(true)

  const processCallback = async (provider: string, searchParams: URLSearchParams) => {
    setLoading(true)
    clearError()

    try {
      // 获取 URL 参数
      const code = searchParams.get('code')
      const state = searchParams.get('state')
      const error = searchParams.get('error')

      // 检查是否有错误
      if (error) {
        throw new Error(parseOAuthError(error))
      }

      // 检查是否有授权码
      if (!code) {
        throw new Error('未获取到授权码')
      }

      // 处理回调
      const response = await authApi.handleOAuthCallback(provider, code, state || undefined)

      // 设置认证状态
      setAuth(response.user, !!response.token)

      // 跳转到相应页面
      if (response.isNewUser) {
        router.push('/profile/complete')
      } else {
        router.push('/dashboard')
      }
    } catch (err) {
      console.error('OAuth callback error:', err)

      let errorMessage = 'OAuth 授权失败'

      if (err instanceof Error) {
        errorMessage = err.message
      }

      setError(errorMessage)
      setLoading(false)
      throw err
    }
  }

  return {
    loading,
    error: null,
    processCallback,
  }
}
