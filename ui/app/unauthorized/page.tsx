'use client'

import { ShieldX } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useRouter } from 'next/navigation'

export default function UnauthorizedPage() {
  const router = useRouter()

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 via-blue-50/30 to-indigo-50/50">
      <div className="text-center space-y-6 max-w-md px-6">
        <div className="flex justify-center">
          <div className="rounded-full bg-red-100 p-4">
            <ShieldX className="h-12 w-12 text-red-500" />
          </div>
        </div>

        <div className="space-y-2">
          <h1 className="text-2xl font-bold text-slate-900">无权限访问</h1>
          <p className="text-slate-500">
            您的登录会话已过期或无权访问该资源，请重新登录。
          </p>
        </div>

        <div className="flex gap-3 justify-center">
          <Button
            variant="outline"
            onClick={() => router.push('/home')}
          >
            返回首页
          </Button>
          <Button
            onClick={() => router.push('/login')}
          >
            重新登录
          </Button>
        </div>
      </div>
    </div>
  )
}
