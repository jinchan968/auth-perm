package container

import (
	"context"
	"crypto/tls"
	"log"
	"strings"

	"auth-perm/config"
	"auth-perm/internal/common/errors"
	controllerHttp "auth-perm/internal/controller/http"
	"auth-perm/internal/controller/middleware"
	"auth-perm/internal/domain/auth"
	authHandler "auth-perm/internal/domain/auth/handler"
	"auth-perm/internal/domain/auth/service"
	"auth-perm/internal/domain/cache"
	"auth-perm/internal/domain/journal"
	journalHandler "auth-perm/internal/domain/journal/handler"
	"auth-perm/internal/domain/newshock"
	newshockHandler "auth-perm/internal/domain/newshock/handler"
	newshockService "auth-perm/internal/domain/newshock/service"
	"auth-perm/internal/domain/permission"
	permHandler "auth-perm/internal/domain/permission/handler"
	permissionService "auth-perm/internal/domain/permission/service"
	"auth-perm/internal/domain/tenant"
	tenantHandler "auth-perm/internal/domain/tenant/handler"
	tenantService "auth-perm/internal/domain/tenant/service"
	"auth-perm/internal/domain/todo"
	todoHandler "auth-perm/internal/domain/todo/handler"
	todoService "auth-perm/internal/domain/todo/service"
	infraCache "auth-perm/internal/infra/cache"
	"auth-perm/internal/infra/code_gen"
	"auth-perm/internal/infra/llm"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

// Container 依赖注入容器
type Container struct {
	*dig.Container
}

// Scheduler 定时任务调度器接口，供 main.go 通过 DI 取得并控制生命周期
type Scheduler interface {
	Start(ctx context.Context)
}

// CompositeScheduler 组合多个调度器，统一启动
type CompositeScheduler struct {
	schedulers []Scheduler
}

// NewCompositeScheduler 创建组合调度器
func NewCompositeScheduler(schedulers ...Scheduler) *CompositeScheduler {
	return &CompositeScheduler{schedulers: schedulers}
}

// Start 并发启动所有调度器，任一退出则全部停止
func (cs *CompositeScheduler) Start(ctx context.Context) {
	for _, s := range cs.schedulers {
		go s.Start(ctx)
	}
	<-ctx.Done()
}

// NewContainer 创建新的容器
func NewContainer() *Container {
	return &Container{
		Container: dig.New(),
	}
}

// BuildBaseContainer 构建共享基础容器（config、LLM、DB、Redis、cache、code_gen、所有 domain 模块）。
// API 和 Worker 容器都基于此函数构建，避免重复注册。
func BuildBaseContainer(cfg *config.Config) (*dig.Container, error) {
	container := dig.New()

	// 注册基础设施
	if err := registerInfra(container, cfg); err != nil {
		return nil, err
	}

	// 注册所有领域模块（含 repo、service、scheduler）
	if err := registerDomains(container); err != nil {
		return nil, err
	}

	return container, nil
}

// BuildAPIContainer 构建 API 服务容器。
// 在基础容器之上额外注册 HTTP 处理器和 Gin 路由引擎。
func BuildAPIContainer(cfg *config.Config) (*dig.Container, error) {
	container, err := BuildBaseContainer(cfg)
	if err != nil {
		return nil, err
	}

	// 注册HTTP处理器和中间件
	if err := registerHandlers(container); err != nil {
		return nil, err
	}

	// 注册Gin引擎
	if err := registerGinEngine(container); err != nil {
		return nil, err
	}

	return container, nil
}

// BuildWorkerContainer 构建 Worker（定时任务）服务容器。
// 在基础容器之上额外注册 CompositeScheduler，供 worker main.go 启动。
func BuildWorkerContainer(cfg *config.Config) (*dig.Container, error) {
	container, err := BuildBaseContainer(cfg)
	if err != nil {
		return nil, err
	}

	// 将多个调度器组合为一个 Scheduler，供 worker main.go 统一启动
	if err := container.Provide(func(
		todoScheduler *todoService.TodoScheduler,
		rssScheduler *newshockService.RSSScheduler,
		scoringScheduler *newshockService.ScoringScheduler,
		pmScheduler *newshockService.PolymarketScheduler,
		stockListScheduler *newshockService.StockListScheduler,
		dailyDataScheduler *newshockService.DailyDataScheduler,
		conceptScheduler *newshockService.ConceptScheduler,
		f10Scheduler *newshockService.F10DataScheduler,
		newsScheduler *newshockService.StockNewsScheduler,
	) Scheduler {
		return NewCompositeScheduler(todoScheduler, rssScheduler, scoringScheduler, pmScheduler, stockListScheduler, dailyDataScheduler, conceptScheduler, f10Scheduler, newsScheduler)
	}); err != nil {
		return nil, err
	}

	return container, nil
}

// registerInfra 注册基础设施依赖：config、LLM、DB、Redis、cache、code_gen
func registerInfra(container *dig.Container, cfg *config.Config) error {
	// 注册配置
	if err := container.Provide(func() *config.Config {
		return cfg
	}); err != nil {
		return err
	}

	// 注册 LLM 客户端
	if err := container.Provide(func(cfg *config.Config) *llm.Client {
		return llm.NewClient(&cfg.LLM)
	}); err != nil {
		return err
	}

	// 注册数据库连接
	if err := registerDatabase(container); err != nil {
		return err
	}

	// 注册Redis连接
	if err := registerRedis(container); err != nil {
		return err
	}

	// 注册缓存
	if err := registerCache(container); err != nil {
		return err
	}

	// 注册Code生成器
	if err := registerCodeGenerator(container); err != nil {
		return err
	}

	return nil
}

// registerDatabase 注册数据库相关依赖
func registerDatabase(container *dig.Container) error {
	return container.Provide(func(cfg *config.Config) (*gorm.DB, error) {
		db, err := NewDatabase(cfg)
		if err != nil {
			return nil, err
		}
		log.Println("Database registered successfully")
		return db, nil
	})
}

// registerRedis 注册Redis相关依赖
func registerRedis(container *dig.Container) error {
	return container.Provide(func(cfg *config.Config) (*redis.Client, error) {
		opts := &redis.Options{
			Addr:     cfg.Redis.GetAddr(),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			PoolSize: cfg.Redis.PoolSize,
		}
		if cfg.Redis.UseTLS {
			opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}

		client := redis.NewClient(opts)

		// 测试连接
		ctx := context.Background()
		if err := client.Ping(ctx).Err(); err != nil {
			return nil, err
		}

		log.Println("Redis registered successfully")
		return client, nil
	})
}

// registerCache 注册缓存相关依赖
func registerCache(container *dig.Container) error {
	// 注册缓存接口
	if err := container.Provide(func(cfg *config.Config, client *redis.Client) (infraCache.Cache, error) {
		switch cfg.Cache.Type {
		case "redis":
			log.Println("Using Redis Cache")
			return infraCache.NewRedisCache(client, ""), nil
		case "memory":
			log.Println("Using Memory (LRU) Cache")
			return infraCache.NewLRUCache(1000), nil
		default:
			return nil, errors.NewInternalErrorF("无效的缓存类型: %s", cfg.Cache.Type)
		}
	}); err != nil {
		return err
	}

	// 注册RedisClient包装器（用于特殊Redis操作）
	return container.Provide(func(cfg *config.Config, client *redis.Client) (*infraCache.RedisCache, error) {
		switch cfg.Cache.Type {
		case "redis":
			return infraCache.NewRedisCache(client, ""), nil
		default:
			return nil, errors.NewInternalErrorF("Redis客户端包装器仅支持Redis缓存类型，当前类型: %s", cfg.Cache.Type)
		}
	})
}

// registerCodeGenerator 注册Code生成器
func registerCodeGenerator(container *dig.Container) error {
	return container.Provide(func(client *redis.Client) code_gen.CodeGenerator {
		log.Println("CodeGenerator registered successfully")
		return code_gen.NewCodeGenerator(client)
	})
}

// registerDomains 注册所有领域模块（repo、service、scheduler）。
// 每个 domain 的 Register*Domain 函数会将该领域的 repo 和 service 注册到容器。
// scheduler 也在此注册（由各 domain 的 module.go 中 Provide），但 CompositeScheduler
// 的组装在 BuildWorkerContainer 中完成，API 容器不需要调度器。
func registerDomains(container *dig.Container) error {
	if err := cache.RegisterCacheDomain(container); err != nil {
		return err
	}
	if err := auth.RegisterAuthDomain(container); err != nil {
		return err
	}
	if err := permission.RegisterPermissionDomain(container); err != nil {
		return err
	}

	// 注册跨域接口适配：SessionService → tenant/service.SessionInvalidator
	if err := container.Provide(func(sessionSvc *service.SessionService) tenantService.SessionInvalidator {
		return sessionSvc
	}); err != nil {
		return err
	}

	if err := tenant.RegisterTenantDomain(container); err != nil {
		return err
	}
	if err := todo.RegisterTodoDomain(container); err != nil {
		return err
	}
	if err := journal.RegisterJournalDomain(container); err != nil {
		return err
	}
	if err := newshock.RegisterNewshockDomain(container); err != nil {
		return err
	}

	log.Println("Domain modules registered successfully")
	return nil
}

// registerHandlers 注册HTTP处理器
func registerHandlers(container *dig.Container) error {
	// 注册认证处理器
	if err := container.Provide(func(
		authService *service.AuthService,
		loginService *service.LoginService,
		registerService *service.RegisterService,
		sessionService *service.SessionService,
		emailService *service.EmailService,
		totpService *service.TOTPService,
		security *service.SecurityService,
		oauthService *service.OAuthService,
		passwordService *service.PasswordService,
		deviceService service.DeviceService,
		securityLogService *service.SecurityLogService,
	) *authHandler.AuthHandler {
		return authHandler.NewAuthHandler(authService, loginService, registerService, sessionService, emailService, totpService, security, oauthService, passwordService, deviceService, securityLogService)
	}); err != nil {
		return err
	}

	// 注册邮箱处理器
	if err := container.Provide(authHandler.NewEmailHandler); err != nil {
		return err
	}

	// 注册密码处理器
	if err := container.Provide(func(
		authService *service.AuthService,
		emailService *service.EmailService,
		passwordService *service.PasswordService,
	) *authHandler.PasswordHandler {
		return authHandler.NewPasswordHandler(authService, emailService, passwordService)
	}); err != nil {
		return err
	}

	// 注册TOTP处理器
	if err := container.Provide(authHandler.NewTOTPHandler); err != nil {
		return err
	}

	// 注册OAuth处理器
	if err := container.Provide(authHandler.NewOAuthHandler); err != nil {
		return err
	}

	// 注册权限处理器
	if err := container.Provide(permHandler.NewPermissionHandler); err != nil {
		return err
	}

	// 注册权限资源处理器
	if err := container.Provide(permHandler.NewPermissionResourceHandler); err != nil {
		return err
	}

	// 注册组织处理器
	if err := container.Provide(permHandler.NewOrganizationHandler); err != nil {
		return err
	}

	// 注册租户处理器
	if err := container.Provide(func(ts *tenantService.TenantService) *tenantHandler.TenantHandler {
		return tenantHandler.NewTenantHandler(ts)
	}); err != nil {
		return err
	}

	// 注册用户管理处理器
	if err := container.Provide(func(
		authService *service.AuthService,
		registerService *service.RegisterService,
	) *authHandler.UserHandler {
		return authHandler.NewUserHandler(authService, registerService)
	}); err != nil {
		return err
	}

	// 注册资源权限处理器
	if err := container.Provide(func(
		cfg *config.Config,
		permService *permissionService.PermissionService,
	) *authHandler.ResourceHandler {
		return authHandler.NewResourceHandler(cfg, permService)
	}); err != nil {
		return err
	}

	log.Println("HTTP handlers registered successfully")
	return nil
}

// registerGinEngine 注册Gin引擎
func registerGinEngine(container *dig.Container) error {
	return container.Provide(func(
		cfg *config.Config,
		redisClient *redis.Client,
		authH *authHandler.AuthHandler,
		emailH *authHandler.EmailHandler,
		passwordH *authHandler.PasswordHandler,
		totpH *authHandler.TOTPHandler,
		oauthH *authHandler.OAuthHandler,
		permissionH *permHandler.PermissionHandler,
		permissionResourceH *permHandler.PermissionResourceHandler,
		organizationH *permHandler.OrganizationHandler,
		tenantH *tenantHandler.TenantHandler,
		userH *authHandler.UserHandler,
		resourceH *authHandler.ResourceHandler,
		thTodoHandler *todoHandler.TodoHandler,
		jhJournalHandler *journalHandler.JournalHandler,
		nsNewshockHandler *newshockHandler.NewshockHandler,
		authService *service.AuthService,
		loginService *service.LoginService,
		permSvc *permissionService.PermissionService,
	) *gin.Engine {
		gin.SetMode(cfg.Server.Mode)
		engine := gin.New()

		engine.Use(middleware.RecoveryMiddleware())
		engine.Use(middleware.LoggingMiddleware(cfg))
		engine.Use(middleware.RouteLoggingMiddleware())
		engine.Use(middleware.CORSMiddleware(corsConfig(cfg)))
		rateLimitConfig := middleware.DefaultRateLimitConfig(redisClient)
		engine.Use(middleware.RateLimitMiddleware(rateLimitConfig))

		engine.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})

		controllerHttp.RegisterRoutes(engine, cfg, authH, emailH, passwordH, totpH, oauthH, permissionH, permissionResourceH, organizationH, tenantH, userH, resourceH, thTodoHandler, jhJournalHandler, nsNewshockHandler, authService, loginService, permSvc)

		log.Println("Gin engine registered successfully")
		return engine
	})
}

// corsConfig 构建 CORS 配置，合并 localhost 默认源与配置中额外源
func corsConfig(cfg *config.Config) middleware.Config {
	config := middleware.DefaultConfig()
	if cfg.Server.CORSOrigins != "" {
		origins := strings.Split(cfg.Server.CORSOrigins, ",")
		for _, o := range origins {
			o = strings.TrimSpace(o)
			if o != "" {
				config.AllowOrigins = append(config.AllowOrigins, o)
			}
		}
	}
	return config
}
