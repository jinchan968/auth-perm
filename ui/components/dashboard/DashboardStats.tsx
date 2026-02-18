'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Activity, Smartphone, Shield, TrendingUp, TrendingDown, Minus } from 'lucide-react'

interface StatCardProps {
  title: string
  value: string | number
  description: string
  icon: React.ReactNode
  trend?: 'up' | 'down' | 'neutral'
  gradientFrom: string
  gradientTo: string
  delay?: string
}

function StatCard({ title, value, description, icon, trend, gradientFrom, gradientTo, delay }: StatCardProps) {
  return (
    <Card
      variant="glass"
      className="animate-slide-up hover-lift group"
      style={{ animationDelay: delay }}
    >
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-medium text-muted-foreground flex items-center justify-between">
          <span>{title}</span>
          <div className={`w-10 h-10 rounded-xl bg-gradient-to-br ${gradientFrom} ${gradientTo} flex items-center justify-center shadow-lg group-hover:scale-110 transition-transform duration-300`}>
            {icon}
          </div>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-end gap-2">
          <p className={`text-4xl font-bold bg-gradient-to-r ${gradientFrom} ${gradientTo} bg-clip-text text-transparent`}>
            {value}
          </p>
          {trend && (
            <div className={`flex items-center gap-0.5 text-xs font-medium mb-1.5 ${trend === 'up' ? 'text-success' :
                trend === 'down' ? 'text-destructive' :
                  'text-muted-foreground'
              }`}>
              {trend === 'up' && <TrendingUp className="w-3 h-3" />}
              {trend === 'down' && <TrendingDown className="w-3 h-3" />}
              {trend === 'neutral' && <Minus className="w-3 h-3" />}
            </div>
          )}
        </div>
        <p className="text-sm text-muted-foreground mt-1">{description}</p>
      </CardContent>
    </Card>
  )
}

export function DashboardStats() {
  return (
    <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-6">
      <StatCard
        title="登录会话"
        value={1}
        description="当前活跃会话"
        icon={<Activity className="w-5 h-5 text-white" />}
        trend="neutral"
        gradientFrom="from-primary"
        gradientTo="to-blue-400"
        delay="0ms"
      />

      <StatCard
        title="已授权设备"
        value={1}
        description="已授权设备数量"
        icon={<Smartphone className="w-5 h-5 text-white" />}
        trend="up"
        gradientFrom="from-accent"
        gradientTo="to-purple-400"
        delay="100ms"
      />

      <StatCard
        title="安全评分"
        value={85}
        description="安全评分（满分100）"
        icon={<Shield className="w-5 h-5 text-white" />}
        trend="up"
        gradientFrom="from-success"
        gradientTo="to-emerald-400"
        delay="200ms"
      />
    </div>
  )
}
