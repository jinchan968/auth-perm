package handler

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

// OrganizationHandler 组织处理器
type OrganizationHandler struct {
	orgService *permissionService.OrganizationService
}

// NewOrganizationHandler 创建组织处理器
func NewOrganizationHandler(os *permissionService.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{orgService: os}
}

func (h *OrganizationHandler) Create(c *gin.Context) {
	tenantID, err := controllerUtil.GetTenantID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "获取租户ID失败", err.Error())
		return
	}
	var req controllerVo.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	result, err := h.orgService.Create(c.Request.Context(), &param.CreateOrganizationParams{
		TenantID: tenantID, ParentID: req.ParentID, Code: req.Code,
		Name: req.Name, Description: req.Description, SortOrder: req.SortOrder,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建组织失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *OrganizationHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "组织ID不能为空", "")
		return
	}
	var req controllerVo.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	result, err := h.orgService.Update(c.Request.Context(), &param.UpdateOrganizationParams{
		ID: id, Name: req.Name, Description: req.Description,
		ParentID: req.ParentID, IsActive: req.IsActive, SortOrder: req.SortOrder,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新组织失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *OrganizationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "组织ID不能为空", "")
		return
	}
	if err := h.orgService.Delete(c.Request.Context(), &param.DeleteOrganizationParams{ID: id}); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除组织失败", err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *OrganizationHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "组织ID不能为空", "")
		return
	}
	result, err := h.orgService.Get(c.Request.Context(), &param.GetOrganizationParams{ID: id})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取组织失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *OrganizationHandler) List(c *gin.Context) {
	tenantID, err := controllerUtil.GetTenantID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "获取租户ID失败", err.Error())
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	params := &param.ListOrganizationsParams{
		TenantID: tenantID, ParentID: c.Query("parent_id"),
		Keyword: c.Query("keyword"), Page: page, Size: size,
	}
	result, total, err := h.orgService.List(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取组织列表失败", err.Error())
		return
	}
	response.Success(c, gin.H{"data": result, "total": total, "page": params.Page, "size": params.Size})
}

func (h *OrganizationHandler) GetTree(c *gin.Context) {
	tenantID, err := controllerUtil.GetTenantID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "获取租户ID失败", err.Error())
		return
	}
	result, err := h.orgService.GetTree(c.Request.Context(), tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取组织树失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *OrganizationHandler) AssignAccountToOrg(c *gin.Context) {
	tenantID, err := controllerUtil.GetTenantID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "获取租户ID失败", err.Error())
		return
	}
	var req controllerVo.AssignAccountToOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if err := h.orgService.AssignAccountToOrg(c.Request.Context(), &param.AssignAccountToOrgParams{
		AccountID: req.AccountID, OrgID: req.OrgID, TenantID: tenantID,
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "分配账户到组织失败", err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *OrganizationHandler) RemoveAccountFromOrg(c *gin.Context) {
	var req controllerVo.RemoveAccountFromOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if err := h.orgService.RemoveAccountFromOrg(c.Request.Context(), &param.RemoveAccountFromOrgParams{
		AccountID: req.AccountID, OrgID: req.OrgID,
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "从组织移除账户失败", err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *OrganizationHandler) GetUserOrganizations(c *gin.Context) {
	accountID := c.Param("accountId")
	if accountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "")
		return
	}
	result, err := h.orgService.GetUserOrganizations(c.Request.Context(), accountID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取用户组织失败", err.Error())
		return
	}
	response.Success(c, result)
}
