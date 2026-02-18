'use client'

import { useState, useEffect } from 'react'
import { Save, Check } from 'lucide-react'
import { listRoles, assignRoleToAccount, getUserRoles } from '@/lib/api/role'
import { RoleListItem } from '@/types/permission'
import { User } from '@/lib/api/auth'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'

// 用户列表（临时模拟数据，需要后端支持）
interface UserWithRoles {
  id: string
  username: string
  email: string
  role_ids: string[]
}

export default function UserRolesPage() {
  const { user: currentUser } = useAuthStore()

  const [roles, setRoles] = useState<RoleListItem[]>([])
  const [users, setUsers] = useState<UserWithRoles[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null)

  const tenantId = currentUser?.tenant_id || ''

  // 获取角色列表
  useEffect(() => {
    const fetchRoles = async () => {
      try {
        const data = await listRoles({ tenant_id: tenantId, size: 1000 })
        setRoles(data.data || [])
      } catch (err) {
        console.error('Failed to fetch roles:', err)
      }
    }
    if (tenantId) {
      fetchRoles()
    }
  }, [tenantId])

  // 模拟用户列表（需要后端 API 支持）
  useEffect(() => {
    const fetchUsers = async () => {
      setLoading(true)
      try {
        // TODO: 需要后端提供租户用户列表 API
        // 临时使用当前用户作为示例
        if (currentUser) {
          // 获取当前用户的角色
          let userRoles: string[] = []
          try {
            const rolesData = await getUserRoles(currentUser.id)
            userRoles = rolesData.map(r => r.id)
          } catch (e) {
            // 忽略错误
          }

          setUsers([{
            id: currentUser.id,
            username: currentUser.username || currentUser.name || '当前用户',
            email: currentUser.email || '',
            role_ids: userRoles,
          }])
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch users')
      } finally {
        setLoading(false)
      }
    }
    if (currentUser) {
      fetchUsers()
    }
  }, [currentUser])

  const handleToggleRole = (userId: string, roleId: string) => {
    setUsers(prev => prev.map(user => {
      if (user.id === userId) {
        const hasRole = user.role_ids.includes(roleId)
        return {
          ...user,
          role_ids: hasRole
            ? user.role_ids.filter(id => id !== roleId)
            : [...user.role_ids, roleId]
        }
      }
      return user
    }))
  }

  const handleSave = async () => {
    if (!selectedUserId) {
      setError('请选择要分配角色的用户')
      return
    }

    const user = users.find(u => u.id === selectedUserId)
    if (!user) return

    setSaving(true)
    setError('')
    try {
      await assignRoleToAccount({
        account_id: user.id,
        role_ids: user.role_ids,
        tenant_id: tenantId,
      })
      alert('角色分配成功')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to assign roles')
    } finally {
      setSaving(false)
    }
  }

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '权限管理', href: '/permissions' },
    { label: '用户角色分配' },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-slate-50">
      <header className="bg-white/95 backdrop-blur-xl border-b border-slate-200/20 shadow-sm sticky top-0 z-10">
        <div className="px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
              Auth-Perm
            </h1>
            <AvatarDropdown user={currentUser ?? null} />
          </div>
        </div>
      </header>

      <div className="flex">
        <DashboardSidebar pathname="/permissions" />
        <main className="flex-1 p-8">
          <Breadcrumb items={breadcrumbItems} />

          <div className="flex justify-between items-center mb-6 mt-4">
            <h2 className="text-xl font-semibold">用户角色分配</h2>
          </div>

          {error && (
            <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>
          )}

          {loading ? (
            <div className="text-center py-8">加载中...</div>
          ) : (
            <>
              {/* 提示信息 */}
              <Card className="mb-4">
                <CardContent className="pt-6">
                  <div className="text-sm text-yellow-600">
                    注意：当前仅显示当前登录用户。完整的用户角色分配功能需要后端提供租户用户列表 API。
                  </div>
                </CardContent>
              </Card>

              {/* 用户列表 */}
              <Card className="mb-4">
                <CardHeader>
                  <CardTitle>选择用户</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    {users.map(user => (
                      <div
                        key={user.id}
                        className={`p-3 border rounded cursor-pointer hover:bg-gray-50 ${
                          selectedUserId === user.id ? 'border-blue-500 bg-blue-50' : ''
                        }`}
                        onClick={() => setSelectedUserId(user.id)}
                      >
                        <div className="font-medium">{user.username}</div>
                        <div className="text-sm text-gray-500">{user.email}</div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>

              {/* 角色列表 */}
              <Card className="mb-4">
                <CardHeader>
                  <CardTitle>可分配角色</CardTitle>
                </CardHeader>
                <CardContent>
                  {roles.length === 0 ? (
                    <div className="text-gray-500 text-center py-4">暂无可分配的角色</div>
                  ) : (
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
                      {roles.map(role => {
                        const isAssigned = selectedUserId && users.find(u => u.id === selectedUserId)?.role_ids.includes(role.id)
                        return (
                          <div
                            key={role.id}
                            className={`flex items-center gap-2 p-3 border rounded cursor-pointer hover:bg-gray-50 ${
                              isAssigned ? 'border-blue-500 bg-blue-50' : ''
                            }`}
                            onClick={() => selectedUserId && handleToggleRole(selectedUserId, role.id)}
                          >
                            <div className={`w-4 h-4 border rounded flex items-center justify-center ${
                              isAssigned ? 'bg-blue-600 border-blue-600' : 'border-gray-300'
                            }`}>
                              {isAssigned && <Check className="w-3 h-3 text-white" />}
                            </div>
                            <div>
                              <div className="text-sm font-medium">{role.name}</div>
                              <div className="text-xs text-gray-400">{role.code}</div>
                            </div>
                          </div>
                        )
                      })}
                    </div>
                  )}
                </CardContent>
              </Card>

              {/* 保存按钮 */}
              <Button onClick={handleSave} disabled={saving || !selectedUserId}>
                <Save className="h-4 w-4 mr-1" />
                {saving ? '保存中...' : '保存'}
              </Button>
            </>
          )}
        </main>
      </div>
    </div>
  )
}
