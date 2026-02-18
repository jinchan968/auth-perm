package cache

import "errors"

// 缓存错误定义
var (
	// ErrKeyNotFound 键不存在
	ErrKeyNotFound = errors.New("cache key not found")

	// ErrCacheFull 缓存已满
	ErrCacheFull = errors.New("cache is full")

	// ErrInvalidKey 无效的键
	ErrInvalidKey = errors.New("invalid cache key")

	// ErrInvalidValue 无效的值
	ErrInvalidValue = errors.New("invalid cache value")

	// ErrTTLExpired TTL已过期
	ErrTTLExpired = errors.New("cache TTL expired")

	// ErrSerializationFailed 序列化失败
	ErrSerializationFailed = errors.New("cache serialization failed")

	// ErrDeserializationFailed 反序列化失败
	ErrDeserializationFailed = errors.New("cache deserialization failed")

	// ErrRedisConnection Redis连接错误
	ErrRedisConnection = errors.New("redis connection error")

	// ErrRedisTimeout Redis超时错误
	ErrRedisTimeout = errors.New("redis timeout error")

	// ErrQueueFull 队列已满
	ErrQueueFull = errors.New("cache queue is full")

	// ErrWorkerStopped 工作协程已停止
	ErrWorkerStopped = errors.New("cache worker is stopped")
)

// IsKeyNotFound FUTURE: 键不存在错误检查 - 在实现错误处理时使用
func IsKeyNotFound(err error) bool {
	return errors.Is(err, ErrKeyNotFound)
}

// IsCacheFull FUTURE: 缓存满错误检查 - 在实现缓存管理时使用
func IsCacheFull(err error) bool {
	return errors.Is(err, ErrCacheFull)
}

// IsInvalidKey FUTURE: 无效键错误检查 - 在实现键验证时使用
func IsInvalidKey(err error) bool {
	return errors.Is(err, ErrInvalidKey)
}

// IsInvalidValue FUTURE: 无效值错误检查 - 在实现值验证时使用
func IsInvalidValue(err error) bool {
	return errors.Is(err, ErrInvalidValue)
}

// IsTTLExpired FUTURE: TTL过期错误检查 - 在实现过期处理时使用
func IsTTLExpired(err error) bool {
	return errors.Is(err, ErrTTLExpired)
}

// IsSerializationFailed FUTURE: 序列化失败错误检查 - 在实现数据序列化时使用
func IsSerializationFailed(err error) bool {
	return errors.Is(err, ErrSerializationFailed)
}

// IsDeserializationFailed FUTURE: 反序列化失败错误检查 - 在实现数据反序列化时使用
func IsDeserializationFailed(err error) bool {
	return errors.Is(err, ErrDeserializationFailed)
}

// IsRedisConnection FUTURE: Redis连接错误检查 - 在实现连接管理时使用
func IsRedisConnection(err error) bool {
	return errors.Is(err, ErrRedisConnection)
}

// IsRedisTimeout FUTURE: Redis超时错误检查 - 在实现超时处理时使用
func IsRedisTimeout(err error) bool {
	return errors.Is(err, ErrRedisTimeout)
}

// IsQueueFull FUTURE: 队列满错误检查 - 在实现队列管理时使用
func IsQueueFull(err error) bool {
	return errors.Is(err, ErrQueueFull)
}

// IsWorkerStopped FUTURE: 工作器停止错误检查 - 在实现工作器管理时使用
func IsWorkerStopped(err error) bool {
	return errors.Is(err, ErrWorkerStopped)
}
