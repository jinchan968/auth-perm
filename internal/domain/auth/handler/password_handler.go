package handler

import (
	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/controller/vo"
	"auth-perm/internal/domain/auth/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PasswordHandler 密码重置处理器
type PasswordHandler struct {
	authService     *service.AuthService
	emailService    *service.EmailService
	passwordService *service.PasswordService
}

// NewPasswordHandler 创建密码重置处理器
func NewPasswordHandler(
	authService *service.AuthService,
	emailService *service.EmailService,
	passwordService *service.PasswordService,
) *PasswordHandler {
	return &PasswordHandler{
		authService:     authService,
		emailService:    emailService,
		passwordService: passwordService,
	}
}

// RequestPasswordReset 请求密码重置
func (h *PasswordHandler) RequestPasswordReset(c *gin.Context) {
	var req vo.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 1. 验证邮箱格式
	if err := h.emailService.ValidateEmailFormat(req.Email); err != nil {
		response.Error(c, http.StatusBadRequest, "邮箱格式错误", err.Error())
		return
	}

	// 2. 请求密码重置
	if err := h.passwordService.RequestPasswordReset(c.Request.Context(), req.Email); err != nil {
		response.Error(c, http.StatusInternalServerError, "请求密码重置失败", err.Error())
		return
	}

	// 3. 返回成功响应（出于安全考虑，即使邮箱不存在也返回成功）
	response.Success(c, gin.H{
		"success": true,
		"message": "如果该邮箱已注册，您将收到密码重置邮件",
	})
}

// ResetPassword 重置密码
func (h *PasswordHandler) ResetPassword(c *gin.Context) {
	var req vo.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 1. 验证必要参数
	if req.Token == "" {
		response.Error(c, http.StatusBadRequest, "重置令牌不能为空", "token is required")
		return
	}
	if req.NewPassword == "" {
		response.Error(c, http.StatusBadRequest, "新密码不能为空", "new_password is required")
		return
	}

	// 2. 重置密码
	if err := h.passwordService.ResetPasswordWithToken(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		response.Error(c, http.StatusBadRequest, "密码重置失败", err.Error())
		return
	}

	// 3. 返回成功响应
	response.Success(c, gin.H{
		"success": true,
		"message": "密码重置成功，请使用新密码登录",
	})
}
