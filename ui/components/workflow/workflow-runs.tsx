'use client'

import { useEffect, useState, useCallback } from 'react'
import { Clock, CheckCircle, XCircle, Loader2 } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { useTenant } from '@/lib/tenant-context'
import { listRuns } from '@/lib/api/workflow'
import type { WorkflowRun } from '@/types/workflow'

interface WorkflowRunsProps {
  workflowId?: string
}

export default function WorkflowRuns({ workflowId }: WorkflowRunsProps) {
  const { tenantId } = useTenant()
  const [runs, setRuns] = useState<WorkflowRun[]>([])
  const [loading, setLoading] = useState(false)

  const fetchRuns = useCallback(async () => {
    if (!tenantId || !workflowId) return
    setLoading(true)
    try {
      const result = await listRuns(workflowId, tenantId, { page: 1, size: 20 })
      setRuns(result.data || [])
    } finally {
      setLoading(false)
    }
  }, [tenantId, workflowId])

  useEffect(() => {
    fetchRuns()
  }, [tenantId, workflowId, fetchRuns])

  if (!workflowId) {
    return (
      <div className="text-center py-12 text-slate-400">
        <Clock className="h-12 w-12 mx-auto mb-3 opacity-50" />
        <p>请先选择一个工作流</p>
      </div>
    )
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />
      case 'running':
        return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />
      default:
        return <Clock className="h-4 w-4 text-slate-400" />
    }
  }

  return (
    <div className="space-y-4">
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
                <div className="text-right">
                  <span className="text-xs text-slate-500">{run.duration_ms}ms</span>
                </div>
              </div>
            </CardContent>
          </Card>
        ))
      )}
    </div>
  )
}
