package constant

import "time"

// ==================== 登录相关常量 ====================

// Cookie配置常量
const (
	// CookieExpiryRememberMe Cookie过期时间（秒）
	CookieExpiryRememberMe = 7 * 24 * 60 * 60 // 7天
	CookieExpiryDefault    = 24 * 60 * 60     // 1天

	// CookiePath Cookie配置
	CookiePath     = "/"
	CookieDomain   = ""
	CookieSecure   = false
	CookieHTTPOnly = true

	// LoginSuccessMessage 消息常量
	LoginSuccessMessage    = "登录成功"
	RegisterSuccessMessage = "注册成功"
	LogoutSuccessMessage   = "登出成功"
)

// GetSessionExpiry 获取会话过期时间
func GetSessionExpiry(rememberMe bool) time.Duration {
	if rememberMe {
		return TokenExpiryRememberMe
	}
	return TokenExpiryDefault
}

// GetCookieExpiry 获取Cookie过期时间（秒）
func GetCookieExpiry(rememberMe bool) int {
	if rememberMe {
		return CookieExpiryRememberMe
	}
	return CookieExpiryDefault
}
