import { User } from '@/lib/api/auth'

export interface ContactInfo {
  type: string
  value: string
}

/**
 * 获取用户联系方式显示信息的自定义Hook
 * @param user 用户对象
 * @returns 联系方式信息（类型和值）
 */
export function getContactInfo(user: User | null): ContactInfo | null {
  if (!user) return null

  const isPhoneLogin = user.profile?.phone && user.profile.phone !== ''
  const isEmailLogin = user.email && user.email.includes('@')

  if (isPhoneLogin && !isEmailLogin) {
    // 手机号登录
    return {
      type: '手机',
      value: user.profile?.phone || '',
    }
  } else if (isEmailLogin) {
    // 邮箱登录
    return {
      type: '邮箱',
      value: user.email,
    }
  }
  return null
}
