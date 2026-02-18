'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Save } from 'lucide-react'
import { createPermission } from '@/lib/api/permission'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'
import { useTenant } from '@/lib/tenant-context'

export default function NewPermissionPage() {
  const router = useRouter()
  const { user } = useAuthStore()

  // 使用统一的租户上下文
  const { tenants, selectedTenantId, setSelectedTenantId, tenantId, loading: tenantLoading } = useTenant()

  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const [formData, setFormData] = useState({
    name: '',
    description: '',
  })

  const handleSave = async () => {
    if (!formData.name) {
      setError('请填写必填字段')
      return
    }

    setSaving(true)
    setError('')
    try {
      await createPermission({
        tenant_id: tenantId,
        name: formData.name,
        description: formData.description,
      })
      router.push('/permissions')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create permission')
    } finally {
      setSaving(false)
    }
  }

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '权限管理', href: '/permissions' },
    { label: '新建权限' },
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

          <div className="flex justify-between items-center mb-6 mt-4">
            <h2 className="text-xl font-semibold">新建权限</h2>
            <div className="flex gap-2">
              <Button onClick={handleSave} disabled={saving}>
                <Save className="h-4 w-4 mr-1" />
                {saving ? '保存中...' : '保存'}
              </Button>
              <Link href="/permissions">
                <Button variant="outline">返回</Button>
              </Link>
            </div>
          </div>

          {error && (
            <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>
          )}

          <Card>
            <CardHeader>
              <CardTitle>权限信息</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Tenant Selector */}
              <div className="space-y-2">
                <Label htmlFor="tenant">租户 *</Label>
                <Select
                  value={selectedTenantId}
                  onValueChange={(value) => setSelectedTenantId(value)}
                  disabled={tenantLoading}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择租户" />
                  </SelectTrigger>
                  <SelectContent>
                    {tenants.map((tenant) => (
                      <SelectItem key={tenant.id} value={tenant.id}>
                        {tenant.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="name">权限名称 *</Label>
                <Input
                  id="name"
                  placeholder="如: 查看用户"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">描述</Label>
                <Input
                  id="description"
                  placeholder="可选描述"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                />
              </div>
            </CardContent>
          </Card>
        </main>
      </div>
    </div>
  )
}
