'use client'

import { ShellLayout } from '@/components/layout/shell-layout'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import { SecurityLogsList } from '@/components/security-logs/SecurityLogsList'

export default function SecurityLogsPage() {
  const breadcrumbItems = [
    { label: '首页', href: '/home' },
    { label: '仪表盘', href: '/dashboard' },
    { label: '安全日志' },
  ]

  return (
    <ShellLayout pathname="/security-logs">
      <Breadcrumb items={breadcrumbItems} />
      
      <div className="mt-4">
        <SecurityLogsList showFilters={true} pageSize={5} />
      </div>
    </ShellLayout>
  )
}
