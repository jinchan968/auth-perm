package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/utils"
	authConstant "auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/param"
	"auth-perm/internal/domain/auth/repo"
	"auth-perm/internal/domain/auth/validator"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService 密码管理服务
type PasswordService struct {
	userRepo          *repo.UserRepo
	accountRepo       *repo.AccountRepo
	auditRepo         *repo.AuditLogRepo
	cache             *CacheService
	passwordValidator *validator.PasswordValidator
}

// NewPasswordService 创建密码管理服务
func NewPasswordService(
	userRepo *repo.UserRepo,
	accountRepo *repo.AccountRepo,
	auditRepo *repo.AuditLogRepo,
	cacheService *CacheService,
) *PasswordService {
	return &PasswordService{
		userRepo:          userRepo,
		accountRepo:       accountRepo,
		auditRepo:         auditRepo,
		cache:             cacheService,
		passwordValidator: validator.NewPasswordValidator(),
	}
}

// ChangePassword 修改密码
func (s *PasswordService) ChangePassword(ctx context.Context, params *param.ChangePasswordParams) error {
	// 验证参数
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// 查找用户
	user, err := s.userRepo.FindByID(ctx, params.UserID)
	if err != nil {
		return errors.WrapBizError(err, "查找用户失败")
	}
	if user == nil {
		return errors.NewNotFoundError("用户不存在")
	}

	// 查找用户的email账户
	accounts, err := s.accountRepo.FindByUserID(ctx, params.UserID)
	if err != nil {
		return errors.WrapBizError(err, "查找账户列表失败")
	}

	var emailAccount *dm.AccountDO
	for _, account := range accounts {
		if account.AccountType == authConstant.AccountTypeEmail {
			emailAccount = account
			break
		}
	}

	if emailAccount == nil {
		return errors.NewNotFoundError("邮箱账户不存在")
	}

	// 验证旧密码（在User层面）
	userDTO := user.ToDTO()
	if userDTO.PasswordHash == "" {
		return errors.NewAuthError("用户未设置密码")
	}

	// 清理输入密码前后空格
	cleanOldPassword := strings.TrimSpace(params.OldPassword)

	// 使用bcrypt验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(userDTO.PasswordHash), []byte(cleanOldPassword)); err != nil {
		return errors.NewAuthError("旧密码错误")
	}

	// 设置新密码（在User层面）
	if err := userDTO.SetPassword(params.NewPassword); err != nil {
		return errors.WrapBizError(err, "设置新密码失败")
	}

	// 保存用户
	if err := s.userRepo.Save(ctx, dm.UserFromDTO(userDTO)); err != nil {
		return errors.WrapBizError(err, "保存用户失败")
	}

	return nil
}

// ResetPassword 重置密码
func (s *PasswordService) ResetPassword(ctx context.Context, identifier, newPassword string) error {
	// 查找账户（支持邮箱和手机号）
	var account *dm.AccountDO
	var err error

	// 判断是邮箱还是手机号
	if strings.Contains(identifier, "@") {
		// 邮箱
		account, err = s.accountRepo.FindByEmail(ctx, identifier)
	} else {
		// 手机号 - 通过用户名查找用户
		user, err := s.userRepo.FindByUsername(ctx, identifier)
		if err != nil {
			return errors.WrapBizError(err, "查找用户失败")
		}
		if user == nil {
			return errors.NewNotFoundError("用户不存在")
		}

		// 查找用户的账户
		userDTO := user.ToDTO()
		accounts, err := s.accountRepo.FindByUserID(ctx, userDTO.GetID())
		if err != nil {
			return errors.WrapBizError(err, "查找账户失败")
		}

		// 找到第一个邮箱账户或手机号账户
		for _, acc := range accounts {
			if acc.User != nil && (acc.User.Email == identifier || acc.User.Phone == identifier) {
				account = acc
				break
			}
		}
	}

	if err != nil {
		return errors.WrapBizError(err, "查找账户失败")
	}
	if account == nil {
		return errors.NewNotFoundError("账户不存在")
	}

	// 设置新密码（在User层面）
	user, err := s.userRepo.FindByID(ctx, account.UserID)
	if err != nil {
		return errors.WrapBizError(err, "查找用户失败")
	}

	userDTO := user.ToDTO()
	if err := userDTO.SetPassword(newPassword); err != nil {
		return errors.WrapBizError(err, "设置新密码失败")
	}

	// 保存用户
	if err := s.userRepo.Save(ctx, dm.UserFromDTO(userDTO)); err != nil {
		return errors.WrapBizError(err, "保存用户失败")
	}

	// 异步记录审计日志
	s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
		Action:       "reset_password",
		ResourceType: "account",
		ResourceID:   account.ID,
		NewValues: dto.AuditLogValuesDTO{
			ChangedFields: map[string]interface{}{"identifier": identifier},
		},
		Success: true,
	})

	return nil
}

// RequestPasswordReset 请求密码重置
func (s *PasswordService) RequestPasswordReset(ctx context.Context, email string) error {
	// 1. 查找账户
	account, err := s.accountRepo.FindByEmail(ctx, email)
	if err != nil {
		// 出于安全考虑，即使账户不存在也返回成功
		return nil
	}

	// 2. 检查频率限制（24小时内只能请求1次）
	if s.cache != nil {
		key := fmt.Sprintf("password_reset:%s", email)
		exists, err := s.cache.Get(ctx, key)
		if err == nil && exists != "" {
			return errors.NewBusinessError("24小时内已发送过重置邮件，请稍后再试")
		}
	}

	// 3. 生成随机令牌
	token, err := dto.GenerateSecureToken(32)
	if err != nil {
		return errors.WrapBizError(err, "生成重置令牌失败")
	}

	// 4. 对令牌进行哈希（数据库中只存储哈希）
	tokenHash := utils.HashToken(token)

	// 5. 设置过期时间（24小时）
	expiresAt := time.Now().Add(24 * time.Hour)

	// 6. 保存令牌到数据库
	if err := s.accountRepo.UpdateResetPasswordToken(ctx, account.ID, tokenHash, expiresAt); err != nil {
		return errors.WrapBizError(err, "保存重置令牌失败")
	}

	// 7. 设置频率限制
	if s.cache != nil {
		key := fmt.Sprintf("password_reset:%s", email)
		if err := s.cache.SetWithTTL(ctx, key, "1", 24*time.Hour); err != nil {
			log.Printf("警告：设置密码重置频率限制失败: %v", err)
		}
	}

	// 8. TODO: 发送重置邮件（实际应用中应该调用邮件服务）
	// 这里仅记录日志，实际应用中需要集成邮件服务
	log.Printf("密码重置令牌: %s (仅用于测试，生产环境中不应记录敏感信息)", token)
	log.Printf("重置链接: https://example.com/reset-password?token=%s", token)

	return nil
}

// ResetPasswordWithToken 使用令牌重置密码
func (s *PasswordService) ResetPasswordWithToken(ctx context.Context, tokenHash, newPassword string) error {
	// 1. 查找账户
	account, err := s.accountRepo.FindByResetToken(ctx, tokenHash)
	if err != nil {
		return err
	}

	// 2. 验证密码强度
	if err := s.passwordValidator.ValidatePasswordStrength(newPassword); err != nil {
		return errors.WrapBizError(err, "密码不符合安全要求")
	}

	// 3. 设置新密码（在User层面）
	user, err := s.userRepo.FindByID(ctx, account.UserID)
	if err != nil {
		return errors.WrapBizError(err, "查找用户失败")
	}

	userDTO := user.ToDTO()
	if err := userDTO.SetPassword(newPassword); err != nil {
		return errors.WrapBizError(err, "设置新密码失败")
	}

	// 4. 保存用户
	if err := s.userRepo.Save(ctx, dm.UserFromDTO(userDTO)); err != nil {
		return errors.WrapBizError(err, "更新密码失败")
	}

	// 6. 清除重置令牌
	if err := s.accountRepo.ClearResetPasswordToken(ctx, account.ID); err != nil {
		log.Printf("警告：清除重置令牌失败: %v", err)
	}

	// 7. 记录审计日志
	if s.auditRepo != nil {
		// TODO: 记录密码重置审计日志
		_ = s.auditRepo
	}

	return nil
}
