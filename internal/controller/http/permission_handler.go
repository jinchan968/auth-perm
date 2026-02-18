package http

import (
	"auth-perm/internal/common/dto/response"
	controllerUtil "auth-perm/internal/controller/util"
	controllerVo "auth-perm/internal/controller/vo"
	"auth-perm/internal/domain/permission/param"
	permissionService "auth-perm/internal/domain/permission/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	permissionService *permissionService.PermissionService
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(ps *permissionService.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: ps,
	}
}

// CheckPermission 检查账户权限
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	// 使用util工具获取认证信息
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}

	var req controllerVo.CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	allowed, err := h.permissionService.CheckPermission(c.Request.Context(), accountID, req.PermissionCode)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查权限失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"allowed":         allowed,
		"permission_code": req.PermissionCode,
	})
}

// CheckRole 检查账户角色
func (h *PermissionHandler) CheckRole(c *gin.Context) {
	// 使用util工具获取认证信息
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}

	roleCode := c.Param("role")
	if roleCode == "" {
		response.Error(c, http.StatusBadRequest, "角色代码不能为空", "")
		return
	}

	hasRole, err := h.permissionService.CheckRole(c.Request.Context(), accountID, roleCode)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查角色失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"has_role":  hasRole,
		"role_code": roleCode,
	})
}

// GetPermissions 获取账户所有权限
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	// 使用util工具获取认证信息
	currentAccountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}

	// 获取目标账户ID
	targetAccountID := c.Query("account_id")
	if targetAccountID == "" {
		targetAccountID = currentAccountID
	}

	// 构建参数
	params := param.NewGetUserDataWithAuthCheckParams(currentAccountID, targetAccountID)

	// 调用权限服务
	permissions, err := h.permissionService.GetAccountPermissionsWithAuthCheck(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取账户权限失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id":  targetAccountID,
		"permissions": permissions,
	})
}

// GetRoles 获取账户所有角色
func (h *PermissionHandler) GetRoles(c *gin.Context) {
	// 使用util工具获取认证信息
	currentAccountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}

	// 获取目标账户ID
	targetAccountID := c.Query("account_id")
	if targetAccountID == "" {
		targetAccountID = currentAccountID
	}

	// 构建参数
	params := param.NewGetUserDataWithAuthCheckParams(currentAccountID, targetAccountID)

	// 调用权限服务
	roles, err := h.permissionService.GetAccountRolesWithAuthCheck(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取账户角色失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id": targetAccountID,
		"roles":      roles,
	})
}

// CheckAnyPermission 检查账户是否拥有任意一个权限
func (h *PermissionHandler) CheckAnyPermission(c *gin.Context) {
	var req struct {
		AccountID   string   `json:"account_id"`
		Permissions []string `json:"permissions" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	targetAccountID := req.AccountID
	if targetAccountID == "" {
		// 使用util工具获取认证信息
		accountID, err := controllerUtil.GetAccountID(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, err.Error(), "")
			return
		}
		targetAccountID = accountID
	}

	hasPermission, err := h.permissionService.CheckAnyPermission(c.Request.Context(), targetAccountID, req.Permissions)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查权限失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id":     targetAccountID,
		"has_permission": hasPermission,
		"permissions":    req.Permissions,
	})
}

// CheckAllPermissions 检查账户是否拥有所有权限
func (h *PermissionHandler) CheckAllPermissions(c *gin.Context) {
	var req struct {
		AccountID   string   `json:"account_id"`
		Permissions []string `json:"permissions" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	targetAccountID := req.AccountID
	if targetAccountID == "" {
		// 使用util工具获取认证信息
		accountID, err := controllerUtil.GetAccountID(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, err.Error(), "")
			return
		}
		targetAccountID = accountID
	}

	hasPermission, err := h.permissionService.CheckAllPermissions(c.Request.Context(), targetAccountID, req.Permissions)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查权限失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id":     targetAccountID,
		"has_permission": hasPermission,
		"permissions":    req.Permissions,
	})
}

// CheckAnyRole 检查账户是否拥有任意一个角色
func (h *PermissionHandler) CheckAnyRole(c *gin.Context) {
	var req struct {
		AccountID string   `json:"account_id"`
		RoleCodes []string `json:"role_codes" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	targetAccountID := req.AccountID
	if targetAccountID == "" {
		// 从中间件获取账户ID
		accountID, err := controllerUtil.GetAccountID(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, err.Error(), "")
			return
		}
		targetAccountID = accountID
	}

	hasRole, err := h.permissionService.CheckAnyRole(c.Request.Context(), targetAccountID, req.RoleCodes)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查角色失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id": targetAccountID,
		"has_role":   hasRole,
		"role_codes": req.RoleCodes,
	})
}

// CheckAllRoles 检查账户是否拥有所有角色
func (h *PermissionHandler) CheckAllRoles(c *gin.Context) {
	var req struct {
		AccountID string   `json:"account_id"`
		RoleCodes []string `json:"role_codes" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	targetAccountID := req.AccountID
	if targetAccountID == "" {
		// 从中间件获取账户ID
		accountID, err := controllerUtil.GetAccountID(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, err.Error(), "")
			return
		}
		targetAccountID = accountID
	}

	hasRole, err := h.permissionService.CheckAllRoles(c.Request.Context(), targetAccountID, req.RoleCodes)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查角色失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id": targetAccountID,
		"has_role":   hasRole,
		"role_codes": req.RoleCodes,
	})
}

// CheckOrgPermission 检查组织权限
func (h *PermissionHandler) CheckOrgPermission(c *gin.Context) {
	orgID := c.Param("org_id")
	if orgID == "" {
		response.Error(c, http.StatusBadRequest, "组织ID不能为空", "")
		return
	}

	var req controllerVo.CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 使用util工具获取认证信息
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}

	checkOrgParams := param.NewCheckOrgPermissionParams(accountID, orgID, req.PermissionCode)
	hasPermission, err := h.permissionService.CheckOrgPermission(c.Request.Context(), checkOrgParams)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查组织权限失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id":      accountID,
		"org_id":          orgID,
		"permission_code": req.PermissionCode,
		"has_permission":  hasPermission,
	})
}

// IsOrgAdmin 检查是否为组织管理员
func (h *PermissionHandler) IsOrgAdmin(c *gin.Context) {
	orgID := c.Param("org_id")
	if orgID == "" {
		response.Error(c, http.StatusBadRequest, "组织ID不能为空", "")
		return
	}

	// 使用util工具获取认证信息
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}

	isAdminParams := param.NewIsOrgAdminParams(accountID, orgID)
	isOrgAdmin, err := h.permissionService.IsOrgAdmin(c.Request.Context(), isAdminParams)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查组织管理员失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id":   accountID,
		"org_id":       orgID,
		"is_org_admin": isOrgAdmin,
	})
}

// IsSuperAdmin 检查是否为超级管理员
func (h *PermissionHandler) IsSuperAdmin(c *gin.Context) {
	// 使用util工具获取认证信息
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}

	// 如果指定了account_id参数，且有管理权限，可以检查其他账户
	targetAccountID := accountID
	if accountIDParam := c.Query("account_id"); accountIDParam != "" {
		// 检查是否有查看其他账户权限的权限
		canView, err := h.permissionService.CheckPermission(c.Request.Context(), accountID, "users.read")
		if err != nil || !canView {
			response.Error(c, http.StatusForbidden, "没有权限检查其他账户", "")
			return
		}
		targetAccountID = accountIDParam
	}

	isAdminParams := param.NewIsSystemAdminParams(targetAccountID)
	isSuperAdmin, err := h.permissionService.IsSystemAdmin(c.Request.Context(), isAdminParams)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查超级管理员失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id":     targetAccountID,
		"is_super_admin": isSuperAdmin,
	})
}

// GetEffectivePermissions 获取有效权限
func (h *PermissionHandler) GetEffectivePermissions(c *gin.Context) {
	// 使用util工具获取认证信息
	currentAccountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}

	// 获取目标账户ID
	targetAccountID := c.Query("account_id")
	if targetAccountID == "" {
		targetAccountID = currentAccountID
	}

	// 构建参数
	params := param.NewGetUserDataWithAuthCheckParams(currentAccountID, targetAccountID)

	// 调用权限服务
	permissions, roles, err := h.permissionService.GetEffectivePermissionsWithAuthCheck(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取权限和角色失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id":  targetAccountID,
		"permissions": permissions,
		"roles":       roles,
	})
}

// ==================== Permission CRUD ====================

// CreatePermission 创建权限
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req controllerVo.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.CreatePermissionParams{
		TenantID:    req.TenantID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    req.IsSystem,
	}

	result, err := h.permissionService.CreatePermission(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建权限失败", err.Error())
		return
	}

	response.Success(c, result)
}

// UpdatePermission 更新权限
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	var req controllerVo.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.UpdatePermissionParams{
		ID:          req.ID,
		TenantID:    req.TenantID,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	result, err := h.permissionService.UpdatePermission(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新权限失败", err.Error())
		return
	}

	response.Success(c, result)
}

// DeletePermission 删除权限
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.Query("tenant_id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "权限ID不能为空", "")
		return
	}

	params := &param.DeletePermissionParams{ID: id, TenantID: tenantID}

	err := h.permissionService.DeletePermission(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "删除权限失败", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// GetPermission 获取权限详情
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.Query("tenant_id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "权限ID不能为空", "")
		return
	}

	result, err := h.permissionService.GetPermission(c.Request.Context(), id, tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取权限失败", err.Error())
		return
	}

	response.Success(c, result)
}

// ListPermissions 获取权限列表
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	code := c.Query("code")
	name := c.Query("name")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	params := param.NewListPermissionParams(tenantID, page, pageSize)
	params.Code = code
	params.Name = name

	results, total, err := h.permissionService.ListPermissions(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取权限列表失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":  results,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// ==================== Role CRUD ====================

// CreateRole 创建角色
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req controllerVo.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.CreateRoleParams{
		TenantID:    req.TenantID,
		OrgID:       req.OrgID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Priority:    req.Priority,
		IsSystem:    req.IsSystem,
	}

	result, err := h.permissionService.CreateRole(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建角色失败", err.Error())
		return
	}

	response.Success(c, result)
}

// UpdateRole 更新角色
func (h *PermissionHandler) UpdateRole(c *gin.Context) {
	var req controllerVo.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.UpdateRoleParams{
		ID:          req.ID,
		TenantID:    req.TenantID,
		Name:        req.Name,
		Description: req.Description,
		Priority:    req.Priority,
		IsActive:    req.IsActive,
	}

	result, err := h.permissionService.UpdateRole(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新角色失败", err.Error())
		return
	}

	response.Success(c, result)
}

// DeleteRole 删除角色
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.Query("tenant_id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "角色ID不能为空", "")
		return
	}

	params := &param.DeleteRoleParams{ID: id, TenantID: tenantID}

	err := h.permissionService.DeleteRole(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "删除角色失败", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// GetRole 获取角色详情
func (h *PermissionHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.Query("tenant_id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "角色ID不能为空", "")
		return
	}

	result, err := h.permissionService.GetRole(c.Request.Context(), id, tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取角色失败", err.Error())
		return
	}

	response.Success(c, result)
}

// ListRoles 获取角色列表
func (h *PermissionHandler) ListRolesHandler(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	orgID := c.Query("org_id")
	code := c.Query("code")
	name := c.Query("name")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	params := param.NewListRoleParams(tenantID, page, pageSize)
	params.OrgID = orgID
	params.Code = code
	params.Name = name

	results, total, err := h.permissionService.ListRoles(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取角色列表失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":  results,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// ==================== Role-Permission 关联管理 ====================

// AssignPermissionToRole 分配权限给角色
func (h *PermissionHandler) AssignPermissionToRole(c *gin.Context) {
	var req controllerVo.AssignPermissionToRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.AssignPermissionToRoleParams{
		RoleID:        req.RoleID,
		PermissionIDs: req.PermissionIDs,
		TenantID:      req.TenantID,
	}

	err := h.permissionService.AssignPermissionToRole(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "分配权限失败", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "分配成功"})
}

// RemovePermissionFromRole 移除角色权限
func (h *PermissionHandler) RemovePermissionFromRole(c *gin.Context) {
	var req controllerVo.RemovePermissionFromRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.RemovePermissionFromRoleParams{
		RoleID:       req.RoleID,
		PermissionID: req.PermissionID,
		TenantID:     req.TenantID,
	}

	err := h.permissionService.RemovePermissionFromRole(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "移除权限失败", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "移除成功"})
}

// GetRolePermissions 获取角色的权限列表
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	tenantID := c.Query("tenant_id")
	if roleID == "" {
		response.Error(c, http.StatusBadRequest, "角色ID不能为空", "")
		return
	}

	results, err := h.permissionService.GetRolePermissions(c.Request.Context(), roleID, tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取角色权限失败", err.Error())
		return
	}

	response.Success(c, gin.H{"data": results})
}

// ==================== Account-Role 关联管理 ====================

// AssignRoleToAccount 分配角色给账户
func (h *PermissionHandler) AssignRoleToAccount(c *gin.Context) {
	var req controllerVo.AssignRoleToAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.AssignRoleToAccountParams{
		AccountID: req.AccountID,
		RoleIDs:   req.RoleIDs,
		TenantID:  req.TenantID,
	}

	err := h.permissionService.AssignRoleToAccount(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "分配角色失败", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "分配成功"})
}

// RemoveRoleFromAccount 移除账户角色
func (h *PermissionHandler) RemoveRoleFromAccount(c *gin.Context) {
	var req controllerVo.RemoveRoleFromAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.RemoveRoleFromAccountParams{
		AccountID: req.AccountID,
		RoleID:    req.RoleID,
		TenantID:  req.TenantID,
	}

	err := h.permissionService.RemoveRoleFromAccount(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "移除角色失败", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "移除成功"})
}
