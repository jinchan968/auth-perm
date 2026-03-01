'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { Save, Check } from 'lucide-react'
import { User } from '@/lib/api/auth'
import { getUser } from '@/lib/api/user'
import { listRoles, assignRoleToAccount, getUserRoles } from '@/lib/api/role'
import { AccountListItem } from '@/types/user'
import { RoleListItem } from '@/types/permission'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
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
  const [saveSuccessOpen, setSaveSuccessOpen] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (accountId === 'new' || !accountId) {
      router.replace('/permissions?tab=users')
      return
    }
    const fetchData = async () => {
      if (!selectedTenantId) return
      setLoading(true)
      setError('')
      try {
        const [user, rolesData, userRolesData] = await Promise.all([
          getUser(accountId, selectedTenantId),
          listRoles({ tenant_id: selectedTenantId, size: 1000 }),
          getUserRoles(accountId, selectedTenantId),
        ])
        setUserDetail(user)
        setRoles(rolesData.data || [])
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
      prev.includes(roleId) ? prev.filter((id) => id !== roleId) : [...prev, roleId]
    )
  }

  const handleSave = async () => {
    if (!selectedTenantId) { setError('租户ID缺失'); return }
    setSaving(true)
    setError('')
    try {
      await assignRoleToAccount({
        account_id: accountId,
        role_ids: assignedRoleIds,
        tenant_id: selectedTenantId,
      })
      setSaveSuccessOpen(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存失败')
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
            <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error || '用户不存在'}</div>
            <Link href="/permissions?tab=users">
              <Button variant="outline">返回列表</Button>
            </Link>
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

          {/* ── 标题栏，与角色详情保持一致 ── */}
          <div className="flex justify-between items-center mb-6 mt-4">
            <h2 className="text-xl font-semibold">
              用户角色分配 - {userDetail.username || userDetail.nickname || accountId}
            </h2>
            <div className="flex gap-2">
              <Button onClick={handleSave} disabled={saving}>
                <Save className="h-4 w-4 mr-1" />
                {saving ? '保存中...' : '保存'}
              </Button>
              <Link href="/permissions?tab=users">
                <Button variant="outline">返回</Button>
              </Link>
            </div>
          </div>

          {error && (
            <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>
          )}

          {/* ── 用户基本信息 ── */}
          <Card className="mb-6">
            <CardHeader>
              <CardTitle className="text-base">基本信息</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <InfoRow label="用户名" value={userDetail.username} />
                <InfoRow label="昵称" value={userDetail.nickname} />
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
                return (
                  <Card
                    key={role.id}
                    className={`cursor-pointer transition-all hover:shadow-md ${
                      isAssigned ? 'ring-2 ring-blue-500 bg-blue-50/50' : ''
                    }`}
                    onClick={() => handleToggleRole(role.id)}
                  >
                    <CardHeader className="pb-2">
                      <CardTitle className="text-base flex items-center gap-2">
                        <div className={`w-4 h-4 border rounded flex items-center justify-center flex-shrink-0 ${
                          isAssigned ? 'bg-blue-600 border-blue-600' : 'border-gray-300'
                        }`}>
                          {isAssigned && <Check className="w-3 h-3 text-white" />}
                        </div>
                        {role.name}
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-xs text-gray-400 font-mono">{role.code}</div>
                      {role.description && (
                        <div className="text-xs text-gray-500 mt-1 line-clamp-2">{role.description}</div>
                      )}
                    </CardContent>
                  </Card>
                )
              })}
            </div>
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
