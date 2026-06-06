import { Save, Play, CheckCircle, RotateCcw, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PermGuard } from '@/components/ui/perm-guard'
import type { Node, Edge } from '@xyflow/react'

interface Props {
  nodes: Node[]
  edges: Edge[]
  setNodes: React.Dispatch<React.SetStateAction<Node[]>>
  setEdges: React.Dispatch<React.SetStateAction<Edge[]>>
  onReset: () => void
  onSave: () => void
  onValidate: () => void
  onRun: () => void
  saving?: boolean
  validating?: boolean
  running?: boolean
  dirty?: boolean
}

export default function WorkflowToolbar({
  onSave,
  onValidate,
  onRun,
  onReset,
  saving,
  validating,
  running,
  dirty,
}: Props) {
  return (
    <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex flex-wrap items-center justify-center gap-1.5 sm:gap-2 bg-white rounded-lg shadow-lg p-2 border">
      <PermGuard button="workflow.write">
        <Button size="sm" variant={dirty ? 'default' : 'outline'} onClick={onSave} disabled={saving}>
          {saving ? <Loader2 className="h-4 w-4 mr-1 animate-spin" /> : <Save className="h-4 w-4 mr-1" />}
          {dirty ? '保存*' : '保存'}
        </Button>
      </PermGuard>
      <Button size="sm" variant="outline" onClick={onValidate} disabled={validating}>
        {validating ? <Loader2 className="h-4 w-4 mr-1 animate-spin" /> : <CheckCircle className="h-4 w-4 mr-1" />}
        校验
      </Button>
      <PermGuard button="workflow.write">
        <Button size="sm" onClick={onRun} disabled={running || saving}>
          {running ? <Loader2 className="h-4 w-4 mr-1 animate-spin" /> : <Play className="h-4 w-4 mr-1" />}
          运行
        </Button>
      </PermGuard>
      <Button size="sm" variant="ghost" onClick={onReset}>
        <RotateCcw className="h-4 w-4 mr-1" />
        重置
      </Button>
    </div>
  )
}
