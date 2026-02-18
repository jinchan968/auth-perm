package constant

// URL 常量定义
const (
	// BaseURL 基础URL（可根据环境变量覆盖）
	BaseURL = "http://localhost:8080"

	// VerificationEmailPath 邮箱验证路径
	VerificationEmailPath = "/api/v1/auth/verify-email"

	// ResetPasswordPath 密码重置路径
	ResetPasswordPath = "/api/v1/auth/reset-password"

	// SendVerificationEmailPath 发送验证邮件路径
	SendVerificationEmailPath = "/api/v1/auth/send-verification-email"
)

// FullURL 生成完整URL
func FullURL(path string) string {
	return BaseURL + path
}
