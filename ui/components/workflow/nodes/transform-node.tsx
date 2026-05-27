import { Handle, Position, type NodeProps } from '@xyflow/react'
import { Wrench } from 'lucide-react'
import { getStatusColor } from './node-utils'

export default function TransformNode({ data, selected }: NodeProps) {
  const d = data as Record<string, string | undefined>
  const statusColor = getStatusColor(d.status)

  return (
    <div
      className={`bg-white rounded-lg border-2 p-3 min-w-[200px] shadow-sm ${statusColor} ${
        selected ? 'ring-2 ring-primary' : ''
      }`}
    >
      <div className="flex items-center gap-2 mb-2">
        <Wrench className="h-4 w-4 text-purple-500" />
        <span className="text-xs font-semibold text-slate-700">Transform</span>
      </div>
      <div className="text-xs text-slate-500 truncate">
        {d.operation || '未配置'}
      </div>
      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />
    </div>
  )
}
