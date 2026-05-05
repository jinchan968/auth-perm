package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/infra/cache"
)

// Service 泛型缓存服务 — 提供与业务无关的缓存存取能力
type Service struct {
	cache        cache.Cache
	redis        *cache.RedisCache
	doubleDelete *cache.DoubleDeleteCache
}

// NewService 创建泛型缓存服务
func NewService(c cache.Cache, redisClient *cache.RedisCache) *Service {
	return &Service{
		cache:        c,
		redis:        redisClient,
		doubleDelete: cache.NewDoubleDeleteCache(c, 500*time.Millisecond, 3),
	}
}

// Get 读取字符串值
func (s *Service) Get(ctx context.Context, key string) (string, error) {
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if dataStr, ok := data.(string); ok {
		return dataStr, nil
	}
	return "", errors.NewBusinessError("Invalid data type in cache")
}

// GetBytes 读取原始字节，由调用方自行反序列化
func (s *Service) GetBytes(ctx context.Context, key string) ([]byte, error) {
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if dataStr, ok := data.(string); ok {
		return []byte(dataStr), nil
	}
	return nil, errors.NewBusinessError("Invalid data type in cache")
}

// GetJSON 读取并反序列化为目标类型
func (s *Service) GetJSON(ctx context.Context, key string, target interface{}) error {
	data, err := s.GetBytes(ctx, key)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return errors.WrapBizError(err, "Failed to unmarshal cache data")
	}
	return nil
}

// Set 写入字符串值
func (s *Service) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.cache.Set(ctx, key, value, ttl)
}

// SetBytes 写入原始字节
func (s *Service) SetBytes(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	return s.cache.Set(ctx, key, string(data), ttl)
}

// SetJSON 序列化为 JSON 后写入缓存
func (s *Service) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.WrapBizError(err, "Failed to marshal cache data")
	}
	return s.SetBytes(ctx, key, data, ttl)
}

// Delete 删除缓存
func (s *Service) Delete(ctx context.Context, key string) error {
	return s.cache.Delete(ctx, key)
}

// DoubleDelete 延迟双删
func (s *Service) DoubleDelete(ctx context.Context, key string) error {
	return s.doubleDelete.Delete(ctx, key)
}

// DoubleDeleteKeys 批量延迟双删
func (s *Service) DoubleDeleteKeys(ctx context.Context, keys []string) error {
	for _, key := range keys {
		_ = s.doubleDelete.Delete(ctx, key)
	}
	return nil
}

// Exists 检查 key 是否存在
func (s *Service) Exists(ctx context.Context, key string) (bool, error) {
	return s.cache.Exists(ctx, key)
}

// IncrementBy 原子递增
func (s *Service) IncrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	return s.redis.IncrementBy(ctx, key, delta)
}

// Expire 设置过期时间
func (s *Service) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.redis.Expire(ctx, key, ttl)
}

// GetInt 读取整数值
func (s *Service) GetInt(ctx context.Context, key string) (int, error) {
	data, err := s.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	var result int
	if _, err := fmt.Sscanf(data, "%d", &result); err != nil {
		return 0, err
	}
	return result, nil
}

// GetInt64 读取 int64 值
func (s *Service) GetInt64(ctx context.Context, key string) (int64, error) {
	data, err := s.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	var result int64
	if _, err := fmt.Sscanf(data, "%d", &result); err != nil {
		return 0, err
	}
	return result, nil
}
