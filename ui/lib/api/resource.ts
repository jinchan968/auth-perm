import { apiClient } from './client'

export interface ResourceItem {
  id: string
  permission_id: string
  resource_id: string
  resource_type: 'menu' | 'button' | 'api_path' | 'field' | 'other'
  resource_name: string
  tenant_id: string
  created_at: string
  updated_at: string
}

export interface MyResourcesResponse {
  resources: ResourceItem[]
}

/**
 * 获取当前用户可访问的资源清单
 * 用于前端权限控制（菜单、按钮显隐）
 */
export async function getMyResources(): Promise<ResourceItem[]> {
  const response = await apiClient.get<MyResourcesResponse>('/auth/my-resources')
  return response.resources || []
}
