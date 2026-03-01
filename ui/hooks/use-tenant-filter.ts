import { useState, useEffect } from 'react'
import { listTenants } from '@/lib/api/tenant'
import { TenantListItem } from '@/types/tenant'

/**
 * useTenantFilter
 *
 * 封装"显示全部租户 / 仅活跃租户"的过滤逻辑，避免在多个页面重复声明
 * 相同的 state + useEffect 组合。
 *
 * @param fallback - 请求失败时的兜底列表（通常来自 useTenant() 的 tenants）
 * @returns
 *   - filteredTenants  当前过滤后的租户列表
 *   - showAllTenants   是否显示全部租户
 *   - setShowAllTenants 切换显示模式
 *   - tenantListLoading 租户列表加载状态
 */
export function useTenantFilter(fallback: TenantListItem[]) {
  const [showAllTenants, setShowAllTenants] = useState(false)
  const [filteredTenants, setFilteredTenants] = useState<TenantListItem[]>(fallback)
  const [tenantListLoading, setTenantListLoading] = useState(false)

  // 当 showAllTenants 切换时重新拉取租户列表
  useEffect(() => {
    let cancelled = false

    const fetch = async () => {
      setTenantListLoading(true)
      try {
        const status = showAllTenants ? undefined : 'active'
        const data = await listTenants({ page: 1, size: 100, status })
        if (!cancelled) setFilteredTenants(data.data || [])
      } catch (err) {
        console.error('Failed to fetch tenants:', err)
        if (!cancelled) setFilteredTenants(fallback)
      } finally {
        if (!cancelled) setTenantListLoading(false)
      }
    }

    fetch().then(() => null)
    return () => { cancelled = true }
  }, [showAllTenants]) // eslint-disable-line react-hooks/exhaustive-deps

  // 当上下文 tenants（兜底列表）加载完成后同步一次（仅在未切换到"全部"时）
  useEffect(() => {
    if (!showAllTenants && !tenantListLoading && fallback.length > 0) {
      setFilteredTenants(fallback)
    }
  }, [fallback]) // eslint-disable-line react-hooks/exhaustive-deps

  return { filteredTenants, showAllTenants, setShowAllTenants, tenantListLoading }
}
