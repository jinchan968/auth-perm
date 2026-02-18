'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

export default function RolesPage() {
  const router = useRouter()

  // Redirect to the tabbed permissions page with roles tab
  useEffect(() => {
    router.replace('/permissions?tab=roles')
  }, [router])

  return null
}
