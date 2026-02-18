'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Shield } from 'lucide-react'
import { User } from '@/lib/api/auth'

interface RoleCardProps {
  user: User
}

export function RoleCard({ user }: RoleCardProps) {
  return (
    <Card className="backdrop-blur-xl bg-white/95 border-slate-200/20 shadow-lg">
      <CardHeader>
        <CardTitle className="flex items-center text-slate-800">
          <Shield className="h-5 w-5 mr-2 text-indigo-600" />
          角色权限
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <p className="text-sm text-slate-600">角色</p>
          <div className="flex flex-wrap gap-2">
            {user.roles && user.roles.length > 0 ? (
              user.roles.map((role) => (
                <span
                  key={role}
                  className="px-2 py-1 text-xs bg-gradient-to-r from-blue-600/10 to-indigo-600/10 text-blue-700 rounded-full"
                >
                  {role}
                </span>
              ))
            ) : (
              <span className="text-sm text-slate-500">无角色</span>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
