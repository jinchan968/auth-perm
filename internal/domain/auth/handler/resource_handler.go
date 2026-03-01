package handler

import (
	"fmt"
	"net/http"

	"auth-perm/config"
	"auth-perm/internal/common/dto/response"
	controllerUtil "auth-perm/internal/controller/util"
	"auth-perm/internal/domain/auth/service"
	permissionService "auth-perm/internal/domain/permission/service"

	"github.com/gin-gonic/gin"
)

// ResourceHandler 资源权限处理器
type ResourceHandler struct {
	cfg               *config.Config
	authService       *service.AuthService
	permissionService *permissionService.PermissionService
}

// NewResourceHandler 创建资源权限处理器
func NewResourceHandler(cfg *config.Config, authService *service.AuthService, permService *permissionService.PermissionService) *ResourceHandler {
	return &ResourceHandler{
		cfg:               cfg,
		authService:       authService,
		permissionService: permService,
	}
}

// GetMyResources 获取当前用户可访问的资源清单
// GET /api/v1/auth/my-resources
// 用于前端权限控制：返回当前用户拥有的所有资源（菜单、按钮、API路径）
// 超管返回全量资源
func (h *ResourceHandler) GetMyResources(c *gin.Context) {
	// 获取账户ID
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", "无法获取账户信息")
		return
	}

	// 获取 username：优先从 gin context（Redis 缓存），兜底从 DB 查询
	usernameStr := h.resolveUsername(c)

	var resources interface{}

	// 超管：返回全量资源
	if h.cfg.Server.SuperAdmin != "" && usernameStr == h.cfg.Server.SuperAdmin {
		allResources, err := h.permissionService.GetAllResourcesForSuperAdmin(c.Request.Context())
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "获取全量资源失败", err.Error())
			return
		}
		resources = allResources
		c.Set("is_super_admin", true)
	} else {
		// 普通用户：根据权限返回资源
		resourceList, err := h.permissionService.GetAccountResourcesDetailed(c.Request.Context(), accountID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "获取资源清单失败", err.Error())
			return
		}
		resources = resourceList
	}

	response.Success(c, gin.H{
		"resources": resources,
	})
}

// resolveUsername 获取当前用户的 username
// 优先从 gin context（Redis session 缓存），如果为空则回退查 DB
func (h *ResourceHandler) resolveUsername(c *gin.Context) string {
	// 1. 优先从 context 获取（来自 Redis 缓存，零 DB 查询）
	if username, exists := c.Get("username"); exists {
		if s, ok := username.(string); ok && s != "" {
			return s
		}
	}

	// 2. 兜底：通过 user_id 查 DB（旧 session 缓存没有 username 的情况）
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		return ""
	}

	user, err := h.authService.FindUserByID(c.Request.Context(), userIDStr)
	if err != nil || user == nil {
		fmt.Printf("ResourceHandler: Failed to resolve username for user_id=%s: %v\n", userIDStr, err)
		return ""
	}

	return user.Username
}
