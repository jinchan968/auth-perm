'use client'

import { useState } from 'react'
import { Loader2, Plus } from 'lucide-react'
import { createTenant } from '@/lib/api/tenant'
import { TenantPlan } from '@/types/tenant'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'
import { DetailActionBar } from '@/components/ui/detail-action-bar'
import { DetailPageHeader } from '@/components/ui/detail-page-header'
import { showError } from '@/lib/toast'
import { useNavigationTransition } from '@/components/providers/navigation-transition-provider'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export default function NewTenantPage() {
  const { user } = useAuthStore()
  const tenantsListHref = '/tenants'
  const { navigateWithTransition } = useNavigationTransition()

  const [loading, setLoading] = useState(false)

  const [name, setName] = useState('')
  const [plan, setPlan] = useState<TenantPlan>('free')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      showError('请输入租户名称')
      return
    }

    setLoading(true)
    try {
      await createTenant({ name, plan })
      navigateWithTransition(tenantsListHref)
    } catch (err) {
      showError(err instanceof Error ? err.message : '创建失败')
    } finally {
      setLoading(false)
    }
  }

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '租户管理', href: '/tenants' },
    { label: '新建租户' },
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
        <DashboardSidebar pathname="/tenants" />
        <main className="flex-1 p-8">
          <Breadcrumb items={breadcrumbItems} />

          <div className="max-w-2xl mx-auto mt-4">
            <DetailPageHeader
              title="新建租户"
              description="请填写租户基本信息"
              actions={
                <DetailActionBar returnHref={tenantsListHref} returnLabel="返回列表">
                  <Button type="submit" form="new-tenant-form" disabled={loading} className="min-w-[132px] active:scale-100">
                    {loading ? (
                      <>
                        <Loader2 className="h-4 w-4 animate-spin" />
                        创建中...
                      </>
                    ) : (
                      <>
                        <Plus className="h-4 w-4" />
                        创建
                      </>
                    )}
                  </Button>
                </DetailActionBar>
              }
            />
            <Card>
              <CardHeader>
                <CardTitle>租户信息</CardTitle>
              </CardHeader>
              <CardContent>
                <form id="new-tenant-form" onSubmit={handleSubmit} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="name">租户名称 *</Label>
                    <Input
                      id="name"
                      placeholder="请输入租户名称"
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      disabled={loading}
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="plan">套餐</Label>
                    <Select value={plan} onValueChange={(value) => setPlan(value as TenantPlan)}>
                      <SelectTrigger>
                        <SelectValue placeholder="选择套餐" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="free">免费版</SelectItem>
                        <SelectItem value="basic">基础版</SelectItem>
                        <SelectItem value="pro">专业版</SelectItem>
                        <SelectItem value="enterprise">企业版</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                </form>
              </CardContent>
            </Card>
          </div>
        </main>
      </div>
    </div>
  )
}
