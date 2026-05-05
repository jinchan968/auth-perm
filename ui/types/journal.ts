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
