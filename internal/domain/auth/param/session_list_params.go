package param

import (
	commonConstant "auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
)

// GetSessionsParams 获取会话列表参数
type GetSessionsParams struct {
	UserID   string
	Page     int
	PageSize int
}

// NewGetSessionsParams 创建获取会话列表参数
func NewGetSessionsParams(userID string, page, pageSize int) *GetSessionsParams {
	return &GetSessionsParams{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	}
}

// Validate 验证获取会话列表参数
func (p *GetSessionsParams) Validate() error {
	if p.UserID == "" {
		return errors.NewValidationError("用户ID不能为空")
	}

	if p.Page < 1 {
		p.Page = 1
	}

	if p.PageSize < 1 {
		p.PageSize = commonConstant.DefaultPageSize
	}

	if p.PageSize > commonConstant.MaxPageSize {
		p.PageSize = commonConstant.MaxPageSize
	}

	return nil
}

// GetPagination 获取分页信息
func (p *GetSessionsParams) GetPagination() (int, int) {
	return p.Page, p.PageSize
}
