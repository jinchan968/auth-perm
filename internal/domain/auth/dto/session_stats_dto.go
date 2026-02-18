package dto

import "time"

// SessionStatsDTO 会话统计信息
type SessionStatsDTO struct {
	TotalSessions        int64            `json:"total_sessions"`
	ActiveSessions       int64            `json:"active_sessions"`
	ExpiredSessions      int64            `json:"expired_sessions"`
	AverageTTL           time.Duration    `json:"average_ttl"`
	DeviceDistribution   map[string]int64 `json:"device_distribution"`
	PlatformDistribution map[string]int64 `json:"platform_distribution"`
}
