'use client'

import { useState, useEffect, Suspense } from 'react'
import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import { Plus } from 'lucide-react'
import { listPermissions } from '@/lib/api/permission'
import { listRoles, createRole } from '@/lib/api/role'
import { listUsers, createUser } from '@/lib/api/user'
import { PermissionListItem, RoleListItem } from '@/types/permission'
import { AccountListItem } from '@/types/user'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { AppModal } from '@/components/ui/app-modal'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useTenant } from '@/lib/tenant-context'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'
import { useTenantFilter } from '@/hooks/use-tenant-filter'

type TabType = 'permissions' | 'roles' | 'users'

function PermissionsPageContent() {
  const searchParams = useSearchParams()
  const { user } = useAuthStore()

  // 使用统一的租户上下文
  const { tenants, selectedTenantId, setSelectedTenantId, tenantId, loading: tenantLoading } = useTenant()

  // Tab state - read from URL query param
  const initialTab = searchParams.get('tab') === 'roles' ? 'roles' : searchParams.get('tab') === 'users' ? 'users' : 'permissions'
  const [activeTab, setActiveTab] = useState<TabType>(initialTab)

  // Permissions state
  const [permissions, setPermissions] = useState<PermissionListItem[]>([])
  const [loading, setLoading] = useState(true)

  // Roles state
  const [roles, setRoles] = useState<RoleListItem[]>([])
  const [rolesLoading, setRolesLoading] = useState(false)

  // Users state
  const [users, setUsers] = useState<AccountListItem[]>([])
  const [usersLoading, setUsersLoading] = useState(false)

  // Common state
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [size] = useState(10)
  const [error, setError] = useState('')

  // 租户过滤（封装了 showAllTenants / filteredTenants / tenantListLoading 的逻辑）
  const { filteredTenants, showAllTenants, setShowAllTenants, tenantListLoading } = useTenantFilter(tenants)

  // Role creation modal state
  const [roleModalOpen, setRoleModalOpen] = useState(false)
  const [roleSaving, setRoleSaving] = useState(false)
  const [roleFormData, setRoleFormData] = useState({
    name: '',
    description: '',
  })

  // User creation modal state
  const [userModalOpen, setUserModalOpen] = useState(false)
  const [userCreating, setUserCreating] = useState(false)
  const [userFormData, setUserFormData] = useState({
    identifier_type: 'email' as 'email' | 'phone',
    email: '',
    phone: '',
    username: '',
    password: '',
    confirm_password: '',
    nickname: '',
  })
  const [userFormError, setUserFormError] = useState('')

  const fetchPermissions = async (pageOverride?: number) => {
    if (!tenantId) return
    const targetPage = pageOverride ?? page
    setLoading(true)
    setError('')
    try {
      const data = await listPermissions({ tenant_id: tenantId, keyword, page: targetPage, size })
      setPermissions(data.data)
      setTotal(data.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch permissions')
    } finally {
      setLoading(false)
    }
  }

  const fetchRoles = async (pageOverride?: number) => {
    if (!tenantId) return
    const targetPage = pageOverride ?? page
    setRolesLoading(true)
    setError('')
    try {
      const data = await listRoles({ tenant_id: tenantId, keyword, page: targetPage, size })
      setRoles(data.data)
      setTotal(data.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch roles')
    } finally {
      setRolesLoading(false)
    }
  }

  const fetchUsers = async (pageOverride?: number) => {
    if (!tenantId) return
    const targetPage = pageOverride ?? page
    setUsersLoading(true)
    setError('')
    try {
      const data = await listUsers({ tenant_id: tenantId, keyword, page: targetPage, page_size: size })
      setUsers(data.data || [])
      setTotal(data.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch users')
    } finally {
      setUsersLoading(false)
    }
  }

  // 确保在 filteredTenants 更新后，如果没有选中租户则自动选中第一个 active 租户
  useEffect(() => {
    if (filteredTenants.length > 0 && !selectedTenantId) {
      const activeTenant = filteredTenants.find((t) => t.status === 'active')
      if (activeTenant) {
        setSelectedTenantId(activeTenant.id)
      } else {
        setSelectedTenantId(filteredTenants[0].id)
      }
    }
  }, [filteredTenants])

  // 统一的数据获取 effect - 合并所有触发条件
  useEffect(() => {
    if (!tenantId) return

    if (activeTab === 'permissions') {
      fetchPermissions()
    } else if (activeTab === 'roles') {
      fetchRoles()
    } else if (activeTab === 'users') {
      fetchUsers()
    }
  }, [activeTab, page, tenantId, keyword])

  // Tab 切换时重置页码
  const handleTabChange = (tab: TabType) => {
    setActiveTab(tab)
    setPage(1)
    setKeyword('')
    setError('')
  }

  // 搜索时重置页码
  const handleSearch = () => {
    setPage(1)
    if (activeTab === 'permissions') {
      fetchPermissions(1)
    } else if (activeTab === 'roles') {
      fetchRoles(1)
    } else {
      fetchUsers(1)
    }
  }

  // Role modal handlers
  const handleOpenRoleModal = () => {
    setRoleFormData({ name: '', description: '' })
    setRoleModalOpen(true)
  }

  const handleCloseRoleModal = () => {
    setRoleModalOpen(false)
    setRoleFormData({ name: '', description: '' })
  }

  const handleSaveRole = async () => {
    if (!roleFormData.name) {
      setError('请填写必填字段')
      return
    }

    setRoleSaving(true)
    setError('')
    try {
      await createRole({
        tenant_id: tenantId,
        name: roleFormData.name,
        description: roleFormData.description,
      })
      handleCloseRoleModal()
      // Refresh roles list
      if (activeTab === 'roles') {
        fetchRoles(page)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save role')
    } finally {
      setRoleSaving(false)
    }
  }

  // User creation handlers
  const handleOpenUserModal = () => {
    setUserFormData({
      identifier_type: 'email',
      email: '',
      phone: '',
      username: '',
      password: '',
      confirm_password: '',
      nickname: '',
    })
    setUserFormError('')
    setUserModalOpen(true)
  }

  const handleCloseUserModal = () => {
    setUserModalOpen(false)
    setUserFormError('')
  }

  const handleCreateUser = async () => {
    if (!selectedTenantId) {
      setUserFormError('请选择租户')
      return
    }
    if (!userFormData.username || !userFormData.password || !userFormData.confirm_password) {
      setUserFormError('请填写必填项')
      return
    }
    if (userFormData.password !== userFormData.confirm_password) {
      setUserFormError('两次密码输入不一致')
      return
    }
    if (userFormData.identifier_type === 'email' && !userFormData.email) {
      setUserFormError('请输入邮箱')
      return
    }
    if (userFormData.identifier_type === 'phone' && !userFormData.phone) {
      setUserFormError('请输入手机号')
      return
    }
    setUserCreating(true)
    setUserFormError('')
    try {
      await createUser({
        ...userFormData,
        tenant_id: selectedTenantId,
      })
      handleCloseUserModal()
      fetchUsers()
    } catch (err) {
      setUserFormError(err instanceof Error ? err.message : '创建用户失败')
    } finally {
      setUserCreating(false)
    }
  }

  const getStatusBadge = (isActive: boolean) => {
    return isActive ? (
      <Badge variant="default">启用</Badge>
    ) : (
      <Badge variant="secondary">禁用</Badge>
    )
  }

  const totalPages = Math.ceil(total / size)

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '权限管理' },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-slate-50">
      <header className="bg-white/95 backdrop-blur-xl border-b border-slate-200/20 shadow-sm sticky top-0 z-10">
        <div className="px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
              Auth-Perm
            </h1>
            <AvatarDropdown user={user ?? null} />
          </div>
        </div>
      </header>

      <div className="flex">
        <DashboardSidebar pathname="/permissions" />
        <main className="flex-1 p-8">
          <Breadcrumb items={breadcrumbItems} />

          {/* Tabs */}
          <div className="flex gap-2 mb-6 mt-4">
            <Button
              variant={activeTab === 'permissions' ? 'default' : 'outline'}
              onClick={() => handleTabChange('permissions')}
            >
              权限列表
            </Button>
            <Button
              variant={activeTab === 'roles' ? 'default' : 'outline'}
              onClick={() => handleTabChange('roles')}
            >
              角色列表
            </Button>
            <Button
              variant={activeTab === 'users' ? 'default' : 'outline'}
              onClick={() => handleTabChange('users')}
            >
              用户列表
            </Button>
          </div>

          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">
              {activeTab === 'permissions' ? '权限列表' : activeTab === 'roles' ? '角色列表' : '用户列表'}
            </h2>
            {activeTab === 'permissions' ? (
              <Link href="/permissions/new">
                <Button>新建权限</Button>
              </Link>
            ) : activeTab === 'roles' ? (
              <Button onClick={handleOpenRoleModal}>新建角色</Button>
            ) : (
              <Button onClick={handleOpenUserModal}>
                <Plus className="h-4 w-4 mr-1" />
                新增用户
              </Button>
            )}
          </div>

          <Card>
            <CardContent className="pt-6">
              {/* Tenant Filter and Search */}
              <div className="flex gap-2 items-center mb-4">
                <label className="flex items-center gap-2 text-sm cursor-pointer">
                  <input
                    type="checkbox"
                    checked={showAllTenants}
                    onChange={(e) => setShowAllTenants(e.target.checked)}
                    className="w-4 h-4 rounded border-gray-300"
                  />
                  显示全部租户
                </label>
                <Select
                  value={selectedTenantId || user?.tenant_id || ''}
                  onValueChange={(value) => {
                    setSelectedTenantId(value)
                    setPage(1)
                  }}
                  disabled={tenantLoading || tenantListLoading}
                >
                  <SelectTrigger className="w-[200px]">
                    <SelectValue placeholder="选择租户" />
                  </SelectTrigger>
                  <SelectContent>
                    {filteredTenants.map((tenant) => (
                      <SelectItem key={tenant.id} value={tenant.id}>
                        {tenant.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Input
                  placeholder={activeTab === 'permissions' ? '搜索权限名称或代码...' : '搜索角色名称或代码...'}
                  value={keyword}
                  onChange={(e) => setKeyword(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                  className="max-w-xs"
                />
                <Button onClick={handleSearch}>搜索</Button>
              </div>

              {/* Error */}
              {error && (
                <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>
              )}

              {/* Permissions Table */}
              {activeTab === 'permissions' && (
                <div className="border rounded">
                  <table className="w-full">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">名称</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">代码</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">描述</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">系统权限</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">状态</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {loading ? (
                        <tr>
                          <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                            加载中...
                          </td>
                        </tr>
                      ) : permissions.length === 0 ? (
                        <tr>
                          <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                            暂无数据
                          </td>
                        </tr>
                      ) : (
                        permissions.map((permission) => (
                          <tr key={permission.id} className="border-t">
                            <td className="px-4 py-2">{permission.name}</td>
                            <td className="px-4 py-2 font-mono text-sm">{permission.code}</td>
                            <td className="px-4 py-2">{permission.resource}</td>
                            <td className="px-4 py-2 text-gray-500 truncate max-w-xs">{permission.description || '-'}</td>
                            <td className="px-4 py-2">
                              {permission.is_system ? (
                                <Badge variant="outline">系统</Badge>
                              ) : (
                                <span className="text-gray-400">-</span>
                              )}
                            </td>
                            <td className="px-4 py-2">{getStatusBadge(permission.is_active)}</td>
                            <td className="px-4 py-2">
                              <Link href={`/permissions/${permission.id}`}>
                                <Button variant="ghost" size="sm">查看</Button>
                              </Link>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Roles Table */}
              {activeTab === 'roles' && (
                <div className="border rounded">
                  <table className="w-full">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">名称</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">代码</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">描述</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">权限数量</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">状态</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {rolesLoading ? (
                        <tr>
                          <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                            加载中...
                          </td>
                        </tr>
                      ) : roles.length === 0 ? (
                        <tr>
                          <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                            暂无数据
                          </td>
                        </tr>
                      ) : (
                        roles.map((role) => (
                          <tr key={role.id} className="border-t">
                            <td className="px-4 py-2">{role.name}</td>
                            <td className="px-4 py-2 font-mono text-sm">{role.code}</td>
                            <td className="px-4 py-2 text-gray-500 truncate max-w-xs">{role.description || '-'}</td>
                            <td className="px-4 py-2">{role.permission_count || 0}</td>
                            <td className="px-4 py-2">{getStatusBadge(role.is_active)}</td>
                            <td className="px-4 py-2">
                              <Link href={`/permissions/roles/${role.id}`}>
                                <Button variant="ghost" size="sm">查看</Button>
                              </Link>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Users Table */}
              {activeTab === 'users' && (
                <div className="border rounded">
                  <table className="w-full">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">用户名</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">昵称</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">邮箱</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">手机号</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">状态</th>
                        <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {usersLoading ? (
                        <tr>
                          <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                            加载中...
                          </td>
                        </tr>
                      ) : users.length === 0 ? (
                        <tr>
                          <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                            暂无数据
                          </td>
                        </tr>
                      ) : (
                        users.map((user) => (
                          <tr key={user.account_id} className="border-t">
                            <td className="px-4 py-2">{user.username || '-'}</td>
                            <td className="px-4 py-2">{user.nickname || '-'}</td>
                            <td className="px-4 py-2">{user.email || '-'}</td>
                            <td className="px-4 py-2">{user.phone || '-'}</td>
                            <td className="px-4 py-2">
                              {user.user_status === 'active' ? (
                                <Badge variant="default">启用</Badge>
                              ) : (
                                <Badge variant="secondary">禁用</Badge>
                              )}
                            </td>
                            <td className="px-4 py-2">
                              <Link href={`/permissions/users/${user.account_id}`}>
                                <Button variant="ghost" size="sm">查看</Button>
                              </Link>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Pagination */}
              {total > 0 && (
                <div className="flex justify-between items-center mt-4">
                  <div className="text-sm text-gray-500">
                    共 {total} 条记录，第 {page}/{totalPages} 页
                  </div>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={page === 1}
                      onClick={() => setPage(page - 1)}
                    >
                      上一页
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={page >= totalPages}
                      onClick={() => setPage(page + 1)}
                    >
                      下一页
                    </Button>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Role Creation Modal */}
          <AppModal open={roleModalOpen} onClose={handleCloseRoleModal} title="新建角色">
            <div className="p-6 space-y-4">
              <div className="space-y-2">
                <Label htmlFor="roleName">角色名称 *</Label>
                <Input
                  id="roleName"
                  placeholder="如: 管理员"
                  value={roleFormData.name}
                  onChange={(e) => setRoleFormData({ ...roleFormData, name: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="roleDesc">描述</Label>
                <Input
                  id="roleDesc"
                  placeholder="可选描述"
                  value={roleFormData.description}
                  onChange={(e) => setRoleFormData({ ...roleFormData, description: e.target.value })}
                />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <Button variant="outline" onClick={handleCloseRoleModal}>取消</Button>
                <Button onClick={handleSaveRole} disabled={roleSaving}>
                  {roleSaving ? '保存中...' : '保存'}
                </Button>
              </div>
            </div>
          </AppModal>

          {/* User Creation Modal */}
          <AppModal open={userModalOpen} onClose={handleCloseUserModal} title="新增用户">
            <div className="p-6 space-y-4">
              <div>
                <Label>标识符类型 *</Label>
                <Select
                  value={userFormData.identifier_type}
                  onValueChange={(value: 'email' | 'phone') =>
                    setUserFormData({ ...userFormData, identifier_type: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="email">邮箱</SelectItem>
                    <SelectItem value="phone">手机号</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {userFormData.identifier_type === 'email' ? (
                <div>
                  <Label>邮箱 *</Label>
                  <Input
                    type="email"
                    value={userFormData.email}
                    onChange={(e) => setUserFormData({ ...userFormData, email: e.target.value })}
                    placeholder="user@example.com"
                  />
                </div>
              ) : (
                <div>
                  <Label>手机号 *</Label>
                  <Input
                    type="tel"
                    value={userFormData.phone}
                    onChange={(e) => setUserFormData({ ...userFormData, phone: e.target.value })}
                    placeholder="13800138000"
                  />
                </div>
              )}

              <div>
                <Label>用户名 *</Label>
                <Input
                  value={userFormData.username}
                  onChange={(e) => setUserFormData({ ...userFormData, username: e.target.value })}
                  placeholder="username"
                />
              </div>

              <div>
                <Label>昵称</Label>
                <Input
                  value={userFormData.nickname}
                  onChange={(e) => setUserFormData({ ...userFormData, nickname: e.target.value })}
                  placeholder="昵称（可选）"
                />
              </div>

              <div>
                <Label>密码 *</Label>
                <Input
                  type="password"
                  value={userFormData.password}
                  onChange={(e) => setUserFormData({ ...userFormData, password: e.target.value })}
                  placeholder="至少6位"
                />
              </div>

              <div>
                <Label>确认密码 *</Label>
                <Input
                  type="password"
                  value={userFormData.confirm_password}
                  onChange={(e) =>
                    setUserFormData({ ...userFormData, confirm_password: e.target.value })
                  }
                  placeholder="再次输入密码"
                />
              </div>

              {userFormError && (
                <div className="bg-red-50 text-red-600 p-3 rounded text-sm">{userFormError}</div>
              )}

              <div className="flex justify-end gap-2 pt-2">
                <Button variant="outline" onClick={handleCloseUserModal}>取消</Button>
                <Button onClick={handleCreateUser} disabled={userCreating}>
                  {userCreating ? '创建中...' : '创建'}
                </Button>
              </div>
            </div>
          </AppModal>
        </main>
      </div>
    </div>
  )
}

// Wrapper component for Suspense boundary with useSearchParams
export default function PermissionsPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-slate-50">
        <header className="bg-white/95 backdrop-blur-xl border-b border-slate-200/20 shadow-sm sticky top-0 z-10">
          <div className="px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between items-center h-16">
              <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
                Auth-Perm
              </h1>
            </div>
          </div>
        </header>
        <div className="flex">
          <DashboardSidebar pathname="/permissions" />
          <main className="flex-1 p-8">
            <div className="text-center">加载中...</div>
          </main>
        </div>
      </div>
    }>
      <PermissionsPageContent />
    </Suspense>
  )
}
