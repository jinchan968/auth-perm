package dto

import (
	"auth-perm/internal/common/constant"
	"auth-perm/internal/common/errors"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

// SessionDTO 会话数据传输对象（包含业务逻辑）
type SessionDTO struct {
	// 基本信息
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	AccountID string `json:"account_id"`
	TenantID  string `json:"tenant_id"` // 租户ID，支持多租户
	TokenHash string `json:"token_hash"`

	// 设备信息
	DeviceInfo DeviceInfoDTO `json:"device_info"`
	IPAddress  *string       `json:"ip_address"`
	UserAgent  *string       `json:"user_agent"`

	// 时间信息
	ExpiresAt    time.Time `json:"expires_at"`
	LastActiveAt time.Time `json:"last_active_at"`

	// 状态信息
	IsActive bool `json:"is_active"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeviceInfo 设备信息结构
type DeviceInfo struct {
	Platform    string `json:"platform"` // 操作系统
	Browser     string `json:"browser"`  // 浏览器
	Device      string `json:"device"`   // 设备类型
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	Fingerprint string `json:"fingerprint"` // 设备指纹
}

// NewSessionDTO 创建会话DTO
func NewSessionDTO(userID, accountID, tenantID string, expiresAt time.Time) *SessionDTO {
	now := time.Now()
	return &SessionDTO{
		ID:           "",
		UserID:       userID,
		AccountID:    accountID,
		TenantID:     tenantID,
		TokenHash:    "",
		DeviceInfo:   DeviceInfoDTO{},
		ExpiresAt:    expiresAt,
		LastActiveAt: now,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ==================== Getter 方法 ====================

// GetID 获取会话ID
func (s *SessionDTO) GetID() string {
	return s.ID
}

// GetUserID 获取用户ID
func (s *SessionDTO) GetUserID() string {
	return s.UserID
}

// GetAccountID 获取账户ID
func (s *SessionDTO) GetAccountID() string {
	return s.AccountID
}

// GetTokenHash 获取令牌哈希
func (s *SessionDTO) GetTokenHash() string {
	return s.TokenHash
}

// GetIPAddress 获取IP地址
func (s *SessionDTO) GetIPAddress() *string {
	return s.IPAddress
}

// GetUserAgent 获取用户代理
func (s *SessionDTO) GetUserAgent() *string {
	return s.UserAgent
}

// GetDeviceInfo 获取设备信息
func (s *SessionDTO) GetDeviceInfo() *DeviceInfoDTO {
	return &s.DeviceInfo
}

// GetExpiresAt 获取过期时间
func (s *SessionDTO) GetExpiresAt() time.Time {
	return s.ExpiresAt
}

// GetLastActiveAt 获取最后活动时间
func (s *SessionDTO) GetLastActiveAt() time.Time {
	return s.LastActiveAt
}

// GetIsActive 获取是否活跃
func (s *SessionDTO) GetIsActive() bool {
	return s.IsActive
}

// GetCreatedAt 获取创建时间
func (s *SessionDTO) GetCreatedAt() time.Time {
	return s.CreatedAt
}

// GetUpdatedAt 获取更新时间
func (s *SessionDTO) GetUpdatedAt() time.Time {
	return s.UpdatedAt
}

// GetTenantID 获取租户ID
func (s *SessionDTO) GetTenantID() string {
	return s.TenantID
}

// ==================== Setter 方法 ====================

// SetTokenHash 设置令牌哈希
func (s *SessionDTO) SetTokenHash(tokenHash string) {
	s.TokenHash = tokenHash
	s.UpdatedAt = time.Now()
}

// ==================== 业务方法 ====================

// SetDeviceInfo 设置设备信息
func (s *SessionDTO) SetDeviceInfo(ipAddress, userAgent string) error {
	s.IPAddress = &ipAddress
	s.UserAgent = &userAgent

	// 解析设备信息
	deviceInfo := parseDeviceInfo(userAgent, ipAddress)
	s.DeviceInfo = deviceInfo
	s.UpdatedAt = time.Now()
	return nil
}

// UpdateLastActivity 更新最后活动时间
func (s *SessionDTO) UpdateLastActivity() {
	s.LastActiveAt = time.Now()
	s.UpdatedAt = time.Now()
}

// IsExpired 检查会话是否已过期
func (s *SessionDTO) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid 检查会话是否有效
func (s *SessionDTO) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// Invalidate 使会话失效
func (s *SessionDTO) Invalidate(reason string) {
	s.IsActive = false
	s.UpdatedAt = time.Now()

	// 记录失效原因到设备信息中
	// 注意：DeviceInfoDTO结构简单，使用Context存储附加信息
	s.DeviceInfo.Fingerprint = fmt.Sprintf("%s-invalidated-%s", s.DeviceInfo.Fingerprint, reason)
}

// Extend 延长会话有效期
func (s *SessionDTO) Extend(duration time.Duration) error {
	if !s.IsActive {
		return errors.NewBusinessError("无法延长非活跃会话的有效期")
	}

	newExpiresAt := time.Now().Add(duration)
	if newExpiresAt.Before(s.ExpiresAt) {
		return errors.NewBusinessError("新的过期时间不能早于当前时间")
	}

	s.ExpiresAt = newExpiresAt
	s.UpdatedAt = time.Now()
	return nil
}

// GetTTL 获取会话剩余生存时间
func (s *SessionDTO) GetTTL() time.Duration {
	if s.IsExpired() {
		return 0
	}
	return s.ExpiresAt.Sub(time.Now())
}

// IsSameDevice 检查是否为同一设备
func (s *SessionDTO) IsSameDevice(ipAddress, userAgent string) bool {
	currentDeviceInfo := parseDeviceInfo(userAgent, ipAddress)
	storedDeviceInfo := s.GetDeviceInfo()

	return storedDeviceInfo.Fingerprint == currentDeviceInfo.Fingerprint
}

// GetSessionAge 获取会话年龄
func (s *SessionDTO) GetSessionAge() time.Duration {
	return time.Since(s.CreatedAt)
}

// GetIdleTime 获取空闲时间
func (s *SessionDTO) GetIdleTime() time.Duration {
	return time.Since(s.LastActiveAt)
}

// IsIdleTooLong 检查是否空闲时间过长
func (s *SessionDTO) IsIdleTooLong(maxIdleTime time.Duration) bool {
	return s.GetIdleTime() > maxIdleTime
}

// IsSecureSession 检查是否为安全会话
func (s *SessionDTO) IsSecureSession() bool {
	// 检查IP地址是否为内网地址
	if s.IPAddress != nil {
		if isPrivateIP(*s.IPAddress) {
			return false
		}
	}

	// 检查设备信息
	deviceInfo := s.GetDeviceInfo()
	if deviceInfo.Fingerprint == "" {
		return false
	}

	// 检查会话年龄（超过24小时的会话被认为不太安全）
	if s.GetSessionAge() > constant.SessionExpiryDefault {
		return false
	}

	return true
}

// ToJSON 转换为JSON
func (s *SessionDTO) ToJSON() string {
	sessionData := map[string]interface{}{
		"id":             s.ID,
		"user_id":        s.UserID,
		"account_id":     s.AccountID,
		"device_info":    s.GetDeviceInfo(),
		"expires_at":     s.ExpiresAt,
		"last_active_at": s.LastActiveAt,
		"is_active":      s.IsActive,
		"created_at":     s.CreatedAt,
		"ttl_seconds":    int(s.GetTTL().Seconds()),
		"age_seconds":    int(s.GetSessionAge().Seconds()),
		"idle_seconds":   int(s.GetIdleTime().Seconds()),
		"is_secure":      s.IsSecureSession(),
	}

	jsonData, _ := json.Marshal(sessionData)
	return string(jsonData)
}

// 私有方法

// parseDeviceInfo FUTURE: 设备信息解析 - 在实现设备解析时使用
func parseDeviceInfo(userAgent, ipAddress string) DeviceInfoDTO {
	deviceInfo := DeviceInfoDTO{
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	// 解析User-Agent
	ua := strings.ToLower(userAgent)

	// 检测操作系统
	if strings.Contains(ua, "windows") {
		deviceInfo.Platform = "Windows"
	} else if strings.Contains(ua, "mac") || strings.Contains(ua, "darwin") {
		deviceInfo.Platform = "macOS"
	} else if strings.Contains(ua, "linux") {
		deviceInfo.Platform = "Linux"
	} else if strings.Contains(ua, "android") {
		deviceInfo.Platform = "Android"
	} else if strings.Contains(ua, "ios") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		deviceInfo.Platform = "iOS"
	} else {
		deviceInfo.Platform = "Unknown"
	}

	// 检测浏览器
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") {
		deviceInfo.Browser = "Chrome"
	} else if strings.Contains(ua, "firefox") {
		deviceInfo.Browser = "Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		deviceInfo.Browser = "Safari"
	} else if strings.Contains(ua, "edg") {
		deviceInfo.Browser = "Edge"
	} else if strings.Contains(ua, "opera") {
		deviceInfo.Browser = "Opera"
	} else {
		deviceInfo.Browser = "Unknown"
	}

	// 检测设备类型
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		deviceInfo.Device = "Mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		deviceInfo.Device = "Tablet"
	} else {
		deviceInfo.Device = "Desktop"
	}

	// 生成设备指纹
	deviceInfo.Fingerprint = generateDeviceFingerprint(deviceInfo)

	return deviceInfo
}

// generateDeviceFingerprint 生成设备指纹
func generateDeviceFingerprint(info DeviceInfoDTO) string {
	// 使用设备信息的组合生成唯一指纹
	fingerprintData := fmt.Sprintf("%s|%s|%s", info.Platform, info.Browser, info.Device)
	return fmt.Sprintf("%x", fingerprintData) // 简化实现
}

// isPrivateIP FUTURE: 私有IP检查 - 在实现IP白名单功能时使用
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return true // 无效IP视为私有
	}

	// 检查私有IP范围
	privateIPBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8", // localhost
	}

	for _, block := range privateIPBlocks {
		_, blockNet, _ := net.ParseCIDR(block)
		if blockNet.Contains(ip) {
			return true
		}
	}

	return false
}

// SessionStats 会话统计信息
type SessionStats struct {
	TotalSessions        int            `json:"total_sessions"`
	ActiveSessions       int            `json:"active_sessions"`
	ExpiredSessions      int            `json:"expired_sessions"`
	AverageSessionTTL    time.Duration  `json:"average_session_ttl"`
	DeviceDistribution   map[string]int `json:"device_distribution"`
	PlatformDistribution map[string]int `json:"platform_distribution"`
}

// GetSessionStats 获取会话统计（从多个会话计算）
func GetSessionStats(sessions []*SessionDTO) SessionStats {
	stats := SessionStats{
		TotalSessions:        len(sessions),
		DeviceDistribution:   make(map[string]int),
		PlatformDistribution: make(map[string]int),
	}

	if len(sessions) == 0 {
		return stats
	}

	var totalTTL time.Duration
	activeCount := 0
	expiredCount := 0

	for _, session := range sessions {
		// 统计活跃/过期会话
		if session.IsValid() {
			activeCount++
		} else if session.IsExpired() {
			expiredCount++
		}

		// 统计TTL
		totalTTL += session.GetTTL()

		// 统计设备分布
		deviceInfo := session.GetDeviceInfo()
		if deviceInfo.Device != "" {
			stats.DeviceDistribution[deviceInfo.Device]++
		}

		// 统计平台分布
		if deviceInfo.Platform != "" {
			stats.PlatformDistribution[deviceInfo.Platform]++
		}
	}

	stats.ActiveSessions = activeCount
	stats.ExpiredSessions = expiredCount

	if len(sessions) > 0 {
		stats.AverageSessionTTL = totalTTL / time.Duration(len(sessions))
	}

	return stats
}
