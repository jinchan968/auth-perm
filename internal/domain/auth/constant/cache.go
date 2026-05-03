package constant

// CacheKey 缓存键常量
const (
	// 用户缓存键前缀
	UserCacheKeyPrefix = "auth:user:"

	// 账户缓存键前缀
	AccountCacheKeyPrefix = "auth:account:"

	// 会话缓存键前缀
	SessionCacheKeyPrefix = "auth:session:"

	// 权限缓存键前缀
	PermissionCacheKeyPrefix = "auth:permission:"

	// 登录失败缓存键前缀
	FailedLoginCacheKeyPrefix = "auth:failed_login:"

	// TokenHash缓存键前缀
	TokenHashCacheKeyPrefix = "auth:token_hash:"

	// 密码重置缓存键前缀
	PasswordResetCacheKeyPrefix = "password_reset:"

	// TOTP密钥缓存键前缀
	TOTPSecretCacheKeyPrefix = "auth:totp:secret:"

	// TOTP失败尝试缓存键前缀
	TOTPSecretFailedAttemptsKeyPrefix = "auth:totp:failed:"

	// 登录尝试缓存键前缀
	LoginAttemptCacheKeyPrefix = "auth:login_attempt:"

	// 账户资源缓存键前缀（按资源类型分组）
	ResourceCacheKeyPrefix = "auth:resource:"
)

// CacheKeyGenerator 缓存键生成器
type CacheKeyGenerator struct{}

// NewCacheKeyGenerator 创建缓存键生成器
func NewCacheKeyGenerator() *CacheKeyGenerator {
	return &CacheKeyGenerator{}
}

// UserCacheKey 生成用户缓存键
func (g *CacheKeyGenerator) UserCacheKey(userID string) string {
	return UserCacheKeyPrefix + userID
}

// AccountCacheKey 生成账户缓存键
func (g *CacheKeyGenerator) AccountCacheKey(accountID string) string {
	return AccountCacheKeyPrefix + accountID
}

// SessionCacheKey 生成会话缓存键（包含租户ID以支持多租户）
func (g *CacheKeyGenerator) SessionCacheKey(sessionID, tenantID string) string {
	return SessionCacheKeyPrefix + tenantID + ":" + sessionID
}

// PermissionCacheKey 生成权限缓存键
func (g *CacheKeyGenerator) PermissionCacheKey(userID string) string {
	return PermissionCacheKeyPrefix + userID
}

// FailedLoginCacheKey 生成登录失败缓存键
func (g *CacheKeyGenerator) FailedLoginCacheKey(accountID, ip string) string {
	return FailedLoginCacheKeyPrefix + accountID + ":" + ip
}

// TokenHashCacheKey 生成tokenHash缓存键
func (g *CacheKeyGenerator) TokenHashCacheKey(tokenHash string) string {
	return TokenHashCacheKeyPrefix + tokenHash
}

// PasswordResetCacheKey 生成密码重置缓存键
func (g *CacheKeyGenerator) PasswordResetCacheKey(email string) string {
	return PasswordResetCacheKeyPrefix + email
}

// TOTPSecretCacheKey 生成TOTP密钥缓存键
func (g *CacheKeyGenerator) TOTPSecretCacheKey(accountID string) string {
	return TOTPSecretCacheKeyPrefix + accountID
}

// TOTPSecretFailedAttemptsKey 生成TOTP失败尝试缓存键
func (g *CacheKeyGenerator) TOTPSecretFailedAttemptsKey(accountID string) string {
	return TOTPSecretFailedAttemptsKeyPrefix + accountID
}

// LoginAttemptCacheKey 生成登录尝试缓存键
func (g *CacheKeyGenerator) LoginAttemptCacheKey(identifier string) string {
	return LoginAttemptCacheKeyPrefix + identifier
}

// ResourceCacheKey 生成账户资源缓存键（按资源类型）
func (g *CacheKeyGenerator) ResourceCacheKey(accountID, resourceType string) string {
	return ResourceCacheKeyPrefix + accountID + ":" + resourceType
}
