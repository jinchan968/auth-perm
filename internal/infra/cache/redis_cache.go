package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	if prefix == "" {
		prefix = ""
	}
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// Get 获取缓存值
func (r *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	fullKey := r.getFullKey(key)
	val, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}

	// 尝试解析JSON
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		// 如果不是JSON，直接返回字符串
		return val, nil
	}

	return result, nil
}

// Set 设置缓存值
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := r.getFullKey(key)

	// 序列化值
	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		val = string(jsonBytes)
	}

	// 设置缓存
	err := r.client.Set(ctx, fullKey, val, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Delete 删除缓存
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := r.getFullKey(key)
	err := r.client.Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.getFullKey(key)
	result, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return result > 0, nil
}

// SetNX 设置键值（仅当键不存在时）
func (r *RedisCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	fullKey := r.getFullKey(key)

	// 序列化值
	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return false, fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		val = string(jsonBytes)
	}

	result, err := r.client.SetNX(ctx, fullKey, val, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to setNX key %s: %w", key, err)
	}

	return result, nil
}

// GetTTL 获取键的剩余生存时间
func (r *RedisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := r.getFullKey(key)
	ttl, err := r.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return ttl, nil
}

// Expire 设置键的过期时间
func (r *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := r.getFullKey(key)
	err := r.client.Expire(ctx, fullKey, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration for key %s: %w", key, err)
	}
	return nil
}

// Increment 递增数值
func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	fullKey := r.getFullKey(key)
	result, err := r.client.Incr(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return result, nil
}

// IncrementBy 递增指定值
func (r *RedisCache) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	fullKey := r.getFullKey(key)
	result, err := r.client.IncrBy(ctx, fullKey, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, value, err)
	}
	return result, nil
}

// Decrement 递减数值
func (r *RedisCache) Decrement(ctx context.Context, key string) (int64, error) {
	fullKey := r.getFullKey(key)
	result, err := r.client.Decr(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}
	return result, nil
}

// DecrementBy 递减指定值
func (r *RedisCache) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	fullKey := r.getFullKey(key)
	result, err := r.client.DecrBy(ctx, fullKey, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s by %d: %w", key, value, err)
	}
	return result, nil
}

// GetMultiple 获取多个键的值
func (r *RedisCache) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// 构建完整的键列表
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.getFullKey(key)
	}

	// 批量获取
	values, err := r.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple keys: %w", err)
	}

	result := make(map[string]interface{})
	for i, key := range keys {
		if values[i] != nil {
			valStr := values[i].(string)
			var val interface{}
			if json.Unmarshal([]byte(valStr), &val) == nil {
				result[key] = val
			} else {
				result[key] = valStr
			}
		}
	}

	return result, nil
}

// SetMultiple 设置多个键值对
func (r *RedisCache) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()

	for key, value := range items {
		fullKey := r.getFullKey(key)

		// 序列化值
		var val string
		switch v := value.(type) {
		case string:
			val = v
		case []byte:
			val = string(v)
		default:
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
			}
			val = string(jsonBytes)
		}

		pipe.Set(ctx, fullKey, val, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set multiple keys: %w", err)
	}

	return nil
}

// DeleteMultiple 删除多个键
func (r *RedisCache) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.getFullKey(key)
	}

	err := r.client.Del(ctx, fullKeys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete multiple keys: %w", err)
	}

	return nil
}

// GetJSON 获取JSON值并解析到指定结构
func (r *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	fullKey := r.getFullKey(key)
	val, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON for key %s: %w", key, err)
	}

	return nil
}

// SetJSON 设置JSON值
func (r *RedisCache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := r.getFullKey(key)

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for key %s: %w", key, err)
	}

	err = r.client.Set(ctx, fullKey, jsonBytes, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set JSON for key %s: %w", key, err)
	}

	return nil
}

// GetList 获取列表
func (r *RedisCache) GetList(ctx context.Context, key string) ([]string, error) {
	fullKey := r.getFullKey(key)
	result, err := r.client.LRange(ctx, fullKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get list for key %s: %w", key, err)
	}
	return result, nil
}

// SetList 设置列表
func (r *RedisCache) SetList(ctx context.Context, key string, values []string, ttl time.Duration) error {
	if len(values) == 0 {
		return nil
	}

	fullKey := r.getFullKey(key)

	interfaceValues := make([]interface{}, len(values))
	for i, v := range values {
		interfaceValues[i] = v
	}

	pipe := r.client.Pipeline()
	pipe.Del(ctx, fullKey)
	pipe.RPush(ctx, fullKey, interfaceValues...)
	if ttl > 0 {
		pipe.Expire(ctx, fullKey, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set list for key %s: %w", key, err)
	}

	return nil
}

// PushToList 向列表推入值
func (r *RedisCache) PushToList(ctx context.Context, key string, values ...string) error {
	fullKey := r.getFullKey(key)
	interfaceValues := make([]interface{}, len(values))
	for i, v := range values {
		interfaceValues[i] = v
	}
	err := r.client.RPush(ctx, fullKey, interfaceValues...).Err()
	if err != nil {
		return fmt.Errorf("failed to push to list for key %s: %w", key, err)
	}
	return nil
}

// PopFromList 从列表弹出值
func (r *RedisCache) PopFromList(ctx context.Context, key string) (string, error) {
	fullKey := r.getFullKey(key)
	result, err := r.client.LPop(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("failed to pop from list for key %s: %w", key, err)
	}
	return result, nil
}

// GetSet 获取并设置新值
func (r *RedisCache) GetSet(ctx context.Context, key string, value interface{}) (interface{}, error) {
	fullKey := r.getFullKey(key)

	// 序列化新值
	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		val = string(jsonBytes)
	}

	result, err := r.client.GetSet(ctx, fullKey, val).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to getSet for key %s: %w", key, err)
	}

	// 尝试解析返回值
	var parsedResult interface{}
	if json.Unmarshal([]byte(result), &parsedResult) == nil {
		return parsedResult, nil
	}

	return result, nil
}

// Ping 检查Redis连接
func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close 关闭Redis连接
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// getFullKey 获取完整的键名
func (r *RedisCache) getFullKey(key string) string {
	return r.prefix + key
}

// GetStats 获取Redis统计信息
func (r *RedisCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.Info(ctx, "memory", "keyspace", "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis info: %w", err)
	}

	return map[string]interface{}{
		"info": info,
	}, nil
}
