'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { listTenants } from '@/lib/api/tenant'
import { TenantListItem, TenantStatus, TenantPlan } from '@/types/tenant'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export default function TenantsPage() {
  const [tenants, setTenants] = useState<TenantListItem[]>([])
  const [loading, setLoading] = useState(true)
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [size] = useState(10)
  const [error, setError] = useState('')

  const fetchTenants = async () => {
    setLoading(true)
    setError('')
    try {
      const data = await listTenants({ keyword, page, size })
      setTenants(data.data)
      setTotal(data.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch tenants')
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

  const getStatusBadge = (status: TenantStatus) => {
    const variants: Record<TenantStatus, 'default' | 'secondary' | 'destructive'> = {
      active: 'default',
      suspended: 'secondary',
      deleted: 'destructive',
    }
    const labels: Record<TenantStatus, string> = {
      active: '活跃',
      suspended: '已暂停',
      deleted: '已删除',
    }
    return <Badge variant={variants[status]}>{labels[status]}</Badge>
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

  return (
    <div className="container mx-auto py-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">租户管理</h1>
        <Link href="/tenants/new">
          <Button>新建租户</Button>
        </Link>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>租户列表</CardTitle>
        </CardHeader>
        <CardContent>
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

          {/* Error */}
          {error && (
            <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>
          )}

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
                      <td className="px-4 py-2">{getStatusBadge(tenant.status)}</td>
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
    </div>
  )
}
