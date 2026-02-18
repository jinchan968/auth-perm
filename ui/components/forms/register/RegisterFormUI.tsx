'use client'

import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Mail, Phone, User, Lock } from 'lucide-react'
import { useEffect } from 'react'
import { IdentifierTypeSelector } from './IdentifierTypeSelector'

const registerSchema = z.object({
  identifier_type: z.enum(['email', 'phone'], {
    required_error: '请选择注册方式',
  }),
  email: z.string().optional(),
  phone: z.string().optional(),
  username: z.string().min(3, '用户名至少3位'),
  password: z.string().min(6, '密码至少6位'),
  confirm_password: z.string().min(6, '请确认密码'),
}).refine((data) => {
  if (data.identifier_type === 'email' && !data.email) {
    return false
  }
  if (data.identifier_type === 'phone' && !data.phone) {
    return false
  }
  return true
}, {
  message: '请填写邮箱或手机号',
  path: ['email'],
}).refine((data) => data.password === data.confirm_password, {
  message: '两次输入的密码不一致',
  path: ['confirm_password'],
})

type RegisterForm = z.infer<typeof registerSchema>

interface RegisterFormUIProps {
  error: string | null
  onSubmit: (data: RegisterForm) => Promise<void>
  onClearError: () => void
}

export function RegisterFormUI({ error, onSubmit, onClearError }: RegisterFormUIProps) {
  const form = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      identifier_type: 'email' as 'email' | 'phone',
    },
  })

  useEffect(() => {
    if (!form.getValues('identifier_type')) {
      form.setValue('identifier_type', 'email')
    }
  }, [form])

  const identifierType = form.watch('identifier_type')

  return (
    <FormProvider {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        {/* 错误提示 */}
        {error && (
          <div className="mb-3 p-2.5 text-xs text-red-700 bg-red-50 border border-red-200 rounded-lg">
            <div className="flex items-center justify-between">
              <div className="flex items-center">
                <div className="w-1.5 h-1.5 rounded-full bg-red-500 mr-2" />
                {error}
              </div>
              <button
                type="button"
                onClick={onClearError}
                className="text-red-500 hover:text-red-700 ml-2"
              >
                ×
              </button>
            </div>
          </div>
        )}

        {/* 注册方式选择 */}
        <IdentifierTypeSelector
          errors={form.formState.errors.identifier_type}
        />

      {/* 邮箱输入 */}
      {(identifierType === 'email' || !identifierType) && (
        <div className="space-y-1.5">
          <Label htmlFor="email" className="text-xs font-medium text-slate-700 dark:text-slate-200">
            邮箱
          </Label>
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Mail className="h-4 w-4 text-slate-400" />
            </div>
            <Input
              id="email"
              type="email"
              {...form.register('email')}
              className="w-full pl-9 pr-3 py-2.5 bg-white border border-slate-300 text-slate-900 placeholder-slate-400 rounded-lg focus:border-blue-500 focus:ring-1 focus:ring-blue-500 text-sm"
              placeholder="请输入邮箱"
            />
            {form.formState.errors.email && (
              <p className="mt-1.5 text-xs text-red-600 flex items-center">
                <div className="w-1 h-1 rounded-full bg-red-600 mr-1.5" />
                {form.formState.errors.email.message}
              </p>
            )}
          </div>
        </div>
      )}

      {/* 手机号输入 */}
      {identifierType === 'phone' && (
        <div className="space-y-1.5">
          <Label htmlFor="phone" className="text-xs font-medium text-slate-700 dark:text-slate-200">
            手机号
          </Label>
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Phone className="h-4 w-4 text-slate-400" />
            </div>
            <Input
              id="phone"
              type="tel"
              {...form.register('phone')}
              className="w-full pl-9 pr-3 py-2.5 bg-white border border-slate-300 text-slate-900 placeholder-slate-400 rounded-lg focus:border-blue-500 focus:ring-1 focus:ring-blue-500 text-sm"
              placeholder="请输入手机号"
            />
            {form.formState.errors.phone && (
              <p className="mt-1.5 text-xs text-red-600 flex items-center">
                <div className="w-1 h-1 rounded-full bg-red-600 mr-1.5" />
                {form.formState.errors.phone.message}
              </p>
            )}
          </div>
        </div>
      )}

      {/* 用户名输入 */}
      <div className="space-y-1.5">
        <Label htmlFor="username" className="text-xs font-medium text-slate-700 dark:text-slate-200">
          用户名
        </Label>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <User className="h-4 w-4 text-slate-400" />
          </div>
          <Input
            id="username"
            {...form.register('username')}
            className="w-full pl-9 pr-3 py-2.5 bg-white border border-slate-300 text-slate-900 placeholder-slate-400 rounded-lg focus:border-blue-500 focus:ring-1 focus:ring-blue-500 text-sm"
            placeholder="请输入用户名"
          />
          {form.formState.errors.username && (
            <p className="mt-1.5 text-xs text-red-600 flex items-center">
              <div className="w-1 h-1 rounded-full bg-red-600 mr-1.5" />
              {form.formState.errors.username.message}
            </p>
          )}
        </div>
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
            type="password"
            {...form.register('password')}
            className="w-full pl-9 pr-3 py-2.5 bg-white border border-slate-300 text-slate-900 placeholder-slate-400 rounded-lg focus:border-blue-500 focus:ring-1 focus:ring-blue-500 text-sm"
            placeholder="请输入密码"
          />
          {form.formState.errors.password && (
            <p className="mt-1.5 text-xs text-red-600 flex items-center">
              <div className="w-1 h-1 rounded-full bg-red-600 mr-1.5" />
              {form.formState.errors.password.message}
            </p>
          )}
        </div>
      </div>

      {/* 确认密码输入 */}
      <div className="space-y-1.5">
        <Label htmlFor="confirm_password" className="text-xs font-medium text-slate-700 dark:text-slate-200">
          确认密码
        </Label>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Lock className="h-4 w-4 text-slate-400" />
          </div>
          <Input
            id="confirm_password"
            type="password"
            {...form.register('confirm_password')}
            className="w-full pl-9 pr-3 py-2.5 bg-white border border-slate-300 text-slate-900 placeholder-slate-400 rounded-lg focus:border-blue-500 focus:ring-1 focus:ring-blue-500 text-sm"
            placeholder="请再次输入密码"
          />
          {form.formState.errors.confirm_password && (
            <p className="mt-1.5 text-xs text-red-600 flex items-center">
              <div className="w-1 h-1 rounded-full bg-red-600 mr-1.5" />
              {form.formState.errors.confirm_password.message}
            </p>
          )}
        </div>
      </div>

      {/* 注册按钮 */}
      <Button
        type="submit"
        className="w-full py-2.5 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white font-medium rounded-lg shadow-lg text-sm"
      >
        注册
      </Button>
      </form>
    </FormProvider>
  )
}
