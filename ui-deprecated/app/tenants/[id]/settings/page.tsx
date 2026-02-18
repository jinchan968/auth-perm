'use client'

import { useState, useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'
import Link from 'next/link'
import { getTenantSettings, updateTenantSettings } from '@/lib/api/tenant'
import { FeaturesConfig, QuotaConfig } from '@/types/tenant'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Label } from '@/components/ui/label'

export default function TenantSettingsPage() {
  const router = useRouter()
  const params = useParams()
  const id = params.id as string

  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const [features, setFeatures] = useState<FeaturesConfig>({
    email_verification: true,
    oauth_login: true,
    totp_enabled: true,
    session_limit: true,
    password_complexity: true,
  })

  const [quota, setQuota] = useState<QuotaConfig>({
    max_users: -1,
    max_roles: 100,
    max_organizations: 50,
    max_sessions_per_user: 5,
    api_rate_limit: 1000,
  })

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const settings = await getTenantSettings(id)
        setFeatures(settings.features)
        setQuota(settings.quota)
      } catch (err) {
        setError(err instanceof Error ? err.message : '获取设置失败')
      } finally {
        setLoading(false)
      }
    }
    fetchSettings()
  }, [id])

  const handleSave = async () => {
    setError('')
    setSuccess('')
    setSaving(true)

    try {
      await updateTenantSettings(id, { settings: { features, quota } })
      setSuccess('设置已保存')
      router.refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const toggleFeature = (key: keyof FeaturesConfig) => {
    setFeatures({ ...features, [key]: !features[key] })
  }

  const updateQuota = (key: keyof QuotaConfig, value: string) => {
    const numValue = value === '' ? -1 : parseInt(value, 10)
    if (isNaN(numValue)) return
    setQuota({ ...quota, [key]: numValue })
  }

  if (loading) {
    return (
      <div className="container mx-auto py-6">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6">
      <div className="max-w-4xl mx-auto">
        <div className="flex items-center gap-4 mb-6">
          <Link href={`/tenants/${id}`}>
            <Button variant="ghost" size="sm">← 返回</Button>
          </Link>
          <h1 className="text-2xl font-bold">租户设置</h1>
        </div>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded mb-4">{error}</div>
        )}
        {success && (
          <div className="bg-green-50 text-green-600 p-3 rounded mb-4">{success}</div>
        )}

        <div className="space-y-6">
          {/* Features */}
          <Card>
            <CardHeader>
              <CardTitle>功能开关</CardTitle>
              <CardDescription>控制租户可以使用的功能</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2">
                <div className="flex items-center justify-between p-3 border rounded">
                  <div>
                    <div className="font-medium">邮箱验证</div>
                    <div className="text-sm text-gray-500">要求用户验证邮箱</div>
                  </div>
                  <button
                    onClick={() => toggleFeature('email_verification')}
                    className={`w-12 h-6 rounded-full transition-colors ${
                      features.email_verification ? 'bg-blue-600' : 'bg-gray-300'
                    }`}
                  >
                    <div
                      className={`w-5 h-5 bg-white rounded-full transition-transform ${
                        features.email_verification ? 'translate-x-6' : 'translate-x-0.5'
                      }`}
                    />
                  </button>
                </div>

                <div className="flex items-center justify-between p-3 border rounded">
                  <div>
                    <div className="font-medium">OAuth登录</div>
                    <div className="text-sm text-gray-500">允许第三方OAuth登录</div>
                  </div>
                  <button
                    onClick={() => toggleFeature('oauth_login')}
                    className={`w-12 h-6 rounded-full transition-colors ${
                      features.oauth_login ? 'bg-blue-600' : 'bg-gray-300'
                    }`}
                  >
                    <div
                      className={`w-5 h-5 bg-white rounded-full transition-transform ${
                        features.oauth_login ? 'translate-x-6' : 'translate-x-0.5'
                      }`}
                    />
                  </button>
                </div>

                <div className="flex items-center justify-between p-3 border rounded">
                  <div>
                    <div className="font-medium">TOTP双因子</div>
                    <div className="text-sm text-gray-500">启用TOTP认证</div>
                  </div>
                  <button
                    onClick={() => toggleFeature('totp_enabled')}
                    className={`w-12 h-6 rounded-full transition-colors ${
                      features.totp_enabled ? 'bg-blue-600' : 'bg-gray-300'
                    }`}
                  >
                    <div
                      className={`w-5 h-5 bg-white rounded-full transition-transform ${
                        features.totp_enabled ? 'translate-x-6' : 'translate-x-0.5'
                      }`}
                    />
                  </button>
                </div>

                <div className="flex items-center justify-between p-3 border rounded">
                  <div>
                    <div className="font-medium">会话限制</div>
                    <div className="text-sm text-gray-500">限制用户会话数量</div>
                  </div>
                  <button
                    onClick={() => toggleFeature('session_limit')}
                    className={`w-12 h-6 rounded-full transition-colors ${
                      features.session_limit ? 'bg-blue-600' : 'bg-gray-300'
                    }`}
                  >
                    <div
                      className={`w-5 h-5 bg-white rounded-full transition-transform ${
                        features.session_limit ? 'translate-x-6' : 'translate-x-0.5'
                      }`}
                    />
                  </button>
                </div>

                <div className="flex items-center justify-between p-3 border rounded">
                  <div>
                    <div className="font-medium">密码复杂度</div>
                    <div className="text-sm text-gray-500">强制密码复杂度要求</div>
                  </div>
                  <button
                    onClick={() => toggleFeature('password_complexity')}
                    className={`w-12 h-6 rounded-full transition-colors ${
                      features.password_complexity ? 'bg-blue-600' : 'bg-gray-300'
                    }`}
                  >
                    <div
                      className={`w-5 h-5 bg-white rounded-full transition-transform ${
                        features.password_complexity ? 'translate-x-6' : 'translate-x-0.5'
                      }`}
                    />
                  </button>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Quota */}
          <Card>
            <CardHeader>
              <CardTitle>配额限制</CardTitle>
              <CardDescription>设置租户的资源配额限制（-1表示无限制）</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="max_users">最大用户数</Label>
                  <Input
                    id="max_users"
                    type="number"
                    value={quota.max_users === -1 ? '' : quota.max_users}
                    onChange={(e) => updateQuota('max_users', e.target.value)}
                    placeholder="-1 表示无限制"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="max_roles">最大角色数</Label>
                  <Input
                    id="max_roles"
                    type="number"
                    value={quota.max_roles}
                    onChange={(e) => updateQuota('max_roles', e.target.value)}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="max_organizations">最大组织数</Label>
                  <Input
                    id="max_organizations"
                    type="number"
                    value={quota.max_organizations}
                    onChange={(e) => updateQuota('max_organizations', e.target.value)}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="max_sessions_per_user">每用户最大会话数</Label>
                  <Input
                    id="max_sessions_per_user"
                    type="number"
                    value={quota.max_sessions_per_user}
                    onChange={(e) => updateQuota('max_sessions_per_user', e.target.value)}
                  />
                </div>

                <div className="space-y-2 md:col-span-2">
                  <Label htmlFor="api_rate_limit">API速率限制（次/分钟）</Label>
                  <Input
                    id="api_rate_limit"
                    type="number"
                    value={quota.api_rate_limit}
                    onChange={(e) => updateQuota('api_rate_limit', e.target.value)}
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          <div className="flex justify-end gap-2">
            <Link href={`/tenants/${id}`}>
              <Button variant="outline">取消</Button>
            </Link>
            <Button onClick={handleSave} disabled={saving}>
              {saving ? '保存中...' : '保存设置'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
