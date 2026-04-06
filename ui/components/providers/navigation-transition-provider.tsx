'use client'

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { usePathname, useRouter } from 'next/navigation'

interface NavigateWithTransitionOptions {
  replace?: boolean
  delayMs?: number
  onBeforeNavigate?: () => void
}

interface NavigationTransitionContextValue {
  isNavigating: boolean
  navigateWithTransition: (href: string, options?: NavigateWithTransitionOptions) => void
}

const EXIT_TRANSITION_DELAY_MS = 90
const OVERLAY_HIDE_DELAY_MS = 100
const FAILSAFE_RESET_DELAY_MS = 1200

const NavigationTransitionContext = createContext<NavigationTransitionContextValue | undefined>(undefined)

export function NavigationTransitionProvider({ children }: { children: ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const [isNavigating, setIsNavigating] = useState(false)
  const isNavigatingRef = useRef(false)
  const navigationTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const overlayTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const failsafeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const clearTimers = useCallback(() => {
    if (navigationTimerRef.current) {
      clearTimeout(navigationTimerRef.current)
      navigationTimerRef.current = null
    }
    if (overlayTimerRef.current) {
      clearTimeout(overlayTimerRef.current)
      overlayTimerRef.current = null
    }
    if (failsafeTimerRef.current) {
      clearTimeout(failsafeTimerRef.current)
      failsafeTimerRef.current = null
    }
  }, [])

  const resetNavigationState = useCallback(() => {
    clearTimers()
    isNavigatingRef.current = false
    setIsNavigating(false)
  }, [clearTimers])

  const navigateWithTransition = useCallback(
    (href: string, options?: NavigateWithTransitionOptions) => {
      if (!href || isNavigatingRef.current) {
        return
      }

      options?.onBeforeNavigate?.()

      isNavigatingRef.current = true
      setIsNavigating(true)
      router.prefetch(href)

      navigationTimerRef.current = setTimeout(() => {
        if (options?.replace) {
          router.replace(href)
        } else {
          router.push(href)
        }
      }, options?.delayMs ?? EXIT_TRANSITION_DELAY_MS)

      failsafeTimerRef.current = setTimeout(() => {
        resetNavigationState()
      }, FAILSAFE_RESET_DELAY_MS)
    },
    [resetNavigationState, router]
  )

  useEffect(() => {
    if (!isNavigatingRef.current) {
      return
    }

    overlayTimerRef.current = setTimeout(() => {
      resetNavigationState()
    }, OVERLAY_HIDE_DELAY_MS)

    return () => {
      if (overlayTimerRef.current) {
        clearTimeout(overlayTimerRef.current)
        overlayTimerRef.current = null
      }
    }
  }, [pathname, resetNavigationState])

  useEffect(() => {
    return () => {
      clearTimers()
    }
  }, [clearTimers])

  const contextValue = useMemo(
    () => ({
      isNavigating,
      navigateWithTransition,
    }),
    [isNavigating, navigateWithTransition]
  )

  return (
    <NavigationTransitionContext.Provider value={contextValue}>
      {children}
      <div
        aria-hidden="true"
        className={`pointer-events-none fixed inset-0 z-[120] bg-slate-950/4 transition-opacity duration-180 ease-in-out ${
          isNavigating ? 'opacity-100' : 'opacity-0'
        }`}
      />
    </NavigationTransitionContext.Provider>
  )
}

export function useNavigationTransition() {
  const context = useContext(NavigationTransitionContext)
  if (!context) {
    throw new Error('useNavigationTransition must be used within NavigationTransitionProvider')
  }
  return context
}

