package response

import (
	"auth-perm/internal/common/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// SuccessWithMessage FUTURE: 带消息的成功响应 - 在实现带消息的成功响应时使用
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  message,
		Data: data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpCode int, message string, err string) {
	c.JSON(httpCode, Response{
		Code:  httpCode,
		Msg:   message,
		Error: err,
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message, "")
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message, "")
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message, "")
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message, "")
}

// InternalError 500错误
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message, "")
}

// ErrorFromAppError 基于AppError设置错误响应
// ErrorFromAppError FUTURE: 应用错误响应 - 在实现错误响应时使用
func ErrorFromAppError(c *gin.Context, appErr *errors.AppError) {
	if appErr == nil {
		InternalError(c, "未知错误")
		return
	}

	// 获取HTTP状态码
	httpStatus := errors.GetHTTPStatus(appErr)

	// 设置响应
	c.JSON(httpStatus, Response{
		Code:  httpStatus,
		Msg:   appErr.Message,
		Error: appErr.Details,
	})
}

// ErrorWithMessage 基于错误对象设置错误响应
// ErrorWithMessage FUTURE: 带消息的错误响应 - 在实现带消息的错误响应时使用
func ErrorWithMessage(c *gin.Context, err error, defaultMessage string) {
	if err == nil {
		InternalError(c, defaultMessage)
		return
	}

	// 尝试转换为AppError
	appErr := errors.FromError(err)

	// 如果转换成功，使用AppError的响应方式
	if appErr != nil {
		ErrorFromAppError(c, appErr)
		return
	}

	// 转换失败，使用默认的内部错误响应
	InternalError(c, defaultMessage)
}

// HandleError 处理错误并设置响应
// HandleError FUTURE: 错误处理 - 在实现统一错误处理时使用
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return // 没有错误，不需要处理
	}

	// 自动解析错误并设置响应
	ErrorWithMessage(c, err, "操作失败")
}

// ValidationErrorFromAppError FUTURE: 验证错误响应 - 在实现验证错误响应时使用
func ValidationErrorFromAppError(c *gin.Context, appErr *errors.AppError) {
	if appErr == nil {
		BadRequest(c, "验证失败")
		return
	}

	// 验证错误使用400状态码
	c.JSON(http.StatusBadRequest, Response{
		Code:  http.StatusBadRequest,
		Msg:   appErr.Message,
		Error: appErr.Details,
	})
}

// AuthErrorFromAppError FUTURE: 认证错误响应 - 在实现认证错误响应时使用
func AuthErrorFromAppError(c *gin.Context, appErr *errors.AppError) {
	if appErr == nil {
		Unauthorized(c, "认证失败")
		return
	}

	// 认证错误使用401状态码
	c.JSON(http.StatusUnauthorized, Response{
		Code:  http.StatusUnauthorized,
		Msg:   appErr.Message,
		Error: appErr.Details,
	})
}

// PermissionErrorFromAppError FUTURE: 权限错误响应 - 在实现权限错误响应时使用
func PermissionErrorFromAppError(c *gin.Context, appErr *errors.AppError) {
	if appErr == nil {
		Forbidden(c, "权限不足")
		return
	}

	// 权限错误使用403状态码
	c.JSON(http.StatusForbidden, Response{
		Code:  http.StatusForbidden,
		Msg:   appErr.Message,
		Error: appErr.Details,
	})
}

// NotFoundErrorFromAppError FUTURE: 未找到错误响应 - 在实现未找到错误响应时使用
func NotFoundErrorFromAppError(c *gin.Context, appErr *errors.AppError) {
	if appErr == nil {
		NotFound(c, "资源不存在")
		return
	}

	// 未找到错误使用404状态码
	c.JSON(http.StatusNotFound, Response{
		Code:  http.StatusNotFound,
		Msg:   appErr.Message,
		Error: appErr.Details,
	})
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Code     int         `json:"code"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int64       `json:"total"`
}

// SuccessWithPagination FUTURE: 分页成功响应 - 在实现分页响应时使用
func SuccessWithPagination(c *gin.Context, data interface{}, page, pageSize int, total int64) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Code:     0,
		Message:  "success",
		Data:     data,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}

// ValidationError 验证错误响应
func ValidationError(c *gin.Context, errors map[string]string) {
	Error(c, http.StatusBadRequest, "请求参数验证失败", formatValidationErrors(errors))
}

// formatValidationErrors 格式化验证错误
func formatValidationErrors(errors map[string]string) string {
	var result string
	for field, msg := range errors {
		if result != "" {
			result += "; "
		}
		result += field + ": " + msg
	}
	return result
}
