package param

import (
	"auth-perm/internal/common/errors"
)

// GitHubOAuthCallbackParams GitHub OAuth回调参数
type GitHubOAuthCallbackParams struct {
	Code      string
	IPAddress string
	UserAgent string
}

// NewGitHubOAuthCallbackParams 创建GitHub OAuth回调参数
func NewGitHubOAuthCallbackParams(code, ipAddress, userAgent string) *GitHubOAuthCallbackParams {
	return &GitHubOAuthCallbackParams{
		Code:      code,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

// Validate 验证GitHub OAuth回调参数
func (p *GitHubOAuthCallbackParams) Validate() error {
	if p.Code == "" {
		return errors.NewValidationError("OAuth code不能为空")
	}
	return nil
}

// GoogleOAuthCallbackParams Google OAuth回调参数
type GoogleOAuthCallbackParams struct {
	Code      string
	IPAddress string
	UserAgent string
}

// NewGoogleOAuthCallbackParams 创建Google OAuth回调参数
func NewGoogleOAuthCallbackParams(code, ipAddress, userAgent string) *GoogleOAuthCallbackParams {
	return &GoogleOAuthCallbackParams{
		Code:      code,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

// Validate 验证Google OAuth回调参数
func (p *GoogleOAuthCallbackParams) Validate() error {
	if p.Code == "" {
		return errors.NewValidationError("OAuth code不能为空")
	}
	return nil
}

// WeChatOAuthCallbackParams 微信OAuth回调参数
type WeChatOAuthCallbackParams struct {
	Code      string
	IPAddress string
	UserAgent string
}

// NewWeChatOAuthCallbackParams 创建微信OAuth回调参数
func NewWeChatOAuthCallbackParams(code, ipAddress, userAgent string) *WeChatOAuthCallbackParams {
	return &WeChatOAuthCallbackParams{
		Code:      code,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

// Validate 验证微信OAuth回调参数
func (p *WeChatOAuthCallbackParams) Validate() error {
	if p.Code == "" {
		return errors.NewValidationError("OAuth code不能为空")
	}
	return nil
}
