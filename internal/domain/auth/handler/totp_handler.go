package handler

import (
	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/controller/vo"
	"auth-perm/internal/domain/auth/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TOTPHandler TOTP处理器
type TOTPHandler struct {
	totpService *service.TOTPService
	authService *service.AuthService
}

// NewTOTPHandler 创建TOTP处理器
func NewTOTPHandler(
	totpService *service.TOTPService,
	authService *service.AuthService,
) *TOTPHandler {
	return &TOTPHandler{
		totpService: totpService,
		authService: authService,
	}
}

// TOTPSetupInit 初始化TOTP设置
func (h *TOTPHandler) TOTPSetupInit(c *gin.Context) {
	var req vo.TOTPSetupInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 验证账户ID
	if req.AccountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "account_id is required")
		return
	}

	// 初始化TOTP设置
	setupResp, err := h.totpService.SetupTOTP(req.AccountID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "初始化TOTP失败", err.Error())
		return
	}

	// 返回设置信息
	response.Success(c, gin.H{
		"account_id":   req.AccountID,
		"secret":       setupResp.Secret,
		"qr_code":      setupResp.QRCode,
		"uri":          setupResp.URI,
		"backup_codes": setupResp.BackupCodes,
	})
}

// TOTPSetupVerify 验证TOTP设置
func (h *TOTPHandler) TOTPSetupVerify(c *gin.Context) {
	var req vo.TOTPSetupVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 验证必要参数
	if req.AccountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "account_id is required")
		return
	}
	if req.Secret == "" {
		response.Error(c, http.StatusBadRequest, "密钥不能为空", "secret is required")
		return
	}
	if req.Token == "" {
		response.Error(c, http.StatusBadRequest, "令牌不能为空", "token is required")
		return
	}

	// 验证TOTP令牌
	verifyResp, err := h.totpService.VerifyTOTPSetup(req.AccountID, req.Secret, req.Token)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "验证失败", err.Error())
		return
	}

	if !verifyResp.Success {
		response.Error(c, http.StatusBadRequest, "TOTP验证失败", verifyResp.Message)
		return
	}

	// 返回验证结果
	response.Success(c, gin.H{
		"success":      verifyResp.Success,
		"message":      verifyResp.Message,
		"account_id":   req.AccountID,
		"backup_codes": verifyResp.BackupCodes,
	})
}

// TOTPEnable 启用TOTP
func (h *TOTPHandler) TOTPEnable(c *gin.Context) {
	var req vo.TOTPEnableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 验证账户ID
	if req.AccountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "account_id is required")
		return
	}

	// 启用TOTP
	enableResp, err := h.totpService.EnableTOTP(req.AccountID, req.Secret, req.BackupCodes)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "启用TOTP失败", err.Error())
		return
	}

	// 返回成功响应
	response.Success(c, gin.H{
		"success":      enableResp.Success,
		"message":      enableResp.Message,
		"backup_codes": enableResp.BackupCodes,
		"account_id":   req.AccountID,
	})
}

// TOTPDisable 禁用TOTP
func (h *TOTPHandler) TOTPDisable(c *gin.Context) {
	var req vo.TOTPDisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 验证账户ID
	if req.AccountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "account_id is required")
		return
	}

	// 禁用TOTP
	disableResp, err := h.totpService.DisableTOTP(req.AccountID, req.Password)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "禁用TOTP失败", err.Error())
		return
	}

	// 返回成功响应
	response.Success(c, gin.H{
		"success":    disableResp.Success,
		"message":    disableResp.Message,
		"account_id": req.AccountID,
	})
}

// TOTPVerify 验证TOTP
func (h *TOTPHandler) TOTPVerify(c *gin.Context) {
	var req vo.TOTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 验证账户ID
	if req.AccountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "account_id is required")
		return
	}

	// 验证令牌
	if req.Token == "" {
		response.Error(c, http.StatusBadRequest, "令牌不能为空", "token is required")
		return
	}

	// 验证TOTP令牌
	verifyResp, err := h.totpService.VerifyTOTP(req.AccountID, req.Token, false)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "验证失败", err.Error())
		return
	}

	// 返回验证结果
	response.Success(c, gin.H{
		"valid":              verifyResp.Valid,
		"message":            verifyResp.Message,
		"account_id":         req.AccountID,
		"remaining_attempts": verifyResp.RemainingAttempts,
	})
}

// TOTPBackupCode 使用备份码
func (h *TOTPHandler) TOTPBackupCode(c *gin.Context) {
	var req vo.TOTPBackupCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 验证账户ID
	if req.AccountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "account_id is required")
		return
	}

	// 验证备份码
	if req.Code == "" {
		response.Error(c, http.StatusBadRequest, "备份码不能为空", "code is required")
		return
	}

	// 验证备份码（使用useBackup=true）
	verifyResp, err := h.totpService.VerifyTOTP(req.AccountID, req.Code, true)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "验证失败", err.Error())
		return
	}

	// 返回验证结果
	response.Success(c, gin.H{
		"valid":              verifyResp.Valid,
		"message":            verifyResp.Message,
		"account_id":         req.AccountID,
		"remaining_attempts": verifyResp.RemainingAttempts,
		"remaining_codes":    verifyResp.RemainingCodes,
	})
}

// TOTPStatus 获取TOTP状态
func (h *TOTPHandler) TOTPStatus(c *gin.Context) {
	var req vo.TOTPStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 验证账户ID
	if req.AccountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "account_id is required")
		return
	}

	// 获取TOTP状态
	status, err := h.totpService.GetTOTPStatus(req.AccountID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取TOTP状态失败", err.Error())
		return
	}

	// 返回状态信息
	response.Success(c, gin.H{
		"enabled":          status.Enabled,
		"has_backup_codes": status.HasBackupCodes,
		"remaining_codes":  status.RemainingCodes,
		"setup_at":         status.SetupAt,
		"account_id":       req.AccountID,
	})
}

// TOTPChangeSecret 更改TOTP密钥
func (h *TOTPHandler) TOTPChangeSecret(c *gin.Context) {
	var req vo.TOTPChangeSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 验证账户ID
	if req.AccountID == "" {
		response.Error(c, http.StatusBadRequest, "账户ID不能为空", "account_id is required")
		return
	}

	// 更改TOTP密钥
	changeResp, err := h.totpService.ChangeTOTPSecret(req.AccountID, req.Token, req.NewSecret, req.BackupCodes)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更改TOTP密钥失败", err.Error())
		return
	}

	// 返回新密钥信息
	response.Success(c, gin.H{
		"success":      changeResp.Success,
		"message":      changeResp.Message,
		"backup_codes": changeResp.BackupCodes,
		"account_id":   req.AccountID,
	})
}
