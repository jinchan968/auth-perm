import { RegisterForm } from "@/components/forms/register"
import { AnimatedBackground } from "@/components/ui/animated-background"

export default function RegisterPage() {
  return (
    <div className="relative min-h-screen overflow-hidden">
      {/* 动态背景 */}
      <AnimatedBackground />

      {/* 主要内容 */}
      <div className="relative z-10 flex min-h-screen items-center justify-center p-4">
        <RegisterForm />
      </div>

      {/* 装饰性元素 - 商务蓝色系 */}
      <div className="absolute top-10 left-10 w-72 h-72 bg-blue-600/20 rounded-full mix-blend-multiply filter blur-xl"></div>
      <div className="absolute top-10 right-10 w-72 h-72 bg-indigo-600/20 rounded-full mix-blend-multiply filter blur-xl"></div>
      <div className="absolute -bottom-8 left-20 w-72 h-72 bg-slate-600/20 rounded-full mix-blend-multiply filter blur-xl"></div>
    </div>
  )
}
