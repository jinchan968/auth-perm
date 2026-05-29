package handler

import (
	"net/http"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/controller/util"
	"auth-perm/internal/domain/multimodal/constant"
	"auth-perm/internal/domain/multimodal/service"
	"auth-perm/internal/domain/multimodal/vo"

	"github.com/gin-gonic/gin"
)

type MultimodalHandler struct {
	svc *service.MultimodalService
}

func NewMultimodalHandler(svc *service.MultimodalService) *MultimodalHandler {
	return &MultimodalHandler{svc: svc}
}

func (h *MultimodalHandler) RecognizeImage(c *gin.Context) {
	tenantID, err := util.GetTenantID(c)
	if err != nil || tenantID == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	var req vo.RecognizeImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	if len(req.Images) > constant.MaxImages {
		response.Error(c, http.StatusBadRequest, "图片数量不能超过5张", "")
		return
	}
	maxBase64Size := constant.MaxImageSizeMB * 1024 * 1024 * 4 / 3
	for _, img := range req.Images {
		if len(img) > maxBase64Size {
			response.Error(c, http.StatusBadRequest, "单张图片大小不能超过10MB", "")
			return
		}
	}

	result, err := h.svc.RecognizeImage(c.Request.Context(), tenantID, req.Prompt, req.Images)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "识图失败", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *MultimodalHandler) GenerateImage(c *gin.Context) {
	tenantID, err := util.GetTenantID(c)
	if err != nil || tenantID == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	var req vo.GenerateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	result, err := h.svc.GenerateImage(c.Request.Context(), tenantID, req.Prompt, req.Style)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "生成提示词失败", err.Error())
		return
	}

	response.Success(c, result)
}
