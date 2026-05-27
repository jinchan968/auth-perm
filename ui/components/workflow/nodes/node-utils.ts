const statusColorMap: Record<string, string> = {
  idle: 'border-slate-200',
  running: 'border-blue-400 animate-pulse',
  success: 'border-green-400',
  error: 'border-red-400',
}

export function getStatusColor(status?: string): string {
  return statusColorMap[status || 'idle']
}
