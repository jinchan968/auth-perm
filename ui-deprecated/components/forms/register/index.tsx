'use client'

import { Lock } from 'lucide-react'
import { RegisterFormUI } from './RegisterFormUI'
import { useRegisterLogic } from './useRegisterLogic'

export function RegisterForm() {
  const { error, onSubmit, onClearError } = useRegisterLogic()

  return (
    <div className="w-full max-w-sm">
      <div className="relative">
        <div className="absolute -inset-0.5 bg-gradient-to-r from-blue-600 via-indigo-600 to-slate-600 rounded-2xl blur opacity-20" />

        <div className="relative backdrop-blur-xl bg-white/95 dark:bg-slate-900/95 rounded-2xl p-6 border border-slate-200/20 shadow-2xl">
          <div className="text-center mb-5">
            <div className="inline-flex items-center justify-center w-12 h-12 mx-auto mb-3 rounded-full bg-gradient-to-r from-blue-600 to-indigo-600">
              <Lock className="w-6 h-6 text-white" />
            </div>
            <h1 className="text-2xl font-bold text-slate-800 dark:text-white">
              创建账户
            </h1>
            <p className="text-slate-600 dark:text-slate-300 text-xs mt-1">
              注册您的新账户
            </p>
          </div>

          <RegisterFormUI
            error={error}
            onSubmit={onSubmit}
            onClearError={onClearError}
          />

          <div className="mt-5 text-center">
            <p className="text-slate-600 text-xs">
              已有账户？
              <a
                href="/login"
                className="ml-1 text-blue-600 hover:text-blue-700 font-medium"
              >
                立即登录
              </a>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
