package handler

import (
	"auth-perm/internal/common/constant"
	"auth-perm/internal/common/dto/response"
	controllerUtil "auth-perm/internal/controller/util"
	controllerVo "auth-perm/internal/controller/vo"
	authConstant "auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/param"
	"auth-perm/internal/domain/auth/service"
	authVo "auth-perm/internal/domain/auth/vo"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService        *service.AuthService
	loginService       *service.LoginService
	registerService    *service.RegisterService
	invitationService  *service.RegistrationInvitationService
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
	invitationService *service.RegistrationInvitationService,
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
		invitationService:  invitationService,
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

func (h *AuthHandler) Login(c *gin.Context) {
	var req controllerVo.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if req.IPAddress == "" {
		req.IPAddress = extractClientIP(c)
	}
	if req.UserAgent == "" {
		req.UserAgent = c.GetHeader("User-Agent")
	}
	loginParams := req.ToLoginParams()
	user, account, err := h.loginService.Login(c.Request.Context(), loginParams)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "认证失败", err.Error())
		return
	}
	sessionParams := param.NewSessionTokenParams(c.Request.Context(), user, account, req.IPAddress, req.UserAgent, req.TenantID, req.RememberMe)
	loginResult, err := h.loginService.CreateSessionAndToken(sessionParams)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建会话失败", err.Error())
		return
	}
	cookieExpiry := constant.GetCookieExpiry(req.RememberMe)
	c.SetCookie(constant.CookieAuthToken, loginResult.Token, cookieExpiry, constant.CookiePath, constant.CookieDomain, constant.CookieSecure, constant.CookieHTTPOnly)
	loginResp := &controllerVo.LoginResponse{}
	loginResp.FromUserDTO(loginResult.User, loginResult.Account, loginResult.Token, "登录成功", loginResult.Session.GetExpiresAt())
	response.Success(c, loginResp)
}

func (h *AuthHandler) Register(c *gin.Context) {
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
	user, account, err := h.registerService.RegisterWithInvitation(c.Request.Context(), registerParams, req.InviteCode)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "注册失败", err.Error())
		return
	}
	identifier, _ := req.GetIdentifier()
	loginParams := param.NewLoginParams(identifier, req.Password, "", c.GetHeader("User-Agent"), extractClientIP(c), account.TenantID, false)
	loginUser, loginAccount, err := h.loginService.Login(c.Request.Context(), loginParams)
	if err != nil {
		userResp := &controllerVo.UserResponse{}
		userResp.FromUserDTO(user)
		accountResp := &controllerVo.AccountResponse{}
		accountResp.FromAccountDTO(account)
		response.Success(c, gin.H{"message": "注册成功，请手动登录", "user": userResp, "account": accountResp})
		return
	}
	sessionParams := param.NewSessionTokenParams(c.Request.Context(), loginUser, loginAccount, extractClientIP(c), c.GetHeader("User-Agent"), account.TenantID, false)
	loginResult, err := h.loginService.CreateSessionAndToken(sessionParams)
	if err != nil {
		userResp := &controllerVo.UserResponse{}
		userResp.FromUserDTO(user)
		accountResp := &controllerVo.AccountResponse{}
		accountResp.FromAccountDTO(account)
		response.Success(c, gin.H{"message": "注册成功，请手动登录", "user": userResp, "account": accountResp})
		return
	}
	cookieExpiry := constant.GetCookieExpiry(false)
	c.SetCookie(constant.CookieAuthToken, loginResult.Token, cookieExpiry, constant.CookiePath, constant.CookieDomain, constant.CookieSecure, constant.CookieHTTPOnly)
	loginResp := &controllerVo.LoginResponse{}
	loginResp.FromUserDTO(loginResult.User, loginResult.Account, loginResult.Token, "注册并登录成功", loginResult.Session.GetExpiresAt())
	response.Success(c, loginResp)
}

func (h *AuthHandler) ListInvitations(c *gin.Context) {
	page, pageSize, _ := controllerUtil.GetPaginationParams(c)
	result, err := h.invitationService.List(c.Request.Context(), c.Query("tenant_id"), c.GetString("tenant_id"), c.Query("status"), page, pageSize, c.GetBool("is_super_admin"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取邀请码列表失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AuthHandler) CreateInvitation(c *gin.Context) {
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	var req authVo.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	baseURL := c.GetHeader("Origin")
	if baseURL == "" {
		baseURL = c.GetHeader("Referer")
	}

	result, err := h.invitationService.Create(c.Request.Context(), req.TenantID, accountID, c.GetString("tenant_id"), c.GetBool("is_super_admin"), req.ExpiresAt, baseURL)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "生成邀请码失败", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AuthHandler) InvalidateInvitation(c *gin.Context) {
	accountID, err := controllerUtil.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	if err := h.invitationService.Invalidate(c.Request.Context(), c.Param("id"), accountID, c.GetString("tenant_id"), c.GetBool("is_super_admin")); err != nil {
		response.Error(c, http.StatusBadRequest, "失效邀请码失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "邀请码已失效"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token, err := c.Cookie(constant.CookieAuthToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未登录", "not logged in")
		return
	}
	if err := h.sessionService.LogoutByToken(c.Request.Context(), token, false, authConstant.ReasonUserLogout); err != nil {
		response.Error(c, http.StatusInternalServerError, "登出失败", err.Error())
		return
	}
	c.SetCookie(constant.CookieAuthToken, "", -1, constant.CookiePath, constant.CookieDomain, constant.CookieSecure, constant.CookieHTTPOnly)
	response.Success(c, gin.H{"message": "登出成功"})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(constant.CookieAuthToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未登录", "not logged in")
		return
	}
	newToken, err := h.loginService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "刷新token失败", err.Error())
		return
	}
	c.SetCookie(constant.CookieAuthToken, newToken, int(constant.TokenExpiryDefault.Seconds()), constant.CookiePath, constant.CookieDomain, constant.CookieSecure, constant.CookieHTTPOnly)
	response.Success(c, gin.H{"token": newToken})
}

func (h *AuthHandler) LogoutAllByTenant(c *gin.Context) {
	var req controllerVo.LogoutAllByTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if err := h.sessionService.LogoutAllByTenant(c.Request.Context(), req.TenantID, authConstant.ReasonAdminForceLogout); err != nil {
		response.Error(c, http.StatusInternalServerError, "登出失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "已登出该租户下的所有会话", "tenant_id": req.TenantID})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未授权", "unauthorized")
		return
	}
	user, err := h.authService.FindUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在", err.Error())
		return
	}
	userResp := &controllerVo.UserResponse{}
	userResp.FromUserDTO(user)
	response.Success(c, userResp)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req controllerVo.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未授权", "unauthorized")
		return
	}
	updatedUser, err := h.authService.UpdateProfile(c.Request.Context(), param.NewUpdateProfileParams(userID, req.Nickname, req.Phone, req.Avatar))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "更新失败", err.Error())
		return
	}
	userResp := &controllerVo.UserResponse{}
	userResp.FromUserDTO(updatedUser)
	response.Success(c, userResp)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req controllerVo.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未授权", "unauthorized")
		return
	}
	if err := h.passwordService.ChangePassword(c.Request.Context(), param.NewChangePasswordParams(userID, "", req.OldPassword, req.NewPassword, req.NewPassword)); err != nil {
		response.Error(c, http.StatusBadRequest, "修改密码失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "密码修改成功"})
}

func (h *AuthHandler) GetSessions(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未授权", "unauthorized")
		return
	}
	page, pageSize, _ := controllerUtil.GetPaginationParams(c)
	sessions, pagination, err := h.sessionService.GetUserSessions(c.Request.Context(), param.NewGetSessionsParams(userID, page, pageSize))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取会话列表失败", err.Error())
		return
	}
	sessionResponses := make([]controllerVo.SessionResponse, len(sessions))
	for i, session := range sessions {
		deviceInfo := session.GetDeviceInfo()
		sessionResponses[i] = controllerVo.SessionResponse{
			SessionID: session.GetID(),
			DeviceInfo: controllerVo.SessionDeviceInfo{
				Platform: deviceInfo.Platform, Browser: deviceInfo.Browser,
				Device: deviceInfo.Device, IPAddress: deviceInfo.IPAddress, UserAgent: deviceInfo.UserAgent,
			},
			IsActive:   session.IsValid(),
			CreatedAt:  session.GetCreatedAt(),
			LastActive: session.GetLastActiveAt(),
			ExpiresAt:  session.GetExpiresAt(),
		}
	}
	response.Success(c, controllerVo.SessionsListResponse{
		PaginatedResponse: controllerVo.PaginatedResponse{Total: int64(len(sessionResponses)), Page: pagination.Page, PageSize: pagination.PageSize},
		Sessions:          sessionResponses,
	})
}

func (h *AuthHandler) ValidateToken(c *gin.Context) {
	token, err := c.Cookie(constant.CookieAuthToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未登录", "not logged in")
		return
	}
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		response.Error(c, http.StatusBadRequest, "无效的token", "invalid token")
		return
	}
	session, err := h.loginService.ValidateSession(c.Request.Context(), parts[0])
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "会话无效", err.Error())
		return
	}
	response.Success(c, gin.H{"valid": true, "session_id": session.ID, "user_id": session.GetUserID(), "tenant_id": session.GetTenantID(), "expires_at": session.GetExpiresAt()})
}

func (h *AuthHandler) NotImplemented(c *gin.Context) {
	response.Error(c, http.StatusNotImplemented, "此功能尚未实现", "not implemented")
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req controllerVo.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if err := h.passwordService.RequestPasswordReset(c.Request.Context(), req.Identifier); err != nil {
		response.Error(c, http.StatusInternalServerError, "发送重置密码邮件失败", err.Error())
		return
	}
	response.Success(c, controllerVo.PasswordResetResponse{Message: "密码重置邮件已发送", Sent: true})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req controllerVo.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		response.Error(c, http.StatusBadRequest, "密码验证失败", "两次输入的密码不一致")
		return
	}
	if err := h.passwordService.ResetPasswordWithToken(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		response.Error(c, http.StatusBadRequest, "重置密码失败", err.Error())
		return
	}
	response.Success(c, controllerVo.ResetPasswordResponse{Success: true, Message: "密码重置成功"})
}

func (h *AuthHandler) RevokeSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		response.Error(c, http.StatusBadRequest, "请求参数错误", "sessionId不能为空")
		return
	}
	if err := h.sessionService.Logout(c.Request.Context(), sessionID, false, authConstant.ReasonUserRevokeSession); err != nil {
		response.Error(c, http.StatusBadRequest, "撤销会话失败", err.Error())
		return
	}
	response.Success(c, controllerVo.RevokeSessionResponse{Message: "会话已撤销", SessionID: sessionID})
}

func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}
	if err := h.sessionService.LogoutAllByUser(c.Request.Context(), userID, authConstant.ReasonUserRevokeAllSessions); err != nil {
		response.Error(c, http.StatusBadRequest, "撤销所有会话失败", err.Error())
		return
	}
	response.Success(c, controllerVo.RevokeAllSessionsResponse{Message: "所有会话已撤销", UserID: userID})
}

func (h *AuthHandler) GetDevices(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}
	var req controllerVo.GetDevicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}
	devicesDTO, err := h.deviceService.GetDevices(c.Request.Context(), userID, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取设备列表失败", err.Error())
		return
	}
	responseData := &controllerVo.DevicesListResponse{}
	responseData.FromDevicesListDTO(devicesDTO)
	response.Success(c, responseData)
}

func (h *AuthHandler) RevokeDevice(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		response.Error(c, http.StatusBadRequest, "请求参数错误", "deviceId不能为空")
		return
	}
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}
	revokedCount, err := h.deviceService.RevokeDevice(c.Request.Context(), userID, deviceID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "撤销设备失败", err.Error())
		return
	}
	response.Success(c, controllerVo.RevokeDeviceResponse{Message: "设备已撤销", DeviceID: deviceID, RevokedCount: revokedCount})
}

func (h *AuthHandler) TrustDevice(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		response.Error(c, http.StatusBadRequest, "请求参数错误", "deviceId不能为空")
		return
	}
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}
	if err := h.deviceService.TrustDevice(c.Request.Context(), userID, deviceID, authConstant.ReasonUserTrustDevice); err != nil {
		response.Error(c, http.StatusInternalServerError, "保存设备信任状态失败", err.Error())
		return
	}
	response.Success(c, controllerVo.TrustDeviceResponse{Message: "设备已标记为信任", DeviceID: deviceID, Trusted: true})
}

func (h *AuthHandler) UnTrustDevice(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		response.Error(c, http.StatusBadRequest, "请求参数错误", "deviceId不能为空")
		return
	}
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}
	if err := h.deviceService.UnTrustDevice(c.Request.Context(), userID, deviceID, authConstant.ReasonUserUntrustDevice); err != nil {
		response.Error(c, http.StatusInternalServerError, "取消设备信任状态失败", err.Error())
		return
	}
	response.Success(c, controllerVo.UnTrustDeviceResponse{Message: "设备信任已取消", DeviceID: deviceID, Trusted: false})
}

func (h *AuthHandler) GetSecurityLogs(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "未认证", "用户未登录")
		return
	}
	var req controllerVo.GetSecurityLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}
	securityLogsDTO, err := h.securityLogService.GetSecurityLogs(c.Request.Context(), userID, req.Action, req.StartDate, req.EndDate, req.Search, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取安全日志失败", err.Error())
		return
	}
	responseData := &controllerVo.SecurityLogsListResponse{}
	responseData.FromSecurityLogsListDTO(securityLogsDTO)
	response.Success(c, responseData)
}

// extractClientIP 从 X-Forwarded-For 头提取客户端真实 IP。
// Cloudflare 等代理会追加多个 IP（逗号分隔），PostgreSQL inet 类型只接受单个 IP。
func extractClientIP(c *gin.Context) string {
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		// 取第一个 IP（最靠近客户端的）
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	return c.ClientIP()
}
