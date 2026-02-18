'use client'

import { Lock } from 'lucide-react'
import { LoginFormUI } from './LoginFormUI'
import { OAuthButtons } from './OAuthButtons'
import { useLoginLogic } from './useLoginLogic'

export function LoginForm() {
  const { isLoading, error, onSubmit, onClearError } = useLoginLogic()

  return (
    <div className="w-full max-w-sm">
      <div className="relative">
        {/* 背景光晕 */}
        <div className="absolute -inset-0.5 bg-gradient-to-r from-blue-600 via-indigo-600 to-slate-600 rounded-2xl blur opacity-20" />

        {/* 登录卡片 */}
        <div className="relative backdrop-blur-xl bg-white/95 dark:bg-slate-900/95 rounded-2xl p-6 border border-slate-200/20 shadow-2xl">
          {/* 标题区域 */}
          <div className="text-center mb-5">
            <div className="inline-flex items-center justify-center w-12 h-12 mx-auto mb-3 rounded-full bg-gradient-to-r from-blue-600 to-indigo-600">
              <Lock className="w-6 h-6 text-white" />
            </div>
            <h1 className="text-2xl font-bold text-slate-800 dark:text-white">
              欢迎回来
            </h1>
            <p className="text-slate-600 dark:text-slate-300 text-xs mt-1">
              登录到您的账户
            </p>
          </div>

          {/* 登录表单 */}
          <LoginFormUI
            isLoading={isLoading}
            error={error}
            onSubmit={onSubmit}
            onClearError={onClearError}
          />

          {/* 分割线 */}
          <div className="relative my-4">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-slate-300" />
            </div>
            <div className="relative flex justify-center text-xs">
              <span className="px-3 bg-white text-slate-500">或者使用第三方账号登录</span>
            </div>
          </div>

          {/* OAuth 按钮 */}
          <OAuthButtons onClearError={onClearError} />

          {/* 注册链接 */}
          <div className="mt-5 text-center">
            <p className="text-slate-600 text-xs">
              还没有账户？
              <a
                href="/register"
                className="ml-1 text-blue-600 hover:text-blue-700 font-medium"
              >
                立即注册
              </a>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
