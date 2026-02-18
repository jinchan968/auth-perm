'use client'

import { useState, useEffect, Suspense } from 'react'
import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import { listPermissions } from '@/lib/api/permission'
import { listRoles, createRole } from '@/lib/api/role'
import { PermissionListItem, RoleListItem } from '@/types/permission'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { listTenants } from '@/lib/api/tenant'
import { TenantListItem } from '@/types/tenant'
import { useTenant } from '@/lib/tenant-context'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'

type TabType = 'permissions' | 'roles'

function PermissionsPageContent() {
  const searchParams = useSearchParams()
  const { user } = useAuthStore()

  // 使用统一的租户上下文
  const { tenants, selectedTenantId, setSelectedTenantId, tenantId, loading: tenantLoading } = useTenant()

  // Tab state - read from URL query param
  const initialTab = searchParams.get('tab') === 'roles' ? 'roles' : 'permissions'
  const [activeTab, setActiveTab] = useState<TabType>(initialTab)

  // Permissions state
  const [permissions, setPermissions] = useState<PermissionListItem[]>([])
  const [loading, setLoading] = useState(true)

  // Roles state
  const [roles, setRoles] = useState<RoleListItem[]>([])
  const [rolesLoading, setRolesLoading] = useState(false)

  // Common state
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [size] = useState(10)
  const [error, setError] = useState('')

  // Tenant keyword for filtering (local state)
  const [tenantKeyword, setTenantKeyword] = useState('')

  // 租户过滤状态
  const [showAllTenants, setShowAllTenants] = useState(false)
  const [filteredTenants, setFilteredTenants] = useState<TenantListItem[]>([])
  const [tenantListLoading, setTenantListLoading] = useState(false)

  // Role creation modal state
  const [roleModalOpen, setRoleModalOpen] = useState(false)
  const [roleSaving, setRoleSaving] = useState(false)
  const [roleFormData, setRoleFormData] = useState({
    name: '',
    description: '',
  })

  const fetchPermissions = async () => {
    if (!tenantId) return
    setLoading(true)
    setError('')
    try {
      const data = await listPermissions({ tenant_id: tenantId, keyword, page, size })
      setPermissions(data.data)
      setTotal(data.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch permissions')
    } finally {
      setLoading(false)
    }
  }

  const fetchRoles = async () => {
    if (!tenantId) return
    setRolesLoading(true)
    setError('')
    try {
      const data = await listRoles({ tenant_id: tenantId, keyword, page, size })
      setRoles(data.data)
      setTotal(data.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch roles')
    } finally {
      setRolesLoading(false)
    }
  }

  // 租户过滤 - 当 showAllTenants 变化时重新获取租户列表
  useEffect(() => {
    const fetchFilteredTenants = async () => {
      setTenantListLoading(true)
      try {
        // 不传 status 时后端默认返回 active，传 status=all 时返回全部
        const status = showAllTenants ? undefined : 'active'
        const data = await listTenants({ page: 1, size: 100, status })
        setFilteredTenants(data.data || [])
      } catch (err) {
        console.error('Failed to fetch tenants:', err)
        // 失败时使用缓存的租户列表
        setFilteredTenants(tenants)
      } finally {
        setTenantListLoading(false)
      }
    }
    fetchFilteredTenants()
  }, [showAllTenants])

  // 初始化 filteredTenants 当 tenants 从上下文加载完成时
  useEffect(() => {
    if (tenants.length > 0 && filteredTenants.length === 0 && !showAllTenants) {
      setFilteredTenants(tenants)
    }
  }, [tenants])

  // 确保在 filteredTenants 更新后，如果没有选中租户则自动选中第一个 active 租户
  useEffect(() => {
    if (filteredTenants.length > 0 && !selectedTenantId) {
      const activeTenant = filteredTenants.find((t) => t.status === 'active')
      if (activeTenant) {
        setSelectedTenantId(activeTenant.id)
      } else if (filteredTenants.length > 0) {
        setSelectedTenantId(filteredTenants[0].id)
      }
    }
  }, [filteredTenants])

  // 统一的数据获取 effect - 合并所有触发条件
  useEffect(() => {
    if (!tenantId) return

    if (activeTab === 'permissions') {
      fetchPermissions()
    } else {
      fetchRoles()
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
        fetchRoles()
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save role')
    } finally {
      setRoleSaving(false)
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
          </div>

          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">
              {activeTab === 'permissions' ? '权限列表' : '角色列表'}
            </h2>
            {activeTab === 'permissions' ? (
              <Link href="/permissions/new">
                <Button>新建权限</Button>
              </Link>
            ) : (
              <Button onClick={handleOpenRoleModal}>新建角色</Button>
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
                          <td colSpan={7} className="px-4 py-8 text-center text-gray-500">
                            加载中...
                          </td>
                        </tr>
                      ) : permissions.length === 0 ? (
                        <tr>
                          <td colSpan={7} className="px-4 py-8 text-center text-gray-500">
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
          {roleModalOpen && (
            <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
              <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <h3 className="text-lg font-semibold mb-4">新建角色</h3>
                <div className="space-y-4">
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
                </div>
                <div className="flex justify-end gap-2 mt-6">
                  <Button variant="outline" onClick={handleCloseRoleModal}>取消</Button>
                  <Button onClick={handleSaveRole} disabled={roleSaving}>
                    {roleSaving ? '保存中...' : '保存'}
                  </Button>
                </div>
              </div>
            </div>
          )}
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
