'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { EditProfileModal } from '@/components/profile/edit-profile-modal'
import { Settings } from 'lucide-react'
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
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-16 w-16 border-t-2 border-b-2 border-blue-600"></div>
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
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-slate-50">
      <DashboardHeader user={user} />

      <div className="flex">
        <DashboardSidebar pathname="/dashboard" />

        <main className="flex-1 p-8">
          <Breadcrumb items={breadcrumbItems} />
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-4">
            <UserInfoCard user={user} />
            <RoleCard user={user} />
            
            <Card className="backdrop-blur-xl bg-white/95 border-slate-200/20 shadow-lg">
              <CardHeader>
                <CardTitle className="flex items-center text-slate-800">
                  <Settings className="h-5 w-5 mr-2 text-slate-600" />
                  快捷操作
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button
                  variant="outline"
                  className="w-full justify-start border-slate-300 hover:bg-slate-50"
                  onClick={() => setIsEditProfileOpen(true)}
                >
                  编辑个人资料
                </Button>
                <Button
                  variant="outline"
                  className="w-full justify-start border-slate-300 hover:bg-slate-50"
                  disabled
                >
                  修改密码
                </Button>
              </CardContent>
            </Card>
          </div>

          <DashboardStats />

          <div className="mt-6">
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
