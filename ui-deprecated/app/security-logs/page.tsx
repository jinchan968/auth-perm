'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'
import { DashboardHeader } from '@/components/dashboard/DashboardHeader'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { SecurityLogsList } from '@/components/security-logs/SecurityLogsList'

export default function SecurityLogsPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading, isReady } = useAuthStore()

  useEffect(() => {
    if (isReady && !isLoading && !isAuthenticated) {
      router.push('/login')
    }
  }, [isReady, isLoading, isAuthenticated, router])

  if (!isReady || isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-16 w-16 border-t-2 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (!isAuthenticated || !user) {
    return null
  }

  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '仪表盘', href: '/dashboard' },
    { label: '安全日志' },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-slate-50">
      <DashboardHeader user={user} />

      <div className="flex">
        <DashboardSidebar pathname="/security-logs" />

        <main className="flex-1 p-8">
          <Breadcrumb items={breadcrumbItems} />
          
          <div className="mt-4">
            <SecurityLogsList showFilters={true} pageSize={5} />
          </div>
        </main>
      </div>
    </div>
  )
}
