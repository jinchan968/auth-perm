package service

import (
	"context"
	"log"

	"auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	permissionConstant "auth-perm/internal/domain/permission/constant"
	"auth-perm/internal/domain/permission/dto"
	"auth-perm/internal/domain/permission/param"
)

// CheckAPIPermission 检查API访问权限（支持资源关联）
// 内部委托给 CheckResourcePermission 统一方法
func (s *PermissionService) CheckAPIPermission(ctx context.Context, accountID, apiPath string) (bool, error) {
	params := param.NewCheckResourcePermissionParams(accountID, apiPath, permissionConstant.ResourceTypeAPIPath)
	return s.CheckResourcePermission(ctx, params)
}

// CheckMenuPermission 检查菜单访问权限（支持资源关联）
// 内部委托给 CheckResourcePermission 统一方法
func (s *PermissionService) CheckMenuPermission(ctx context.Context, accountID, menuID string) (bool, error) {
	params := param.NewCheckResourcePermissionParams(accountID, menuID, permissionConstant.ResourceTypeMenu)
	return s.CheckResourcePermission(ctx, params)
}

// CheckButtonPermission 检查按钮访问权限（支持资源关联）
// 内部委托给 CheckResourcePermission 统一方法
func (s *PermissionService) CheckButtonPermission(ctx context.Context, accountID, buttonID string) (bool, error) {
	params := param.NewCheckResourcePermissionParams(accountID, buttonID, permissionConstant.ResourceTypeButton)
	return s.CheckResourcePermission(ctx, params)
}

// CheckFieldPermission 检查字段访问权限（支持资源关联）
// 内部委托给 CheckResourcePermission 统一方法
func (s *PermissionService) CheckFieldPermission(ctx context.Context, accountID, fieldID string) (bool, error) {
	params := param.NewCheckResourcePermissionParams(accountID, fieldID, permissionConstant.ResourceTypeField)
	return s.CheckResourcePermission(ctx, params)
}

// GetAccountResources 获取账户可访问的资源（支持资源关联，带缓存）
func (s *PermissionService) GetAccountResources(ctx context.Context, params *param.GetAccountResourcesParams) ([]string, error) {
	account, err := s.authService.FindAccountByID(ctx, params.AccountID)
	if err != nil {
		return nil, errors.WrapBizError(err, "获取账户失败")
	}

	// 检查账户状态
	if !account.IsActive() {
		return nil, errors.NewBusinessError("账户未激活")
	}

	// 优先从缓存获取资源列表
	if params.ResourceType != "" {
		if cachedResources, err := s.cache.GetAccountResources(ctx, params.AccountID, params.ResourceType); err == nil {
			return cachedResources, nil
		}
	}

	// 获取账户的所有权限
	permissionCodes, err := s.getAccountPermissions(ctx, params.AccountID)
	if err != nil {
		return nil, err
	}

	if len(permissionCodes) == 0 {
		// 缓存空结果
		if params.ResourceType != "" {
			if err := s.cache.SetAccountResources(ctx, params.AccountID, params.ResourceType, []string{}, constant.CacheTTLPermission); err != nil {
				log.Printf("WARN: Failed to set resources cache for account %s: %v", params.AccountID, err)
			}
		}
		return []string{}, nil
	}

	// 获取权限ID列表
	permissions, err := s.permissionRepo.FindByCodes(ctx, permissionCodes)
	if err != nil {
		return nil, errors.WrapBizError(err, "获取权限失败")
	}

	if len(permissions) == 0 {
		if params.ResourceType != "" {
			if err := s.cache.SetAccountResources(ctx, params.AccountID, params.ResourceType, []string{}, constant.CacheTTLPermission); err != nil {
				log.Printf("WARN: Failed to set resources cache for account %s: %v", params.AccountID, err)
			}
		}
		return []string{}, nil
	}

	permissionIDs := make([]string, 0, len(permissions))
	for _, p := range permissions {
		permissionIDs = append(permissionIDs, p.ID)
	}

	// 获取权限资源映射
	resourceMap, err := s.permissionResourceRepo.GetPermissionResourceMap(ctx, permissionIDs)
	if err != nil {
		return nil, errors.WrapBizError(err, "获取权限资源失败")
	}

	// 收集资源ID
	resourceSet := make(map[string]bool)
	for _, permID := range permissionIDs {
		if resources, ok := resourceMap[permID]; ok {
			for _, r := range resources {
				// 按资源类型过滤
				if params.ResourceType != "" && r.ResourceType != params.ResourceType {
					continue
				}
				resourceSet[r.ResourceID] = true
			}
		}
	}

	// 转换为切片
	result := make([]string, 0, len(resourceSet))
	for id := range resourceSet {
		result = append(result, id)
	}

	// 写入缓存
	if params.ResourceType != "" {
		if err := s.cache.SetAccountResources(ctx, params.AccountID, params.ResourceType, result, constant.CacheTTLPermission); err != nil {
			log.Printf("WARN: Failed to set resources cache for account %s: %v", params.AccountID, err)
		}
	}

	return result, nil
}

// GetAccountResourcesDetailed 获取账户可访问的资源（详细信息，含 resource_id、resource_type、resource_name）
// 用于前端权限控制，返回完整资源列表供菜单/按钮显隐判断
func (s *PermissionService) GetAccountResourcesDetailed(ctx context.Context, accountID string) ([]*dto.PermissionResourceDTO, error) {
	account, err := s.authService.FindAccountByID(ctx, accountID)
	if err != nil {
		return nil, errors.WrapBizError(err, "获取账户失败")
	}

	// 检查账户状态
	if !account.IsActive() {
		return nil, errors.NewBusinessError("账户未激活")
	}

	// 获取账户的所有权限
	permissionCodes, err := s.getAccountPermissions(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if len(permissionCodes) == 0 {
		return []*dto.PermissionResourceDTO{}, nil
	}

	// 获取权限ID列表
	permissions, err := s.permissionRepo.FindByCodes(ctx, permissionCodes)
	if err != nil {
		return nil, errors.WrapBizError(err, "获取权限失败")
	}

	if len(permissions) == 0 {
		return []*dto.PermissionResourceDTO{}, nil
	}

	permissionIDs := make([]string, 0, len(permissions))
	for _, p := range permissions {
		permissionIDs = append(permissionIDs, p.ID)
	}

	// 获取权限资源映射
	resourceMap, err := s.permissionResourceRepo.GetPermissionResourceMap(ctx, permissionIDs)
	if err != nil {
		return nil, errors.WrapBizError(err, "获取权限资源失败")
	}

	// 收集所有资源，去重（按 resource_id + resource_type 组合）
	resourceSet := make(map[string]*dto.PermissionResourceDTO)
	for _, permID := range permissionIDs {
		if resources, ok := resourceMap[permID]; ok {
			for _, r := range resources {
				key := r.ResourceType + ":" + r.ResourceID
				if _, exists := resourceSet[key]; !exists {
					resourceSet[key] = r.ToDTO()
				}
			}
		}
	}

	// 转换为切片
	result := make([]*dto.PermissionResourceDTO, 0, len(resourceSet))
	for _, res := range resourceSet {
		result = append(result, res)
	}

	return result, nil
}

// GetAllResourcesForSuperAdmin 获取所有权限资源（超管专用）
func (s *PermissionService) GetAllResourcesForSuperAdmin(ctx context.Context) ([]*dto.PermissionResourceDTO, error) {
	resources, err := s.permissionResourceRepo.FindAllResources(ctx)
	if err != nil {
		return nil, errors.WrapBizError(err, "获取全部资源失败")
	}

	// 去重（按 resource_id + resource_type 组合）
	resourceSet := make(map[string]*dto.PermissionResourceDTO)
	for _, r := range resources {
		key := r.ResourceType + ":" + r.ResourceID
		if _, exists := resourceSet[key]; !exists {
			resourceSet[key] = r.ToDTO()
		}
	}

	result := make([]*dto.PermissionResourceDTO, 0, len(resourceSet))
	for _, res := range resourceSet {
		result = append(result, res)
	}

	return result, nil
}

// CheckResourcePermission 检查资源权限（统一方法）
// 支持所有资源类型：api_path, menu, button, field, other
func (s *PermissionService) CheckResourcePermission(ctx context.Context, params *param.CheckResourcePermissionParams) (bool, error) {
	if err := params.Validate(); err != nil {
		return false, errors.NewValidationError(err.Error())
	}

	account, err := s.authService.FindAccountByID(ctx, params.AccountID)
	if err != nil {
		return false, errors.WrapBizError(err, "获取账户失败")
	}

	// 检查账户状态
	if !account.IsActive() {
		return false, errors.NewBusinessError("账户未激活")
	}

	// 使用统一的 FindResources 方法查找资源（支持通配符）
	resources, err := s.permissionResourceRepo.FindResources(ctx, params.ResourceType, params.ResourceID, false)
	if err != nil {
		return false, errors.WrapBizError(err, "查询权限资源失败")
	}

	if len(resources) == 0 {
		return false, nil
	}

	// 获取账户权限
	permissionCodes, err := s.getAccountPermissions(ctx, params.AccountID)
	if err != nil {
		return false, err
	}

	// 检查账户是否拥有该权限
	permissionSet := make(map[string]bool)
	for _, code := range permissionCodes {
		permissionSet[code] = true
	}

	for _, resource := range resources {
		// 获取权限信息
		permission, err := s.permissionRepo.FindByID(ctx, resource.PermissionID)
		if err != nil || permission == nil {
			continue
		}
		if permissionSet[permission.Code] && permission.IsActive {
			return true, nil
		}
	}

	return false, nil
}

// GetPermissionWithResources 获取权限及其关联的资源
func (s *PermissionService) GetPermissionWithResources(ctx context.Context, permissionID string) (*dto.PermissionDTO, []*dto.PermissionResourceDTO, error) {
	// 获取权限
	permission, err := s.permissionRepo.FindByID(ctx, permissionID)
	if err != nil {
		return nil, nil, errors.WrapBizError(err, "获取权限失败")
	}
	if permission == nil {
		return nil, nil, errors.NewBusinessError("权限不存在")
	}

	// 转换为 DTO
	permissionDTO := permission.ToDTO()

	// 获取关联资源
	resources, err := s.permissionResourceRepo.FindByPermissionID(ctx, permissionID)
	if err != nil {
		return nil, nil, errors.WrapBizError(err, "获取权限资源失败")
	}

	objects := make([]*dto.PermissionResourceDTO, 0, len(resources))
	for _, r := range resources {
		objects = append(objects, r.ToDTO())
	}

	return permissionDTO, objects, nil
}
