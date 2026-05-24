export type Period = '晨' | '上午' | '下午' | '晚' | '夜'

export type Weather = '晴' | '多云' | '雨' | '雪' | '雾' | '风'

export interface Tag {
  id: string
  tenant_id: string
  account_id: string
  name: string
  color: string
  created_at: string
  updated_at: string
}

export interface Correction {
  id: string
  content: string
  created_at: string
}

export interface Entry {
  id: string
  tenant_id: string
  account_id: string
  parent_id?: string
  title?: string
  content: string
  weather?: Weather
  location?: string
  period: Period
  entry_date: string
  tags?: Tag[]
  corrections?: Correction[]
  created_at: string
  updated_at: string
}

export interface EntryListResponse {
  data: Entry[]
  total: number
  page: number
  page_size: number
}

export interface TagListResponse {
  data: Tag[]
}

export interface CreateEntryRequest {
  title?: string
  content: string
  weather?: Weather
  location?: string
  period: Period
  entry_date: string
  tag_ids?: string[]
}

export interface AddCorrectionRequest {
  content: string
}

export interface UpdateTagsRequest {
  tag_ids: string[]
}

export interface CreateTagRequest {
  name: string
  color?: string
}

export interface UpdateTagRequest {
  name?: string
  color?: string
}

// -------- Template --------

export interface Template {
  id: string
  tenant_id: string
  account_id: string
  name: string
  content?: string
  tags: string[]
  created_at: string
  updated_at: string
}

export interface TemplateListResponse {
  data: Template[]
  total: number
  page: number
  page_size: number
}

export interface CreateTemplateRequest {
  name: string
  content?: string
  tags?: string[]
}

export interface UpdateTemplateRequest {
  name?: string
  content?: string
  tags?: string[]
}

// -------- AI Prediction --------

export interface AIPredictionResult {
  model_id: string
  model_name: string
  content: string
  duration_ms: number
  error?: string
}

export interface CreateAIPredictionRequest {
  question: string
  system_prompt?: string
  models?: string[]
  reasoning_mode?: 'low' | 'medium' | 'high'
}

export interface AIPredictionResponse {
  results: AIPredictionResult[]
  prediction_id: string
}

export interface AIPredictionHistoryItem {
  id: string
  question: string
  created_at: string
}

export interface AIPredictionHistoryListResponse {
  data: AIPredictionHistoryItem[]
  total: number
  page: number
  page_size: number
}

export interface AIModelInfo {
  id: string
  name: string
}

export interface AIPredictionDetail {
  id: string
  question: string
  system_prompt: string
  reasoning_mode: string
  results: AIPredictionResult[]
  model_snapshot: string[]
  created_at: string
}

export interface AIPredictionModelsResponse {
  defaults: string[]
  replaceable: AIModelInfo[]
  all: AIModelInfo[]
  default_system_prompt: string
  daily_limit: number
}

export interface AIPredictionQuotasResponse {
  daily_limit: number
  remaining: Record<string, number>
}
