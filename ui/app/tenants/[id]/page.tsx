'use client'

import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { Save, Edit2, X } from 'lucide-react'
import { getTenant, updateTenant, changeTenantStatus } from '@/lib/api/tenant'
import { Tenant, TenantStatus, TenantPlan, TenantSettings, FeaturesConfig, QuotaConfig } from '@/types/tenant'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { ShellLayout } from '@/components/layout/shell-layout'
import { useAuthStore } from '@/store/auth-store'
import { showError } from '@/lib/toast'
import { DetailActionBar } from '@/components/ui/detail-action-bar'
import { DetailPageHeader } from '@/components/ui/detail-page-header'
import { ListReturnButton } from '@/components/ui/list-return-button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export default function TenantDetailPage() {
  const params = useParams()
  const id = params.id as string
  const tenantsListHref = '/tenants'

  const { user } = useAuthStore()

  const [tenant, setTenant] = useState<Tenant | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadFailed, setLoadFailed] = useState(false)
  const [actionLoading, setActionLoading] = useState(false)

  // Edit mode state
  const [isEditing, setIsEditing] = useState(false)
  const [saving, setSaving] = useState(false)

  // Edit form data
  const [formData, setFormData] = useState({
    name: '',
    status: '' as TenantStatus | '',
    plan: '' as TenantPlan | '',
    expire_at: '',
    settings: null as TenantSettings | null,
  })

  useEffect(() => {
    const fetchTenant = async () => {
      setLoading(true)
      try {
        const data = await getTenant(id)
        setTenant(data)
      } catch (err) {
        showError(err instanceof Error ? err.message : 'Failed to fetch tenant')
        setLoadFailed(true)
      } finally {
        setLoading(false)
      }
    }
    fetchTenant()
  }, [id])

  // Initialize form data when entering edit mode
  useEffect(() => {
    if (isEditing && tenant) {
      setFormData({
        name: tenant.name,
        status: tenant.status,
        plan: tenant.plan,
        expire_at: tenant.expire_at || '',
        settings: { ...tenant.settings },
      })
    }
  }, [isEditing, tenant])

  const handleToggleEdit = () => {
    if (!isEditing) {
      // Enter edit mode - initialize form data
      if (tenant) {
        setFormData({
          name: tenant.name,
          status: tenant.status,
          plan: tenant.plan,
          expire_at: tenant.expire_at || '',
          settings: { ...tenant.settings },
        })
      }
    }
    setIsEditing(!isEditing)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const request: any = {}
      if (formData.name !== tenant?.name) request.name = formData.name
      if (formData.status !== tenant?.status) request.status = formData.status
      if (formData.plan !== tenant?.plan) request.plan = formData.plan
      if (formData.expire_at !== (tenant?.expire_at || '')) request.expire_at = formData.expire_at || undefined
      if (formData.settings) request.settings = formData.settings

      const updated = await updateTenant(id, request)
      setTenant(updated)
      setIsEditing(false)
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Failed to save tenant')
    } finally {
      setSaving(false)
    }
  }

  const handleFeatureChange = (key: keyof FeaturesConfig, value: boolean) => {
    if (!formData.settings) return
    setFormData({
      ...formData,
      settings: {
        ...formData.settings,
        features: {
          ...formData.settings.features,
          [key]: value,
        },
      },
    })
  }

  const handleQuotaChange = (key: keyof QuotaConfig, value: string) => {
    if (!formData.settings) return
    const numValue = value === '' ? 0 : parseInt(value, 10)
    setFormData({
      ...formData,
      settings: {
        ...formData.settings,
        quota: {
          ...formData.settings.quota,
          [key]: isNaN(numValue) ? 0 : numValue,
        },
      },
    })
  }


  const handleSuspend = async () => {
    if (!confirm('确定要禁用此租户吗？')) return
    setActionLoading(true)
    setLoadFailed(false)
    try {
      await changeTenantStatus(id, 'suspended')
      // Refresh tenant data
      const data = await getTenant(id)
      setTenant(data)
    } catch (err) {
      showError(err instanceof Error ? err.message : '禁用租户失败')
    } finally {
      setActionLoading(false)
    }
  }

  const handleActivate = async () => {
    setActionLoading(true)
    setLoadFailed(false)
    try {
      await changeTenantStatus(id, 'active')
      // Refresh tenant data
      const data = await getTenant(id)
      setTenant(data)
    } catch (err) {
      showError(err instanceof Error ? err.message : '启用租户失败')
    } finally {
      setActionLoading(false)
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

  const featureLabels: Record<keyof FeaturesConfig, string> = {
    email_verification: '邮箱验证',
    oauth_login: 'OAuth登录',
    totp_enabled: 'TOTP两步验证',
    session_limit: '会话限制',
    password_complexity: '密码复杂度',
  }

  const quotaLabels: Record<keyof QuotaConfig, string> = {
    max_users: '最大用户数',
    max_roles: '最大角色数',
    max_organizations: '最大组织数',
    max_sessions_per_user: '最大会话数',
    api_rate_limit: 'API速率限制',
  }

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '租户管理', href: '/tenants' },
    { label: tenant?.name || '租户详情' },
  ]

  return (
    <ShellLayout pathname="/tenants/[id]">
      {loading ? (
        <div className="text-center">加载中...</div>
      ) : loadFailed && !tenant ? (
        <ListReturnButton href={tenantsListHref} label="返回列表" />
      ) : !tenant ? (
        <>
          <div className="text-center">租户不存在</div>
          <ListReturnButton href={tenantsListHref} label="返回列表" />
        </>
      ) : (
        <>
          <Breadcrumb items={breadcrumbItems} />

          <DetailPageHeader
            title={tenant.name}
            actions={
              <DetailActionBar returnHref={tenantsListHref} returnLabel="返回">
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
                ) : (
                  <Button onClick={handleToggleEdit}>
                    <Edit2 className="h-4 w-4 mr-1" />
                    编辑
                  </Button>
                )}
                {tenant.status === 'active' && (
                  <Button variant="outline" onClick={handleSuspend} disabled={actionLoading}>
                    {actionLoading ? '禁用中...' : '禁用'}
                  </Button>
                )}
                {tenant.status === 'deleted' && (
                  <Button variant="default" onClick={handleActivate} disabled={actionLoading}>
                    {actionLoading ? '启用中...' : '启用'}
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
                  <div className="font-mono text-sm">{tenant.id}</div>
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
                      <div>{tenant.code}</div>
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="status">状态</Label>
                      <Select
                        value={formData.status}
                        onValueChange={(value) => setFormData({ ...formData, status: value as TenantStatus })}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="active">活跃</SelectItem>
                          <SelectItem value="suspended">已暂停</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="plan">套餐</Label>
                      <Select
                        value={formData.plan}
                        onValueChange={(value) => setFormData({ ...formData, plan: value as TenantPlan })}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="free">免费版</SelectItem>
                          <SelectItem value="basic">基础版</SelectItem>
                          <SelectItem value="pro">专业版</SelectItem>
                          <SelectItem value="enterprise">企业版</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="expire_at">过期时间</Label>
                      <Input
                        id="expire_at"
                        type="date"
                        value={formData.expire_at ? formData.expire_at.split('T')[0] : ''}
                        onChange={(e) => setFormData({ ...formData, expire_at: e.target.value })}
                      />
                    </div>
                  </>
                ) : (
                  // View mode
                  <>
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
                  </>
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

            {/* 功能开关 */}
            <Card className="md:col-span-2">
              <CardHeader>
                <CardTitle>功能开关</CardTitle>
              </CardHeader>
              <CardContent>
                {isEditing && formData.settings ? (
                  // Edit mode
                  <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                    {(Object.keys(formData.settings.features) as Array<keyof FeaturesConfig>).map((key) => (
                      <div key={key} className="flex items-center justify-between p-3 border rounded-lg">
                        <Label htmlFor={key} className="text-sm cursor-pointer">
                          {featureLabels[key]}
                        </Label>
                        <Switch
                          id={key}
                          checked={formData.settings!.features[key]}
                          onCheckedChange={(checked) => handleFeatureChange(key, checked)}
                        />
                      </div>
                    ))}
                  </div>
                ) : (
                  // View mode
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
                )}
              </CardContent>
            </Card>

            {/* 配额限制 */}
            <Card className="md:col-span-2">
              <CardHeader>
                <CardTitle>配额限制</CardTitle>
              </CardHeader>
              <CardContent>
                {isEditing && formData.settings ? (
                  // Edit mode
                  <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                    {(Object.keys(formData.settings.quota) as Array<keyof QuotaConfig>).map((key) => (
                      <div key={key} className="space-y-2">
                        <Label htmlFor={key} className="text-xs text-gray-500">
                          {quotaLabels[key]}
                        </Label>
                        <Input
                          id={key}
                          type="number"
                          value={formData.settings!.quota[key]}
                          onChange={(e) => handleQuotaChange(key, e.target.value)}
                          className="h-9"
                        />
                      </div>
                    ))}
                  </div>
                ) : (
                  // View mode
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
                )}
              </CardContent>
            </Card>
          </div>
        </>
      )}
    </ShellLayout>
  )
}
