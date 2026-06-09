import { apiClient } from './client'
import type { Workflow, WorkflowListResponse, WorkflowRun, WorkflowRunNode, FlowGraph } from '@/types/workflow'

const BASE = '/workflow'

function withQuery(path: string, params: Record<string, string | number | undefined>) {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== '') {
      query.set(key, String(value))
    }
  })
  const suffix = query.toString()
  return suffix ? `${path}?${suffix}` : path
}

export async function listWorkflows(params: { tenant_id: string; page?: number; size?: number; type?: string }) {
  return apiClient.get<WorkflowListResponse>(withQuery(BASE, {
    tenant_id: params.tenant_id,
    page: params.page,
    size: params.size,
    type: params.type,
  }))
}

export async function createWorkflow(data: {
  tenant_id: string
  name: string
  description?: string
  flow_json: FlowGraph
  template_id?: string
}) {
  return apiClient.post<Workflow>(BASE, data)
}

export async function getWorkflow(id: string, tenant_id: string) {
  return apiClient.get<Workflow>(withQuery(`${BASE}/${encodeURIComponent(id)}`, { tenant_id }))
}

export async function updateWorkflow(
  id: string,
  data: {
    tenant_id: string
    name: string
    description?: string
    flow_json: FlowGraph
    status?: string
  }
) {
  return apiClient.put<Workflow>(`${BASE}/${id}`, data)
}

export async function deleteWorkflow(id: string, tenant_id: string) {
  return apiClient.delete(withQuery(`${BASE}/${encodeURIComponent(id)}`, { tenant_id }))
}

export async function executeWorkflow(
  id: string,
  data: { tenant_id: string; input_text?: string; input_json?: Record<string, unknown> },
  mode: 'sync' | 'async' = 'sync'
) {
  if (mode === 'async') {
    return apiClient.post<{ run_id: string }>(withQuery(`${BASE}/${encodeURIComponent(id)}/execute`, { mode: 'async' }), data)
  }
  return apiClient.post<WorkflowRun>(withQuery(`${BASE}/${encodeURIComponent(id)}/execute`, { mode: 'sync' }), data)
}

export async function validateWorkflow(id: string, tenant_id: string) {
  return apiClient.post<{ valid: boolean; errors: Array<{ node_id?: string; message: string; level: string }> }>(
    withQuery(`${BASE}/${encodeURIComponent(id)}/validate`, { tenant_id }),
    {}
  )
}

export async function cloneWorkflow(id: string, tenant_id: string) {
  return apiClient.post<Workflow>(withQuery(`${BASE}/${encodeURIComponent(id)}/clone`, { tenant_id }), {})
}

export async function listTemplates(tenant_id: string) {
  return apiClient.get<{ data: Workflow[] }>(withQuery(`${BASE}/templates`, { tenant_id }))
}

export async function listRuns(workflowId: string, tenantId: string, params: { page?: number; size?: number }) {
  return apiClient.get<{ data: WorkflowRun[]; total: number; page: number; size: number }>(
    withQuery(`${BASE}/${encodeURIComponent(workflowId)}/runs`, {
      tenant_id: tenantId,
      page: params.page,
      size: params.size,
    })
  )
}

export async function getRun(workflowId: string, runId: string, tenantId: string) {
  return apiClient.get<WorkflowRun>(
    withQuery(`${BASE}/${encodeURIComponent(workflowId)}/runs/${encodeURIComponent(runId)}`, { tenant_id: tenantId })
  )
}

export async function getRunNodes(workflowId: string, runId: string, tenantId: string) {
  return apiClient.get<{ data: WorkflowRunNode[] }>(
    withQuery(`${BASE}/${encodeURIComponent(workflowId)}/runs/${encodeURIComponent(runId)}/nodes`, { tenant_id: tenantId })
  )
}

export async function cancelRun(workflowId: string, runId: string, tenantId: string) {
  return apiClient.post<{ message: string }>(
    withQuery(`${BASE}/${encodeURIComponent(workflowId)}/runs/${encodeURIComponent(runId)}/cancel`, { tenant_id: tenantId }),
    {}
  )
}
