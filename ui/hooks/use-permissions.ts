'use client'

import { create } from 'zustand'
import { useEffect } from 'react'
import { getMyResources, ResourceItem } from '@/lib/api/resource'

interface PermissionsState {
  resources: ResourceItem[]
  menus: Set<string>
  buttons: Set<string>
  apiPaths: Set<string>
  loading: boolean
  loaded: boolean
  error: string | null

  fetchResources: () => Promise<void>
  hasMenu: (menuId: string) => boolean
  hasButton: (buttonId: string) => boolean
  hasAnyMenu: (menuIds: string[]) => boolean
  hasAnyButton: (buttonIds: string[]) => boolean
  clear: () => void
}

export const usePermissionsStore = create<PermissionsState>((set, get) => ({
  resources: [],
  menus: new Set(),
  buttons: new Set(),
  apiPaths: new Set(),
  loading: false,
  loaded: false,
  error: null,

  fetchResources: async () => {
    const { loaded, loading } = get()
    if (loaded || loading) return

    set({ loading: true, error: null })

    try {
      const resources = await getMyResources()

      const menus = new Set<string>()
      const buttons = new Set<string>()
      const apiPaths = new Set<string>()

      resources.forEach((res) => {
        if (res.resource_type === 'menu') {
          menus.add(res.resource_id)
        } else if (res.resource_type === 'button') {
          buttons.add(res.resource_id)
        } else if (res.resource_type === 'api_path') {
          apiPaths.add(res.resource_id)
        }
      })

      set({
        resources,
        menus,
        buttons,
        apiPaths,
        loading: false,
        loaded: true,
      })
    } catch (err) {
      // 403 或其他错误：视为无任何权限
      console.error('Failed to fetch permissions:', err)
      set({
        resources: [],
        menus: new Set(),
        buttons: new Set(),
        apiPaths: new Set(),
        loading: false,
        loaded: true,
        error: err instanceof Error ? err.message : '获取权限失败',
      })
    }
  },

  hasMenu: (menuId: string) => {
    return get().menus.has(menuId)
  },

  hasButton: (buttonId: string) => {
    return get().buttons.has(buttonId)
  },

  hasAnyMenu: (menuIds: string[]) => {
    const { menus } = get()
    return menuIds.some((id) => menus.has(id))
  },

  hasAnyButton: (buttonIds: string[]) => {
    const { buttons } = get()
    return buttonIds.some((id) => buttons.has(id))
  },

  clear: () => {
    set({
      resources: [],
      menus: new Set(),
      buttons: new Set(),
      apiPaths: new Set(),
      loading: false,
      loaded: false,
      error: null,
    })
  },
}))

/**
 * 权限 Hook
 * 用于前端权限控制：菜单、按钮显隐判断
 */
export function usePermissions() {
  const {
    loading,
    loaded,
    error,
    fetchResources,
    hasMenu,
    hasButton,
    hasAnyMenu,
    hasAnyButton,
  } = usePermissionsStore()

  // 首次调用时自动加载权限
  useEffect(() => {
    if (!loaded && !loading) {
      fetchResources()
    }
  }, [loaded, loading, fetchResources])

  return {
    loading,
    loaded,
    error,
    hasMenu,
    hasButton,
    hasAnyMenu,
    hasAnyButton,
    refetch: fetchResources,
  }
}
