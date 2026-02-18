'use client'

/**
 * 日志服务
 * 提供统一的日志输出，生产环境自动禁用敏感日志
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error'

interface LogEntry {
  level: LogLevel
  message: string
  timestamp: string
  data?: unknown[]
}

const isProduction = process.env.NODE_ENV === 'production'

/**
 * 简单日志服务
 * 生产环境自动禁用 debug 和 info 级别日志
 */
export class Logger {
  private static logs: LogEntry[] = []
  private static readonly MAX_LOGS = 100

  /**
   * Debug 级别日志（仅开发环境显示）
   */
  static debug(message: string, ...args: unknown[]): void {
    if (isProduction) return
    this.log('debug', message, args)
    console.debug(`[DEBUG] ${message}`, ...args)
  }

  /**
   * Info 级别日志
   */
  static info(message: string, ...args: unknown[]): void {
    if (isProduction) return
    this.log('info', message, args)
    console.info(`[INFO] ${message}`, ...args)
  }

  /**
   * Warning 级别日志
   */
  static warn(message: string, ...args: unknown[]): void {
    this.log('warn', message, args)
    console.warn(`[WARN] ${message}`, ...args)
  }

  /**
   * Error 级别日志
   */
  static error(message: string, ...args: unknown[]): void {
    this.log('error', message, args)
    console.error(`[ERROR] ${message}`, ...args)
  }

  /**
   * 记录日志条目
   */
  private static log(level: LogLevel, message: string, data?: unknown[]): void {
    this.logs.push({
      level,
      message,
      timestamp: new Date().toISOString(),
      data,
    })

    if (this.logs.length > this.MAX_LOGS) {
      this.logs.shift()
    }
  }

  /**
   * 获取所有日志（仅开发环境使用）
   */
  static getLogs(): LogEntry[] {
    return [...this.logs]
  }

  /**
   * 清除所有日志
   */
  static clear(): void {
    this.logs = []
  }

  /**
   * 导出日志（用于错误报告）
   */
  static export(): string {
    return JSON.stringify(this.logs, null, 2)
  }
}

// 便捷方法
export const logger = Logger
