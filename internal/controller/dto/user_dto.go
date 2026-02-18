package dto

import (
	"regexp"
	"strings"

	"auth-perm/internal/common/constant"

	"github.com/go-playground/validator/v10"
)

// IdentifierType 用户标识类型
type IdentifierType string

const (
	IdentifierTypePhone IdentifierType = "phone"
	IdentifierTypeEmail IdentifierType = "email"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username        string         `json:"username" binding:"required,min=3,max=50,alphanumhyphen"`
	IdentifierType  IdentifierType `json:"identifier_type" binding:"required,oneof=phone email"`
	IdentifierValue string         `json:"identifier_value" binding:"required"`
}

// UpdateUserIdentifierRequest 更新用户标识请求
type UpdateUserIdentifierRequest struct {
	IdentifierType  IdentifierType `json:"identifier_type" binding:"required,oneof=phone email"`
	IdentifierValue string         `json:"identifier_value" binding:"required"`
}

// RegisterUserRequest 注册用户请求（包含密码）
type RegisterUserRequest struct {
	Username        string         `json:"username" binding:"required,min=3,max=50,alphanumhyphen"`
	Password        string         `json:"password" binding:"required,min=6"`
	IdentifierType  IdentifierType `json:"identifier_type" binding:"required,oneof=phone email"`
	IdentifierValue string         `json:"identifier_value" binding:"required"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	IdentifierValue string `json:"identifier_value" binding:"required"`
	Password        string `json:"password" binding:"required"`
	IdentifierType  string `json:"identifier_type"` // 可选，如果不提供会自动识别
}

// 自定义验证器
var validate = validator.New()

// init 注册自定义验证函数
func init() {
	// 注册手机号验证
	_ = validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := strings.TrimSpace(fl.Field().String())
		return regexp.MustCompile(constant.RegexPhone).MatchString(phone)
	})

	// 注册邮箱验证
	_ = validate.RegisterValidation("email", func(fl validator.FieldLevel) bool {
		email := strings.TrimSpace(fl.Field().String())
		return regexp.MustCompile(constant.RegexEmail).MatchString(email)
	})

	// 注册用户名验证（字母、数字、下划线、连字符）
	_ = validate.RegisterValidation("alphanumhyphen", func(fl validator.FieldLevel) bool {
		value := strings.TrimSpace(fl.Field().String())
		return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(value)
	})

	// 注册用户名长度验证（使用常量）
	_ = validate.RegisterValidation("username_length", func(fl validator.FieldLevel) bool {
		value := strings.TrimSpace(fl.Field().String())
		return len(value) >= constant.MinUsernameLength && len(value) <= constant.MaxUsernameLength
	})

	// 注册昵称长度验证（使用常量）
	_ = validate.RegisterValidation("nickname_length", func(fl validator.FieldLevel) bool {
		value := strings.TrimSpace(fl.Field().String())
		return len(value) <= constant.MaxNicknameLength
	})

	// 注册密码长度验证（使用常量）
	_ = validate.RegisterValidation("password_length", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return len(value) >= constant.MinPasswordLength && len(value) <= constant.MaxPasswordLength
	})
}

// Validate 验证CreateUserRequest
func (r *CreateUserRequest) Validate() error {
	// 先进行标准验证
	if err := validate.Struct(r); err != nil {
		return err
	}

	// 根据identifier_type验证对应的格式
	switch r.IdentifierType {
	case IdentifierTypePhone:
		if !regexp.MustCompile(constant.RegexPhone).MatchString(strings.TrimSpace(r.IdentifierValue)) {
			return &ValidationError{
				Field:   "identifier_value",
				Message: "invalid phone number format, must be 11-digit Chinese mobile number",
			}
		}
	case IdentifierTypeEmail:
		if !regexp.MustCompile(constant.RegexEmail).MatchString(strings.TrimSpace(r.IdentifierValue)) {
			return &ValidationError{
				Field:   "identifier_value",
				Message: "invalid email format",
			}
		}
	}

	return nil
}

// Validate 验证UpdateUserIdentifierRequest
func (r *UpdateUserIdentifierRequest) Validate() error {
	// 先进行标准验证
	if err := validate.Struct(r); err != nil {
		return err
	}

	// 根据identifier_type验证对应的格式
	switch r.IdentifierType {
	case IdentifierTypePhone:
		if !regexp.MustCompile(constant.RegexPhone).MatchString(strings.TrimSpace(r.IdentifierValue)) {
			return &ValidationError{
				Field:   "identifier_value",
				Message: "invalid phone number format, must be 11-digit Chinese mobile number",
			}
		}
	case IdentifierTypeEmail:
		if !regexp.MustCompile(constant.RegexEmail).MatchString(strings.TrimSpace(r.IdentifierValue)) {
			return &ValidationError{
				Field:   "identifier_value",
				Message: "invalid email format",
			}
		}
	}

	return nil
}

// Validate 验证RegisterUserRequest
func (r *RegisterUserRequest) Validate() error {
	// 先进行标准验证
	if err := validate.Struct(r); err != nil {
		return err
	}

	// 根据identifier_type验证对应的格式
	switch r.IdentifierType {
	case IdentifierTypePhone:
		if !regexp.MustCompile(constant.RegexPhone).MatchString(strings.TrimSpace(r.IdentifierValue)) {
			return &ValidationError{
				Field:   "identifier_value",
				Message: "invalid phone number format, must be 11-digit Chinese mobile number",
			}
		}
	case IdentifierTypeEmail:
		if !regexp.MustCompile(constant.RegexEmail).MatchString(strings.TrimSpace(r.IdentifierValue)) {
			return &ValidationError{
				Field:   "identifier_value",
				Message: "invalid email format",
			}
		}
	}

	// 验证密码强度
	if !isStrongPassword(r.Password) {
		return &ValidationError{
			Field:   "password",
			Message: "password must contain both letters and numbers",
		}
	}

	return nil
}

// Validate 验证LoginRequest
func (r *LoginRequest) Validate() error {
	// 先进行标准验证
	if err := validate.Struct(r); err != nil {
		return err
	}

	// 如果提供了identifier_type，验证格式
	if r.IdentifierType != "" {
		switch IdentifierType(r.IdentifierType) {
		case IdentifierTypePhone:
			if !regexp.MustCompile(constant.RegexPhone).MatchString(strings.TrimSpace(r.IdentifierValue)) {
				return &ValidationError{
					Field:   "identifier_value",
					Message: "invalid phone number format, must be 11-digit Chinese mobile number",
				}
			}
		case IdentifierTypeEmail:
			if !regexp.MustCompile(constant.RegexEmail).MatchString(strings.TrimSpace(r.IdentifierValue)) {
				return &ValidationError{
					Field:   "identifier_value",
					Message: "invalid email format",
				}
			}
		default:
			return &ValidationError{
				Field:   "identifier_type",
				Message: "invalid identifier type, must be phone or email",
			}
		}
	}

	return nil
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// isStrongPassword 检查密码强度
func isStrongPassword(password string) bool {
	hasLetter := false
	hasNumber := false

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z':
			hasLetter = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}
	}

	return hasLetter && hasNumber
}

// AutoDetectIdentifierType FUTURE: 标识符类型自动检测 - 在实现标识符检测时使用
func AutoDetectIdentifierType(value string) (IdentifierType, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", &ValidationError{
			Field:   "identifier_value",
			Message: "identifier value cannot be empty",
		}
	}

	// 检查是否为手机号
	if regexp.MustCompile(constant.RegexPhone).MatchString(value) {
		return IdentifierTypePhone, nil
	}

	// 检查是否为邮箱
	if regexp.MustCompile(constant.RegexEmail).MatchString(value) {
		return IdentifierTypeEmail, nil
	}

	// 无法识别
	return "", &ValidationError{
		Field:   "identifier_value",
		Message: "cannot detect identifier type, please specify manually",
	}
}
