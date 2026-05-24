package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/controller/util"
	"auth-perm/internal/domain/journal/constant"
	"auth-perm/internal/domain/journal/service"
	"auth-perm/internal/domain/journal/vo"
)

type AIPredictionHandler struct {
	svc *service.AIPredictionService
}

func NewAIPredictionHandler(svc *service.AIPredictionService) *AIPredictionHandler {
	return &AIPredictionHandler{svc: svc}
}

func (h *AIPredictionHandler) CreatePrediction(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req vo.CreateAIPredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	result, err := h.svc.CreatePrediction(c.Request.Context(), tenantID, auth.AccountID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *AIPredictionHandler) ListPredictions(c *gin.Context) {
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

	result, err := h.svc.ListPredictions(c.Request.Context(), tenantID, auth.AccountID, page, pageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *AIPredictionHandler) GetPrediction(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	result, err := h.svc.GetPrediction(c.Request.Context(), id, auth.AccountID, tenantID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *AIPredictionHandler) ListModels(c *gin.Context) {
	_, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	models := make([]vo.AIModelVO, 0, len(constant.Models))
	for _, m := range constant.Models {
		models = append(models, vo.AIModelVO{ID: m.ID, Name: m.Name})
	}

	replaceable := make([]vo.AIModelVO, 0, len(constant.ReplaceableModelIDs))
	for _, id := range constant.ReplaceableModelIDs {
		for _, m := range constant.Models {
			if m.ID == id {
				replaceable = append(replaceable, vo.AIModelVO{ID: m.ID, Name: m.Name})
			}
		}
	}

	response.Success(c, vo.ListModelsResponse{
		Defaults:            constant.DefaultModelIDs,
		Replaceable:         replaceable,
		All:                 models,
		DefaultSystemPrompt: constant.DefaultSystemPrompt,
		DailyLimit:          constant.DailyCallLimitPerModel,
	})
}

func (h *AIPredictionHandler) GetQuotas(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	quotas, err := h.svc.GetQuotas(c.Request.Context(), tenantID, auth.AccountID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{
		"daily_limit": constant.DailyCallLimitPerModel,
		"remaining":   quotas,
	})
}
