'use client'

import { authApi } from '@/lib/api/auth'

export interface TokenInfo {
  token: string
  expiresAt: number
  refreshToken?: string
}

/**
 * Token管理器 - 负责Token的自动刷新和过期检测
 * 解决页面刷新后登录态丢失问题
 */
export class TokenManager {
  private static instance: TokenManager
  private refreshPromise: Promise<boolean> | null = null
  private readonly TOKEN_REFRESH_THRESHOLD = 5 * 60 * 1000 // 提前5分钟刷新

  private constructor() {}

  public static getInstance(): TokenManager {
    if (!TokenManager.instance) {
      TokenManager.instance = new TokenManager()
    }
    return TokenManager.instance
  }

  /**
   * 检查Token是否即将过期
   */
  public isTokenExpiringSoon(expiresAt: number): boolean {
    const now = Date.now()
    return expiresAt - now < this.TOKEN_REFRESH_THRESHOLD
  }

  /**
   * 验证当前Token是否有效
   * @returns true-有效, false-无效, null-未知(网络错误等)
   */
  public async validateToken(): Promise<boolean | null> {
    try {
      // 使用getProfile替代validate，因为validate接口需要额外的数据库查询
      // 如果getProfile能返回用户信息，说明token是有效的
      const user = await authApi.getProfile()
      return !!user
    } catch (error) {
      // 网络错误或其他非认证错误
      console.warn('Token validation failed:', error)
      return null
    }
  }

  /**
   * 刷新Token
   * 使用单例模式避免并发刷新
   */
  public async refreshToken(): Promise<boolean> {
    // 如果正在刷新，等待之前的刷新完成
    if (this.refreshPromise) {
      return this.refreshPromise
    }

    this.refreshPromise = this.doRefreshToken()
    const result = await this.refreshPromise
    this.refreshPromise = null
    return result
  }

  private async doRefreshToken(): Promise<boolean> {
    try {
      // 调用刷新接口，后端会从Cookie读取token并刷新
      // 不需要传递refreshToken参数，接口会自动从Cookie读取
      const response = await authApi.refreshToken()
      if (response && response.token) {
        // 刷新成功，Cookie已经由后端更新
        console.log('Token refreshed successfully')
        return true
      }
      return false
    } catch (error) {
      console.warn('Token refresh failed:', error)
      return false
    }
  }

  /**
   * 智能Token检查
   * 先检查是否即将过期，如果过期则尝试刷新
   */
  public async checkAndRefreshToken(expiresAt?: number): Promise<boolean> {
    // 如果没有过期时间信息，直接验证
    if (!expiresAt) {
      const isValid = await this.validateToken()
      return isValid === true
    }

    // 检查是否即将过期
    if (this.isTokenExpiringSoon(expiresAt)) {
      console.log('Token is expiring soon, attempting to refresh...')
      const refreshed = await this.refreshToken()
      if (refreshed) {
        return true
      }
    }

    // 验证当前Token
    const isValid = await this.validateToken()
    return isValid === true
  }

  /**
   * 清除Token相关信息
   */
  public clearToken(): void {
    this.refreshPromise = null
  }
}

// 导出单例实例
export const tokenManager = TokenManager.getInstance()
