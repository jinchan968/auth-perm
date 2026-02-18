'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Shield, Crown, Users, Key } from 'lucide-react'
import { User } from '@/lib/api/auth'

interface RoleCardProps {
  user: User
}

// 角色图标映射
const roleIcons: Record<string, React.ReactNode> = {
  admin: <Crown className="w-3 h-3" />,
  user: <Users className="w-3 h-3" />,
  default: <Key className="w-3 h-3" />,
}

// 角色颜色映射
const roleColors: Record<string, string> = {
  admin: 'from-amber-500 to-orange-500 text-white',
  superadmin: 'from-purple-500 to-pink-500 text-white',
  user: 'from-primary/20 to-accent/20 text-primary',
  default: 'from-muted to-muted text-muted-foreground',
}

export function RoleCard({ user }: RoleCardProps) {
  const getRoleColor = (role: string) => {
    const lowerRole = role.toLowerCase()
    return roleColors[lowerRole] || roleColors.default
  }

  const getRoleIcon = (role: string) => {
    const lowerRole = role.toLowerCase()
    return roleIcons[lowerRole] || roleIcons.default
  }

  return (
    <Card variant="glass" className="animate-slide-up hover-lift" style={{ animationDelay: '50ms' }}>
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center text-foreground">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-accent to-purple-400 flex items-center justify-center mr-3 shadow-md">
            <Shield className="h-4 w-4 text-white" />
          </div>
          角色权限
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          <p className="text-sm text-muted-foreground">当前角色</p>
          <div className="flex flex-wrap gap-2">
            {user.roles && user.roles.length > 0 ? (
              user.roles.map((role) => (
                <span
                  key={role}
                  className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-gradient-to-r ${getRoleColor(role)} rounded-full shadow-sm transition-all duration-200 hover:scale-105 hover:shadow-md`}
                >
                  {getRoleIcon(role)}
                  {role}
                </span>
              ))
            ) : (
              <span className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs text-muted-foreground bg-muted rounded-full">
                <Users className="w-3 h-3" />
                无角色
              </span>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
