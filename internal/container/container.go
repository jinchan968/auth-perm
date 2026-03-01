package container

import (
	"context"
	"log"

	"auth-perm/config"
	"auth-perm/internal/common/errors"
	controllerHttp "auth-perm/internal/controller/http"
	"auth-perm/internal/controller/middleware"
	"auth-perm/internal/domain/auth"
	authHandler "auth-perm/internal/domain/auth/handler"
	"auth-perm/internal/domain/auth/service"
	"auth-perm/internal/domain/permission"
	permHandler "auth-perm/internal/domain/permission/handler"
	permissionService "auth-perm/internal/domain/permission/service"
	"auth-perm/internal/domain/tenant"
	tenantHandler "auth-perm/internal/domain/tenant/handler"
	tenantService "auth-perm/internal/domain/tenant/service"
	"auth-perm/internal/domain/todo"
	todoHandler "auth-perm/internal/domain/todo/handler"
	todoService "auth-perm/internal/domain/todo/service"
	"auth-perm/internal/infra/cache"
	"auth-perm/internal/infra/code_gen"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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

// NewContainer 创建新的容器
func NewContainer() *Container {
	return &Container{
		Container: dig.New(),
	}
}

// BuildContainer 构建完整的依赖注入容器
func BuildContainer(cfg *config.Config) (*dig.Container, error) {
	container := dig.New()

	// 注册配置
	if err := container.Provide(func() *config.Config {
		return cfg
	}); err != nil {
		return nil, err
	}

	// 注册数据库连接
	if err := registerDatabase(container); err != nil {
		return nil, err
	}

	// 注册Redis连接
	if err := registerRedis(container); err != nil {
		return nil, err
	}

	// 注册缓存
	if err := registerCache(container); err != nil {
		return nil, err
	}

	// 注册Code生成器
	if err := registerCodeGenerator(container); err != nil {
		return nil, err
	}

	// 注册应用服务
	if err := registerApplicationServices(container); err != nil {
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
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.GetAddr(),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			PoolSize: cfg.Redis.PoolSize,
		})

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
	if err := container.Provide(func(cfg *config.Config, client *redis.Client) (cache.Cache, error) {
		switch cfg.Cache.Type {
		case "redis":
			log.Println("Using Redis Cache")
			return cache.NewRedisCache(client, "cache:"), nil
		case "memory":
			log.Println("Using Memory (LRU) Cache")
			return cache.NewLRUCache(1000), nil
		default:
			return nil, errors.NewInternalErrorF("无效的缓存类型: %s", cfg.Cache.Type)
		}
	}); err != nil {
		return err
	}

	// 注册RedisClient包装器（用于特殊Redis操作）
	return container.Provide(func(cfg *config.Config, client *redis.Client) (*cache.RedisCache, error) {
		switch cfg.Cache.Type {
		case "redis":
			return cache.NewRedisCache(client, "auth:"), nil
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

// registerApplicationServices 注册应用服务
func registerApplicationServices(container *dig.Container) error {
	// 应用服务可以在这里注册
	// 注册领域模块
	if err := auth.RegisterAuthDomain(container); err != nil {
		return err
	}
	if err := permission.RegisterPermissionDomain(container); err != nil {
		return err
	}

	// 注册跨域接口适配：SessionService → tenant/service.SessionInvalidator
	// SessionService 实现了 InvalidateTenantSessions 方法，满足 tenant 域的 SessionInvalidator 接口
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
	// 将 TodoScheduler 绑定到 Scheduler 接口，供 main.go 通过接口注入
	if err := container.Provide(func(s *todoService.TodoScheduler) Scheduler {
		return s
	}); err != nil {
		return err
	}
	log.Println("Application services registered successfully")
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
		authService *service.AuthService,
		permService *permissionService.PermissionService,
	) *authHandler.ResourceHandler {
		return authHandler.NewResourceHandler(cfg, authService, permService)
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
		authService *service.AuthService,
		loginService *service.LoginService,
		permSvc *permissionService.PermissionService,
	) *gin.Engine {
		gin.SetMode(cfg.Server.Mode)
		engine := gin.New()

		engine.Use(middleware.RecoveryMiddleware())
		engine.Use(middleware.LoggingMiddleware(cfg))
		engine.Use(middleware.RouteLoggingMiddleware())
		engine.Use(middleware.CORSMiddleware(middleware.DefaultConfig()))
		rateLimitConfig := middleware.DefaultRateLimitConfig(redisClient)
		engine.Use(middleware.RateLimitMiddleware(rateLimitConfig))

		engine.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})

		controllerHttp.RegisterRoutes(engine, cfg, authH, emailH, passwordH, totpH, oauthH, permissionH, permissionResourceH, organizationH, tenantH, userH, resourceH, thTodoHandler, authService, loginService, permSvc)

		log.Println("Gin engine registered successfully")
		return engine
	})
}
