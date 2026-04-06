import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface DetailPageHeaderProps {
  title: ReactNode
  description?: ReactNode
  actions?: ReactNode
  className?: string
}

export function DetailPageHeader({ title, description, actions, className }: DetailPageHeaderProps) {
  return (
    <div className={cn('mb-6 mt-4 flex flex-col gap-4 md:flex-row md:items-start md:justify-between', className)}>
      <div className="min-w-0 space-y-1">
        <h2 className="text-xl font-semibold text-slate-900">{title}</h2>
        {description ? <p className="text-sm text-slate-500">{description}</p> : null}
      </div>
      {actions ? <div className="shrink-0">{actions}</div> : null}
    </div>
  )
}

