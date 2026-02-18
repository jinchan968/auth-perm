'use client'

import { useState, useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'
import Link from 'next/link'
import { getTenant, deleteTenant } from '@/lib/api/tenant'
import { Tenant, TenantStatus, TenantPlan } from '@/types/tenant'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export default function TenantDetailPage() {
  const router = useRouter()
  const params = useParams()
  const id = params.id as string

  const [tenant, setTenant] = useState<Tenant | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    const fetchTenant = async () => {
      setLoading(true)
      try {
        const data = await getTenant(id)
        setTenant(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch tenant')
      } finally {
        setLoading(false)
      }
    }
    fetchTenant()
  }, [id])

  const handleDelete = async () => {
    if (!confirm('确定要删除此租户吗？')) return

    setDeleting(true)
    try {
      await deleteTenant(id)
      router.push('/tenants')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete tenant')
    } finally {
      setDeleting(false)
    }
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

  if (loading) {
    return (
      <div className="container mx-auto py-6">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container mx-auto py-6">
        <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>
        <Link href="/tenants">
          <Button variant="outline">返回列表</Button>
        </Link>
      </div>
    )
  }

  if (!tenant) {
    return (
      <div className="container mx-auto py-6">
        <div className="text-center">租户不存在</div>
        <Link href="/tenants">
          <Button variant="outline">返回列表</Button>
        </Link>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6">
      <div className="flex justify-between items-center mb-6">
        <div className="flex items-center gap-4">
          <Link href="/tenants">
            <Button variant="ghost" size="sm">← 返回</Button>
          </Link>
          <h1 className="text-2xl font-bold">{tenant.name}</h1>
        </div>
        <div className="flex gap-2">
          <Link href={`/tenants/${id}/settings`}>
            <Button variant="outline">设置</Button>
          </Link>
          <Link href={`/tenants/${id}/edit`}>
            <Button variant="outline">编辑</Button>
          </Link>
          <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
            {deleting ? '删除中...' : '删除'}
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>基本信息</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <div className="text-sm text-gray-500">ID</div>
              <div className="font-mono text-sm">{tenant.id}</div>
            </div>
            <div>
              <div className="text-sm text-gray-500">名称</div>
              <div>{tenant.name}</div>
            </div>
            <div>
              <div className="text-sm text-gray-500">代码</div>
              <div>{tenant.code}</div>
            </div>
            <div>
              <div className="text-sm text-gray-500">状态</div>
              <div>{getStatusBadge(tenant.status)}</div>
            </div>
            <div>
              <div className="text-sm text-gray-500">套餐</div>
              <div>{getPlanBadge(tenant.plan)}</div>
            </div>
            {tenant.expire_at && (
              <div>
                <div className="text-sm text-gray-500">过期时间</div>
                <div>{new Date(tenant.expire_at).toLocaleDateString('zh-CN')}</div>
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
              <div>{new Date(tenant.created_at).toLocaleString('zh-CN')}</div>
            </div>
            <div>
              <div className="text-sm text-gray-500">更新时间</div>
              <div>{new Date(tenant.updated_at).toLocaleString('zh-CN')}</div>
            </div>
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle>功能开关</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
              <div className="flex items-center justify-between p-2 border rounded">
                <span className="text-sm">邮箱验证</span>
                <span>{tenant.settings.features.email_verification ? '✓' : '✗'}</span>
              </div>
              <div className="flex items-center justify-between p-2 border rounded">
                <span className="text-sm">OAuth登录</span>
                <span>{tenant.settings.features.oauth_login ? '✓' : '✗'}</span>
              </div>
              <div className="flex items-center justify-between p-2 border rounded">
                <span className="text-sm">TOTP</span>
                <span>{tenant.settings.features.totp_enabled ? '✓' : '✗'}</span>
              </div>
              <div className="flex items-center justify-between p-2 border rounded">
                <span className="text-sm">会话限制</span>
                <span>{tenant.settings.features.session_limit ? '✓' : '✗'}</span>
              </div>
              <div className="flex items-center justify-between p-2 border rounded">
                <span className="text-sm">密码复杂度</span>
                <span>{tenant.settings.features.password_complexity ? '✓' : '✗'}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle>配额限制</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
              <div className="p-2 border rounded text-center">
                <div className="text-2xl font-bold">
                  {tenant.settings.quota.max_users === -1 ? '∞' : tenant.settings.quota.max_users}
                </div>
                <div className="text-sm text-gray-500">最大用户数</div>
              </div>
              <div className="p-2 border rounded text-center">
                <div className="text-2xl font-bold">{tenant.settings.quota.max_roles}</div>
                <div className="text-sm text-gray-500">最大角色数</div>
              </div>
              <div className="p-2 border rounded text-center">
                <div className="text-2xl font-bold">{tenant.settings.quota.max_organizations}</div>
                <div className="text-sm text-gray-500">最大组织数</div>
              </div>
              <div className="p-2 border rounded text-center">
                <div className="text-2xl font-bold">{tenant.settings.quota.max_sessions_per_user}</div>
                <div className="text-sm text-gray-500">最大会话数</div>
              </div>
              <div className="p-2 border rounded text-center">
                <div className="text-2xl font-bold">{tenant.settings.quota.api_rate_limit}</div>
                <div className="text-sm text-gray-500">API速率限制</div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
