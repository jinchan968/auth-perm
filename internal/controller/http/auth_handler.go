package http

import (
	"auth-perm/internal/common/constant"
	"auth-perm/internal/common/dto/response"
	controllerUtil "auth-perm/internal/controller/util"
	controllerVo "auth-perm/internal/controller/vo"
	"auth-perm/internal/domain/auth/param"
	"auth-perm/internal/domain/auth/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService        *service.AuthService
	loginService       *service.LoginService
	registerService    *service.RegisterService
	sessionService     *service.SessionService
	emailService       *service.EmailService
	totpService        *service.TOTPService
	security           *service.SecurityService
	oauthService       *service.OAuthService
	passwordService    *service.PasswordService
	deviceService      service.DeviceService
	securityLogService *service.SecurityLogService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(
	authService *service.AuthService,
	loginService *service.LoginService,
	registerService *service.RegisterService,
	sessionService *service.SessionService,
	emailService *service.EmailService,
	totpService *service.TOTPService,
	security *service.SecurityService,
	oauthService *service.OAuthService,
	passwordService *service.PasswordService,
	deviceService service.DeviceService,
	securityLogService *service.SecurityLogService,
) *AuthHandler {
	return &AuthHandler{
		authService:        authService,
		loginService:       loginService,
		registerService:    registerService,
		sessionService:     sessionService,
		emailService:       emailService,
		totpService:        totpService,
		security:           security,
		oauthService:       oauthService,
		passwordService:    passwordService,
		deviceService:      deviceService,
		securityLogService: securityLogService,
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 支持邮箱或手机号登录，返回访问令牌和会话信息
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param LoginRequest body vo.LoginRequest true "登录请求"
// @Success 200 {object} vo.LoginResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 401 {object} response.ErrorResponse "认证失败"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req controllerVo.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 获取客户端信息
	// 优先使用请求体中的IP，其次使用X-Forwarded-For头，最后使用ClientIP
	if req.IPAddress == "" {
		req.IPAddress = c.GetHeader("X-Forwarded-For")
		if req.IPAddress == "" {
			req.IPAddress = c.ClientIP()
		}
	}
	if req.UserAgent == "" {
		req.UserAgent = c.GetHeader("User-Agent")
	}

	// 调用登录服务进行登录
	loginParams := req.ToLoginParams()
	user, account, err := h.loginService.Login(c.Request.Context(), loginParams)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "认证失败", err.Error())
		return
	}

	// 创建会话并生成token
	sessionParams := param.NewSessionTokenParams(
		c.Request.Context(),
		user,
		account,
		req.IPAddress,
		req.UserAgent,
		req.TenantID,
		req.RememberMe,
	)
	loginResult, err := h.loginService.CreateSessionAndToken(sessionParams)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建会话失败", err.Error())
		return
	}

	// 设置Cookie
	cookieExpiry := constant.GetCookieExpiry(req.RememberMe)
	c.SetCookie(
		constant.CookieAuthToken,
		loginResult.Token,
		cookieExpiry,
		constant.CookiePath,
		constant.CookieDomain,
		constant.CookieSecure,
		constant.CookieHTTPOnly,
	)

	loginResp := &controllerVo.LoginResponse{}
	loginResp.FromUserDTO(loginResult.User, loginResult.Account, loginResult.Token, "登录成功", loginResult.Session.GetExpiresAt())
	response.Success(c, loginResp)
}

// Register 用户注册
// @Summary 用户注册
// @Description 支持邮箱或手机号注册，创建新用户账户
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param RegisterRequest body vo.RegisterRequest true "注册请求"
// @Success 200 {object} vo.UserResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req controllerVo.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 调用注册服务进行注册
	registerParams, err := req.ToRegisterParams()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	user, account, err := h.registerService.Register(c.Request.Context(), registerParams)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "注册失败", err.Error())
		return
	}

	// 注册成功后自动登录
	identifier, _ := req.GetIdentifier()
	loginParams := param.NewLoginParams(
		identifier,
		req.Password,
		"",
		c.GetHeader("User-Agent"),
		c.ClientIP(),
		req.TenantID,
		false,
	)

	// 执行登录
	loginUser, loginAccount, err := h.loginService.Login(c.Request.Context(), loginParams)
	if err != nil {
		// 注册成功但登录失败，返回注册成功信息
		userResp := &controllerVo.UserResponse{}
		userResp.FromUserDTO(user)
		accountResp := &controllerVo.AccountResponse{}
		accountResp.FromAccountDTO(account)
		response.Success(c, gin.H{
			"message": "注册成功，请手动登录",
			"user":    userResp,
			"account": accountResp,
		})
		return
	}

	// 创建会话和token
	sessionParams := param.NewSessionTokenParams(
		c.Request.Context(),
		loginUser,
		loginAccount,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		req.TenantID,
		false,
	)
	loginResult, err := h.loginService.CreateSessionAndToken(sessionParams)
	if err != nil {
		// 注册成功但创建会话失败，返回注册成功信息
		userResp := &controllerVo.UserResponse{}
		userResp.FromUserDTO(user)
		accountResp := &controllerVo.AccountResponse{}
		accountResp.FromAccountDTO(account)
		response.Success(c, gin.H{
			"message": "注册成功，请手动登录",
			"user":    userResp,
			"account": accountResp,
		})
		return
	}

	// 设置Cookie
	cookieExpiry := constant.GetCookieExpiry(false)
	c.SetCookie(
		constant.CookieAuthToken,
		loginResult.Token,
		cookieExpiry,
		constant.CookiePath,
		constant.CookieDomain,
		constant.CookieSecure,
		constant.CookieHTTPOnly,
	)

	// 返回登录成功响应
	loginResp := &controllerVo.LoginResponse{}
	loginResp.FromUserDTO(loginResult.User, loginResult.Account, loginResult.Token, "注册并登录成功", loginResult.Session.GetExpiresAt())
	response.Success(c, loginResp)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户主动登出，使当前会话失效
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param LogoutRequest body vo.LogoutRequest false "登出请求"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 获取token
	token, err := c.Cookie(constant.CookieAuthToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未登录", "not logged in")
		return
	}

	// 解析token获取sessionID
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		response.Error(c, http.StatusBadRequest, "无效的token", "invalid token")
		return
	}
	sessionID := parts[1]

	// 登出
	if err := h.sessionService.Logout(c.Request.Context(), sessionID, false, "user_logout"); err != nil {
		response.Error(c, http.StatusInternalServerError, "登出失败", err.Error())
		return
	}

	// 清除Cookie
	c.SetCookie(
		constant.CookieAuthToken,
		"",
		-1,
		constant.CookiePath,
		constant.CookieDomain,
		constant.CookieSecure,
		constant.CookieHTTPOnly,
	)

	response.Success(c, gin.H{
		"message": "登出成功",
	})
}

// RefreshToken 刷新token
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 401 {object} response.ErrorResponse "未认证或刷新令牌无效"
// @Router /api/v1/auth/token/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从Cookie获取刷新token
	refreshToken, err := c.Cookie(constant.CookieAuthToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未登录", "not logged in")
		return
	}

	// 刷新token
	newToken, err := h.loginService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "刷新token失败", err.Error())
		return
	}

	// 设置新Cookie
	c.SetCookie(
		constant.CookieAuthToken,
		newToken,
		int(constant.TokenExpiryDefault.Seconds()),
		constant.CookiePath,
		constant.CookieDomain,
		constant.CookieSecure,
		constant.CookieHTTPOnly,
	)

	response.Success(c, gin.H{
		"token": newToken,
	})
}

// LogoutAllByTenant 管理员按租户登出所有会话
func (h *AuthHandler) LogoutAllByTenant(c *gin.Context) {
	var req controllerVo.LogoutAllByTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 登出所有会话
	if err := h.sessionService.LogoutAllByTenant(c.Request.Context(), req.TenantID, "admin_force_logout"); err != nil {
		response.Error(c, http.StatusInternalServerError, "登出失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"message":   "已登出该租户下的所有会话",
		"tenant_id": req.TenantID,
	})
}

// GetProfile 获取用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} vo.UserResponse "成功响应"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// 从上下文中获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未授权", "unauthorized")
		return
	}

	// 获取用户信息
	user, err := h.authService.FindUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在", err.Error())
		return
	}

	userResp := &controllerVo.UserResponse{}
	userResp.FromUserDTO(user)
	response.Success(c, userResp)
}

// UpdateProfile 更新个人信息
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req controllerVo.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 从上下文中获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未授权", "unauthorized")
		return
	}

	// 构建更新参数
	updateParams := param.NewUpdateProfileParams(userID, req.Nickname, req.Phone, req.Avatar)

	// 更新用户信息
	updatedUser, err := h.authService.UpdateProfile(c.Request.Context(), updateParams)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "更新失败", err.Error())
		return
	}

	userResp := &controllerVo.UserResponse{}
	userResp.FromUserDTO(updatedUser)
	response.Success(c, userResp)
}

// ChangePassword 修改密码
// @Summary 修改用户密码
// @Description 用户修改自己的登录密码，需要验证旧密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ChangePasswordRequest body vo.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误或旧密码错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/auth/password/change [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req controllerVo.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 从上下文中获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未授权", "unauthorized")
		return
	}

	// 构建密码修改参数
	changeParams := param.NewChangePasswordParams(userID, "", req.OldPassword, req.NewPassword, req.NewPassword)

	// 修改密码
	if err := h.passwordService.ChangePassword(c.Request.Context(), changeParams); err != nil {
		response.Error(c, http.StatusBadRequest, "修改密码失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "密码修改成功",
	})
}

// GetSessions 获取用户会话列表
// @Summary 获取用户会话列表
// @Description 获取当前用户的所有活跃会话列表
// @Tags 会话管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认20，最大100"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/auth/sessions [get]
func (h *AuthHandler) GetSessions(c *gin.Context) {
	// 从上下文中获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未授权", "unauthorized")
		return
	}

	// 获取分页参数
	page, pageSize, _ := controllerUtil.GetPaginationParams(c)

	// 获取会话列表参数
	getSessionsParams := param.NewGetSessionsParams(userID, page, pageSize)

	// 获取会话列表
	sessions, pagination, err := h.sessionService.GetUserSessions(c.Request.Context(), getSessionsParams)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取会话列表失败", err.Error())
		return
	}

	// 转换为响应格式
	sessionResponses := make([]controllerVo.SessionResponse, len(sessions))
	for i, session := range sessions {
		deviceInfo := session.GetDeviceInfo()
		sessionResponses[i] = controllerVo.SessionResponse{
			SessionID: session.GetID(),
			DeviceInfo: controllerVo.SessionDeviceInfo{
				Platform:  deviceInfo.Platform,
				Browser:   deviceInfo.Browser,
				Device:    deviceInfo.Device,
				IPAddress: deviceInfo.IPAddress,
				UserAgent: deviceInfo.UserAgent,
			},
			IsActive:   session.IsValid(),
			CreatedAt:  session.GetCreatedAt(),
			LastActive: session.GetLastActiveAt(),
			ExpiresAt:  session.GetExpiresAt(),
		}
	}

	// 返回分页结果
	responseData := controllerVo.SessionsListResponse{
		PaginatedResponse: controllerVo.PaginatedResponse{
			Total:    int64(len(sessionResponses)),
			Page:     pagination.Page,
			PageSize: pagination.PageSize,
		},
		Sessions: sessionResponses,
	}
	response.Success(c, responseData)
}

// ValidateToken 验证token
// @Summary 验证访问令牌
// @Description 验证当前访问令牌是否有效，返回用户信息
// @Tags 认证管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 401 {object} response.ErrorResponse "未认证或令牌无效"
// @Router /api/v1/auth/token/validate [get]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	// 获取token
	token, err := c.Cookie(constant.CookieAuthToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未登录", "not logged in")
		return
	}

	// 解析token获取sessionID
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		response.Error(c, http.StatusBadRequest, "无效的token", "invalid token")
		return
	}
	sessionID := parts[1]

	// 验证会话
	session, err := h.loginService.ValidateSession(c.Request.Context(), parts[0])
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "会话无效", err.Error())
		return
	}

	response.Success(c, gin.H{
		"valid":      true,
		"session_id": sessionID,
		"user_id":    session.GetUserID(),
		"tenant_id":  session.GetTenantID(),
		"expires_at": session.GetExpiresAt(),
	})
}

// NotImplemented a placeholder for future implementation
func (h *AuthHandler) NotImplemented(c *gin.Context) {
	response.Error(c, http.StatusNotImplemented, "此功能尚未实现", "not implemented")
}

// ForgotPassword 忘记密码，发送密码重置邮件
// @Summary 发送密码重置邮件
// @Description 用户忘记密码时，通过邮箱或手机号发送密码重置邮件
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param ForgotPasswordRequest body vo.ForgotPasswordRequest true "忘记密码请求"
// @Success 200 {object} vo.PasswordResetResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 500 {object} response.ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/password/forgot [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req controllerVo.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 调用密码重置服务
	err := h.passwordService.RequestPasswordReset(c.Request.Context(), req.Identifier)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "发送重置密码邮件失败", err.Error())
		return
	}

	response.Success(c, controllerVo.PasswordResetResponse{
		Message: "密码重置邮件已发送",
		Sent:    true,
	})
}

// ResetPassword 重置密码，使用token验证
// @Summary 重置用户密码
// @Description 使用密码重置token重置用户密码，需要验证两次密码输入一致
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param ResetPasswordRequest body vo.ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} vo.ResetPasswordResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误或密码验证失败"
// @Router /api/v1/auth/password/reset [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req controllerVo.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 验证新密码
	if req.NewPassword != req.ConfirmPassword {
		response.Error(c, http.StatusBadRequest, "密码验证失败", "两次输入的密码不一致")
		return
	}

	// 调用密码重置服务
	err := h.passwordService.ResetPasswordWithToken(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "重置密码失败", err.Error())
		return
	}

	response.Success(c, controllerVo.ResetPasswordResponse{
		Success: true,
		Message: "密码重置成功",
	})
}

// RevokeSession 撤销单个会话
// @Summary 撤销指定的会话
// @Description 用户主动撤销指定的会话，使其立即失效
// @Tags 会话管理
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "会话ID"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/auth/sessions/{sessionId}/revoke [delete]
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	// 从URL参数获取会话ID
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		response.Error(c, http.StatusBadRequest, "请求参数错误", "sessionId不能为空")
		return
	}

	// 调用会话管理服务
	err := h.sessionService.Logout(c.Request.Context(), sessionID, false, "用户主动撤销会话")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "撤销会话失败", err.Error())
		return
	}

	response.Success(c, controllerVo.RevokeSessionResponse{
		Message:   "会话已撤销",
		SessionID: sessionID,
	})
}

// RevokeAllSessions 撤销所有会话
// @Summary 撤销用户所有会话
// @Description 用户主动撤销所有会话，使所有会话立即失效
// @Tags 会话管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/auth/sessions/revoke-all [delete]
func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	// 获取用户信息
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}

	// 调用会话管理服务撤销所有会话
	err := h.sessionService.LogoutAllByUser(c.Request.Context(), userID, "用户主动撤销所有会话")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "撤销所有会话失败", err.Error())
		return
	}

	response.Success(c, controllerVo.RevokeAllSessionsResponse{
		Message: "所有会话已撤销",
		UserID:  userID,
	})
}

// GetDevices 获取设备列表
// @Summary 获取用户的设备列表
// @Description 获取用户所有登录设备列表，包括设备信息、信任状态等，支持分页查询
// @Tags 设备管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认20，最大100"
// @Success 200 {object} response.SuccessResponse "成功响应，返回设备列表及分页信息"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/auth/devices [get]
func (h *AuthHandler) GetDevices(c *gin.Context) {
	// 获取用户信息
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}

	// 解析查询参数
	var req controllerVo.GetDevicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// 调用设备服务获取设备列表
	devicesDTO, err := h.deviceService.GetDevices(c.Request.Context(), userID, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取设备列表失败", err.Error())
		return
	}

	// 转换为响应VO
	responseData := &controllerVo.DevicesListResponse{}
	responseData.FromDevicesListDTO(devicesDTO)
	response.Success(c, responseData)
}

// RevokeDevice 撤销设备
// @Summary 撤销指定设备
// @Description 撤销指定设备的所有会话，使其立即失效
// @Tags 设备管理
// @Produce json
// @Security BearerAuth
// @Param deviceId path string true "设备ID（设备指纹）"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/auth/devices/{deviceId}/revoke [delete]
func (h *AuthHandler) RevokeDevice(c *gin.Context) {
	// 从URL参数获取设备ID
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		response.Error(c, http.StatusBadRequest, "请求参数错误", "deviceId不能为空")
		return
	}

	// 获取用户信息
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}

	// 获取用户的所有会话
	params := param.NewGetSessionsParams(userID, 1, 1000) // 获取足够多的会话
	sessions, _, err := h.sessionService.GetUserSessions(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取会话列表失败", err.Error())
		return
	}

	// 查找并撤销该设备的所有会话
	revokedCount := 0
	for _, session := range sessions {
		deviceInfo := session.GetDeviceInfo()
		if deviceInfo != nil && deviceInfo.Fingerprint == deviceID {
			// 撤销这个会话
			err := h.sessionService.Logout(c.Request.Context(), session.ID, false, "用户撤销设备")
			if err == nil {
				revokedCount++
			}
		}
	}

	response.Success(c, controllerVo.RevokeDeviceResponse{
		Message:      "设备已撤销",
		DeviceID:     deviceID,
		RevokedCount: revokedCount,
	})
}

// TrustDevice 信任设备
// @Summary 标记设备为信任
// @Description 将指定设备标记为信任设备，用于安全策略管理
// @Tags 设备管理
// @Produce json
// @Security BearerAuth
// @Param deviceId path string true "设备ID（设备指纹）"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Failure 404 {object} response.ErrorResponse "设备不存在"
// @Router /api/v1/auth/devices/{deviceId}/trust [post]
func (h *AuthHandler) TrustDevice(c *gin.Context) {
	// 从URL参数获取设备ID
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		response.Error(c, http.StatusBadRequest, "请求参数错误", "deviceId不能为空")
		return
	}

	// 获取用户信息
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}

	// 调用设备服务信任设备
	err := h.deviceService.TrustDevice(c.Request.Context(), userID, deviceID, "用户主动信任")
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "保存设备信任状态失败", err.Error())
		return
	}

	response.Success(c, controllerVo.TrustDeviceResponse{
		Message:  "设备已标记为信任",
		DeviceID: deviceID,
		Trusted:  true,
	})
}

// UnTrustDevice 取消信任设备
// @Summary 取消设备信任标记
// @Description 取消指定设备的信任状态，用于安全策略管理
// @Tags 设备管理
// @Produce json
// @Security BearerAuth
// @Param deviceId path string true "设备ID（设备指纹）"
// @Success 200 {object} response.SuccessResponse "成功响应"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Failure 404 {object} response.ErrorResponse "设备不存在"
// @Router /api/v1/auth/devices/{deviceId}/untrust [delete]
func (h *AuthHandler) UnTrustDevice(c *gin.Context) {
	// 从URL参数获取设备ID
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		response.Error(c, http.StatusBadRequest, "请求参数错误", "deviceId不能为空")
		return
	}

	// 获取用户信息
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}

	// 调用设备服务取消设备信任
	err := h.deviceService.UnTrustDevice(c.Request.Context(), userID, deviceID, "用户主动取消信任")
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "取消设备信任状态失败", err.Error())
		return
	}

	response.Success(c, controllerVo.UnTrustDeviceResponse{
		Message:  "设备信任已取消",
		DeviceID: deviceID,
		Trusted:  false,
	})
}

// GetSecurityLogs 获取安全日志
// @Summary 获取安全审计日志
// @Description 获取用户的安全相关日志，支持按操作类型和时间范围过滤
// @Tags 安全审计
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "开始日期 (RFC3339格式，如: 2023-01-01T00:00:00Z)"
// @Param end_date query string false "结束日期 (RFC3339格式，如: 2023-01-31T23:59:59Z)"
// @Param action query string false "操作类型过滤 (如: login, logout, password_reset)"
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认20，最大100"
// @Success 200 {object} response.SuccessResponse "成功响应，返回安全日志列表及分页信息"
// @Failure 400 {object} response.ErrorResponse "请求参数错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Router /api/v1/auth/security/logs [get]
func (h *AuthHandler) GetSecurityLogs(c *gin.Context) {
	// 获取用户信息
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}

	// 解析查询参数
	var req controllerVo.GetSecurityLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// 调用安全日志服务获取安全日志
	securityLogsDTO, err := h.securityLogService.GetSecurityLogs(
		c.Request.Context(),
		userID,
		req.Action,
		req.StartDate,
		req.EndDate,
		req.Search,
		req.Page,
		req.PageSize,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取安全日志失败", err.Error())
		return
	}

	// 转换为响应VO
	responseData := &controllerVo.SecurityLogsListResponse{}
	responseData.FromSecurityLogsListDTO(securityLogsDTO)
	response.Success(c, responseData)
}
