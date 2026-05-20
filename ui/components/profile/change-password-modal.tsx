'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { authApi } from '@/lib/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { AppModal } from '@/components/ui/app-modal'
import { showError, showSuccess } from '@/lib/toast'

const changePasswordSchema = z
  .object({
    old_password: z.string().min(1, '请输入当前密码'),
    new_password: z.string().min(6, '新密码至少6位'),
    confirm_password: z.string().min(1, '请确认新密码'),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: '两次输入的密码不一致',
    path: ['confirm_password'],
  })

type ChangePasswordForm = z.infer<typeof changePasswordSchema>

interface ChangePasswordModalProps {
  isOpen: boolean
  onClose: () => void
}

export function ChangePasswordModal({ isOpen, onClose }: ChangePasswordModalProps) {
  const [saving, setSaving] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ChangePasswordForm>({
    resolver: zodResolver(changePasswordSchema),
  })

  const onSubmit = async (data: ChangePasswordForm) => {
    setSaving(true)
    try {
      await authApi.changePassword({
        old_password: data.old_password,
        new_password: data.new_password,
        confirm_password: data.confirm_password,
      })
      showSuccess('密码修改成功')
      reset()
      onClose()
    } catch (err) {
      showError(err instanceof Error ? err.message : '修改密码失败')
    } finally {
      setSaving(false)
    }
  }

  const handleClose = () => {
    reset()
    onClose()
  }

  return (
    <AppModal open={isOpen} onClose={handleClose} title="修改密码">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="old_password">当前密码</Label>
          <Input
            id="old_password"
            type="password"
            {...register('old_password')}
          />
          {errors.old_password && (
            <p className="text-sm text-red-500">{errors.old_password.message}</p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="new_password">新密码</Label>
          <Input
            id="new_password"
            type="password"
            {...register('new_password')}
          />
          {errors.new_password && (
            <p className="text-sm text-red-500">{errors.new_password.message}</p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="confirm_password">确认新密码</Label>
          <Input
            id="confirm_password"
            type="password"
            {...register('confirm_password')}
          />
          {errors.confirm_password && (
            <p className="text-sm text-red-500">{errors.confirm_password.message}</p>
          )}
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="outline" onClick={handleClose} disabled={saving}>
            取消
          </Button>
          <Button type="submit" disabled={saving}>
            {saving ? '提交中...' : '确认修改'}
          </Button>
        </div>
      </form>
    </AppModal>
  )
}
