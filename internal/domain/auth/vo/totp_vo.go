package vo

import "time"

// TOTPSetupInitRequest 初始化2FA设置请求
type TOTPSetupInitRequest struct {
	AccountID string `json:"account_id" binding:"required"`
}

// Validate 验证请求参数
func (r *TOTPSetupInitRequest) Validate() error {
	if r.AccountID == "" {
		return NewValidationError("account_id is required")
	}
	return nil
}

// TOTPSetupInitResponse 初始化2FA设置响应
type TOTPSetupInitResponse struct {
	Secret      string   `json:"secret"`
	QRCode      string   `json:"qr_code"`
	URI         string   `json:"uri"`
	BackupCodes []string `json:"backup_codes"`
}

// TOTPSetupVerifyRequest 验证2FA设置请求
type TOTPSetupVerifyRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Secret    string `json:"secret" binding:"required"`
	Token     string `json:"token" binding:"required"`
}

// Validate 验证请求参数
func (r *TOTPSetupVerifyRequest) Validate() error {
	if r.AccountID == "" {
		return NewValidationError("account_id is required")
	}
	if r.Secret == "" {
		return NewValidationError("secret is required")
	}
	if r.Token == "" {
		return NewValidationError("token is required")
	}
	return nil
}

// TOTPSetupVerifyResponse 验证2FA设置响应
type TOTPSetupVerifyResponse struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message"`
	BackupCodes []string `json:"backup_codes,omitempty"`
}

// TOTPEnableRequest 启用2FA请求
type TOTPEnableRequest struct {
	AccountID   string   `json:"account_id" binding:"required"`
	Secret      string   `json:"secret" binding:"required"`
	Token       string   `json:"token" binding:"required"`
	BackupCodes []string `json:"backup_codes" binding:"required"`
}

// Validate 验证请求参数
func (r *TOTPEnableRequest) Validate() error {
	if r.AccountID == "" {
		return NewValidationError("account_id is required")
	}
	if r.Secret == "" {
		return NewValidationError("secret is required")
	}
	if r.Token == "" {
		return NewValidationError("token is required")
	}
	if len(r.BackupCodes) == 0 {
		return NewValidationError("backup_codes is required")
	}
	return nil
}

// TOTPEnableResponse 启用2FA响应
type TOTPEnableResponse struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message"`
	BackupCodes []string `json:"backup_codes,omitempty"`
}

// TOTPDisableRequest 禁用2FA请求
type TOTPDisableRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Token     string `json:"token" binding:"required"`
}

// Validate 验证请求参数
func (r *TOTPDisableRequest) Validate() error {
	if r.AccountID == "" {
		return NewValidationError("account_id is required")
	}
	if r.Password == "" {
		return NewValidationError("password is required")
	}
	if r.Token == "" {
		return NewValidationError("token is required")
	}
	return nil
}

// TOTPDisableResponse 禁用2FA响应
type TOTPDisableResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TOTPVerifyRequest 验证2FA请求
type TOTPVerifyRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Token     string `json:"token" binding:"required"`
	UseBackup bool   `json:"use_backup"`
}

// Validate 验证请求参数
func (r *TOTPVerifyRequest) Validate() error {
	if r.AccountID == "" {
		return NewValidationError("account_id is required")
	}
	if r.Token == "" {
		return NewValidationError("token is required")
	}
	return nil
}

// TOTPVerifyResponse 验证2FA响应
type TOTPVerifyResponse struct {
	Valid             bool   `json:"valid"`
	Message           string `json:"message"`
	RemainingAttempts int    `json:"remaining_attempts"`
	RemainingCodes    int    `json:"remaining_codes,omitempty"`
}

// TOTPBackupCodeRequest 使用备份码请求
type TOTPBackupCodeRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

// Validate 验证请求参数
func (r *TOTPBackupCodeRequest) Validate() error {
	if r.AccountID == "" {
		return NewValidationError("account_id is required")
	}
	if r.Code == "" {
		return NewValidationError("code is required")
	}
	return nil
}

// TOTPBackupCodeResponse 使用备份码响应
type TOTPBackupCodeResponse struct {
	Valid          bool   `json:"valid"`
	Message        string `json:"message"`
	Success        bool   `json:"success"`
	RemainingCodes int    `json:"remaining_codes"`
}

// TOTPStatusRequest 获取2FA状态请求
type TOTPStatusRequest struct {
	AccountID string `json:"account_id" binding:"required"`
}

// Validate 验证请求参数
func (r *TOTPStatusRequest) Validate() error {
	if r.AccountID == "" {
		return NewValidationError("account_id is required")
	}
	return nil
}

// TOTPStatusResponse 获取2FA状态响应
type TOTPStatusResponse struct {
	Enabled        bool       `json:"enabled"`
	HasBackupCodes bool       `json:"has_backup_codes"`
	RemainingCodes int        `json:"remaining_codes"`
	SetupAt        *time.Time `json:"setup_at,omitempty"`
}

// TOTPChangeSecretRequest 更换2FA密钥请求
type TOTPChangeSecretRequest struct {
	AccountID   string   `json:"account_id" binding:"required"`
	Token       string   `json:"token" binding:"required"`
	NewSecret   string   `json:"new_secret" binding:"required"`
	BackupCodes []string `json:"backup_codes" binding:"required"`
}

// Validate 验证请求参数
func (r *TOTPChangeSecretRequest) Validate() error {
	if r.AccountID == "" {
		return NewValidationError("account_id is required")
	}
	if r.Token == "" {
		return NewValidationError("token is required")
	}
	if r.NewSecret == "" {
		return NewValidationError("new_secret is required")
	}
	if len(r.BackupCodes) == 0 {
		return NewValidationError("backup_codes is required")
	}
	return nil
}

// TOTPChangeSecretResponse 更换2FA密钥响应
type TOTPChangeSecretResponse struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message"`
	BackupCodes []string `json:"backup_codes"`
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError 创建验证错误
func NewValidationError(message string) error {
	return &ValidationError{
		Message: message,
	}
}
