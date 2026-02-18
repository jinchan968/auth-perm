package service

import (
	"context"
	"time"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/repo"
)

// SecurityLogsListDTO 安全日志列表DTO（供Controller使用）
type SecurityLogsListDTO struct {
	Logs     []*dto.AuditLogEntryDTO
	Total    int64
	Page     int
	PageSize int
}

// SecurityLogService 安全日志服务
type SecurityLogService struct {
	auditLogRepo *repo.AuditLogRepo
	userRepo     *repo.UserRepo
}

// NewSecurityLogService 创建安全日志服务
func NewSecurityLogService(
	auditLogRepo *repo.AuditLogRepo,
	userRepo *repo.UserRepo,
) *SecurityLogService {
	return &SecurityLogService{
		auditLogRepo: auditLogRepo,
		userRepo:     userRepo,
	}
}

// GetSecurityLogs 获取安全日志（供Controller调用）
func (s *SecurityLogService) GetSecurityLogs(ctx context.Context, userID, action, startDate, endDate, search string, page, pageSize int) (*SecurityLogsListDTO, error) {
	pagination := &model.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// 如果action为"all"或空，重置为空字符串表示不过滤
	if action == "all" || action == "" {
		action = ""
	}

	logs, err := s.auditLogRepo.FindLoginLogs(ctx, userID, pagination, search, action)
	if err != nil {
		return nil, err
	}

	// 转换为 DTO
	logDTOs := make([]*dto.AuditLogEntryDTO, len(logs))
	for i, log := range logs {
		logDTOs[i] = log.ToDTO()
	}

	// 获取总数（带搜索条件和action过滤）
	total, err := s.auditLogRepo.CountLoginLogs(ctx, userID, action)
	if err != nil {
		return nil, err
	}

	return &SecurityLogsListDTO{
		Logs:     logDTOs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetUserLoginLogs 获取用户的登录日志
func (s *SecurityLogService) GetUserLoginLogs(ctx context.Context, userID string, pagination *model.Pagination) ([]*dto.AuditLogEntryDTO, error) {
	logs, err := s.auditLogRepo.FindLoginLogs(ctx, userID, pagination, "", "")
	if err != nil {
		return nil, err
	}

	// 转换为 DTO
	result := make([]*dto.AuditLogEntryDTO, len(logs))
	for i, log := range logs {
		result[i] = log.ToDTO()
	}

	return result, nil
}

// GetAllLoginLogs 获取所有登录日志（管理员用）
func (s *SecurityLogService) GetAllLoginLogs(ctx context.Context, tenantID string, pagination *model.Pagination, filter *repo.LoginLogFilter) ([]*dto.AuditLogEntryDTO, int64, error) {
	logs, total, err := s.auditLogRepo.FindAllLoginLogs(ctx, tenantID, pagination, filter)
	if err != nil {
		return nil, 0, err
	}

	// 转换为 DTO
	result := make([]*dto.AuditLogEntryDTO, len(logs))
	for i, log := range logs {
		result[i] = log.ToDTO()
	}

	return result, total, nil
}

// GetLoginLogStats 获取用户的登录日志统计
func (s *SecurityLogService) GetLoginLogStats(ctx context.Context, userID string) (*dto.LoginLogStatsDTO, error) {
	count, err := s.auditLogRepo.CountLoginLogs(ctx, userID, "")
	if err != nil {
		return nil, err
	}

	// 获取最近7天的统计
	weekAgo := time.Now().AddDate(0, 0, -7)
	stats, err := s.auditLogRepo.GetActionStats(ctx, userID, weekAgo, time.Now())
	if err != nil {
		return nil, err
	}

	return &dto.LoginLogStatsDTO{
		TotalCount: count,
		WeekStats:  stats,
	}, nil
}

// GetLoginLogByID 根据ID获取登录日志
func (s *SecurityLogService) GetLoginLogByID(ctx context.Context, logID string) (*dto.AuditLogEntryDTO, error) {
	log, err := s.auditLogRepo.FindByID(ctx, logID)
	if err != nil {
		return nil, err
	}

	// 检查是否是登录相关日志
	loginActions := map[string]bool{
		"login": true, "logout": true, "create_session": true, "refresh_token": true,
	}
	if !loginActions[log.Action] {
		return nil, errors.NewNotFoundError("登录日志不存在")
	}

	return log.ToDTO(), nil
}
