package validator

import (
	"auth-perm/internal/common/errors"
	authConstant "auth-perm/internal/domain/auth/constant"
)

// PasswordValidator 密码校验器
type PasswordValidator struct{}

// NewPasswordValidator 创建密码校验器
func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{}
}

// ValidatePasswordStrength 验证密码强度
func (v *PasswordValidator) ValidatePasswordStrength(password string) error {
	if len(password) < authConstant.PasswordMinLength {
		return errors.NewValidationErrorF("密码长度不能少于%d个字符", authConstant.PasswordMinLength)
	}
	// TODO: 添加更复杂的密码强度检查
	return nil
}

// ValidatePasswordMatch 验证密码匹配
func (v *PasswordValidator) ValidatePasswordMatch(password, confirmPassword string) error {
	if password != confirmPassword {
		return errors.NewValidationError("两次输入的密码不一致")
	}
	return nil
}
