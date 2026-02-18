package auth

import (
	"fmt"
	"log"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"auth-perm/config"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/repo"
	"auth-perm/internal/domain/auth/service"
	"auth-perm/internal/domain/auth/validator"
	"auth-perm/internal/infra/cache"
)

// RegisterAuthDomain 注册认证领域的所有依赖
func RegisterAuthDomain(container *dig.Container) error {
	// 注册仓储
	if err := container.Provide(func(db *gorm.DB) *repo.UserRepo {
		return repo.NewUserRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.AccountRepo {
		return repo.NewAccountRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.SessionRepo {
		return repo.NewSessionRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) *repo.AuditLogRepo {
		return repo.NewAuditLogRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) repo.DeviceTrustRepo {
		return repo.NewDeviceTrustRepo(db)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(cfg *config.Config) *repo.OAuthRepo {
		// 从配置中读取OAuth配置，并构造回调URL
		githubRedirectURL := cfg.OAuth.GitHub.RedirectURL
		if githubRedirectURL == "" {
			githubRedirectURL = fmt.Sprintf("https://%s:%d/api/v1/auth/oauth/github/callback", cfg.Server.Host, cfg.Server.Port)
		}

		googleRedirectURL := cfg.OAuth.Google.RedirectURL
		if googleRedirectURL == "" {
			googleRedirectURL = fmt.Sprintf("https://%s:%d/api/v1/auth/oauth/google/callback", cfg.Server.Host, cfg.Server.Port)
		}

		wechatRedirectURL := cfg.OAuth.WeChat.RedirectURL
		if wechatRedirectURL == "" {
			wechatRedirectURL = fmt.Sprintf("https://%s:%d/api/v1/auth/oauth/wechat/callback", cfg.Server.Host, cfg.Server.Port)
		}

		return repo.NewOAuthRepo(
			cfg.OAuth.GitHub.ClientID,     // GitHub Client ID
			cfg.OAuth.GitHub.ClientSecret, // GitHub Client Secret
			githubRedirectURL,             // GitHub Redirect URL
			cfg.OAuth.Google.ClientID,     // Google Client ID
			cfg.OAuth.Google.ClientSecret, // Google Client Secret
			googleRedirectURL,             // Google Redirect URL
			cfg.OAuth.WeChat.AppID,        // WeChat App ID
			cfg.OAuth.WeChat.AppSecret,    // WeChat App Secret
			wechatRedirectURL,             // WeChat Redirect URL
		)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(db *gorm.DB) repo.TOTPSecretRepo {
		return repo.NewTOTPSecretRepository(db)
	}); err != nil {
		return err
	}

	// 注册缓存服务
	if err := container.Provide(func(
		cacheSvc cache.Cache,
		redisClient *cache.RedisCache,
	) *service.CacheService {
		return service.NewCacheService(cacheSvc, redisClient)
	}); err != nil {
		return err
	}

	// 注册防暴力破解服务
	if err := container.Provide(func(
		cacheSvc cache.Cache,
	) *service.BruteForceService {
		return service.NewBruteForceService(cacheSvc)
	}); err != nil {
		return err
	}

	// 注册邮件服务
	if err := container.Provide(func(cfg *config.Config) *service.EmailService {
		// 从配置中读取SMTP配置
		emailConfig := dto.EmailConfig{
			Host:     cfg.SMTP.Host,
			Port:     cfg.SMTP.Port,
			Username: cfg.SMTP.Username,
			Password: cfg.SMTP.Password,
			From:     cfg.SMTP.From,
			FromName: cfg.SMTP.FromName,
		}
		return service.NewEmailService(emailConfig)
	}); err != nil {
		return err
	}

	// 注册密码策略
	if err := container.Provide(func() *validator.PasswordPolicy {
		return validator.DefaultPasswordPolicy()
	}); err != nil {
		return err
	}

	// 注册安全服务
	if err := container.Provide(func(
		cache *service.CacheService,
		auditRepo *repo.AuditLogRepo,
		policy *validator.PasswordPolicy,
	) *service.SecurityService {
		return service.NewSecurityService(cache, auditRepo, policy)
	}); err != nil {
		return err
	}

	// 注册TOTP服务
	if err := container.Provide(func(
		totpRepo repo.TOTPSecretRepo,
		accountRepo *repo.AccountRepo,
		userRepo *repo.UserRepo,
		cache *service.CacheService,
		security *service.SecurityService,
	) *service.TOTPService {
		return service.NewTOTPService(totpRepo, accountRepo, userRepo, cache, security)
	}); err != nil {
		return err
	}

	// 注册OAuth服务
	if err := container.Provide(func(
		userRepo *repo.UserRepo,
		accountRepo *repo.AccountRepo,
		sessionRepo *repo.SessionRepo,
		oauthRepo *repo.OAuthRepo,
		auditRepo *repo.AuditLogRepo,
		cache *service.CacheService,
	) *service.OAuthService {
		return service.NewOAuthService(userRepo, accountRepo, sessionRepo, oauthRepo, auditRepo, cache)
	}); err != nil {
		return err
	}

	// 注册密码服务
	if err := container.Provide(func(
		userRepo *repo.UserRepo,
		accountRepo *repo.AccountRepo,
		auditRepo *repo.AuditLogRepo,
		cache *service.CacheService,
	) *service.PasswordService {
		return service.NewPasswordService(userRepo, accountRepo, auditRepo, cache)
	}); err != nil {
		return err
	}

	// 注册领域服务
	if err := container.Provide(func(
		userRepo *repo.UserRepo,
		accountRepo *repo.AccountRepo,
		sessionRepo *repo.SessionRepo,
		oauthRepo *repo.OAuthRepo,
		auditRepo *repo.AuditLogRepo,
		cache *service.CacheService,
	) *service.AuthService {
		authSvc := service.NewAuthService(userRepo, accountRepo, sessionRepo, oauthRepo, auditRepo, cache)
		log.Println("AuthService registered successfully")
		return authSvc
	}); err != nil {
		return err
	}

	// 注册登录服务
	if err := container.Provide(func(
		userRepo *repo.UserRepo,
		accountRepo *repo.AccountRepo,
		sessionRepo *repo.SessionRepo,
		oauthRepo *repo.OAuthRepo,
		auditRepo *repo.AuditLogRepo,
		cache *service.CacheService,
		bruteForce *service.BruteForceService,
	) *service.LoginService {
		loginSvc := service.NewLoginService(
			userRepo,
			accountRepo,
			sessionRepo,
			oauthRepo,
			auditRepo,
			cache,
			bruteForce,
		)
		log.Println("LoginService registered successfully")
		return loginSvc
	}); err != nil {
		return err
	}

	// 注册会话服务
	if err := container.Provide(func(
		sessionRepo *repo.SessionRepo,
		auditRepo *repo.AuditLogRepo,
		cache *service.CacheService,
	) *service.SessionService {
		sessionSvc := service.NewSessionService(sessionRepo, auditRepo, cache)
		log.Println("SessionService registered successfully")
		return sessionSvc
	}); err != nil {
		return err
	}

	// 注册注册服务
	if err := container.Provide(func(
		userRepo *repo.UserRepo,
		accountRepo *repo.AccountRepo,
		auditRepo *repo.AuditLogRepo,
	) *service.RegisterService {
		registerSvc := service.NewRegisterService(userRepo, accountRepo, auditRepo)
		log.Println("RegisterService registered successfully")
		return registerSvc
	}); err != nil {
		return err
	}

	// 注册设备服务
	if err := container.Provide(func(
		sessionRepo *repo.SessionRepo,
		deviceTrustRepo repo.DeviceTrustRepo,
	) service.DeviceService {
		deviceSvc := service.NewDeviceService(sessionRepo, deviceTrustRepo)
		log.Println("DeviceService registered successfully")
		return deviceSvc
	}); err != nil {
		return err
	}

	// 注册安全日志服务
	if err := container.Provide(func(
		auditRepo *repo.AuditLogRepo,
		userRepo *repo.UserRepo,
	) *service.SecurityLogService {
		securityLogSvc := service.NewSecurityLogService(auditRepo, userRepo)
		log.Println("SecurityLogService registered successfully")
		return securityLogSvc
	}); err != nil {
		return err
	}

	log.Println("Auth domain registered successfully")
	return nil
}
