package vo

import "time"

// CreateInvitationRequest 创建注册邀请码请求
type CreateInvitationRequest struct {
	TenantID  string     `json:"tenant_id"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// InvitationResponse 注册邀请码响应
type InvitationResponse struct {
	ID                     string     `json:"id"`
	CodePreview            string     `json:"code_preview"`
	TenantID               string     `json:"tenant_id"`
	Status                 string     `json:"status"`
	ExpiresAt              time.Time  `json:"expires_at"`
	UsedAt                 *time.Time `json:"used_at,omitempty"`
	UsedByAccountID        string     `json:"used_by_account_id,omitempty"`
	CreatedByAccountID     string     `json:"created_by_account_id"`
	InvalidatedAt          *time.Time `json:"invalidated_at,omitempty"`
	InvalidatedByAccountID string     `json:"invalidated_by_account_id,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// CreateInvitationResponse 创建邀请码响应，明文邀请码仅创建时返回。
type CreateInvitationResponse struct {
	InvitationResponse
	InviteCode string `json:"invite_code"`
	InviteURL  string `json:"invite_url"`
}

// InvitationListResponse 邀请码列表响应
type InvitationListResponse struct {
	Data  []*InvitationResponse `json:"data"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Size  int                   `json:"size"`
}
