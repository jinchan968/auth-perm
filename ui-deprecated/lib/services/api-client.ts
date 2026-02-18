'use client'

import { ApiError } from '@/lib/api/client'

/**
 * 后端响应包装类型
 */
interface ApiResponse<T> {
  code: number
  msg: string
  data: T
}

/**
 * API 客户端配置
 */
interface ApiClientConfig {
  baseURL?: string
  timeout?: number
}

/**
 * 简化版 API 客户端
 * 统一处理请求和响应，封装 token 注入逻辑
 * 消除 api/client.ts 中的冗余逻辑
 */
export class SimpleApiClient {
  private readonly baseURL: string

  constructor(config: ApiClientConfig = {}) {
    // 始终通过 Next.js API 代理调用后端
    this.baseURL = config.baseURL || '/api'
  }

  /**
   * 构建请求 URL
   */
  private buildUrl(endpoint: string): string {
    const cleanEndpoint = endpoint.replace(/^\/+/, '')
    return `${this.baseURL}/${cleanEndpoint}`
  }

  /**
   * 获取认证 Token
   */
  private getAuthToken(): string | null {
    if (typeof window === 'undefined') return null
    return localStorage.getItem('auth_token')
  }

  /**
   * 发起请求
   */
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = this.buildUrl(endpoint)
    const token = this.getAuthToken()

    const config: RequestInit = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    }

    // 注入认证 Token
    if (token) {
      (config.headers as Record<string, string>)['x-auth-token'] = token
    }

    try {
      const response = await fetch(url, config)
      return await this.handleResponse<T>(response)
    } catch (error) {
      if (error instanceof ApiError) throw error
      throw new ApiError(
        error instanceof Error ? error.message : '网络请求失败',
        0,
        'NETWORK_ERROR'
      )
    }
  }

  /**
   * 处理响应
   */
  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      let errorMessage = '请求失败'
      let errorCode = 'UNKNOWN_ERROR'

      try {
        const errorData = await response.json()
        errorMessage = errorData.msg || errorData.message || errorData.error || errorMessage
        errorCode = errorData.code || errorCode
      } catch {
        errorMessage = `请求失败: ${response.status} ${response.statusText}`
      }

      throw new ApiError(errorMessage, response.status, errorCode)
    }

    // 处理 JSON 响应
    const contentType = response.headers.get('content-type')
    if (contentType && contentType.includes('application/json')) {
      const jsonData = await response.json()
      // 解包 { code, msg, data } 结构
      if (jsonData && typeof jsonData === 'object' && 'data' in jsonData) {
        return (jsonData as ApiResponse<T>).data as T
      }
      return jsonData as T
    }

    return {} as T
  }

  // ==================== 公开方法 ====================

  async get<T>(endpoint: string, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET', ...options })
  }

  async post<T>(endpoint: string, data?: unknown, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    })
  }

  async put<T>(endpoint: string, data?: unknown, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    })
  }

  async patch<T>(endpoint: string, data?: unknown, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    })
  }

  async delete<T>(endpoint: string, options?: RequestInit): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE', ...options })
  }
}

// 导出默认实例
export const apiClient = new SimpleApiClient()

// 导出类型
export type { ApiResponse, ApiClientConfig }
