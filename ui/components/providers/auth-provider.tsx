
'use client'

import React, { useEffect, useRef, useCallback } from 'react'
import { useAuthStore, hydrateAuthStore } from '@/store/auth-store'

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { user, validateAuthStatus, setReady } = useAuthStore()
  const hasCheckedAuth = useRef(false)
  const isChecking = useRef(false)

  const checkAuth = useCallback(async () => {
    if (isChecking.current) return
    isChecking.current = true

    try {
      // 检查localStorage中是否有token（仅用于日志记录）
      if (typeof window !== 'undefined') {
        const storedToken = localStorage.getItem('auth_token')
        console.log('AuthProvider: Checking localStorage token:', storedToken ? 'found' : 'not found')
      }

      // 如果没有用户信息，无法验证，直接返回
      if (!user) {
        console.log('AuthProvider: No user found, skipping auth check')
        return
      }

      // 使用新的智能认证检查
      const isValid = await validateAuthStatus()

      if (!isValid && user) {
        // 认证失败，清除本地状态
        console.log('AuthProvider: Auth validation failed, clearing auth state')
        useAuthStore.persist.clearStorage()
      }
    } catch (error) {
      console.error('Auth check error:', error)
      // 错误情况下不清除状态，等待下次检查
    } finally {
      isChecking.current = false
      hasCheckedAuth.current = true
    }
  }, [validateAuthStatus, user])

  useEffect(() => {
    // 第一步：水合本地存储数据
    hydrateAuthStore()

    // 第二步：延迟标记为就绪，让水合完成
    const timer = setTimeout(() => {
      setReady(true)
    }, 100)

    return () => clearTimeout(timer)
  }, [setReady])

  useEffect(() => {
    // 第三步：执行认证检查
    if (!hasCheckedAuth.current) {
      checkAuth()
    }
  }, [checkAuth])

  return <>{children}</>
}
