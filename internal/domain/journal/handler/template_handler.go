package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/controller/util"
	"auth-perm/internal/domain/journal/dto"
	"auth-perm/internal/domain/journal/service"
	"auth-perm/internal/domain/journal/vo"
)

// TemplateHandler 模板处理器
type TemplateHandler struct {
	templateSvc *service.TemplateService
}

// NewTemplateHandler 创建模板处理器
func NewTemplateHandler(templateSvc *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateSvc: templateSvc}
}

// ListTemplates 获取模板列表
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	page, pageSize, _ := util.GetPaginationParams(c)
	name := c.Query("name")
	tag := c.Query("tag")

	params := &dto.ListTemplateParams{
		TenantID: tenantID,
		Page:     page,
		PageSize: pageSize,
		Name:     name,
		Tag:      tag,
	}

	templates, total, err := h.templateSvc.List(c.Request.Context(), params)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{
		"data":      dto.FromTemplateDOList(templates),
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTemplate 获取模板详情
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	templateID := c.Param("id")
	if templateID == "" {
		response.Error(c, http.StatusBadRequest, "模板ID不能为空", "")
		return
	}

	template, err := h.templateSvc.Get(c.Request.Context(), templateID, tenantID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, dto.FromTemplateDO(template))
}

// CreateTemplate 创建模板
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req vo.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	params := &dto.CreateTemplateParams{
		TenantID:  tenantID,
		AccountID: auth.AccountID,
		Name:      req.Name,
		Content:   req.Content,
		Tags:      req.Tags,
	}

	template, err := h.templateSvc.Create(c.Request.Context(), params)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, dto.FromTemplateDO(template))
}

// UpdateTemplate 更新模板
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	templateID := c.Param("id")
	if templateID == "" {
		response.Error(c, http.StatusBadRequest, "模板ID不能为空", "")
		return
	}

	var req vo.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	params := &dto.UpdateTemplateParams{
		ID:        templateID,
		TenantID:  tenantID,
		AccountID: auth.AccountID,
		Name:      req.Name,
		Content:   req.Content,
		Tags:      req.Tags,
	}

	template, err := h.templateSvc.Update(c.Request.Context(), params)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, dto.FromTemplateDO(template))
}

// DeleteTemplate 删除模板
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	templateID := c.Param("id")
	if templateID == "" {
		response.Error(c, http.StatusBadRequest, "模板ID不能为空", "")
		return
	}

	err = h.templateSvc.Delete(c.Request.Context(), templateID, tenantID, auth.AccountID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}