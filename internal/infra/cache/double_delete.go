package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// CacheAdapter 缓存适配器接口
type CacheAdapter interface {
	Delete(ctx context.Context, key string) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

// DoubleDeleteCache 延迟双删缓存策略
type DoubleDeleteCache struct {
	cache          CacheAdapter
	delay          time.Duration
	pendingJobs    chan *deleteJob
	mu             sync.RWMutex
	workers        int
	stopChan       chan struct{}
	wg             sync.WaitGroup
	totalDeletes   int64
	delayedDeletes int64
}

// deleteJob 删除任务
type deleteJob struct {
	key      string
	delay    time.Duration
	attempt  int
	maxRetry int
}

// NewDoubleDeleteCache 创建延迟双删缓存
func NewDoubleDeleteCache(cache CacheAdapter, delay time.Duration, workers int) *DoubleDeleteCache {
	if cache == nil {
		panic("cache adapter cannot be nil")
	}

	if delay <= 0 {
		delay = 500 * time.Millisecond // 默认延迟500ms
	}

	if workers <= 0 {
		workers = 3 // 默认3个工作协程
	}

	ddc := &DoubleDeleteCache{
		cache:          cache,
		delay:          delay,
		pendingJobs:    make(chan *deleteJob, 1000),
		workers:        workers,
		stopChan:       make(chan struct{}),
		totalDeletes:   0,
		delayedDeletes: 0,
	}

	// 启动工作协程
	ddc.start()

	return ddc
}

// Delete 执行延迟双删
func (ddc *DoubleDeleteCache) Delete(ctx context.Context, key string) error {
	// 增加总删除计数
	ddc.mu.Lock()
	ddc.totalDeletes++
	ddc.mu.Unlock()

	// 1. 立即删除缓存
	if err := ddc.cache.Delete(ctx, key); err != nil {
		log.Printf("Failed to delete cache key %s (first attempt): %v", key, err)
		// 即使第一次删除失败，也继续执行延迟删除
	}

	// 2. 调度延迟删除
	job := &deleteJob{
		key:      key,
		delay:    ddc.delay,
		attempt:  0,
		maxRetry: 3,
	}

	select {
	case ddc.pendingJobs <- job:
		// 成功调度
	default:
		// 队列满了，记录警告
		log.Printf("Warning: double delete queue is full, skipping delayed delete for key %s", key)
	}

	return nil
}

// DeleteWithCustomDelay 使用自定义延迟的双删
func (ddc *DoubleDeleteCache) DeleteWithCustomDelay(ctx context.Context, key string, delay time.Duration) error {
	// 1. 立即删除缓存
	if err := ddc.cache.Delete(ctx, key); err != nil {
		log.Printf("Failed to delete cache key %s (first attempt): %v", key, err)
	}

	// 2. 调度延迟删除
	job := &deleteJob{
		key:      key,
		delay:    delay,
		attempt:  0,
		maxRetry: 3,
	}

	select {
	case ddc.pendingJobs <- job:
	default:
		log.Printf("Warning: double delete queue is full, skipping delayed delete for key %s", key)
	}

	return nil
}

// BatchDelete 批量双删
func (ddc *DoubleDeleteCache) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// 1. 立即批量删除
	for _, key := range keys {
		if err := ddc.cache.Delete(ctx, key); err != nil {
			log.Printf("Failed to delete cache key %s (first attempt): %v", key, err)
		}
	}

	// 2. 调度延迟删除
	for _, key := range keys {
		job := &deleteJob{
			key:      key,
			delay:    ddc.delay,
			attempt:  0,
			maxRetry: 3,
		}

		select {
		case ddc.pendingJobs <- job:
		default:
			log.Printf("Warning: double delete queue is full, skipping delayed delete for key %s", key)
		}
	}

	return nil
}

// Set 删除并重新设置缓存
func (ddc *DoubleDeleteCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 1. 删除旧的缓存
	if err := ddc.Delete(ctx, key); err != nil {
		log.Printf("Failed to delete old cache key %s before set: %v", key, err)
	}

	// 2. 设置新的缓存
	if err := ddc.cache.Set(ctx, key, value, ttl); err != nil {
		return fmt.Errorf("failed to set cache key %s: %w", key, err)
	}

	return nil
}

// GetStats 获取延迟双删统计信息
func (ddc *DoubleDeleteCache) GetStats() map[string]interface{} {
	ddc.mu.RLock()
	defer ddc.mu.RUnlock()

	return map[string]interface{}{
		"queue_length":   len(ddc.pendingJobs),
		"workers":        ddc.workers,
		"default_delay":  ddc.delay.String(),
		"max_queue_size": cap(ddc.pendingJobs),
	}
}

// Stop 停止延迟双删服务
func (ddc *DoubleDeleteCache) Stop() {
	close(ddc.stopChan)
	ddc.wg.Wait()
}

// start 启动工作协程
func (ddc *DoubleDeleteCache) start() {
	for i := 0; i < ddc.workers; i++ {
		ddc.wg.Add(1)
		go ddc.worker(i)
	}
}

// worker 工作协程
func (ddc *DoubleDeleteCache) worker(workerID int) {
	defer ddc.wg.Done()

	log.Printf("Double delete worker %d started", workerID)

	for {
		select {
		case job := <-ddc.pendingJobs:
			ddc.processJob(workerID, job)

		case <-ddc.stopChan:
			log.Printf("Double delete worker %d stopping", workerID)
			return
		}
	}
}

// processJob 处理删除任务
func (ddc *DoubleDeleteCache) processJob(workerID int, job *deleteJob) {
	// 等待延迟时间
	time.Sleep(job.delay)

	// 执行延迟删除
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ddc.cache.Delete(ctx, job.key); err != nil {
		log.Printf("Worker %d: failed to delete cache key %s (delayed attempt %d): %v",
			workerID, job.key, job.attempt+1, err)

		// 重试机制
		if job.attempt < job.maxRetry {
			job.attempt++
			job.delay = job.delay * 2 // 指数退避

			// 重新调度任务
			select {
			case ddc.pendingJobs <- job:
				log.Printf("Worker %d: rescheduled delete job for key %s (attempt %d)",
					workerID, job.key, job.attempt)
			default:
				log.Printf("Worker %d: failed to reschedule delete job for key %s (queue full)",
					workerID, job.key)
			}
		} else {
			log.Printf("Worker %d: exhausted retries for key %s", workerID, job.key)
		}
	} else {
		// 删除成功，增加延迟删除计数
		ddc.mu.Lock()
		ddc.delayedDeletes++
		ddc.mu.Unlock()

		log.Printf("Worker %d: successfully deleted cache key %s", workerID, job.key)
	}
}

// SetDelay 设置默认延迟时间
func (ddc *DoubleDeleteCache) SetDelay(delay time.Duration) {
	ddc.mu.Lock()
	defer ddc.mu.Unlock()
	ddc.delay = delay
}

// GetDelay 获取默认延迟时间
func (ddc *DoubleDeleteCache) GetDelay() time.Duration {
	ddc.mu.RLock()
	defer ddc.mu.RUnlock()
	return ddc.delay
}

// FlushQueue 清空待处理的任务队列
func (ddc *DoubleDeleteCache) FlushQueue() int {
	count := 0
	for {
		select {
		case <-ddc.pendingJobs:
			count++
		default:
			return count
		}
	}
}

// WarmupCache 预热缓存（在数据变更后立即重新加载到缓存）
func (ddc *DoubleDeleteCache) WarmupCache(ctx context.Context, key string, loader func() (interface{}, error), ttl time.Duration) error {
	// 1. 删除旧缓存
	if err := ddc.Delete(ctx, key); err != nil {
		log.Printf("Failed to delete cache key %s during warmup: %v", key, err)
	}

	// 2. 从数据源重新加载
	value, err := loader()
	if err != nil {
		return fmt.Errorf("failed to load data for key %s: %w", key, err)
	}

	// 3. 设置新缓存
	if err := ddc.cache.Set(ctx, key, value, ttl); err != nil {
		return fmt.Errorf("failed to set cache key %s during warmup: %w", key, err)
	}

	return nil
}

// InvalidatePattern 按模式失效缓存
func (ddc *DoubleDeleteCache) InvalidatePattern(ctx context.Context, pattern string, keysProvider func(string) ([]string, error)) error {
	// 1. 获取匹配模式的键列表
	keys, err := keysProvider(pattern)
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	// 2. 批量双删
	return ddc.BatchDelete(ctx, keys)
}

// CacheWarmer 缓存预热器
type CacheWarmer struct {
	ddc *DoubleDeleteCache
}

// NewCacheWarmer 创建缓存预热器
func NewCacheWarmer(ddc *DoubleDeleteCache) *CacheWarmer {
	return &CacheWarmer{ddc: ddc}
}

// Warmup 预热单个键
func (cw *CacheWarmer) Warmup(ctx context.Context, key string, loader func() (interface{}, error), ttl time.Duration) error {
	return cw.ddc.WarmupCache(ctx, key, loader, ttl)
}

// WarmupBatch 批量预热
func (cw *CacheWarmer) WarmupBatch(ctx context.Context, warmupJobs map[string]WarmupJob) error {
	for key, job := range warmupJobs {
		if err := cw.Warmup(ctx, key, job.Loader, job.TTL); err != nil {
			log.Printf("Failed to warmup cache key %s: %v", key, err)
			// 继续处理其他键
		}
	}
	return nil
}

// WarmupJob 预热任务
type WarmupJob struct {
	Loader func() (interface{}, error)
	TTL    time.Duration
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	TotalDeletes     int64         `json:"total_deletes"`
	DelayedDeletes   int64         `json:"delayed_deletes"`
	AverageDelay     time.Duration `json:"average_delay"`
	MaxQueueLength   int           `json:"max_queue_length"`
	CurrentQueueSize int           `json:"current_queue_size"`
}

// GetMetrics 获取性能指标
func (ddc *DoubleDeleteCache) GetMetrics() PerformanceMetrics {
	ddc.mu.RLock()
	defer ddc.mu.RUnlock()

	return PerformanceMetrics{
		TotalDeletes:     ddc.totalDeletes,   // 使用实际计数器值
		DelayedDeletes:   ddc.delayedDeletes, // 使用实际计数器值
		AverageDelay:     ddc.delay,
		MaxQueueLength:   cap(ddc.pendingJobs),
		CurrentQueueSize: len(ddc.pendingJobs),
	}
}
