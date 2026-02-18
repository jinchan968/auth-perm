'use client'

import { AvatarDropdown } from '@/components/ui/avatar-dropdown'
import { User } from '@/lib/api/auth'

interface DashboardHeaderProps {
  user: User
}

export function DashboardHeader({ user }: DashboardHeaderProps) {
  return (
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
  )
}
