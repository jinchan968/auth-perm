'use client'

import { Button, type ButtonProps } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useNavigationTransition } from '@/components/providers/navigation-transition-provider'

interface ListReturnButtonProps extends Omit<ButtonProps, 'onClick'> {
  href: string
  label?: string
  onBeforeNavigate?: () => void
  replace?: boolean
  delayMs?: number
}

export function ListReturnButton({
  href,
  label = '返回列表',
  className,
  disabled,
  onBeforeNavigate,
  replace,
  delayMs,
  ...buttonProps
}: ListReturnButtonProps) {
  const { isNavigating, navigateWithTransition } = useNavigationTransition()

  return (
    <Button
      type="button"
      variant="outline"
      className={cn('active:scale-100 transition-all duration-200', className)}
      disabled={disabled || isNavigating}
      onClick={() =>
        navigateWithTransition(href, {
          replace,
          delayMs,
          onBeforeNavigate,
        })
      }
      {...buttonProps}
    >
      {label}
    </Button>
  )
}

