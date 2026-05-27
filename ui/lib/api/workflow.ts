import { apiClient } from './client'
import type { Workflow, WorkflowListResponse, WorkflowRun, WorkflowRunNode, FlowGraph } from '@/types/workflow'

const BASE = '/workflow'

export async function listWorkflows(params: { tenant_id: string; page?: number; size?: number; type?: string }) {
  const query = new URLSearchParams()
  query.set('tenant_id', params.tenant_id)
  if (params.page) query.set('page', String(params.page))
  if (params.size) query.set('size', String(params.size))
  if (params.type) query.set('type', params.type)
  return apiClient.get<WorkflowListResponse>(`${BASE}?${query.toString()}`)
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
  return apiClient.get<Workflow>(`${BASE}/${id}?tenant_id=${tenant_id}`)
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
  return apiClient.delete(`${BASE}/${id}?tenant_id=${tenant_id}`)
}

export async function executeWorkflow(
  id: string,
  data: { tenant_id: string; input_text?: string; input_json?: Record<string, unknown> },
  mode: 'sync' | 'async' = 'sync'
) {
  if (mode === 'async') {
    return apiClient.post<{ run_id: string }>(`${BASE}/${id}/execute?mode=async`, data)
  }
  return apiClient.post<WorkflowRun>(`${BASE}/${id}/execute?mode=sync`, data)
}

export async function validateWorkflow(id: string, tenant_id: string) {
  return apiClient.post<{ valid: boolean; errors: Array<{ node_id?: string; message: string; level: string }> }>(
    `${BASE}/${id}/validate?tenant_id=${tenant_id}`,
    {}
  )
}

export async function cloneWorkflow(id: string, tenant_id: string) {
  return apiClient.post<Workflow>(`${BASE}/${id}/clone?tenant_id=${tenant_id}`, {})
}

export async function listTemplates(tenant_id: string) {
  return apiClient.get<{ data: Workflow[] }>(`${BASE}/templates?tenant_id=${tenant_id}`)
}

export async function listRuns(workflowId: string, tenantId: string, params: { page?: number; size?: number }) {
  const query = new URLSearchParams()
  query.set('tenant_id', tenantId)
  if (params.page) query.set('page', String(params.page))
  if (params.size) query.set('size', String(params.size))
  return apiClient.get<{ data: WorkflowRun[]; total: number; page: number; size: number }>(
    `${BASE}/${workflowId}/runs?${query.toString()}`
  )
}

export async function getRun(workflowId: string, runId: string, tenantId: string) {
  return apiClient.get<WorkflowRun>(`${BASE}/${workflowId}/runs/${runId}?tenant_id=${tenantId}`)
}

export async function getRunNodes(workflowId: string, runId: string, tenantId: string) {
  return apiClient.get<{ data: WorkflowRunNode[] }>(`${BASE}/${workflowId}/runs/${runId}/nodes?tenant_id=${tenantId}`)
}

export async function cancelRun(workflowId: string, runId: string, tenantId: string) {
  return apiClient.post<{ message: string }>(`${BASE}/${workflowId}/runs/${runId}/cancel?tenant_id=${tenantId}`, {})
}
