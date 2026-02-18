package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/utils"
	"auth-perm/internal/domain/auth/dto"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// GitHubUser GitHub用户信息
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GoogleUser Google用户信息
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// WeChatTokenResponse 微信token响应
type WeChatTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// WeChatUser 微信用户信息
type WeChatUser struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	Nickname   string `json:"nickname"`
	HeadImgURL string `json:"headimgurl"`
	Sex        int    `json:"sex"`
	Country    string `json:"country"`
	Province   string `json:"province"`
	City       string `json:"city"`
}

// OAuthProvider OAuth提供商接口
type OAuthProvider interface {
	GetUserInfo(ctx context.Context, code string) (*dto.OAuthUserInfoDTO, error)
	ValidateToken(ctx context.Context, token string) (*dto.OAuthUserInfoDTO, error)
}

// GitHubProvider GitHub OAuth提供商
type GitHubProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// NewGitHubProvider FUTURE: GitHub OAuth提供商创建 - 在实现GitHub OAuth时使用
func NewGitHubProvider(clientID, clientSecret, redirectURL string) *GitHubProvider {
	return &GitHubProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

func (p *GitHubProvider) GetUserInfo(ctx context.Context, code string) (*dto.OAuthUserInfoDTO, error) {
	// 实现GitHub OAuth
	// 1. 使用code交换access_token
	// 2. 调用GitHub API获取用户信息
	// 3. 返回标准化用户数据

	// 配置OAuth2
	config := &oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		RedirectURL:  p.RedirectURL,
		Scopes: []string{
			"read:user",
			"user:email",
		},
		Endpoint: github.Endpoint,
	}

	// 交换访问令牌
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.WrapBizError(err, "GitHub OAuth token exchange failed")
	}

	// 创建OAuth客户端
	oauthClient := utils.NewOAuthClient(
		utils.HTTPClientConfig{
			BaseURL:   "https://api.github.com",
			Timeout:   10 * time.Second,
			UserAgent: "auth-perm/1.0",
		},
		token.AccessToken,
	)

	// 获取用户信息
	var githubUser GitHubUser
	if err := oauthClient.GetJSONWithToken(ctx, "/user", nil, &githubUser); err != nil {
		return nil, errors.WrapBizError(err, "Failed to fetch GitHub user info")
	}

	// 获取用户邮箱（如果用户没有公开邮箱）
	var userEmail string
	if githubUser.Email == "" {
		var emailList []struct {
			Email   string `json:"email"`
			Primary bool   `json:"primary"`
		}
		if err := oauthClient.GetJSONWithToken(ctx, "/user/emails", nil, &emailList); err == nil {
			for _, e := range emailList {
				if e.Primary {
					userEmail = e.Email
					break
				}
			}
		}
	} else {
		userEmail = githubUser.Email
	}

	// 返回标准化用户数据
	return &dto.OAuthUserInfoDTO{
		Provider:     "github",
		ProviderID:   fmt.Sprintf("%d", githubUser.ID),
		Email:        userEmail,
		Name:         githubUser.Name,
		Username:     githubUser.Login,
		AvatarURL:    githubUser.AvatarURL,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
	}, nil
}

func (p *GitHubProvider) ValidateToken(ctx context.Context, token string) (*dto.OAuthUserInfoDTO, error) {
	// 实现GitHub token验证
	oauthClient := utils.NewOAuthClient(
		utils.HTTPClientConfig{
			BaseURL:   "https://api.github.com",
			Timeout:   10 * time.Second,
			UserAgent: "auth-perm/1.0",
		},
		token,
	)

	var githubUser GitHubUser
	if err := oauthClient.GetJSONWithToken(ctx, "/user", nil, &githubUser); err != nil {
		return nil, errors.WrapBizError(err, "Failed to validate GitHub token")
	}

	// 返回标准化用户数据
	return &dto.OAuthUserInfoDTO{
		Provider:    "github",
		ProviderID:  fmt.Sprintf("%d", githubUser.ID),
		Email:       githubUser.Email,
		Name:        githubUser.Name,
		Username:    githubUser.Login,
		AvatarURL:   githubUser.AvatarURL,
		AccessToken: token,
	}, nil
}

// GoogleProvider Google OAuth提供商
type GoogleProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// NewGoogleProvider FUTURE: Google OAuth提供商创建 - 在实现Google OAuth时使用
func NewGoogleProvider(clientID, clientSecret, redirectURL string) *GoogleProvider {
	return &GoogleProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

func (p *GoogleProvider) GetUserInfo(ctx context.Context, code string) (*dto.OAuthUserInfoDTO, error) {
	// 实现Google OAuth
	// 1. 使用code交换access_token
	// 2. 调用Google API获取用户信息
	// 3. 返回标准化用户数据

	// 配置OAuth2
	config := &oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		RedirectURL:  p.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// 交换访问令牌
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.WrapBizError(err, "Google OAuth token exchange failed")
	}

	// 创建OAuth客户端
	oauthClient := utils.NewOAuthClient(
		utils.HTTPClientConfig{
			BaseURL:   "https://www.googleapis.com/oauth2/v2",
			Timeout:   10 * time.Second,
			UserAgent: "auth-perm/1.0",
		},
		token.AccessToken,
	)

	// 获取用户信息
	var googleUser GoogleUser
	if err := oauthClient.GetJSONWithToken(ctx, "/userinfo", nil, &googleUser); err != nil {
		return nil, errors.WrapBizError(err, "Failed to fetch Google user info")
	}

	// 检查邮箱是否已验证
	if !googleUser.VerifiedEmail {
		return nil, errors.NewBusinessError("Google email not verified")
	}

	// 返回标准化用户数据
	return &dto.OAuthUserInfoDTO{
		Provider:     "google",
		ProviderID:   googleUser.ID,
		Email:        googleUser.Email,
		Name:         googleUser.Name,
		Username:     googleUser.Email, // Google没有username，使用邮箱
		AvatarURL:    googleUser.Picture,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
	}, nil
}

func (p *GoogleProvider) ValidateToken(ctx context.Context, token string) (*dto.OAuthUserInfoDTO, error) {
	// 实现Google token验证
	oauthClient := utils.NewOAuthClient(
		utils.HTTPClientConfig{
			BaseURL:   "https://www.googleapis.com/oauth2/v2",
			Timeout:   10 * time.Second,
			UserAgent: "auth-perm/1.0",
		},
		token,
	)

	var googleUser GoogleUser
	if err := oauthClient.GetJSONWithToken(ctx, "/userinfo", nil, &googleUser); err != nil {
		return nil, errors.WrapBizError(err, "Failed to validate Google token")
	}

	// 检查邮箱是否已验证
	if !googleUser.VerifiedEmail {
		return nil, errors.NewBusinessError("Google email not verified")
	}

	// 返回标准化用户数据
	return &dto.OAuthUserInfoDTO{
		Provider:    "google",
		ProviderID:  googleUser.ID,
		Email:       googleUser.Email,
		Name:        googleUser.Name,
		Username:    googleUser.Email,
		AvatarURL:   googleUser.Picture,
		AccessToken: token,
	}, nil
}

// WeChatProvider 微信OAuth提供商
type WeChatProvider struct {
	AppID       string
	AppSecret   string
	RedirectURL string
}

// NewWeChatProvider FUTURE: WeChat OAuth提供商创建 - 在实现WeChat OAuth时使用
func NewWeChatProvider(appID, appSecret, redirectURL string) *WeChatProvider {
	return &WeChatProvider{
		AppID:       appID,
		AppSecret:   appSecret,
		RedirectURL: redirectURL,
	}
}

func (p *WeChatProvider) GetUserInfo(ctx context.Context, code string) (*dto.OAuthUserInfoDTO, error) {
	// 创建HTTP客户端
	client := utils.NewHTTPClient(utils.HTTPClientConfig{
		BaseURL:   "https://api.weixin.qq.com/sns",
		Timeout:   10 * time.Second,
		UserAgent: "auth-perm/1.0",
	})

	// 1. 使用code交换access_token
	tokenParams := map[string]string{
		"appid":      p.AppID,
		"secret":     p.AppSecret,
		"code":       code,
		"grant_type": "authorization_code",
	}

	var tokenResp WeChatTokenResponse
	if err := client.GetJSON(ctx, "/oauth2/access_token", tokenParams, &tokenResp); err != nil {
		return nil, errors.WrapBizError(err, "WeChat OAuth token exchange failed")
	}

	if tokenResp.ErrCode != 0 {
		return nil, errors.NewBusinessError(fmt.Sprintf("WeChat OAuth failed: %s", tokenResp.ErrMsg))
	}

	// 2. 使用access_token获取用户信息
	userParams := map[string]string{
		"access_token": tokenResp.AccessToken,
		"openid":       tokenResp.OpenID,
		"lang":         "zh_CN",
	}

	var wechatUser WeChatUser
	if err := client.GetJSON(ctx, "/userinfo", userParams, &wechatUser); err != nil {
		return nil, errors.WrapBizError(err, "Failed to fetch WeChat user info")
	}

	// 3. 返回标准化用户数据（使用虚拟邮箱）
	virtualEmail := fmt.Sprintf("wechat_%s@oauth.local", wechatUser.OpenID)

	return &dto.OAuthUserInfoDTO{
		Provider:     "wechat",
		ProviderID:   wechatUser.OpenID,
		Email:        virtualEmail,
		Name:         wechatUser.Nickname,
		Username:     wechatUser.Nickname,
		AvatarURL:    wechatUser.HeadImgURL,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

func (p *WeChatProvider) ValidateToken(ctx context.Context, token string) (*dto.OAuthUserInfoDTO, error) {
	// 微信token验证需要access_token和openid，格式为 "access_token:openid"
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return nil, errors.NewValidationError("微信token格式无效，需要格式：access_token:openid")
	}

	accessToken := parts[0]
	openID := parts[1]

	// 创建HTTP客户端
	client := utils.NewHTTPClient(utils.HTTPClientConfig{
		BaseURL:   "https://api.weixin.qq.com/sns",
		Timeout:   10 * time.Second,
		UserAgent: "auth-perm/1.0",
	})

	// 1. 验证access_token是否有效
	authParams := map[string]string{
		"access_token": accessToken,
		"openid":       openID,
	}

	var authResult struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := client.GetJSON(ctx, "/auth", authParams, &authResult); err != nil {
		return nil, errors.WrapBizError(err, "Failed to validate WeChat token")
	}

	if authResult.ErrCode != 0 {
		return nil, errors.NewBusinessError(fmt.Sprintf("微信token无效: %s", authResult.ErrMsg))
	}

	// 2. 获取用户信息
	userParams := map[string]string{
		"access_token": accessToken,
		"openid":       openID,
		"lang":         "zh_CN",
	}

	var wechatUser WeChatUser
	if err := client.GetJSON(ctx, "/userinfo", userParams, &wechatUser); err != nil {
		return nil, errors.WrapBizError(err, "Failed to fetch WeChat user info")
	}

	// 3. 返回标准化用户数据
	virtualEmail := fmt.Sprintf("wechat_%s@oauth.local", wechatUser.OpenID)

	return &dto.OAuthUserInfoDTO{
		Provider:    "wechat",
		ProviderID:  wechatUser.OpenID,
		Email:       virtualEmail,
		Name:        wechatUser.Nickname,
		Username:    wechatUser.Nickname,
		AvatarURL:   wechatUser.HeadImgURL,
		AccessToken: accessToken,
	}, nil
}

// OAuthRepo OAuth仓储
type OAuthRepo struct {
	githubProvider *GitHubProvider
	googleProvider *GoogleProvider
	wechatProvider *WeChatProvider
}

// NewOAuthRepo 创建OAuth仓储
func NewOAuthRepo(githubClientID, githubClientSecret, githubRedirectURL,
	googleClientID, googleClientSecret, googleRedirectURL,
	wechatAppID, wechatAppSecret, wechatRedirectURL string) *OAuthRepo {
	return &OAuthRepo{
		githubProvider: NewGitHubProvider(githubClientID, githubClientSecret, githubRedirectURL),
		googleProvider: NewGoogleProvider(googleClientID, googleClientSecret, googleRedirectURL),
		wechatProvider: NewWeChatProvider(wechatAppID, wechatAppSecret, wechatRedirectURL),
	}
}

// GetUserInfo 获取用户信息
func (r *OAuthRepo) GetUserInfo(ctx context.Context, provider, code, redirectURI string) (*dto.OAuthUserInfoDTO, error) {
	switch provider {
	case "github":
		return r.githubProvider.GetUserInfo(ctx, code)
	case "google":
		return r.googleProvider.GetUserInfo(ctx, code)
	case "wechat":
		return r.wechatProvider.GetUserInfo(ctx, code)
	default:
		return nil, errors.NewBusinessError("Unsupported OAuth provider: " + provider)
	}
}

// ValidateToken 验证令牌
func (r *OAuthRepo) ValidateToken(ctx context.Context, provider, token string) (*dto.OAuthUserInfoDTO, error) {
	switch provider {
	case "github":
		return r.githubProvider.ValidateToken(ctx, token)
	case "google":
		return r.googleProvider.ValidateToken(ctx, token)
	case "wechat":
		return r.wechatProvider.ValidateToken(ctx, token)
	default:
		return nil, errors.NewBusinessError("Unsupported OAuth provider: " + provider)
	}
}

// GetUserEmail 获取用户邮箱
func (r *OAuthRepo) GetUserEmail(ctx context.Context, provider, token string) (string, error) {
	userInfo, err := r.ValidateToken(ctx, provider, token)
	if err != nil {
		return "", err
	}
	return userInfo.Email, nil
}
