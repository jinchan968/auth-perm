'use client'

/**
 * Token 存储服务
 * 统一管理所有 Token 相关的 localStorage 操作
 * 消除 token 操作分散在多处的问题
 */
export class TokenStorageService {
  private static readonly TOKEN_KEY = 'auth_token'
  private static readonly TOKEN_EXPIRY_KEY = 'auth_token_expires_at'
  private static readonly THRESHOLD = 5 * 60 * 1000 // 5分钟

  // ==================== Token 存取 ====================

  /**
   * 存储 Token（localStorage + cookie）
   * cookie 用于跨端口共享（如 newshock 在 3001 端口访问）
   */
  static setToken(token: string, expiresAt?: number): void {
    if (typeof window === 'undefined') return
    localStorage.setItem(this.TOKEN_KEY, token)
    if (expiresAt) {
      localStorage.setItem(this.TOKEN_EXPIRY_KEY, expiresAt.toString())
    }
    // 同步写入 cookie，供其他端口的前端（如 newshock）读取
    const maxAge = expiresAt ? Math.floor((expiresAt - Date.now()) / 1000) : 7 * 24 * 60 * 60
    const secure = typeof window !== 'undefined' && window.location.protocol === 'https:' ? '; secure' : ''
    document.cookie = `${this.TOKEN_KEY}=${token}; path=/; max-age=${maxAge}; samesite=lax${secure}`
  }

  /**
   * 获取 Token（优先从 Cookie，获取不到再从 localStorage）
   */
  static getToken(): string | null {
    if (typeof window === 'undefined') return null
    
    // 优先从 Cookie 获取
    const cookieToken = this.getTokenFromCookie()
    if (cookieToken) {
      return cookieToken
    }
    
    // 降级到 localStorage
    return localStorage.getItem(this.TOKEN_KEY)
  }

  /**
   * 从 Cookie 获取 Token
   */
  static getTokenFromCookie(): string | null {
    if (typeof window === 'undefined') return null
    
    const name = process.env.NEXT_PUBLIC_AUTH_COOKIE_NAME || 'auth_token'
    const value = `; ${document.cookie}`
    const parts = value.split(`; ${name}=`)
    
    if (parts.length === 2) {
      const token = parts.pop()?.split(';').shift()
      return token || null
    }
    
    return null
  }

  /**
   * 清除 Token（localStorage 和 Cookie）
   */
  static clearToken(): void {
    if (typeof window === 'undefined') return
    localStorage.removeItem(this.TOKEN_KEY)
    this.clearTokenFromCookie()
  }

  /**
   * 从 Cookie 清除 Token
   */
  static clearTokenFromCookie(): void {
    if (typeof window === 'undefined') return
    
    const name = process.env.NEXT_PUBLIC_AUTH_COOKIE_NAME || 'auth_token'
    document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`
  }

  /**
   * 检查 Token 是否存在
   */
  static hasToken(): boolean {
    return this.getToken() !== null
  }

  // ==================== Token 过期时间 ====================

  /**
   * 设置 Token 过期时间
   */
  static setTokenExpiry(expiresAt: number): void {
    if (typeof window === 'undefined') return
    localStorage.setItem(this.TOKEN_EXPIRY_KEY, expiresAt.toString())
  }

  /**
   * 获取 Token 过期时间
   */
  static getTokenExpiry(): number | null {
    if (typeof window === 'undefined') return null
    const expiry = localStorage.getItem(this.TOKEN_EXPIRY_KEY)
    return expiry ? parseInt(expiry, 10) : null
  }

  /**
   * 清除 Token 过期时间
   */
  static clearTokenExpiry(): void {
    if (typeof window === 'undefined') return
    localStorage.removeItem(this.TOKEN_EXPIRY_KEY)
  }

  // ==================== Token 验证 ====================

  /**
   * 检查 Token 是否即将过期（5分钟内）
   */
  static isTokenExpiringSoon(): boolean {
    const expiresAt = this.getTokenExpiry()
    if (!expiresAt) return false
    return expiresAt - Date.now() < this.THRESHOLD
  }

  /**
   * 检查 Token 是否已过期
   */
  static isTokenExpired(): boolean {
    const expiresAt = this.getTokenExpiry()
    if (!expiresAt) return false
    return Date.now() > expiresAt
  }

  // ==================== 批量操作 ====================

  /**
   * 清除所有认证相关的存储
   */
  static clearAll(): void {
    this.clearToken()
    this.clearTokenExpiry()
  }

  /**
   * 获取完整的认证信息
   */
  static getAuthInfo(): { token: string | null; expiresAt: number | null; isValid: boolean } {
    const token = this.getToken()
    const expiresAt = this.getTokenExpiry()
    const isValid = !!token && (expiresAt ? !this.isTokenExpired() : true)
    return { token, expiresAt, isValid }
  }
}

// 导出便捷方法
export const tokenStorage = TokenStorageService
