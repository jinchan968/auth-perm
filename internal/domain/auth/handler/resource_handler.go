package handler

import (
	"net/http"

	"auth-perm/config"
	"auth-perm/internal/common/dto/response"
	controllerUtil "auth-perm/internal/controller/util"
	permissionService "auth-perm/internal/domain/permission/service"

	"github.com/gin-gonic/gin"
)

// ResourceHandler 资源权限处理器
type ResourceHandler struct {
	cfg               *config.Config
	permissionService *permissionService.PermissionService
}

// NewResourceHandler 创建资源权限处理器
func NewResourceHandler(cfg *config.Config, permService *permissionService.PermissionService) *ResourceHandler {
	return &ResourceHandler{
		cfg:               cfg,
		permissionService: permService,
	}
}

// GetMyResources 获取当前用户可访问的资源清单
func (h *ResourceHandler) GetMyResources(c *gin.Context) {
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", "无法获取账户信息")
		return
	}

	isSuperAdmin := h.cfg.Server.SuperAdmin != "" && c.GetString("username") == h.cfg.Server.SuperAdmin
	resources, err := h.permissionService.GetMyResources(c.Request.Context(), accountID, isSuperAdmin)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取资源清单失败", err.Error())
		return
	}

	if isSuperAdmin {
		c.Set("is_super_admin", true)
	}

	response.Success(c, gin.H{
		"resources": resources,
	})
}
