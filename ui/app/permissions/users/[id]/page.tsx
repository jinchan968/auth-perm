'use client'

import { useState, useEffect, useRef } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { Save, Check, Eye, Loader2 } from 'lucide-react'
import { User } from '@/lib/api/auth'
import { getUser } from '@/lib/api/user'
import { listRoles, assignRoleToAccount, getUserRoles, getRolePermissions } from '@/lib/api/role'
import { listPermissionResources, type PermissionResource } from '@/lib/api/permission-resource'
import { AccountListItem } from '@/types/user'
import { PermissionListItem, RoleListItem } from '@/types/permission'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { DetailActionBar } from '@/components/ui/detail-action-bar'
import { DetailPageHeader } from '@/components/ui/detail-page-header'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { useTenant } from '@/lib/tenant-context'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'
import { ListReturnButton } from '@/components/ui/list-return-button'
import { showError } from '@/lib/toast'

// 公共 Header，与角色详情页保持一致
function PageHeader({ user }: { user: User | null }) {
  return (
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
  )
}

function InfoRow({ label, value }: { label: string; value?: string | null }) {
  return (
    <div>
      <div className="text-gray-500 mb-1">{label}</div>
      <div className="font-medium">{value || '-'}</div>
    </div>
  )
}

interface RoleDetailCacheEntry {
  permissions: PermissionListItem[]
  resources: Record<string, PermissionResource[]>
}

export default function UserDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { user: currentUser } = useAuthStore()
  const { selectedTenantId, tenants } = useTenant()
  const usersListHref = '/permissions?tab=users'

  const accountId = params.id as string

  const [userDetail, setUserDetail] = useState<AccountListItem | null>(null)
  const [roles, setRoles] = useState<RoleListItem[]>([])
  const [assignedRoleIds, setAssignedRoleIds] = useState<string[]>([])
  const [activeRoleId, setActiveRoleId] = useState<string | null>(null)
  const [activeRolePermissions, setActiveRolePermissions] = useState<PermissionListItem[]>([])
  const [activeRoleResources, setActiveRoleResources] = useState<Record<string, PermissionResource[]>>({})
  const [roleDetailLoading, setRoleDetailLoading] = useState(false)
  const [hasLoadedRoleDetails, setHasLoadedRoleDetails] = useState(false)
  const [roleDetailError, setRoleDetailError] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [saveSuccessOpen, setSaveSuccessOpen] = useState(false)
  
  const roleDetailCacheRef = useRef<Record<string, RoleDetailCacheEntry>>({})

  const refreshAssignedRoles = async (tenantId: string) => {
    const userRolesData = await getUserRoles(accountId, tenantId)
    const nextAssignedRoleIds = userRolesData.map((role) => role.id)

    setAssignedRoleIds(nextAssignedRoleIds)
    setActiveRoleId((prev) => {
      if (prev) {
        return prev
      }

      return nextAssignedRoleIds[0] || null
    })
  }

  useEffect(() => {
    if (accountId === 'new' || !accountId) {
      router.replace(usersListHref)
      return
    }
    const fetchData = async () => {
      if (!selectedTenantId) return
      setLoading(true)
      try {
        const [user, rolesData, userRolesData] = await Promise.all([
          getUser(accountId, selectedTenantId),
          listRoles({ tenant_id: selectedTenantId, size: 1000 }),
          getUserRoles(accountId, selectedTenantId),
        ])
        setUserDetail(user)
        setRoles(rolesData.data || [])
        const nextAssignedRoleIds = userRolesData.map((r) => r.id)
        setAssignedRoleIds(nextAssignedRoleIds)
        setActiveRoleId((prev) => {
          if (prev && rolesData.data?.some((role) => role.id === prev)) {
            return prev
          }

          return nextAssignedRoleIds[0] || rolesData.data?.[0]?.id || null
        })
      } catch (err) {
        showError(err instanceof Error ? err.message : '加载数据失败')
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [selectedTenantId, accountId, router])

  useEffect(() => {
    roleDetailCacheRef.current = {}
  }, [selectedTenantId, accountId])

  useEffect(() => {
    const fetchActiveRoleDetails = async () => {
      if (!selectedTenantId || !activeRoleId) {
        setActiveRolePermissions([])
        setActiveRoleResources({})
        setRoleDetailError('')
        setHasLoadedRoleDetails(false)
        return
      }

      const cacheKey = `${selectedTenantId}:${activeRoleId}`
      const cachedDetails = roleDetailCacheRef.current[cacheKey]
      if (cachedDetails) {
        setActiveRolePermissions(cachedDetails.permissions)
        setActiveRoleResources(cachedDetails.resources)
        setRoleDetailError('')
        setHasLoadedRoleDetails(true)
        setRoleDetailLoading(false)
        return
      }

      setRoleDetailLoading(true)
      setRoleDetailError('')

      try {
        const permissions = await getRolePermissions(activeRoleId, selectedTenantId)
        setActiveRolePermissions(permissions)

        if (permissions.length === 0) {
          roleDetailCacheRef.current[cacheKey] = {
            permissions: [],
            resources: {},
          }
          setActiveRoleResources({})
          setHasLoadedRoleDetails(true)
          return
        }

        const resourcesEntries = await Promise.all(
          permissions.map(async (permission) => {
            const response = await listPermissionResources(permission.id, {
              tenant_id: selectedTenantId,
              size: 100,
            })

            return [permission.id, response.data || []] as const
          })
        )

        const nextResources = Object.fromEntries(resourcesEntries)
        setActiveRoleResources(nextResources)
        roleDetailCacheRef.current[cacheKey] = {
          permissions,
          resources: nextResources,
        }
        setHasLoadedRoleDetails(true)
      } catch (err) {
        setActiveRolePermissions([])
        setActiveRoleResources({})
        showError(err instanceof Error ? err.message : '加载角色权限与资源失败')
        setRoleDetailError(err instanceof Error ? err.message : '加载角色权限与资源失败')
        setHasLoadedRoleDetails(true)
      } finally {
        setRoleDetailLoading(false)
      }
    }

    fetchActiveRoleDetails()
  }, [activeRoleId, selectedTenantId])

  const handleSelectRole = (roleId: string) => {
    setActiveRoleId(roleId)
  }

  const handleToggleAssignedRole = (roleId: string) => {
    setAssignedRoleIds((prev) =>
      prev.includes(roleId) ? prev.filter((id) => id !== roleId) : [...prev, roleId]
    )
  }

  const handleSave = async () => {
    if (!selectedTenantId) { showError('租户ID缺失'); return }
    setSaving(true)
    try {
      await assignRoleToAccount({
        account_id: accountId,
        role_ids: assignedRoleIds,
        tenant_id: selectedTenantId,
      })

      await refreshAssignedRoles(selectedTenantId)
      setSaveSuccessOpen(true)
    } catch (err) {
      showError(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '权限管理', href: '/permissions?tab=users' },
    { label: '用户列表', href: '/permissions?tab=users' },
    { label: '用户详情' },
  ]

  const userTenant = userDetail
    ? tenants.find((tenant) => tenant.id === userDetail.tenant_id)
    : null
  const activeRole = activeRoleId
    ? roles.find((role) => role.id === activeRoleId) || null
    : null

  const getResourceTypeLabel = (resourceType: string) => {
    switch (resourceType) {
      case 'api_path':
        return 'API'
      case 'menu':
        return '菜单'
      case 'button':
        return '按钮'
      default:
        return resourceType
    }
  }

  // ── Loading ──
  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-slate-50">
        <PageHeader user={currentUser} />
        <div className="flex">
          <DashboardSidebar pathname="/permissions" />
          <main className="flex-1 p-8">
            <div className="text-center py-8 text-slate-400">加载中...</div>
          </main>
        </div>
      </div>
    )
  }

  // ── Not found / error ──
  if (!userDetail) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-slate-50">
        <PageHeader user={currentUser} />
        <div className="flex">
          <DashboardSidebar pathname="/permissions" />
          <main className="flex-1 p-8">
            <ListReturnButton href={usersListHref} label="返回列表" />
          </main>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-slate-50">
      <PageHeader user={currentUser} />

      <div className="flex">
        <DashboardSidebar pathname="/permissions" />
        <main className="flex-1 p-8">
          <Breadcrumb items={breadcrumbItems} />

          <DetailPageHeader
            title={`用户角色分配 - ${userDetail.username || userDetail.nickname || accountId}`}
            actions={
              <DetailActionBar returnHref={usersListHref} returnLabel="返回">
                <Button onClick={handleSave} disabled={saving}>
                  <Save className="h-4 w-4 mr-1" />
                  {saving ? '保存中...' : '保存'}
                </Button>
              </DetailActionBar>
            }
          />

          {/* ── 用户基本信息 ── */}
          <Card className="mb-6">
            <CardHeader>
              <CardTitle className="text-base">基本信息</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <InfoRow label="用户名" value={userDetail.username} />
                <InfoRow label="昵称" value={userDetail.nickname} />
                <InfoRow
                  label="所属租户"
                  value={userTenant ? `${userTenant.name} (${userTenant.code})` : userDetail.tenant_id}
                />
                <InfoRow label="邮箱" value={userDetail.email} />
                <InfoRow label="手机号" value={userDetail.phone} />
                <div>
                  <div className="text-gray-500 mb-1">账户类型</div>
                  <Badge variant="outline">{userDetail.account_type}</Badge>
                </div>
                <div>
                  <div className="text-gray-500 mb-1">账户状态</div>
                  <Badge variant={userDetail.account_status === 'active' ? 'default' : 'secondary'}>
                    {userDetail.account_status === 'active' ? '活跃'
                      : userDetail.account_status === 'inactive' ? '停用' : '暂停'}
                  </Badge>
                </div>
                <div>
                  <div className="text-gray-500 mb-1">邮箱验证</div>
                  <Badge variant={userDetail.email_verified ? 'default' : 'outline'}>
                    {userDetail.email_verified ? '已验证' : '未验证'}
                  </Badge>
                </div>
                <InfoRow
                  label="最后登录"
                  value={userDetail.last_login_at
                    ? new Date(userDetail.last_login_at).toLocaleString('zh-CN')
                    : '从未登录'}
                />
                <InfoRow label="创建时间" value={new Date(userDetail.created_at).toLocaleString('zh-CN')} />
                <InfoRow label="更新时间" value={new Date(userDetail.updated_at).toLocaleString('zh-CN')} />
              </div>
            </CardContent>
          </Card>

          {/* ── 角色分配，卡片风格与权限分配一致 ── */}
          <div className="text-sm text-gray-500 mb-4">
            已选择 {assignedRoleIds.length} / {roles.length} 个角色
          </div>
          <div className="mb-4 text-xs text-slate-500">
            使用“查看详情”浏览角色权限与资源，使用“分配角色/取消分配”调整当前用户的角色关联。
          </div>

          {roles.length === 0 ? (
            <Card>
              <CardContent className="py-8 text-center text-gray-500">
                当前租户暂无可分配的角色
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {roles.map((role) => {
                const isAssigned = assignedRoleIds.includes(role.id)
                const isActive = activeRoleId === role.id

                return (
                  <Card
                    key={role.id}
                    className={`transition-all hover:shadow-md ${
                      isActive ? 'border-indigo-300 ring-2 ring-indigo-500 shadow-md' : 'border-slate-200'
                    }`}
                  >
                    <CardHeader className="pb-2">
                      <CardTitle className="text-base flex items-center gap-2">
                        {role.name}
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="flex items-center justify-between gap-2">
                        <div className="text-xs text-gray-400 font-mono">{role.code}</div>
                        <div className="flex items-center gap-2">
                          {isAssigned && <Badge variant="outline">已分配</Badge>}
                          {isActive && <Badge variant="secondary">当前查看</Badge>}
                        </div>
                      </div>
                      {role.description && (
                        <div className="text-xs text-gray-500 mt-1 line-clamp-2">{role.description}</div>
                      )}
                      <div className="mt-4 flex flex-wrap items-center gap-2">
                        <Button
                          type="button"
                          size="sm"
                          variant={isActive ? 'default' : 'outline'}
                          onClick={() => handleSelectRole(role.id)}
                        >
                          <Eye className="h-4 w-4" />
                          {isActive ? '查看中' : '查看详情'}
                        </Button>
                        <Button
                          type="button"
                          size="sm"
                          variant={isAssigned ? 'secondary' : 'outline'}
                          onClick={() => handleToggleAssignedRole(role.id)}
                        >
                          <Check className="h-4 w-4" />
                          {isAssigned ? '取消分配' : '分配角色'}
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                )
              })}
            </div>
          )}

          {activeRole && (
            <Card className="mt-6">
              <CardHeader>
                <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                  <div>
                    <CardTitle className="text-base">角色权限与资源</CardTitle>
                    <div className="mt-1 text-sm text-gray-500">
                      当前查看：{activeRole.name}
                      {activeRole.description ? ` · ${activeRole.description}` : ''}
                    </div>
                  </div>
                  <div className="flex flex-wrap items-center gap-2 text-xs text-gray-500">
                    <Badge variant={assignedRoleIds.includes(activeRole.id) ? 'default' : 'outline'}>
                      {assignedRoleIds.includes(activeRole.id) ? '已分配' : '未分配'}
                    </Badge>
                    <span>权限 {activeRolePermissions.length} 项</span>
                    {roleDetailLoading && hasLoadedRoleDetails && (
                      <span className="inline-flex items-center gap-1 text-indigo-600">
                        <Loader2 className="h-3 w-3 animate-spin" />
                        更新中
                      </span>
                    )}
                  </div>
                </div>
              </CardHeader>
              <CardContent className="min-h-[220px]">
                {!hasLoadedRoleDetails && roleDetailLoading ? (
                  <div className="flex min-h-[188px] items-center justify-center gap-2 py-8 text-sm text-slate-500">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    加载角色权限与资源中...
                  </div>
                ) : (
                  <div
                    className={`transition-opacity duration-200 ease-in-out ${
                      roleDetailLoading && hasLoadedRoleDetails ? 'opacity-55' : 'opacity-100'
                    }`}
                  >
                    {activeRolePermissions.length === 0 ? (
                      <div className="rounded border border-dashed p-6 text-center text-sm text-gray-500">
                        当前角色暂无权限或资源配置
                      </div>
                    ) : (
                      <div className="space-y-4">
                        {activeRolePermissions.map((permission) => {
                          const resources = activeRoleResources[permission.id] || []

                          return (
                            <div key={permission.id} className="rounded-lg border bg-white p-4">
                              <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                                <div className="space-y-1">
                                  <div className="flex flex-wrap items-center gap-2">
                                    <span className="font-medium text-slate-900">{permission.name}</span>
                                    <Badge variant="outline">{permission.resource || '未分类资源'}</Badge>
                                    {!permission.is_active && <Badge variant="secondary">已禁用</Badge>}
                                  </div>
                                  <div className="font-mono text-xs text-slate-500">{permission.code}</div>
                                  {permission.description && (
                                    <div className="text-sm text-slate-600">{permission.description}</div>
                                  )}
                                </div>
                                <div className="text-xs text-slate-500">
                                  关联资源 {resources.length} 项
                                </div>
                              </div>

                              <div className="mt-4 space-y-2">
                                {resources.length === 0 ? (
                                  <div className="text-sm text-gray-500">暂无关联资源</div>
                                ) : (
                                  resources.map((resource) => (
                                    <div
                                      key={resource.id}
                                      className="flex flex-col gap-2 rounded-md bg-slate-50 px-3 py-2 md:flex-row md:items-center md:justify-between"
                                    >
                                      <div className="space-y-1">
                                        <div className="flex flex-wrap items-center gap-2">
                                          <Badge variant="secondary">{getResourceTypeLabel(resource.resource_type)}</Badge>
                                          <span className="text-sm font-medium text-slate-900">
                                            {resource.resource_name}
                                          </span>
                                        </div>
                                        <div className="font-mono text-xs text-slate-500">
                                          {resource.resource_id}
                                        </div>
                                      </div>
                                    </div>
                                  ))
                                )}
                              </div>
                            </div>
                          )
                        })}
                      </div>
                    )}
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </main>
      </div>

      {/* 保存成功弹窗（替换原来的 alert）*/}
      <Dialog open={saveSuccessOpen} onOpenChange={setSaveSuccessOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>保存成功</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-slate-600 py-2">角色分配已保存。</p>
          <DialogFooter>
            <Button onClick={() => setSaveSuccessOpen(false)}>确定</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
