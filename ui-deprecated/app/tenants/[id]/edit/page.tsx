'use client'

import { useState, useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'
import Link from 'next/link'
import { getTenant, updateTenant } from '@/lib/api/tenant'
import { TenantPlan, TenantStatus } from '@/types/tenant'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export default function EditTenantPage() {
  const router = useRouter()
  const params = useParams()
  const id = params.id as string

  const [loading, setLoading] = useState(false)
  const [fetching, setFetching] = useState(true)
  const [error, setError] = useState('')

  const [name, setName] = useState('')
  const [status, setStatus] = useState<TenantStatus>('active')
  const [plan, setPlan] = useState<TenantPlan>('free')

  useEffect(() => {
    const fetchTenant = async () => {
      try {
        const tenant = await getTenant(id)
        setName(tenant.name)
        setStatus(tenant.status)
        setPlan(tenant.plan)
      } catch (err) {
        setError(err instanceof Error ? err.message : '获取租户信息失败')
      } finally {
        setFetching(false)
      }
    }
    fetchTenant()
  }, [id])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!name.trim()) {
      setError('请输入租户名称')
      return
    }

    setLoading(true)
    try {
      await updateTenant(id, { name, status, plan })
      router.push(`/tenants/${id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : '更新失败')
    } finally {
      setLoading(false)
    }
  }

  if (fetching) {
    return (
      <div className="container mx-auto py-6">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6">
      <div className="max-w-2xl mx-auto">
        <div className="flex items-center gap-4 mb-6">
          <Link href={`/tenants/${id}`}>
            <Button variant="ghost" size="sm">← 返回</Button>
          </Link>
          <h1 className="text-2xl font-bold">编辑租户</h1>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>租户信息</CardTitle>
            <CardDescription>请修改租户信息</CardDescription>
          </CardHeader>
          <CardContent>
            {error && (
              <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
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
                <Label htmlFor="status">状态</Label>
                <Select value={status} onValueChange={(value) => setStatus(value as TenantStatus)}>
                  <SelectTrigger>
                    <SelectValue placeholder="选择状态" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">活跃</SelectItem>
                    <SelectItem value="suspended">已暂停</SelectItem>
                    <SelectItem value="deleted">已禁用</SelectItem>
                  </SelectContent>
                </Select>
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

              <div className="flex justify-end gap-2 pt-4">
                <Link href={`/tenants/${id}`}>
                  <Button type="button" variant="outline" disabled={loading}>
                    取消
                  </Button>
                </Link>
                <Button type="submit" disabled={loading}>
                  {loading ? '保存中...' : '保存'}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
