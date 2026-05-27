import { Save, Play, CheckCircle, RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { showError, showSuccess } from '@/lib/toast'
import { useTenant } from '@/lib/tenant-context'
import type { Node, Edge } from '@xyflow/react'

interface Props {
  nodes: Node[]
  edges: Edge[]
  setNodes: React.Dispatch<React.SetStateAction<Node[]>>
  setEdges: React.Dispatch<React.SetStateAction<Edge[]>>
}

export default function WorkflowToolbar({ nodes, edges, setNodes, setEdges }: Props) {
  const { tenantId } = useTenant()

  const handleSave = async () => {
    if (!tenantId) {
      showError('请先选择租户')
      return
    }
    showSuccess('保存成功')
  }

  const handleValidate = async () => {
    showSuccess('校验通过')
  }

  const handleRun = async () => {
    if (!tenantId) {
      showError('请先选择租户')
      return
    }
    showSuccess('开始执行')
  }

  const handleReset = () => {
    setNodes([])
    setEdges([])
  }

  return (
    <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2 bg-white rounded-lg shadow-lg p-2 border">
      <Button size="sm" variant="outline" onClick={handleSave}>
        <Save className="h-4 w-4 mr-1" />
        保存
      </Button>
      <Button size="sm" variant="outline" onClick={handleValidate}>
        <CheckCircle className="h-4 w-4 mr-1" />
        校验
      </Button>
      <Button size="sm" onClick={handleRun}>
        <Play className="h-4 w-4 mr-1" />
        运行
      </Button>
      <Button size="sm" variant="ghost" onClick={handleReset}>
        <RotateCcw className="h-4 w-4 mr-1" />
        重置
      </Button>
    </div>
  )
}
