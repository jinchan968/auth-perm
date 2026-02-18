'use client'

import { ApiError } from '@/lib/api/client'

/**
 * 错误类型枚举
 */
export enum ErrorType {
  NETWORK_ERROR = 'NETWORK_ERROR',
  AUTH_ERROR = 'AUTH_ERROR',
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  SERVER_ERROR = 'SERVER_ERROR',
  UNKNOWN_ERROR = 'UNKNOWN_ERROR',
}

/**
 * 分类后的错误信息
 */
export interface ErrorInfo {
  type: ErrorType
  title: string
  message: string
  retryable: boolean
  action?: string
}

/**
 * 友好的错误消息映射
 */
const ERROR_MESSAGES: Record<number | string, { title: string; message: string; action?: string }> = {
  0: { title: '网络错误', message: '无法连接到服务器，请检查网络连接', action: '请稍后重试' },
  400: { title: '请求错误', message: '提交的请求有误，请检查输入内容' },
  401: { title: '登录过期', message: '您的登录状态已过期，请重新登录', action: '点击重新登录' },
  403: { title: '权限不足', message: '您没有权限执行此操作' },
  404: { title: '未找到', message: '请求的资源不存在' },
  422: { title: '数据验证错误', message: '提交的数据不符合要求' },
  500: { title: '服务器错误', message: '服务器内部错误，请稍后重试', action: '联系管理员' },
  502: { title: '服务不可用', message: '服务器暂时无法响应', action: '请稍后重试' },
  NETWORK_ERROR: { title: '网络错误', message: '网络连接失败，请检查网络设置', action: '点击重试' },
}

/**
 * 统一错误处理器
 * 提供分类错误和友好提示
 */
export class ErrorHandler {
  /**
   * 分类错误并返回友好的错误信息
   */
  static classify(error: unknown): ErrorInfo {
    if (error instanceof ApiError) {
      const status = error.status
      if (ERROR_MESSAGES[status]) {
        const config = ERROR_MESSAGES[status]
        return {
          type: this.getErrorType(status),
          title: config.title,
          message: error.message || config.message,
          retryable: status >= 500 || status === 0,
          action: config.action,
        }
      }
      return {
        type: this.getErrorType(status),
        title: '操作失败',
        message: error.message || '发生了未知错误',
        retryable: status >= 500 || status === 0,
      }
    }

    if (error instanceof TypeError && error.message.includes('fetch')) {
      return {
        type: ErrorType.NETWORK_ERROR,
        title: ERROR_MESSAGES.NETWORK_ERROR.title,
        message: ERROR_MESSAGES.NETWORK_ERROR.message,
        retryable: true,
        action: ERROR_MESSAGES.NETWORK_ERROR.action,
      }
    }

    return {
      type: ErrorType.UNKNOWN_ERROR,
      title: '发生错误',
      message: error instanceof Error ? error.message : '发生了未知错误',
      retryable: false,
    }
  }

  /**
   * 根据 HTTP 状态码获取错误类型
   */
  private static getErrorType(status: number): ErrorType {
    if (status === 401 || status === 403) return ErrorType.AUTH_ERROR
    if (status === 400 || status === 422) return ErrorType.VALIDATION_ERROR
    if (status >= 500) return ErrorType.SERVER_ERROR
    if (status === 0) return ErrorType.NETWORK_ERROR
    return ErrorType.UNKNOWN_ERROR
  }

  /**
   * 获取用户友好的错误消息
   */
  static getUserMessage(error: unknown): string {
    const info = this.classify(error)
    return info.action ? `${info.message}（${info.action}）` : info.message
  }

  /**
   * 判断错误是否可重试
   */
  static isRetryable(error: unknown): boolean {
    return this.classify(error).retryable
  }

  /**
   * 判断是否为认证错误
   */
  static isAuthError(error: unknown): boolean {
    return this.classify(error).type === ErrorType.AUTH_ERROR
  }
}

// 导出便捷方法
export const getErrorInfo = ErrorHandler.classify
export const getUserMessage = ErrorHandler.getUserMessage
export const isRetryableError = ErrorHandler.isRetryable
export const isAuthError = ErrorHandler.isAuthError
