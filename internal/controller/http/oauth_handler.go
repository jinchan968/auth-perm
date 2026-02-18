package http

import (
	"auth-perm/internal/common/constant"
	"auth-perm/internal/common/dto/response"
	controllerVo "auth-perm/internal/controller/vo"
	"auth-perm/internal/domain/auth/param"
	"auth-perm/internal/domain/auth/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// OAuthHandler OAuth处理器
type OAuthHandler struct {
	authService  *service.AuthService
	oauthService *service.OAuthService
}

// NewOAuthHandler 创建OAuth处理器
func NewOAuthHandler(
	authService *service.AuthService,
	oauthService *service.OAuthService,
) *OAuthHandler {
	return &OAuthHandler{
		authService:  authService,
		oauthService: oauthService,
	}
}

// GitHubCallback GitHub OAuth回调
func (h *OAuthHandler) GitHubCallback(c *gin.Context) {
	var req controllerVo.OAuthCallbackRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 构建回调参数
	callbackParams := param.NewGitHubOAuthCallbackParams(req.Code, ipAddress, userAgent)

	// 处理OAuth回调
	user, account, session, err := h.oauthService.GitHubOAuthCallback(c.Request.Context(), callbackParams)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "OAuth认证失败", err.Error())
		return
	}

	// 生成token
	token := session.TokenHash + ":" + session.ID

	// 设置Cookie
	c.SetCookie(
		constant.CookieAuthToken,
		token,
		int(constant.TokenExpiryDefault.Seconds()),
		constant.CookiePath,
		constant.CookieDomain,
		constant.CookieSecure,
		constant.CookieHTTPOnly,
	)

	// 返回成功响应
	loginResp := &controllerVo.LoginResponse{}
	loginResp.FromUserDTO(user, account, token, "登录成功", session.GetExpiresAt())
	response.Success(c, loginResp)
}

// GoogleCallback Google OAuth回调
func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	var req controllerVo.OAuthCallbackRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 构建回调参数
	callbackParams := param.NewGoogleOAuthCallbackParams(req.Code, ipAddress, userAgent)

	// 处理OAuth回调
	user, account, session, err := h.oauthService.GoogleOAuthCallback(c.Request.Context(), callbackParams)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "OAuth认证失败", err.Error())
		return
	}

	// 生成token
	token := session.TokenHash + ":" + session.ID

	// 设置Cookie
	c.SetCookie(
		constant.CookieAuthToken,
		token,
		int(constant.TokenExpiryDefault.Seconds()),
		constant.CookiePath,
		constant.CookieDomain,
		constant.CookieSecure,
		constant.CookieHTTPOnly,
	)

	// 返回成功响应
	loginResp := &controllerVo.LoginResponse{}
	loginResp.FromUserDTO(user, account, token, "登录成功", session.GetExpiresAt())
	response.Success(c, loginResp)
}

// WeChatCallback 微信OAuth回调
func (h *OAuthHandler) WeChatCallback(c *gin.Context) {
	var req controllerVo.OAuthCallbackRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 构建回调参数
	callbackParams := param.NewWeChatOAuthCallbackParams(req.Code, ipAddress, userAgent)

	// 处理OAuth回调
	user, account, session, err := h.oauthService.WeChatOAuthCallback(c.Request.Context(), callbackParams)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "OAuth认证失败", err.Error())
		return
	}

	// 生成token
	token := session.TokenHash + ":" + session.ID

	// 设置Cookie
	c.SetCookie(
		constant.CookieAuthToken,
		token,
		int(constant.TokenExpiryDefault.Seconds()),
		constant.CookiePath,
		constant.CookieDomain,
		constant.CookieSecure,
		constant.CookieHTTPOnly,
	)

	// 返回成功响应
	loginResp := &controllerVo.LoginResponse{}
	loginResp.FromUserDTO(user, account, token, "登录成功", session.GetExpiresAt())
	response.Success(c, loginResp)
}
