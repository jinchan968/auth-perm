package service

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	commonConstant "auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/utils"
	"auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/param"
	"auth-perm/internal/domain/auth/repo"

	"golang.org/x/crypto/bcrypt"
)

// LoginService 登录服务
type LoginService struct {
	userRepo          *repo.UserRepo
	accountRepo       *repo.AccountRepo
	sessionRepo       *repo.SessionRepo
	oauthRepo         *repo.OAuthRepo
	auditRepo         *repo.AuditLogRepo
	cache             *CacheService
	bruteForceService *BruteForceService
}

// NewLoginService 创建登录服务
func NewLoginService(
	userRepo *repo.UserRepo,
	accountRepo *repo.AccountRepo,
	sessionRepo *repo.SessionRepo,
	oauthRepo *repo.OAuthRepo,
	auditRepo *repo.AuditLogRepo,
	cacheService *CacheService,
	bruteForceService *BruteForceService,
) *LoginService {

	return &LoginService{
		userRepo:          userRepo,
		accountRepo:       accountRepo,
		sessionRepo:       sessionRepo,
		oauthRepo:         oauthRepo,
		auditRepo:         auditRepo,
		cache:             cacheService,
		bruteForceService: bruteForceService,
	}
}

// Login 邮箱/手机号登录
func (s *LoginService) Login(ctx context.Context, params *param.LoginParams) (*dto.UserDTO, *dto.AccountDTO, error) {
	// 检查登录尝试次数
	allowed, remaining, err := s.bruteForceService.CheckLoginAttempt(ctx, params.Identifier)
	if err != nil {
		// 记录登录失败
		_ = s.bruteForceService.RecordFailedLogin(ctx, params.Identifier, params.IPAddress, constant.LoginFailureReasonCheckAttemptFailed)
		return nil, nil, err
	}
	if !allowed {
		// 记录登录失败
		_ = s.bruteForceService.RecordFailedLogin(ctx, params.Identifier, params.IPAddress, constant.LoginFailureReasonTooManyAttempts)
		return nil, nil, errors.NewAuthErrorF("登录尝试次数过多，剩余尝试次数: %d", remaining)
	}

	// 获取租户ID（支持多租户登录）
	tenantID := params.TenantID
	if tenantID == "" {
		// 如果没有传递tenant_id，使用默认租户ID
		tenantID = commonConstant.DefaultTenantID
	}

	// 查找账户（自动识别邮箱或手机号）
	account, err := s.accountRepo.FindByIdentifier(ctx, params.Identifier)
	if err != nil {
		// 记录登录失败
		_ = s.bruteForceService.RecordFailedLogin(ctx, params.Identifier, params.IPAddress, constant.LoginFailureReasonAccountLookupFailed)
		return nil, nil, errors.WrapBizError(err, "查找账户失败")
	}

	// 查找用户
	user, err := s.userRepo.FindByID(ctx, account.UserID)
	if err != nil {
		// 记录登录失败
		_ = s.bruteForceService.RecordFailedLogin(ctx, params.Identifier, params.IPAddress, constant.LoginFailureReasonUserLookupFailed)
		return nil, nil, errors.WrapBizError(err, "查找用户失败")
	}

	// 检查用户状态
	userDTO := user.ToDTO()
	if !userDTO.IsActive() {
		// 记录登录失败
		_ = s.bruteForceService.RecordFailedLogin(ctx, params.Identifier, params.IPAddress, constant.LoginFailureReasonUserInactive)
		return nil, nil, errors.NewAuthError("用户已被禁用")
	}

	// 验证用户密码（在User层面）
	if userDTO.PasswordHash == "" {
		// 记录登录失败
		_ = s.bruteForceService.RecordFailedLogin(ctx, params.Identifier, params.IPAddress, constant.LoginFailureReasonPasswordNotSet)
		return nil, nil, errors.NewAuthError("用户名或密码错误")
	}

	// 清理输入密码前后空格（防止隐藏字符导致验证失败）
	cleanPassword := strings.TrimSpace(params.Password)

	// 使用bcrypt验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(userDTO.PasswordHash), []byte(cleanPassword)); err != nil {
		// 记录登录失败
		_ = s.bruteForceService.RecordFailedLogin(ctx, params.Identifier, params.IPAddress, constant.LoginFailureReasonPasswordIncorrect)
		// 返回统一的错误消息，避免泄露具体错误信息
		return nil, nil, errors.NewAuthError("用户名或密码错误")
	}

	// 检查账户状态
	accountDTO := account.ToDTO()
	if !accountDTO.IsActive() {
		// 记录登录失败
		_ = s.bruteForceService.RecordFailedLogin(ctx, params.Identifier, params.IPAddress, constant.LoginFailureReasonAccountInactive)
		return nil, nil, errors.NewAuthError("账户已被禁用")
	}

	// 记录登录成功（清除失败记录）
	_ = s.bruteForceService.RecordSuccessfulLogin(ctx, params.Identifier)

	// 更新最后登录时间
	accountDTO.UpdateLastLogin()
	// TODO: 需要重构accountDTO转换为accountDO的逻辑
	// accountDO := transformer.ToAccountDO(accountDTO)
	// if err := s.accountRepo.Save(ctx, accountDO); err != nil {
	// 	return nil, nil, errors.WrapBizError(err, "更新最后登录时间失败")
	// }

	// 记录审计日志
	s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
		Action:       constant.ActionLogin,
		ResourceType: constant.AuditResourceAccount,
		ResourceID:   account.ID,
		UserID:       user.ID,
		Success:      true,
		IPAddress:    params.IPAddress,
		UserAgent:    params.UserAgent,
	})

	return userDTO, accountDTO, nil
}

// CreateSessionAndToken 创建会话并生成token
func (s *LoginService) CreateSessionAndToken(
	params *param.SessionTokenParams,
) (*dto.LoginResult, error) {
	// 创建会话
	expiresIn := commonConstant.TokenExpiryDefault
	if params.RememberMe {
		expiresIn = commonConstant.TokenExpiryRememberMe
	}

	sessionParams := param.NewCreateSessionParamsWithUsername(
		params.User.GetID(),
		params.Account.ID,
		params.TenantID,
		params.User.Username, // 传入 username
		params.IPAddress,
		params.UserAgent,
		expiresIn,
	)
	session, err := s.CreateSession(params.Context, sessionParams)
	if err != nil {
		return nil, errors.WrapBizError(err, "创建会话失败")
	}

	// 生成token
	token := session.TokenHash + ":" + session.ID

	return &dto.LoginResult{
		User:    params.User,
		Account: params.Account,
		Token:   token,
		Session: session,
	}, nil
}

// CreateSession 创建会话
func (s *LoginService) CreateSession(ctx context.Context, params *param.CreateSessionParams) (*dto.SessionDTO, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 会话固定攻击防护：在创建新会话前，使该用户在当前租户下的所有旧会话失效
	if err := s.sessionRepo.InvalidateUserTenantSessions(ctx, params.UserID, params.TenantID); err != nil {
		s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
			Action:       constant.ActionInvalidateSessionsError,
			ResourceType: constant.AuditResourceSession,
			ResourceID:   params.UserID,
			Success:      false,
		})
	}

	// 如果有缓存，清理该用户在当前租户下的旧会话缓存
	oldSessions, err := s.sessionRepo.FindByUserIDAndTenantID(ctx, params.UserID, params.TenantID)
	if err == nil && len(oldSessions) > 0 {
		for _, oldSession := range oldSessions {
			_ = s.cache.DeleteSession(ctx, oldSession.ID, oldSession.TenantID)
			_ = s.cache.Delete(ctx, s.cache.TokenHashCacheKey(oldSession.TokenHash))
		}
	}

	// 创建新会话
	sessionDTO := dto.NewSessionDTO(params.UserID, params.AccountID, params.TenantID, time.Now().Add(params.ExpiresIn))
	sessionDTO.Username = params.Username // 设置 username 用于超管判断

	// 生成 token 和 tokenHash
	token, err := dto.GenerateSecureToken(commonConstant.DefaultTokenLength)
	if err != nil {
		return nil, errors.WrapBizError(err, "生成会话令牌失败")
	}
	tokenHash := utils.HashToken(token)
	sessionDTO.SetTokenHash(tokenHash)

	// 设置设备信息
	if err := sessionDTO.SetDeviceInfo(params.IPAddress, params.UserAgent); err != nil {
		return nil, errors.WrapBizError(err, "设置设备信息失败")
	}

	// 保存会话到数据库
	sessionDO := dm.SessionFromDTO(sessionDTO)
	if err := s.sessionRepo.Save(ctx, sessionDO); err != nil {
		return nil, errors.WrapBizError(err, "保存会话失败")
	}

	// 更新 DTO 中的 ID（从数据库返回的 DO 中获取）
	sessionDTO.ID = sessionDO.ID

	// 缓存会话信息（完整session数据，7天TTL）
	ttl := time.Until(sessionDTO.ExpiresAt)
	if ttl > commonConstant.TokenExpiryRememberMe {
		ttl = commonConstant.TokenExpiryRememberMe
	}

	err = s.cache.SetSession(ctx, sessionDTO, ttl)
	if err != nil {
		return nil, errors.WrapBizError(err, "缓存会话失败")
	}

	// 异步记录审计日志
	s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
		Action:       constant.ActionCreateSession,
		ResourceType: constant.AuditResourceSession,
		ResourceID:   sessionDTO.ID,
		Success:      true,
	})

	return sessionDTO, nil
}

// ValidateSession 验证会话（优化版：优先使用缓存，减少数据库查询）
func (s *LoginService) ValidateSession(ctx context.Context, tokenHash string) (*dto.SessionDTO, error) {
	// 优先从缓存查找完整session信息
	tokenKey := s.cache.TokenHashCacheKey(tokenHash)
	dataStr, err := s.cache.Get(ctx, tokenKey)
	if err == nil {
		var cache dto.SessionCache
		if json.Unmarshal([]byte(dataStr), &cache) == nil && cache.IsValid() {
			return cache.ToSessionDTO(tokenHash), nil
		}
	}

	// 缓存未命中，从数据库查找（兜底方案）
	session, err := s.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.WrapBizError(err, "查找会话失败")
	}
	if session == nil {
		return nil, errors.NewNotFoundError("会话不存在")
	}
	sessionDTO := session.ToDTO()

	// 验证会话是否有效
	if !sessionDTO.IsValid() {
		_ = s.cache.DeleteSession(ctx, sessionDTO.ID, sessionDTO.GetTenantID())
		return nil, errors.NewAuthError("会话无效或已过期")
	}

	// 更新最后活动时间
	sessionDTO.UpdateLastActivity()
	if err := s.sessionRepo.Save(ctx, dm.SessionFromDTO(sessionDTO)); err != nil {
		log.Printf("更新会话最后活动时间失败: %v", err)
	}

	// 更新缓存（缓存完整session信息）
	remainingTTL := time.Until(sessionDTO.ExpiresAt)
	if remainingTTL > 0 {
		if err := s.cache.SetSession(ctx, sessionDTO, remainingTTL); err != nil {
			log.Printf("更新session缓存失败: %v", err)
		}
	}

	return sessionDTO, nil
}

// RefreshToken 刷新令牌
func (s *LoginService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// 通过token hash查找会话
	session, err := s.sessionRepo.FindByTokenHash(ctx, refreshToken)
	if err != nil {
		return "", errors.WrapBizError(err, "查找会话失败")
	}
	if session == nil {
		return "", errors.NewAuthError("无效的刷新令牌")
	}

	// 验证会话是否有效
	sessionDTO := session.ToDTO()
	if !sessionDTO.IsValid() {
		return "", errors.NewAuthError("会话已过期")
	}

	// 生成新的访问令牌
	newToken, err := dto.GenerateSecureToken(commonConstant.DefaultTokenLength)
	if err != nil {
		return "", errors.WrapBizError(err, "生成新令牌失败")
	}

	// 更新会话的token hash
	newTokenHash := utils.HashToken(newToken)
	sessionDTO.TokenHash = newTokenHash
	sessionDTO.UpdateLastActivity()

	// 延长会话有效期（如果需要）
	if err := sessionDTO.Extend(commonConstant.SessionExpiryDefault); err != nil {
		return "", errors.WrapBizError(err, "延长会话失败")
	}

	// 删除旧token的缓存
	_ = s.cache.DeleteSession(ctx, sessionDTO.ID, sessionDTO.GetTenantID())
	_ = s.cache.Delete(ctx, s.cache.TokenHashCacheKey(session.TokenHash))

	// 保存会话到数据库
	if err := s.sessionRepo.Save(ctx, dm.SessionFromDTO(sessionDTO)); err != nil {
		return "", errors.WrapBizError(err, "更新会话失败")
	}

	// 更新缓存（使用新token信息）
	remainingTTL := time.Until(sessionDTO.ExpiresAt)
	if remainingTTL > 0 {
		if err := s.cache.SetSession(ctx, sessionDTO, remainingTTL); err != nil {
			log.Printf("更新session缓存失败: %v", err)
		}
	}

	// 异步记录审计日志
	s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
		Action:       constant.ActionRefreshToken,
		ResourceType: constant.AuditResourceSession,
		ResourceID:   session.ID,
		Success:      true,
	})

	return newToken, nil
}
