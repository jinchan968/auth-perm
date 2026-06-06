'use client'

import { useEffect, useState, useCallback } from 'react'
import { Clock, CheckCircle, XCircle, Loader2, WifiOff, Wifi } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { useTenant } from '@/lib/tenant-context'
import { listRuns } from '@/lib/api/workflow'
import { showError } from '@/lib/toast'
import { useWorkflowWS } from '@/hooks/use-workflow-ws'
import { tokenStorage } from '@/lib/services/token-storage'
import type { Workflow, WorkflowRun } from '@/types/workflow'
import WorkflowSelector from './workflow-selector'

interface WorkflowRunsProps {
  workflowId?: string
}

export default function WorkflowRuns({ workflowId }: WorkflowRunsProps) {
  const { tenantId } = useTenant()
  const [selectedWorkflowId, setSelectedWorkflowId] = useState<string | null>(workflowId || null)
  const [runs, setRuns] = useState<WorkflowRun[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setSelectedWorkflowId(workflowId || null)
  }, [workflowId])

  const effectiveWorkflowId = selectedWorkflowId

  const fetchRuns = useCallback(async () => {
    if (!tenantId || !effectiveWorkflowId) {
      setRuns([])
      return
    }
    setLoading(true)
    try {
      const result = await listRuns(effectiveWorkflowId, tenantId, { page: 1, size: 20 })
      setRuns(result.data || [])
    } catch {
      showError('获取运行记录失败')
    } finally {
      setLoading(false)
    }
  }, [tenantId, effectiveWorkflowId])

  useEffect(() => {
    fetchRuns()
  }, [fetchRuns])

  const getActiveRunId = (): string | null => {
    const running = runs.find((r) => r.status === 'running' || r.status === 'pending')
    return running?.id ?? null
  }

  const token = typeof window !== 'undefined' ? tokenStorage.getToken() : null
  const { connected } = useWorkflowWS({
    runId: getActiveRunId(),
    tenantId: tenantId || '',
    token: token || '',
    onMessage: () => {
      fetchRuns()
    },
  })

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />
      case 'running':
        return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />
      case 'pending':
        return <Clock className="h-4 w-4 text-amber-400" />
      case 'cancelled':
        return <XCircle className="h-4 w-4 text-slate-400" />
      default:
        return <Clock className="h-4 w-4 text-slate-400" />
    }
  }

  const getStatusLabel = (status: string) => {
    const map: Record<string, string> = {
      pending: '等待中',
      running: '运行中',
      success: '成功',
      failed: '失败',
      cancelled: '已取消',
    }
    return map[status] || status
  }

  const handleSelectWorkflow = (workflow: Workflow) => {
    setSelectedWorkflowId(workflow.id)
  }

  if (!effectiveWorkflowId) {
    return (
      <div className="space-y-4">
        <WorkflowSelector
          tenantId={tenantId}
          selectedWorkflowId={selectedWorkflowId}
          onSelect={handleSelectWorkflow}
          onNew={() => setSelectedWorkflowId(null)}
        />
        <div className="text-center py-12 text-slate-400">
          <Clock className="h-12 w-12 mx-auto mb-3 opacity-50" />
          <p>请选择工作流查看运行历史</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <WorkflowSelector
        tenantId={tenantId}
        selectedWorkflowId={effectiveWorkflowId}
        onSelect={handleSelectWorkflow}
        onNew={() => setSelectedWorkflowId(null)}
      />
      <div className="flex items-center gap-2 text-xs text-slate-400">
        {connected ? (
          <><Wifi className="h-3.5 w-3.5 text-green-500" /> <span>实时连接中</span></>
        ) : (
          <><WifiOff className="h-3.5 w-3.5 text-slate-300" /> <span>离线模式</span></>
        )}
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-slate-400" />
        </div>
      ) : runs.length === 0 ? (
        <div className="text-center py-12 text-slate-400">
          <Clock className="h-12 w-12 mx-auto mb-3 opacity-50" />
          <p>暂无运行记录</p>
        </div>
      ) : (
        runs.map((run) => (
          <Card key={run.id} className="hover:shadow-md transition-shadow cursor-pointer">
            <CardContent className="py-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  {getStatusIcon(run.status)}
                  <div>
                    <p className="text-sm font-medium">{run.id.slice(0, 8)}</p>
                    <p className="text-xs text-slate-400">
                      {run.started_at ? new Date(run.started_at).toLocaleString() : '—'}
                    </p>
                  </div>
                </div>
                <div className="text-right flex flex-col items-end gap-0.5">
                  <span className="text-xs font-medium text-slate-600">{getStatusLabel(run.status)}</span>
                  <span className="text-xs text-slate-400">{run.duration_ms}ms</span>
                </div>
              </div>
            </CardContent>
          </Card>
        ))
      )}
    </div>
  )
}
