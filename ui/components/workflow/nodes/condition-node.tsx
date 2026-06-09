import { memo } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import { GitFork } from 'lucide-react'
import { getStatusColor } from './node-utils'

export default memo(function ConditionNode({ data, selected }: NodeProps) {
  const d = data as Record<string, string | undefined>
  const statusColor = getStatusColor(d.status)
  const branches = (d.branches as Array<{ handle: string; label?: string }> | undefined) || []
  const defaultHandle = (d.default_handle as string | undefined) || 'default'
  const handleCount = branches.length + 1
  const minHeight = Math.max(72, 42 + handleCount * 24)

  const handles = [
    ...branches.map((b, i) => (
      <Handle
        key={b.handle}
        type="source"
        position={Position.Right}
        id={b.handle}
        style={{ top: `${30 + i * 24}px` }}
      />
    )),
    <Handle
      key={defaultHandle}
      type="source"
      position={Position.Right}
      id={defaultHandle}
      style={{ top: `${30 + branches.length * 24}px` }}
    />,
  ]

  return (
    <div
      className={`bg-white rounded-lg border-2 p-3 min-w-[200px] shadow-sm ${statusColor} ${
        selected ? 'ring-2 ring-primary' : ''
      }`}
      style={{ minHeight }}
    >
      <div className="flex items-center gap-2 mb-2">
        <GitFork className="h-4 w-4 text-amber-500" />
        <span className="text-xs font-semibold text-slate-700">Condition</span>
      </div>
      <div className="text-xs text-slate-500">
        {branches.length > 0
          ? branches.map((b) => b.label || b.handle).join(' / ') + (defaultHandle ? ` / ${defaultHandle}` : '')
          : '条件分支'}
      </div>
      <Handle type="target" position={Position.Left} />
      {handles}
    </div>
  )
})
