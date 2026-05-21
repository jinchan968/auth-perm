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
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from '@/components/ui/dialog'
import { useAuthStore } from '@/store/auth-store'
import { useRouter } from 'next/navigation'
import { showError } from '@/lib/toast'

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
  const [sessionExpired, setSessionExpired] = useState(false)
  const { setUser } = useAuthStore()
  const router = useRouter()

  const handleReLogin = async () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token')
    }
    useAuthStore.persist.clearStorage()
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
      setSessionExpired(false)
    }
  }, [isOpen, user, form])

  const onSubmit = async (data: EditProfileForm) => {
    setIsLoading(true)
    setSessionExpired(false)
    try {
      const updatedUser = await authApi.updateProfile({
        ...data,
        avatar: data.avatar || '',
      })
      setUser(updatedUser)
      onClose()
      alert('个人资料更新成功！')
    } catch (err) {
      if (err instanceof Error && err.message.includes('登录状态已过期')) {
        setSessionExpired(true)
        showError('您的登录状态已过期，请重新登录')
      } else {
        showError(err instanceof Error ? err.message : '更新失败，请重试')
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => { if (!open) onClose() }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>编辑个人资料</DialogTitle>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          {sessionExpired && (
            <div className="p-3 text-sm text-red-700 bg-red-50 border-2 border-red-400 rounded-lg">
              <div className="flex items-center justify-between">
                <span>登录状态已过期</span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleReLogin}
                  className="ml-2 text-xs h-6 border-red-300 text-red-600 hover:bg-red-50"
                >
                  重新登录
                </Button>
              </div>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="nickname" className="text-sm font-medium text-slate-700">
              昵称 <span className="text-red-500">*</span>
            </Label>
            <Input id="nickname" {...form.register('nickname')} placeholder="请输入昵称" />
            {form.formState.errors.nickname && (
              <p className="text-xs text-red-600">{form.formState.errors.nickname.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="phone" className="text-sm font-medium text-slate-700">手机号</Label>
            <Input id="phone" {...form.register('phone')} placeholder="请输入手机号" />
            {form.formState.errors.phone && (
              <p className="text-xs text-red-600">{form.formState.errors.phone.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="avatar" className="text-sm font-medium text-slate-700">头像URL</Label>
            <Input id="avatar" {...form.register('avatar')} placeholder="请输入头像URL" />
            {form.formState.errors.avatar && (
              <p className="text-xs text-red-600">{form.formState.errors.avatar.message}</p>
            )}
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose} disabled={isLoading}>
              取消
            </Button>
            <Button
              type="submit"
              disabled={isLoading}
              className="bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700"
            >
              {isLoading ? '保存中...' : '保存'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
