package dto

import (
	"fmt"
	"time"
)

// TOTPSecretStatus TOTP密钥状态
type TOTPSecretStatus string

const (
	TOTPStatusDisabled TOTPSecretStatus = "disabled"
	TOTPStatusEnabled  TOTPSecretStatus = "enabled"
	TOTPStatusPending  TOTPSecretStatus = "pending"
)

// String 转换为字符串
func (s TOTPSecretStatus) String() string {
	return string(s)
}

// IsValid 检查状态是否有效
func (s TOTPSecretStatus) IsValid() bool {
	switch s {
	case TOTPStatusDisabled, TOTPStatusEnabled, TOTPStatusPending:
		return true
	default:
		return false
	}
}

// TOTPSecretDTO TOTP密钥数据传输对象
type TOTPSecretDTO struct {
	ID          string     `json:"id"`
	AccountID   string     `json:"account_id"`
	Secret      string     `json:"secret"`
	Algorithm   string     `json:"algorithm"`
	Digits      int        `json:"digits"`
	Period      int        `json:"period"`
	IsEnabled   bool       `json:"is_enabled"`
	BackupCodes []string   `json:"backup_codes"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// NewTOTPSecretDTO FUTURE: TOTP密钥DTO创建 - 在实现TOTP功能时使用
func NewTOTPSecretDTO(accountID, secret string) *TOTPSecretDTO {
	now := time.Now()
	return &TOTPSecretDTO{
		AccountID:   accountID,
		Secret:      secret,
		Algorithm:   "SHA1",
		Digits:      6,
		Period:      30,
		IsEnabled:   false,
		BackupCodes: []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Enable 启用TOTP
func (t *TOTPSecretDTO) Enable() {
	t.IsEnabled = true
	t.UpdatedAt = time.Now()
}

// Disable 禁用TOTP
func (t *TOTPSecretDTO) Disable() {
	t.IsEnabled = false
	t.UpdatedAt = time.Now()
}

// UpdateSecret 更新TOTP密钥
func (t *TOTPSecretDTO) UpdateSecret(secret string) {
	t.Secret = secret
	t.UpdatedAt = time.Now()
}

// SetBackupCodes 设置备份码
func (t *TOTPSecretDTO) SetBackupCodes(codes []string) {
	t.BackupCodes = codes
	t.UpdatedAt = time.Now()
}

// UseBackupCode 使用备份码
func (t *TOTPSecretDTO) UseBackupCode(code string) bool {
	for i, c := range t.BackupCodes {
		if c == code {
			// 移除已使用的备份码
			t.BackupCodes = append(t.BackupCodes[:i], t.BackupCodes[i+1:]...)
			t.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// HasBackupCodes 检查是否有备份码
func (t *TOTPSecretDTO) HasBackupCodes() bool {
	return len(t.BackupCodes) > 0
}

// IsValid 检查TOTP密钥是否有效
func (t *TOTPSecretDTO) IsValid() bool {
	return t.Secret != "" && t.AccountID != ""
}

// GetURI 获取TOTP URI
func (t *TOTPSecretDTO) GetURI(issuer, accountName string) string {
	// 格式：otpauth://totp/Issuer:Account?secret=SECRET&issuer=Issuer&algorithm=SHA1&digits=6&period=30
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=%s&digits=%d&period=%d",
		issuer, accountName, t.Secret, issuer, t.Algorithm, t.Digits, t.Period)
}

// IsTOTPEnabled 检查TOTP是否启用
func (t *TOTPSecretDTO) IsTOTPEnabled() bool {
	return t.IsEnabled
}

// GetStatus 获取状态
func (t *TOTPSecretDTO) GetStatus() TOTPSecretStatus {
	if !t.IsEnabled {
		return TOTPStatusDisabled
	}
	return TOTPStatusEnabled
}

// TOTPValidationResult TOTP验证结果
type TOTPValidationResult struct {
	Valid         bool   `json:"valid"`
	UsedBackup    bool   `json:"used_backup"`
	Message       string `json:"message"`
	Attempts      int    `json:"attempts"`
	WindowSize    int    `json:"window_size"`
	RemainingCode int    `json:"remaining_code,omitempty"`
}

// NewTOTPValidationResult FUTURE: TOTP验证结果创建 - 在实现TOTP验证时使用
func NewTOTPValidationResult(valid bool, usedBackup bool, message string) *TOTPValidationResult {
	return &TOTPValidationResult{
		Valid:      valid,
		UsedBackup: usedBackup,
		Message:    message,
		Attempts:   0,
		WindowSize: 0,
	}
}

// Success 标记为成功
func (r *TOTPValidationResult) Success(message string) *TOTPValidationResult {
	r.Valid = true
	r.Message = message
	return r
}

// Fail 标记为失败
func (r *TOTPValidationResult) Fail(message string) *TOTPValidationResult {
	r.Valid = false
	r.Message = message
	return r
}

// IsSuccess 检查是否成功
func (r *TOTPValidationResult) IsSuccess() bool {
	return r.Valid
}

// IsFailure 检查是否失败
func (r *TOTPValidationResult) IsFailure() bool {
	return !r.Valid
}
