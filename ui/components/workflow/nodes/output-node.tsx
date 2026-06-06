import { memo } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import { Workflow } from 'lucide-react'
import { getStatusColor } from './node-utils'

const formatLabels: Record<string, string> = {
  raw: '原始文本',
  json: 'JSON',
  markdown: 'Markdown',
}

export default memo(function OutputNode({ data, selected }: NodeProps) {
  const d = data as Record<string, string | undefined>
  const statusColor = getStatusColor(d.status)

  return (
    <div
      className={`bg-white rounded-lg border-2 p-3 min-w-[200px] shadow-sm ${statusColor} ${
        selected ? 'ring-2 ring-primary' : ''
      }`}
    >
      <div className="flex items-center gap-2 mb-2">
        <Workflow className="h-4 w-4 text-slate-500" />
        <span className="text-xs font-semibold text-slate-700">Output</span>
      </div>
      <div className="text-xs text-slate-500 truncate">
        {d.format ? formatLabels[d.format] : '未配置'}
      </div>
      <Handle type="target" position={Position.Left} />
    </div>
  )
})