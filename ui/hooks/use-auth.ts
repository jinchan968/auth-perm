'use client'

import { useAuthStore } from '@/store/auth-store'
import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

export function useAuth() {
  const { user, isAuthenticated, isLoading, isReady, setAuth, setReady } = useAuthStore()
  const router = useRouter()

  useEffect(() => {
    if (isReady && !isLoading && !isAuthenticated) {
      router.push('/login')
    }
  }, [isReady, isLoading, isAuthenticated, router])

  return {
    user,
    isAuthenticated,
    isLoading,
    isReady,
    setAuth,
    setReady,
  }
}
