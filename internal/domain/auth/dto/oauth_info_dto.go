package dto

import "time"

// OAuthUserInfoDTO OAuth用户信息
type OAuthUserInfoDTO struct {
	ID           string                 `json:"id"`
	Email        string                 `json:"email"`
	Name         string                 `json:"name"`
	Username     string                 `json:"username"`
	Avatar       string                 `json:"avatar"`
	AvatarURL    string                 `json:"avatar_url"`
	Provider     string                 `json:"provider"`
	ProviderID   string                 `json:"provider_id"`
	AccessToken  string                 `json:"access_token"`
	RefreshToken string                 `json:"refresh_token"`
	ExpiresAt    time.Time              `json:"expires_at"`
	RawData      map[string]interface{} `json:"raw_data"`
}
