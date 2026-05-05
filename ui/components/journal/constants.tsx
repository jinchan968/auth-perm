import {
  Sun, Cloud, CloudRain, CloudSnow, CloudFog, Wind,
} from 'lucide-react'
import type { Period, Weather } from '@/types/journal'

export const TAG_COLORS = [
  '#6366f1', // 靛蓝
  '#3b82f6', // 蓝
  '#10b981', // 翠绿
  '#f59e0b', // 琥珀
  '#ef4444', // 红
  '#ec4899', // 粉
  '#8b5cf6', // 紫
  '#14b8a6', // 青绿
]

export const PERIODS: { value: Period; label: string; emoji: string; color: string; bgClass: string }[] = [
  { value: '晨', label: '晨', emoji: '🌅', color: '#d97706', bgClass: 'bg-amber-100 text-amber-800 border-amber-200' },
  { value: '上午', label: '上午', emoji: '☀️', color: '#2563eb', bgClass: 'bg-sky-100 text-sky-800 border-sky-200' },
  { value: '下午', label: '下午', emoji: '🌤️', color: '#ea580c', bgClass: 'bg-orange-100 text-orange-800 border-orange-200' },
  { value: '晚', label: '晚', emoji: '🌆', color: '#6d28d9', bgClass: 'bg-violet-100 text-violet-800 border-violet-200' },
  { value: '夜', label: '夜', emoji: '🌙', color: '#1e3a5f', bgClass: 'bg-slate-700 text-slate-100 border-slate-500' },
]

export const WEATHERS: { value: Weather; label: string; icon: React.ReactNode }[] = [
  { value: '晴', label: '晴', icon: <Sun className="h-3.5 w-3.5" /> },
  { value: '多云', label: '多云', icon: <Cloud className="h-3.5 w-3.5" /> },
  { value: '雨', label: '雨', icon: <CloudRain className="h-3.5 w-3.5" /> },
  { value: '雪', label: '雪', icon: <CloudSnow className="h-3.5 w-3.5" /> },
  { value: '雾', label: '雾', icon: <CloudFog className="h-3.5 w-3.5" /> },
  { value: '风', label: '风', icon: <Wind className="h-3.5 w-3.5" /> },
]

export const WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六']

export function charLen(s: string): number {
  return Array.from(s).length
}

export function inferPeriod(): Period {
  const h = new Date().getHours()
  if (h >= 6 && h < 9) return '晨'
  if (h >= 9 && h < 12) return '上午'
  if (h >= 12 && h < 18) return '下午'
  if (h >= 18 && h < 24) return '晚'
  return '夜'
}

export function formatDate(d: Date): string {
  return d.toISOString().slice(0, 10)
}

export function addDays(d: Date, n: number): Date {
  const r = new Date(d)
  r.setDate(r.getDate() + n)
  return r
}

export function formatErrMsg(e: unknown, fallback: string): string {
  if (e instanceof Error) return e.message
  return fallback
}

export const PAGE_SIZE = 20