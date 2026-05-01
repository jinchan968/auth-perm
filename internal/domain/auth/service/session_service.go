package service

import (
	"context"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/auth/constant"
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
			reasonParam = constant.ReasonUserRequestAllTenants
		}
		return s.logoutWithAllTenants(ctx, session.UserID, reasonParam)
	}

	// 登出单个会话
	reasonParam := reason
	if reasonParam == "" {
		reasonParam = constant.ReasonUserLogout
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
		Action:       constant.ActionLogout,
		ResourceType: constant.AuditResourceSession,
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
			Action:       constant.ActionLogoutAllByTenant,
			ResourceType: constant.AuditResourceSession,
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
		reasonParam = constant.ReasonUserLogoutAll
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
			Action:       constant.ActionLogoutAllByUser,
			ResourceType: constant.AuditResourceSession,
			ResourceID:   session.ID,
			NewValues: dto.AuditLogValuesDTO{
				ChangedFields: map[string]interface{}{"reason": reason},
			},
			Success: true,
		})
	}

	return nil
}

// InvalidateTenantSessions 使指定租户下的所有会话失效并清理缓存
// 实现 tenant/service.SessionInvalidator 接口
func (s *SessionService) InvalidateTenantSessions(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return errors.NewValidationError("租户ID不能为空")
	}

	// 先查找会话用于清理缓存
	sessions, err := s.sessionRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return errors.WrapBizError(err, "查找租户会话失败")
	}

	if len(sessions) == 0 {
		return nil
	}

	// 批量使会话失效
	if err := s.sessionRepo.InvalidateTenantSessions(ctx, tenantID); err != nil {
		return errors.WrapBizError(err, "批量失效租户会话失败")
	}

	// 清理缓存
	if s.cache != nil {
		for _, session := range sessions {
			sessionDTO := session.ToDTO()
			_ = s.cache.DeleteSession(ctx, sessionDTO.ID, sessionDTO.GetTenantID())
		}
	}

	return nil
}

// CleanExpiredSessionsWithCache 清理过期会话（含缓存清理）
func (s *SessionService) CleanExpiredSessionsWithCache(ctx context.Context) (int64, error) {
	// 先查找过期会话用于清理缓存
	expiredSessions, err := s.sessionRepo.FindExpiredSessions(ctx)
	if err != nil {
		return 0, errors.WrapBizError(err, "查找过期会话失败")
	}

	// 清理缓存
	if s.cache != nil && len(expiredSessions) > 0 {
		for _, session := range expiredSessions {
			sessionDTO := session.ToDTO()
			_ = s.cache.DeleteSession(ctx, sessionDTO.ID, sessionDTO.GetTenantID())
		}
	}

	// 物理删除过期会话
	count, err := s.sessionRepo.CleanExpiredSessions(ctx)
	if err != nil {
		return 0, errors.WrapBizError(err, "清理过期会话失败")
	}

	return count, nil
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
