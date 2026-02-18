package service

import (
	"context"

	"auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	permissionDM "auth-perm/internal/domain/permission/dm"
	"auth-perm/internal/domain/permission/dto"
	"auth-perm/internal/domain/permission/param"
	"auth-perm/internal/domain/permission/repo"

	authService "auth-perm/internal/domain/auth/service"
	"auth-perm/internal/infra/code_gen"
)

// PermissionService 权限领域服务
type PermissionService struct {
	authService            *authService.AuthService
	permissionRepo         *repo.PermissionRepo
	permissionResourceRepo *repo.PermissionResourceRepo
	cache                  *authService.CacheService
	codeGen                code_gen.CodeGenerator
}

// NewPermissionService FUTURE: 权限服务创建 - 在实现权限管理时使用
func NewPermissionService(authService *authService.AuthService, permissionRepo *repo.PermissionRepo, cache *authService.CacheService, codeGen code_gen.CodeGenerator) *PermissionService {
	return &PermissionService{
		authService:    authService,
		permissionRepo: permissionRepo,
		cache:          cache,
		codeGen:        codeGen,
	}
}

// NewPermissionServiceWithResourceRepo 创建带权限资源仓储的服务
func NewPermissionServiceWithResourceRepo(authService *authService.AuthService, permissionRepo *repo.PermissionRepo, permissionResourceRepo *repo.PermissionResourceRepo, cache *authService.CacheService, codeGen code_gen.CodeGenerator) *PermissionService {
	return &PermissionService{
		authService:            authService,
		permissionRepo:         permissionRepo,
		permissionResourceRepo: permissionResourceRepo,
		cache:                  cache,
		codeGen:                codeGen,
	}
}

// getAccountPermissions 获取账户权限（带缓存）
func (s *PermissionService) getAccountPermissions(ctx context.Context, accountID string) ([]string, error) {
	// 优先从缓存获取
	if s.cache != nil {
		if cachedPerms, err := s.cache.GetPermissions(ctx, accountID); err == nil {
			return cachedPerms, nil
		}
	}

	// 缓存未命中，从数据库查询
	roles, err := s.permissionRepo.FindRolesByAccountID(ctx, accountID)
	if err != nil {
		return nil, errors.WrapBizError(err, "获取账户角色失败")
	}

	if len(roles) == 0 {
		return []string{}, nil
	}

	// 获取所有角色ID
	roleIDs := make([]string, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.ID
	}

	// 使用权限仓储查询权限
	permissions, err := s.permissionRepo.FindPermissionsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, errors.WrapBizError(err, "获取权限失败")
	}

	// 提取权限代码
	permissionCodes := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionCodes[i] = perm.Code
	}

	// 写入缓存（10分钟TTL）
	if s.cache != nil {
		s.cache.SetPermissions(ctx, accountID, permissionCodes, constant.CacheTTLPermission)
	}

	return permissionCodes, nil
}

// CheckPermission 检查账户是否拥有指定权限
func (s *PermissionService) CheckPermission(ctx context.Context, accountID, permissionCode string) (bool, error) {
	account, err := s.authService.FindAccountByID(ctx, accountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户失败")
	}

	// 检查账户状态
	if !account.IsActive() {
		return false, errors.NewBusinessError("账户未激活")
	}

	// 获取账户权限（带缓存）
	permissionCodes, err := s.getAccountPermissions(ctx, accountID)
	if err != nil {
		return false, err
	}

	for _, code := range permissionCodes {
		if code == permissionCode {
			return true, nil
		}
	}

	return false, nil
}

// CheckAnyPermission 检查账户是否拥有任意一个权限
func (s *PermissionService) CheckAnyPermission(ctx context.Context, accountID string, permissions []string) (bool, error) {
	account, err := s.authService.FindAccountByID(ctx, accountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户失败")
	}
	// 检查账户状态
	if !account.IsActive() {
		return false, errors.NewBusinessError("账户未激活")
	}

	// 获取账户权限（带缓存）
	permissionCodes, err := s.getAccountPermissions(ctx, accountID)
	if err != nil {
		return false, err
	}

	// 检查是否拥有任意一个权限
	for _, p := range permissions {
		for _, code := range permissionCodes {
			if code == p {
				return true, nil
			}
		}
	}

	return false, nil
}

// CheckAllPermissions 检查账户是否拥有所有权限
func (s *PermissionService) CheckAllPermissions(ctx context.Context, accountID string, permissions []string) (bool, error) {
	account, err := s.authService.FindAccountByID(ctx, accountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户失败")
	}
	// 检查账户状态
	if !account.IsActive() {
		return false, errors.NewBusinessError("账户未激活")
	}

	// 获取账户权限（带缓存）
	permissionCodes, err := s.getAccountPermissions(ctx, accountID)
	if err != nil {
		return false, err
	}

	// 检查是否拥有所有权限
	for _, p := range permissions {
		hasPermission := false
		for _, code := range permissionCodes {
			if code == p {
				hasPermission = true
				break
			}
		}
		if !hasPermission {
			return false, nil
		}
	}

	return true, nil
}

// CheckRole 检查账户是否拥有指定角色
func (s *PermissionService) CheckRole(ctx context.Context, accountID, roleCode string) (bool, error) {
	account, err := s.authService.FindAccountByID(ctx, accountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户失败")
	}
	// 检查账户状态
	if !account.IsActive() {
		return false, errors.NewBusinessError("账户未激活")
	}

	roles, err := s.permissionRepo.FindRolesByAccountID(ctx, accountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户角色失败")
	}

	for _, role := range roles {
		if role.Code == roleCode && role.IsActive {
			return true, nil
		}
	}

	return false, nil
}

// CheckAnyRole 检查账户是否拥有任意一个角色
func (s *PermissionService) CheckAnyRole(ctx context.Context, accountID string, roleCodes []string) (bool, error) {
	account, err := s.authService.FindAccountByID(ctx, accountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户失败")
	}
	// 检查账户状态
	if !account.IsActive() {
		return false, errors.NewBusinessError("账户未激活")
	}

	roles, err := s.permissionRepo.FindRolesByAccountID(ctx, accountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户角色失败")
	}

	roleCodeSet := make(map[string]struct{})
	for _, rc := range roleCodes {
		roleCodeSet[rc] = struct{}{}
	}

	for _, role := range roles {
		if _, exists := roleCodeSet[role.Code]; exists && role.IsActive {
			return true, nil
		}
	}

	return false, nil
}

// CheckAllRoles 检查账户是否拥有所有角色
func (s *PermissionService) CheckAllRoles(ctx context.Context, accountID string, roleCodes []string) (bool, error) {
	account, err := s.authService.FindAccountByID(ctx, accountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户失败")
	}
	// 检查账户状态
	if !account.IsActive() {
		return false, errors.NewBusinessError("账户未激活")
	}

	roles, err := s.permissionRepo.FindRolesByAccountID(ctx, accountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户角色失败")
	}

	var userRoles = make(map[string]struct{})
	for _, r := range roles {
		if r.IsActive {
			userRoles[r.Code] = struct{}{}
		}
	}

	for _, rc := range roleCodes {
		if _, exists := userRoles[rc]; !exists {
			return false, nil
		}
	}

	return true, nil
}

// GetAccountPermissions 获取账户的所有权限
func (s *PermissionService) GetAccountPermissions(ctx context.Context, accountID string) ([]string, error) {
	roles, err := s.permissionRepo.FindRolesByAccountID(ctx, accountID)
	if err != nil {
		return nil, errors.WrapBizError(err, "获取账户角色失败")
	}

	if len(roles) == 0 {
		return []string{}, nil
	}

	// 获取所有角色ID
	roleIDs := make([]string, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.ID
	}

	// 使用权限仓储查询权限
	permissions, err := s.permissionRepo.FindPermissionsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, errors.WrapBizError(err, "获取权限失败")
	}

	// 提取权限代码
	permissionSet := make(map[string]struct{})
	for _, perm := range permissions {
		if perm.IsActive {
			permissionSet[perm.Code] = struct{}{}
		}
	}

	result := make([]string, 0, len(permissionSet))
	for p := range permissionSet {
		result = append(result, p)
	}

	return result, nil
}

// GetAccountRoles 获取账户的所有角色
func (s *PermissionService) GetAccountRoles(ctx context.Context, accountID string) ([]*dto.RoleDTO, error) {
	var roles []*permissionDM.RoleDO
	roles, err := s.permissionRepo.FindRolesByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	roleDTOs := make([]*dto.RoleDTO, len(roles))
	for i, role := range roles {
		roleDTOs[i] = role.ToDTO()
	}
	return roleDTOs, nil
}

// CheckOrgPermission 检查账户在组织中的权限
func (s *PermissionService) CheckOrgPermission(ctx context.Context, params *param.CheckOrgPermissionParams) (bool, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return false, errors.NewValidationError(err.Error())
	}

	// This logic might need to be more complex, involving checking if an account is in an org first.
	// For now, we assume roles are org-specific.
	account, err := s.authService.FindAccountByID(ctx, params.AccountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户失败")
	}
	// 检查账户状态
	if !account.IsActive() {
		return false, errors.NewBusinessError("账户未激活")
	}

	roles, err := s.permissionRepo.FindRolesByAccountID(ctx, params.AccountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户角色失败")
	}

	if len(roles) == 0 {
		return false, nil
	}

	// 获取相关角色ID（全局角色或指定组织的角色）
	var relevantRoleIDs []string
	for _, role := range roles {
		if role.OrgID == params.OrgID || role.OrgID == "" {
			relevantRoleIDs = append(relevantRoleIDs, role.ID)
		}
	}

	if len(relevantRoleIDs) == 0 {
		return false, nil
	}

	// 使用权限仓储查询权限
	permissions, err := s.permissionRepo.FindPermissionsByRoleIDs(ctx, relevantRoleIDs)
	if err != nil {
		return false, errors.WrapBizError(err, "获取权限失败")
	}

	for _, perm := range permissions {
		if perm.Code == params.PermissionCode && perm.IsActive {
			return true, nil
		}
	}

	return false, nil
}

// IsSystemAdmin 检查是否为系统管理员
func (s *PermissionService) IsSystemAdmin(ctx context.Context, params *param.IsAdminParams) (bool, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return false, errors.NewValidationError(err.Error())
	}
	return s.CheckAnyRole(ctx, params.AccountID, []string{"super_admin", "admin"})
}

// IsOrgAdmin 检查是否为组织管理员
func (s *PermissionService) IsOrgAdmin(ctx context.Context, params *param.IsAdminParams) (bool, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return false, errors.NewValidationError(err.Error())
	}

	roles, err := s.permissionRepo.FindRolesByAccountID(ctx, params.AccountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户角色失败")
	}

	isOrgAdmin := false
	for _, role := range roles {
		if role.IsActive && role.Code == "org_admin" && role.OrgID == params.OrgID {
			isOrgAdmin = true
			break
		}
	}

	if isOrgAdmin {
		return true, nil
	}

	// A system admin is also an admin of any org
	return s.IsSystemAdmin(ctx, param.NewIsSystemAdminParams(params.AccountID))
}

// GetAccountPermissionsWithAuthCheck 获取账户权限（带权限检查）
func (s *PermissionService) GetAccountPermissionsWithAuthCheck(ctx context.Context, params *param.GetUserDataWithAuthCheckParams) ([]string, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 检查是否有查看其他账户权限的权限
	if !params.IsSelf() {
		canView, err := s.CheckPermission(ctx, params.CurrentAccountID, "users.read")
		if err != nil || !canView {
			return nil, errors.NewPermissionError("没有权限查看其他账户的权限")
		}
	}

	// 获取账户权限
	return s.GetAccountPermissions(ctx, params.TargetAccountID)
}

// GetAccountRolesWithAuthCheck 获取账户角色（带权限检查）
func (s *PermissionService) GetAccountRolesWithAuthCheck(ctx context.Context, params *param.GetUserDataWithAuthCheckParams) ([]*dto.RoleDTO, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 检查是否有查看其他账户角色的权限
	if !params.IsSelf() {
		canView, err := s.CheckPermission(ctx, params.CurrentAccountID, "users.read")
		if err != nil || !canView {
			return nil, errors.NewPermissionError("没有权限查看其他账户的角色")
		}
	}

	// 获取账户角色
	return s.GetAccountRoles(ctx, params.TargetAccountID)
}

// GetEffectivePermissionsWithAuthCheck 获取有效权限（带权限检查）
func (s *PermissionService) GetEffectivePermissionsWithAuthCheck(ctx context.Context, params *param.GetUserDataWithAuthCheckParams) ([]string, []*dto.RoleDTO, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, nil, errors.NewValidationError(err.Error())
	}

	// 检查是否有查看其他账户权限的权限
	if !params.IsSelf() {
		canView, err := s.CheckPermission(ctx, params.CurrentAccountID, "users.read")
		if err != nil || !canView {
			return nil, nil, errors.NewPermissionError("没有权限查看其他账户的权限")
		}
	}

	// 获取权限
	permissions, err := s.GetAccountPermissions(ctx, params.TargetAccountID)
	if err != nil {
		return nil, nil, errors.WrapBizError(err, "获取权限失败")
	}

	// 获取角色
	roles, err := s.GetAccountRoles(ctx, params.TargetAccountID)
	if err != nil {
		return nil, nil, errors.WrapBizError(err, "获取角色失败")
	}

	return permissions, roles, nil
}

// ==================== 权限资源关联管理 ====================

// CreatePermissionResource 创建权限资源关联
func (s *PermissionService) CreatePermissionResource(ctx context.Context, params *param.CreatePermissionResourceParams) (*dto.PermissionResourceDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 检查权限是否存在
	permission, err := s.permissionRepo.FindByID(ctx, params.PermissionID)
	if err != nil {
		return nil, errors.WrapBizError(err, "权限不存在")
	}
	if permission == nil {
		return nil, errors.NewBusinessError("权限不存在")
	}

	// 检查是否已存在
	existing, err := s.permissionResourceRepo.FindByPermissionAndResource(ctx, params.PermissionID, params.ResourceID, params.ResourceType)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询关联失败")
	}
	if existing != nil {
		return nil, errors.NewBusinessError("该资源关联已存在")
	}

	// 创建关联
	pr := permissionDM.NewPermissionResource(params.PermissionID, params.ResourceID, params.ResourceType, params.ResourceName, params.TenantID)
	if err := s.permissionResourceRepo.Create(ctx, pr); err != nil {
		return nil, errors.WrapBizError(err, "创建关联失败")
	}

	return pr.ToDTO(), nil
}

// CreatePermissionResourcesBatch 批量创建权限资源关联
func (s *PermissionService) CreatePermissionResourcesBatch(ctx context.Context, params *param.BindPermissionResourcesParams) ([]*dto.PermissionResourceDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 检查权限是否存在
	permission, err := s.permissionRepo.FindByID(ctx, params.PermissionID)
	if err != nil {
		return nil, errors.WrapBizError(err, "权限不存在")
	}
	if permission == nil {
		return nil, errors.NewBusinessError("权限不存在")
	}

	// 批量创建
	resources := make([]*permissionDM.PermissionResourceDO, 0, len(params.Resources))
	for _, r := range params.Resources {
		pr := permissionDM.NewPermissionResource(params.PermissionID, r.ResourceID, r.ResourceType, r.ResourceName, params.TenantID)
		resources = append(resources, pr)
	}

	if err := s.permissionResourceRepo.CreateBatch(ctx, resources); err != nil {
		return nil, errors.WrapBizError(err, "批量创建关联失败")
	}

	// 转换为DTO
	dtos := make([]*dto.PermissionResourceDTO, 0, len(resources))
	for _, r := range resources {
		dtos = append(dtos, r.ToDTO())
	}

	return dtos, nil
}

// UpdatePermissionResource 更新权限资源关联
func (s *PermissionService) UpdatePermissionResource(ctx context.Context, params *param.UpdatePermissionResourceParams) (*dto.PermissionResourceDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 查找现有关联
	resources, err := s.permissionResourceRepo.FindByPermissionID(ctx, "")
	if err != nil {
		return nil, errors.WrapBizError(err, "查询关联失败")
	}

	var existing *permissionDM.PermissionResourceDO
	for _, r := range resources {
		if r.ID == params.ID {
			existing = r
			break
		}
	}

	if existing == nil {
		return nil, errors.NewBusinessError("关联不存在")
	}

	// 更新
	existing.ResourceName = params.ResourceName
	if err := s.permissionResourceRepo.Update(ctx, existing); err != nil {
		return nil, errors.WrapBizError(err, "更新关联失败")
	}

	return existing.ToDTO(), nil
}

// DeletePermissionResource 删除权限资源关联
func (s *PermissionService) DeletePermissionResource(ctx context.Context, id string) error {
	if id == "" {
		return errors.NewValidationError("ID不能为空")
	}

	if err := s.permissionResourceRepo.DeleteByID(ctx, id); err != nil {
		return errors.WrapBizError(err, "删除关联失败")
	}

	return nil
}

// GetPermissionResources 获取权限的所有资源
func (s *PermissionService) GetPermissionResources(ctx context.Context, params *param.ListPermissionResourceParams) ([]*dto.PermissionResourceDTO, int64, error) {
	if err := params.Validate(); err != nil {
		return nil, 0, errors.NewValidationError(err.Error())
	}

	resources, err := s.permissionResourceRepo.FindByPermissionID(ctx, params.PermissionID)
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询关联失败")
	}

	// 按资源类型过滤
	if params.ResourceType != "" {
		filtered := make([]*permissionDM.PermissionResourceDO, 0)
		for _, r := range resources {
			if r.ResourceType == params.ResourceType {
				filtered = append(filtered, r)
			}
		}
		resources = filtered
	}

	// 转换为DTO
	dtos := make([]*dto.PermissionResourceDTO, 0, len(resources))
	for _, r := range resources {
		dtos = append(dtos, r.ToDTO())
	}

	return dtos, int64(len(dtos)), nil
}

// BindPermissionResources 绑定权限资源（批量）
func (s *PermissionService) BindPermissionResources(ctx context.Context, params *param.BindPermissionResourcesParams) ([]*dto.PermissionResourceDTO, error) {
	return s.CreatePermissionResourcesBatch(ctx, params)
}

// UnbindPermissionResources 解绑权限资源
func (s *PermissionService) UnbindPermissionResources(ctx context.Context, params *param.UnbindPermissionResourcesParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// 批量删除
	for _, resourceID := range params.ResourceIDs {
		if err := s.permissionResourceRepo.DeleteByResource(ctx, resourceID, ""); err != nil {
			return errors.WrapBizError(err, "解绑资源失败")
		}
	}

	return nil
}

// ==================== Permission 管理 ====================

// CreatePermission 创建权限
func (s *PermissionService) CreatePermission(ctx context.Context, params *param.CreatePermissionParams) (*dto.PermissionDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 自动生成权限代码
	code := params.Code
	if code == "" {
		var err error
		code, err = code_gen.GenerateCodeWithDB(s.codeGen, "P", func() (string, error) {
			return s.permissionRepo.FindMaxPermissionCodeByPrefix(ctx, params.TenantID, "P")
		}, params.TenantID)
		if err != nil {
			return nil, errors.WrapBizError(err, "生成权限代码失败")
		}
	}

	// 检查权限代码是否已存在
	existing, err := s.permissionRepo.FindPermissionByCode(ctx, params.TenantID, code)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询权限失败")
	}
	if existing != nil {
		return nil, errors.NewBusinessError("权限代码已存在")
	}

	// 创建权限
	permission := permissionDM.NewPermission(params.TenantID, code, params.Name)
	permission.Description = params.Description
	permission.IsSystem = params.IsSystem

	if err := s.permissionRepo.SavePermission(ctx, permission); err != nil {
		return nil, errors.WrapBizError(err, "创建权限失败")
	}

	return permission.ToDTO(), nil
}

// UpdatePermission 更新权限
func (s *PermissionService) UpdatePermission(ctx context.Context, params *param.UpdatePermissionParams) (*dto.PermissionDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 查询权限
	permission, err := s.permissionRepo.FindPermissionByID(ctx, params.ID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询权限失败")
	}
	if permission == nil {
		return nil, errors.NewNotFoundError("权限不存在")
	}

	// 租户隔离验证
	if permission.TenantID != params.TenantID {
		return nil, errors.NewPermissionError("无权操作此权限")
	}

	// 检查是否为系统权限
	if permission.IsSystem {
		return nil, errors.NewBusinessError("系统权限不能被修改")
	}

	// 更新字段
	if params.Name != "" {
		permission.Name = params.Name
	}
	if params.Description != "" {
		permission.Description = params.Description
	}
	if params.IsActive != nil {
		permission.IsActive = *params.IsActive
	}

	if err := s.permissionRepo.SavePermission(ctx, permission); err != nil {
		return nil, errors.WrapBizError(err, "更新权限失败")
	}

	return permission.ToDTO(), nil
}

// DeletePermission 删除权限
func (s *PermissionService) DeletePermission(ctx context.Context, params *param.DeletePermissionParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// 查询权限
	permission, err := s.permissionRepo.FindPermissionByID(ctx, params.ID)
	if err != nil {
		return errors.WrapBizError(err, "查询权限失败")
	}
	if permission == nil {
		return errors.NewNotFoundError("权限不存在")
	}

	// 租户隔离验证（如果params.TenantID为空，则使用权限的TenantID）
	tenantID := params.TenantID
	if tenantID == "" {
		tenantID = permission.TenantID
	}
	if permission.TenantID != tenantID {
		return errors.NewPermissionError("无权操作此权限")
	}

	// 检查是否为系统权限
	if permission.IsSystem {
		return errors.NewBusinessError("系统权限不能被删除")
	}

	// 检查是否有关联的角色
	rolePerms, err := s.permissionRepo.GetRolePermissions(ctx, permission.ID)
	if err != nil {
		return errors.WrapBizError(err, "检查权限关联失败")
	}
	if len(rolePerms) > 0 {
		return errors.NewBusinessError("该权限已关联角色，无法删除")
	}

	if err := s.permissionRepo.DeletePermission(ctx, params.ID); err != nil {
		return errors.WrapBizError(err, "删除权限失败")
	}

	// 清除缓存
	if s.cache != nil {
		s.cache.DeletePermissions(ctx, "")
	}

	return nil
}

// GetPermission 获取权限详情
func (s *PermissionService) GetPermission(ctx context.Context, id, tenantID string) (*dto.PermissionDTO, error) {
	if id == "" {
		return nil, errors.NewValidationError("权限ID不能为空")
	}

	permission, err := s.permissionRepo.FindPermissionByID(ctx, id)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询权限失败")
	}
	if permission == nil {
		return nil, errors.NewNotFoundError("权限不存在")
	}

	// 租户隔离验证（如果提供了 tenantID）
	if tenantID != "" && permission.TenantID != tenantID {
		return nil, errors.NewPermissionError("无权访问此权限")
	}

	return permission.ToDTO(), nil
}

// ListPermissions 获取权限列表
func (s *PermissionService) ListPermissions(ctx context.Context, params *param.ListPermissionParams) ([]*dto.PermissionDTO, int64, error) {
	if err := params.Validate(); err != nil {
		return nil, 0, errors.NewValidationError(err.Error())
	}

	permissions, err := s.permissionRepo.FindPermissionsByTenantID(ctx, params.TenantID, nil)
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询权限列表失败")
	}

	// 过滤
	var filtered []*permissionDM.PermissionDO
	for _, p := range permissions {
		if params.Code != "" && p.Code != params.Code {
			continue
		}
		if params.Name != "" && p.Name != params.Name {
			continue
		}
		if params.IsActive != nil && p.IsActive != *params.IsActive {
			continue
		}
		if params.IsSystem != nil && p.IsSystem != *params.IsSystem {
			continue
		}
		if !params.IncludeAll && !p.IsActive {
			continue
		}
		filtered = append(filtered, p)
	}

	// 分页
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize
	if start > len(filtered) {
		filtered = []*permissionDM.PermissionDO{}
	} else {
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[start:end]
	}

	// 转换为DTO
	dtos := make([]*dto.PermissionDTO, 0, len(filtered))
	for _, p := range filtered {
		dtos = append(dtos, p.ToDTO())
	}

	return dtos, int64(len(filtered)), nil
}

// ==================== Role 管理 ====================

// CreateRole 创建角色
func (s *PermissionService) CreateRole(ctx context.Context, params *param.CreateRoleParams) (*dto.RoleDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 自动生成角色代码
	code := params.Code
	if code == "" {
		var err error
		code, err = code_gen.GenerateCodeWithDB(s.codeGen, "R", func() (string, error) {
			return s.permissionRepo.FindMaxRoleCodeByPrefix(ctx, params.TenantID, "R")
		}, params.TenantID)
		if err != nil {
			return nil, errors.WrapBizError(err, "生成角色代码失败")
		}
	}

	// 检查角色代码是否已存在
	existing, err := s.permissionRepo.FindRoleByCode(ctx, params.TenantID, code)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询角色失败")
	}
	if existing != nil {
		return nil, errors.NewBusinessError("角色代码已存在")
	}

	// 创建角色
	role := permissionDM.NewRole(params.TenantID, code, params.Name)
	role.OrgID = params.OrgID
	role.Description = params.Description
	role.Priority = params.Priority
	role.IsSystem = params.IsSystem

	if err := s.permissionRepo.SaveRole(ctx, role); err != nil {
		return nil, errors.WrapBizError(err, "创建角色失败")
	}

	return role.ToDTO(), nil
}

// UpdateRole 更新角色
func (s *PermissionService) UpdateRole(ctx context.Context, params *param.UpdateRoleParams) (*dto.RoleDTO, error) {
	if err := params.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// 查询角色
	role, err := s.permissionRepo.FindRoleByID(ctx, params.ID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询角色失败")
	}
	if role == nil {
		return nil, errors.NewNotFoundError("角色不存在")
	}

	// 租户隔离验证
	if role.TenantID != params.TenantID {
		return nil, errors.NewPermissionError("无权操作此角色")
	}

	// 检查是否为系统角色
	if role.IsSystem {
		return nil, errors.NewBusinessError("系统角色不能被修改")
	}

	// 查询关联的账户ID列表，用于失效缓存
	accountIDs, err := s.permissionRepo.FindAccountIDsByRoleIDs(ctx, []string{params.ID})
	if err != nil {
		accountIDs = []string{}
	}

	// 更新字段
	if params.Name != "" {
		role.Name = params.Name
	}
	if params.Description != "" {
		role.Description = params.Description
	}
	if params.Priority != 0 {
		role.Priority = params.Priority
	}
	if params.IsActive != nil {
		role.IsActive = *params.IsActive
	}

	if err := s.permissionRepo.SaveRole(ctx, role); err != nil {
		return nil, errors.WrapBizError(err, "更新角色失败")
	}

	// 失效相关账户的权限缓存
	if len(accountIDs) > 0 && s.cache != nil {
		_ = s.cache.DeletePermissionsByAccountIDs(ctx, accountIDs)
	}

	return role.ToDTO(), nil
}

// DeleteRole 删除角色
func (s *PermissionService) DeleteRole(ctx context.Context, params *param.DeleteRoleParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// 查询角色
	role, err := s.permissionRepo.FindRoleByID(ctx, params.ID)
	if err != nil {
		return errors.WrapBizError(err, "查询角色失败")
	}
	if role == nil {
		return errors.NewNotFoundError("角色不存在")
	}

	// 租户隔离验证（如果params.TenantID为空，则使用角色的TenantID）
	tenantID := params.TenantID
	if tenantID == "" {
		tenantID = role.TenantID
	}
	if role.TenantID != tenantID {
		return errors.NewPermissionError("无权操作此角色")
	}

	// 检查是否为系统角色
	if role.IsSystem {
		return errors.NewBusinessError("系统角色不能被删除")
	}

	// 查询关联的账户ID列表，用于失效缓存
	accountIDs, err := s.permissionRepo.FindAccountIDsByRoleIDs(ctx, []string{params.ID})
	if err != nil {
		// 记录错误但继续执行，不阻塞删除操作
		accountIDs = []string{}
	}

	if err := s.permissionRepo.DeleteRole(ctx, params.ID); err != nil {
		return errors.WrapBizError(err, "删除角色失败")
	}

	// 失效相关账户的权限缓存
	if len(accountIDs) > 0 && s.cache != nil {
		_ = s.cache.DeletePermissionsByAccountIDs(ctx, accountIDs)
	}

	return nil
}

// GetRole 获取角色详情
func (s *PermissionService) GetRole(ctx context.Context, id, tenantID string) (*dto.RoleDTO, error) {
	if id == "" {
		return nil, errors.NewValidationError("角色ID不能为空")
	}

	role, err := s.permissionRepo.FindRoleByID(ctx, id)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询角色失败")
	}
	if role == nil {
		return nil, errors.NewNotFoundError("角色不存在")
	}

	// 租户隔离验证：只有当两者都有值且不匹配时才拒绝
	if tenantID != "" && role.TenantID != "" && role.TenantID != tenantID {
		return nil, errors.NewPermissionError("无权访问此角色")
	}

	roleDTO := role.ToDTO()

	// 确定查询权限的租户ID：优先使用角色的租户ID，如果角色是全局的则使用传入的tenantID
	queryTenantID := role.TenantID
	if queryTenantID == "" {
		queryTenantID = tenantID
	}

	// 查询租户下全量权限
	allPermissions, err := s.permissionRepo.FindPermissionsByTenantID(ctx, queryTenantID, nil)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询权限列表失败")
	}

	// 获取角色已关联的权限ID
	rolePermIDs, err := s.permissionRepo.GetRolePermissions(ctx, id)
	if err != nil {
		rolePermIDs = []string{}
	}

	// 构建已关联权限ID的 map
	rolePermMap := make(map[string]bool)
	for _, pid := range rolePermIDs {
		rolePermMap[pid] = true
	}

	// 转换权限并标记关联状态
	permissions := make([]*dto.PermissionDTO, 0, len(allPermissions))
	for _, perm := range allPermissions {
		permDTO := perm.ToDTO()
		if rolePermMap[perm.ID] {
			permDTO.IsSelected = true
		}
		permissions = append(permissions, permDTO)
	}

	roleDTO.Permissions = permissions
	roleDTO.PermissionCount = len(permissions)

	return roleDTO, nil
}

// ListRoles 获取角色列表
func (s *PermissionService) ListRoles(ctx context.Context, params *param.ListRoleParams) ([]*dto.RoleDTO, int64, error) {
	if err := params.Validate(); err != nil {
		return nil, 0, errors.NewValidationError(err.Error())
	}

	roles, err := s.permissionRepo.FindRolesByTenantID(ctx, params.TenantID, nil)
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询角色列表失败")
	}

	// 过滤
	var filtered []*permissionDM.RoleDO
	for _, r := range roles {
		if params.OrgID != "" && r.OrgID != params.OrgID {
			continue
		}
		if params.Code != "" && r.Code != params.Code {
			continue
		}
		if params.Name != "" && r.Name != params.Name {
			continue
		}
		if params.IsActive != nil && r.IsActive != *params.IsActive {
			continue
		}
		if params.IsSystem != nil && r.IsSystem != *params.IsSystem {
			continue
		}
		if !params.IncludeAll && !r.IsActive {
			continue
		}
		filtered = append(filtered, r)
	}

	// 分页
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize
	if start > len(filtered) {
		filtered = []*permissionDM.RoleDO{}
	} else {
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[start:end]
	}

	// 转换为DTO
	dtos := make([]*dto.RoleDTO, 0, len(filtered))
	for _, r := range filtered {
		roleDTO := r.ToDTO()
		// 获取权限数量
		permIDs, err := s.permissionRepo.GetRolePermissions(ctx, r.ID)
		if err == nil && len(permIDs) > 0 {
			roleDTO.PermissionCount = len(permIDs)
		}
		dtos = append(dtos, roleDTO)
	}

	return dtos, int64(len(filtered)), nil
}

// ==================== Role-Permission 关联管理 ====================

// AssignPermissionToRole 分配权限给角色
func (s *PermissionService) AssignPermissionToRole(ctx context.Context, params *param.AssignPermissionToRoleParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// 查询角色
	role, err := s.permissionRepo.FindRoleByID(ctx, params.RoleID)
	if err != nil {
		return errors.WrapBizError(err, "查询角色失败")
	}
	if role == nil {
		return errors.NewNotFoundError("角色不存在")
	}

	// 查询关联的账户ID列表，用于失效缓存
	accountIDs, err := s.permissionRepo.FindAccountIDsByRoleIDs(ctx, []string{params.RoleID})
	if err != nil {
		accountIDs = []string{}
	}

	// 分配权限
	for _, permissionID := range params.PermissionIDs {
		if err := s.permissionRepo.AssignPermissionToRole(ctx, params.RoleID, permissionID, params.TenantID); err != nil {
			return errors.WrapBizError(err, "分配权限失败")
		}
	}

	// 失效相关账户的权限缓存
	if len(accountIDs) > 0 && s.cache != nil {
		_ = s.cache.DeletePermissionsByAccountIDs(ctx, accountIDs)
	}

	return nil
}

// RemovePermissionFromRole 移除角色权限
func (s *PermissionService) RemovePermissionFromRole(ctx context.Context, params *param.RemovePermissionFromRoleParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// 查询关联的账户ID列表，用于失效缓存
	accountIDs, err := s.permissionRepo.FindAccountIDsByRoleIDs(ctx, []string{params.RoleID})
	if err != nil {
		accountIDs = []string{}
	}

	if err := s.permissionRepo.RemovePermissionFromRole(ctx, params.RoleID, params.PermissionID); err != nil {
		return errors.WrapBizError(err, "移除权限失败")
	}

	// 失效相关账户的权限缓存
	if len(accountIDs) > 0 && s.cache != nil {
		_ = s.cache.DeletePermissionsByAccountIDs(ctx, accountIDs)
	}

	return nil
}

// GetRolePermissions 获取角色的权限列表
func (s *PermissionService) GetRolePermissions(ctx context.Context, roleID, tenantID string) ([]*dto.PermissionDTO, error) {
	if roleID == "" {
		return nil, errors.NewValidationError("角色ID不能为空")
	}

	// 验证角色属于该租户
	role, err := s.permissionRepo.FindRoleByID(ctx, roleID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询角色失败")
	}
	if role == nil {
		return nil, errors.NewNotFoundError("角色不存在")
	}
	// 租户隔离验证：只有当两者都有值且不匹配时才拒绝
	if tenantID != "" && role.TenantID != "" && role.TenantID != tenantID {
		return nil, errors.NewPermissionError("无权操作此角色")
	}

	permissionIDs, err := s.permissionRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询角色权限失败")
	}

	if len(permissionIDs) == 0 {
		return []*dto.PermissionDTO{}, nil
	}

	// 查询权限详情
	permissions, err := s.permissionRepo.FindByIDs(ctx, permissionIDs)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询权限失败")
	}

	dtos := make([]*dto.PermissionDTO, 0, len(permissions))
	for _, p := range permissions {
		dtos = append(dtos, p.ToDTO())
	}

	return dtos, nil
}

// ==================== Account-Role 关联管理 ====================

// AssignRoleToAccount 分配角色给账户
func (s *PermissionService) AssignRoleToAccount(ctx context.Context, params *param.AssignRoleToAccountParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// 分配角色
	for _, roleID := range params.RoleIDs {
		if err := s.permissionRepo.AssignRoleToAccount(ctx, params.AccountID, roleID, params.TenantID); err != nil {
			return errors.WrapBizError(err, "分配角色失败")
		}
	}

	// 清除缓存
	if s.cache != nil {
		s.cache.DeletePermissions(ctx, params.AccountID)
	}

	return nil
}

// RemoveRoleFromAccount 移除账户角色
func (s *PermissionService) RemoveRoleFromAccount(ctx context.Context, params *param.RemoveRoleFromAccountParams) error {
	if err := params.Validate(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	if err := s.permissionRepo.RemoveRoleFromAccount(ctx, params.AccountID, params.RoleID); err != nil {
		return errors.WrapBizError(err, "移除角色失败")
	}

	// 清除缓存
	if s.cache != nil {
		s.cache.DeletePermissions(ctx, params.AccountID)
	}

	return nil
}

// GetAccountRolesByAccountID 获取账户的角色列表
func (s *PermissionService) GetAccountRolesByAccountID(ctx context.Context, accountID string) ([]*dto.RoleDTO, error) {
	if accountID == "" {
		return nil, errors.NewValidationError("账户ID不能为空")
	}

	roles, err := s.permissionRepo.FindRolesByAccountID(ctx, accountID)
	if err != nil {
		return nil, errors.WrapBizError(err, "查询账户角色失败")
	}

	dtos := make([]*dto.RoleDTO, 0, len(roles))
	for _, r := range roles {
		dtos = append(dtos, r.ToDTO())
	}

	return dtos, nil
}
