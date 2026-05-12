package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeInternal   ErrorType = "INTERNAL"
	ErrorTypeNotFound   ErrorType = "NOT_FOUND"
	ErrorTypeAuth       ErrorType = "AUTH"
	ErrorTypeValidation ErrorType = "VALIDATION"
	ErrorTypePermission ErrorType = "PERMISSION"
	ErrorTypeBusiness   ErrorType = "BUSINESS"
)

func Is(err, target error) bool {
	return errors.Is(err, target)
}

// AppError 应用错误
type AppError struct {
	Type     ErrorType `json:"type"`
	Code     int       `json:"code"`
	Message  string    `json:"message"`
	Details  string    `json:"details,omitempty"`
	Internal error     `json:"-"` // 内部错误，不暴露给客户端
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap 支持错误链
func (e *AppError) Unwrap() error {
	return e.Internal
}

// NewInternalError 创建内部错误
func NewInternalError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Code:    http.StatusInternalServerError,
		Message: message,
	}
}

// NewInternalErrorWithDetails 创建带详情的内部错误
func NewInternalErrorWithDetails(message, details string, internal error) *AppError {
	return &AppError{
		Type:     ErrorTypeInternal,
		Code:     http.StatusInternalServerError,
		Message:  message,
		Details:  details,
		Internal: internal,
	}
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Code:    http.StatusNotFound,
		Message: message,
	}
}

// NewAuthError 创建认证错误
func NewAuthError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeAuth,
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

// NewValidationError 创建验证错误
func NewValidationError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

// NewValidationErrorWithDetails 创建带详情的验证错误
func NewValidationErrorWithDetails(message, details string) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Code:    http.StatusBadRequest,
		Message: message,
		Details: details,
	}
}

// NewPermissionError 创建权限错误
func NewPermissionError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypePermission,
		Code:    http.StatusForbidden,
		Message: message,
	}
}

// NewBusinessError 创建业务错误（默认400状态码）
func NewBusinessError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeBusiness,
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

// NewBusinessErrorWithDetails 创建带详情的业务错误
func NewBusinessErrorWithDetails(message, details string) *AppError {
	return &AppError{
		Type:    ErrorTypeBusiness,
		Code:    http.StatusBadRequest,
		Message: message,
		Details: details,
	}
}

// NewDBError 创建数据库错误
func NewDBError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Code:    http.StatusInternalServerError,
		Message: message,
	}
}

// NewDBErrorWithDetails 创建带详情的数据库错误
func NewDBErrorWithDetails(message, details string, internal error) *AppError {
	return &AppError{
		Type:     ErrorTypeInternal,
		Code:     http.StatusInternalServerError,
		Message:  message,
		Details:  details,
		Internal: internal,
	}
}

// NewErrorF 格式化错误
func NewErrorF(errorType ErrorType, httpCode int, format string, args ...interface{}) *AppError {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}
	return &AppError{
		Type:    errorType,
		Code:    httpCode,
		Message: message,
	}
}

// NewInternalErrorF 格式化创建内部错误
func NewInternalErrorF(format string, args ...interface{}) *AppError {
	return NewErrorF(ErrorTypeInternal, http.StatusInternalServerError, format, args...)
}

// NewNotFoundErrorF 格式化创建未找到错误
func NewNotFoundErrorF(format string, args ...interface{}) *AppError {
	return NewErrorF(ErrorTypeNotFound, http.StatusNotFound, format, args...)
}

// NewAuthErrorF 格式化认证错误
func NewAuthErrorF(format string, args ...interface{}) *AppError {
	return NewErrorF(ErrorTypeAuth, http.StatusUnauthorized, format, args...)
}

// NewValidationErrorF 格式化验证错误
func NewValidationErrorF(format string, args ...interface{}) *AppError {
	return NewErrorF(ErrorTypeValidation, http.StatusBadRequest, format, args...)
}

// NewPermissionErrorF 格式化权限错误
func NewPermissionErrorF(format string, args ...interface{}) *AppError {
	return NewErrorF(ErrorTypePermission, http.StatusForbidden, format, args...)
}

// NewBusinessErrorF 格式化创建业务错误
func NewBusinessErrorF(format string, args ...interface{}) *AppError {
	return NewErrorF(ErrorTypeBusiness, http.StatusBadRequest, format, args...)
}

// IsNotFoundError 检查是否为未找到错误
func IsNotFoundError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeNotFound
	}
	return false
}

// GetHTTPStatus 获取HTTP状态码
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return http.StatusInternalServerError
}

// WrapBizError 包装错误
func WrapBizError(err error, message string) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		// 如果已经是AppError，添加包装信息
		return &AppError{
			Type:     appErr.Type,
			Code:     appErr.Code,
			Message:  message + ": " + appErr.Message,
			Details:  appErr.Details,
			Internal: appErr.Internal,
		}
	}

	// 包装普通错误
	return NewBusinessErrorWithDetails(message, err.Error())
}

// FromError 从标准错误创建 AppError
// 自动检测错误类型并设置合适的 HTTP 状态码
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// 根据错误类型自动判断
	return NewInternalErrorWithDetails("未知错误", err.Error(), err)
}
