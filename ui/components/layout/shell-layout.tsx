'use client'

import { useState, useEffect, useCallback } from 'react'
import { Menu, X } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'
import { useAuthStore } from '@/store/auth-store'
interface ShellLayoutProps {
  pathname: string
  children: React.ReactNode
}

export function ShellLayout({ pathname, children }: ShellLayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const router = useRouter()
  const { user, isAuthenticated, isLoading, isReady } = useAuthStore()

  useEffect(() => {
    if (isReady && !isLoading && (!isAuthenticated || !user)) {
      router.push('/login')
    }
  }, [isReady, isLoading, isAuthenticated, user, router])

  const closeSidebar = useCallback(() => setSidebarOpen(false), [])

  useEffect(() => {
    if (sidebarOpen) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => {
      document.body.style.overflow = ''
    }
  }, [sidebarOpen])

  if (!isReady || isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-16 w-16 border-t-2 border-b-2 border-blue-600" />
      </div>
    )
  }

  if (!isAuthenticated || !user) {
    return null
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-slate-50">
      {/* Header */}
      <header className="bg-white/95 backdrop-blur-xl border-b border-slate-200/20 shadow-sm sticky top-0 z-10">
        <div className="px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center gap-3">
              {/* Hamburger — visible only on mobile */}
              <button
                className="lg:hidden p-2 -ml-2 rounded-lg text-slate-600 hover:bg-slate-100 transition-colors"
                onClick={() => setSidebarOpen(true)}
                aria-label="打开菜单"
              >
                <Menu className="h-5 w-5" />
              </button>
              <h1 className="text-xl sm:text-2xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
                Auth-Perm
              </h1>
            </div>
            <AvatarDropdown user={user} />
          </div>
        </div>
      </header>

      <div className="flex">
        {/* Desktop sidebar — hidden on mobile */}
        <div className="hidden lg:block">
          <DashboardSidebar pathname={pathname} />
        </div>

        {/* Main content */}
        <main className="flex-1 min-w-0 p-4 sm:p-6 lg:p-8 overflow-x-auto">
          {children}
        </main>
      </div>

      {/* Mobile drawer overlay */}
      {sidebarOpen && (
        <div className="lg:hidden fixed inset-0 z-50">
          {/* Backdrop */}
          <div
            className="absolute inset-0 bg-black/30 backdrop-blur-sm transition-opacity"
            onClick={closeSidebar}
          />
          {/* Drawer */}
          <aside className="absolute left-0 top-0 bottom-0 w-72 bg-white shadow-2xl animate-slide-in-left overflow-y-auto">
            <div className="flex justify-end p-3">
              <button
                className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors"
                onClick={closeSidebar}
                aria-label="关闭菜单"
              >
                <X className="h-5 w-5" />
              </button>
            </div>
            <DashboardSidebar pathname={pathname} />
          </aside>
        </div>
      )}
    </div>
  )
}
