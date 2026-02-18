'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { createTenant } from '@/lib/api/tenant'
import { TenantPlan } from '@/types/tenant'
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

export default function NewTenantPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const [name, setName] = useState('')
  const [code, setCode] = useState('')
  const [plan, setPlan] = useState<TenantPlan>('free')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!name.trim()) {
      setError('请输入租户名称')
      return
    }
    if (!code.trim()) {
      setError('请输入租户代码')
      return
    }

    setLoading(true)
    try {
      await createTenant({ name, code, plan })
      router.push('/tenants')
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="container mx-auto py-6">
      <div className="max-w-2xl mx-auto">
        <div className="flex items-center gap-4 mb-6">
          <Link href="/tenants">
            <Button variant="ghost" size="sm">← 返回</Button>
          </Link>
          <h1 className="text-2xl font-bold">新建租户</h1>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>租户信息</CardTitle>
            <CardDescription>请填写租户基本信息</CardDescription>
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
                <Label htmlFor="code">租户代码 *</Label>
                <Input
                  id="code"
                  placeholder="请输入租户代码（唯一）"
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
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
              </div>

                </Select>
              <div className="flex justify-end gap-2 pt-4">
                <Link href="/tenants">
                  <Button type="button" variant="outline" disabled={loading}>
                    取消
                  </Button>
                </Link>
                <Button type="submit" disabled={loading}>
                  {loading ? '创建中...' : '创建'}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
