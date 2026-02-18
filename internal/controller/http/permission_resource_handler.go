package http

import (
	"auth-perm/internal/common/dto/response"
	controllerUtil "auth-perm/internal/controller/util"
	"auth-perm/internal/domain/permission/param"
	permissionService "auth-perm/internal/domain/permission/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PermissionResourceHandler 权限资源处理器
type PermissionResourceHandler struct {
	permissionService *permissionService.PermissionService
}

// NewPermissionResourceHandler 创建权限资源处理器
func NewPermissionResourceHandler(ps *permissionService.PermissionService) *PermissionResourceHandler {
	return &PermissionResourceHandler{
		permissionService: ps,
	}
}

// Create 创建权限资源关联
// POST /api/v1/permissions/resources
func (h *PermissionResourceHandler) Create(c *gin.Context) {
	var req param.CreatePermissionResourceParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	if req.PermissionID == "" {
		response.Error(c, http.StatusBadRequest, "权限ID不能为空", "")
		return
	}

	// 如果请求体中没有 tenant_id，则从请求头获取（可选）
	if req.TenantID == "" {
		if tenantID, err := controllerUtil.GetTenantID(c); err == nil {
			req.TenantID = tenantID
		}
		// 允许 tenant_id 为空，不强制要求
	}

	result, err := h.permissionService.CreatePermissionResource(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建权限资源关联失败", err.Error())
		return
	}

	response.Success(c, result)
}

// CreateBatch 批量创建权限资源关联
// POST /api/v1/permissions/resources/batch
func (h *PermissionResourceHandler) CreateBatch(c *gin.Context) {
	var req param.BindPermissionResourcesParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	if req.PermissionID == "" {
		response.Error(c, http.StatusBadRequest, "权限ID不能为空", "")
		return
	}

	// 如果请求体中没有 tenant_id，则从请求头获取（可选）
	if req.TenantID == "" {
		if tenantID, err := controllerUtil.GetTenantID(c); err == nil {
			req.TenantID = tenantID
		}
	}

	result, err := h.permissionService.CreatePermissionResourcesBatch(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "批量创建权限资源关联失败", err.Error())
		return
	}

	response.Success(c, result)
}

// Update 更新权限资源关联
// PUT /api/v1/permissions/:permissionId/resources/:resourceId
func (h *PermissionResourceHandler) Update(c *gin.Context) {
	resourceID := c.Param("resourceId")
	if resourceID == "" {
		response.Error(c, http.StatusBadRequest, "资源ID不能为空", "")
		return
	}

	var req param.UpdatePermissionResourceParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	req.ID = resourceID

	result, err := h.permissionService.UpdatePermissionResource(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新权限资源关联失败", err.Error())
		return
	}

	response.Success(c, result)
}

// Delete 删除权限资源关联
// DELETE /api/v1/permissions/:permissionId/resources/:resourceId
func (h *PermissionResourceHandler) Delete(c *gin.Context) {
	resourceID := c.Param("resourceId")
	if resourceID == "" {
		response.Error(c, http.StatusBadRequest, "资源ID不能为空", "")
		return
	}

	// 这里需要通过ID删除，所以需要先找到对应的关联ID
	// 简化处理：直接使用 resourceID 作为关联ID
	err := h.permissionService.DeletePermissionResource(c.Request.Context(), resourceID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "删除权限资源关联失败", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// List 获取权限的所有资源
// GET /api/v1/permissions/resources?permission_id=xxx
func (h *PermissionResourceHandler) List(c *gin.Context) {
	permissionID := c.Query("permission_id")
	if permissionID == "" {
		response.Error(c, http.StatusBadRequest, "权限ID不能为空", "")
		return
	}

	resourceType := c.Query("resource_type")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	params := param.NewListPermissionResourceParams(permissionID, resourceType, page, pageSize)

	resources, total, err := h.permissionService.GetPermissionResources(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取权限资源列表失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":  resources,
		"total": total,
		"page":  params.Page,
		"size":  params.PageSize,
	})
}

// Bind 绑定权限资源（批量）
// POST /api/v1/permissions/resources/bind
func (h *PermissionResourceHandler) Bind(c *gin.Context) {
	var req param.BindPermissionResourcesParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	if req.PermissionID == "" {
		response.Error(c, http.StatusBadRequest, "权限ID不能为空", "")
		return
	}

	// 如果请求体中没有 tenant_id，则从请求头获取（可选）
	if req.TenantID == "" {
		if tenantID, err := controllerUtil.GetTenantID(c); err == nil {
			req.TenantID = tenantID
		}
	}

	result, err := h.permissionService.BindPermissionResources(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "绑定权限资源失败", err.Error())
		return
	}

	response.Success(c, result)
}

// Unbind 解绑权限资源
// POST /api/v1/permissions/resources/unbind
func (h *PermissionResourceHandler) Unbind(c *gin.Context) {
	var req param.UnbindPermissionResourcesParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	if req.PermissionID == "" {
		response.Error(c, http.StatusBadRequest, "权限ID不能为空", "")
		return
	}

	err := h.permissionService.UnbindPermissionResources(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "解绑权限资源失败", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "解绑成功"})
}

// CheckResourcePermission 检查资源权限
// POST /api/v1/permissions/check-resource
func (h *PermissionResourceHandler) CheckResourcePermission(c *gin.Context) {
	var req param.CheckResourcePermissionParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	allowed, err := h.permissionService.CheckResourcePermission(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "检查资源权限失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"allowed": allowed,
	})
}

// GetAccountResources 获取账户可访问的资源
// GET /api/v1/permissions/account-resources
func (h *PermissionResourceHandler) GetAccountResources(c *gin.Context) {
	accountID := c.Query("account_id")
	resourceType := c.Query("resource_type")

	if accountID == "" {
		// 如果未提供，尝试从上下文获取
		var err error
		accountID, err = controllerUtil.GetAccountID(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "账户ID不能为空", "")
			return
		}
	}

	params := param.NewGetAccountResourcesParams(accountID, resourceType)

	resources, err := h.permissionService.GetAccountResources(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取账户可访问资源失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"resources": resources,
	})
}
