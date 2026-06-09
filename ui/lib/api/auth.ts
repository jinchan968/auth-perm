import { apiClient, ApiError } from './client'
import { ErrorClassifier, ErrorType } from '@/lib/auth/error-classifier'

// ===== 类型定义 =====

export interface LoginRequest {
  identifier: string
  password: string
}

export interface RegisterRequest {
  identifier_type: 'email' | 'phone'
  email?: string
  phone?: string
  username: string
  password: string
  confirm_password: string
  invite_code: string
}

// 后端登录响应结构 (LoginResponse)
export interface AuthResponse {
  username: string
  nickname: string
  avatar: string
  status: string
  account_id: string
  email_verified: boolean
  token: string
  message: string
  expires_at: string
}

export interface User {
  id: string
  email: string
  username: string
  name: string
  roles: string[]
  tenant_id?: string
  avatar?: string
  profile?: {
    phone?: string
    bio?: string
    createdAt?: string
    updatedAt?: string
    identifierType?: string
    identifierValue?: string
    status?: string
  }
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
  confirm_password: string
}

export interface ForgotPasswordRequest {
  email: string
}

export interface ResetPasswordRequest {
  token: string
  newPassword: string
  confirmPassword: string
}

export interface OAuthProvider {
  id: 'github' | 'google' | 'wechat'
  name: string
  icon: string
  color: string
}

export interface OAuthUrlResponse {
  url: string
  state?: string
}

export interface OAuthCallbackResponse {
  token: string
  refreshToken?: string
  user: User
  isNewUser?: boolean
}

export interface SessionInfo {
  id: string
  deviceInfo: {
    browser: string
    os: string
    device: string
    ip: string
  }
  location?: string
  lastActivity: string
  isCurrent: boolean
}

export interface DeviceInfo {
  id: string
  name: string
  browser: string
  os: string
  ip: string
  location?: string
  lastActivity: string
  isTrusted: boolean
}

// ===== 认证 API =====

export const authApi = {
  // Generic get method for any endpoint
  get: async <T>(endpoint: string): Promise<T> => {
    return apiClient.get<T>(endpoint)
  },

  // Note: login and register are now handled by internal Next.js API routes
  // to facilitate setting HttpOnly cookies. These functions might be deprecated
  // if all calls are migrated to the internal routes.
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    return apiClient.post<AuthResponse>('/auth/public/login', data)
  },

  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    return apiClient.post<AuthResponse>('/auth/public/register', data)
  },

  // This is called by the server-side logout route.
  // The client-side logout just calls `fetch('/api/auth/logout')`
  logout: async (): Promise<void> => {
    return apiClient.post('/auth/logout', {})
  },
  
  // This is now handled by the cookie's maxAge. Refresh token logic might need re-evaluation.
  refreshToken: async (refreshToken?: string): Promise<AuthResponse> => {
    // 后端会从Cookie读取token，不需要传递参数
    return apiClient.post<AuthResponse>('/auth/refresh', {})
  },

  // User information (no token needed from client)
  getProfile: async (): Promise<User> => {
    return apiClient.get<User>('/auth/profile')
  },

  updateProfile: async (data: Partial<User>): Promise<User> => {
    return apiClient.patch<User>('/auth/profile', data)
  },

  // Password management
  changePassword: async (data: ChangePasswordRequest): Promise<void> => {
    return apiClient.post('/auth/change-password', data)
  },

  forgotPassword: async (data: ForgotPasswordRequest): Promise<void> => {
    return apiClient.post('/auth/forgot-password', data)
  },

  resetPassword: async (data: ResetPasswordRequest): Promise<void> => {
    return apiClient.post('/auth/reset-password', data)
  },

  // OAuth (these are public or handled via redirects and proxies)
  getOAuthUrl: async (provider: string): Promise<OAuthUrlResponse> => {
    return apiClient.get<OAuthUrlResponse>(`/auth/oauth/${provider}/url`)
  },

  handleOAuthCallback: async (
    provider: string,
    code: string,
    state?: string
  ): Promise<OAuthCallbackResponse> => {
    return apiClient.post<OAuthCallbackResponse>(
      `/auth/oauth/${provider}/callback`,
      { code, state }
    )
  },

  // 2FA Management
  enable2FA: async (): Promise<{ qrCode: string; secret: string }> => {
    return apiClient.post('/auth/2fa/enable', {})
  },

  verify2FA: async (code: string): Promise<void> => {
    return apiClient.post('/auth/2fa/verify', { code })
  },

  disable2FA: async (code: string): Promise<void> => {
    return apiClient.post('/auth/2fa/disable', { code })
  },

  // Session Management
  getSessions: async (): Promise<SessionInfo[]> => {
    return apiClient.get<SessionInfo[]>('/auth/sessions')
  },

  revokeSession: async (sessionId: string): Promise<void> => {
    return apiClient.delete(`/auth/sessions/${sessionId}`)
  },

  revokeAllSessions: async (): Promise<void> => {
    return apiClient.delete('/auth/sessions/all')
  },

  // Device Management
  getDevices: async (): Promise<DeviceInfo[]> => {
    return apiClient.get<DeviceInfo[]>('/auth/devices')
  },

  revokeDevice: async (deviceId: string): Promise<void> => {
    return apiClient.delete(`/auth/devices/${deviceId}`)
  },

  trustDevice: async (deviceId: string): Promise<void> => {
    return apiClient.post(`/auth/devices/${deviceId}/trust`, {})
  },

  untrustDevice: async (deviceId: string): Promise<void> => {
    return apiClient.delete(`/auth/devices/${deviceId}/trust`)
  },

  // Account Security
  getSecurityLogs: async (): Promise<any[]> => {
    return apiClient.get<any[]>('/auth/security/logs')
  },

  // Validate token (now cookie based)
  validateToken: async (): Promise<{ valid: boolean; user?: User }> => {
    return apiClient.get('/auth/validate')
  },
}

// ===== OAuth 提供商配置 =====

export const OAUTH_PROVIDERS: OAuthProvider[] = [
  {
    id: 'wechat',
    name: '微信',
    icon: 'wechat',
    color: '#07C160',
  },
  {
    id: 'google',
    name: 'Google',
    icon: 'google',
    color: '#4285F4',
  },
  {
    id: 'github',
    name: 'GitHub',
    icon: 'github',
    color: '#24292E',
  },
]

// ===== 工具函数 =====

// 解析 OAuth 错误
export const parseOAuthError = (error: string): string => {
  const errorMap: Record<string, string> = {
    'access_denied': '用户取消了授权',
    'invalid_request': '无效的请求参数',
    'unauthorized_client': '客户端未授权',
    'unsupported_response_type': '不支持的响应类型',
    'invalid_scope': '无效的权限范围',
    'server_error': '服务器内部错误',
    'temporarily_unavailable': '服务暂时不可用',
  }

  return errorMap[error] || 'OAuth 授权失败'
}

// 检查错误是否为认证错误
export const isAuthError = (error: unknown): boolean => {
  if (error instanceof ApiError) {
    return error.status === 401 || error.status === 403
  }
  return false
}

// 检查错误是否为网络错误
export const isNetworkError = (error: unknown): boolean => {
  if (error instanceof ApiError) {
    return error.code === 'NETWORK_ERROR' || error.status === 0
  }
  return false
}

/**
 * 带认证的API调用 - 自动处理401错误
 * 如果收到401错误，自动尝试刷新token后重试
 */
export const callAuthApi = async <T>(
  apiCall: () => Promise<T>,
  retryOnAuthError: boolean = true
): Promise<T> => {
  try {
    return await apiCall()
  } catch (error) {
    // 如果是401错误且允许重试
    if (retryOnAuthError && isAuthError(error)) {
      try {
        await authApi.refreshToken()
        return await apiCall()
      } catch (refreshError) {
        console.error('Token refresh failed:', refreshError)
        // 刷新失败，抛出原始错误
        throw error
      }
    }

    // 非认证错误或其他情况，直接抛出
    throw error
  }
}

/**
 * 安全获取用户信息
 * 使用智能错误处理，网络错误不会清除认证状态
 */
export const safeGetProfile = async (): Promise<User | null> => {
  try {
    const user = await authApi.getProfile()
    return user
  } catch (error) {
    const classifiedError = ErrorClassifier.classify(error)

    // 网络错误或服务器错误，返回null但不抛出
    if (classifiedError.type === ErrorType.NETWORK_ERROR || classifiedError.type === ErrorType.SERVER_ERROR) {
      console.warn('Temporary error getting profile:', classifiedError.message)
      return null
    }

    // 认证错误，抛出错误
    if (classifiedError.type === ErrorType.AUTH_ERROR) {
      console.warn('Auth error getting profile:', classifiedError.message)
      throw error
    }

    // 其他错误，抛出
    throw error
  }
}
