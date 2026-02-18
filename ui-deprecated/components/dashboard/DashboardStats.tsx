'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function DashboardStats() {
  return (
    <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-6">
      <Card className="backdrop-blur-xl bg-white/95 border-slate-200/20 shadow-lg">
        <CardHeader>
          <CardTitle className="text-lg text-slate-800">登录会话</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-3xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">1</p>
          <p className="text-sm text-slate-600 mt-2">当前活跃会话</p>
        </CardContent>
      </Card>

      <Card className="backdrop-blur-xl bg-white/95 border-slate-200/20 shadow-lg">
        <CardHeader>
          <CardTitle className="text-lg text-slate-800">已授权设备</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-3xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">1</p>
          <p className="text-sm text-slate-600 mt-2">已授权设备数量</p>
        </CardContent>
      </Card>

      <Card className="backdrop-blur-xl bg-white/95 border-slate-200/20 shadow-lg">
        <CardHeader>
          <CardTitle className="text-lg text-slate-800">安全评分</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-3xl font-bold text-green-600">85</p>
          <p className="text-sm text-slate-600 mt-2">安全评分（满分100）</p>
        </CardContent>
      </Card>
    </div>
  )
}
