
'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Home, LayoutDashboard } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { User } from '@/lib/api/auth'
import { ContactInfo } from '@/hooks/get-contact-info'

interface HomeContentProps {
  user: User | null
  contactInfo: ContactInfo | null
}

export function HomeContent({ user, contactInfo }: HomeContentProps) {
  const router = useRouter()

  const handleNavigateToDashboard = () => {
    router.push('/dashboard')
  }

  return (
    <main className="flex-1 p-8">
      <div className="flex items-center justify-center min-h-[60vh]">
        <Card className="backdrop-blur-xl bg-white/95 border-slate-200/20 shadow-lg max-w-2xl w-full">
          <CardHeader className="text-center">
            <div className="mx-auto w-16 h-16 bg-gradient-to-br from-blue-100 to-indigo-100 rounded-full flex items-center justify-center mb-4">
              <Home className="w-8 h-8 text-blue-600" />
            </div>
            <CardTitle className="text-2xl text-slate-800">欢迎来到 Auth-Perm 系统</CardTitle>
          </CardHeader>
          <CardContent className="text-center space-y-4">
            <p className="text-slate-600">
              您已成功登录系统，当前位于首页
            </p>
            <div className="flex flex-col sm:flex-row gap-3 justify-center items-center">
              <Button
                variant="outline"
                className="border-slate-300 hover:bg-slate-50"
                onClick={handleNavigateToDashboard}
              >
                <LayoutDashboard className="h-4 w-4 mr-2" />
                进入仪表盘
              </Button>
            </div>
            <div className="pt-4 border-t border-slate-200">
              <p className="text-sm text-slate-500">
                用户名：{user?.username || ''}
              </p>
              {contactInfo && (
                <p className="text-sm text-slate-500">
                  {contactInfo.type}：{contactInfo.value}
                </p>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  )
}
