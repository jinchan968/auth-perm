package vo

// SimpleMessageResponse 简单消息响应
type SimpleMessageResponse struct {
	Message string `json:"message"`
}

// RevokeSessionResponse 撤销会话响应
type RevokeSessionResponse struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

// RevokeAllSessionsResponse 撤销所有会话响应
type RevokeAllSessionsResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

// RevokeDeviceResponse 撤销设备响应
type RevokeDeviceResponse struct {
	Message      string `json:"message"`
	DeviceID     string `json:"device_id"`
	RevokedCount int    `json:"revoked_count"`
}

// TrustDeviceResponse 信任设备响应
type TrustDeviceResponse struct {
	Message  string `json:"message"`
	DeviceID string `json:"device_id"`
	Trusted  bool   `json:"trusted"`
}

// UnTrustDeviceResponse 取消信任设备响应
type UnTrustDeviceResponse struct {
	Message  string `json:"message"`
	DeviceID string `json:"device_id"`
	Trusted  bool   `json:"trusted"`
}
