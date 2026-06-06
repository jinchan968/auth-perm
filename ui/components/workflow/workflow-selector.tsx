'use client'

import { useCallback, useEffect, useState } from 'react'
import { FileText, Loader2, Plus, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { listWorkflows } from '@/lib/api/workflow'
import { showError } from '@/lib/toast'
import type { Workflow } from '@/types/workflow'

interface WorkflowSelectorProps {
  tenantId: string
  selectedWorkflowId?: string | null
  refreshKey?: number
  onSelect: (workflow: Workflow) => void
  onNew: () => void
  onWorkflowsChange?: (workflows: Workflow[]) => void
}

export default function WorkflowSelector({
  tenantId,
  selectedWorkflowId,
  refreshKey,
  onSelect,
  onNew,
  onWorkflowsChange,
}: WorkflowSelectorProps) {
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [loading, setLoading] = useState(false)

  const fetchWorkflows = useCallback(async () => {
    if (!tenantId) return
    setLoading(true)
    try {
      const result = await listWorkflows({ tenant_id: tenantId, page: 1, size: 100 })
      const next = result.data || []
      setWorkflows(next)
      onWorkflowsChange?.(next)
    } catch {
      showError('获取工作流列表失败')
    } finally {
      setLoading(false)
    }
  }, [tenantId, onWorkflowsChange])

  useEffect(() => {
    fetchWorkflows()
  }, [fetchWorkflows, refreshKey])

  const handleSelect = (id: string) => {
    const workflow = workflows.find((item) => item.id === id)
    if (workflow) {
      onSelect(workflow)
    }
  }

  return (
    <div className="mb-3 flex flex-col gap-2 rounded-lg border bg-white p-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex min-w-0 flex-1 items-center gap-2">
        <FileText className="h-4 w-4 shrink-0 text-slate-500" />
        <Select value={selectedWorkflowId || ''} onValueChange={handleSelect} disabled={loading || workflows.length === 0}>
          <SelectTrigger className="h-9 min-w-0 flex-1">
            <SelectValue placeholder={loading ? '加载工作流...' : '选择已有工作流'} />
          </SelectTrigger>
          <SelectContent>
            {workflows.map((workflow) => (
              <SelectItem key={workflow.id} value={workflow.id}>
                {workflow.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="flex items-center gap-2">
        <Button size="sm" variant="outline" onClick={fetchWorkflows} disabled={loading}>
          {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
        </Button>
        <Button size="sm" variant="outline" onClick={onNew}>
          <Plus className="h-4 w-4 mr-1" />
          新建
        </Button>
      </div>
    </div>
  )
}
