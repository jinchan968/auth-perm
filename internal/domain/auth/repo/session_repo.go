package repo

import (
	"context"
	stdErr "errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/model"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
)

// SessionRepo 会话仓储
type SessionRepo struct {
	db *gorm.DB
}

// NewSessionRepo 创建会话仓储
func NewSessionRepo(db *gorm.DB) *SessionRepo {
	return &SessionRepo{
		db: db,
	}
}

// FindByID 根据ID查找会话
func (r *SessionRepo) FindByID(ctx context.Context, id string) (*dm.SessionDO, error) {
	var session dm.SessionDO
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&session).Error

	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("会话不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找会话失败")
	}

	return &session, nil
}

// FindByTokenHash 根据令牌哈希查找会话
func (r *SessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*dm.SessionDO, error) {
	var session dm.SessionDO
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND is_active = ? AND expires_at > ?", tokenHash, true, time.Now()).
		First(&session).Error

	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundError("会话不存在或已过期")
		}
		return nil, errors.WrapBizError(err, "通过令牌查找会话失败")
	}

	return &session, nil
}

// FindByUserID 根据用户ID查找会话
func (r *SessionRepo) FindByUserID(ctx context.Context, userID string, pagination *model.Pagination) ([]*dm.SessionDO, error) {
	var sessions []*dm.SessionDO

	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID)

	query = r.applyPagination(query, pagination)

	err := query.Find(&sessions).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "通过用户查找会话列表失败")
	}

	return sessions, nil
}

// Save 保存会话（使用 GORM Upsert 语法）
func (r *SessionRepo) Save(ctx context.Context, session *dm.SessionDO) error {
	// 使用 ON CONFLICT DO UPDATE 避免重复键错误
	return r.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: false,
			UpdateAll: true,
		},
	).Create(session).Error
}

// Delete 删除会话
func (r *SessionRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&dm.SessionDO{}, "id = ?", id).Error
}

// FindByAccountID 根据账户ID查找会话
func (r *SessionRepo) FindByAccountID(ctx context.Context, accountID string) ([]*dm.SessionDO, error) {
	var sessions []*dm.SessionDO
	err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("created_at DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "通过账户查找会话列表失败")
	}

	return sessions, nil
}

// FindByTenantID 根据租户ID查找会话
func (r *SessionRepo) FindByTenantID(ctx context.Context, tenantID string) ([]*dm.SessionDO, error) {
	var sessions []*dm.SessionDO
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "通过租户查找会话列表失败")
	}

	return sessions, nil
}

// FindActiveSessions 查找所有活跃会话
func (r *SessionRepo) FindActiveSessions(ctx context.Context) ([]*dm.SessionDO, error) {
	var sessions []*dm.SessionDO
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND expires_at > ?", true, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "查找活跃会话失败")
	}

	return sessions, nil
}

// FindExpiredSessions 查找所有过期会话
func (r *SessionRepo) FindExpiredSessions(ctx context.Context) ([]*dm.SessionDO, error) {
	var sessions []*dm.SessionDO
	err := r.db.WithContext(ctx).
		Where("expires_at <= ? OR is_active = ?", time.Now(), false).
		Find(&sessions).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "查找过期会话失败")
	}

	return sessions, nil
}

// CleanExpiredSessions 清理过期会话
func (r *SessionRepo) CleanExpiredSessions(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&dm.SessionDO{})

	if result.Error != nil {
		return 0, errors.WrapBizError(result.Error, "清理过期会话失败")
	}

	return result.RowsAffected, nil
}

// InvalidateUserSessions 使用户的所有会话失效
func (r *SessionRepo) InvalidateUserSessions(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&dm.SessionDO{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error
}

// InvalidateUserTenantSessions 使用户在指定租户下的所有会话失效
func (r *SessionRepo) InvalidateUserTenantSessions(ctx context.Context, userID, tenantID string) error {
	return r.db.WithContext(ctx).
		Model(&dm.SessionDO{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error
}

// FindByUserIDAndTenantID 根据用户ID和租户ID查找会话
func (r *SessionRepo) FindByUserIDAndTenantID(ctx context.Context, userID, tenantID string) ([]*dm.SessionDO, error) {
	var sessions []*dm.SessionDO
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Order("created_at DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "通过用户和租户查找会话列表失败")
	}

	return sessions, nil
}

// InvalidateAccountSessions 使账户的所有会话失效
func (r *SessionRepo) InvalidateAccountSessions(ctx context.Context, accountID string) error {
	return r.db.WithContext(ctx).
		Model(&dm.SessionDO{}).
		Where("account_id = ?", accountID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error
}

// InvalidateTenantSessions 使租户的所有会话失效
func (r *SessionRepo) InvalidateTenantSessions(ctx context.Context, tenantID string) error {
	return r.db.WithContext(ctx).
		Model(&dm.SessionDO{}).
		Where("tenant_id = ?", tenantID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error
}

// GetSessionStats 获取会话统计
func (r *SessionRepo) GetSessionStats(ctx context.Context) (*dto.SessionStatsDTO, error) {
	var sessions []*dm.SessionDO

	// 获取所有会话
	err := r.db.WithContext(ctx).Find(&sessions).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "获取会话统计失败")
	}

	// 计算统计信息
	stats := &dto.SessionStatsDTO{
		TotalSessions:        int64(len(sessions)),
		DeviceDistribution:   make(map[string]int64),
		PlatformDistribution: make(map[string]int64),
	}

	var activeCount, expiredCount int64

	for _, session := range sessions {
		if session.IsActive {
			activeCount++
		} else {
			expiredCount++
		}

		// 从DeviceInfo中提取设备信息
		deviceInfo := session.GetDeviceInfo()
		if deviceInfo != nil {
			if deviceInfo.Device != "" {
				stats.DeviceDistribution[deviceInfo.Device]++
			}
			if deviceInfo.Platform != "" {
				stats.PlatformDistribution[deviceInfo.Platform]++
			}
		}
	}

	stats.ActiveSessions = activeCount
	stats.ExpiredSessions = expiredCount

	// 计算平均TTL
	if len(sessions) > 0 {
		var totalTTL time.Duration
		for _, session := range sessions {
			totalTTL += session.ExpiresAt.Sub(session.CreatedAt)
		}
		stats.AverageTTL = totalTTL / time.Duration(len(sessions))
	}

	return stats, nil
}

// ==================== 辅助方法 ====================

// applyPagination 应用分页和排序
func (r *SessionRepo) applyPagination(query *gorm.DB, pagination *model.Pagination) *gorm.DB {
	if pagination == nil {
		return query
	}

	query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())

	if pagination.SortBy != "" {
		order := pagination.SortBy
		if pagination.SortDesc {
			order += " DESC"
		} else {
			order += " ASC"
		}
		query = query.Order(order)
	}

	return query
}
