package constant

import "time"

// SecurityConfig 安全配置常量
const (
	// MaxFailedAttempts 最大失败尝试次数
	MaxFailedAttempts = 5

	// LockoutDuration 锁定时间
	LockoutDuration = 15 * time.Minute

	// ResetAttemptWindow 重置失败尝试的时间窗口
	ResetAttemptWindow = 30 * time.Minute
)
