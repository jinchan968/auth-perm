
'use client'

import { Button } from '@/components/ui/button'
import { Home, LayoutDashboard, Building2 } from 'lucide-react'
import { useRouter } from 'next/navigation'

interface DashboardSidebarProps {
  pathname: string
}

export function DashboardSidebar({ pathname }: DashboardSidebarProps) {
  const router = useRouter()

  const isHomeActive = pathname === '/home'
  const isDashboardActive = pathname === '/dashboard'
  const isTenantsActive = pathname.startsWith('/tenants')

  return (
    <aside className="w-64 bg-white/95 backdrop-blur-xl border-r border-slate-200/20 shadow-sm min-h-screen">
      <div className="p-4">
        <Button
          variant={isHomeActive ? "secondary" : "ghost"}
          className={`w-full justify-start mt-2 ${isHomeActive ? 'bg-gradient-to-r from-blue-600 to-indigo-600 text-white' : 'text-slate-600 hover:bg-slate-100'}`}
          onClick={() => router.push('/home')}
        >
          <Home className="h-4 w-4 mr-2" />
          首页
        </Button>

        <Button
          variant={isDashboardActive ? "secondary" : "ghost"}
          className={`w-full justify-start mt-2 ${isDashboardActive ? 'bg-gradient-to-r from-blue-600 to-indigo-600 text-white' : 'text-slate-600 hover:bg-slate-100'}`}
          onClick={() => router.push('/dashboard')}
        >
          <LayoutDashboard className="h-4 w-4 mr-2" />
          仪表盘
        </Button>

        <Button
          variant={isTenantsActive ? "secondary" : "ghost"}
          className={`w-full justify-start mt-2 ${isTenantsActive ? 'bg-gradient-to-r from-blue-600 to-indigo-600 text-white' : 'text-slate-600 hover:bg-slate-100'}`}
          onClick={() => router.push('/tenants')}
        >
          <Building2 className="h-4 w-4 mr-2" />
          租户管理
        </Button>
      </div>
    </aside>
  )
}
