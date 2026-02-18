package param

import (
	"context"

	authConstant "auth-perm/internal/domain/auth/constant"
)

// UserCreationParams 用户和账户创建参数
type UserCreationParams struct {
	Context        context.Context
	Username       string
	IdentifierType authConstant.IdentifierType
	Identifier     string
	TenantID       string
	AccountType    authConstant.AccountType
	Password       string
}

// NewUserCreationParams 创建用户和账户参数
func NewUserCreationParams(
	ctx context.Context,
	username string,
	identifierType authConstant.IdentifierType,
	identifier string,
	tenantID string,
	accountType authConstant.AccountType,
	password string,
) *UserCreationParams {
	return &UserCreationParams{
		Context:        ctx,
		Username:       username,
		IdentifierType: identifierType,
		Identifier:     identifier,
		TenantID:       tenantID,
		AccountType:    accountType,
		Password:       password,
	}
}
