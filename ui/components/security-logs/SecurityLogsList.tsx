'use client'

import { useState } from 'react'
import { useLoginLogs } from '@/hooks/api/useSecurityLogs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { SkeletonCard } from '@/components/ui/skeleton-card'
import { 
  Search, Filter, RefreshCw, ChevronLeft, ChevronRight, 
  LogIn, LogOut, AlertTriangle, CheckCircle, XCircle, Activity,
  Monitor, Smartphone, Tablet, Globe, Shield
} from 'lucide-react'
import { format, formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import type { AuditLogEntry } from '@/types/security-log'

const ACTION_LABELS: Record<string, string> = {
  'login': '登录',
  'login_fail': '登录失败',
  'logout': '登出',
  'create_session': '创建会话',
  'refresh_token': '刷新令牌',
  'revoke_session': '撤销会话',
  'password_change': '修改密码',
  'password_reset': '重置密码',
  'mfa_enable': '启用MFA',
  'mfa_disable': '禁用MFA',
  'oauth_link': '绑定OAuth',
  'oauth_unlink': '解绑OAuth',
}

function getActionDisplayValue(action: string): string {
  const actionMap: Record<string, string> = {
    'all': '全部',
    'login': '登录',
    'login_fail': '登录失败',
    'logout': '登出',
    'create_session': '创建会话',
    'refresh_token': '刷新令牌',
  }
  return actionMap[action] || '操作类型'
}

function getActionStyle(action: string): {
  icon: typeof LogIn
  color: 'green' | 'red' | 'yellow' | 'blue' | 'gray'
} {
  const actionLower = action.toLowerCase()
  
  if (actionLower.includes('login') && !actionLower.includes('fail')) {
    return { icon: LogIn, color: 'green' }
  }
  if (actionLower.includes('fail') || actionLower.includes('error')) {
    return { icon: XCircle, color: 'red' }
  }
  if (actionLower.includes('logout')) {
    return { icon: LogOut, color: 'blue' }
  }
  if (actionLower.includes('warn') || actionLower.includes('risk') || actionLower.includes('security')) {
    return { icon: Shield, color: 'yellow' }
  }
  if (actionLower.includes('success') || actionLower.includes('ok')) {
    return { icon: CheckCircle, color: 'green' }
  }
  
  return { icon: Activity, color: 'gray' }
}

function parseDeviceInfo(userAgent: string): { type: string; browser: string } {
  if (!userAgent) return { type: 'unknown', browser: '未知' }
  
  let deviceType = 'desktop'
  let browser = '未知'
  
  // Detect device type
  if (userAgent.includes('Mobile') || userAgent.includes('Android')) {
    deviceType = 'mobile'
  } else if (userAgent.includes('Tablet') || userAgent.includes('iPad')) {
    deviceType = 'tablet'
  }
  
  // Detect browser
  if (userAgent.includes('Chrome')) browser = 'Chrome'
  else if (userAgent.includes('Firefox')) browser = 'Firefox'
  else if (userAgent.includes('Safari') && !userAgent.includes('Chrome')) browser = 'Safari'
  else if (userAgent.includes('Edge')) browser = 'Edge'
  else if (userAgent.includes('WeChat')) browser = '微信'
  
  return { type: deviceType, browser }
}

function getDeviceIcon(type: string) {
  switch (type) {
    case 'mobile': return Smartphone
    case 'tablet': return Tablet
    default: return Monitor
  }
}

interface SecurityLogsListProps {
  showFilters?: boolean
  pageSize?: number
}

export function SecurityLogsList({ showFilters = true, pageSize = 20 }: SecurityLogsListProps) {
  const [page, setPage] = useState(1)
  const [filters, setFilters] = useState({
    action: 'all',  // 使用 'all' 而不是空字符串作为默认值
    search: '',
  })

  const { data, isLoading, error, refetch } = useLoginLogs({
    action: filters.action && filters.action !== 'all' ? filters.action : undefined,
    search: filters.search || undefined,
    page,
    pageSize,
  })

  const handleSearch = (value: string) => {
    setFilters(prev => ({ ...prev, search: value }))
    setPage(1)
  }

  const handleActionFilter = (value: string) => {
    // 直接使用传入的值，不再进行 'all' 到空字符串的转换
    setFilters(prev => ({ ...prev, action: value }))
    setPage(1)
  }

  const handleRefresh = () => {
    refetch()
  }

  const totalPages = data ? Math.ceil(data.total / pageSize) : 0

  return (
    <Card className="backdrop-blur-xl bg-white/95 border-slate-200/20 shadow-lg">
      <CardHeader>
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <CardTitle className="flex items-center text-slate-800">
            <Activity className="h-5 w-5 mr-2 text-slate-600" />
            安全日志
          </CardTitle>
          
          {showFilters && (
            <div className="flex flex-col sm:flex-row gap-2">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-slate-400" />
                <Input
                  placeholder="搜索IP地址、设备..."
                  className="pl-9 w-full sm:w-64"
                  value={filters.search}
                  onChange={(e) => handleSearch(e.target.value)}
                />
              </div>
              
              <Select value={filters.action} onValueChange={handleActionFilter}>
                <SelectTrigger className="w-full sm:w-40">
                  <Filter className="h-4 w-4 mr-2" />
                  <SelectValue placeholder="操作类型">
                    {getActionDisplayValue(filters.action)}
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部</SelectItem>
                  <SelectItem value="login">登录</SelectItem>
                  <SelectItem value="login_fail">登录失败</SelectItem>
                  <SelectItem value="logout">登出</SelectItem>
                  <SelectItem value="create_session">创建会话</SelectItem>
                  <SelectItem value="refresh_token">刷新令牌</SelectItem>
                </SelectContent>
              </Select>
              
              <Button variant="outline" size="icon" onClick={handleRefresh}>
                <RefreshCw className="h-4 w-4" />
              </Button>
            </div>
          )}
        </div>
      </CardHeader>
      
      <CardContent>
        {/* Logs List */}
        {isLoading ? (
          <div className="space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <SkeletonCard key={i} className="h-20" />
            ))}
          </div>
        ) : error ? (
          <div className="text-center py-12">
            <AlertTriangle className="h-12 w-12 mx-auto mb-4 text-yellow-500" />
            <p className="text-slate-600">加载失败，请稍后重试</p>
            <Button variant="outline" className="mt-4" onClick={handleRefresh}>
              <RefreshCw className="h-4 w-4 mr-2" />
              重试
            </Button>
          </div>
        ) : data?.logs.length === 0 ? (
          <div className="text-center py-12">
            <Globe className="h-12 w-12 mx-auto mb-4 text-slate-300" />
            <p className="text-slate-600">暂无安全日志记录</p>
          </div>
        ) : (
          <div className="space-y-3">
            {data?.logs.map((log) => {
              const { icon: Icon, color } = getActionStyle(log.action || '')
              const { type: deviceType, browser } = parseDeviceInfo(log.userAgent || '')
              const DeviceIcon = getDeviceIcon(deviceType)
              const actionLabel = ACTION_LABELS[log.action?.toLowerCase() || ''] || log.action || '未知操作'
              
              return (
                <div
                  key={log.id}
                  className="flex items-center justify-between p-4 rounded-lg border border-slate-100 hover:bg-slate-50/50 transition-colors"
                >
                  <div className="flex items-center space-x-4">
                    <div className={`p-3 rounded-full ${
                      color === 'green' ? 'bg-green-100' :
                      color === 'red' ? 'bg-red-100' :
                      color === 'yellow' ? 'bg-yellow-100' :
                      color === 'blue' ? 'bg-blue-100' : 'bg-gray-100'
                    }`}>
                      <Icon className={`h-5 w-5 ${
                        color === 'green' ? 'text-green-600' :
                        color === 'red' ? 'text-red-600' :
                        color === 'yellow' ? 'text-yellow-600' :
                        color === 'blue' ? 'text-blue-600' : 'text-gray-600'
                      }`} />
                    </div>
                    
                    <div className="space-y-1">
                      <div className="flex items-center space-x-2">
                        <span className="font-medium text-slate-800">{actionLabel}</span>
                        <Badge variant={log.success ? 'default' : 'destructive'} className="text-xs">
                          {log.success ? '成功' : '失败'}
                        </Badge>
                      </div>
                      
                      <div className="flex items-center space-x-4 text-sm text-slate-500">
                        <span className="flex items-center">
                          <Globe className="h-3 w-3 mr-1" />
                          {log.ipAddress || '未知IP'}
                        </span>
                        <span className="flex items-center">
                          <DeviceIcon className="h-3 w-3 mr-1" />
                          {browser}
                        </span>
                        {!log.success && log.errorMessage && (
                          <span className="text-red-500 max-w-xs truncate">
                            {log.errorMessage}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  
                  <div className="text-right">
                    <div className="text-sm text-slate-600">
                      {log.createdAt && format(new Date(log.createdAt), 'yyyy-MM-dd HH:mm')}
                    </div>
                    <div className="text-xs text-slate-400">
                      {log.createdAt && formatDistanceToNow(new Date(log.createdAt), { locale: zhCN, addSuffix: true })}
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        )}
        
        {/* Pagination */}
        {!isLoading && data && data.logs.length > 0 && (
          <div className="flex items-center justify-between mt-6 pt-4 border-t border-slate-100">
            <div className="text-sm text-slate-500">
              共 {data.total} 条记录，第 {page} / {totalPages} 页
            </div>
            
            <div className="flex items-center space-x-2">
              <Button
                variant="outline"
                size="icon"
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page <= 1}
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
              
              {/* Page indicators */}
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                const pageNum = i + 1
                return (
                  <Button
                    key={pageNum}
                    variant={page === pageNum ? 'default' : 'outline'}
                    size="sm"
                    className="w-8 h-8"
                    onClick={() => setPage(pageNum)}
                  >
                    {pageNum}
                  </Button>
                )
              })}
              
              <Button
                variant="outline"
                size="icon"
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
