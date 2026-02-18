'use client'

import { ApiError } from '@/lib/api/client'

/**
 * 错误分类枚举
 */
export enum ErrorType {
  NETWORK_ERROR = 'NETWORK_ERROR',        // 网络错误
  AUTH_ERROR = 'AUTH_ERROR',              // 认证错误(401/403)
  VALIDATION_ERROR = 'VALIDATION_ERROR',  // 参数验证错误(400)
  SERVER_ERROR = 'SERVER_ERROR',          // 服务器错误(5xx)
  UNKNOWN_ERROR = 'UNKNOWN_ERROR'         // 未知错误
}

/**
 * 错误分类结果
 */
export interface ClassifiedError {
  type: ErrorType
  isRetryable: boolean
  message: string
  originalError: unknown
}

/**
 * 错误分类器
 * 将不同类型的错误分类，以便采取不同的处理策略
 */
export class ErrorClassifier {
  /**
   * 分类错误
   */
  public static classify(error: unknown): ClassifiedError {
    // ApiError类型
    if (error instanceof ApiError) {
      return this.classifyApiError(error)
    }

    // 网络错误 (TypeError通常表示网络问题)
    if (error instanceof TypeError && error.message.includes('fetch')) {
      return {
        type: ErrorType.NETWORK_ERROR,
        isRetryable: true,
        message: '网络连接失败，请检查网络后重试',
        originalError: error
      }
    }

    // 其他未知错误
    return {
      type: ErrorType.UNKNOWN_ERROR,
      isRetryable: false,
      message: '发生未知错误，请稍后重试',
      originalError: error
    }
  }

  /**
   * 分类ApiError
   */
  private static classifyApiError(error: ApiError): ClassifiedError {
    const status = error.status

    // 401 Unauthorized - 认证错误
    if (status === 401) {
      return {
        type: ErrorType.AUTH_ERROR,
        isRetryable: false,
        message: '认证已过期，请重新登录',
        originalError: error
      }
    }

    // 403 Forbidden - 权限错误
    if (status === 403) {
      return {
        type: ErrorType.AUTH_ERROR,
        isRetryable: false,
        message: '权限不足，无法访问此资源',
        originalError: error
      }
    }

    // 400 Bad Request - 参数验证错误
    if (status === 400) {
      return {
        type: ErrorType.VALIDATION_ERROR,
        isRetryable: false,
        message: error.message || '请求参数错误',
        originalError: error
      }
    }

    // 5xx Server Error - 服务器错误
    if (status >= 500 && status < 600) {
      return {
        type: ErrorType.SERVER_ERROR,
        isRetryable: true,
        message: '服务器开小差，请稍后重试',
        originalError: error
      }
    }

    // 网络错误 (status为0或code为NETWORK_ERROR)
    if (status === 0 || error.code === 'NETWORK_ERROR') {
      return {
        type: ErrorType.NETWORK_ERROR,
        isRetryable: true,
        message: '网络连接失败，请检查网络后重试',
        originalError: error
      }
    }

    // 其他错误
    return {
      type: ErrorType.UNKNOWN_ERROR,
      isRetryable: false,
      message: error.message || '操作失败，请稍后重试',
      originalError: error
    }
  }

  /**
   * 判断是否为网络错误
   */
  public static isNetworkError(error: unknown): boolean {
    return this.classify(error).type === ErrorType.NETWORK_ERROR
  }

  /**
   * 判断是否为认证错误
   */
  public static isAuthError(error: unknown): boolean {
    return this.classify(error).type === ErrorType.AUTH_ERROR
  }

  /**
   * 判断是否可重试
   */
  public static isRetryable(error: unknown): boolean {
    return this.classify(error).isRetryable
  }

  /**
   * 获取用户友好的错误消息
   */
  public static getErrorMessage(error: unknown): string {
    return this.classify(error).message
  }
}
