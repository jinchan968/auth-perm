package handler

import (
	"net/http"
	"strconv"
	"strings"

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
	passwordService *service.PasswordService
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(authService *service.AuthService, registerService *service.RegisterService, passwordService *service.PasswordService) *UserHandler {
	return &UserHandler{authService: authService, registerService: registerService, passwordService: passwordService}
}

// duplicateKeyMsg 将唯一约束冲突错误转换为用户友好的提示信息。
func duplicateKeyMsg(errMsg string) string {
	if !strings.Contains(errMsg, "duplicate key") && !strings.Contains(errMsg, "23505") {
		return ""
	}
	switch {
	case strings.Contains(errMsg, "idx_users_email"), strings.Contains(errMsg, "\"email\""):
		return "该邮箱已被注册"
	case strings.Contains(errMsg, "idx_users_phone"), strings.Contains(errMsg, "\"phone\""):
		return "该手机号已被注册"
	case strings.Contains(errMsg, "idx_users_username"), strings.Contains(errMsg, "\"username\""):
		return "该用户名已被使用"
	case strings.Contains(errMsg, "idx_users_identifier"):
		return "该账号标识已被注册"
	default:
		return "该用户已存在"
	}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return
	}
	keyword := c.Query("keyword")
	statusStr := c.Query("status")
	accountTypeStr := c.Query("account_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	query := &dto.AccountSearchQueryDTO{
		TenantID: tenantID,
		Keyword:  keyword,
		Pagination: &model.Pagination{
			Page: page, PageSize: pageSize, SortBy: "accounts.created_at", SortDesc: true,
		},
	}
	if statusStr != "" {
		s := constant.AccountStatus(statusStr)
		query.Status = &s
	}
	if accountTypeStr != "" {
		t := constant.AccountType(accountTypeStr)
		query.AccountType = &t
	}
	users, total, err := h.authService.ListAccountsByTenant(c.Request.Context(), query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取用户列表失败", err.Error())
		return
	}
	userResponses := make([]controllerVo.UserWithAccountResponse, len(users))
	for i, user := range users {
		userResponses[i].FromUserWithAccountDTO(user)
	}
	response.Success(c, gin.H{"data": userResponses, "total": total, "page": page, "size": pageSize})
}

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
	user, err := h.authService.GetAccountWithUser(c.Request.Context(), accountID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在", err.Error())
		return
	}
	if user.TenantID != tenantID {
		response.Error(c, http.StatusForbidden, "无权访问此用户", "")
		return
	}
	var userResponse controllerVo.UserWithAccountResponse
	userResponse.FromUserWithAccountDTO(user)
	response.Success(c, userResponse)
}

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
	if err := h.authService.UpdateAccountStatus(c.Request.Context(), accountID, req.TenantID, req.Status); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新用户状态失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "状态更新成功"})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req controllerVo.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	registerParams, err := req.ToRegisterParams()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	user, account, err := h.registerService.Register(c.Request.Context(), registerParams)
	if err != nil {
		errMsg := err.Error()
		if msg := duplicateKeyMsg(errMsg); msg != "" {
			response.Error(c, http.StatusConflict, msg, errMsg)
			return
		}
		response.Error(c, http.StatusBadRequest, "创建用户失败", errMsg)
		return
	}
	userWithAccount := &dto.UserWithAccountDTO{
		AccountID: account.ID, TenantID: account.TenantID, AccountType: account.AccountType,
		AccountStatus: account.Status, EmailVerified: account.EmailVerified, LastLoginAt: account.GetLastLoginAt(),
		UserID: user.ID, Username: user.Username, Nickname: user.Nickname, Avatar: user.Avatar,
		Email: user.Email, Phone: user.Phone, UserStatus: user.Status,
		CreatedAt: account.CreatedAt, UpdatedAt: account.UpdatedAt,
	}
	var userResponse controllerVo.UserWithAccountResponse
	userResponse.FromUserWithAccountDTO(userWithAccount)
	response.Success(c, gin.H{"message": "用户创建成功", "data": userResponse})
}

func (h *UserHandler) GetUserAccounts(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, http.StatusBadRequest, "用户ID不能为空", "")
		return
	}
	currentAccountID, err := util.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	currentAccount, err := h.authService.FindAccountByID(c.Request.Context(), currentAccountID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取当前账户信息失败", err.Error())
		return
	}
	if currentAccount.UserID != userID {
		response.Error(c, http.StatusForbidden, "无权访问其他用户的账户", "")
		return
	}
	response.Success(c, gin.H{"message": "此功能暂未实现", "user_id": userID})
}

func (h *UserHandler) ResetUserPassword(c *gin.Context) {
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

	var req controllerVo.AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		response.Error(c, http.StatusBadRequest, "两次输入的密码不一致", "")
		return
	}

	if len(req.NewPassword) < 6 {
		response.Error(c, http.StatusBadRequest, "密码长度不能少于6位", "")
		return
	}

	userWithAccount, err := h.authService.GetAccountWithUser(c.Request.Context(), accountID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "账户不存在", err.Error())
		return
	}

	if userWithAccount.TenantID != tenantID {
		response.Error(c, http.StatusForbidden, "无权重置此用户密码", "")
		return
	}

	identifier := ""
	if userWithAccount.Email != "" {
		identifier = userWithAccount.Email
	} else if userWithAccount.Phone != "" {
		identifier = userWithAccount.Phone
	}

	if identifier == "" {
		response.Error(c, http.StatusBadRequest, "用户邮箱和手机号均为空，无法重置密码", "")
		return
	}

	err = h.passwordService.ResetPassword(c.Request.Context(), identifier, req.NewPassword, constant.ActionAdminResetPassword)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "重置密码失败", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "密码重置成功，已使该用户所有登录会话失效"})
}
