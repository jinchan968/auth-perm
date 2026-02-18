import { useEffect, useState } from 'react'

/**
 * 页面hydration状态检查的Hook
 * @returns hydration状态
 */
export function useHydration() {
  const [isHydrated, setIsHydrated] = useState(false)

  useEffect(() => {
    setIsHydrated(true)
  }, [])

  return { isHydrated }
}
