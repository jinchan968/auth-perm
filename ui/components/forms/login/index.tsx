'use client'

import { Lock, Sparkles } from 'lucide-react'
import { LoginFormUI } from './LoginFormUI'
import { OAuthButtons } from './OAuthButtons'
import { useLoginLogic } from './useLoginLogic'

export function LoginForm() {
  const { isLoading, error, onSubmit, onClearError } = useLoginLogic()

  return (
    <div className="w-full max-w-sm animate-scale-in">
      <div className="relative">
        {/* 背景光晕 - 增强版 */}
        <div className="absolute -inset-1 bg-gradient-to-r from-primary via-accent to-primary rounded-2xl blur-lg opacity-30 animate-pulse-subtle" />
        <div className="absolute -inset-0.5 bg-gradient-to-r from-primary/50 to-accent/50 rounded-2xl opacity-20" />

        {/* 登录卡片 - 玻璃拟态 */}
        <div className="relative glass-card rounded-2xl p-7 shadow-2xl">
          {/* 装饰角标 */}
          <div className="absolute -top-3 -right-3 w-6 h-6 bg-gradient-to-br from-primary to-accent rounded-full flex items-center justify-center shadow-lg">
            <Sparkles className="w-3 h-3 text-white" />
          </div>

          {/* 标题区域 */}
          <div className="text-center mb-6">
            <div className="inline-flex items-center justify-center w-14 h-14 mx-auto mb-4 rounded-xl bg-gradient-to-br from-primary to-accent shadow-lg shadow-primary/30">
              <Lock className="w-7 h-7 text-white" />
            </div>
            <h1 className="text-2xl font-bold text-foreground">
              欢迎回来
            </h1>
            <p className="text-muted-foreground text-sm mt-1.5">
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

          {/* 分割线 - 增强版 */}
          <div className="relative my-6">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-border" />
            </div>
            <div className="relative flex justify-center text-xs">
              <span className="px-4 bg-background text-muted-foreground rounded-full">
                或者使用第三方账号登录
              </span>
            </div>
          </div>

          {/* OAuth 按钮 */}
          <OAuthButtons onClearError={onClearError} />

          {/* 注册链接 */}
          <div className="mt-6 text-center">
            <p className="text-muted-foreground text-sm">
              还没有账户？
              <a
                href="/register"
                className="ml-1 text-primary hover:text-primary/80 font-medium transition-colors duration-200 hover:underline underline-offset-4"
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
