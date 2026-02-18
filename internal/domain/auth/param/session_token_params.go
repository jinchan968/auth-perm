package param

import (
	"auth-perm/internal/domain/auth/dto"
	"context"
)

// SessionTokenParams 会话和令牌创建参数
type SessionTokenParams struct {
	Context    context.Context
	User       *dto.UserDTO
	Account    *dto.AccountDTO
	IPAddress  string
	UserAgent  string
	TenantID   string
	RememberMe bool
}

// NewSessionTokenParams 创建会话和令牌参数
func NewSessionTokenParams(
	ctx context.Context,
	user *dto.UserDTO,
	account *dto.AccountDTO,
	ipAddress,
	userAgent string,
	tenantID string,
	rememberMe bool,
) *SessionTokenParams {
	return &SessionTokenParams{
		Context:    ctx,
		User:       user,
		Account:    account,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		TenantID:   tenantID,
		RememberMe: rememberMe,
	}
}
