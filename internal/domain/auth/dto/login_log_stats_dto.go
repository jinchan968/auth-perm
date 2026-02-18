package dto

// LoginLogStatsDTO 登录日志统计
type LoginLogStatsDTO struct {
	TotalCount int64            `json:"total_count"` // 总数
	WeekStats  map[string]int64 `json:"week_stats"`  // 最近7天的统计
}

// LoginLogFiltersDTO 登录日志筛选条件（API用）
type LoginLogFiltersDTO struct {
	UserID     string `json:"user_id,omitempty"`     // 用户ID（管理员用）
	Action     string `json:"action,omitempty"`      // 操作类型
	StartTime  string `json:"start_time,omitempty"`  // 开始时间（RFC3339格式）
	EndTime    string `json:"end_time,omitempty"`    // 结束时间（RFC3339格式）
	Success    *bool  `json:"success,omitempty"`     // 是否成功
	SearchText string `json:"search_text,omitempty"` // 搜索文本
}

// LoginLogsListResponse 登录日志列表响应
type LoginLogsListResponse struct {
	Data       []*AuditLogEntryDTO `json:"data"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}
