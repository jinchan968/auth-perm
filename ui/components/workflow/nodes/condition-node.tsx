import { Handle, Position, type NodeProps } from '@xyflow/react'
import { GitFork } from 'lucide-react'
import { getStatusColor } from './node-utils'

export default function ConditionNode({ data, selected }: NodeProps) {
  const status = (data as Record<string, string | undefined>)?.status
  const statusColor = getStatusColor(status)

  return (
    <div
      className={`bg-white rounded-lg border-2 p-3 min-w-[200px] shadow-sm ${statusColor} ${
        selected ? 'ring-2 ring-primary' : ''
      }`}
    >
      <div className="flex items-center gap-2 mb-2">
        <GitFork className="h-4 w-4 text-amber-500" />
        <span className="text-xs font-semibold text-slate-700">Condition</span>
      </div>
      <div className="text-xs text-slate-500">条件分支</div>
      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />
    </div>
  )
}
