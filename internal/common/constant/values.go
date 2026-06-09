package constant

import "time"

// ==================== 通用类型别名 ====================

// HTTPMethod HTTP方法类型
type HTTPMethod string

// LogLevel 日志级别类型
type LogLevel string

// CacheStrategy 缓存策略类型
type CacheStrategy string

// ==================== HTTP相关常量 ====================

// HTTPMethod 常量
const (
	MethodGET     HTTPMethod = "GET"
	MethodPOST    HTTPMethod = "POST"
	MethodPUT     HTTPMethod = "PUT"
	MethodDELETE  HTTPMethod = "DELETE"
	MethodPATCH   HTTPMethod = "PATCH"
	MethodHEAD    HTTPMethod = "HEAD"
	MethodOPTIONS HTTPMethod = "OPTIONS"
)

// ==================== 时间相关常量 ====================

// TokenExpiry Token过期时间
const (
	TokenExpiryDefault    = 24 * time.Hour     // 默认24小时
	TokenExpiryRememberMe = 7 * 24 * time.Hour // 记住我7天
	TokenExpiryShort      = 1 * time.Hour      // 短时间1小时
)

// SessionExpiry 会话过期时间
const (
	SessionExpiryDefault = 24 * time.Hour      // 默认24小时
	SessionExpiryLong    = 30 * 24 * time.Hour // 长期30天
)

// ==================== 分页相关常量 ====================

// DefaultPageSize 默认分页大小
const (
	DefaultPageSize = 10  // 默认每页10条
	MaxPageSize     = 100 // 最大每页100条
	MinPageSize     = 1   // 最小每页1条
)

// ==================== 缓存相关常量 ====================

// CacheKeyPrefix 缓存键前缀
const (
	CacheKeyUser       = "user:"       // 用户缓存前缀
	CacheKeySession    = "session:"    // 会话缓存前缀
	CacheKeyPermission = "permission:" // 权限缓存前缀
	CacheKeyRole       = "role:"       // 角色缓存前缀
)

// CacheTTL 缓存TTL
const (
	CacheTTLShort      = 5 * time.Minute  // 短时间缓存5分钟
	CacheTTLMedium     = 30 * time.Minute // 中等时间缓存30分钟
	CacheTTLLong       = 2 * time.Hour    // 长时间缓存2小时
	CacheTTLPermission = 10 * time.Minute // 权限缓存10分钟
)

// Token常量定义
const (
	// DefaultTokenLength 默认令牌长度
	DefaultTokenLength = 32

	// EmailVerificationTokenLength 邮箱验证令牌长度
	EmailVerificationTokenLength = 64

	// PasswordResetTokenLength 密码重置令牌长度
	PasswordResetTokenLength = 64
)

// ==================== 数据库相关常量 ====================

// DBErrorCode 数据库错误码
const (
	DBErrorCodeDuplicateKey = "23505" // 唯一键冲突
	DBErrorCodeForeignKey   = "23503" // 外键约束
	DBErrorCodeNotNull      = "23502" // 非空约束
)

// ==================== 字符串相关常量 ====================

// DefaultValue 默认值
const (
	DefaultTenantID = "default" // 默认租户ID
	DefaultAvatar   = ""        // 默认头像
	DefaultNickname = ""        // 默认昵称
	EmptyString     = ""        // 空字符串
)

// ==================== 验证相关常量 ====================

// ValidationRegex 验证正则
const (
	RegexPhone    = `^1[3-9]\d{9}$`                                    // 手机号正则
	RegexEmail    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` // 邮箱正则
	RegexUsername = `^[a-zA-Z0-9_]{3,50}$`                             // 用户名正则
)

// ValidationLength 验证长度
const (
	MinUsernameLength = 3   // 最小用户名长度
	MaxUsernameLength = 50  // 最大用户名长度
	MaxNicknameLength = 100 // 最大昵称长度
)

// ReservedUsernames 保留用户名（禁止注册）
var ReservedUsernames = []string{
	"admin", "administrator", "root", "system", "sys",
	"superuser", "super", "user", "guest", "anonymous",
	"owner", "support", "help", "service", "api",
	"www", "mail", "email", "test", "testing",
	"demo", "default", "null", "void", "undefined",
}

// CommonWeakPasswords 常见弱密码（禁止使用）
var CommonWeakPasswords = []string{
	"123456", "password", "123456789", "12345678", "12345",
	"1234567", "1234567890", "qwerty", "abc123", "111111",
	"123123", "admin", "letmein", "welcome", "monkey",
	"1234", "password1", "123", "000000", "iloveyou",
	"1q2w3e4r", "zaq12wsx", "dragon", "sunshine", "princess",
}

// ==================== 日志相关常量 ====================

// LogLevel 常量
const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// ==================== 速率限制相关常量 ====================

const (
	RateLimitDefaultRequests  = 100
	RateLimitDefaultWindow    = 60
	RateLimitAuthRequests     = 10
	RateLimitAuthWindow       = 60
	RateLimitLoginRequests    = 5
	RateLimitLoginWindow      = 300
	RateLimitRegisterRequests = 3
	RateLimitRegisterWindow   = 600
)

// ==================== Cookie相关常量 ====================

const (
	// CookieAuthToken Cookie名称
	CookieAuthToken = "auth_token" // 认证令牌Cookie名称
)

// ==================== CORS相关常量 ====================

const (
	// 允许的源
	Localhost3000       = "http://localhost:3000"
	Localhost8080       = "http://localhost:8080"
	SecureLocalhost3000 = "https://localhost:3000"
	SecureLocalhost8080 = "https://localhost:8080"

	// 允许的HTTP头部
	HeaderOrigin        = "Origin"
	HeaderContentType   = "Content-Type"
	HeaderAccept        = "Accept"
	HeaderAuthorization = "Authorization"
	HeaderAuthToken     = "x-auth-token"
	HeaderTenantID      = "X-Tenant-ID"
	HeaderRequestID     = "X-Request-ID"

	// 暴露的HTTP头部
	ExposeTotalCount = "X-Total-Count"
	ExposeRequestID  = "X-Request-ID"

	// CORS配置
	CORSMaxAge = 24 * time.Hour
)

// ==================== 日志相关常量 ====================

const (
	MaxLogBodySize    = 1024
	UnknownRequestID  = "unknown"
	DefaultStackSize  = 4 << 10
	DetailedStackSize = 8 << 10
)

// ==================== HTTP通用常量 ====================

const (
	StatusCode2xx = 200
	StatusCode4xx = 400
	StatusCode5xx = 500

	BodySize1KB = 1024
	BodySize4KB = 4 << 10
	BodySize8KB = 8 << 10
)
