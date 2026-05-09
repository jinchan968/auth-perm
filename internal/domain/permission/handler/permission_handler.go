package handler

import (
	"auth-perm/internal/common/dto/response"
	controllerUtil "auth-perm/internal/controller/util"
	controllerVo "auth-perm/internal/controller/vo"
	permissionConstant "auth-perm/internal/domain/permission/constant"
	"auth-perm/internal/domain/permission/param"
	permissionService "auth-perm/internal/domain/permission/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	permissionService *permissionService.PermissionService
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(ps *permissionService.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: ps}
}

func (h *PermissionHandler) CheckPermission(c *gin.Context) {
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
	response.Success(c, gin.H{"allowed": allowed, "permission_code": req.PermissionCode})
}

func (h *PermissionHandler) CheckRole(c *gin.Context) {
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
	response.Success(c, gin.H{"has_role": hasRole, "role_code": roleCode})
}

func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	currentAccountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}
	targetAccountID := c.Query("account_id")
	if targetAccountID == "" {
		targetAccountID = currentAccountID
	}
	permissions, err := h.permissionService.GetAccountPermissionsWithAuthCheck(c.Request.Context(), param.NewGetUserDataWithAuthCheckParams(currentAccountID, targetAccountID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取账户权限失败", err.Error())
		return
	}
	response.Success(c, gin.H{"account_id": targetAccountID, "permissions": permissions})
}

func (h *PermissionHandler) GetRoles(c *gin.Context) {
	currentAccountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}
	targetAccountID := c.Query("account_id")
	if targetAccountID == "" {
		targetAccountID = currentAccountID
	}
	roles, err := h.permissionService.GetAccountRolesWithAuthCheck(c.Request.Context(), param.NewGetUserDataWithAuthCheckParams(currentAccountID, targetAccountID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取账户角色失败", err.Error())
		return
	}
	response.Success(c, gin.H{"account_id": targetAccountID, "roles": roles})
}

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
		id, err := controllerUtil.GetAccountID(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, err.Error(), "")
			return
		}
		targetAccountID = id
	}
	hasPermission, err := h.permissionService.CheckAnyPermission(c.Request.Context(), targetAccountID, req.Permissions)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查权限失败", err.Error())
		return
	}
	response.Success(c, gin.H{"account_id": targetAccountID, "has_permission": hasPermission, "permissions": req.Permissions})
}

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
		id, err := controllerUtil.GetAccountID(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, err.Error(), "")
			return
		}
		targetAccountID = id
	}
	hasPermission, err := h.permissionService.CheckAllPermissions(c.Request.Context(), targetAccountID, req.Permissions)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查权限失败", err.Error())
		return
	}
	response.Success(c, gin.H{"account_id": targetAccountID, "has_permission": hasPermission, "permissions": req.Permissions})
}

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
		id, err := controllerUtil.GetAccountID(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, err.Error(), "")
			return
		}
		targetAccountID = id
	}
	hasRole, err := h.permissionService.CheckAnyRole(c.Request.Context(), targetAccountID, req.RoleCodes)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查角色失败", err.Error())
		return
	}
	response.Success(c, gin.H{"account_id": targetAccountID, "has_role": hasRole, "role_codes": req.RoleCodes})
}

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
		id, err := controllerUtil.GetAccountID(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, err.Error(), "")
			return
		}
		targetAccountID = id
	}
	hasRole, err := h.permissionService.CheckAllRoles(c.Request.Context(), targetAccountID, req.RoleCodes)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查角色失败", err.Error())
		return
	}
	response.Success(c, gin.H{"account_id": targetAccountID, "has_role": hasRole, "role_codes": req.RoleCodes})
}

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
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}
	hasPermission, err := h.permissionService.CheckOrgPermission(c.Request.Context(), param.NewCheckOrgPermissionParams(accountID, orgID, req.PermissionCode))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查组织权限失败", err.Error())
		return
	}
	response.Success(c, gin.H{"account_id": accountID, "org_id": orgID, "permission_code": req.PermissionCode, "has_permission": hasPermission})
}

func (h *PermissionHandler) IsOrgAdmin(c *gin.Context) {
	orgID := c.Param("org_id")
	if orgID == "" {
		response.Error(c, http.StatusBadRequest, "组织ID不能为空", "")
		return
	}
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}
	isOrgAdmin, err := h.permissionService.IsOrgAdmin(c.Request.Context(), param.NewIsOrgAdminParams(accountID, orgID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查组织管理员失败", err.Error())
		return
	}
	response.Success(c, gin.H{"account_id": accountID, "org_id": orgID, "is_org_admin": isOrgAdmin})
}

func (h *PermissionHandler) IsSuperAdmin(c *gin.Context) {
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}
	targetAccountID := accountID
	if accountIDParam := c.Query("account_id"); accountIDParam != "" {
		canView, err := h.permissionService.CheckPermission(c.Request.Context(), accountID, permissionConstant.PermissionCodeUsersRead)
		if err != nil || !canView {
			response.Error(c, http.StatusForbidden, "没有权限检查其他账户", "")
			return
		}
		targetAccountID = accountIDParam
	}
	isSuperAdmin, err := h.permissionService.IsSystemAdmin(c.Request.Context(), param.NewIsSystemAdminParams(targetAccountID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查超级管理员失败", err.Error())
		return
	}
	response.Success(c, gin.H{"account_id": targetAccountID, "is_super_admin": isSuperAdmin})
}

func (h *PermissionHandler) GetEffectivePermissions(c *gin.Context) {
	currentAccountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		return
	}
	targetAccountID := c.Query("account_id")
	if targetAccountID == "" {
		targetAccountID = currentAccountID
	}
	permissions, roles, err := h.permissionService.GetEffectivePermissionsWithAuthCheck(c.Request.Context(), param.NewGetUserDataWithAuthCheckParams(currentAccountID, targetAccountID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取权限和角色失败", err.Error())
		return
	}
	response.Success(c, gin.H{"account_id": targetAccountID, "permissions": permissions, "roles": roles})
}

// ==================== Permission CRUD ====================

func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req controllerVo.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	result, err := h.permissionService.CreatePermission(c.Request.Context(), &param.CreatePermissionParams{
		TenantID: req.TenantID, Code: req.Code, Name: req.Name, Description: req.Description, IsSystem: req.IsSystem,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建权限失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	var req controllerVo.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	result, err := h.permissionService.UpdatePermission(c.Request.Context(), &param.UpdatePermissionParams{
		ID: req.ID, TenantID: req.TenantID, Name: req.Name, Description: req.Description, IsActive: req.IsActive,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新权限失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "权限ID不能为空", "")
		return
	}
	if err := h.permissionService.DeletePermission(c.Request.Context(), &param.DeletePermissionParams{ID: id, TenantID: c.Query("tenant_id")}); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除权限失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "权限ID不能为空", "")
		return
	}
	result, err := h.permissionService.GetPermission(c.Request.Context(), id, c.Query("tenant_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取权限失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	page, pageSize, _ := controllerUtil.GetPaginationParams(c)
	params := param.NewListPermissionParams(c.Query("tenant_id"), page, pageSize)
	params.Code = c.Query("code")
	params.Name = c.Query("name")
	results, total, err := h.permissionService.ListPermissions(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取权限列表失败", err.Error())
		return
	}
	response.Success(c, gin.H{"data": results, "total": total, "page": page, "size": pageSize})
}

// ==================== Role CRUD ====================

func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req controllerVo.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	result, err := h.permissionService.CreateRole(c.Request.Context(), &param.CreateRoleParams{
		TenantID: req.TenantID, OrgID: req.OrgID, Code: req.Code, Name: req.Name,
		Description: req.Description, Priority: req.Priority, IsSystem: req.IsSystem,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建角色失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *PermissionHandler) UpdateRole(c *gin.Context) {
	var req controllerVo.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	result, err := h.permissionService.UpdateRole(c.Request.Context(), &param.UpdateRoleParams{
		ID: req.ID, TenantID: req.TenantID, Name: req.Name, Description: req.Description,
		Priority: req.Priority, IsActive: req.IsActive,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新角色失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "角色ID不能为空", "")
		return
	}
	if err := h.permissionService.DeleteRole(c.Request.Context(), &param.DeleteRoleParams{ID: id, TenantID: c.Query("tenant_id")}); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除角色失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *PermissionHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "角色ID不能为空", "")
		return
	}
	result, err := h.permissionService.GetRole(c.Request.Context(), id, c.Query("tenant_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取角色失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *PermissionHandler) ListRolesHandler(c *gin.Context) {
	page, pageSize, _ := controllerUtil.GetPaginationParams(c)
	params := param.NewListRoleParams(c.Query("tenant_id"), page, pageSize)
	params.OrgID = c.Query("org_id")
	params.Code = c.Query("code")
	params.Name = c.Query("name")
	results, total, err := h.permissionService.ListRoles(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取角色列表失败", err.Error())
		return
	}
	response.Success(c, gin.H{"data": results, "total": total, "page": page, "size": pageSize})
}

// ==================== Role-Permission ====================

func (h *PermissionHandler) AssignPermissionToRole(c *gin.Context) {
	var req controllerVo.AssignPermissionToRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if err := h.permissionService.AssignPermissionToRole(c.Request.Context(), &param.AssignPermissionToRoleParams{
		RoleID: req.RoleID, PermissionIDs: req.PermissionIDs, TenantID: req.TenantID,
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "分配权限失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "分配成功"})
}

func (h *PermissionHandler) RemovePermissionFromRole(c *gin.Context) {
	var req controllerVo.RemovePermissionFromRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if err := h.permissionService.RemovePermissionFromRole(c.Request.Context(), &param.RemovePermissionFromRoleParams{
		RoleID: req.RoleID, PermissionID: req.PermissionID, TenantID: req.TenantID,
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "移除权限失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "移除成功"})
}

func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		response.Error(c, http.StatusBadRequest, "角色ID不能为空", "")
		return
	}
	results, err := h.permissionService.GetRolePermissions(c.Request.Context(), roleID, c.Query("tenant_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取角色权限失败", err.Error())
		return
	}
	response.Success(c, gin.H{"data": results})
}

// ==================== Account-Role ====================

func (h *PermissionHandler) AssignRoleToAccount(c *gin.Context) {
	accountID := c.Param("accountId")
	if accountID == "" {
		response.Error(c, http.StatusBadRequest, "请求参数错误", "账户ID不能为空")
		return
	}

	var req controllerVo.AssignRoleToAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if err := h.permissionService.AssignRoleToAccount(c.Request.Context(), &param.AssignRoleToAccountParams{
		AccountID: accountID, RoleIDs: req.RoleIDs, TenantID: req.TenantID,
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "分配角色失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "分配成功"})
}

func (h *PermissionHandler) RemoveRoleFromAccount(c *gin.Context) {
	accountID := c.Param("accountId")
	roleID := c.Param("roleId")
	if accountID == "" || roleID == "" {
		response.Error(c, http.StatusBadRequest, "请求参数错误", "账户ID和角色ID不能为空")
		return
	}

	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "请求参数错误", "租户ID不能为空")
		return
	}

	if err := h.permissionService.RemoveRoleFromAccount(c.Request.Context(), &param.RemoveRoleFromAccountParams{
		AccountID: accountID, RoleID: roleID, TenantID: tenantID,
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "移除角色失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "移除成功"})
}

func (h *PermissionHandler) GetAccountRoles(c *gin.Context) {
	accountID := c.Param("accountId")
	if accountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "")
		return
	}
	roles, err := h.permissionService.GetAccountRolesByAccountID(c.Request.Context(), accountID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取账户角色失败", err.Error())
		return
	}
	response.Success(c, roles)
}
