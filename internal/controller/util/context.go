package util

import (
	"fmt"
	"net/http"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/common/errors"

	"github.com/gin-gonic/gin"
)

// AuthInfo 认证信息结构体
type AuthInfo struct {
	UserID    string `json:"user_id"`
	AccountID string `json:"account_id"`
	SessionID string `json:"session_id"`
}

// GetAuthInfo 从gin.Context中获取认证信息
// 如果未认证或认证信息不存在，返回错误
func GetAuthInfo(c *gin.Context) (*AuthInfo, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, errors.NewAuthError("未认证，请先登录")
	}

	accountID, exists := c.Get("account_id")
	if !exists {
		return nil, errors.NewAuthError("账户信息不存在")
	}

	sessionID, exists := c.Get("session_id")
	if !exists {
		return nil, errors.NewAuthError("会话信息不存在")
	}

	return &AuthInfo{
		UserID:    userID.(string),
		AccountID: accountID.(string),
		SessionID: sessionID.(string),
	}, nil
}

func GetToken(c *gin.Context) (string, error) {
	token, exists := c.Get("token")
	if !exists {
		return "", errors.NewAuthError("未认证，请先登录")
	}
	return token.(string), nil
}

// GetUserID 获取用户ID
func GetUserID(c *gin.Context) (string, error) {
	authInfo, err := GetAuthInfo(c)
	if err != nil {
		return "", err
	}
	return authInfo.UserID, nil
}

// GetAccountID 获取账户ID
func GetAccountID(c *gin.Context) (string, error) {
	authInfo, err := GetAuthInfo(c)
	if err != nil {
		return "", err
	}
	return authInfo.AccountID, nil
}

// GetSessionID 获取会话ID
func GetSessionID(c *gin.Context) (string, error) {
	authInfo, err := GetAuthInfo(c)
	if err != nil {
		return "", err
	}
	return authInfo.SessionID, nil
}

// GetTenantID 获取租户ID
func GetTenantID(c *gin.Context) (string, error) {
	tenantID, exists := c.Get("tenant_id")
	if !exists || tenantID == "" {
		return "", errors.NewBusinessError("租户ID不存在")
	}
	return tenantID.(string), nil
}

// MustGetUserID 获取用户ID（panic如果未认证）
func MustGetUserID(c *gin.Context) string {
	userID, err := GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		c.Abort()
		panic(err)
	}
	return userID
}

// MustGetAccountID 获取账户ID（panic如果未认证）
func MustGetAccountID(c *gin.Context) string {
	accountID, err := GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		c.Abort()
		panic(err)
	}
	return accountID
}

// MustGetSessionID 获取会话ID（panic如果未认证）
func MustGetSessionID(c *gin.Context) string {
	sessionID, err := GetSessionID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), "")
		c.Abort()
		panic(err)
	}
	return sessionID
}

// GetPaginationParams 获取分页参数
func GetPaginationParams(c *gin.Context) (page, pageSize int, err error) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page = parseInt(pageStr, 1)
	pageSize = parseInt(pageSizeStr, 10)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大页面大小
	}

	return page, pageSize, nil
}

// GetSortParams 获取排序参数
func GetSortParams(c *gin.Context) (sortBy, sortOrder string) {
	sortBy = c.DefaultQuery("sort_by", "created_at")
	sortOrder = c.DefaultQuery("sort_order", "desc")

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return sortBy, sortOrder
}

// GetQueryParams 获取查询参数
func GetQueryParams(c *gin.Context) map[string]string {
	params := make(map[string]string)

	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	return params
}

// parseInt 解析字符串为int，失败时返回默认值
func parseInt(s string, defaultValue int) int {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
