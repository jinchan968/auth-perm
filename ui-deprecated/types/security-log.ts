/**
 * Security Log Type Definitions
 * 安全日志类型定义
 */

/**
 * Audit log entry - 审计日志条目
 */
export interface AuditLogEntry {
  /** 日志ID */
  id: string
  /** 租户ID */
  tenantId: string
  /** 用户ID */
  userId: string
  /** 操作类型 (login, logout, create_session, refresh_token, etc.) */
  action: string
  /** 资源类型 */
  resourceType: string
  /** 资源ID */
  resourceId: string
  /** 操作前的旧值 */
  oldValues: AuditLogValues | null
  /** 操作后的新值 */
  newValues: AuditLogValues | null
  /** IP地址 */
  ipAddress: string
  /** 用户代理 */
  userAgent: string
  /** 操作是否成功 */
  success: boolean
  /** 错误信息 */
  errorMessage: string
  /** 创建时间 */
  createdAt: string
}

/**
 * Audit log values - 审计日志值变化
 */
export interface AuditLogValues {
  /** 变更的字段 */
  changedFields?: Record<string, unknown>
  /** 上下文信息 */
  context?: Record<string, unknown>
  /** 元数据 */
  metadata?: Record<string, unknown>
  /** 标签 */
  tags?: string[]
  /** IP地址 */
  ipAddress?: string
  /** 用户代理 */
  userAgent?: string
  /** 会话ID */
  sessionId?: string
  /** 请求ID */
  requestId?: string
  /** 关联ID */
  correlationId?: string
}

/**
 * Login log filters - 登录日志过滤条件
 */
export interface LoginLogFilters {
  /** 操作类型 */
  action?: string
  /** 开始日期 (RFC3339格式) */
  startDate?: string
  /** 结束日期 (RFC3339格式) */
  endDate?: string
  /** 搜索关键词 (IP, UserAgent) */
  search?: string
  /** 页码 */
  page?: number
  /** 每页数量 */
  pageSize?: number
}

/**
 * Login logs list response - 登录日志列表响应
 */
export interface LoginLogsResponse {
  /** 日志列表 */
  logs: AuditLogEntry[]
  /** 总数量 */
  total: number
  /** 当前页码 */
  page: number
  /** 每页数量 */
  pageSize: number
}

/**
 * Login log stats - 登录日志统计
 */
export interface LoginLogStats {
  /** 操作统计 */
  actionStats: Record<string, number>
  /** 错误统计 */
  errorStats: Record<string, number>
  /** 成功/失败统计 */
  successStats: {
    success: number
    failed: number
  }
  /** 资源统计 */
  resourceStats: Record<string, number>
  /** 今日登录数 */
  todayLogins: number
  /** 今日失败数 */
  todayFailed: number
}

/**
 * All login logs response (for admin) - 所有登录日志响应
 */
export interface AllLoginLogsResponse extends LoginLogsResponse {
  /** 日志列表 (扩展字段) */
  logs: AuditLogEntry[]
}

/**
 * Security log item for dashboard - 仪表盘安全日志项
 */
export interface SecurityLogItem {
  /** 动作描述 */
  action: string
  /** 动作图标 */
  icon: 'login' | 'logout' | 'warning' | 'success' | 'error'
  /** 动作颜色 */
  color: 'green' | 'red' | 'yellow' | 'blue' | 'gray'
  /** 时间 */
  time: string
  /** 详细信息 */
  details: string
  /** IP地址 */
  ipAddress?: string
  /** 位置 (可选) */
  location?: string
  /** 设备信息 (可选) */
  device?: string
}
