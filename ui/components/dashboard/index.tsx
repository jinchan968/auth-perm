'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { EditProfileModal } from '@/components/profile/edit-profile-modal'
import { Settings, Edit2, KeyRound, ChevronRight } from 'lucide-react'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'
import { Button } from '@/components/ui/button'
import { DashboardHeader } from './DashboardHeader'
import { UserInfoCard } from './UserInfoCard'
import { RoleCard } from './RoleCard'
import { DashboardStats } from './DashboardStats'
import { SecurityLogsCard } from './SecurityLogsCard'

export function DashboardPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading, isReady } = useAuthStore()
  const [isEditProfileOpen, setIsEditProfileOpen] = useState(false)

  if (!isReady || isLoading) {
    return (
      <div className="flex items-center justify-center h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-indigo-50/50">
        <div className="relative">
          <div className="w-20 h-20 rounded-full border-4 border-primary/20 border-t-primary animate-spin" />
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-accent animate-pulse" />
          </div>
        </div>
      </div>
    )
  }

  if (!isAuthenticated || !user) {
    if (typeof window !== 'undefined') {
      router.push('/login')
    }
    return null
  }

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '仪表盘' },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-indigo-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-slate-950">
      {/* 背景装饰 */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-primary/5 rounded-full blur-3xl" />
        <div className="absolute top-1/2 -left-40 w-80 h-80 bg-accent/5 rounded-full blur-3xl" />
        <div className="absolute -bottom-40 right-1/4 w-80 h-80 bg-primary/5 rounded-full blur-3xl" />
      </div>

      <DashboardHeader user={user} />

      <div className="relative flex">
        <DashboardSidebar pathname="/dashboard" />

        <main className="flex-1 p-8 animate-fade-in">
          <Breadcrumb items={breadcrumbItems} />

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-6">
            <UserInfoCard user={user} />
            <RoleCard user={user} />

            {/* 快捷操作卡片 - 增强版 */}
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
                  disabled
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
        </main>
      </div>

      <EditProfileModal
        isOpen={isEditProfileOpen}
        onClose={() => setIsEditProfileOpen(false)}
        user={user}
      />
    </div>
  )
}
