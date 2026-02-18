'use client'

import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { User } from '@/lib/api/auth'
import { authApi } from '@/lib/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuthStore } from '@/store/auth-store'
import { X } from 'lucide-react'
import { useRouter } from 'next/navigation'

const editProfileSchema = z.object({
  nickname: z.string().min(1, '请输入昵称').max(50, '昵称过长'),
  phone: z.string().optional(),
  avatar: z.string().url('请输入有效的头像URL').optional().or(z.literal('')),
})

type EditProfileForm = z.infer<typeof editProfileSchema>

interface EditProfileModalProps {
  isOpen: boolean
  onClose: () => void
  user: User
}

export function EditProfileModal({ isOpen, onClose, user }: EditProfileModalProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isVisible, setIsVisible] = useState(false)
  const { setUser } = useAuthStore()
  const router = useRouter()

  const handleReLogin = async () => {
    // 清除认证状态
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token')
    }
    useAuthStore.persist.clearStorage()

    // 跳转到登录页
    router.push('/login')
    router.refresh()
  }

  const form = useForm<EditProfileForm>({
    resolver: zodResolver(editProfileSchema),
    defaultValues: {
      nickname: user.name || '',
      phone: user.profile?.phone || '',
      avatar: user.avatar || '',
    },
  })

  useEffect(() => {
    if (isOpen) {
      form.reset({
        nickname: user.name || '',
        phone: user.profile?.phone || '',
        avatar: user.avatar || '',
      })
      // 延迟显示以触发动画
      const timer = setTimeout(() => setIsVisible(true), 10)
      return () => clearTimeout(timer)
    } else {
      setIsVisible(false)
    }
  }, [isOpen, user, form])

  const onSubmit = async (data: EditProfileForm) => {
    setIsLoading(true)
    setError(null)

    try {
      const updatedUser = await authApi.updateProfile({
        ...data,
        // nickname: data.nickname,
        // phone: data.phone || '',
        avatar: data.avatar || '',
      })

      // 更新本地用户状态
      setUser(updatedUser)

      // 关闭模态框
      onClose()

      // 显示成功提示
      alert('个人资料更新成功！')
    } catch (err) {
      console.error('Failed to update profile:', err)

      // 检查是否为认证错误
      if (err instanceof Error && err.message.includes('登录状态已过期')) {
        setError('您的登录状态已过期，请重新登录')
      } else {
        setError(err instanceof Error ? err.message : '更新失败，请重试')
      }
    } finally {
      setIsLoading(false)
    }
  }

  if (!isOpen) return null

  return (
    <div
      className={`fixed inset-0 z-[70] flex items-center justify-center transition-all duration-300 ease-out ${
        isVisible ? 'opacity-100' : 'opacity-0'
      }`}
    >
      {/* Backdrop */}
      <div
        className={`absolute inset-0 z-[70] transition-all duration-300 ease-out ${
          isVisible
            ? 'bg-black/50 backdrop-blur-sm'
            : 'bg-black/0 backdrop-blur-none'
        }`}
        onClick={onClose}
      />

      {/* Modal */}
      <div
        className={`relative z-[80] w-full max-w-md bg-white rounded-2xl shadow-2xl transition-all duration-300 ease-out transform origin-center ${
          isVisible
            ? 'opacity-100 scale-100 translate-y-0'
            : 'opacity-0 scale-95 translate-y-4'
        }`}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-slate-200">
          <h2 className="text-xl font-semibold text-slate-900">编辑个人资料</h2>
          <button
            onClick={onClose}
            className="p-2 text-slate-400 hover:text-slate-600 rounded-full hover:bg-slate-100/80 transition-all duration-200 hover:scale-110"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={form.handleSubmit(onSubmit)} className="p-6 space-y-4">
          {error && (
            <div className="p-3 text-sm text-red-700 bg-red-50 border border-red-200 rounded-lg">
              <div className="flex items-start justify-between">
                <span>{error}</span>
                {error.includes('登录状态已过期') && (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={handleReLogin}
                    className="ml-2 text-xs h-6 border-red-300 text-red-600 hover:bg-red-50"
                  >
                    重新登录
                  </Button>
                )}
              </div>
            </div>
          )}

          {/* Nickname */}
          <div className="space-y-2">
            <Label htmlFor="nickname" className="text-sm font-medium text-slate-700">
              昵称 <span className="text-red-500">*</span>
            </Label>
            <Input
              id="nickname"
              {...form.register('nickname')}
              className="w-full"
              placeholder="请输入昵称"
            />
            {form.formState.errors.nickname && (
              <p className="text-xs text-red-600">
                {form.formState.errors.nickname.message}
              </p>
            )}
          </div>

          {/* Phone */}
          <div className="space-y-2">
            <Label htmlFor="phone" className="text-sm font-medium text-slate-700">
              手机号
            </Label>
            <Input
              id="phone"
              {...form.register('phone')}
              className="w-full"
              placeholder="请输入手机号"
            />
            {form.formState.errors.phone && (
              <p className="text-xs text-red-600">
                {form.formState.errors.phone.message}
              </p>
            )}
          </div>

          {/* Avatar */}
          <div className="space-y-2">
            <Label htmlFor="avatar" className="text-sm font-medium text-slate-700">
              头像URL
            </Label>
            <Input
              id="avatar"
              {...form.register('avatar')}
              className="w-full"
              placeholder="请输入头像URL"
            />
            {form.formState.errors.avatar && (
              <p className="text-xs text-red-600">
                {form.formState.errors.avatar.message}
              </p>
            )}
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-3 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={isLoading}
            >
              取消
            </Button>
            <Button
              type="submit"
              disabled={isLoading}
              className="bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700"
            >
              {isLoading ? '保存中...' : '保存'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
