'use client'

export function AnimatedBackground() {
  return (
    <div className="absolute inset-0 overflow-hidden">
      {/* 商务风格背景 - 深蓝灰色系 */}
      <div className="absolute inset-0 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900" />

      {/* 网格背景 */}
      <div
        className="absolute inset-0 opacity-5"
        style={{
          backgroundImage: `
            linear-gradient(rgba(148, 163, 184, 0.4) 1px, transparent 1px),
            linear-gradient(90deg, rgba(148, 163, 184, 0.4) 1px, transparent 1px)
          `,
          backgroundSize: '50px 50px',
        }}
      />

      {/* 静态几何图形 - 商务风格 */}
      <div className="absolute top-1/4 left-1/4 w-64 h-64 border border-slate-700/30 rounded-full" />
      <div className="absolute top-1/3 right-1/4 w-48 h-48 border border-slate-700/30 rotate-45" />
      <div className="absolute bottom-1/4 left-1/3 w-56 h-56 border border-slate-700/30 rounded-lg" />
    </div>
  )
}
