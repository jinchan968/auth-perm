package handler

import (
	"auth-perm/internal/common/dto/response"
	controllerUtil "auth-perm/internal/controller/util"
	"auth-perm/internal/domain/permission/param"
	permissionService "auth-perm/internal/domain/permission/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PermissionResourceHandler 权限资源处理器
type PermissionResourceHandler struct {
	permissionService *permissionService.PermissionService
}

// NewPermissionResourceHandler 创建权限资源处理器
func NewPermissionResourceHandler(ps *permissionService.PermissionService) *PermissionResourceHandler {
	return &PermissionResourceHandler{permissionService: ps}
}

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
	if req.TenantID == "" {
		if tenantID, err := controllerUtil.GetTenantID(c); err == nil {
			req.TenantID = tenantID
		}
	}
	result, err := h.permissionService.CreatePermissionResource(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建权限资源关联失败", err.Error())
		return
	}
	response.Success(c, result)
}

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

func (h *PermissionResourceHandler) Delete(c *gin.Context) {
	resourceID := c.Param("resourceId")
	if resourceID == "" {
		response.Error(c, http.StatusBadRequest, "资源ID不能为空", "")
		return
	}
	if err := h.permissionService.DeletePermissionResource(c.Request.Context(), resourceID); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除权限资源关联失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *PermissionResourceHandler) List(c *gin.Context) {
	permissionID := c.Query("permission_id")
	if permissionID == "" {
		response.Error(c, http.StatusBadRequest, "权限ID不能为空", "")
		return
	}
	page, pageSize, _ := controllerUtil.GetPaginationParams(c)
	params := param.NewListPermissionResourceParams(permissionID, c.Query("resource_type"), page, pageSize)
	resources, total, err := h.permissionService.GetPermissionResources(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取权限资源列表失败", err.Error())
		return
	}
	response.Success(c, gin.H{"data": resources, "total": total, "page": params.Page, "size": params.PageSize})
}

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
	if err := h.permissionService.UnbindPermissionResources(c.Request.Context(), &req); err != nil {
		response.Error(c, http.StatusInternalServerError, "解绑权限资源失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "解绑成功"})
}

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
	response.Success(c, gin.H{"allowed": allowed})
}

func (h *PermissionResourceHandler) GetAccountResources(c *gin.Context) {
	accountID := c.Query("account_id")
	if accountID == "" {
		var err error
		accountID, err = controllerUtil.GetAccountID(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "账户ID不能为空", "")
			return
		}
	}
	resources, err := h.permissionService.GetAccountResources(c.Request.Context(), param.NewGetAccountResourcesParams(accountID, c.Query("resource_type")))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取账户可访问资源失败", err.Error())
		return
	}
	response.Success(c, gin.H{"resources": resources})
}
