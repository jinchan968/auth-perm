
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
      if (!user) return

      const isValid = await validateAuthStatus()

      if (!isValid && user) {
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
