package param

import (
	"auth-perm/internal/common/errors"
)

// UpdateProfileParams 更新用户资料参数
type UpdateProfileParams struct {
	UserID   string
	Nickname string
	Phone    string
	Avatar   string
}

// NewUpdateProfileParams 创建更新用户资料参数
func NewUpdateProfileParams(userID, nickname, phone, avatar string) *UpdateProfileParams {
	return &UpdateProfileParams{
		UserID:   userID,
		Nickname: nickname,
		Phone:    phone,
		Avatar:   avatar,
	}
}

// Validate 验证更新用户资料参数
func (p *UpdateProfileParams) Validate() error {
	if p.UserID == "" {
		return errors.NewValidationError("用户ID不能为空")
	}

	return nil
}
