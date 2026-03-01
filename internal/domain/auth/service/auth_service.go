package service

import (
	"context"
	"log"
	"time"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/param"
	"auth-perm/internal/domain/auth/repo"

	commonConstant "auth-perm/internal/common/constant"
	authConstant "auth-perm/internal/domain/auth/constant"
)

// AuthService 认证服务
type AuthService struct {
	userRepo          *repo.UserRepo
	accountRepo       *repo.AccountRepo
	sessionRepo       *repo.SessionRepo
	oauthRepo         *repo.OAuthRepo
	auditRepo         *repo.AuditLogRepo
	cache             *CacheService
	bruteForceService *BruteForceService
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo *repo.UserRepo,
	accountRepo *repo.AccountRepo,
	sessionRepo *repo.SessionRepo,
	oauthRepo *repo.OAuthRepo,
	auditRepo *repo.AuditLogRepo,
	cacheService *CacheService,
) *AuthService {
	bruteForceService := NewBruteForceService(cacheService.cache)

	return &AuthService{
		userRepo:          userRepo,
		accountRepo:       accountRepo,
		sessionRepo:       sessionRepo,
		oauthRepo:         oauthRepo,
		auditRepo:         auditRepo,
		cache:             cacheService,
		bruteForceService: bruteForceService,
	}
}

// FindUserByID finds a user by their ID.
func (s *AuthService) FindUserByID(ctx context.Context, userID string) (*dto.UserDTO, error) {
	// 优先从缓存获取
	if s.cache != nil {
		if cachedUser, err := s.cache.GetUserByID(ctx, userID); err == nil {
			return cachedUser, nil
		}
	}

	// 缓存未命中，从数据库查询
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	userDTO := user.ToDTO()

	// 写入缓存（24小时TTL）
	if s.cache != nil {
		err := s.cache.SetUser(ctx, userDTO, commonConstant.SessionExpiryDefault)
		if err != nil {
			return nil, err
		}
	}

	return userDTO, nil
}

// FindAccountByID 根据账户ID查找账户
func (s *AuthService) FindAccountByID(ctx context.Context, accountID string) (*dto.AccountDTO, error) {
	// 优先从缓存获取
	if s.cache != nil {
		if cachedAccount, err := s.cache.GetAccountByID(ctx, accountID); err == nil {
			return cachedAccount, nil
		}
	}

	// 缓存未命中，从数据库查询
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, nil
	}

	accountDTO := account.ToDTO()

	// 写入缓存（24小时TTL）
	if s.cache != nil {
		err := s.cache.SetAccount(ctx, accountDTO, commonConstant.SessionExpiryDefault)
		if err != nil {
			return nil, err
		}
	}

	return accountDTO, nil
}

// FindAccountByEmail 根据邮箱查找账户
func (s *AuthService) FindAccountByEmail(ctx context.Context, email string) (*dto.AccountDTO, error) {
	account, err := s.accountRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return account.ToDTO(), nil
}

// UpdateEmailVerificationStatus 更新邮箱验证状态
func (s *AuthService) UpdateEmailVerificationStatus(ctx context.Context, email string, verified bool) error {
	account, err := s.accountRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	return s.accountRepo.UpdateEmailVerificationStatus(ctx, account.ID, verified)
}

// UpdateProfile 更新用户资料
func (s *AuthService) UpdateProfile(ctx context.Context, params *param.UpdateProfileParams) (*dto.UserDTO, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 查找用户
	user, err := s.userRepo.FindByID(ctx, params.UserID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查找用户失败")
	}
	if user == nil {
		return nil, errors.NewNotFoundError("用户不存在")
	}

	// 更新用户信息
	userDTO := user.ToDTO()
	userDTO.SetNickname(params.Nickname)
	userDTO.SetPhone(params.Phone)
	userDTO.SetAvatar(params.Avatar)
	userDTO.UpdatedAt = time.Now()

	// 保存到数据库
	if err := s.userRepo.Save(ctx, dm.UserFromDTO(userDTO)); err != nil {
		return nil, errors.WrapBizError(err, "更新用户信息失败")
	}

	// 更新缓存
	if s.cache != nil {
		err := s.cache.SetUser(ctx, userDTO, commonConstant.SessionExpiryDefault)
		if err != nil {
			log.Printf("更新用户缓存失败: %v", err)
		}
	}

	return userDTO, nil
}

// ==================== 用户管理方法 ====================

// ListAccountsByTenant 根据租户ID列出账户（含用户信息）
func (s *AuthService) ListAccountsByTenant(ctx context.Context, query *dto.AccountSearchQueryDTO) ([]*dto.UserWithAccountDTO, int64, error) {
	// 调用仓储搜索账户
	accounts, total, err := s.accountRepo.SearchAccountsWithCount(ctx, query)
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "搜索账户失败")
	}

	if len(accounts) == 0 {
		return []*dto.UserWithAccountDTO{}, total, nil
	}

	// 收集所有用户ID
	userIDs := make([]string, 0, len(accounts))
	for _, account := range accounts {
		userIDs = append(userIDs, account.UserID)
	}

	// 批量加载用户信息
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "批量加载用户信息失败")
	}

	// 构建用户ID到用户的映射
	userMap := make(map[string]*dm.UserDO)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// 组合账户和用户信息
	result := make([]*dto.UserWithAccountDTO, 0, len(accounts))
	for _, account := range accounts {
		user, exists := userMap[account.UserID]
		if !exists {
			continue
		}

		result = append(result, &dto.UserWithAccountDTO{
			AccountID:     account.ID,
			TenantID:      account.TenantID,
			AccountType:   account.AccountType,
			AccountStatus: account.Status,
			EmailVerified: account.EmailVerified,
			LastLoginAt:   account.LastLoginAt,
			UserID:        user.ID,
			Username:      dm.StrVal(user.Username),
			Nickname:      dm.StrVal(user.Nickname),
			Avatar:        dm.StrVal(user.Avatar),
			Email:         dm.StrVal(user.Email),
			Phone:         dm.StrVal(user.Phone),
			UserStatus:    user.Status,
			CreatedAt:     account.CreatedAt,
			UpdatedAt:     account.UpdatedAt,
		})
	}

	return result, total, nil
}

// GetAccountWithUser 获取账户详情（含用户信息）
func (s *AuthService) GetAccountWithUser(ctx context.Context, accountID string) (*dto.UserWithAccountDTO, error) {
	// 查找账户
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查找账户失败")
	}

	// 查找用户
	user, err := s.userRepo.FindByID(ctx, account.UserID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查找用户失败")
	}

	return &dto.UserWithAccountDTO{
		AccountID:     account.ID,
		TenantID:      account.TenantID,
		AccountType:   account.AccountType,
		AccountStatus: account.Status,
		EmailVerified: account.EmailVerified,
		LastLoginAt:   account.LastLoginAt,
		UserID:        user.ID,
		Username:      dm.StrVal(user.Username),
		Nickname:      dm.StrVal(user.Nickname),
		Avatar:        dm.StrVal(user.Avatar),
		Email:         dm.StrVal(user.Email),
		Phone:         dm.StrVal(user.Phone),
		UserStatus:    user.Status,
		CreatedAt:     account.CreatedAt,
		UpdatedAt:     account.UpdatedAt,
	}, nil
}

// UpdateAccountStatus 更新账户状态（带租户校验）
func (s *AuthService) UpdateAccountStatus(ctx context.Context, accountID, tenantID string, status authConstant.AccountStatus) error {
	// 查找账户
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return errors.WrapBizError(err, "查找账户失败")
	}

	// 校验租户归属（防越权）
	if account.TenantID != tenantID {
		return errors.NewPermissionError("无权操作此账户")
	}

	// 更新状态
	if err := s.accountRepo.UpdateStatus(ctx, accountID, status); err != nil {
		return errors.WrapBizError(err, "更新账户状态失败")
	}

	// 清除账户缓存
	if s.cache != nil {
		if err := s.cache.DeleteAccount(ctx, accountID); err != nil {
			log.Printf("清除账户缓存失败: %v", err)
		}
	}

	// 当账户被禁用/暂停时，使该账户所有会话失效
	if !status.IsActive() {
		// 先查找会话用于清理缓存
		sessions, err := s.sessionRepo.FindByAccountID(ctx, accountID)
		if err != nil {
			log.Printf("查找账户会话失败: %v", err)
		}

		// 批量使会话失效
		if err := s.sessionRepo.InvalidateAccountSessions(ctx, accountID); err != nil {
			log.Printf("使账户会话失效失败: %v", err)
		}

		// 清理会话缓存
		if s.cache != nil && len(sessions) > 0 {
			for _, session := range sessions {
				sessionDTO := session.ToDTO()
				_ = s.cache.DeleteSession(ctx, sessionDTO.ID, sessionDTO.GetTenantID())
			}
		}

		log.Printf("账户 %s 已被设置为 %s，所有会话已失效", accountID, status)
	}

	return nil
}
