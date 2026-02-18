'use client'

import { useState, useRef, useEffect } from 'react'
import { User as UserIcon, Settings, LogOut } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/store/auth-store'
import { createPortal } from 'react-dom'
import { getContactInfo } from '@/hooks/get-contact-info'
import { User } from '@/lib/api/auth'
import { EditProfileModal } from '@/components/profile/edit-profile-modal'

interface AvatarDropdownProps {
  user: User | null
}

export function AvatarDropdown({ user }: AvatarDropdownProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [isEditProfileOpen, setIsEditProfileOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const router = useRouter()
  const logout = useAuthStore((state) => state.logout)
  const contactInfo = getContactInfo(user || null)
  const closeTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  const handleLogout = async () => {
    await logout()
    router.push('/login')
  }

  const handleMenuClick = (action: string) => {
    setIsOpen(false)
    if (action === 'profile') {
      setIsEditProfileOpen(true)
    }
  }

  const getInitials = (name: string | undefined) => {
    return name?.charAt(0).toUpperCase() || ''
  }

  const handleMouseEnter = () => {
    if (closeTimeoutRef.current) {
      clearTimeout(closeTimeoutRef.current)
      closeTimeoutRef.current = null
    }
    setIsOpen(true)
  }

  const handleMouseLeave = () => {
    closeTimeoutRef.current = setTimeout(() => {
      setIsOpen(false)
    }, 150)
  }

  useEffect(() => {
    return () => {
      if (closeTimeoutRef.current) {
        clearTimeout(closeTimeoutRef.current)
      }
    }
  }, [])

  // 当编辑资料模态框打开时，关闭头像下拉菜单
  useEffect(() => {
    if (isEditProfileOpen) {
      setIsOpen(false)
    }
  }, [isEditProfileOpen])

  return (
    <div
      className="relative"
      ref={dropdownRef}
      onMouseEnter={isEditProfileOpen ? undefined : handleMouseEnter}
      onMouseLeave={isEditProfileOpen ? undefined : handleMouseLeave}
    >
      <button
        className={`flex items-center space-x-2 p-2 rounded-lg transition-all duration-200 ${
          isEditProfileOpen
            ? 'cursor-default'
            : 'hover:bg-slate-100/50 hover:shadow-sm'
        }`}
      >
        <div className="w-8 h-8 rounded-full bg-gradient-to-r from-blue-600 to-indigo-600 flex items-center justify-center text-white font-semibold text-sm shadow-sm">
          {getInitials(user?.name)}
        </div>
      </button>

      {isOpen && !isEditProfileOpen && dropdownRef.current && createPortal(
        <div
          className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-xl border border-slate-200 py-1 z-[60] transition-opacity duration-200"
          style={{
            position: 'fixed',
            top: dropdownRef.current.getBoundingClientRect().bottom + window.scrollY,
            right: 16,
            left: 'auto',
          }}
          onMouseEnter={handleMouseEnter}
          onMouseLeave={handleMouseLeave}
        >
          <div className="px-4 py-2 border-b border-slate-200">
            <p className="text-sm font-medium text-slate-900 truncate">{user?.name || ''}</p>
            <p className="text-xs text-slate-500 truncate">
              {contactInfo?.value || ''}
            </p>
          </div>

          <button
            onClick={() => handleMenuClick('profile')}
            className="w-full px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-100/80 hover:text-slate-900 transition-all duration-150 flex items-center group"
          >
            <UserIcon className="h-4 w-4 mr-2 group-hover:scale-110 transition-transform duration-150" />
            编辑资料
          </button>

          <button
            onClick={() => handleMenuClick('password')}
            className="w-full px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-100/80 hover:text-slate-900 transition-all duration-150 flex items-center group"
          >
            <Settings className="h-4 w-4 mr-2 group-hover:scale-110 transition-transform duration-150" />
            修改密码
          </button>

          <button
            onClick={handleLogout}
            className="w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50 hover:text-red-700 transition-all duration-150 flex items-center group"
          >
            <LogOut className="h-4 w-4 mr-2 group-hover:scale-110 transition-transform duration-150" />
            退出
          </button>
        </div>,
        document.body
      )}

      {/* Edit Profile Modal */}
      {user && (
        <EditProfileModal
          isOpen={isEditProfileOpen}
          onClose={() => setIsEditProfileOpen(false)}
          user={user}
        />
      )}
    </div>
  )
}
