package handler

import (
	"fmt"
	"time"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/controller/vo"
	"auth-perm/internal/domain/auth/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// EmailHandler 邮箱验证处理器
type EmailHandler struct {
	authService  *service.AuthService
	emailService *service.EmailService
}

// NewEmailHandler 创建邮箱验证处理器
func NewEmailHandler(
	authService *service.AuthService,
	emailService *service.EmailService,
) *EmailHandler {
	return &EmailHandler{
		authService:  authService,
		emailService: emailService,
	}
}

// SendVerificationEmail 发送验证邮件
func (h *EmailHandler) SendVerificationEmail(c *gin.Context) {
	var req vo.EmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 1. 验证邮箱格式
	if err := h.emailService.ValidateEmailFormat(req.Email); err != nil {
		response.Error(c, http.StatusBadRequest, "邮箱格式错误", err.Error())
		return
	}

	// 2. 查找用户
	account, err := h.authService.FindAccountByEmail(c.Request.Context(), req.Email)
	if err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在", err.Error())
		return
	}

	// 3. 检查是否已验证
	if account.IsEmailVerified() {
		response.Success(c, vo.EmailVerificationResponse{
			Message: "邮箱已经验证过了",
			Sent:    false,
		})
		return
	}

	// 4. 生成验证令牌
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	token := h.emailService.GenerateVerificationToken(req.Email, timestamp)

	// 5. 发送验证邮件
	if err := h.emailService.SendVerificationEmail(c.Request.Context(), req.Email, "用户", token); err != nil {
		response.Error(c, http.StatusInternalServerError, "发送验证邮件失败", err.Error())
		return
	}

	// 6. 返回成功响应
	response.Success(c, vo.EmailVerificationResponse{
		Message: "验证邮件已发送，请检查您的邮箱",
		Sent:    true,
	})
}

// VerifyEmail 验证邮箱
func (h *EmailHandler) VerifyEmail(c *gin.Context) {
	var req vo.VerifyEmailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 1. 检查必要参数
	if req.Token == "" {
		response.Error(c, http.StatusBadRequest, "验证令牌不能为空", "token is required")
		return
	}

	// 2. 从查询参数获取email和timestamp
	email := c.Query("email")
	timestamp := c.Query("timestamp")

	if email == "" {
		response.Error(c, http.StatusBadRequest, "邮箱地址不能为空", "email is required")
		return
	}
	if timestamp == "" {
		response.Error(c, http.StatusBadRequest, "时间戳不能为空", "timestamp is required")
		return
	}

	// 3. 验证邮箱格式
	if err := h.emailService.ValidateEmailFormat(email); err != nil {
		response.Error(c, http.StatusBadRequest, "邮箱格式错误", err.Error())
		return
	}

	// 4. 验证令牌
	isValid, err := h.emailService.VerifyToken(email, req.Token, timestamp)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "令牌验证失败", err.Error())
		return
	}
	if !isValid {
		response.Error(c, http.StatusBadRequest, "验证令牌无效或已过期", "invalid or expired token")
		return
	}

	// 5. 查找并更新账户
	ctx := c.Request.Context()
	account, err := h.authService.FindAccountByEmail(ctx, email)
	if err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在", err.Error())
		return
	}
	_ = account // 简化处理，仅用于验证用户存在

	// 6. 更新邮箱验证状态
	if err := h.authService.UpdateEmailVerificationStatus(ctx, email, true); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新邮箱验证状态失败", err.Error())
		return
	}

	// 返回成功响应
	response.Success(c, vo.VerifyEmailResponse{
		Success:  true,
		Message:  "邮箱验证成功",
		Verified: true,
		Email:    email,
	})
}

// ResendVerificationEmail 重新发送验证邮件
func (h *EmailHandler) ResendVerificationEmail(c *gin.Context) {
	var req vo.EmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 1. 验证邮箱格式
	if err := h.emailService.ValidateEmailFormat(req.Email); err != nil {
		response.Error(c, http.StatusBadRequest, "邮箱格式错误", err.Error())
		return
	}

	// 2. 生成验证令牌
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	token := h.emailService.GenerateVerificationToken(req.Email, timestamp)

	// 3. 发送验证邮件
	if err := h.emailService.SendVerificationEmail(c.Request.Context(), req.Email, "用户", token); err != nil {
		response.Error(c, http.StatusInternalServerError, "发送验证邮件失败", err.Error())
		return
	}

	// 4. 返回成功响应
	response.Success(c, vo.EmailVerificationResponse{
		Message: "验证邮件已发送，请检查您的邮箱",
		Sent:    true,
	})
}
