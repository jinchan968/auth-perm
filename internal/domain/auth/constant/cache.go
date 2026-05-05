package constant

import cacheConstant "auth-perm/internal/domain/cache/constant"

const (
	UserCacheKeyPrefix                = "auth:user:"
	AccountCacheKeyPrefix             = "auth:account:"
	SessionCacheKeyPrefix             = "auth:session:"
	FailedLoginCacheKeyPrefix         = "auth:failed_login:"
	TokenHashCacheKeyPrefix           = "auth:token_hash:"
	PasswordResetCacheKeyPrefix       = "auth:password_reset:"
	TOTPSecretCacheKeyPrefix          = "auth:totp:secret:"
	TOTPSecretFailedAttemptsKeyPrefix = "auth:totp:failed:"
	LoginAttemptCacheKeyPrefix        = "auth:login_attempt:"
)

// CacheKeyGenerator auth 域缓存键生成器
type CacheKeyGenerator struct{}

func NewCacheKeyGenerator() *CacheKeyGenerator {
	return &CacheKeyGenerator{}
}

func (g *CacheKeyGenerator) UserCacheKey(userID string) string {
	return cacheConstant.BuildKey(UserCacheKeyPrefix, userID)
}

func (g *CacheKeyGenerator) AccountCacheKey(accountID string) string {
	return cacheConstant.BuildKey(AccountCacheKeyPrefix, accountID)
}

func (g *CacheKeyGenerator) SessionCacheKey(sessionID, tenantID string) string {
	return cacheConstant.BuildKey(SessionCacheKeyPrefix, tenantID, sessionID)
}

func (g *CacheKeyGenerator) FailedLoginCacheKey(accountID, ip string) string {
	return cacheConstant.BuildKey(FailedLoginCacheKeyPrefix, accountID, ip)
}

func (g *CacheKeyGenerator) TokenHashCacheKey(tokenHash string) string {
	return cacheConstant.BuildKey(TokenHashCacheKeyPrefix, tokenHash)
}

func (g *CacheKeyGenerator) PasswordResetCacheKey(email string) string {
	return cacheConstant.BuildKey(PasswordResetCacheKeyPrefix, email)
}

func (g *CacheKeyGenerator) TOTPSecretCacheKey(accountID string) string {
	return cacheConstant.BuildKey(TOTPSecretCacheKeyPrefix, accountID)
}

func (g *CacheKeyGenerator) TOTPSecretFailedAttemptsKey(accountID string) string {
	return cacheConstant.BuildKey(TOTPSecretFailedAttemptsKeyPrefix, accountID)
}

func (g *CacheKeyGenerator) LoginAttemptCacheKey(identifier string) string {
	return cacheConstant.BuildKey(LoginAttemptCacheKeyPrefix, identifier)
}
