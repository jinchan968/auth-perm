package handler

import (
	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/domain/tenant/dto"
	"auth-perm/internal/domain/tenant/param"
	tenantService "auth-perm/internal/domain/tenant/service"
	"auth-perm/internal/domain/tenant/vo"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TenantHandler 租户处理器
type TenantHandler struct {
	tenantService *tenantService.TenantService
}

// NewTenantHandler 创建租户处理器
func NewTenantHandler(ts *tenantService.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: ts}
}

// Create 创建租户
func (h *TenantHandler) Create(c *gin.Context) {
	var req vo.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.CreateTenantParams{
		Name:     req.Name,
		Code:     req.Code,
		Plan:     req.Plan,
		ExpireAt: req.ExpireAt,
	}

	result, err := h.tenantService.Create(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建租户失败", err.Error())
		return
	}

	response.Success(c, h.toResponse(result))
}

// Update 更新租户
func (h *TenantHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	var req vo.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.UpdateTenantParams{
		ID:       id,
		Name:     req.Name,
		Status:   req.Status,
		Plan:     req.Plan,
		ExpireAt: req.ExpireAt,
		Settings: req.Settings,
	}

	result, err := h.tenantService.Update(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新租户失败", err.Error())
		return
	}

	response.Success(c, h.toResponse(result))
}

// Delete 删除租户
func (h *TenantHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	params := &param.DeleteTenantParams{ID: id}

	err := h.tenantService.Delete(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "删除租户失败", err.Error())
		return
	}

	response.Success(c, nil)
}

// Enable 启用租户
func (h *TenantHandler) Enable(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	params := &param.EnableTenantParams{ID: id}

	err := h.tenantService.Enable(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "启用租户失败", err.Error())
		return
	}

	response.Success(c, nil)
}

// Get 获取租户详情
func (h *TenantHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	params := &param.GetTenantParams{ID: id}

	result, err := h.tenantService.Get(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取租户失败", err.Error())
		return
	}

	response.Success(c, h.toResponse(result))
}

// List 列出租户
func (h *TenantHandler) List(c *gin.Context) {
	keyword := c.Query("keyword")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	params := &param.ListTenantsParams{
		Keyword: keyword,
		Status:  status,
		Page:    page,
		Size:    size,
	}

	result, total, err := h.tenantService.List(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取租户列表失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":  h.toListResponse(result),
		"total": total,
		"page":  params.Page,
		"size":  params.Size,
	})
}

// GetSettings 获取租户设置
func (h *TenantHandler) GetSettings(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	params := &param.GetTenantParams{ID: id}

	result, err := h.tenantService.Get(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取租户设置失败", err.Error())
		return
	}

	response.Success(c, result.Settings)
}

// UpdateSettings 更新租户设置
func (h *TenantHandler) UpdateSettings(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	var req vo.UpdateTenantSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	params := &param.UpdateTenantSettingsParams{
		ID:       id,
		Settings: &req.Settings,
	}

	result, err := h.tenantService.UpdateSettings(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新租户设置失败", err.Error())
		return
	}

	response.Success(c, h.toResponse(result))
}

// ChangeStatus 变更租户状态
func (h *TenantHandler) ChangeStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	var req vo.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	err := h.tenantService.ChangeStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "变更租户状态失败", err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 私有方法 ====================

// toResponse 转换为响应
func (h *TenantHandler) toResponse(dto *dto.TenantDTO) *vo.TenantResponse {
	if dto == nil {
		return nil
	}
	return &vo.TenantResponse{
		ID:        dto.ID,
		Name:      dto.Name,
		Code:      dto.Code,
		Status:    dto.Status,
		Plan:      dto.Plan,
		ExpireAt:  dto.ExpireAt,
		Settings:  dto.Settings,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

// toListResponse 转换为列表响应
func (h *TenantHandler) toListResponse(dtos []*dto.TenantListItemDTO) []vo.TenantListItemResponse {
	if dtos == nil {
		return nil
	}
	result := make([]vo.TenantListItemResponse, len(dtos))
	for i, dto := range dtos {
		result[i] = vo.TenantListItemResponse{
			ID:        dto.ID,
			Name:      dto.Name,
			Code:      dto.Code,
			Status:    dto.Status,
			Plan:      dto.Plan,
			ExpireAt:  dto.ExpireAt,
			UserCount: dto.UserCount,
			CreatedAt: dto.CreatedAt,
		}
	}
	return result
}
