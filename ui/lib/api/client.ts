const IS_SERVER = typeof window === 'undefined'
const API_BASE_URL = IS_SERVER
  ? process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'
  : '/api'

// API 错误类
export class ApiError extends Error {
  status: number
  code?: string
  details?: any

  constructor(message: string, status: number, code?: string, details?: any) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }
}

// 导入 Zustand store
let authStore: any = null
if (typeof window !== 'undefined') {
  try {
    // 动态导入以避免 SSR 问题
    const storeModule = require('@/store/auth-store')
    authStore = storeModule.useAuthStore
  } catch (e) {
    console.warn('Failed to import auth store:', e)
  }
}

/**
 * 获取认证令牌
 * 优先从 localStorage 获取，其次从 Zustand store 获取
 */
async function getAuthToken(): Promise<string | null> {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('auth_token')
    if (token) {
      return token
    }

    if (authStore) {
      const state = authStore.getState()
      if (state.isAuthenticated && state.user) {
        return 'NO_TOKEN_NEEDED'
      }
    }
  }

  return null
}

// HTTP 状态码
const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  UNPROCESSABLE_ENTITY: 422,
  INTERNAL_SERVER_ERROR: 500,
} as const

// 请求拦截器
const requestInterceptors: Array<(config: RequestInit) => RequestInit> = []
// 响应拦截器
const responseInterceptors: Array<(response: Response) => Response> = []

// 添加请求拦截器
export const addRequestInterceptor = (interceptor: (config: RequestInit) => RequestInit) => {
  requestInterceptors.push(interceptor)
}

// 添加响应拦截器
export const addResponseInterceptor = (interceptor: (response: Response) => Response) => {
  responseInterceptors.push(interceptor)
}

// 公开路由列表 - 这些路由收到认证错误时不重定向
const PUBLIC_ROUTES = ['/login', '/register', '/forgot-password', '/reset-password', '/unauthorized']

// 处理 401/403 鉴权响应的函数
function handleAuthRedirect(response: Response): Response {
  // 只在客户端执行
  if (typeof window === 'undefined') {
    return response
  }

  const currentPath = window.location.pathname
  const isPublicRoute = PUBLIC_ROUTES.some(route => currentPath.startsWith(route))

  if (isPublicRoute) {
    return response
  }

  if (response.status === HTTP_STATUS.UNAUTHORIZED) {
    window.location.href = '/login'
    return response
  }

  if (response.status === HTTP_STATUS.FORBIDDEN) {
    if (currentPath !== '/unauthorized') {
      window.location.href = '/unauthorized'
    }
  }

  return response
}

// 注册鉴权跳转处理拦截器
addResponseInterceptor(handleAuthRedirect)

// 通用 API 客户端类
class ApiClient {
  private readonly baseURL: string
  private readonly defaultHeaders: Record<string, string>

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL
    this.defaultHeaders = {
      'Content-Type': 'application/json',
    }
  }

  private buildUrl(endpoint: string): string {
    // 移除 endpoint 开头的斜杠，避免双斜杠
    const cleanEndpoint = endpoint.replace(/^\/+/, '')
    return `${this.baseURL}/${cleanEndpoint}`
  }

  private applyRequestInterceptors(config: RequestInit): RequestInit {
    return requestInterceptors.reduce((acc, interceptor) => interceptor(acc), config)
  }

  private async applyResponseInterceptors(response: Response): Promise<Response> {
    return responseInterceptors.reduce(async (acc, interceptor) => {
      const resolved = await acc
      return interceptor(resolved)
    }, Promise.resolve(response))
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    // 应用响应拦截器
    const processedResponse = await this.applyResponseInterceptors(response)

    // 检查响应状态
    if (!processedResponse.ok) {
      let errorMessage = '请求失败'
      let errorCode = 'UNKNOWN_ERROR'
      let errorDetails = null

      try {
        const errorData = await processedResponse.json()
        // 后端返回格式: { code, msg, error, data }
        let msg = errorData.msg || errorData.message || ''
        const detail = errorData.error || ''
        if (msg && detail && msg !== detail) {
          errorMessage = `${msg}: ${detail}`
        } else {
          errorMessage = msg || detail || errorMessage
        }
        errorCode = errorData.code || errorCode
        errorDetails = errorData.details || errorData
      } catch (e) {
        // 如果解析 JSON 失败，使用默认消息
        errorMessage = `请求失败: ${processedResponse.status} ${processedResponse.statusText}`
      }

      throw new ApiError(errorMessage, processedResponse.status, errorCode, errorDetails)
    }

    // 处理空响应
    const contentType = processedResponse.headers.get('content-type')
    if (contentType && contentType.includes('application/json')) {
      const jsonData = await processedResponse.json()
      // 后端响应被包装在 { code, msg, data } 结构中，需要解包
      if (jsonData && typeof jsonData === 'object' && 'data' in jsonData) {
        return jsonData.data as T
      }
      return jsonData as T
    }

    return {} as T
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = this.buildUrl(endpoint)

    // 合并配置
    const config: RequestInit = {
      headers: {
        ...this.defaultHeaders,
        ...options.headers,
      },
      ...options,
    }

    if (typeof window !== 'undefined') {
      const token = await getAuthToken()
      if (token === 'NO_TOKEN_NEEDED') {
        throw new ApiError(
          '您的登录状态已过期，请重新登录',
          401,
          'AUTH_TOKEN_MISSING'
        )
      } else if (token) {
        config.headers = {
          ...config.headers,
          'x-auth-token': token,
        }
      }
    }

    const processedConfig = this.applyRequestInterceptors(config)

    try {
      const response = await fetch(url, processedConfig)
      return await this.handleResponse<T>(response)
    } catch (error) {
      if (error instanceof ApiError) {
        throw error
      }

      // 网络错误或其他错误
      throw new ApiError(
        error instanceof Error ? error.message : '网络请求失败',
        0,
        'NETWORK_ERROR'
      )
    }
  }

  // GET 请求
  async get<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'GET',
      ...options,
    })
  }

  // POST 请求
  async post<T>(
    endpoint: string,
    data?: any,
    options: RequestInit = {}
  ): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      headers: {
        ...options.headers,
      },
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  // PUT 请求
  async put<T>(
    endpoint: string,
    data?: any,
    options: RequestInit = {}
  ): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      headers: {
        ...options.headers,
      },
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  // PATCH 请求
  async patch<T>(
    endpoint: string,
    data?: any,
    options: RequestInit = {}
  ): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PATCH',
      headers: {
        ...options.headers,
      },
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  // DELETE 请求
  async delete<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'DELETE',
      ...options,
    })
  }

  // 设置默认请求头
  setDefaultHeader(key: string, value: string): void {
    this.defaultHeaders[key] = value
  }

  // 移除默认请求头
  removeDefaultHeader(key: string): void {
    delete this.defaultHeaders[key]
  }

  // 获取基础 URL
  getBaseURL(): string {
    return this.baseURL
  }
}

// 创建默认实例
export const apiClient = new ApiClient()

// 创建带认证的客户端实例
export const createAuthClient = (token: string): ApiClient => {
  const client = new ApiClient()
  client.setDefaultHeader('Authorization', `Bearer ${token}`)
  return client
}

// 导出常量
export { HTTP_STATUS }
