'use client'

import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import { User } from '@/lib/api/auth'
import { tokenStorage } from '@/lib/services/token-storage'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  isReady: boolean
  error: string | null

  setAuth: (user: User | null, isAuthenticated: boolean, expiresAt?: number) => void
  setUser: (user: User) => void
  logout: () => Promise<void>
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
  setReady: (ready: boolean) => void
  validateAuthStatus: () => Promise<boolean>
}

const STORAGE_KEY = 'auth-storage'

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      isReady: false,
      error: null,

      setAuth: (user, isAuthenticated, expiresAt) => {
        if (isAuthenticated && expiresAt) {
          tokenStorage.setTokenExpiry(expiresAt)
        }

        set({
          user,
          isAuthenticated,
          isLoading: false,
          isReady: true,
          error: null,
        })
      },

      setUser: (user) => {
        set({ user })
      },

      logout: async () => {
        try {
          await fetch('/api/auth/logout', { method: 'POST' })
        } catch (error) {
          console.warn('Logout 请求失败:', error)
        } finally {
          tokenStorage.clearAll()
          useAuthStore.persist.clearStorage()
          
          set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
            isReady: false,
            error: null,
          })
        }
      },

      setLoading: (isLoading) => set({ isLoading }),

      setError: (error) => set({ error, isLoading: false }),

      clearError: () => set({ error: null }),

      setReady: (ready) => set({ isReady: ready }),

      validateAuthStatus: async (): Promise<boolean> => {
        const { user } = get()
        if (!user) return false

        const { isValid } = tokenStorage.getAuthInfo()
        
        if (isValid) {
          set({ isAuthenticated: true })
          return true
        }

        // Token 无效，清除状态（不在 store 内部调用 logout）
        tokenStorage.clearAll()
        useAuthStore.persist.clearStorage()
        
        set({
          user: null,
          isAuthenticated: false,
          isLoading: false,
          isReady: false,
          error: null,
        })
        
        return false
      },
    }),
    {
      name: STORAGE_KEY,
      storage: createJSONStorage(() => localStorage),
      skipHydration: true,
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.setReady(true)
        }
      },
    }
  )
)

export const hydrateAuthStore = () => {
  useAuthStore.persist.rehydrate()
}
