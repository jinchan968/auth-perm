package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"auth-perm/internal/common/errors"
	authConstant "auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dto"

	"auth-perm/internal/infra/cache"
)

// CacheService 缓存服务
type CacheService struct {
	cache           cache.Cache
	redis           *cache.RedisCache // 保持对Redis缓存的访问，用于特殊操作
	keyGenerator    *authConstant.CacheKeyGenerator
	permissionCache *cache.DoubleDeleteCache // 用于权限缓存的延迟双删
}

// NewCacheService 创建缓存服务
func NewCacheService(c cache.Cache, redisClient *cache.RedisCache) *CacheService {
	// 创建权限缓存的延迟双删包装器
	permissionCache := cache.NewDoubleDeleteCache(c, 500*time.Millisecond, 3)

	return &CacheService{
		cache:           c,
		redis:           redisClient,
		keyGenerator:    authConstant.NewCacheKeyGenerator(),
		permissionCache: permissionCache,
	}
}

// GetUserByID 根据ID获取用户（带缓存）
func (s *CacheService) GetUserByID(ctx context.Context, userID string) (*dto.UserDTO, error) {
	key := s.keyGenerator.UserCacheKey(userID)

	// 先查缓存
	data, err := s.cache.Get(ctx, key)
	if err == nil {
		if dataStr, ok := data.(string); ok {
			var userDTO dto.UserDTO
			if err := json.Unmarshal([]byte(dataStr), &userDTO); err == nil {
				return &userDTO, nil
			}
		}
	}

	// 缓存未命中，返回错误提示（实际应用中应该调用数据库）
	return nil, errors.NewBusinessError("User not found in cache")
}

// SetUser 缓存用户信息
func (s *CacheService) SetUser(ctx context.Context, user *dto.UserDTO, ttl time.Duration) error {
	key := s.keyGenerator.UserCacheKey(user.GetID())

	data, err := json.Marshal(user)
	if err != nil {
		return errors.WrapBizError(err, "Failed to marshal user data")
	}

	return s.cache.Set(ctx, key, data, ttl)
}

// DeleteUser 删除用户缓存
func (s *CacheService) DeleteUser(ctx context.Context, userID string) error {
	key := s.keyGenerator.UserCacheKey(userID)
	return s.cache.Delete(ctx, key)
}

// GetAccountByID 根据ID获取账户（带缓存）
func (s *CacheService) GetAccountByID(ctx context.Context, accountID string) (*dto.AccountDTO, error) {
	key := s.keyGenerator.AccountCacheKey(accountID)

	// 先查缓存
	data, err := s.cache.Get(ctx, key)
	if err == nil {
		if dataStr, ok := data.(string); ok {
			var accountDTO dto.AccountDTO
			if err := json.Unmarshal([]byte(dataStr), &accountDTO); err == nil {
				return &accountDTO, nil
			}
		}
	}

	// 缓存未命中，返回错误提示（实际应用中应该调用数据库）
	return nil, errors.NewBusinessError("Account not found in cache")
}

// SetAccount 缓存账户信息
func (s *CacheService) SetAccount(ctx context.Context, account *dto.AccountDTO, ttl time.Duration) error {
	key := s.keyGenerator.AccountCacheKey(account.ID)

	data, err := json.Marshal(account)
	if err != nil {
		return errors.WrapBizError(err, "Failed to marshal account data")
	}

	return s.cache.Set(ctx, key, data, ttl)
}

// DeleteAccount 删除账户缓存
func (s *CacheService) DeleteAccount(ctx context.Context, accountID string) error {
	key := s.keyGenerator.AccountCacheKey(accountID)
	return s.cache.Delete(ctx, key)
}

// GetSession 获取会话（带缓存）
func (s *CacheService) GetSession(ctx context.Context, sessionID, tenantID string) (string, error) {
	key := s.keyGenerator.SessionCacheKey(sessionID, tenantID)

	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return "", errors.WrapBizError(err, "Failed to get session from cache")
	}

	if dataStr, ok := data.(string); ok {
		return dataStr, nil
	}
	return "", errors.NewBusinessError("Invalid session data in cache")
}

// SetSession 缓存会话（优化版：存储完整session信息）
func (s *CacheService) SetSession(ctx context.Context, session *dto.SessionDTO, ttl time.Duration) error {
	// 构建完整session信息
	sessionData := map[string]interface{}{
		"session_id": session.ID,
		"token_hash": session.TokenHash,
		"user_id":    session.UserID,
		"account_id": session.AccountID,
		"tenant_id":  session.GetTenantID(), // 添加TenantID支持多租户
		"expires_at": session.ExpiresAt.Unix(),
		"is_active":  session.IsActive,
		"ip_address": session.IPAddress,
		"user_agent": session.UserAgent,
		"created_at": session.CreatedAt.Unix(),
		"updated_at": session.UpdatedAt.Unix(),
	}

	// JSON序列化
	data, err := json.Marshal(sessionData)
	if err != nil {
		return errors.WrapBizError(err, "序列化session失败")
	}

	// 存储tokenHash → session完整信息
	tokenKey := s.keyGenerator.TokenHashCacheKey(session.TokenHash)
	if err := s.cache.Set(ctx, tokenKey, data, ttl); err != nil {
		return errors.WrapBizError(err, "缓存session失败")
	}

	// 存储sessionID → session完整信息（便于主动查找）
	sessionKey := s.keyGenerator.SessionCacheKey(session.ID, session.GetTenantID())
	return s.cache.Set(ctx, sessionKey, data, ttl)
}

// GetSessionIDByTokenHash 通过tokenHash获取sessionID
func (s *CacheService) GetSessionIDByTokenHash(ctx context.Context, tokenHash string) (string, error) {
	key := s.keyGenerator.TokenHashCacheKey(tokenHash)
	sessionID, err := s.cache.Get(ctx, key)
	if err != nil {
		return "", errors.WrapBizError(err, "Failed to get session ID from cache")
	}

	if sessionIDStr, ok := sessionID.(string); ok {
		return sessionIDStr, nil
	}
	return "", errors.NewBusinessError("Invalid session ID data in cache")
}

// DeleteSession 删除会话缓存
func (s *CacheService) DeleteSession(ctx context.Context, sessionID, tenantID string) error {
	// 先获取tokenHash
	key := s.keyGenerator.SessionCacheKey(sessionID, tenantID)
	tokenHash, err := s.cache.Get(ctx, key)
	if err == nil {
		// 删除tokenHash映射
		if tokenHashStr, ok := tokenHash.(string); ok {
			tokenKey := s.keyGenerator.TokenHashCacheKey(tokenHashStr)
			_ = s.cache.Delete(ctx, tokenKey)
		}
	}
	// 删除session缓存
	return s.cache.Delete(ctx, key)
}

// Exists 检查缓存项是否存在
func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	return s.cache.Exists(ctx, key)
}

// GetPermissions 获取用户权限（带缓存）
func (s *CacheService) GetPermissions(ctx context.Context, userID string) ([]string, error) {
	key := s.keyGenerator.PermissionCacheKey(userID)

	data, err := s.cache.Get(ctx, key)
	if err == nil {
		if dataStr, ok := data.(string); ok {
			var permissions []string
			if err := json.Unmarshal([]byte(dataStr), &permissions); err == nil {
				return permissions, nil
			}
		}
	}

	return nil, errors.NewBusinessError("Permissions not found in cache")
}

// SetPermissions 缓存用户权限
func (s *CacheService) SetPermissions(ctx context.Context, userID string, permissions []string, ttl time.Duration) error {
	key := s.keyGenerator.PermissionCacheKey(userID)

	data, err := json.Marshal(permissions)
	if err != nil {
		return errors.WrapBizError(err, "Failed to marshal permissions data")
	}

	return s.cache.Set(ctx, key, data, ttl)
}

// DeletePermissions 删除权限缓存（使用延迟双删）
func (s *CacheService) DeletePermissions(ctx context.Context, userID string) error {
	key := s.keyGenerator.PermissionCacheKey(userID)
	return s.permissionCache.Delete(ctx, key)
}

// DeletePermissionsByAccountIDs 批量删除多个账户的权限缓存（使用延迟双删）
func (s *CacheService) DeletePermissionsByAccountIDs(ctx context.Context, accountIDs []string) error {
	if len(accountIDs) == 0 {
		return nil
	}
	for _, accountID := range accountIDs {
		key := s.keyGenerator.PermissionCacheKey(accountID)
		// 使用延迟双删策略
		if err := s.permissionCache.Delete(ctx, key); err != nil {
			// 继续删除其他键，不因单个错误中断
			continue
		}
	}
	return nil
}

// IncrementFailedLogin 增加登录失败计数
func (s *CacheService) IncrementFailedLogin(ctx context.Context, accountID, ip string) (int64, error) {
	key := s.keyGenerator.FailedLoginCacheKey(accountID, ip)

	// 使用redis.IncrementBy方法
	count, err := s.redis.IncrementBy(ctx, key, 1)
	if err != nil {
		return 0, errors.WrapBizError(err, "Failed to increment failed login count")
	}

	// 设置过期时间为15分钟
	if count == 1 {
		_ = s.redis.Expire(ctx, key, 15*time.Minute)
	}

	return count, nil
}

// ResetFailedLogin 重置登录失败计数
func (s *CacheService) ResetFailedLogin(ctx context.Context, accountID, ip string) error {
	key := s.keyGenerator.FailedLoginCacheKey(accountID, ip)
	return s.cache.Delete(ctx, key)
}

// GetFailedLoginCount 获取登录失败计数
func (s *CacheService) GetFailedLoginCount(ctx context.Context, accountID, ip string) (int64, error) {
	key := s.keyGenerator.FailedLoginCacheKey(accountID, ip)

	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return 0, nil // 返回0而不是错误，当键不存在时
	}

	if countStr, ok := data.(string); ok {
		if count, err := parseInt64(countStr); err == nil {
			return count, nil
		}
	}

	return 0, nil
}

// parseInt64 解析字符串为int64
func parseInt64(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// IsLocked 检查账户是否被锁定
func (s *CacheService) IsLocked(ctx context.Context, accountID, ip string) (bool, error) {
	count, err := s.GetFailedLoginCount(ctx, accountID, ip)
	if err != nil {
		return false, err
	}

	// 连续5次失败锁定账户
	return count >= 5, nil
}

// SetWithTTL 设置键值对并指定TTL
func (s *CacheService) SetWithTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.cache.Set(ctx, key, value, ttl)
}

// PasswordResetCacheKey 生成密码重置缓存键
func (s *CacheService) PasswordResetCacheKey(email string) string {
	return s.keyGenerator.PasswordResetCacheKey(email)
}

// Get 获取值
func (s *CacheService) Get(ctx context.Context, key string) (string, error) {
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if dataStr, ok := data.(string); ok {
		return dataStr, nil
	}
	return "", errors.NewBusinessError("Invalid data type in cache")
}

// Delete 删除键
func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.cache.Delete(ctx, key)
}

// FlushAll 清空所有缓存（仅用于测试）
func (s *CacheService) FlushAll(ctx context.Context) error {
	// FlushAll是Redis特有的操作，在通用cache接口中不可用
	// 在实际项目中，应该通过管理接口或直接访问Redis客户端来执行此操作
	// 这里返回错误提示
	return errors.NewBusinessError("FlushAll operation is not supported through the generic cache interface")
}

// TOTPSecretCacheKey 生成TOTP密钥缓存键
func (s *CacheService) TOTPSecretCacheKey(accountID string) string {
	return s.keyGenerator.TOTPSecretCacheKey(accountID)
}

// TOTPSecretFailedAttemptsKey 生成TOTP失败尝试缓存键
func (s *CacheService) TOTPSecretFailedAttemptsKey(accountID string) string {
	return s.keyGenerator.TOTPSecretFailedAttemptsKey(accountID)
}

// GetTOTPSecret 获取TOTP密钥（带缓存）
func (s *CacheService) GetTOTPSecret(accountID string) (*dto.TOTPSecretDTO, error) {
	ctx := context.Background()
	key := s.TOTPSecretCacheKey(accountID)

	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil, nil // 返回nil而不是错误，当键不存在时
	}

	if dataStr, ok := data.(string); ok {
		var totpSecret *dto.TOTPSecretDTO
		if err := json.Unmarshal([]byte(dataStr), &totpSecret); err == nil {
			return totpSecret, nil
		}
	}

	return nil, nil
}

// SetTOTPSecret 设置TOTP密钥（带缓存）
func (s *CacheService) SetTOTPSecret(accountID string, totpSecret *dto.TOTPSecretDTO, ttl time.Duration) error {
	ctx := context.Background()
	key := s.TOTPSecretCacheKey(accountID)

	data, err := json.Marshal(totpSecret)
	if err != nil {
		return errors.WrapBizError(err, "Failed to marshal TOTP secret")
	}

	return s.cache.Set(ctx, key, data, ttl)
}

// DeleteTOTPSecret 删除TOTP密钥缓存
func (s *CacheService) DeleteTOTPSecret(accountID string) error {
	ctx := context.Background()
	key := s.TOTPSecretCacheKey(accountID)
	return s.cache.Delete(ctx, key)
}

// GetTOTPFailedAttempts 获取TOTP失败尝试次数
func (s *CacheService) GetTOTPFailedAttempts(accountID string) (int64, error) {
	ctx := context.Background()
	key := s.TOTPSecretFailedAttemptsKey(accountID)

	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return 0, nil // 返回0而不是错误，当键不存在时
	}

	if countStr, ok := data.(string); ok {
		if count, err := parseInt64(countStr); err == nil {
			return count, nil
		}
	}

	return 0, nil
}

// IncrementTOTPFailedAttempts 增加TOTP失败尝试次数
func (s *CacheService) IncrementTOTPFailedAttempts(accountID string) (int64, error) {
	ctx := context.Background()
	key := s.TOTPSecretFailedAttemptsKey(accountID)

	count, err := s.redis.IncrementBy(ctx, key, 1)
	if err != nil {
		return 0, errors.WrapBizError(err, "Failed to increment TOTP failed attempts")
	}

	// 设置过期时间为15分钟
	if count == 1 {
		_ = s.redis.Expire(ctx, key, 15*time.Minute)
	}

	return count, nil
}

// ClearTOTPFailedAttempts 清除TOTP失败尝试次数
func (s *CacheService) ClearTOTPFailedAttempts(accountID string) error {
	ctx := context.Background()
	key := s.TOTPSecretFailedAttemptsKey(accountID)
	return s.cache.Delete(ctx, key)
}
