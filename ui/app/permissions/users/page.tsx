'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Plus, Search } from 'lucide-react'
import { listUsers, createUser, updateUserStatus } from '@/lib/api/user'
import { AccountListItem, AccountStatus, AccountType } from '@/types/user'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { useTenant } from '@/lib/tenant-context'
import { ShellLayout } from '@/components/layout/shell-layout'
import { showError } from '@/lib/toast'

export default function UsersPage() {
  const { selectedTenantId, tenantId, loading: tenantLoading } = useTenant()

  const [users, setUsers] = useState<AccountListItem[]>([])
  const [loading, setLoading] = useState(true)
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [size] = useState(10)

  // 创建用户对话框
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [creating, setCreating] = useState(false)
  const [createForm, setCreateForm] = useState({
    identifier_type: 'email' as 'email' | 'phone',
    email: '',
    phone: '',
    username: '',
    password: '',
    confirm_password: '',
    nickname: '',
  })

  const fetchUsers = async () => {
    if (!selectedTenantId) return
    setLoading(true)
    try {
      const data = await listUsers({
        tenant_id: selectedTenantId,
        keyword,
        page,
        page_size: size,
      })
      setUsers(data.data || [])
      setTotal(data.total)
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Failed to fetch users')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (selectedTenantId) {
      fetchUsers()
    }
  }, [selectedTenantId, page])

  const handleSearch = () => {
    setPage(1)
    fetchUsers()
  }

  const handleStatusChange = async (accountId: string, newStatus: AccountStatus) => {
    try {
      await updateUserStatus(accountId, {
        tenant_id: selectedTenantId,
        status: newStatus,
      })
      // 更新本地状态
      setUsers((prev) =>
        prev.map((u) =>
          u.account_id === accountId ? { ...u, account_status: newStatus } : u
        )
      )
    } catch (err) {
      showError(err instanceof Error ? err.message : '状态更新失败')
    }
  }

  const handleCreateUser = async () => {
    // 验证
    if (!createForm.username || !createForm.password || !createForm.confirm_password) {
      showError('请填写必填项')
      return
    }
    if (createForm.password !== createForm.confirm_password) {
      showError('两次密码输入不一致')
      return
    }
    if (createForm.identifier_type === 'email' && !createForm.email) {
      showError('请输入邮箱')
      return
    }
    if (createForm.identifier_type === 'phone' && !createForm.phone) {
      showError('请输入手机号')
      return
    }

    setCreating(true)
    try {
      await createUser({
        ...createForm,
        tenant_id: selectedTenantId,
      })
      setCreateDialogOpen(false)
      setCreateForm({
        identifier_type: 'email',
        email: '',
        phone: '',
        username: '',
        password: '',
        confirm_password: '',
        nickname: '',
      })
      fetchUsers()
    } catch (err) {
      showError(err instanceof Error ? err.message : '创建用户失败')
    } finally {
      setCreating(false)
    }
  }

  const getStatusBadge = (status: AccountStatus) => {
    const variants: Record<AccountStatus, 'default' | 'secondary' | 'outline'> = {
      active: 'default',
      inactive: 'secondary',
      suspended: 'outline',
    }
    const labels: Record<AccountStatus, string> = {
      active: '活跃',
      inactive: '停用',
      suspended: '暂停',
    }
    return <Badge variant={variants[status]}>{labels[status]}</Badge>
  }

  const getAccountTypeBadge = (type: AccountType) => {
    const labels: Record<AccountType, string> = {
      email: '邮箱',
      phone: '手机',
      github: 'GitHub',
      google: 'Google',
      wechat: '微信',
    }
    return <Badge variant="outline">{labels[type] || type}</Badge>
  }

  const totalPages = Math.ceil(total / size)

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '权限管理', href: '/permissions' },
    { label: '用户管理' },
  ]

  return (
    <ShellLayout pathname="/permissions">
          <Breadcrumb items={breadcrumbItems} />

          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mb-6 mt-4">
            <h2 className="text-xl font-semibold">用户列表</h2>
            <Button onClick={() => setCreateDialogOpen(true)}>
              <Plus className="h-4 w-4 mr-1" />
              新建用户
            </Button>
          </div>

          <Card>
            <CardContent className="pt-6">
              {/* Search */}
              <div className="flex flex-wrap gap-2 mb-4">
                <Input
                  placeholder="搜索用户名、邮箱或手机号..."
                  value={keyword}
                  onChange={(e) => setKeyword(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                  className="w-full sm:w-auto sm:max-w-xs"
                />
                <Button onClick={handleSearch}>
                  <Search className="h-4 w-4 mr-1" />
                  搜索
                </Button>
              </div>

              {/* Table */}
              <div className="border rounded overflow-x-auto">
                <table className="w-full min-w-[700px]">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">用户名</th>
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">邮箱</th>
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">账户类型</th>
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">状态</th>
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-500">最后登录</th>
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
                    ) : users.length === 0 ? (
                      <tr>
                        <td colSpan={7} className="px-4 py-8 text-center text-gray-500">
                          暂无数据
                        </td>
                      </tr>
                    ) : (
                      users.map((user) => (
                        <tr key={user.account_id} className="border-t hover:bg-gray-50">
                          <td className="px-4 py-2">
                            <div className="font-medium">{user.username}</div>
                            {user.nickname && (
                              <div className="text-xs text-gray-400">{user.nickname}</div>
                            )}
                          </td>
                          <td className="px-4 py-2 text-sm">{user.email || user.phone || '-'}</td>
                          <td className="px-4 py-2">{getAccountTypeBadge(user.account_type)}</td>
                          <td className="px-4 py-2">
                            <Select
                              value={user.account_status}
                              onValueChange={(value: AccountStatus) =>
                                handleStatusChange(user.account_id, value)
                              }
                            >
                              <SelectTrigger className="w-[100px]">
                                <SelectValue>{getStatusBadge(user.account_status)}</SelectValue>
                              </SelectTrigger>
                              <SelectContent>
                                <SelectItem value="active">活跃</SelectItem>
                                <SelectItem value="inactive">停用</SelectItem>
                                <SelectItem value="suspended">暂停</SelectItem>
                              </SelectContent>
                            </Select>
                          </td>
                          <td className="px-4 py-2 text-sm">
                            {user.last_login_at
                              ? new Date(user.last_login_at).toLocaleString('zh-CN')
                              : '从未登录'}
                          </td>
                          <td className="px-4 py-2 text-sm">
                            {new Date(user.created_at).toLocaleDateString('zh-CN')}
                          </td>
                          <td className="px-4 py-2">
                            <Link href={`/permissions/users/${user.account_id}`}>
                              <Button variant="ghost" size="sm">详情</Button>
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

          {/* Create User Dialog */}
          <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>新建用户</DialogTitle>
              </DialogHeader>
              <div className="space-y-4">
                <div>
                  <Label>标识符类型 *</Label>
                  <Select
                    value={createForm.identifier_type}
                    onValueChange={(value: 'email' | 'phone') =>
                      setCreateForm({ ...createForm, identifier_type: value })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="email">邮箱</SelectItem>
                      <SelectItem value="phone">手机号</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {createForm.identifier_type === 'email' ? (
                  <div>
                    <Label>邮箱 *</Label>
                    <Input
                      type="email"
                      value={createForm.email}
                      onChange={(e) => setCreateForm({ ...createForm, email: e.target.value })}
                      placeholder="user@example.com"
                    />
                  </div>
                ) : (
                  <div>
                    <Label>手机号 *</Label>
                    <Input
                      type="tel"
                      value={createForm.phone}
                      onChange={(e) => setCreateForm({ ...createForm, phone: e.target.value })}
                      placeholder="13800138000"
                    />
                  </div>
                )}

                <div>
                  <Label>用户名 *</Label>
                  <Input
                    value={createForm.username}
                    onChange={(e) => setCreateForm({ ...createForm, username: e.target.value })}
                    placeholder="username"
                  />
                </div>

                <div>
                  <Label>昵称</Label>
                  <Input
                    value={createForm.nickname}
                    onChange={(e) => setCreateForm({ ...createForm, nickname: e.target.value })}
                    placeholder="昵称（可选）"
                  />
                </div>

                <div>
                  <Label>密码 *</Label>
                  <Input
                    type="password"
                    value={createForm.password}
                    onChange={(e) => setCreateForm({ ...createForm, password: e.target.value })}
                    placeholder="至少6位"
                  />
                </div>

                <div>
                  <Label>确认密码 *</Label>
                  <Input
                    type="password"
                    value={createForm.confirm_password}
                    onChange={(e) =>
                      setCreateForm({ ...createForm, confirm_password: e.target.value })
                    }
                    placeholder="再次输入密码"
                  />
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setCreateDialogOpen(false)}>
                  取消
                </Button>
                <Button onClick={handleCreateUser} disabled={creating}>
                  {creating ? '创建中...' : '创建'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
    </ShellLayout>
  )
}
