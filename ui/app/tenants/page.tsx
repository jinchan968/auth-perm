'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { listTenants, changeTenantStatus } from '@/lib/api/tenant'
import { TenantListItem, TenantStatus, TenantPlan } from '@/types/tenant'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { showError } from '@/lib/toast'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'
import { CreateTenantModal } from '@/components/tenants/create-tenant-modal'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export default function TenantsPage() {
  const { user } = useAuthStore()

  const [tenants, setTenants] = useState<TenantListItem[]>([])
  const [loading, setLoading] = useState(true)
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [size] = useState(10)
  const [modalOpen, setModalOpen] = useState(false)
  const [updatingStatusId, setUpdatingStatusId] = useState<string | null>(null)

  const fetchTenants = async () => {
    setLoading(true)
    try {
      const data = await listTenants({ keyword, page, size })
      setTenants(data.data)
      setTotal(data.total)
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Failed to fetch tenants')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTenants()
  }, [page])

  const handleSearch = () => {
    setPage(1)
    fetchTenants()
  }

  const handleStatusChange = async (tenantId: string, newStatus: TenantStatus) => {
    setUpdatingStatusId(tenantId)
    try {
      await changeTenantStatus(tenantId, newStatus)
      // 更新本地状态
      setTenants((prev) =>
        prev.map((t) => (t.id === tenantId ? { ...t, status: newStatus } : t))
      )
    } catch (err) {
      showError(err instanceof Error ? err.message : '状态更新失败')
    } finally {
      setUpdatingStatusId(null)
    }
  }

  const getPlanBadge = (plan: TenantPlan) => {
    const variants: Record<TenantPlan, 'default' | 'secondary' | 'outline'> = {
      free: 'outline',
      basic: 'secondary',
      pro: 'default',
      enterprise: 'default',
    }
    const labels: Record<TenantPlan, string> = {
      free: '免费版',
      basic: '基础版',
      pro: '专业版',
      enterprise: '企业版',
    }
    return <Badge variant={variants[plan]}>{labels[plan]}</Badge>
  }

  const totalPages = Math.ceil(total / size)

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '租户管理' },
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

          <div className="flex justify-between items-center mb-6 mt-4">
            <h2 className="text-xl font-semibold">租户列表</h2>
            <Button onClick={() => setModalOpen(true)}>新建租户</Button>
          </div>

          <Card>
            <CardContent className="pt-6">
              {/* Search */}
              <div className="flex gap-2 mb-4">
                <Input
                  placeholder="搜索租户名称或代码..."
                  value={keyword}
                  onChange={(e) => setKeyword(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                  className="max-w-xs"
                />
                <Button onClick={handleSearch}>搜索</Button>
              </div>

              {/* Table */}
              <div className="border rounded">
                <table className="w-full">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">名称</th>
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">代码</th>
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">状态</th>
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">套餐</th>
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">用户数</th>
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">创建时间</th>
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
                    ) : tenants.length === 0 ? (
                      <tr>
                        <td colSpan={7} className="px-4 py-8 text-center text-gray-500">
                          暂无数据
                        </td>
                      </tr>
                    ) : (
                      tenants.map((tenant) => (
                        <tr key={tenant.id} className="border-t">
                          <td className="px-4 py-2">{tenant.name}</td>
                          <td className="px-4 py-2">{tenant.code}</td>
                          <td className="px-4 py-2">
                            <Select
                              value={tenant.status}
                              onValueChange={(value: TenantStatus) =>
                                handleStatusChange(tenant.id, value)
                              }
                              disabled={updatingStatusId === tenant.id}
                            >
                              <SelectTrigger className="w-[110px]">
                                <SelectValue>
                                  <span className="flex items-center gap-2">
                                    <span className={`w-2 h-2 rounded-full ${
                                      tenant.status === 'active' ? 'bg-green-500' :
                                      tenant.status === 'suspended' ? 'bg-yellow-500' : 'bg-red-500'
                                    }`}></span>
                                    {tenant.status === 'active' ? '活跃' :
                                     tenant.status === 'suspended' ? '停用' : '禁用'}
                                  </span>
                                </SelectValue>
                              </SelectTrigger>
                              <SelectContent>
                                <SelectItem value="active">
                                  <span className="flex items-center gap-2">
                                    <span className="w-2 h-2 rounded-full bg-green-500"></span>
                                    活跃
                                  </span>
                                </SelectItem>
                                <SelectItem value="suspended">
                                  <span className="flex items-center gap-2">
                                    <span className="w-2 h-2 rounded-full bg-yellow-500"></span>
                                    停用
                                  </span>
                                </SelectItem>
                                <SelectItem value="deleted">
                                  <span className="flex items-center gap-2">
                                    <span className="w-2 h-2 rounded-full bg-red-500"></span>
                                    禁用
                                  </span>
                                </SelectItem>
                              </SelectContent>
                            </Select>
                          </td>
                          <td className="px-4 py-2">{getPlanBadge(tenant.plan)}</td>
                          <td className="px-4 py-2">{tenant.user_count}</td>
                          <td className="px-4 py-2">
                            {new Date(tenant.created_at).toLocaleDateString('zh-CN')}
                          </td>
                          <td className="px-4 py-2">
                            <Link href={`/tenants/${tenant.id}`}>
                              <Button variant="ghost" size="sm">查看</Button>
                            </Link>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

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

          <CreateTenantModal
            open={modalOpen}
            onOpenChange={setModalOpen}
            onSuccess={() => fetchTenants()}
          />
        </main>
      </div>
    </div>
  )
}
