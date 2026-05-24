import { apiClient } from './client'
import {
  Entry,
  Tag,
  EntryListResponse,
  TagListResponse,
  CreateEntryRequest,
  AddCorrectionRequest,
  UpdateTagsRequest,
  CreateTagRequest,
  UpdateTagRequest,
  Template,
  TemplateListResponse,
  CreateTemplateRequest,
  UpdateTemplateRequest,
  AIPredictionResponse,
  CreateAIPredictionRequest,
  AIPredictionResult,
  AIPredictionHistoryListResponse,
  AIPredictionModelsResponse,
  AIPredictionHistoryItem,
  AIPredictionDetail,
  AIPredictionQuotasResponse,
} from '@/types/journal'

const BASE = '/journal'

// -------- Tags --------

export async function listTags(tenantId: string): Promise<TagListResponse> {
  return apiClient.get<TagListResponse>(`${BASE}/tags?tenant_id=${tenantId}`)
}

export async function createTag(
  req: CreateTagRequest & { tenant_id: string }
): Promise<Tag> {
  return apiClient.post<Tag>(
    `${BASE}/tags?tenant_id=${req.tenant_id}`,
    req
  )
}

export async function updateTag(
  id: string,
  req: UpdateTagRequest & { tenant_id: string }
): Promise<Tag> {
  return apiClient.put<Tag>(
    `${BASE}/tags/${id}?tenant_id=${req.tenant_id}`,
    req
  )
}

export async function deleteTag(id: string, tenantId: string): Promise<void> {
  await apiClient.delete(`${BASE}/tags/${id}?tenant_id=${tenantId}`)
}

// -------- Entries --------

export async function listEntries(params: {
  tenant_id: string
  start_date?: string
  end_date?: string
  page?: number
  page_size?: number
}): Promise<EntryListResponse> {
  const q = new URLSearchParams()
  q.set('tenant_id', params.tenant_id)
  if (params.start_date) q.set('start_date', params.start_date)
  if (params.end_date) q.set('end_date', params.end_date)
  if (params.page) q.set('page', String(params.page))
  if (params.page_size) q.set('page_size', String(params.page_size))
  return apiClient.get<EntryListResponse>(`${BASE}?${q.toString()}`)
}

export async function getEntry(id: string, tenantId: string): Promise<Entry> {
  return apiClient.get<Entry>(`${BASE}/${id}?tenant_id=${tenantId}`)
}

export async function createEntry(
  req: CreateEntryRequest & { tenant_id: string }
): Promise<Entry> {
  return apiClient.post<Entry>(`${BASE}?tenant_id=${req.tenant_id}`, req)
}

export async function addCorrection(
  id: string,
  req: AddCorrectionRequest & { tenant_id: string }
): Promise<Entry> {
  return apiClient.post<Entry>(`${BASE}/${id}/corrections?tenant_id=${req.tenant_id}`, req)
}

export async function updateTags(
  id: string,
  req: UpdateTagsRequest & { tenant_id: string }
): Promise<Entry> {
  return apiClient.put<Entry>(`${BASE}/${id}/tags?tenant_id=${req.tenant_id}`, req)
}

export async function deleteEntry(id: string, tenantId: string): Promise<void> {
  await apiClient.delete(`${BASE}/${id}?tenant_id=${tenantId}`)
}

// -------- Templates --------

export async function listTemplates(params: {
  tenant_id: string
  page?: number
  page_size?: number
  name?: string
  tag?: string
}): Promise<TemplateListResponse> {
  const q = new URLSearchParams()
  q.set('tenant_id', params.tenant_id)
  if (params.page) q.set('page', String(params.page))
  if (params.page_size) q.set('page_size', String(params.page_size))
  if (params.name) q.set('name', params.name)
  if (params.tag) q.set('tag', params.tag)
  return apiClient.get<TemplateListResponse>(`${BASE}/templates?${q.toString()}`)
}

export async function getTemplate(id: string, tenantId: string): Promise<Template> {
  return apiClient.get<Template>(`${BASE}/templates/${id}?tenant_id=${tenantId}`)
}

export async function createTemplate(
  req: CreateTemplateRequest & { tenant_id: string }
): Promise<Template> {
  const { tenant_id, ...body } = req
  return apiClient.post<Template>(`${BASE}/templates?tenant_id=${tenant_id}`, body)
}

export async function updateTemplate(
  id: string,
  req: UpdateTemplateRequest & { tenant_id: string }
): Promise<Template> {
  const { tenant_id, ...body } = req
  return apiClient.put<Template>(`${BASE}/templates/${id}?tenant_id=${tenant_id}`, body)
}

export async function deleteTemplate(id: string, tenantId: string): Promise<void> {
  await apiClient.delete(`${BASE}/templates/${id}?tenant_id=${tenantId}`)
}

// -------- AI Predictions --------

export async function createAIPrediction(
  req: CreateAIPredictionRequest & { tenant_id: string }
): Promise<AIPredictionResponse> {
  const { tenant_id, ...body } = req
  return apiClient.post<AIPredictionResponse>(`${BASE}/ai-predictions?tenant_id=${tenant_id}`, body)
}

export async function listAIPredictions(params: {
  tenant_id: string
  page?: number
  page_size?: number
}): Promise<AIPredictionHistoryListResponse> {
  const q = new URLSearchParams()
  q.set('tenant_id', params.tenant_id)
  if (params.page !== undefined) q.set('page', String(params.page))
  if (params.page_size !== undefined) q.set('page_size', String(params.page_size))
  return apiClient.get<AIPredictionHistoryListResponse>(`${BASE}/ai-predictions?${q.toString()}`)
}

export async function getAIPrediction(id: string, tenantId: string): Promise<AIPredictionDetail> {
  return apiClient.get<AIPredictionDetail>(`${BASE}/ai-predictions/${id}?tenant_id=${tenantId}`)
}

export async function getAIPredictionModels(tenantId: string): Promise<AIPredictionModelsResponse> {
  return apiClient.get<AIPredictionModelsResponse>(`${BASE}/ai-predictions/models?tenant_id=${tenantId}`)
}

export async function getAIPredictionQuotas(tenantId: string): Promise<AIPredictionQuotasResponse> {
  return apiClient.get<AIPredictionQuotasResponse>(`${BASE}/ai-predictions/quotas?tenant_id=${tenantId}`)
}
