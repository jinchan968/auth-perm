'use client'

import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { ListReturnButton } from '@/components/ui/list-return-button'

interface DetailActionBarProps {
  children?: ReactNode
  returnHref: string
  returnLabel?: string
  onBeforeReturn?: () => void
  className?: string
}

export function DetailActionBar({
  children,
  returnHref,
  returnLabel = '返回',
  onBeforeReturn,
  className,
}: DetailActionBarProps) {
  return (
    <div className={cn('flex flex-wrap items-center gap-2', className)}>
      {children}
      <ListReturnButton
        href={returnHref}
        label={returnLabel}
        onBeforeNavigate={onBeforeReturn}
      />
    </div>
  )
}

