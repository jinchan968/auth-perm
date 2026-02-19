
'use client'

import { useRouter } from 'next/navigation'
import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { getContactInfo } from '@/hooks/get-contact-info'
import { useAuthStore } from '@/store/auth-store'
import { DashboardSidebar } from '@/components/layout/dashboard-sidebar'
import { HomeContent } from '@/components/home/home-content'

export default function HomePage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading, isReady } = useAuthStore()

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

  const contactInfo = getContactInfo(user)

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-slate-50">
      <header className="bg-white/95 backdrop-blur-xl border-b border-slate-200/20 shadow-sm sticky top-0 z-10">
        <div className="px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
              Auth-Perm
            </h1>
            <AvatarDropdown user={user} />
          </div>
        </div>
      </header>

      <div className="flex">
        <DashboardSidebar pathname="/home" />
        <main className="flex-1 p-8">
          <HomeContent user={user} contactInfo={contactInfo} />
        </main>
      </div>
    </div>
  )
}
