package service

import (
	"context"
	"log"
	"time"

	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/repo"

	commonConstant "auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/auth/param"
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
