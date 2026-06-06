import { memo } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import { Wrench } from 'lucide-react'
import { getStatusColor } from './node-utils'

const opLabels: Record<string, string> = {
  regex_extract: '正则提取',
  regex_replace: '正则替换',
  trim: '去空白',
  markdown_to_text: '去Markdown',
  extract_json: '提取JSON',
  truncate: '截断',
  to_uppercase: '转大写',
  to_lowercase: '转小写',
}

export default memo(function TransformNode({ data, selected }: NodeProps) {
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
        {d.operation ? opLabels[d.operation] || d.operation : '未配置'}
      </div>
      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />
    </div>
  )
})