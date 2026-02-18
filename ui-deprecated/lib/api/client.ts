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
  // 首先尝试从 localStorage 获取
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('auth_token')
    if (token) {
      console.log('ApiClient: Token from localStorage:', token)
      return token
    }

    // 如果 localStorage 中没有，检查 Zustand store
    if (authStore) {
      const state = authStore.getState()
      console.log('ApiClient: Checking Zustand store for token...')
      console.log('ApiClient: Zustand state:', { user: state.user?.username, isAuthenticated: state.isAuthenticated })

      // 如果用户已认证，但 localStorage 中没有令牌，
      // 这可能是由于旧版本登录或令牌被清除
      if (state.isAuthenticated && state.user) {
        console.log('ApiClient: User is authenticated but no token in localStorage')
        console.log('ApiClient: This might be a problem with the login flow')

        // 尝试从 /api/auth/login 获取新令牌（通过刷新机制）
        // 但这需要用户已经登录，所以我们需要实现一个刷新机制

        // 现在，我们暂时返回一个特殊值，表示需要处理认证
        console.log('ApiClient: Returning special NO_TOKEN value for authenticated user')
        return 'NO_TOKEN_NEEDED' // 特殊标记，表示用户已认证但没有令牌
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
        errorMessage = errorData.msg || errorData.message || errorData.error || errorMessage
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

    console.log('ApiClient: Config before token:', config)

    // Add auth token for API requests
    // This solves the cross-port cookie limitation
    if (typeof window !== 'undefined') {
      console.log('ApiClient: Checking localStorage for token...')
      console.log('ApiClient: All localStorage keys:', Object.keys(localStorage))
      console.log('ApiClient: auth-storage content:', localStorage.getItem('auth-storage'))

      const token = await getAuthToken()
      if (token === 'NO_TOKEN_NEEDED') {
        // 用户已认证但没有令牌，抛出错误
        console.error('ApiClient: User is authenticated but no token found in localStorage')
        console.error('ApiClient: This might be due to an old login session')
        throw new ApiError(
          '您的登录状态已过期，请重新登录',
          401,
          'AUTH_TOKEN_MISSING'
        )
      } else if (token) {
        console.log('ApiClient: Adding x-auth-token header:', token)
        config.headers = {
          ...config.headers,
          'x-auth-token': token,
        }
        console.log('ApiClient: Headers after adding token:', config.headers)
      } else {
        console.log('ApiClient: No token found')
        console.log('ApiClient: Available keys:', Object.keys(localStorage))
      }
    }

    console.log('ApiClient: Config after token:', config)

    // 应用请求拦截器
    const processedConfig = this.applyRequestInterceptors(config)

    console.log('ApiClient: Final config:', processedConfig)
    console.log('ApiClient: Final headers:', processedConfig.headers)

    try {
      console.log('ApiClient: Making fetch request to:', url)
      console.log('ApiClient: Request method:', processedConfig.method)
      console.log('ApiClient: Request headers:', processedConfig.headers)
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
