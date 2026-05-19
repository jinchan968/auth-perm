'use client'

import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { Save, Check, Trash2, Loader2 } from 'lucide-react'
import { getRole, getRolePermissions, assignPermissionsToRole, deleteRole } from '@/lib/api/role'
import { listPermissions as listPermApi } from '@/lib/api/permission'
import { Role, PermissionListItem } from '@/types/permission'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { DetailActionBar } from '@/components/ui/detail-action-bar'
import { DetailPageHeader } from '@/components/ui/detail-page-header'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogTrigger,
} from '@/components/ui/dialog'
import { useTenant } from '@/lib/tenant-context'
import { ShellLayout } from '@/components/layout/shell-layout'
import { ListReturnButton } from '@/components/ui/list-return-button'
import { showError } from '@/lib/toast'
import { useNavigationTransition } from '@/components/providers/navigation-transition-provider'

export default function RolePermissionsPage() {
  const params = useParams()
  const roleId = params.id as string
  const rolesListHref = '/permissions?tab=roles'
  const { navigateWithTransition } = useNavigationTransition()

  // 使用统一的租户上下文（仅需 selectedTenantId）
  const { selectedTenantId } = useTenant()

  const [role, setRole] = useState<Role | null>(null)
  const [allPermissions, setAllPermissions] = useState<PermissionListItem[]>([])
  const [assignedPermissionIds, setAssignedPermissionIds] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [saveSuccessOpen, setSaveSuccessOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [loadFailed, setLoadFailed] = useState(false)


  useEffect(() => {
    const fetchData = async () => {
      if (!selectedTenantId) return
      setLoading(true)
      setLoadFailed(false)
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
        showError(err instanceof Error ? err.message : 'Failed to fetch data')
        setLoadFailed(true)
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
    if (!selectedTenantId) {
      showError('租户ID缺失')
      return
    }

    setSaving(true)
    try {
      await assignPermissionsToRole({
        role_id: roleId,
        permission_ids: assignedPermissionIds,
        tenant_id: selectedTenantId,
      })
      setSaveSuccessOpen(true)
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Failed to save permissions')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <ShellLayout pathname="/permissions">
        <div className="text-center">加载中...</div>
      </ShellLayout>
    )
  }

  if (loadFailed && !role) {
    return (
      <ShellLayout pathname="/permissions">
        <ListReturnButton href={rolesListHref} label="返回列表" />
      </ShellLayout>
    )
  }

  if (!role) {
    return (
      <ShellLayout pathname="/permissions">
        <div className="text-center">角色不存在</div>
        <ListReturnButton href={rolesListHref} label="返回列表" />
      </ShellLayout>
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
    <ShellLayout pathname="/permissions">
          <Breadcrumb items={breadcrumbItems} />

          <DetailPageHeader
            title={`角色权限分配 - ${role.name}`}
            actions={
              <DetailActionBar returnHref={rolesListHref} returnLabel="返回">
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
                            navigateWithTransition(rolesListHref, {
                              onBeforeNavigate: () => setDeleteDialogOpen(false),
                            })
                          } catch (err) {
                            showError(err instanceof Error ? err.message : '删除失败')
                            setDeleteDialogOpen(false)
                          }
                        }}
                      >
                        确认删除
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
                <Button onClick={handleSave} disabled={saving} className="min-w-[132px] active:scale-100">
                  <span className="relative flex items-center justify-center">
                    <span
                      className={`flex items-center justify-center transition-all duration-200 ${
                        saving ? 'opacity-0 -translate-y-1 scale-95' : 'opacity-100 translate-y-0 scale-100'
                      }`}
                    >
                      <Save className="h-4 w-4 mr-1" />
                      <span className="inline-flex min-w-[32px] justify-center">保存</span>
                    </span>
                    <span
                      className={`absolute inset-0 flex items-center justify-center transition-all duration-200 ${
                        saving ? 'opacity-100 translate-y-0 scale-100' : 'opacity-0 translate-y-1 scale-95'
                      }`}
                    >
                      <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                      <span className="inline-flex min-w-[60px] justify-center">保存中...</span>
                    </span>
                  </span>
                </Button>
              </DetailActionBar>
            }
          />

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

      <Dialog open={saveSuccessOpen} onOpenChange={setSaveSuccessOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>保存成功</DialogTitle>
          </DialogHeader>
          <p className="py-2 text-sm text-slate-600">角色权限分配已保存。</p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setSaveSuccessOpen(false)}>
              继续编辑
            </Button>
            <ListReturnButton href={rolesListHref} label="返回列表" onBeforeNavigate={() => setSaveSuccessOpen(false)} />
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </ShellLayout>
  )
}
