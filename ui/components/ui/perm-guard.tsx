'use client'

import { ReactNode } from 'react'
import { usePermissions } from '@/hooks/use-permissions'

interface PermGuardProps {
  /** 菜单权限ID */
  menu?: string
  /** 按钮权限ID */
  button?: string
  /** 无权限时的替代内容，默认 null（不渲染） */
  fallback?: ReactNode
  /** 子元素 */
  children: ReactNode
}

/**
 * 权限守卫组件
 * 根据用户权限控制子组件是否渲染
 *
 * @example
 * // 菜单权限
 * <PermGuard menu="tenants">
 *   <MenuItem>租户管理</MenuItem>
 * </PermGuard>
 *
 * // 按钮权限
 * <PermGuard button="tenants.show_all">
 *   <Checkbox>显示全部租户</Checkbox>
 * </PermGuard>
 */
export function PermGuard({ menu, button, fallback = null, children }: PermGuardProps) {
  const { hasMenu, hasButton, loading, isSuperAdmin } = usePermissions()

  // 加载中：不渲染（避免闪烁）
  if (loading) {
    return null
  }

  // 超管不受权限限制
  if (isSuperAdmin) {
    return <>{children}</>
  }

  // 检查权限
  let hasPermission: boolean

  if (menu) {
    hasPermission = hasMenu(menu)
  } else if (button) {
    hasPermission = hasButton(button)
  } else {
    // 未指定权限类型，默认渲染
    hasPermission = true
  }

  return <>{hasPermission ? children : fallback}</>
}
