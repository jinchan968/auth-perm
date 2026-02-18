package service

import (
	"context"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/permission/dm"
	"auth-perm/internal/domain/permission/dto"
	"auth-perm/internal/domain/permission/param"
	"auth-perm/internal/domain/permission/repo"
)

// OrganizationService 组织服务
type OrganizationService struct {
	orgRepo *repo.OrganizationRepo
}

// NewOrganizationService 创建组织服务
func NewOrganizationService(orgRepo *repo.OrganizationRepo) *OrganizationService {
	return &OrganizationService{orgRepo: orgRepo}
}

// Create 创建组织
func (s *OrganizationService) Create(ctx context.Context, params *param.CreateOrganizationParams) (*dto.OrganizationDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 检查编码是否已存在
	existing, err := s.orgRepo.FindByCode(ctx, params.TenantID, params.Code)
	if err != nil {
		return nil, errors.WrapBizError(err, "检查组织编码失败")
	}
	if existing != nil {
		return nil, errors.NewBusinessError("组织编码已存在")
	}

	// 处理父组织
	var parentID *string
	if params.ParentID != "" {
		parentID = &params.ParentID
		// 验证父组织存在
		parent, err := s.orgRepo.FindByID(ctx, params.ParentID)
		if err != nil {
			return nil, errors.WrapBizError(err, "检查父组织失败")
		}
		if parent == nil {
			return nil, errors.NewNotFoundError("父组织不存在")
		}
	}

	// 创建组织
	org := &dm.OrganizationDO{
		TenantID:    params.TenantID,
		ParentID:    parentID,
		Code:        params.Code,
		Name:        params.Name,
		Description: params.Description,
		SortOrder:   params.SortOrder,
		IsActive:    true,
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, errors.WrapBizError(err, "创建组织失败")
	}

	return s.enrichOrgDTO(ctx, org.ToDTO())
}

// Update 更新组织
func (s *OrganizationService) Update(ctx context.Context, params *param.UpdateOrganizationParams) (*dto.OrganizationDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 查询现有组织
	org, err := s.orgRepo.FindByID(ctx, params.ID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询组织失败")
	}
	if org == nil {
		return nil, errors.NewNotFoundError("组织不存在")
	}

	// 更新字段
	if params.Name != "" {
		org.Name = params.Name
	}
	if params.Description != "" {
		org.Description = params.Description
	}
	if params.ParentID != "" && params.ParentID != *org.ParentID {
		// 验证父组织存在
		if params.ParentID == params.ID {
			return nil, errors.NewBusinessError("不能将组织设置为自己的子组织")
		}
		org.ParentID = &params.ParentID
	}
	if params.IsActive != nil {
		org.IsActive = *params.IsActive
	}
	if params.SortOrder != 0 {
		org.SortOrder = params.SortOrder
	}

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, errors.WrapBizError(err, "更新组织失败")
	}

	return s.enrichOrgDTO(ctx, org.ToDTO())
}

// Delete 删除组织
func (s *OrganizationService) Delete(ctx context.Context, params *param.DeleteOrganizationParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// 查询现有组织
	org, err := s.orgRepo.FindByID(ctx, params.ID)
	if err != nil {
		return errors.WrapBizError(err, "查询组织失败")
	}
	if org == nil {
		return errors.NewNotFoundError("组织不存在")
	}

	// 检查是否有子组织
	hasChildren, err := s.orgRepo.HasChildren(ctx, params.ID)
	if err != nil {
		return errors.WrapBizError(err, "检查子组织失败")
	}
	if hasChildren {
		return errors.NewBusinessError("该组织下存在子组织，无法删除")
	}

	// 检查是否有用户
	userCount, err := s.orgRepo.CountUsersByOrgID(ctx, params.ID)
	if err != nil {
		return errors.WrapBizError(err, "检查用户数量失败")
	}
	if userCount > 0 {
		return errors.NewBusinessError("该组织下存在用户，无法删除")
	}

	return s.orgRepo.Delete(ctx, params.ID)
}

// Get 获取组织详情
func (s *OrganizationService) Get(ctx context.Context, params *param.GetOrganizationParams) (*dto.OrganizationDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	org, err := s.orgRepo.FindByID(ctx, params.ID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询组织失败")
	}
	if org == nil {
		return nil, errors.NewNotFoundError("组织不存在")
	}

	return s.enrichOrgDTO(ctx, org.ToDTO())
}

// List 列出组织
func (s *OrganizationService) List(ctx context.Context, params *param.ListOrganizationsParams) ([]*dto.OrganizationDTO, int64, error) {
	if err := params.Validate(); err != nil {
		return nil, 0, errors.NewValidationError(err.Error())
	}

	pagination := &model.Pagination{
		Page:     params.Page,
		PageSize: params.Size,
	}

	var orgs []*dm.OrganizationDO
	var total int64
	var err error

	if params.ParentID != "" {
		orgs, err = s.orgRepo.FindByParentID(ctx, params.ParentID)
		if err != nil {
			return nil, 0, errors.WrapBizError(err, "查询子组织失败")
		}
		total = int64(len(orgs))
	} else {
		orgs, total, err = s.orgRepo.FindByTenantID(ctx, params.TenantID, pagination)
		if err != nil {
			return nil, 0, errors.WrapBizError(err, "查询组织列表失败")
		}
	}

	// 转换为 DTO
	dtos := make([]*dto.OrganizationDTO, len(orgs))
	for i, org := range orgs {
		dtos[i] = org.ToDTO()
	}

	// 填充统计信息
	for i := range dtos {
		dtos[i] = s.enrichWithStats(ctx, dtos[i])
	}

	return dtos, total, nil
}

// GetTree 获取组织树
func (s *OrganizationService) GetTree(ctx context.Context, tenantID string) ([]*dto.OrganizationTreeNodeDTO, error) {
	// 获取所有顶级组织
	rootOrgs, err := s.orgRepo.FindRootOrgs(ctx, tenantID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询顶级组织失败")
	}

	// 构建树
	trees := make([]*dto.OrganizationTreeNodeDTO, len(rootOrgs))
	for i, org := range rootOrgs {
		trees[i] = s.buildTreeNode(ctx, org)
	}

	return trees, nil
}

// AssignAccountToOrg 分配账户到组织
func (s *OrganizationService) AssignAccountToOrg(ctx context.Context, params *param.AssignAccountToOrgParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// 检查组织是否存在
	org, err := s.orgRepo.FindByID(ctx, params.OrgID)
	if err != nil {
		return errors.WrapBizError(err, "检查组织失败")
	}
	if org == nil {
		return errors.NewNotFoundError("组织不存在")
	}

	// 检查是否已存在关联
	exists, err := s.orgRepo.ExistsAccountOrg(ctx, params.AccountID, params.OrgID)
	if err != nil {
		return errors.WrapBizError(err, "检查关联失败")
	}
	if exists {
		return errors.NewBusinessError("账户已在该组织中")
	}

	return s.orgRepo.AssignAccountToOrg(ctx, params.AccountID, params.OrgID, params.TenantID)
}

// RemoveAccountFromOrg 从组织移除账户
func (s *OrganizationService) RemoveAccountFromOrg(ctx context.Context, params *param.RemoveAccountFromOrgParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	return s.orgRepo.RemoveAccountFromOrg(ctx, params.AccountID, params.OrgID)
}

// GetUserOrganizations 获取用户所属组织列表
func (s *OrganizationService) GetUserOrganizations(ctx context.Context, accountID string) ([]*dto.OrganizationDTO, error) {
	if accountID == "" {
		return nil, errors.NewValidationError("账户ID不能为空")
	}

	orgs, err := s.orgRepo.FindAccountOrgs(ctx, accountID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询用户组织失败")
	}

	dtos := make([]*dto.OrganizationDTO, len(orgs))
	for i, org := range orgs {
		dtos[i] = org.ToDTO()
	}

	return dtos, nil
}

// ==================== 私有方法 ====================

// enrichOrgDTO 填充组织 DTO 的额外信息
func (s *OrganizationService) enrichOrgDTO(ctx context.Context, org *dto.OrganizationDTO) (*dto.OrganizationDTO, error) {
	if org == nil {
		return nil, nil
	}

	// 判断是否为根组织
	org.IsRoot_ = org.ParentID == nil || *org.ParentID == ""

	return s.enrichWithStats(ctx, org), nil
}

// enrichWithStats 填充统计信息
func (s *OrganizationService) enrichWithStats(ctx context.Context, org *dto.OrganizationDTO) *dto.OrganizationDTO {
	if org == nil {
		return nil
	}

	// 填充用户数量
	userCount, _ := s.orgRepo.CountUsersByOrgID(ctx, org.ID)
	org.UserCount = int(userCount)

	// 填充子组织数量
	children, _ := s.orgRepo.FindByParentID(ctx, org.ID)
	org.ChildCount = len(children)
	org.IsLeaf_ = org.ChildCount == 0

	return org
}

// buildTreeNode 构建树节点
func (s *OrganizationService) buildTreeNode(ctx context.Context, org *dm.OrganizationDO) *dto.OrganizationTreeNodeDTO {
	node := &dto.OrganizationTreeNodeDTO{
		ID:          org.ID,
		TenantID:    org.TenantID,
		ParentID:    org.ParentID,
		Name:        org.Name,
		Code:        org.Code,
		Description: org.Description,
		Level:       org.Level,
		Path:        org.Path,
		IsActive:    org.IsActive,
		SortOrder:   org.SortOrder,
		Metadata:    org.Metadata,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
	}

	// 判断是否为根组织
	node.IsRoot = org.ParentID == nil || *org.ParentID == ""

	// 填充统计信息
	userCount, _ := s.orgRepo.CountUsersByOrgID(ctx, org.ID)
	node.UserCount = int(userCount)

	// 递归填充子节点
	children, _ := s.orgRepo.FindByParentID(ctx, org.ID)
	if len(children) > 0 {
		node.Children = make([]dto.OrganizationTreeNodeDTO, len(children))
		for i, child := range children {
			node.Children[i] = *s.buildTreeNode(ctx, child)
		}
		node.IsLeaf = false
	} else {
		node.IsLeaf = true
	}

	return node
}

// CountUsersByOrgID 统计组织用户数量
func (s *OrganizationService) CountUsersByOrgID(ctx context.Context, orgID string) (int, error) {
	if orgID == "" {
		return 0, errors.NewValidationError("组织ID不能为空")
	}

	count, err := s.orgRepo.CountUsersByOrgID(ctx, orgID)
	if err != nil {
		return 0, errors.WrapBizError(err, "查询组织用户数量失败")
	}

	return int(count), nil
}

// CountActiveUsersByOrgID 统计组织活跃用户数量
func (s *OrganizationService) CountActiveUsersByOrgID(ctx context.Context, orgID string, days int) (int, error) {
	if orgID == "" {
		return 0, errors.NewValidationError("组织ID不能为空")
	}

	count, err := s.orgRepo.CountActiveUsersByOrgID(ctx, orgID, days)
	if err != nil {
		return 0, errors.WrapBizError(err, "查询组织活跃用户数量失败")
	}

	return int(count), nil
}

// CheckUserInOrg 检查用户是否属于组织
func (s *OrganizationService) CheckUserInOrg(ctx context.Context, accountID, orgID string) (bool, error) {
	if accountID == "" {
		return false, errors.NewValidationError("账户ID不能为空")
	}
	if orgID == "" {
		return false, errors.NewValidationError("组织ID不能为空")
	}

	exists, err := s.orgRepo.ExistsAccountOrg(ctx, accountID, orgID)
	if err != nil {
		return false, errors.WrapBizError(err, "查询用户组织归属失败")
	}

	return exists, nil
}
