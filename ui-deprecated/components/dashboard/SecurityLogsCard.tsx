'use client'

import { useRecentLoginActivity } from '@/hooks/api/useSecurityLogs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Activity, LogIn, LogOut, AlertTriangle, CheckCircle, XCircle, ArrowRight } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

/**
 * Get icon and color based on action type
 * 根据操作类型获取图标和颜色
 */
function getActionStyle(action: string): {
  icon: typeof LogIn
  color: 'green' | 'red' | 'yellow' | 'blue' | 'gray'
  label: string
} {
  const actionLower = action.toLowerCase()
  
  if (actionLower.includes('login') || actionLower.includes('signin')) {
    return { icon: LogIn, color: 'green', label: '登录' }
  }
  if (actionLower.includes('logout') || actionLower.includes('signout')) {
    return { icon: LogOut, color: 'blue', label: '登出' }
  }
  if (actionLower.includes('fail') || actionLower.includes('error') || actionLower.includes('failed')) {
    return { icon: XCircle, color: 'red', label: '失败' }
  }
  if (actionLower.includes('warn') || actionLower.includes('warning') || actionLower.includes('risk')) {
    return { icon: AlertTriangle, color: 'yellow', label: '警告' }
  }
  if (actionLower.includes('success') || actionLower.includes('ok')) {
    return { icon: CheckCircle, color: 'green', label: '成功' }
  }
  
  return { icon: Activity, color: 'gray', label: action }
}

/**
 * Get color class based on status
 * 根据状态获取颜色类名
 */
function getColorClasses(color: 'green' | 'red' | 'yellow' | 'blue' | 'gray') {
  const colorMap = {
    green: 'text-green-600 bg-green-50 border-green-200',
    red: 'text-red-600 bg-red-50 border-red-200',
    yellow: 'text-yellow-600 bg-yellow-50 border-yellow-200',
    blue: 'text-blue-600 bg-blue-50 border-blue-200',
    gray: 'text-gray-600 bg-gray-50 border-gray-200',
  }
  return colorMap[color]
}

/**
 * Format user agent string
 * 格式化用户代理字符串
 */
function formatUserAgent(userAgent: string): string {
  if (!userAgent) return '未知设备'
  
  // Common browsers
  if (userAgent.includes('Chrome')) return 'Chrome 浏览器'
  if (userAgent.includes('Firefox')) return 'Firefox 浏览器'
  if (userAgent.includes('Safari')) {
    if (userAgent.includes('Mobile')) return 'Safari 移动版'
    return 'Safari 浏览器'
  }
  if (userAgent.includes('Edge')) return 'Edge 浏览器'
  if (userAgent.includes('WeChat')) return '微信内置浏览器'
  
  // Mobile devices
  if (userAgent.includes('Mobile')) return '移动设备'
  if (userAgent.includes('Windows')) return 'Windows 设备'
  if (userAgent.includes('Mac')) return 'Mac 设备'
  
  return '未知设备'
}

export function SecurityLogsCard() {
  const { data: logs, isLoading, error } = useRecentLoginActivity()

  if (isLoading) {
    return (
      <Card className="backdrop-blur-xl bg-white/95 border-slate-200/20 shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center text-slate-800">
            <Activity className="h-5 w-5 mr-2 text-slate-600" />
            最近活动
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-center space-x-4">
                <Skeleton className="h-10 w-10 rounded-full" />
                <div className="space-y-2 flex-1">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-3 w-1/2" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className="backdrop-blur-xl bg-white/95 border-slate-200/20 shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center text-slate-800">
            <Activity className="h-5 w-5 mr-2 text-slate-600" />
            最近活动
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-4 text-slate-500">
            <AlertTriangle className="h-8 w-8 mx-auto mb-2 text-yellow-500" />
            <p>加载失败，请稍后重试</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  const displayLogs = logs?.slice(0, 5) || []

  return (
    <Card className="backdrop-blur-xl bg-white/95 border-slate-200/20 shadow-lg">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="flex items-center text-slate-800">
          <Activity className="h-5 w-5 mr-2 text-slate-600" />
          最近活动
        </CardTitle>
        <Button 
          variant="ghost" 
          size="sm" 
          className="text-slate-600 hover:text-slate-800"
          onClick={() => window.location.href = '/security-logs'}
        >
          查看全部
          <ArrowRight className="h-4 w-4 ml-1" />
        </Button>
      </CardHeader>
      <CardContent>
        {displayLogs.length === 0 ? (
          <div className="text-center py-8 text-slate-500">
            <Activity className="h-12 w-12 mx-auto mb-3 text-slate-300" />
            <p>暂无登录记录</p>
          </div>
        ) : (
          <div className="space-y-3">
            {displayLogs.map((log) => {
              const { icon: Icon, color, label } = getActionStyle(log.action || '')
              const colorClasses = getColorClasses(color)
              const deviceInfo = formatUserAgent(log.userAgent || '')
              const timeAgo = log.createdAt 
                ? formatDistanceToNow(new Date(log.createdAt), { locale: zhCN, addSuffix: true })
                : '未知时间'

              return (
                <div 
                  key={log.id} 
                  className="flex items-center justify-between p-3 rounded-lg bg-slate-50/50 hover:bg-slate-100/50 transition-colors"
                >
                  <div className="flex items-center space-x-3">
                    <div className={`p-2 rounded-full border ${colorClasses}`}>
                      <Icon className="h-4 w-4" />
                    </div>
                    <div>
                      <div className="flex items-center space-x-2">
                        <span className="font-medium text-slate-800">{label}</span>
                        <Badge variant="outline" className="text-xs">
                          {log.success ? '成功' : '失败'}
                        </Badge>
                      </div>
                      <div className="text-sm text-slate-500 flex items-center space-x-2">
                        <span>{deviceInfo}</span>
                        <span>•</span>
                        <span>{log.ipAddress || '未知IP'}</span>
                      </div>
                    </div>
                  </div>
                  <span className="text-xs text-slate-400 whitespace-nowrap">
                    {timeAgo}
                  </span>
                </div>
              )
            })}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
