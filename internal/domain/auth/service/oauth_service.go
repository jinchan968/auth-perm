package service

import (
	"context"
	"time"

	commonConstant "auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	authConstant "auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/param"
	"auth-perm/internal/domain/auth/repo"
	"github.com/google/uuid"
)

// OAuthService OAuth认证服务
type OAuthService struct {
	userRepo    *repo.UserRepo
	accountRepo *repo.AccountRepo
	sessionRepo *repo.SessionRepo
	oauthRepo   *repo.OAuthRepo
	auditRepo   *repo.AuditLogRepo
	cache       *CacheService
}

// NewOAuthService 创建OAuth认证服务
func NewOAuthService(
	userRepo *repo.UserRepo,
	accountRepo *repo.AccountRepo,
	sessionRepo *repo.SessionRepo,
	oauthRepo *repo.OAuthRepo,
	auditRepo *repo.AuditLogRepo,
	cacheService *CacheService,
) *OAuthService {
	return &OAuthService{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		sessionRepo: sessionRepo,
		oauthRepo:   oauthRepo,
		auditRepo:   auditRepo,
		cache:       cacheService,
	}
}

// LoginWithOAuth OAuth登录
func (s *OAuthService) LoginWithOAuth(ctx context.Context, params *param.LoginWithOAuthParams) (*dto.UserDTO, *dto.AccountDTO, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, nil, errors.NewValidationError(err.Error())
	}

	// 查找OAuth账户
	account, err := s.accountRepo.FindByOAuth(ctx, params.Provider, params.OAuthID)
	if err != nil {
		return nil, nil, errors.WrapBizError(err, "查找OAuth账户失败")
	}

	// 账户不存在，创建新账户
	if account == nil {
		// 根据OAuth提供商确定标识类型
		var identifierType authConstant.IdentifierType
		switch params.Provider {
		case authConstant.OAuthProviderGitHub:
			identifierType = authConstant.IdentifierTypeGitHub
		case authConstant.OAuthProviderGoogle:
			identifierType = authConstant.IdentifierTypeGoogle
		case authConstant.OAuthProviderWeChat:
			identifierType = authConstant.IdentifierTypeWeChat
		default:
			return nil, nil, errors.NewValidationError("不支持的OAuth提供商")
		}

		// 创建账户
		accountID := uuid.New().String()
		account = dm.NewAccount(accountID, "", authConstant.AccountType("")) // OAuth账户无邮箱
		account.OAuthProvider = params.Provider
		account.OAuthID = params.OAuthID

		// 创建用户（使用OAuth提供商作为标识类型，OAuthID作为标识值）
		user, err := dto.NewUserDTO(params.Username, identifierType, params.OAuthID)
		if err != nil {
			return nil, nil, errors.WrapBizError(err, "创建用户失败")
		}
		if err := s.userRepo.Save(ctx, dm.UserFromDTO(user)); err != nil {
			return nil, nil, errors.WrapBizError(err, "保存用户失败")
		}

		// 关联用户和账户
		account.UserID = user.GetID()

		// 保存账户
		if err := s.accountRepo.Save(ctx, account); err != nil {
			return nil, nil, errors.WrapBizError(err, "创建OAuth账户失败")
		}

		accountDTO := account.ToDTO()
		accountDTO.VerifyEmail() // OAuth登录默认邮箱已验证

		if err := s.accountRepo.Save(ctx, dm.AccountFromDTO(accountDTO)); err != nil {
			return nil, nil, errors.WrapBizError(err, "保存OAuth账户失败")
		}

		return user, accountDTO, nil
	}

	// 账户已存在，更新最后登录时间
	accountDTO := account.ToDTO()
	accountDTO.UpdateLastLogin()
	if err := s.accountRepo.Save(ctx, dm.AccountFromDTO(accountDTO)); err != nil {
		return nil, nil, errors.WrapBizError(err, "更新最后登录时间失败")
	}

	// 查找用户
	user, err := s.userRepo.FindByID(ctx, account.UserID)
	if err != nil {
		return nil, nil, errors.WrapBizError(err, "查找用户失败")
	}

	return user.ToDTO(), accountDTO, nil
}

// GitHubOAuthCallback GitHub OAuth回调处理
func (s *OAuthService) GitHubOAuthCallback(ctx context.Context, params *param.GitHubOAuthCallbackParams) (*dto.UserDTO, *dto.AccountDTO, *dto.SessionDTO, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, nil, nil, errors.NewValidationError(err.Error())
	}

	// 1. 获取GitHub用户信息
	userInfo, err := s.oauthRepo.GetUserInfo(ctx, "github", params.Code, "")
	if err != nil {
		return nil, nil, nil, errors.WrapBizError(err, "GitHub OAuth认证失败")
	}

	// 2. 检查用户是否已存在
	existingUserDO, err := s.userRepo.FindByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, nil, nil, errors.WrapBizError(err, "查找用户失败")
	}

	var user *dto.UserDTO
	var account *dto.AccountDTO

	if existingUserDO != nil {
		// 2a. 用户已存在，登录
		user = existingUserDO.ToDTO()

		accounts, err := s.accountRepo.FindByUserID(ctx, user.GetID())
		if err != nil {
			return nil, nil, nil, errors.WrapBizError(err, "查找账户失败")
		}

		// 找到GitHub账户
		for _, acc := range accounts {
			if acc.OAuthProvider == "github" {
				account = acc.ToDTO()
				break
			}
		}

		if account == nil {
			return nil, nil, nil, errors.NewBusinessError("未找到GitHub账户")
		}
	} else {
		// 2b. 用户不存在，注册新用户
		loginParams := param.NewLoginWithOAuthParams("github", userInfo.ProviderID, userInfo.Name)
		newUser, newAccount, err := s.LoginWithOAuth(ctx, loginParams)
		if err != nil {
			return nil, nil, nil, errors.WrapBizError(err, "OAuth登录失败")
		}
		user = newUser
		account = newAccount
	}

	// 3. 创建会话
	expiresIn := commonConstant.TokenExpiryDefault
	sessionParams := param.NewCreateSessionParams(user.ID, account.ID, commonConstant.DefaultTenantID, params.IPAddress, params.UserAgent, expiresIn)
	session, err := s.CreateSession(ctx, sessionParams)
	if err != nil {
		return nil, nil, nil, errors.WrapBizError(err, "创建会话失败")
	}

	return user, account, session, nil
}

// GoogleOAuthCallback Google OAuth回调处理
func (s *OAuthService) GoogleOAuthCallback(ctx context.Context, params *param.GoogleOAuthCallbackParams) (*dto.UserDTO, *dto.AccountDTO, *dto.SessionDTO, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, nil, nil, errors.NewValidationError(err.Error())
	}

	// 1. 获取Google用户信息
	userInfo, err := s.oauthRepo.GetUserInfo(ctx, "google", params.Code, "")
	if err != nil {
		return nil, nil, nil, errors.WrapBizError(err, "Google OAuth认证失败")
	}

	// 2. 检查用户是否已存在
	existingUserDO, err := s.userRepo.FindByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, nil, nil, errors.WrapBizError(err, "查找用户失败")
	}

	var user *dto.UserDTO
	var account *dto.AccountDTO

	if existingUserDO != nil {
		// 2a. 用户已存在，登录
		user = existingUserDO.ToDTO()

		accounts, err := s.accountRepo.FindByUserID(ctx, user.GetID())
		if err != nil {
			return nil, nil, nil, errors.WrapBizError(err, "查找账户失败")
		}

		// 找到Google账户
		for _, acc := range accounts {
			if acc.OAuthProvider == "google" {
				account = acc.ToDTO()
				break
			}
		}

		if account == nil {
			return nil, nil, nil, errors.NewBusinessError("未找到Google账户")
		}
	} else {
		// 2b. 用户不存在，注册新用户
		loginParams := param.NewLoginWithOAuthParams("google", userInfo.ProviderID, userInfo.Name)
		newUser, newAccount, err := s.LoginWithOAuth(ctx, loginParams)
		if err != nil {
			return nil, nil, nil, errors.WrapBizError(err, "OAuth登录失败")
		}
		user = newUser
		account = newAccount
	}

	// 3. 创建会话
	expiresIn := commonConstant.TokenExpiryDefault
	sessionParams := param.NewCreateSessionParams(user.ID, account.ID, commonConstant.DefaultTenantID, params.IPAddress, params.UserAgent, expiresIn)
	session, err := s.CreateSession(ctx, sessionParams)
	if err != nil {
		return nil, nil, nil, errors.WrapBizError(err, "创建会话失败")
	}

	return user, account, session, nil
}

// WeChatOAuthCallback 微信OAuth回调处理
func (s *OAuthService) WeChatOAuthCallback(ctx context.Context, params *param.WeChatOAuthCallbackParams) (*dto.UserDTO, *dto.AccountDTO, *dto.SessionDTO, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, nil, nil, errors.NewValidationError(err.Error())
	}

	// 1. 获取微信用户信息
	userInfo, err := s.oauthRepo.GetUserInfo(ctx, "wechat", params.Code, "")
	if err != nil {
		return nil, nil, nil, errors.WrapBizError(err, "微信OAuth认证失败")
	}

	// 2. 检查用户是否已存在（使用虚拟邮箱查找）
	existingUserDO, err := s.userRepo.FindByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, nil, nil, errors.WrapBizError(err, "查找用户失败")
	}

	var user *dto.UserDTO
	var account *dto.AccountDTO

	if existingUserDO != nil {
		// 2a. 用户已存在，登录
		user = existingUserDO.ToDTO()

		accounts, err := s.accountRepo.FindByUserID(ctx, user.GetID())
		if err != nil {
			return nil, nil, nil, errors.WrapBizError(err, "查找账户失败")
		}

		// 找到微信账户
		for _, acc := range accounts {
			if acc.OAuthProvider == "wechat" {
				account = acc.ToDTO()
				break
			}
		}

		if account == nil {
			return nil, nil, nil, errors.NewBusinessError("未找到微信账户")
		}
	} else {
		// 2b. 用户不存在，注册新用户
		loginParams := param.NewLoginWithOAuthParams("wechat", userInfo.ProviderID, userInfo.Name)
		newUser, newAccount, err := s.LoginWithOAuth(ctx, loginParams)
		if err != nil {
			return nil, nil, nil, errors.WrapBizError(err, "微信OAuth登录失败")
		}
		user = newUser
		account = newAccount
	}

	// 3. 创建会话
	expiresIn := commonConstant.TokenExpiryDefault
	sessionParams := param.NewCreateSessionParams(user.ID, account.ID, commonConstant.DefaultTenantID, params.IPAddress, params.UserAgent, expiresIn)
	session, err := s.CreateSession(ctx, sessionParams)
	if err != nil {
		return nil, nil, nil, errors.WrapBizError(err, "创建会话失败")
	}

	return user, account, session, nil
}

// CreateSession 创建会话（OAuth服务使用）
func (s *OAuthService) CreateSession(ctx context.Context, params *param.CreateSessionParams) (*dto.SessionDTO, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 会话固定攻击防护：在创建新会话前，使该用户在当前租户下的所有旧会话失效
	if err := s.sessionRepo.InvalidateUserSessions(ctx, params.UserID); err != nil {
		// 记录错误但不影响登录流程
		s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
			Action:       "invalidate_sessions_error",
			ResourceType: "session",
			ResourceID:   params.UserID,
			Success:      false,
		})
	}

	// 如果有缓存，清理该用户的旧会话缓存
	if s.cache != nil {
		// 查找用户的旧会话并删除缓存
		oldSessions, err := s.sessionRepo.FindByUserID(ctx, params.UserID, nil)
		if err == nil && len(oldSessions) > 0 {
			for _, oldSession := range oldSessions {
				// 删除会话相关缓存
				_ = s.cache.DeleteSession(ctx, oldSession.ID, oldSession.TenantID)
				_ = s.cache.Delete(ctx, s.cache.keyGenerator.TokenHashCacheKey(oldSession.TokenHash))
			}
		}
	}

	// 创建新会话
	sessionDTO := dto.NewSessionDTO(params.UserID, params.AccountID, params.TenantID, time.Now().Add(params.ExpiresIn))

	// 设置设备信息
	if err := sessionDTO.SetDeviceInfo(params.IPAddress, params.UserAgent); err != nil {
		return nil, errors.WrapBizError(err, "设置设备信息失败")
	}

	// 保存会话到数据库
	if err := s.sessionRepo.Save(ctx, dm.SessionFromDTO(sessionDTO)); err != nil {
		return nil, errors.WrapBizError(err, "保存会话失败")
	}

	// 缓存会话信息（完整session数据，7天TTL）
	if s.cache != nil {
		// 计算实际TTL（不超过rememberMe时间）
		ttl := time.Until(sessionDTO.ExpiresAt)
		if ttl > commonConstant.TokenExpiryRememberMe {
			ttl = commonConstant.TokenExpiryRememberMe
		}

		err := s.cache.SetSession(ctx, sessionDTO, ttl)
		if err != nil {
			return nil, errors.WrapBizError(err, "缓存会话失败")
		}
	}

	// 异步记录审计日志
	s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
		Action:       "create_session",
		ResourceType: "session",
		ResourceID:   sessionDTO.ID,
		Success:      true,
	})

	return sessionDTO, nil
}
