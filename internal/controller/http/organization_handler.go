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

// OrganizationHandler 组织处理器
type OrganizationHandler struct {
	orgService *permissionService.OrganizationService
}

// NewOrganizationHandler 创建组织处理器
func NewOrganizationHandler(os *permissionService.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: os,
	}
}

// Create 创建组织
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

	params := &param.CreateOrganizationParams{
		TenantID:    tenantID,
		ParentID:    req.ParentID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}

	result, err := h.orgService.Create(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建组织失败", err.Error())
		return
	}

	response.Success(c, result)
}

// Update 更新组织
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

	params := &param.UpdateOrganizationParams{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	}

	result, err := h.orgService.Update(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新组织失败", err.Error())
		return
	}

	response.Success(c, result)
}

// Delete 删除组织
func (h *OrganizationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "组织ID不能为空", "")
		return
	}

	params := &param.DeleteOrganizationParams{ID: id}

	err := h.orgService.Delete(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "删除组织失败", err.Error())
		return
	}

	response.Success(c, nil)
}

// Get 获取组织详情
func (h *OrganizationHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "组织ID不能为空", "")
		return
	}

	params := &param.GetOrganizationParams{ID: id}

	result, err := h.orgService.Get(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取组织失败", err.Error())
		return
	}

	response.Success(c, result)
}

// List 列出组织
func (h *OrganizationHandler) List(c *gin.Context) {
	tenantID, err := controllerUtil.GetTenantID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "获取租户ID失败", err.Error())
		return
	}

	// 获取查询参数
	parentID := c.Query("parent_id")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	params := &param.ListOrganizationsParams{
		TenantID: tenantID,
		ParentID: parentID,
		Keyword:  keyword,
		Page:     page,
		Size:     size,
	}

	result, total, err := h.orgService.List(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取组织列表失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":  result,
		"total": total,
		"page":  params.Page,
		"size":  params.Size,
	})
}

// GetTree 获取组织树
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

// AssignAccountToOrg 分配账户到组织
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

	params := &param.AssignAccountToOrgParams{
		AccountID: req.AccountID,
		OrgID:     req.OrgID,
		TenantID:  tenantID,
	}

	err = h.orgService.AssignAccountToOrg(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "分配账户到组织失败", err.Error())
		return
	}

	response.Success(c, nil)
}

// RemoveAccountFromOrg 从组织移除账户
func (h *OrganizationHandler) RemoveAccountFromOrg(c *gin.Context) {
	var req controllerVo.RemoveAccountFromOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.RemoveAccountFromOrgParams{
		AccountID: req.AccountID,
		OrgID:     req.OrgID,
	}

	err := h.orgService.RemoveAccountFromOrg(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "从组织移除账户失败", err.Error())
		return
	}

	response.Success(c, nil)
}

// GetUserOrganizations 获取用户所属组织
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
