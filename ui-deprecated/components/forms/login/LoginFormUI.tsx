'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Mail, Lock, Eye, EyeOff } from 'lucide-react'
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
  
  const form = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: { identifier: '', password: '' },
  })

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
      {/* 错误提示 */}
      {error && (
        <div className="p-2.5 text-xs text-red-700 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center justify-between">
            <span>{error}</span>
            <button type="button" onClick={onClearError} className="text-red-500 hover:text-red-700">
              ×
            </button>
          </div>
        </div>
      )}

      {/* 邮箱/用户名输入 */}
      <div className="space-y-1.5">
        <Label htmlFor="identifier" className="text-xs font-medium text-slate-700 dark:text-slate-200">
          邮箱或用户名
        </Label>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Mail className="h-4 w-4 text-slate-400" />
          </div>
          <Input
            id="identifier"
            type="text"
            {...form.register('identifier')}
            className="w-full pl-9 pr-3 py-2.5 bg-white border border-slate-300 text-slate-900 placeholder-slate-400 rounded-lg focus:border-blue-500 focus:ring-1 focus:ring-blue-500 text-sm"
            placeholder="请输入邮箱或用户名"
          />
        </div>
        {form.formState.errors.identifier && (
          <p className="mt-1.5 text-xs text-red-600 flex items-center">
            <div className="w-1 h-1 rounded-full bg-red-600 mr-1.5" />
            {form.formState.errors.identifier.message}
          </p>
        )}
      </div>

      {/* 密码输入 */}
      <div className="space-y-1.5">
        <Label htmlFor="password" className="text-xs font-medium text-slate-700 dark:text-slate-200">
          密码
        </Label>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Lock className="h-4 w-4 text-slate-400" />
          </div>
          <Input
            id="password"
            type={showPassword ? 'text' : 'password'}
            {...form.register('password')}
            className="w-full pl-9 pr-10 py-2.5 bg-white border border-slate-300 text-slate-900 placeholder-slate-400 rounded-lg focus:border-blue-500 focus:ring-1 focus:ring-blue-500 text-sm"
            placeholder="请输入密码"
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-600"
          >
            {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </button>
        </div>
        {form.formState.errors.password && (
          <p className="mt-1.5 text-xs text-red-600 flex items-center">
            <div className="w-1 h-1 rounded-full bg-red-600 mr-1.5" />
            {form.formState.errors.password.message}
          </p>
        )}
      </div>

      {/* 忘记密码 */}
      <div className="flex items-center justify-end -mt-1">
        <a
          href="/forgot-password"
          className="text-xs text-blue-600 hover:text-blue-700"
        >
          忘记密码？
        </a>
      </div>

      {/* 登录按钮 */}
      <Button
        type="submit"
        disabled={isLoading}
        className="w-full py-2.5 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white font-medium rounded-lg shadow-lg disabled:opacity-50 disabled:cursor-not-allowed text-sm"
      >
        {isLoading ? '登录中...' : '登录'}
      </Button>
    </form>
  )
}
