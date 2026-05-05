package service

import (
	"context"
	"encoding/json"
	"time"

	"auth-perm/internal/common/errors"
	authConstant "auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dto"
	cacheService "auth-perm/internal/domain/cache/service"
	permissionConstant "auth-perm/internal/domain/permission/constant"
)

// CacheService auth 域缓存服务 — 封装 auth 域特有的缓存业务逻辑
type CacheService struct {
	cacheSvc *cacheService.Service
	keyGen   *authConstant.CacheKeyGenerator
}

// NewCacheService 创建缓存服务
func NewCacheService(cacheSvc *cacheService.Service) *CacheService {
	return &CacheService{
		cacheSvc: cacheSvc,
		keyGen:   authConstant.NewCacheKeyGenerator(),
	}
}

// TokenHashCacheKey 暴露 tokenHash 缓存键生成
func (s *CacheService) TokenHashCacheKey(hash string) string {
	return s.keyGen.TokenHashCacheKey(hash)
}

// GetUserByID 根据ID获取用户（带缓存）
func (s *CacheService) GetUserByID(ctx context.Context, userID string) (*dto.UserDTO, error) {
	key := s.keyGen.UserCacheKey(userID)

	data, err := s.cacheSvc.GetBytes(ctx, key)
	if err == nil {
		var userDTO dto.UserDTO
		if err := json.Unmarshal(data, &userDTO); err == nil {
			return &userDTO, nil
		}
	}

	return nil, errors.NewBusinessError("User not found in cache")
}

// SetUser 缓存用户信息
func (s *CacheService) SetUser(ctx context.Context, user *dto.UserDTO, ttl time.Duration) error {
	key := s.keyGen.UserCacheKey(user.GetID())

	data, err := json.Marshal(user)
	if err != nil {
		return errors.WrapBizError(err, "Failed to marshal user data")
	}

	return s.cacheSvc.SetBytes(ctx, key, data, ttl)
}

// DeleteUser 删除用户缓存
func (s *CacheService) DeleteUser(ctx context.Context, userID string) error {
	key := s.keyGen.UserCacheKey(userID)
	return s.cacheSvc.Delete(ctx, key)
}

// GetAccountByID 根据ID获取账户（带缓存）
func (s *CacheService) GetAccountByID(ctx context.Context, accountID string) (*dto.AccountDTO, error) {
	key := s.keyGen.AccountCacheKey(accountID)

	data, err := s.cacheSvc.GetBytes(ctx, key)
	if err == nil {
		var accountDTO dto.AccountDTO
		if err := json.Unmarshal(data, &accountDTO); err == nil {
			return &accountDTO, nil
		}
	}

	return nil, errors.NewBusinessError("Account not found in cache")
}

// SetAccount 缓存账户信息
func (s *CacheService) SetAccount(ctx context.Context, account *dto.AccountDTO, ttl time.Duration) error {
	key := s.keyGen.AccountCacheKey(account.ID)

	data, err := json.Marshal(account)
	if err != nil {
		return errors.WrapBizError(err, "Failed to marshal account data")
	}

	return s.cacheSvc.SetBytes(ctx, key, data, ttl)
}

// DeleteAccount 删除账户缓存
func (s *CacheService) DeleteAccount(ctx context.Context, accountID string) error {
	key := s.keyGen.AccountCacheKey(accountID)
	return s.cacheSvc.Delete(ctx, key)
}

// GetSession 获取会话（带缓存）
func (s *CacheService) GetSession(ctx context.Context, sessionID, tenantID string) (string, error) {
	key := s.keyGen.SessionCacheKey(sessionID, tenantID)

	data, err := s.cacheSvc.Get(ctx, key)
	if err != nil {
		return "", errors.WrapBizError(err, "Failed to get session from cache")
	}

	return data, nil
}

// SetSession 缓存会话（存储完整session信息）
func (s *CacheService) SetSession(ctx context.Context, session *dto.SessionDTO, ttl time.Duration) error {
	cache := dto.NewSessionCache(session)

	data, err := json.Marshal(cache)
	if err != nil {
		return errors.WrapBizError(err, "序列化session失败")
	}

	tokenKey := s.keyGen.TokenHashCacheKey(session.TokenHash)
	if err := s.cacheSvc.SetBytes(ctx, tokenKey, data, ttl); err != nil {
		return errors.WrapBizError(err, "缓存session失败")
	}

	sessionKey := s.keyGen.SessionCacheKey(session.ID, session.GetTenantID())
	return s.cacheSvc.SetBytes(ctx, sessionKey, data, ttl)
}

// GetSessionIDByTokenHash 通过tokenHash获取sessionID
func (s *CacheService) GetSessionIDByTokenHash(ctx context.Context, tokenHash string) (string, error) {
	key := s.keyGen.TokenHashCacheKey(tokenHash)
	sessionID, err := s.cacheSvc.Get(ctx, key)
	if err != nil {
		return "", errors.WrapBizError(err, "Failed to get session ID from cache")
	}

	return sessionID, nil
}

// DeleteSession 删除会话缓存
func (s *CacheService) DeleteSession(ctx context.Context, sessionID, tenantID string) error {
	key := s.keyGen.SessionCacheKey(sessionID, tenantID)
	data, err := s.cacheSvc.GetBytes(ctx, key)
	if err == nil {
		var sessionCache dto.SessionCache
		if json.Unmarshal(data, &sessionCache) == nil && sessionCache.TokenHash != "" {
			tokenKey := s.keyGen.TokenHashCacheKey(sessionCache.TokenHash)
			_ = s.cacheSvc.Delete(ctx, tokenKey)
		}
	}
	return s.cacheSvc.Delete(ctx, key)
}

// Exists 检查缓存项是否存在
func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	return s.cacheSvc.Exists(ctx, key)
}

// GetPermissions 获取用户权限（带缓存）
func (s *CacheService) GetPermissions(ctx context.Context, userID string) ([]string, error) {
	key := permissionConstant.PermissionCacheKey(userID)

	data, err := s.cacheSvc.GetBytes(ctx, key)
	if err == nil {
		var permissions []string
		if err := json.Unmarshal(data, &permissions); err == nil {
			return permissions, nil
		}
	}

	return nil, errors.NewBusinessError("Permissions not found in cache")
}

// SetPermissions 缓存用户权限
func (s *CacheService) SetPermissions(ctx context.Context, userID string, permissions []string, ttl time.Duration) error {
	key := permissionConstant.PermissionCacheKey(userID)

	data, err := json.Marshal(permissions)
	if err != nil {
		return errors.WrapBizError(err, "Failed to marshal permissions data")
	}

	return s.cacheSvc.SetBytes(ctx, key, data, ttl)
}

// DeletePermissions 删除权限缓存（延迟双删）
func (s *CacheService) DeletePermissions(ctx context.Context, userID string) error {
	key := permissionConstant.PermissionCacheKey(userID)
	return s.cacheSvc.DoubleDelete(ctx, key)
}

// DeletePermissionsByAccountIDs 批量删除多个账户的权限缓存（延迟双删）
func (s *CacheService) DeletePermissionsByAccountIDs(ctx context.Context, accountIDs []string) error {
	if len(accountIDs) == 0 {
		return nil
	}
	for _, accountID := range accountIDs {
		key := permissionConstant.PermissionCacheKey(accountID)
		_ = s.cacheSvc.DoubleDelete(ctx, key)
	}
	return nil
}

// GetAccountResources 获取账户的资源缓存（按资源类型）
func (s *CacheService) GetAccountResources(ctx context.Context, accountID, resourceType string) ([]string, error) {
	key := permissionConstant.ResourceCacheKey(accountID, resourceType)

	data, err := s.cacheSvc.GetBytes(ctx, key)
	if err == nil {
		var resources []string
		if err := json.Unmarshal(data, &resources); err == nil {
			return resources, nil
		}
	}

	return nil, errors.NewBusinessError("Resources not found in cache")
}

// SetAccountResources 缓存账户的资源列表
func (s *CacheService) SetAccountResources(ctx context.Context, accountID, resourceType string, resources []string, ttl time.Duration) error {
	key := permissionConstant.ResourceCacheKey(accountID, resourceType)

	data, err := json.Marshal(resources)
	if err != nil {
		return errors.WrapBizError(err, "Failed to marshal resources data")
	}

	return s.cacheSvc.SetBytes(ctx, key, data, ttl)
}

// DeleteAccountResources 删除单个账户的所有资源缓存（延迟双删）
func (s *CacheService) DeleteAccountResources(ctx context.Context, accountID string) error {
	for _, rt := range permissionConstant.AllResourceTypes {
		key := permissionConstant.ResourceCacheKey(accountID, rt)
		if err := s.cacheSvc.DoubleDelete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// DeleteAccountResourcesByAccountIDs 批量删除多个账户的所有资源缓存（延迟双删）
func (s *CacheService) DeleteAccountResourcesByAccountIDs(ctx context.Context, accountIDs []string) error {
	if len(accountIDs) == 0 {
		return nil
	}
	for _, accountID := range accountIDs {
		_ = s.DeleteAccountResources(ctx, accountID)
	}
	return nil
}

// IncrementFailedLogin 增加登录失败计数
func (s *CacheService) IncrementFailedLogin(ctx context.Context, accountID, ip string) (int64, error) {
	key := s.keyGen.FailedLoginCacheKey(accountID, ip)

	count, err := s.cacheSvc.IncrementBy(ctx, key, 1)
	if err != nil {
		return 0, errors.WrapBizError(err, "Failed to increment failed login count")
	}

	if count == 1 {
		_ = s.cacheSvc.Expire(ctx, key, 15*time.Minute)
	}

	return count, nil
}

// ResetFailedLogin 重置登录失败计数
func (s *CacheService) ResetFailedLogin(ctx context.Context, accountID, ip string) error {
	key := s.keyGen.FailedLoginCacheKey(accountID, ip)
	return s.cacheSvc.Delete(ctx, key)
}

// GetFailedLoginCount 获取登录失败计数
func (s *CacheService) GetFailedLoginCount(ctx context.Context, accountID, ip string) (int64, error) {
	key := s.keyGen.FailedLoginCacheKey(accountID, ip)

	count, err := s.cacheSvc.GetInt64(ctx, key)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

// IsLocked 检查账户是否被锁定
func (s *CacheService) IsLocked(ctx context.Context, accountID, ip string) (bool, error) {
	count, err := s.GetFailedLoginCount(ctx, accountID, ip)
	if err != nil {
		return false, err
	}

	return count >= 5, nil
}

// SetWithTTL 设置键值对并指定TTL
func (s *CacheService) SetWithTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.cacheSvc.Set(ctx, key, value, ttl)
}

// PasswordResetCacheKey 生成密码重置缓存键
func (s *CacheService) PasswordResetCacheKey(email string) string {
	return s.keyGen.PasswordResetCacheKey(email)
}

// Get 获取值
func (s *CacheService) Get(ctx context.Context, key string) (string, error) {
	return s.cacheSvc.Get(ctx, key)
}

// Delete 删除键
func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.cacheSvc.Delete(ctx, key)
}

// FlushAll 清空所有缓存（仅用于测试）
func (s *CacheService) FlushAll(ctx context.Context) error {
	return errors.NewBusinessError("FlushAll operation is not supported through the generic cache interface")
}

// TOTPSecretCacheKey 生成TOTP密钥缓存键
func (s *CacheService) TOTPSecretCacheKey(accountID string) string {
	return s.keyGen.TOTPSecretCacheKey(accountID)
}

// TOTPSecretFailedAttemptsKey 生成TOTP失败尝试缓存键
func (s *CacheService) TOTPSecretFailedAttemptsKey(accountID string) string {
	return s.keyGen.TOTPSecretFailedAttemptsKey(accountID)
}

// GetTOTPSecret 获取TOTP密钥（带缓存）
func (s *CacheService) GetTOTPSecret(accountID string) (*dto.TOTPSecretDTO, error) {
	ctx := context.Background()
	key := s.TOTPSecretCacheKey(accountID)

	data, err := s.cacheSvc.GetBytes(ctx, key)
	if err != nil {
		return nil, nil
	}

	var totpSecret *dto.TOTPSecretDTO
	if err := json.Unmarshal(data, &totpSecret); err == nil {
		return totpSecret, nil
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

	return s.cacheSvc.SetBytes(ctx, key, data, ttl)
}

// DeleteTOTPSecret 删除TOTP密钥缓存
func (s *CacheService) DeleteTOTPSecret(accountID string) error {
	ctx := context.Background()
	key := s.TOTPSecretCacheKey(accountID)
	return s.cacheSvc.Delete(ctx, key)
}

// GetTOTPFailedAttempts 获取TOTP失败尝试次数
func (s *CacheService) GetTOTPFailedAttempts(accountID string) (int64, error) {
	ctx := context.Background()
	key := s.TOTPSecretFailedAttemptsKey(accountID)

	count, err := s.cacheSvc.GetInt64(ctx, key)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

// IncrementTOTPFailedAttempts 增加TOTP失败尝试次数
func (s *CacheService) IncrementTOTPFailedAttempts(accountID string) (int64, error) {
	ctx := context.Background()
	key := s.TOTPSecretFailedAttemptsKey(accountID)

	count, err := s.cacheSvc.IncrementBy(ctx, key, 1)
	if err != nil {
		return 0, errors.WrapBizError(err, "Failed to increment TOTP failed attempts")
	}

	if count == 1 {
		_ = s.cacheSvc.Expire(ctx, key, 15*time.Minute)
	}

	return count, nil
}

// ClearTOTPFailedAttempts 清除TOTP失败尝试次数
func (s *CacheService) ClearTOTPFailedAttempts(accountID string) error {
	ctx := context.Background()
	key := s.TOTPSecretFailedAttemptsKey(accountID)
	return s.cacheSvc.Delete(ctx, key)
}
