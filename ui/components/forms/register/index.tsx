'use client'

import { UserPlus, Sparkles } from 'lucide-react'
import { RegisterFormUI } from './RegisterFormUI'
import { useRegisterLogic } from './useRegisterLogic'

export function RegisterForm() {
  const { error, onSubmit, onClearError } = useRegisterLogic()

  return (
    <div className="w-full max-w-sm animate-scale-in">
      <div className="relative">
        {/* 背景光晕 - 增强版 */}
        <div className="absolute -inset-1 bg-gradient-to-r from-primary via-accent to-primary rounded-2xl blur-lg opacity-30 animate-pulse-subtle" />
        <div className="absolute -inset-0.5 bg-gradient-to-r from-primary/50 to-accent/50 rounded-2xl opacity-20" />

        {/* 注册卡片 - 玻璃拟态 */}
        <div className="relative glass-card rounded-2xl p-7 shadow-2xl">
          {/* 装饰角标 */}
          <div className="absolute -top-3 -right-3 w-6 h-6 bg-gradient-to-br from-accent to-purple-400 rounded-full flex items-center justify-center shadow-lg">
            <Sparkles className="w-3 h-3 text-white" />
          </div>

          {/* 标题区域 */}
          <div className="text-center mb-6">
            <div className="inline-flex items-center justify-center w-14 h-14 mx-auto mb-4 rounded-xl bg-gradient-to-br from-accent to-purple-400 shadow-lg shadow-accent/30">
              <UserPlus className="w-7 h-7 text-white" />
            </div>
            <h1 className="text-2xl font-bold text-foreground">
              创建账户
            </h1>
            <p className="text-muted-foreground text-sm mt-1.5">
              注册您的新账户
            </p>
          </div>

          {/* 注册表单 */}
          <RegisterFormUI
            error={error}
            onSubmit={onSubmit}
            onClearError={onClearError}
          />

          {/* 登录链接 */}
          <div className="mt-6 text-center">
            <p className="text-muted-foreground text-sm">
              已有账户？
              <a
                href="/login"
                className="ml-1 text-primary hover:text-primary/80 font-medium transition-colors duration-200 hover:underline underline-offset-4"
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
