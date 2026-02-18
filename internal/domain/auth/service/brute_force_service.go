package service

import (
	"context"
	"encoding/json"
	"time"

	"auth-perm/internal/common/errors"
	authConstant "auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/infra/cache"
)

// BruteForceService 防暴力破解服务
type BruteForceService struct {
	cache        cache.Cache
	keyGenerator *authConstant.CacheKeyGenerator
}

// NewBruteForceService 创建防暴力破解服务
func NewBruteForceService(c cache.Cache) *BruteForceService {
	return &BruteForceService{
		cache:        c,
		keyGenerator: authConstant.NewCacheKeyGenerator(),
	}
}

// LoginAttempt 登录尝试记录
type LoginAttempt struct {
	Count       int        `json:"count"`
	FirstTryAt  time.Time  `json:"first_try_at"`
	LastTryAt   time.Time  `json:"last_try_at"`
	LockedUntil *time.Time `json:"locked_until,omitempty"`
}

// CheckLoginAttempt 检查登录尝试
// 返回值: 是否允许登录, 剩余尝试次数, 错误信息
func (s *BruteForceService) CheckLoginAttempt(ctx context.Context, identifier string) (bool, int, error) {
	key := s.keyGenerator.LoginAttemptCacheKey(identifier)

	// 检查缓存是否存在
	exists, err := s.cache.Exists(ctx, key)
	if err != nil || !exists {
		// 缓存中没有记录，允许登录
		return true, 5, nil
	}

	// 获取尝试记录
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return true, 5, nil
	}

	// 解析尝试记录
	var attempt LoginAttempt
	dataBytes, ok := data.([]byte)
	if !ok {
		// 如果不是字节数组，尝试JSON解析
		if dataStr, ok := data.(string); ok {
			dataBytes = []byte(dataStr)
		} else {
			return true, 5, nil
		}
	}

	if err := json.Unmarshal(dataBytes, &attempt); err != nil {
		return true, 5, nil
	}

	// 检查是否被锁定
	if attempt.LockedUntil != nil && time.Now().Before(*attempt.LockedUntil) {
		remaining := int(attempt.LockedUntil.Sub(time.Now()).Minutes())
		return false, 0, errors.NewAuthErrorF("账户已被锁定，请在 %d 分钟后再试", remaining)
	}

	// 检查尝试次数是否超限（5次后锁定15分钟）
	if attempt.Count >= 5 {
		// 锁定账户
		lockedUntil := time.Now().Add(15 * time.Minute)
		attempt.LockedUntil = &lockedUntil
		attempt.LastTryAt = time.Now()

		// 更新缓存
		if err := s.cache.Set(ctx, key, attempt, 30*time.Minute); err != nil {
			return false, 0, errors.WrapBizError(err, "更新锁定状态失败")
		}

		return false, 0, errors.NewAuthError("登录尝试次数过多，账户已被锁定15分钟")
	}

	// 计算剩余尝试次数
	remaining := 5 - attempt.Count
	return true, remaining, nil
}

// RecordFailedLogin 记录登录失败
// 增加 ip 和 reason 参数以支持更完整的审计
func (s *BruteForceService) RecordFailedLogin(ctx context.Context, identifier, ip, reason string) error {
	key := s.keyGenerator.LoginAttemptCacheKey(identifier)

	// 获取当前尝试记录
	var attempt LoginAttempt

	// 检查缓存是否存在
	exists, err := s.cache.Exists(ctx, key)
	if err == nil && exists {
		// 获取尝试记录
		data, err := s.cache.Get(ctx, key)
		if err == nil {
			// 解析尝试记录
			if dataBytes, ok := data.([]byte); ok {
				_ = json.Unmarshal(dataBytes, &attempt)
			} else if dataStr, ok := data.(string); ok {
				_ = json.Unmarshal([]byte(dataStr), &attempt)
			}
		}
	}

	// 如果距离第一次尝试超过30分钟，重置计数
	if attempt.FirstTryAt.IsZero() || time.Since(attempt.FirstTryAt) > 30*time.Minute {
		attempt.Count = 0
		attempt.FirstTryAt = time.Now()
	}

	// 更新尝试次数和时间
	attempt.Count++
	attempt.LastTryAt = time.Now()

	// JSON序列化并设置缓存
	dataBytes, err := json.Marshal(attempt)
	if err != nil {
		return err
	}

	// 设置缓存（30分钟过期）
	return s.cache.Set(ctx, key, string(dataBytes), 30*time.Minute)
}

// RecordSuccessfulLogin 记录登录成功，清除失败记录
func (s *BruteForceService) RecordSuccessfulLogin(ctx context.Context, identifier string) error {
	key := s.keyGenerator.LoginAttemptCacheKey(identifier)
	// 登录成功，清除失败记录
	return s.cache.Delete(ctx, key)
}
