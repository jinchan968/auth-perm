package service

import (
	"context"
	"log"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/tenant/constant"
	"auth-perm/internal/domain/tenant/dm"
	"auth-perm/internal/domain/tenant/dto"
	"auth-perm/internal/domain/tenant/param"
	"auth-perm/internal/domain/tenant/repo"

	"auth-perm/internal/infra/code_gen"
)

// SessionInvalidator 会话失效接口（用于跨域解耦）
type SessionInvalidator interface {
	// InvalidateTenantSessions 使指定租户下的所有会话失效并清理缓存
	InvalidateTenantSessions(ctx context.Context, tenantID string) error
}

// TenantService 租户服务
type TenantService struct {
	tenantRepo         *repo.TenantRepo
	codeGen            code_gen.CodeGenerator
	sessionInvalidator SessionInvalidator
}

// NewTenantService 创建租户服务
func NewTenantService(tenantRepo *repo.TenantRepo, codeGen code_gen.CodeGenerator, sessionInvalidator SessionInvalidator) *TenantService {
	return &TenantService{tenantRepo: tenantRepo, codeGen: codeGen, sessionInvalidator: sessionInvalidator}
}

// Create 创建租户
func (s *TenantService) Create(ctx context.Context, params *param.CreateTenantParams) (*dto.TenantDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 自动生成租户代码
	code, err := code_gen.GenerateCodeWithDB(s.codeGen, "T", func() (string, error) {
		return s.tenantRepo.FindMaxCodeByPrefix(ctx, "T")
	})
	if err != nil {
		return nil, errors.WrapBizError(err, "生成租户代码失败")
	}

	// 检查租户代码是否已存在
	exists, err := s.tenantRepo.ExistsByCode(ctx, code)
	if err != nil {
		return nil, errors.WrapBizError(err, "检查租户代码失败")
	}
	if exists {
		return nil, errors.NewBusinessError("租户代码已存在")
	}

	// 创建租户
	tenant := &dm.TenantDO{
		Name:     params.Name,
		Code:     code,
		Plan:     params.Plan,
		ExpireAt: params.ExpireAt,
		Status:   constant.TenantStatusActive,
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, errors.WrapBizError(err, "创建租户失败")
	}

	return tenant.ToDTO(), nil
}

// Update 更新租户
func (s *TenantService) Update(ctx context.Context, params *param.UpdateTenantParams) (*dto.TenantDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 查询现有租户
	tenant, err := s.tenantRepo.FindByID(ctx, params.ID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询租户失败")
	}
	if tenant == nil {
		return nil, errors.NewNotFoundError("租户不存在")
	}

	// 更新字段
	if params.Name != "" {
		tenant.Name = params.Name
	}
	if params.Status != nil {
		tenant.Status = *params.Status
	}
	if params.Plan != nil {
		tenant.Plan = *params.Plan
	}
	if params.ExpireAt != nil {
		tenant.ExpireAt = params.ExpireAt
	}

	// 更新设置
	if params.Settings != nil {
		if err := s.tenantRepo.UpdateSettings(ctx, params.ID, params.Settings); err != nil {
			return nil, errors.WrapBizError(err, "更新租户设置失败")
		}
		tenant.Settings = *params.Settings
	}

	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, errors.WrapBizError(err, "更新租户失败")
	}

	return tenant.ToDTO(), nil
}

// Delete 删除租户
func (s *TenantService) Delete(ctx context.Context, params *param.DeleteTenantParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// 查询现有租户
	tenant, err := s.tenantRepo.FindByID(ctx, params.ID)
	if err != nil {
		return errors.WrapBizError(err, "查询租户失败")
	}
	if tenant == nil {
		return errors.NewNotFoundError("租户不存在")
	}

	// 检查是否有账户
	accountCount, err := s.tenantRepo.CountAccountsByTenantID(ctx, params.ID)
	if err != nil {
		return errors.WrapBizError(err, "检查账户数量失败")
	}
	if accountCount > 0 {
		return errors.NewBusinessError("该租户下存在账户，无法删除")
	}

	// 删除租户（更新状态为 deleted）
	if err := s.tenantRepo.UpdateStatus(ctx, params.ID, constant.TenantStatusDeleted); err != nil {
		return errors.WrapBizError(err, "删除租户失败")
	}

	// 使该租户下所有会话失效
	if err := s.sessionInvalidator.InvalidateTenantSessions(ctx, params.ID); err != nil {
		log.Printf("警告：使租户 %s 的会话失效失败: %v", params.ID, err)
	}

	return nil
}

// Enable 启用租户
func (s *TenantService) Enable(ctx context.Context, params *param.EnableTenantParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// 查询现有租户
	tenant, err := s.tenantRepo.FindByID(ctx, params.ID)
	if err != nil {
		return errors.WrapBizError(err, "查询租户失败")
	}
	if tenant == nil {
		return errors.NewNotFoundError("租户不存在")
	}

	// 启用租户（更新状态为 active）
	return s.tenantRepo.UpdateStatus(ctx, params.ID, constant.TenantStatusActive)
}

// Get 获取租户详情
func (s *TenantService) Get(ctx context.Context, params *param.GetTenantParams) (*dto.TenantDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	tenant, err := s.tenantRepo.FindByID(ctx, params.ID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询租户失败")
	}
	if tenant == nil {
		return nil, errors.NewNotFoundError("租户不存在")
	}

	return tenant.ToDTO(), nil
}

// GetByCode 根据代码获取租户
func (s *TenantService) GetByCode(ctx context.Context, code string) (*dto.TenantDTO, error) {
	if code == "" {
		return nil, errors.NewValidationError("租户代码不能为空")
	}

	tenant, err := s.tenantRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询租户失败")
	}
	if tenant == nil {
		return nil, errors.NewNotFoundError("租户不存在")
	}

	return tenant.ToDTO(), nil
}

// List 列出租户
func (s *TenantService) List(ctx context.Context, params *param.ListTenantsParams) ([]*dto.TenantListItemDTO, int64, error) {
	if err := params.Validate(); err != nil {
		return nil, 0, errors.NewValidationError(err.Error())
	}

	pagination := &model.Pagination{
		Page:     params.Page,
		PageSize: params.Size,
	}

	listParams := &repo.ListParams{
		Keyword:    params.Keyword,
		Status:     params.Status,
		Pagination: pagination,
	}

	tenants, total, err := s.tenantRepo.List(ctx, listParams)
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询租户列表失败")
	}

	// 转换为 DTO
	dtos := make([]*dto.TenantListItemDTO, len(tenants))
	for i, tenant := range tenants {
		dtos[i] = &dto.TenantListItemDTO{
			ID:        tenant.ID,
			Name:      tenant.Name,
			Code:      tenant.Code,
			Status:    tenant.Status,
			Plan:      tenant.Plan,
			ExpireAt:  tenant.ExpireAt,
			CreatedAt: tenant.CreatedAt,
		}
	}

	return dtos, total, nil
}

// UpdateSettings 更新租户设置
func (s *TenantService) UpdateSettings(ctx context.Context, params *param.UpdateTenantSettingsParams) (*dto.TenantDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 查询现有租户
	tenant, err := s.tenantRepo.FindByID(ctx, params.ID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询租户失败")
	}
	if tenant == nil {
		return nil, errors.NewNotFoundError("租户不存在")
	}

	// 更新设置
	if params.Settings != nil {
		if err := s.tenantRepo.UpdateSettings(ctx, params.ID, params.Settings); err != nil {
			return nil, errors.WrapBizError(err, "更新租户设置失败")
		}
		tenant.Settings = *params.Settings
	}

	return tenant.ToDTO(), nil
}

// ChangeStatus 变更租户状态
func (s *TenantService) ChangeStatus(ctx context.Context, tenantID string, status constant.TenantStatus) error {
	// 验证状态是否有效
	if !status.IsValid() {
		return errors.NewValidationError("无效的状态值，支持: active, suspended, deleted")
	}

	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return errors.WrapBizError(err, "查询租户失败")
	}
	if tenant == nil {
		return errors.NewNotFoundError("租户不存在")
	}

	tenant.Status = status
	return s.tenantRepo.Update(ctx, tenant)
}
