'use client'

import {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from 'react'
import { listTenants } from '@/lib/api/tenant'
import { TenantListItem } from '@/types/tenant'
import { useAuthStore } from '@/store/auth-store'

// 默认租户 ID（与后端 constant.DefaultTenantID = "default" 保持一致）
export const DEFAULT_TENANT_ID = 'default'

interface TenantContextType {
  tenants: TenantListItem[]
  selectedTenantId: string
  setSelectedTenantId: (id: string) => void
  tenantId: string
  loading: boolean
}

const TenantContext = createContext<TenantContextType | undefined>(undefined)

// 模块级别缓存，防止 React Strict Mode 双重请求
let tenantsCache: TenantListItem[] | null = null
let fetchPromise: Promise<TenantListItem[]> | null = null

export function TenantProvider({ children }: { children: ReactNode }) {
  const { user } = useAuthStore()
  const [tenants, setTenants] = useState<TenantListItem[]>(tenantsCache || [])
  const [selectedTenantId, setSelectedTenantId] = useState<string>('')
  const [loading, setLoading] = useState(!tenantsCache)

  // 获取租户列表
  useEffect(() => {
    let isCancelled = false

    const fetchTenants = async () => {
      // 如果已有缓存，直接使用
      if (tenantsCache) {
        setTenants(tenantsCache)
        setLoading(false)
        return
      }

      // 如果已有进行中的请求，复用它（请求去重）
      if (!fetchPromise) {
        // 默认只获取 active 状态的租户
        fetchPromise = listTenants({ page: 1, size: 100, status: 'active' })
          .then((data) => {
            const result = data.data || []
            // 请求成功后立即设置缓存，防止后续请求
            tenantsCache = result
            return result
          })
          .catch((err) => {
            // 请求失败时清空 promise，允许重试
            fetchPromise = null
            throw err
          })
      }

      setLoading(true)
      try {
        const data = await fetchPromise
        if (!isCancelled) {
          setTenants(data)
        }
      } catch (err) {
        if (!isCancelled) {
          console.error('Failed to fetch tenants:', err)
        }
      } finally {
        if (!isCancelled) {
          setLoading(false)
        }
      }
    }
    fetchTenants()

    return () => {
      isCancelled = true
    }
  }, [])

  // 设置默认租户
  useEffect(() => {
    if (tenants.length > 0 && !selectedTenantId) {
      // 优先使用用户所属租户（如果是 active 状态）
      if (user?.tenant_id) {
        const userTenant = tenants.find((t) => t.id === user.tenant_id)
        if (userTenant && userTenant.status === 'active') {
          setSelectedTenantId(user.tenant_id)
          return
        }
      }
      // 否则使用第一个 active 状态的租户
      const activeTenant = tenants.find((t) => t.status === 'active')
      if (activeTenant) {
        setSelectedTenantId(activeTenant.id)
      } else if (tenants.length > 0) {
        // 如果没有 active 租户，使用第一个
        setSelectedTenantId(tenants[0].id)
      }
    }
  }, [tenants, user, selectedTenantId])

  // 计算实际使用的 tenantId
  const tenantId = selectedTenantId || user?.tenant_id || DEFAULT_TENANT_ID

  return (
    <TenantContext.Provider
      value={{
        tenants,
        selectedTenantId,
        setSelectedTenantId,
        tenantId,
        loading,
      }}
    >
      {children}
    </TenantContext.Provider>
  )
}

export function useTenant() {
  const context = useContext(TenantContext)
  if (context === undefined) {
    throw new Error('useTenant must be used within a TenantProvider')
  }
  return context
}
