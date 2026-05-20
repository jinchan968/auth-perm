'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { EditProfileModal } from '@/components/profile/edit-profile-modal'
import { ChangePasswordModal } from '@/components/profile/change-password-modal'
import { Settings, Edit2, KeyRound, ChevronRight } from 'lucide-react'
import { useAuthStore } from '@/store/auth-store'
import { ShellLayout } from '@/components/layout/shell-layout'
import { Button } from '@/components/ui/button'
import { UserInfoCard } from './UserInfoCard'
import { RoleCard } from './RoleCard'
import { DashboardStats } from './DashboardStats'
import { SecurityLogsCard } from './SecurityLogsCard'

export function DashboardPage() {
  const { user } = useAuthStore()
  const [isEditProfileOpen, setIsEditProfileOpen] = useState(false)
  const [isChangePwdOpen, setIsChangePwdOpen] = useState(false)

  if (!user) return null

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '仪表盘' },
  ]

  return (
    <ShellLayout pathname="/dashboard">
      {/* 背景装饰 */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none -z-10">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-primary/5 rounded-full blur-3xl" />
        <div className="absolute top-1/2 -left-40 w-80 h-80 bg-accent/5 rounded-full blur-3xl" />
        <div className="absolute -bottom-40 right-1/4 w-80 h-80 bg-primary/5 rounded-full blur-3xl" />
      </div>

      <Breadcrumb items={breadcrumbItems} />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-6">
        <UserInfoCard user={user} />
        <RoleCard user={user} />

        <Card variant="glass" className="animate-slide-up hover-lift" style={{ animationDelay: '100ms' }}>
          <CardHeader className="pb-4">
            <CardTitle className="flex items-center text-foreground">
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-slate-500 to-slate-600 flex items-center justify-center mr-3 shadow-md">
                <Settings className="h-4 w-4 text-white" />
              </div>
              快捷操作
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Button
              variant="outline"
              className="w-full justify-between group"
              onClick={() => setIsEditProfileOpen(true)}
            >
              <span className="flex items-center gap-2">
                <Edit2 className="w-4 h-4 text-primary" />
                编辑个人资料
              </span>
              <ChevronRight className="w-4 h-4 text-muted-foreground group-hover:translate-x-1 transition-transform" />
            </Button>
            <Button
              variant="outline"
              className="w-full justify-between group"
              onClick={() => setIsChangePwdOpen(true)}
            >
              <span className="flex items-center gap-2">
                <KeyRound className="w-4 h-4 text-muted-foreground" />
                修改密码
              </span>
              <ChevronRight className="w-4 h-4 text-muted-foreground" />
            </Button>
          </CardContent>
        </Card>
      </div>

      <DashboardStats />

      <div className="mt-6 animate-slide-up" style={{ animationDelay: '300ms' }}>
        <SecurityLogsCard />
      </div>

      <EditProfileModal
        isOpen={isEditProfileOpen}
        onClose={() => setIsEditProfileOpen(false)}
        user={user}
      />
      <ChangePasswordModal
        isOpen={isChangePwdOpen}
        onClose={() => setIsChangePwdOpen(false)}
      />
    </ShellLayout>
  )
}
