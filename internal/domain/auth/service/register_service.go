package service

import (
	"context"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/param"
	"auth-perm/internal/domain/auth/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegisterService 注册服务
type RegisterService struct {
	userRepo    *repo.UserRepo
	accountRepo *repo.AccountRepo
	auditRepo   *repo.AuditLogRepo
}

// NewRegisterService 创建注册服务
func NewRegisterService(
	userRepo *repo.UserRepo,
	accountRepo *repo.AccountRepo,
	auditRepo *repo.AuditLogRepo,
) *RegisterService {
	return &RegisterService{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		auditRepo:   auditRepo,
	}
}

// Register 通用注册（支持邮箱和手机号）
func (s *RegisterService) Register(ctx context.Context, params *param.RegisterParams) (*dto.UserDTO, *dto.AccountDTO, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, nil, errors.NewValidationError(err.Error())
	}

	// 根据标识符类型确定账户类型
	accountType := params.IdentifierType.ToAccountType()

	// 查找现有用户（用户名 + 标识符联合查询）
	existingUser, existingAccount, err := s.findExistingUserAndAccount(ctx, params.Username, params.Identifier, params.TenantID)
	if err != nil {
		return nil, nil, errors.WrapBizError(err, "查找用户失败")
	}

	// 用户已存在的情况
	if existingUser != nil {
		// 如果用户已存在但账户不存在，创建账户
		if existingAccount == nil {
			accountDTO, err := s.createAccountForExistingUser(ctx, existingUser.ID, params.TenantID, accountType, params.Password)
			if err != nil {
				return nil, nil, err
			}
			return existingUser.ToDTO(), accountDTO, nil
		}
		// 用户和账户都存在，返回错误（重复注册）
		return nil, nil, errors.NewBusinessError("该账户已注册")
	}

	// 创建新用户和新账户
	creationParams := param.NewUserCreationParams(ctx, params.Username, params.IdentifierType, params.Identifier, params.TenantID, accountType, params.Password)
	return s.createNewUserAndAccount(creationParams)
}

// findExistingUserAndAccount 查找现有用户和账户（使用JOIN优化查询）
func (s *RegisterService) findExistingUserAndAccount(ctx context.Context, username, identifier, tenantID string) (*dm.UserDO, *dm.AccountDO, error) {
	// 通过用户名查找用户
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, nil, nil // 用户不存在，正常情况
		}
		return nil, nil, err
	}

	// 用户存在，检查标识符是否匹配
	if !s.isIdentifierMatch(existingUser, identifier) {
		return existingUser, nil, errors.NewBusinessError("用户名已存在")
	}

	// 查找用户的账户
	existingAccount, err := s.accountRepo.FetchWithTenantUser(ctx, existingUser.ID, tenantID)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return existingUser, nil, nil // 用户存在但账户不存在，正常情况
		}
		return nil, nil, err
	}

	return existingUser, existingAccount, nil
}

// isIdentifierMatch 检查标识符是否匹配用户
func (s *RegisterService) isIdentifierMatch(user *dm.UserDO, identifier string) bool {
	// 检查邮箱匹配
	if user.Email == identifier {
		return true
	}
	// 检查手机号匹配
	if user.Phone == identifier {
		return true
	}
	return false
}

// createAccountForExistingUser 为现有用户创建账户
func (s *RegisterService) createAccountForExistingUser(ctx context.Context, userID, tenantID string, accountType constant.AccountType, password string) (*dto.AccountDTO, error) {
	// 创建账户
	accountDTO := dto.NewAccountDTO(userID, tenantID, accountType)

	// 保存账户（不需要在Account层面设置密码）
	if err := s.accountRepo.Save(ctx, dm.AccountFromDTO(accountDTO)); err != nil {
		return nil, errors.WrapBizError(err, "创建账户失败")
	}

	// 异步记录审计日志
	s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
		Action:       "register",
		ResourceType: "account",
		ResourceID:   accountDTO.ID,
		NewValues: dto.AuditLogValuesDTO{
			ChangedFields: map[string]interface{}{
				"user_id":      userID,
				"tenant_id":    tenantID,
				"account_type": accountType,
			},
		},
		Success: true,
	})

	return accountDTO, nil
}

// createNewUserAndAccount 创建新用户和新账户（使用事务保证数据一致性）
func (s *RegisterService) createNewUserAndAccount(
	params *param.UserCreationParams,
) (*dto.UserDTO, *dto.AccountDTO, error) {
	// 使用数据库事务确保数据一致性
	var user *dto.UserDTO
	var accountDTO *dto.AccountDTO
	var userDO *dm.UserDO
	var accountDO *dm.AccountDO

	// 生成用户ID
	userID := uuid.New().String()

	// 在事务中执行所有数据库操作
	err := s.userRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		var err error

		// 1. 创建用户
		user, err = dto.NewUserDTO(params.Username, params.IdentifierType, params.Identifier)
		if err != nil {
			return errors.WrapBizError(err, "创建用户失败")
		}
		user.WithNewID(userID)

		// 设置密码（在User层面）
		if err := user.SetPassword(params.Password); err != nil {
			// 事务会自动回滚
			return errors.WrapBizError(err, "设置密码失败")
		}

		// 保存用户到数据库
		if err := s.userRepo.SaveWithTx(params.Context, tx, dm.UserFromDTO(user)); err != nil {
			return errors.WrapBizError(err, "保存用户失败")
		}

		// 2. 创建账户
		accountDTO = dto.NewAccountDTO(userID, params.TenantID, params.AccountType)

		// 保存账户到数据库（不需要在Account层面设置密码）
		if err := s.accountRepo.SaveWithTx(params.Context, tx, dm.AccountFromDTO(accountDTO)); err != nil {
			// 事务会自动回滚
			return errors.WrapBizError(err, "创建账户失败")
		}

		// 保存引用，用于事务外使用
		userDO = dm.UserFromDTO(user)
		accountDO = dm.AccountFromDTO(accountDTO)

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// 异步记录审计日志（事务外，不影响主流程）
	s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
		Action:       "register",
		ResourceType: "user",
		ResourceID:   userID,
		NewValues: dto.AuditLogValuesDTO{
			ChangedFields: map[string]interface{}{
				"username":   params.Username,
				"tenant_id":  params.TenantID,
				"identifier": params.Identifier,
			},
		},
		Success: true,
	})

	// 转换回DTO并返回
	return userDO.ToDTO(), accountDO.ToDTO(), nil
}
