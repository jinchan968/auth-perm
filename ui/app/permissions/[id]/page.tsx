'use client'

import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { Save, Edit2, X, Plus, Trash2 } from 'lucide-react'
import { getPermission, deletePermission, updatePermission } from '@/lib/api/permission'
import {
  listPermissionResources,
  createPermissionResource,
  deletePermissionResource,
  PermissionResource,
} from '@/lib/api/permission-resource'
import { Permission } from '@/types/permission'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { DetailActionBar } from '@/components/ui/detail-action-bar'
import { DetailPageHeader } from '@/components/ui/detail-page-header'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'
import { ListReturnButton } from '@/components/ui/list-return-button'
import { showError } from '@/lib/toast'
import { useNavigationTransition } from '@/components/providers/navigation-transition-provider'

export default function PermissionDetailPage() {
  const params = useParams()
  const id = params.id as string
  const permissionsListHref = '/permissions'
  const { navigateWithTransition } = useNavigationTransition()

  const { user } = useAuthStore()
  const tenantId = user?.tenant_id || ''

  const [permission, setPermission] = useState<Permission | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadFailed, setLoadFailed] = useState(false)
  const [deleting, setDeleting] = useState(false)

  // Edit mode state
  const [isEditing, setIsEditing] = useState(false)
  const [saving, setSaving] = useState(false)

  // Edit form data
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    is_active: true,
  })

  // Resource management state
  const [resources, setResources] = useState<PermissionResource[]>([])
  const [resourceLoading, setResourceLoading] = useState(false)
  const [showResourceForm, setShowResourceForm] = useState(false)
  const [resourceForm, setResourceForm] = useState({
    resource_id: '',
    resource_type: 'api_path',
    resource_name: '',
  })
  const [savingResource, setSavingResource] = useState(false)

  useEffect(() => {
    const fetchPermission = async () => {
      setLoading(true)
      try {
        const data = await getPermission(id, tenantId)
        setPermission(data)
      } catch (err) {
        showError(err instanceof Error ? err.message : 'Failed to fetch permission')
        setLoadFailed(true)
      } finally {
        setLoading(false)
      }
    }
    fetchPermission()
  }, [id])

  // Fetch resources when permission is loaded
  useEffect(() => {
    const fetchResources = async () => {
      if (!id) return
      setResourceLoading(true)
      try {
        const data = await listPermissionResources(id, { size: 100, tenant_id: tenantId })
        setResources(data.data || [])
      } catch (err) {
        console.error('Failed to fetch resources:', err)
      } finally {
        setResourceLoading(false)
      }
    }
    fetchResources()
  }, [id, tenantId])

  // Initialize form data when entering edit mode
  useEffect(() => {
    if (isEditing && permission) {
      setFormData({
        name: permission.name,
        description: permission.description || '',
        is_active: permission.is_active,
      })
    }
  }, [isEditing, permission])

  const handleToggleEdit = () => {
    if (!isEditing) {
      // Enter edit mode - initialize form data
      if (permission) {
        setFormData({
          name: permission.name,
          description: permission.description || '',
          is_active: permission.is_active,
        })
      }
    }
    setIsEditing(!isEditing)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const request: any = {}
      if (formData.name !== permission?.name) request.name = formData.name
      if (formData.description !== permission?.description) request.description = formData.description
      if (formData.is_active !== permission?.is_active) request.is_active = formData.is_active

      const updated = await updatePermission(id, tenantId, request)
      setPermission(updated)
      setIsEditing(false)
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Failed to save permission')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!confirm('确定要删除此权限吗？')) return

    setDeleting(true)
    try {
      await deletePermission(id, tenantId)
      navigateWithTransition(permissionsListHref)
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Failed to delete permission')
    } finally {
      setDeleting(false)
    }
  }

  const handleAddResource = async () => {
    if (!resourceForm.resource_id || !resourceForm.resource_name) {
      showError('请填写资源标识和资源名称')
      return
    }
    setSavingResource(true)
    try {
      await createPermissionResource(id, {
        permission_id: id,
        resource_id: resourceForm.resource_id,
        resource_type: resourceForm.resource_type,
        resource_name: resourceForm.resource_name,
        tenant_id: tenantId,
      })
      // Refresh resources
      const data = await listPermissionResources(id, { size: 100, tenant_id: tenantId })
      setResources(data.data || [])
      setShowResourceForm(false)
      setResourceForm({ resource_id: '', resource_type: 'api_path', resource_name: '' })
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Failed to add resource')
    } finally {
      setSavingResource(false)
    }
  }

  const handleDeleteResource = async (resourceId: string) => {
    if (!confirm('确定要删除此资源关联吗？')) return
    try {
      await deletePermissionResource(id, resourceId, tenantId)
      setResources(resources.filter(r => r.id !== resourceId))
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Failed to delete resource')
    }
  }

  const getResourceTypeBadge = (type: string) => {
    const types: Record<string, { label: string; variant: 'default' | 'secondary' | 'outline' }> = {
      api_path: { label: 'API', variant: 'default' },
      menu: { label: '菜单', variant: 'secondary' },
      button: { label: '按钮', variant: 'outline' },
    }
    const config = types[type] || { label: type, variant: 'secondary' }
    return <Badge variant={config.variant}>{config.label}</Badge>
  }

  const getStatusBadge = (isActive: boolean) => {
    return isActive ? (
      <Badge variant="default">启用</Badge>
    ) : (
      <Badge variant="secondary">禁用</Badge>
    )
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

  if (loadFailed && !permission) {
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
            <ListReturnButton href={permissionsListHref} label="返回列表" />
          </main>
        </div>
      </div>
    )
  }

  if (!permission) {
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
            <div className="text-center">权限不存在</div>
            <ListReturnButton href={permissionsListHref} label="返回列表" />
          </main>
        </div>
      </div>
    )
  }

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '权限管理', href: '/permissions' },
    { label: permission?.name || '权限详情' },
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

          <DetailPageHeader
            title={permission.name}
            actions={
              <DetailActionBar returnHref={permissionsListHref} returnLabel="返回">
                {isEditing ? (
                  <>
                    <Button variant="outline" onClick={handleToggleEdit}>
                      <X className="h-4 w-4 mr-1" />
                      取消
                    </Button>
                    <Button onClick={handleSave} disabled={saving}>
                      <Save className="h-4 w-4 mr-1" />
                      {saving ? '保存中...' : '保存'}
                    </Button>
                  </>
                ) : !permission.is_system ? (
                  <Button onClick={handleToggleEdit}>
                    <Edit2 className="h-4 w-4 mr-1" />
                    编辑
                  </Button>
                ) : null}
                {!permission.is_system && (
                  <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
                    {deleting ? '删除中...' : '删除'}
                  </Button>
                )}
              </DetailActionBar>
            }
          />

          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>基本信息</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <div className="text-sm text-gray-500">ID</div>
                  <div className="font-mono text-sm">{permission.id}</div>
                </div>

                {isEditing ? (
                  // Edit mode
                  <>
                    <div className="space-y-2">
                      <Label htmlFor="name">名称</Label>
                      <Input
                        id="name"
                        value={formData.name}
                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      />
                    </div>
                    <div>
                      <div className="text-sm text-gray-500">代码</div>
                      <div className="font-mono">{permission.code}</div>
                    </div>
                    <div>
                      <div className="text-sm text-gray-500">资源</div>
                      <div>{permission.resource}</div>
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="description">描述</Label>
                      <Input
                        id="description"
                        value={formData.description}
                        onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                      />
                    </div>
                    <div className="flex items-center gap-2">
                      <Switch
                        id="is_active"
                        checked={formData.is_active}
                        onCheckedChange={(checked) => setFormData({ ...formData, is_active: checked })}
                      />
                      <Label htmlFor="is_active">启用</Label>
                    </div>
                  </>
                ) : (
                  // View mode
                  <>
                    <div>
                      <div className="text-sm text-gray-500">名称</div>
                      <div>{permission.name}</div>
                    </div>
                    <div>
                      <div className="text-sm text-gray-500">代码</div>
                      <div className="font-mono">{permission.code}</div>
                    </div>
                    <div>
                      <div className="text-sm text-gray-500">资源</div>
                      <div>{permission.resource}</div>
                    </div>
                    <div>
                      <div className="text-sm text-gray-500">描述</div>
                      <div>{permission.description || '-'}</div>
                    </div>
                    <div>
                      <div className="text-sm text-gray-500">系统权限</div>
                      <div>
                        {permission.is_system ? (
                          <Badge variant="outline">是</Badge>
                        ) : (
                          <span>否</span>
                        )}
                      </div>
                    </div>
                    <div>
                      <div className="text-sm text-gray-500">状态</div>
                      <div>{getStatusBadge(permission.is_active)}</div>
                    </div>
                  </>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <div className="flex justify-between items-center">
                  <CardTitle>关联资源</CardTitle>
                  {!permission.is_system && (
                    <Button
                      size="sm"
                      onClick={() => setShowResourceForm(!showResourceForm)}
                    >
                      <Plus className="h-4 w-4 mr-1" />
                      添加资源
                    </Button>
                  )}
                </div>
              </CardHeader>
              <CardContent>
                {/* Add Resource Form */}
                {showResourceForm && (
                  <div className="mb-4 p-4 bg-gray-50 rounded space-y-3">
                    <div className="grid grid-cols-3 gap-3">
                      <div>
                        <Label>资源类型</Label>
                        <Select
                          value={resourceForm.resource_type}
                          onValueChange={(value) => setResourceForm({ ...resourceForm, resource_type: value })}
                        >
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="api_path">API 路径</SelectItem>
                            <SelectItem value="menu">菜单</SelectItem>
                            <SelectItem value="button">按钮</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                      <div>
                        <Label>资源标识 *</Label>
                        <Input
                          placeholder="如: /api/users, menu:users"
                          value={resourceForm.resource_id}
                          onChange={(e) => setResourceForm({ ...resourceForm, resource_id: e.target.value })}
                        />
                      </div>
                      <div>
                        <Label>资源名称 *</Label>
                        <Input
                          placeholder="如: 用户列表API"
                          value={resourceForm.resource_name}
                          onChange={(e) => setResourceForm({ ...resourceForm, resource_name: e.target.value })}
                        />
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button size="sm" onClick={handleAddResource} disabled={savingResource}>
                        {savingResource ? '保存中...' : '保存'}
                      </Button>
                      <Button size="sm" variant="outline" onClick={() => setShowResourceForm(false)}>
                        取消
                      </Button>
                    </div>
                  </div>
                )}

                {/* Resource List */}
                {resourceLoading ? (
                  <div className="text-center py-4 text-gray-500">加载中...</div>
                ) : resources.length === 0 ? (
                  <div className="text-center py-4 text-gray-500">暂无关联资源</div>
                ) : (
                  <div className="border rounded">
                    <table className="w-full">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-3 py-2 text-left text-sm font-medium text-gray-500">类型</th>
                          <th className="px-3 py-2 text-left text-sm font-medium text-gray-500">标识</th>
                          <th className="px-3 py-2 text-left text-sm font-medium text-gray-500">名称</th>
                          <th className="px-3 py-2 text-right text-sm font-medium text-gray-500">操作</th>
                        </tr>
                      </thead>
                      <tbody>
                        {resources.map((resource) => (
                          <tr key={resource.id} className="border-t">
                            <td className="px-3 py-2">{getResourceTypeBadge(resource.resource_type)}</td>
                            <td className="px-3 py-2 font-mono text-sm">{resource.resource_id}</td>
                            <td className="px-3 py-2">{resource.resource_name}</td>
                            <td className="px-3 py-2 text-right">
                              {!permission.is_system && (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleDeleteResource(resource.id)}
                                >
                                  <Trash2 className="h-4 w-4 text-red-500" />
                                </Button>
                              )}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>时间信息</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <div className="text-sm text-gray-500">创建时间</div>
                  <div>{new Date(permission.created_at).toLocaleString('zh-CN')}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-500">更新时间</div>
                  <div>{new Date(permission.updated_at).toLocaleString('zh-CN')}</div>
                </div>
              </CardContent>
            </Card>
          </div>
        </main>
      </div>
    </div>
  )
}
