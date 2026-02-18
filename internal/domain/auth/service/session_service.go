package service

import (
	"context"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
	"auth-perm/internal/domain/auth/param"
	"auth-perm/internal/domain/auth/repo"
)

// SessionService 会话管理服务
type SessionService struct {
	sessionRepo *repo.SessionRepo
	auditRepo   *repo.AuditLogRepo
	cache       *CacheService
}

// NewSessionService 创建会话管理服务
func NewSessionService(
	sessionRepo *repo.SessionRepo,
	auditRepo *repo.AuditLogRepo,
	cacheService *CacheService,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		cache:       cacheService,
	}
}

// Logout 登出
func (s *SessionService) Logout(ctx context.Context, sessionID string, logoutAllTenants bool, reason string) error {
	// 通过sessionID查找会话
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return errors.WrapBizError(err, "查找会话失败")
	}
	if session == nil {
		return errors.NewNotFoundError("会话不存在")
	}

	// 如果需要登出所有租户下的会话
	if logoutAllTenants {
		// 登出用户在所有租户下的会话
		reasonParam := reason
		if reasonParam == "" {
			reasonParam = "user_request_all_tenants"
		}
		return s.logoutWithAllTenants(ctx, session.UserID, reasonParam)
	}

	// 登出单个会话
	reasonParam := reason
	if reasonParam == "" {
		reasonParam = "user_logout"
	}

	// 使会话失效
	sessionDTO := session.ToDTO()
	sessionDTO.Invalidate(reasonParam)
	if err := s.sessionRepo.Save(ctx, dm.SessionFromDTO(sessionDTO)); err != nil {
		return errors.WrapBizError(err, "更新会话失败")
	}

	// 清理缓存
	if s.cache != nil {
		_ = s.cache.DeleteSession(ctx, sessionDTO.ID, sessionDTO.GetTenantID())
	}

	// 异步记录审计日志
	s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
		Action:       "logout",
		ResourceType: "session",
		ResourceID:   session.ID,
		NewValues: dto.AuditLogValuesDTO{
			ChangedFields: map[string]interface{}{"reason": reasonParam},
		},
		Success: true,
	})

	return nil
}

// LogoutAllByTenant 按租户登出所有会话（管理员接口）
func (s *SessionService) LogoutAllByTenant(ctx context.Context, tenantID string, reason string) error {
	if tenantID == "" {
		return errors.NewValidationError("请指定租户ID")
	}

	// 查找该租户下所有会话
	sessions, err := s.sessionRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return errors.WrapBizError(err, "查找租户会话失败")
	}

	if len(sessions) == 0 {
		// 没有会话，直接返回成功
		return nil
	}

	// 批量使会话失效
	if err := s.sessionRepo.InvalidateTenantSessions(ctx, tenantID); err != nil {
		return errors.WrapBizError(err, "批量失效会话失败")
	}

	// 清理缓存
	if s.cache != nil {
		for _, session := range sessions {
			sessionDTO := session.ToDTO()
			_ = s.cache.DeleteSession(ctx, sessionDTO.ID, sessionDTO.GetTenantID())
		}
	}

	// 记录审计日志（批量）
	for _, session := range sessions {
		s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
			Action:       "logout_all_by_tenant",
			ResourceType: "session",
			ResourceID:   session.ID,
			NewValues: dto.AuditLogValuesDTO{
				ChangedFields: map[string]interface{}{"reason": reason},
			},
			Success: true,
		})
	}

	return nil
}

// LogoutAllByUser 登出用户所有会话
func (s *SessionService) LogoutAllByUser(ctx context.Context, userID string, reason string) error {
	reasonParam := reason
	if reasonParam == "" {
		reasonParam = "user_logout_all"
	}
	return s.logoutWithAllTenants(ctx, userID, reasonParam)
}

// logoutWithAllTenants 登出用户在所有租户下的会话
func (s *SessionService) logoutWithAllTenants(ctx context.Context, userID string, reason string) error {
	if userID == "" {
		return errors.NewValidationError("用户ID不能为空")
	}

	// 查找该用户所有会话 todo 这里是否要考虑 sessions返回条数过多情况 如果不考虑考虑 是否考虑在登陆时限制最多登陆tenant数量
	sessions, err := s.sessionRepo.FindByUserID(ctx, userID, nil)
	if err != nil {
		return errors.WrapBizError(err, "查找用户会话失败")
	}

	if len(sessions) == 0 {
		// 没有会话，直接返回成功
		return nil
	}

	// 批量使会话失效
	if err := s.sessionRepo.InvalidateUserSessions(ctx, userID); err != nil {
		return errors.WrapBizError(err, "批量失效会话失败")
	}

	// 清理缓存
	if s.cache != nil {
		for _, session := range sessions {
			sessionDTO := session.ToDTO()
			_ = s.cache.DeleteSession(ctx, sessionDTO.ID, sessionDTO.GetTenantID())
		}
	}

	// 记录审计日志（批量）
	for _, session := range sessions {
		s.auditRepo.LogAsync(&dto.AuditLogEntryDTO{
			Action:       "logout_all_by_user",
			ResourceType: "session",
			ResourceID:   session.ID,
			NewValues: dto.AuditLogValuesDTO{
				ChangedFields: map[string]interface{}{"reason": reason},
			},
			Success: true,
		})
	}

	return nil
}

// GetUserSessions 获取用户会话列表
func (s *SessionService) GetUserSessions(ctx context.Context, params *param.GetSessionsParams) ([]*dto.SessionDTO, *model.Pagination, error) {
	// 验证参数
	if err := params.Validate(); err != nil {
		return nil, nil, errors.NewValidationError(err.Error())
	}

	// 创建分页对象
	page, pageSize := params.GetPagination()
	pagination := &model.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	// 从仓储查询会话
	sessions, err := s.sessionRepo.FindByUserID(ctx, params.UserID, pagination)
	if err != nil {
		return nil, nil, errors.WrapBizError(err, "获取会话列表失败")
	}

	// 转换为DTO
	sessionDTOs := make([]*dto.SessionDTO, len(sessions))
	for i, session := range sessions {
		sessionDTOs[i] = session.ToDTO()
	}

	return sessionDTOs, pagination, nil
}
