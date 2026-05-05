package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/controller/util"
	"auth-perm/internal/domain/journal/repo"
	"auth-perm/internal/domain/journal/service"
	"auth-perm/internal/domain/journal/vo"
)

// JournalHandler HTTP 处理器
type JournalHandler struct {
	svc *service.JournalService
}

func NewJournalHandler(svc *service.JournalService) *JournalHandler {
	return &JournalHandler{svc: svc}
}

// ---------- 标签 ----------

// ListTags GET /api/v1/journal/tags
func (h *JournalHandler) ListTags(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.ListTags(c.Request.Context(), auth.AccountID, tenantID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, result)
}

// CreateTag POST /api/v1/journal/tags
func (h *JournalHandler) CreateTag(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	var req vo.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.CreateTag(c.Request.Context(), auth.AccountID, tenantID, req.Name, req.Color)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, result)
}

// UpdateTag PUT /api/v1/journal/tags/:id
func (h *JournalHandler) UpdateTag(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	id := c.Param("id")
	var req vo.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.UpdateTag(c.Request.Context(), id, auth.AccountID, tenantID, req.Name, req.Color)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, result)
}

// DeleteTag DELETE /api/v1/journal/tags/:id
func (h *JournalHandler) DeleteTag(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	id := c.Param("id")
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteTag(c.Request.Context(), id, auth.AccountID, tenantID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

// ---------- 札记条目 ----------

// ListEntries GET /api/v1/journal
func (h *JournalHandler) ListEntries(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	page, pageSize, _ := util.GetPaginationParams(c)

	var startDate, endDate time.Time
	if sd := c.Query("start_date"); sd != "" {
		parsed, err := time.Parse("2006-01-02", sd)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "开始日期格式错误，请使用 YYYY-MM-DD", err.Error())
			return
		}
		startDate = parsed
	}
	if ed := c.Query("end_date"); ed != "" {
		parsed, err := time.Parse("2006-01-02", ed)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "结束日期格式错误，请使用 YYYY-MM-DD", err.Error())
			return
		}
		endDate = parsed
	}

	params := &repo.JournalQueryParams{
		TenantID:  tenantID,
		AccountID: auth.AccountID,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		PageSize:  pageSize,
	}

	result, err := h.svc.ListEntries(c.Request.Context(), params)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, result)
}

// CreateEntry POST /api/v1/journal
func (h *JournalHandler) CreateEntry(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	var req vo.CreateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	entryDate, err := time.Parse("2006-01-02", req.EntryDate)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "日期格式错误，请使用 YYYY-MM-DD", err.Error())
		return
	}

	params := &vo.CreateEntryParams{
		TenantID:  tenantID,
		AccountID: auth.AccountID,
		Title:     req.Title,
		Content:   req.Content,
		Weather:   req.Weather,
		Location:  req.Location,
		Period:    req.Period,
		EntryDate: entryDate,
		TagIDs:    req.TagIDs,
	}

	result, err := h.svc.CreateEntry(c.Request.Context(), params)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, result)
}

// GetEntry GET /api/v1/journal/:id
func (h *JournalHandler) GetEntry(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	id := c.Param("id")
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.GetEntry(c.Request.Context(), id, auth.AccountID, tenantID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, result)
}

// AddCorrection POST /api/v1/journal/:id/corrections
func (h *JournalHandler) AddCorrection(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	id := c.Param("id")
	var req vo.AddCorrectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.AddCorrection(c.Request.Context(), id, auth.AccountID, tenantID, req.Content)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, result)
}

// UpdateTags PUT /api/v1/journal/:id/tags
func (h *JournalHandler) UpdateTags(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	id := c.Param("id")
	var req vo.UpdateTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.UpdateTags(c.Request.Context(), id, auth.AccountID, tenantID, req.TagIDs)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, result)
}

// DeleteEntry DELETE /api/v1/journal/:id
func (h *JournalHandler) DeleteEntry(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	id := c.Param("id")
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteEntry(c.Request.Context(), id, auth.AccountID, tenantID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

func requireTenantID(c *gin.Context) (string, bool) {
	tenantID, err := util.GetTenantID(c)
	if err != nil || tenantID == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return "", false
	}
	return tenantID, true
}
