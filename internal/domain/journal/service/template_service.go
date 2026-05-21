package service

import (
	"context"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/journal/dm"
	"auth-perm/internal/domain/journal/dto"
	"auth-perm/internal/domain/journal/repo"
)

// TemplateService 模板服务
type TemplateService struct {
	templateRepo *repo.TemplateRepo
}

// NewTemplateService 创建模板服务
func NewTemplateService(templateRepo *repo.TemplateRepo) *TemplateService {
	return &TemplateService{templateRepo: templateRepo}
}

// Create 创建模板
func (s *TemplateService) Create(ctx context.Context, params *dto.CreateTemplateParams) (*dm.JournalTemplateDO, error) {
	template := dm.NewJournalTemplate(
		params.TenantID,
		params.AccountID,
		params.Name,
		params.Content,
		params.Tags,
	)

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// Update 更新模板
func (s *TemplateService) Update(ctx context.Context, params *dto.UpdateTemplateParams) (*dm.JournalTemplateDO, error) {
	template, err := s.templateRepo.FindByID(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	// 检查租户隔离
	if template.TenantID != params.TenantID {
		return nil, errors.NewBusinessError("无权限修改此模板")
	}

	// 检查权限：只能修改自己的模板
	if template.AccountID != params.AccountID {
		return nil, errors.NewBusinessError("无权限修改此模板")
	}

	// 更新字段
	if params.Name != nil {
		if *params.Name == "" {
			return nil, errors.NewBusinessError("模板名称不能为空")
		}
		if len(*params.Name) > 255 {
			return nil, errors.NewBusinessError("模板名称不能超过255个字符")
		}
		template.Name = *params.Name
	}
	if params.Content != nil {
		template.Content = params.Content
	}
	if params.Tags != nil {
		template.Tags = params.Tags
	}

	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// Delete 删除模板
func (s *TemplateService) Delete(ctx context.Context, templateID, tenantID, accountID string) error {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return err
	}

	// 检查租户隔离
	if template.TenantID != tenantID {
		return errors.NewBusinessError("无权限删除此模板")
	}

	// 检查权限：只能删除自己的模板
	if template.AccountID != accountID {
		return errors.NewBusinessError("无权限删除此模板")
	}

	return s.templateRepo.Delete(ctx, templateID)
}

// Get 获取模板详情
func (s *TemplateService) Get(ctx context.Context, templateID, tenantID string) (*dm.JournalTemplateDO, error) {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// 检查租户隔离
	if template.TenantID != tenantID {
		return nil, errors.NewBusinessError("无权限访问此模板")
	}

	return template, nil
}

// List 列表查询
func (s *TemplateService) List(ctx context.Context, params *dto.ListTemplateParams) ([]*dm.JournalTemplateDO, int64, error) {
	return s.templateRepo.ListByTenant(
		ctx,
		params.TenantID,
		params.Page,
		params.PageSize,
		params.Name,
		params.Tag,
	)
}

// ListByAccount 获取用户自己的模板
func (s *TemplateService) ListByAccount(ctx context.Context, tenantID, accountID string) ([]*dm.JournalTemplateDO, error) {
	return s.templateRepo.FindByAccountID(ctx, tenantID, accountID)
}