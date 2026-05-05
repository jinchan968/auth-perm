package cache

import (
	"log"

	"go.uber.org/dig"

	"auth-perm/internal/domain/cache/service"
	"auth-perm/internal/infra/cache"
)

// RegisterCacheDomain 注册缓存领域的所有依赖
func RegisterCacheDomain(container *dig.Container) error {
	// 注册泛型缓存服务
	if err := container.Provide(func(
		cacheSvc cache.Cache,
		redisClient *cache.RedisCache,
	) *service.Service {
		log.Println("CacheService registered successfully")
		return service.NewService(cacheSvc, redisClient)
	}); err != nil {
		return err
	}

	log.Println("Cache domain registered successfully")
	return nil
}
