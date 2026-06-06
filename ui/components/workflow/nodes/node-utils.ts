const statusColorMap: Record<string, string> = {
  idle: 'border-slate-200',
  running: 'border-blue-400 animate-pulse',
  success: 'border-green-400',
  failed: 'border-red-400',
  pending: 'border-slate-200',
  cancelled: 'border-slate-300',
}

const nodeTypeFillMap: Record<string, string> = {
  trigger: '#10b981',
  llm: '#3b82f6',
  condition: '#f59e0b',
  transform: '#a855f7',
  merge: '#f43f5e',
  output: '#64748b',
}

const statusStrokeMap: Record<string, string> = {
  running: '#2563eb',
  success: '#16a34a',
  failed: '#dc2626',
}

const DEFAULT_FILL = '#94a3b8'
const DEFAULT_STROKE = '#cbd5e1'

export function getStatusColor(status?: string): string {
  return statusColorMap[status || 'idle']
}

export function getNodeColor(node: { type?: string }): string {
  return node.type ? nodeTypeFillMap[node.type] || DEFAULT_FILL : DEFAULT_FILL
}

export function getNodeStrokeColor(node: { type?: string; data?: unknown; selected?: boolean }): string {
  if (node.selected) return '#0f172a'
  const data = (node.data ?? {}) as { status?: string }
  if (data.status && statusStrokeMap[data.status]) return statusStrokeMap[data.status]
  return DEFAULT_STROKE
}
