/**
 * 验证 redirect URL 是否安全（防止开放重定向攻击）
 * 允许：相对路径、同 hostname 的 newshock 端口、主 UI 端口
 */
export function isSafeRedirect(url: string): boolean {
  // 相对路径（不以 // 开头）
  if (url.startsWith('/') && !url.startsWith('//')) {
    return true
  }
  try {
    const parsed = new URL(url, window.location.origin)
    const mainPort = process.env.NEXT_PUBLIC_MAIN_PORT || '3000'
    const allowedHosts = [
      `${window.location.hostname}:3001`,
      `${window.location.hostname}:${mainPort}`,
      window.location.host,
    ]
    return allowedHosts.includes(parsed.host)
  } catch {
    return false
  }
}
