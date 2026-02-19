package http

import (
	"net/http"
	"strconv"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/common/model"
	"auth-perm/internal/controller/util"
	controllerVo "auth-perm/internal/controller/vo"
	"auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/service"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	authService     *service.AuthService
	registerService *service.RegisterService
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(
	authService *service.AuthService,
	registerService *service.RegisterService,
) *UserHandler {
	return &UserHandler{
		authService:     authService,
		registerService: registerService,
	}
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 根据租户ID获取用户列表，支持关键词搜索、状态过滤和分页
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenant_id query string true "租户ID"
// @Param keyword query string false "关键词搜索（用户名/邮箱/手机号）"
// @Param status query string false "账户状态过滤"
// @Param account_type query string false "账户类型过滤"
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认10"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 获取查询参数
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	keyword := c.Query("keyword")
	statusStr := c.Query("status")
	accountTypeStr := c.Query("account_type")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := &dto.AccountSearchQueryDTO{
		TenantID: tenantID,
		Keyword:  keyword,
		Pagination: &model.Pagination{
			Page:     page,
			PageSize: pageSize,
			SortBy:   "accounts.created_at",
			SortDesc: true,
		},
	}

	// 状态过滤
	if statusStr != "" {
		status := constant.AccountStatus(statusStr)
		query.Status = &status
	}

	// 账户类型过滤
	if accountTypeStr != "" {
		accountType := constant.AccountType(accountTypeStr)
		query.AccountType = &accountType
	}

	// 调用服务层
	users, total, err := h.authService.ListAccountsByTenant(c.Request.Context(), query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取用户列表失败", err.Error())
		return
	}

	// 转换为响应格式
	userResponses := make([]controllerVo.UserWithAccountResponse, len(users))
	for i, user := range users {
		userResponses[i].FromUserWithAccountDTO(user)
	}

	response.Success(c, gin.H{
		"data":  userResponses,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据账户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "账户ID"
// @Param tenant_id query string true "租户ID"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Failure 404 {object} response.ErrorResponse "用户不存在"
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	accountID := c.Param("id")
	tenantID := c.Query("tenant_id")

	if accountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "")
		return
	}

	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	// 获取用户详情
	user, err := h.authService.GetAccountWithUser(c.Request.Context(), accountID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在", err.Error())
		return
	}

	// 校验租户归属
	if user.TenantID != tenantID {
		response.Error(c, http.StatusForbidden, "无权访问此用户", "")
		return
	}

	var userResponse controllerVo.UserWithAccountResponse
	userResponse.FromUserWithAccountDTO(user)

	response.Success(c, userResponse)
}

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态
// @Description 更新账户状态（启用/停用/暂停）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "账户ID"
// @Param UpdateUserStatusRequest body vo.UpdateUserStatusRequest true "更新状态请求"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/users/{id}/status [patch]
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "")
		return
	}

	var req controllerVo.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	if req.TenantID == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}

	// 更新状态
	err := h.authService.UpdateAccountStatus(c.Request.Context(), accountID, req.TenantID, req.Status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新用户状态失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "状态更新成功",
	})
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户（管理员功能，等同于注册）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param RegisterRequest body vo.RegisterRequest true "创建用户请求"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req controllerVo.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 复用注册逻辑
	registerParams, err := req.ToRegisterParams()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	user, account, err := h.registerService.Register(c.Request.Context(), registerParams)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "创建用户失败", err.Error())
		return
	}

	// 构建响应
	userWithAccount := &dto.UserWithAccountDTO{
		AccountID:     account.ID,
		TenantID:      account.TenantID,
		AccountType:   account.AccountType,
		AccountStatus: account.Status,
		EmailVerified: account.EmailVerified,
		LastLoginAt:   account.GetLastLoginAt(),
		UserID:        user.ID,
		Username:      user.Username,
		Nickname:      user.Nickname,
		Avatar:        user.Avatar,
		Email:         user.Email,
		Phone:         user.Phone,
		UserStatus:    user.Status,
		CreatedAt:     account.CreatedAt,
		UpdatedAt:     account.UpdatedAt,
	}

	var userResponse controllerVo.UserWithAccountResponse
	userResponse.FromUserWithAccountDTO(userWithAccount)

	response.Success(c, gin.H{
		"message": "用户创建成功",
		"data":    userResponse,
	})
}

// GetUserAccounts 获取用户的所有账户
// @Summary 获取用户的所有账户
// @Description 根据用户ID获取该用户在所有租户下的账户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/users/{id}/accounts [get]
func (h *UserHandler) GetUserAccounts(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, http.StatusBadRequest, "用户ID不能为空", "")
		return
	}

	// 获取当前认证的账户ID，用于权限校验
	currentAccountID, err := util.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	// 获取当前账户信息，用于校验是否为同一用户
	currentAccount, err := h.authService.FindAccountByID(c.Request.Context(), currentAccountID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取当前账户信息失败", err.Error())
		return
	}

	// 只允许查看自己的账户列表
	if currentAccount.UserID != userID {
		response.Error(c, http.StatusForbidden, "无权访问其他用户的账户", "")
		return
	}

	response.Success(c, gin.H{
		"message": "此功能暂未实现",
		"user_id": userID,
	})
}
