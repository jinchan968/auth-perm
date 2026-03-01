'use client'

import { useState, useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'
import Link from 'next/link'
import { Save, Check, Trash2 } from 'lucide-react'
import { getRole, getRolePermissions, assignPermissionsToRole, deleteRole } from '@/lib/api/role'
import { listPermissions as listPermApi } from '@/lib/api/permission'
import { Role, PermissionListItem } from '@/types/permission'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogTrigger,
} from '@/components/ui/dialog'
import { useTenant } from '@/lib/tenant-context'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'

export default function RolePermissionsPage() {
  const router = useRouter()
  const params = useParams()
  const roleId = params.id as string
  const { user } = useAuthStore()

  // 使用统一的租户上下文（仅需 selectedTenantId）
  const { selectedTenantId } = useTenant()

  const [role, setRole] = useState<Role | null>(null)
  const [allPermissions, setAllPermissions] = useState<PermissionListItem[]>([])
  const [assignedPermissionIds, setAssignedPermissionIds] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [error, setError] = useState('')


  useEffect(() => {
    const fetchData = async () => {
      if (!selectedTenantId) return
      setLoading(true)
      setError('')
      try {
        const [roleData, permissionsData, assignedData] = await Promise.all([
          getRole(roleId, selectedTenantId),
          listPermApi({ tenant_id: selectedTenantId, size: 1000 }),
          getRolePermissions(roleId, selectedTenantId),
        ])
        setRole(roleData)
        setAllPermissions(permissionsData.data || [])
        setAssignedPermissionIds(assignedData.map((p) => p.id))
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch data')
      } finally {
        setLoading(false)
      }
    }
    if (roleId && selectedTenantId) {
      fetchData().then(()=>{})
    }
  }, [roleId, selectedTenantId])

  const handleTogglePermission = (permissionId: string) => {
    setAssignedPermissionIds((prev) =>
      prev.includes(permissionId)
        ? prev.filter((id) => id !== permissionId)
        : [...prev, permissionId]
    )
  }

  const handleSave = async () => {
    setSaving(true)
    setError('')
    try {
      await assignPermissionsToRole({
        role_id: roleId,
        permission_ids: assignedPermissionIds,
        tenant_id: selectedTenantId,
      })
      router.push('/permissions?tab=roles')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save permissions')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
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
            <div className="text-center">加载中...</div>
          </main>
        </div>
      </div>
    )
  }

  if (error && !role) {
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
            <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>
            <Link href="/permissions">
              <Button variant="outline">返回列表</Button>
            </Link>
          </main>
        </div>
      </div>
    )
  }

  if (!role) {
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
            <div className="text-center">角色不存在</div>
            <Link href="/permissions">
              <Button variant="outline">返回列表</Button>
            </Link>
          </main>
        </div>
      </div>
    )
  }

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '权限管理', href: '/permissions' },
    { label: '角色管理', href: '/permissions' },
    { label: '权限分配' },
  ]

  // Group permissions by resource
  const permissionsByResource = allPermissions.reduce((acc, perm) => {
    const resource = perm.resource || 'other'
    if (!acc[resource]) {
      acc[resource] = []
    }
    acc[resource].push(perm)
    return acc
  }, {} as Record<string, PermissionListItem[]>)

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

          <div className="flex justify-between items-center mb-6 mt-4">
            <h2 className="text-xl font-semibold">角色权限分配 - {role.name}</h2>
            <div className="flex gap-2">
              <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
                <DialogTrigger asChild>
                  <Button variant="destructive">
                    <Trash2 className="h-4 w-4 mr-1" />
                    删除
                  </Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>确认删除角色</DialogTitle>
                  </DialogHeader>
                  <p>
                    确定要删除角色 <strong>{role.name}</strong> 吗？此操作无法撤销。
                  </p>
                  <DialogFooter>
                    <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>取消</Button>
                    <Button
                      variant="destructive"
                      onClick={async () => {
                        try {
                          await deleteRole(roleId, selectedTenantId)
                          router.push('/permissions?tab=roles')
                        } catch (err) {
                          setError(err instanceof Error ? err.message : '删除失败')
                          setDeleteDialogOpen(false)
                        }
                      }}
                    >
                      确认删除
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
              <Button onClick={handleSave} disabled={saving}>
                <Save className="h-4 w-4 mr-1" />
                {saving ? '保存中...' : '保存'}
              </Button>
              <Link href="/permissions?tab=roles">
                <Button variant="outline">返回</Button>
              </Link>
            </div>
          </div>

          {error && (
            <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>
          )}

          <div className="text-sm text-gray-500 mb-4">
            已选择 {assignedPermissionIds.length} / {allPermissions.length} 个权限
          </div>

          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {Object.entries(permissionsByResource).map(([resource, perms]) => (
              <Card key={resource}>
                <CardHeader>
                  <CardTitle className="text-base">{resource}</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    {perms.map((perm) => (
                      <div
                        key={perm.id}
                        className="flex items-center gap-2 cursor-pointer hover:bg-gray-50 p-1 rounded"
                        onClick={() => handleTogglePermission(perm.id)}
                      >
                        <div
                          className={`w-4 h-4 border rounded flex items-center justify-center ${
                            assignedPermissionIds.includes(perm.id)
                              ? 'bg-blue-600 border-blue-600'
                              : 'border-gray-300'
                          }`}
                        >
                          {assignedPermissionIds.includes(perm.id) && (
                            <Check className="w-3 h-3 text-white" />
                          )}
                        </div>
                        <div className="flex-1">
                          <div className="text-sm">{perm.name}</div>
                          <div className="text-xs text-gray-400">{perm.code}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {allPermissions.length === 0 && (
            <Card>
              <CardContent className="py-8 text-center text-gray-500">
                暂无可分配的权限
              </CardContent>
            </Card>
          )}
        </main>
      </div>
    </div>
  )
}
