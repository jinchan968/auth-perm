package param

// TOTPBackupCodeUsageParams TOTP备份码使用记录参数
type TOTPBackupCodeUsageParams struct {
	AccountID string
	Code      string
	IPAddress string
	UserAgent string
	Success   bool
}

// NewTOTPBackupCodeUsageParams 创建TOTP备份码使用记录参数
func NewTOTPBackupCodeUsageParams(accountID, code, ipAddress, userAgent string, success bool) *TOTPBackupCodeUsageParams {
	return &TOTPBackupCodeUsageParams{
		AccountID: accountID,
		Code:      code,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
	}
}
