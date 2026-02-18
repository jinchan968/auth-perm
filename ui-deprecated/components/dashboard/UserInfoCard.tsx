'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { User } from '@/lib/api/auth'
import { User as UserIcon } from 'lucide-react'
import { getContactInfo } from '@/hooks/get-contact-info'

interface UserInfoCardProps {
  user: User
}

export function UserInfoCard({ user }: UserInfoCardProps) {
  const contactInfo = getContactInfo(user)

  return (
    <Card className="backdrop-blur-xl bg-white/95 border-slate-200/20 shadow-lg">
      <CardHeader>
        <CardTitle className="flex items-center text-slate-800">
          <UserIcon className="h-5 w-5 mr-2 text-blue-600" />
          用户信息
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        <div>
          <p className="text-sm text-slate-600">用户名</p>
          <p className="font-medium text-slate-900">{user.username}</p>
        </div>
        {contactInfo && (
          <div>
            <p className="text-sm text-slate-600">{contactInfo.type}</p>
            <p className="font-medium text-slate-900">{contactInfo.value}</p>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
