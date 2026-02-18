'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { User } from '@/lib/api/auth'
import { User as UserIcon, Mail, Phone, CheckCircle2 } from 'lucide-react'
import { getContactInfo } from '@/hooks/get-contact-info'

interface UserInfoCardProps {
  user: User
}

export function UserInfoCard({ user }: UserInfoCardProps) {
  const contactInfo = getContactInfo(user)

  return (
    <Card variant="glass" className="animate-slide-up hover-lift">
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center text-foreground">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary to-accent flex items-center justify-center mr-3 shadow-md">
            <UserIcon className="h-4 w-4 text-white" />
          </div>
          用户信息
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* 头像和用户名 */}
        <div className="flex items-center gap-4">
          <div className="relative">
            <div className="w-16 h-16 rounded-full bg-gradient-to-br from-primary to-accent p-0.5 shadow-lg">
              <div className="w-full h-full rounded-full bg-background flex items-center justify-center text-2xl font-bold text-primary">
                {user.username?.charAt(0).toUpperCase() || 'U'}
              </div>
            </div>
            <div className="absolute -bottom-1 -right-1 w-5 h-5 bg-success rounded-full flex items-center justify-center border-2 border-background">
              <CheckCircle2 className="w-3 h-3 text-white" />
            </div>
          </div>
          <div>
            <p className="font-semibold text-lg text-foreground">{user.username}</p>
            <p className="text-sm text-muted-foreground">已验证用户</p>
          </div>
        </div>

        {/* 联系信息 */}
        {contactInfo && (
          <div className="pt-3 border-t border-border/50">
            <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/50 hover:bg-muted transition-colors">
              {contactInfo.type === '邮箱' ? (
                <Mail className="h-4 w-4 text-primary" />
              ) : (
                <Phone className="h-4 w-4 text-primary" />
              )}
              <div>
                <p className="text-xs text-muted-foreground">{contactInfo.type}</p>
                <p className="font-medium text-foreground">{contactInfo.value}</p>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
