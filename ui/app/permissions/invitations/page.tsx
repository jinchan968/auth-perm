'use client'

import { useState, useEffect, useCallback } from 'react'
import { Plus, Copy, Check, Clock, UserCheck, Ban, AlertCircle } from 'lucide-react'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { PermGuard } from '@/components/ui/perm-guard'
import { ShellLayout } from '@/components/layout/shell-layout'
import { useTenant } from '@/lib/tenant-context'
import { showError, showSuccess } from '@/lib/toast'
import { listInvitations, createInvitation, invalidateInvitation } from '@/lib/api/invitation'
import type { InvitationItem, CreateInvitationResponse } from '@/types/invitation'

export default function InvitationsPage() {
  const { tenantId } = useTenant()
  const [invitations, setInvitations] = useState<InvitationItem[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [creating, setCreating] = useState(false)
  const [createdResult, setCreatedResult] = useState<CreateInvitationResponse | null>(null)
  const [copiedField, setCopiedField] = useState<'code' | 'url' | null>(null)

  const pageSize = 10

  const fetchInvitations = useCallback(async () => {
    if (!tenantId) return
    setLoading(true)
    try {
      const result = await listInvitations({ tenant_id: tenantId, status: statusFilter || undefined, page, size: pageSize })
      setInvitations(result.data || [])
      setTotal(result.total || 0)
    } catch {
      showError('获取邀请码列表失败')
    } finally {
      setLoading(false)
    }
  }, [tenantId, statusFilter, page])

  useEffect(() => {
    fetchInvitations()
  }, [fetchInvitations])

  const handleCreate = async () => {
    if (!tenantId) {
      showError('请先选择租户')
      return
    }
    setCreating(true)
    try {
      const result = await createInvitation({ tenant_id: tenantId })
      setCreatedResult(result)
    } catch {
      showError('创建邀请码失败')
    } finally {
      setCreating(false)
    }
  }

  const handleInvalidate = async (id: string) => {
    try {
      await invalidateInvitation(id)
      showSuccess('邀请码已失效')
      fetchInvitations()
    } catch {
      showError('操作失败')
    }
  }

  const handleCopy = async (text: string, field: 'code' | 'url') => {
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      const ta = document.createElement('textarea')
      ta.value = text
      ta.style.position = 'fixed'
      ta.style.opacity = '0'
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
    }
    setCopiedField(field)
    setTimeout(() => setCopiedField(null), 2000)
  }

  const handleCloseModal = () => {
    setCreateModalOpen(false)
    setCreatedResult(null)
    fetchInvitations()
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge variant="default">有效</Badge>
      case 'used':
        return <Badge variant="secondary">已使用</Badge>
      case 'invalidated':
        return <Badge variant="destructive">已失效</Badge>
      case 'expired':
        return <Badge variant="outline" className="text-amber-600 border-amber-300">已过期</Badge>
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active': return <Clock className="h-4 w-4 text-green-500" />
      case 'used': return <UserCheck className="h-4 w-4 text-blue-500" />
      case 'invalidated': return <Ban className="h-4 w-4 text-red-500" />
      case 'expired': return <AlertCircle className="h-4 w-4 text-amber-500" />
      default: return <Clock className="h-4 w-4 text-slate-400" />
    }
  }

  const totalPages = Math.ceil(total / pageSize)

  return (
    <ShellLayout pathname="/permissions/invitations">
      <Breadcrumb
        items={[
          { label: '首页', href: '/home' },
          { label: '权限管理', href: '/permissions' },
          { label: '邀请码' },
        ]}
      />

      <div className="flex flex-wrap items-center justify-between gap-3 mb-6 mt-4">
        <h2 className="text-xl font-semibold">注册邀请码</h2>
        <PermGuard button="invitations.create">
          <Button onClick={() => setCreateModalOpen(true)}>
            <Plus className="h-4 w-4 mr-1" />
            生成邀请码
          </Button>
        </PermGuard>
      </div>

      {/* Status filter */}
      <div className="flex flex-wrap gap-2 mb-4">
        {['', 'active', 'used', 'invalidated', 'expired'].map((s) => (
          <Button
            key={s}
            variant={statusFilter === s ? 'default' : 'outline'}
            size="sm"
            onClick={() => { setStatusFilter(s); setPage(1) }}
          >
            {s === '' ? '全部' : s === 'active' ? '有效' : s === 'used' ? '已使用' : s === 'invalidated' ? '已失效' : '已过期'}
          </Button>
        ))}
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-slate-200 border-t-primary" />
        </div>
      ) : invitations.length === 0 ? (
        <div className="text-center py-12 text-slate-400">
          <Clock className="h-12 w-12 mx-auto mb-3 opacity-50" />
          <p>暂无邀请码</p>
        </div>
      ) : (
        <div className="space-y-3">
          {invitations.map((inv) => (
            <Card key={inv.id}>
              <CardContent className="py-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    {getStatusIcon(inv.status)}
                    <div>
                      <p className="text-sm font-mono">{inv.code_preview}</p>
                      <p className="text-xs text-slate-400">
                        有效期至 {new Date(inv.expires_at).toLocaleString()}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {getStatusBadge(inv.status)}
                    {inv.status === 'active' && (
                      <PermGuard button="invitations.invalidate">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleInvalidate(inv.id)}
                        >
                          失效
                        </Button>
                      </PermGuard>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 pt-4">
              <Button
                size="sm"
                variant="outline"
                disabled={page <= 1}
                onClick={() => setPage(page - 1)}
              >
                上一页
              </Button>
              <span className="text-sm text-slate-500">{page} / {totalPages}</span>
              <Button
                size="sm"
                variant="outline"
                disabled={page >= totalPages}
                onClick={() => setPage(page + 1)}
              >
                下一页
              </Button>
            </div>
          )}
        </div>
      )}

      {/* Create modal */}
      <Dialog open={createModalOpen} onOpenChange={(open) => { if (!open) handleCloseModal() }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>生成邀请码</DialogTitle>
          </DialogHeader>

          {createdResult ? (
            <div className="space-y-3">
              <div className="rounded-lg bg-green-50 border border-green-200 p-4">
                <p className="text-sm text-green-700 mb-3">
                  邀请码已生成，请立即复制保存。此信息仅在本次展示。
                </p>
                <div className="space-y-2">
                  <div>
                    <Label className="text-xs">邀请码</Label>
                    <div className="flex gap-2 mt-1">
                      <Input
                        readOnly
                        value={createdResult.invite_code}
                        className="font-mono text-sm"
                      />
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleCopy(createdResult.invite_code, 'code')}
                      >
                        {copiedField === 'code' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                      </Button>
                    </div>
                  </div>
                  <div>
                    <Label className="text-xs">注册链接</Label>
                    <div className="flex gap-2 mt-1">
                      <Input
                        readOnly
                        value={createdResult.invite_url}
                        className="text-sm"
                      />
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleCopy(createdResult.invite_url, 'url')}
                      >
                        {copiedField === 'url' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
              <Button onClick={handleCloseModal} className="w-full">
                确认并关闭
              </Button>
            </div>
          ) : (
            <div className="space-y-4">
              <p className="text-sm text-slate-500">
                将为当前租户生成一个注册邀请码，有效期 7 天。
              </p>
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={handleCloseModal}>
                  取消
                </Button>
                <Button onClick={handleCreate} disabled={creating}>
                  {creating ? '生成中...' : '确认生成'}
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </ShellLayout>
  )
}
