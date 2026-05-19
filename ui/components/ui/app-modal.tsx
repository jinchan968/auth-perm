'use client'

/**
 * AppModal — 全局通用 Modal 组件
 *
 * 特性：
 * - createPortal 挂到 document.body，完全脱离父级层叠上下文
 * - 统一的浅色毛玻璃 backdrop（bg-black/20 + backdrop-blur-[2px]）
 * - 进入/退出动画
 * - 点击 backdrop 关闭
 * - 支持自定义最大宽度（maxWidth）
 */

import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

interface AppModalProps {
  open: boolean
  onClose: () => void
  title?: string
  children: React.ReactNode
  /** Tailwind max-w-* class，默认 max-w-md */
  maxWidth?: string
  /** 隐藏默认标题栏（自行渲染 header） */
  hideHeader?: boolean
}

export function AppModal({
  open,
  onClose,
  title,
  children,
  maxWidth = 'max-w-md',
  hideHeader = false,
}: AppModalProps) {
  const [mounted, setMounted] = useState(false)
  const [visible, setVisible] = useState(false)
  // 用于在退出动画播放完毕前保持 portal 挂载
  const [rendered, setRendered] = useState(false)

  useEffect(() => {
    setMounted(true)
    return () => setMounted(false)
  }, [])

  useEffect(() => {
    if (open) {
      setRendered(true)
      // 下一帧触发进入动画
      const t = setTimeout(() => setVisible(true), 10)
      return () => clearTimeout(t)
    } else {
      // 先触发退出动画
      setVisible(false)
      // 等动画播放完（300ms）再卸载
      const t = setTimeout(() => setRendered(false), 300)
      return () => clearTimeout(t)
    }
  }, [open])

  // ESC 关闭
  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open, onClose])

  if (!rendered || !mounted) return null

  return createPortal(
    <>
      {/* Backdrop — fixed 挂到 body，覆盖含 sticky header 在内的整个视口 */}
      <div
        className={cn(
          'fixed inset-0 z-[100] transition-all duration-300 ease-out',
          visible
            ? 'bg-black/20 backdrop-blur-[2px]'
            : 'bg-black/0 backdrop-blur-none pointer-events-none',
        )}
        onClick={onClose}
      />

      {/* 居中容器 */}
      <div
        className={cn(
          'fixed inset-0 z-[101] flex items-center justify-center pointer-events-none p-4',
          'transition-all duration-300 ease-out',
          visible ? 'opacity-100' : 'opacity-0',
        )}
      >
        {/* Modal 卡片 */}
        <div
          className={cn(
            'pointer-events-auto w-full bg-white rounded-2xl shadow-2xl',
            'transition-all duration-300 ease-out transform origin-center',
            maxWidth,
            visible
              ? 'opacity-100 scale-100 translate-y-0'
              : 'opacity-0 scale-95 translate-y-4',
          )}
        >
          {/* 默认 Header */}
          {!hideHeader && (
            <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200">
              {title && (
                <h2 className="text-lg font-semibold text-slate-900">{title}</h2>
              )}
              <button
                onClick={onClose}
                className="ml-auto p-1.5 text-slate-400 hover:text-slate-600 rounded-full hover:bg-slate-100/80 transition-all duration-200 hover:scale-110"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          )}

          {/* 内容区 */}
          <div>{children}</div>
        </div>
      </div>
    </>,
    document.body,
  )
}
