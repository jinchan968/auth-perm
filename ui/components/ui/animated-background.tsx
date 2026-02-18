'use client'

export function AnimatedBackground() {
  return (
    <div className="absolute inset-0 overflow-hidden">
      {/* 现代渐变背景 */}
      <div className="absolute inset-0 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900" />

      {/* 网格背景 - 增强版 */}
      <div
        className="absolute inset-0 opacity-[0.03]"
        style={{
          backgroundImage: `
            linear-gradient(hsl(var(--primary)) 1px, transparent 1px),
            linear-gradient(90deg, hsl(var(--primary)) 1px, transparent 1px)
          `,
          backgroundSize: '60px 60px',
        }}
      />

      {/* 动态渐变光斑 */}
      <div className="absolute top-0 left-1/4 w-[500px] h-[500px] bg-primary/20 rounded-full mix-blend-screen filter blur-[100px] animate-float" />
      <div className="absolute bottom-0 right-1/4 w-[400px] h-[400px] bg-accent/15 rounded-full mix-blend-screen filter blur-[80px] animate-float" style={{ animationDelay: '-3s' }} />
      <div className="absolute top-1/2 left-0 w-[300px] h-[300px] bg-primary/10 rounded-full mix-blend-screen filter blur-[60px] animate-pulse-subtle" />

      {/* 几何装饰 - 现代风格 */}
      <div className="absolute top-1/4 left-1/4 w-72 h-72 border border-white/5 rounded-full animate-pulse-subtle" />
      <div className="absolute top-1/3 right-1/4 w-56 h-56 border border-white/5 rotate-45 animate-float" style={{ animationDelay: '-2s' }} />
      <div className="absolute bottom-1/4 left-1/3 w-64 h-64 border border-white/5 rounded-2xl animate-float" style={{ animationDelay: '-4s' }} />

      {/* 光线效果 */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-px h-[40%] bg-gradient-to-b from-primary/30 via-primary/10 to-transparent" />
      <div className="absolute top-1/4 right-0 w-[30%] h-px bg-gradient-to-l from-accent/20 via-accent/5 to-transparent" />
    </div>
  )
}
