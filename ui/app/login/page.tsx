import { LoginForm } from "@/components/forms/login"
import { AnimatedBackground } from "@/components/ui/animated-background"

export default function LoginPage() {
  return (
    <div className="relative min-h-screen overflow-hidden bg-gradient-to-br from-slate-50 via-blue-50/30 to-indigo-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-slate-950">
      {/* 动态背景 */}
      <AnimatedBackground />

      {/* 网格背景 */}
      <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxwYXRoIGQ9Ik0wIDBoNjB2NjBIMHoiLz48cGF0aCBkPSJNMzAgMzBoMXYxaC0xek0wIDBoMXYxSDB6IiBmaWxsPSJyZ2JhKDAsIDAsIDAsIDAuMDMpIi8+PC9nPjwvc3ZnPg==')] opacity-50" />

      {/* 主要内容 */}
      <div className="relative z-10 flex min-h-screen items-center justify-center p-4">
        <LoginForm />
      </div>

      {/* 装饰性元素 - 使用新主题色 */}
      <div className="absolute top-10 left-10 w-80 h-80 bg-primary/15 rounded-full mix-blend-multiply filter blur-3xl animate-float" />
      <div className="absolute top-20 right-10 w-72 h-72 bg-accent/15 rounded-full mix-blend-multiply filter blur-3xl animate-float" style={{ animationDelay: '-2s' }} />
      <div className="absolute -bottom-20 left-1/4 w-96 h-96 bg-primary/10 rounded-full mix-blend-multiply filter blur-3xl animate-float" style={{ animationDelay: '-4s' }} />
      <div className="absolute bottom-20 right-1/4 w-64 h-64 bg-accent/10 rounded-full mix-blend-multiply filter blur-3xl animate-pulse-subtle" />

      {/* 顶部渐变遮罩 */}
      <div className="absolute top-0 inset-x-0 h-40 bg-gradient-to-b from-background/50 to-transparent pointer-events-none" />
    </div>
  )
}
