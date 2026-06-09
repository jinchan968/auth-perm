package service

import (
	"context"
	stdErr "errors"

	"auth-perm/internal/common/errors"
	authConstant "auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/param"
	"auth-perm/internal/domain/auth/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegisterService 注册服务
type RegisterService struct {
	userRepo       *repo.UserRepo
	accountRepo    *repo.AccountRepo
	auditRepo      *repo.AuditLogRepo
	invitationRepo *repo.RegistrationInvitationRepo
	invitationSvc  *RegistrationInvitationService
}

// NewRegisterService 创建注册服务
func NewRegisterService(
	userRepo *repo.UserRepo,
	accountRepo *repo.AccountRepo,
	auditRepo *repo.AuditLogRepo,
	invitationRepo *repo.RegistrationInvitationRepo,
	invitationSvc *RegistrationInvitationService,
) *RegisterService {
	return &RegisterService{
		userRepo:       userRepo,
		accountRepo:    accountRepo,
		auditRepo:      auditRepo,
		invitationRepo: invitationRepo,
		invitationSvc:  invitationSvc,
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

// RegisterWithInvitation 公开注册入口：必须使用有效邀请码，并在同一事务中消费。
func (s *RegisterService) RegisterWithInvitation(ctx context.Context, params *param.RegisterParams, inviteCode string) (*dto.UserDTO, *dto.AccountDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, nil, errors.NewValidationError(err.Error())
	}
	if inviteCode == "" {
		return nil, nil, errors.NewValidationError("邀请码不能为空")
	}

	var user *dto.UserDTO
	var accountDTO *dto.AccountDTO
	var userDO *dm.UserDO
	var accountDO *dm.AccountDO
	var userID string

	err := s.userRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		invitation, err := s.invitationSvc.findForRegistrationWithTx(ctx, tx, inviteCode)
		if err != nil {
			return err
		}

		if params.TenantID != "" && params.TenantID != invitation.TenantID {
			return errors.NewValidationError("租户 ID 与邀请码不匹配")
		}
		params.TenantID = invitation.TenantID
		accountType := params.IdentifierType.ToAccountType()

		existingUser, existingAccount, err := s.findExistingUserAndAccountWithTx(ctx, tx, params.Username, params.Identifier, params.TenantID)
		if err != nil {
			return errors.WrapBizError(err, "查找用户失败")
		}
		if existingUser != nil {
			if existingAccount != nil {
				return errors.NewBusinessError("该账户已注册")
			}
			accountDTO = dto.NewAccountDTO(existingUser.ID, params.TenantID, accountType)
			if err := s.accountRepo.SaveWithTx(ctx, tx, dm.AccountFromDTO(accountDTO)); err != nil {
				return errors.WrapBizError(err, "创建账户失败")
			}
			userDO = existingUser
			accountDO = dm.AccountFromDTO(accountDTO)
			userID = existingUser.ID
		} else {
			userID = uuid.New().String()
			user, err = dto.NewUserDTO(params.Username, params.IdentifierType, params.Identifier)
			if err != nil {
				return errors.WrapBizError(err, "创建用户失败")
			}
			user.WithNewID(userID)
			if err := user.SetPassword(params.Password); err != nil {
				return errors.WrapBizError(err, "设置密码失败")
			}
			if err := s.userRepo.SaveWithTx(ctx, tx, dm.UserFromDTO(user)); err != nil {
				return errors.WrapBizError(err, "保存用户失败")
			}

			accountDTO = dto.NewAccountDTO(userID, params.TenantID, accountType)
			if err := s.accountRepo.SaveWithTx(ctx, tx, dm.AccountFromDTO(accountDTO)); err != nil {
				return errors.WrapBizError(err, "创建账户失败")
			}
			userDO = dm.UserFromDTO(user)
			accountDO = dm.AccountFromDTO(accountDTO)
		}

		if err := s.invitationRepo.MarkUsedWithTx(ctx, tx, invitation.ID, accountDTO.ID); err != nil {
			return errors.WrapBizError(err, "消费邀请码失败")
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
		Action:       authConstant.ActionRegister,
		ResourceType: authConstant.AuditResourceUser,
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

	return userDO.ToDTO(), accountDO.ToDTO(), nil
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

func (s *RegisterService) findExistingUserAndAccountWithTx(ctx context.Context, tx *gorm.DB, username, identifier, tenantID string) (*dm.UserDO, *dm.AccountDO, error) {
	var existingUser dm.UserDO
	err := tx.WithContext(ctx).Where("username = ?", username).First(&existingUser).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	if !s.isIdentifierMatch(&existingUser, identifier) {
		return &existingUser, nil, errors.NewBusinessError("用户名已存在")
	}

	var existingAccount dm.AccountDO
	err = tx.WithContext(ctx).
		Where("user_id = ?", existingUser.ID).
		Where("tenant_id = ?", tenantID).
		Take(&existingAccount).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return &existingUser, nil, nil
		}
		return nil, nil, err
	}

	return &existingUser, &existingAccount, nil
}

// isIdentifierMatch 检查标识符是否匹配用户
func (s *RegisterService) isIdentifierMatch(user *dm.UserDO, identifier string) bool {
	// 检查邮箱匹配
	if dm.StrVal(user.Email) == identifier {
		return true
	}
	// 检查手机号匹配
	if dm.StrVal(user.Phone) == identifier {
		return true
	}
	return false
}

// createAccountForExistingUser 为现有用户创建账户
func (s *RegisterService) createAccountForExistingUser(ctx context.Context, userID, tenantID string, accountType authConstant.AccountType, password string) (*dto.AccountDTO, error) {
	// 创建账户
	accountDTO := dto.NewAccountDTO(userID, tenantID, accountType)

	// 保存账户（不需要在Account层面设置密码）
	if err := s.accountRepo.Save(ctx, dm.AccountFromDTO(accountDTO)); err != nil {
		return nil, errors.WrapBizError(err, "创建账户失败")
	}

	// 异步记录审计日志
	s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
		Action:       authConstant.ActionRegister,
		ResourceType: authConstant.AuditResourceAccount,
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
		Action:       authConstant.ActionRegister,
		ResourceType: authConstant.AuditResourceUser,
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
