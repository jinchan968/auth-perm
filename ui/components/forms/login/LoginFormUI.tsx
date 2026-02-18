'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Mail, Lock, Eye, EyeOff, Loader2 } from 'lucide-react'
import { useState } from 'react'

const loginSchema = z.object({
  identifier: z.string().min(1, '请输入邮箱或手机号'),
  password: z.string().min(6, '密码至少6位'),
})

type LoginForm = z.infer<typeof loginSchema>

interface LoginFormUIProps {
  isLoading: boolean
  error: string | null
  onSubmit: (data: LoginForm) => Promise<void>
  onClearError: () => void
}

export function LoginFormUI({ isLoading, error, onSubmit, onClearError }: LoginFormUIProps) {
  const [showPassword, setShowPassword] = useState(false)
  const [hasError, setHasError] = useState(false)

  const form = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: { identifier: '', password: '' },
  })

  const handleSubmit = async (data: LoginForm) => {
    setHasError(false)
    try {
      await onSubmit(data)
    } catch {
      setHasError(true)
    }
  }

  return (
    <form
      onSubmit={form.handleSubmit(handleSubmit)}
      className={`space-y-5 ${hasError || error ? 'animate-shake' : ''}`}
    >
      {/* 错误提示 */}
      {error && (
        <div className="p-3 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-lg animate-slide-down">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-destructive animate-pulse-subtle" />
              <span>{error}</span>
            </div>
            <button
              type="button"
              onClick={onClearError}
              className="text-destructive/70 hover:text-destructive transition-colors p-1 hover:bg-destructive/10 rounded"
            >
              ×
            </button>
          </div>
        </div>
      )}

      {/* 邮箱/用户名输入 */}
      <div className="space-y-2">
        <Label htmlFor="identifier" className="text-sm font-medium text-foreground/90">
          邮箱或用户名
        </Label>
        <div className="relative group">
          <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none z-10">
            <Mail className="h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors duration-200" />
          </div>
          <Input
            id="identifier"
            type="text"
            {...form.register('identifier')}
            error={!!form.formState.errors.identifier}
            className="pl-10"
            placeholder="请输入邮箱或用户名"
          />
        </div>
        {form.formState.errors.identifier && (
          <p className="mt-1.5 text-xs text-destructive flex items-center gap-1.5 animate-slide-down">
            <div className="w-1.5 h-1.5 rounded-full bg-destructive" />
            {form.formState.errors.identifier.message}
          </p>
        )}
      </div>

      {/* 密码输入 */}
      <div className="space-y-2">
        <Label htmlFor="password" className="text-sm font-medium text-foreground/90">
          密码
        </Label>
        <div className="relative group">
          <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none z-10">
            <Lock className="h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors duration-200" />
          </div>
          <Input
            id="password"
            type={showPassword ? 'text' : 'password'}
            {...form.register('password')}
            error={!!form.formState.errors.password}
            className="pl-10 pr-11"
            placeholder="请输入密码"
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute inset-y-0 right-0 pr-3.5 flex items-center text-muted-foreground hover:text-foreground transition-colors duration-200"
          >
            {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </button>
        </div>
        {form.formState.errors.password && (
          <p className="mt-1.5 text-xs text-destructive flex items-center gap-1.5 animate-slide-down">
            <div className="w-1.5 h-1.5 rounded-full bg-destructive" />
            {form.formState.errors.password.message}
          </p>
        )}
      </div>

      {/* 忘记密码 */}
      <div className="flex items-center justify-end">
        <a
          href="/forgot-password"
          className="text-sm text-primary hover:text-primary/80 transition-colors duration-200 hover:underline underline-offset-4"
        >
          忘记密码？
        </a>
      </div>

      {/* 登录按钮 */}
      <Button
        type="submit"
        disabled={isLoading}
        className="w-full h-11 text-base font-medium"
        size="lg"
      >
        {isLoading ? (
          <span className="flex items-center gap-2">
            <Loader2 className="h-4 w-4 animate-spin" />
            登录中...
          </span>
        ) : (
          '登录'
        )}
      </Button>
    </form>
  )
}
