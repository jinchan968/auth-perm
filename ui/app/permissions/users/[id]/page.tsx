'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { ArrowLeft, Save, Check } from 'lucide-react'
import { getUser } from '@/lib/api/user'
import { listRoles, assignRoleToAccount, getUserRoles } from '@/lib/api/role'
import { AccountListItem } from '@/types/user'
import { RoleListItem } from '@/types/permission'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { useTenant } from '@/lib/tenant-context'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'

export default function UserDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { user: currentUser } = useAuthStore()
  const { selectedTenantId } = useTenant()

  const accountId = params.id as string

  const [userDetail, setUserDetail] = useState<AccountListItem | null>(null)
  const [roles, setRoles] = useState<RoleListItem[]>([])
  const [assignedRoleIds, setAssignedRoleIds] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    // Guard: redirect if id is invalid
    if (accountId === 'new' || !accountId) {
      router.replace('/permissions?tab=users')
      return
    }

    const fetchData = async () => {
      if (!selectedTenantId || !accountId) return
      setLoading(true)
      setError('')
      try {
        // 获取用户详情
        const user = await getUser(accountId, selectedTenantId)
        setUserDetail(user)

        // 获取可用角色列表
        const rolesData = await listRoles({
          tenant_id: selectedTenantId,
          size: 1000,
        })
        setRoles(rolesData.data || [])

        // 获取已分配的角色
        const userRolesData = await getUserRoles(accountId)
        setAssignedRoleIds(userRolesData.map((r) => r.id))
      } catch (err) {
        setError(err instanceof Error ? err.message : '加载数据失败')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [selectedTenantId, accountId])

  const handleToggleRole = (roleId: string) => {
    setAssignedRoleIds((prev) =>
      prev.includes(roleId)
        ? prev.filter((id) => id !== roleId)
        : [...prev, roleId]
    )
  }

  const handleSave = async () => {
    if (!selectedTenantId) {
      setError('租户ID缺失')
      return
    }

    setSaving(true)
    setError('')
    try {
      await assignRoleToAccount({
        account_id: accountId,
        role_ids: assignedRoleIds,
        tenant_id: selectedTenantId,
      })
      alert('角色分配成功')
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '权限管理', href: '/permissions' },
    { label: '用户列表', href: '/permissions/users' },
    { label: '用户详情' },
  ]

  if (loading) {
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
            <div className="text-center py-8">加载中...</div>
          </main>
        </div>
      </div>
    )
  }

  if (!userDetail) {
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
            <div className="text-center py-8 text-red-600">用户不存在</div>
          </main>
        </div>
      </div>
    )
  }

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

          <div className="flex items-center gap-4 mb-6 mt-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.back()}
            >
              <ArrowLeft className="h-4 w-4 mr-1" />
              返回
            </Button>
            <h2 className="text-xl font-semibold">用户详情</h2>
          </div>

          {error && (
            <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>
          )}

          {/* 用户基本信息 */}
          <Card className="mb-4">
            <CardHeader>
              <CardTitle>基本信息</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-sm text-gray-500">用户名</div>
                  <div className="font-medium">{userDetail.username}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">昵称</div>
                  <div>{userDetail.nickname || '-'}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">邮箱</div>
                  <div>{userDetail.email || '-'}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">手机号</div>
                  <div>{userDetail.phone || '-'}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">账户类型</div>
                  <div>
                    <Badge variant="outline">{userDetail.account_type}</Badge>
                  </div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">账户状态</div>
                  <div>
                    <Badge
                      variant={
                        userDetail.account_status === 'active'
                          ? 'default'
                          : 'secondary'
                      }
                    >
                      {userDetail.account_status === 'active'
                        ? '活跃'
                        : userDetail.account_status === 'inactive'
                        ? '停用'
                        : '暂停'}
                    </Badge>
                  </div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">邮箱验证</div>
                  <div>
                    <Badge variant={userDetail.email_verified ? 'default' : 'outline'}>
                      {userDetail.email_verified ? '已验证' : '未验证'}
                    </Badge>
                  </div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">最后登录</div>
                  <div>
                    {userDetail.last_login_at
                      ? new Date(userDetail.last_login_at).toLocaleString('zh-CN')
                      : '从未登录'}
                  </div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">创建时间</div>
                  <div>
                    {new Date(userDetail.created_at).toLocaleString('zh-CN')}
                  </div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">更新时间</div>
                  <div>
                    {new Date(userDetail.updated_at).toLocaleString('zh-CN')}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* 角色分配 */}
          <Card className="mb-4">
            <CardHeader>
              <CardTitle>角色分配</CardTitle>
            </CardHeader>
            <CardContent>
              {roles.length === 0 ? (
                <div className="text-gray-500 text-center py-4">
                  当前租户暂无可分配的角色
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                  {roles.map((role) => {
                    const isAssigned = assignedRoleIds.includes(role.id)
                    return (
                      <div
                        key={role.id}
                        className={`flex items-center gap-3 p-4 border rounded-lg cursor-pointer transition-all hover:shadow-md ${
                          isAssigned
                            ? 'border-blue-500 bg-blue-50'
                            : 'border-gray-200 hover:border-blue-300'
                        }`}
                        onClick={() => handleToggleRole(role.id)}
                      >
                        <div
                          className={`w-5 h-5 border-2 rounded flex items-center justify-center flex-shrink-0 ${
                            isAssigned
                              ? 'bg-blue-600 border-blue-600'
                              : 'border-gray-300'
                          }`}
                        >
                          {isAssigned && <Check className="w-3 h-3 text-white" />}
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="font-medium truncate">{role.name}</div>
                          <div className="text-xs text-gray-400 truncate">
                            {role.code}
                          </div>
                          {role.description && (
                            <div className="text-xs text-gray-500 mt-1 truncate">
                              {role.description}
                            </div>
                          )}
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </CardContent>
          </Card>

          {/* 保存按钮 */}
          <div className="flex gap-2">
            <Button onClick={handleSave} disabled={saving}>
              <Save className="h-4 w-4 mr-1" />
              {saving ? '保存中...' : '保存角色分配'}
            </Button>
            <Button variant="outline" onClick={() => router.back()}>
              取消
            </Button>
          </div>
        </main>
      </div>
    </div>
  )
}
