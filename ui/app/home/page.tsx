
'use client'

import { ShellLayout } from '@/components/layout/shell-layout'
import { HomeContent } from '@/components/home/home-content'
import { useAuthStore } from '@/store/auth-store'
import { getContactInfo } from '@/hooks/get-contact-info'

export default function HomePage() {
  const { user } = useAuthStore()
  if (!user) return null
  const contactInfo = getContactInfo(user)

  return (
    <ShellLayout pathname="/home">
      <HomeContent user={user} contactInfo={contactInfo} />
    </ShellLayout>
  )
}
