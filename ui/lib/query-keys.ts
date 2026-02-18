/**
 * 统一的 Query Keys 管理
 * 确保所有地方使用相同的查询键
 */

export const queryKeys = {
  // Auth
  auth: {
    profile: ['auth', 'profile'] as const,
    sessions: ['auth', 'sessions'] as const,
    devices: ['auth', 'devices'] as const,
    securityLogs: ['auth', 'security', 'logs'] as const,
  },

  // Common
  refresh: ['refresh'] as const,
}

/**
 * 类型安全的查询键生成器
 */
export function createQueryKeys<T extends Record<string, readonly unknown[]>>(
  keys: T
): T {
  return keys
}
