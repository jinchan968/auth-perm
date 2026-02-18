package param

import (
	"auth-perm/internal/common/errors"
)

// GetUserDataWithAuthCheckParams 获取账户数据参数（带权限检查）
type GetUserDataWithAuthCheckParams struct {
	CurrentAccountID string
	TargetAccountID  string
}

// NewGetUserDataWithAuthCheckParams 创建获取账户数据参数（带权限检查）
func NewGetUserDataWithAuthCheckParams(currentAccountID, targetAccountID string) *GetUserDataWithAuthCheckParams {
	return &GetUserDataWithAuthCheckParams{
		CurrentAccountID: currentAccountID,
		TargetAccountID:  targetAccountID,
	}
}

// Validate 验证获取账户数据参数
func (p *GetUserDataWithAuthCheckParams) Validate() error {
	if p.CurrentAccountID == "" {
		return errors.NewValidationError("当前账户ID不能为空")
	}

	if p.TargetAccountID == "" {
		return errors.NewValidationError("目标账户ID不能为空")
	}

	return nil
}

// IsSelf 检查是否为查看自己的数据
func (p *GetUserDataWithAuthCheckParams) IsSelf() bool {
	return p.CurrentAccountID == p.TargetAccountID
}
