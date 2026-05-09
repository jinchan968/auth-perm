package repo

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/model"
	"auth-perm/internal/common/utils"
	"auth-perm/internal/domain/auth/constant"
	"auth-perm/internal/domain/auth/dm"
	"auth-perm/internal/domain/auth/dto"
)

// AuditLogRepo 审计日志仓储
type AuditLogRepo struct {
	db *gorm.DB
}

// NewAuditLogRepo 创建审计日志仓储
func NewAuditLogRepo(db *gorm.DB) *AuditLogRepo {
	return &AuditLogRepo{
		db: db,
	}
}

// log 记录审计日志（私有方法）
func (r *AuditLogRepo) log(ctx context.Context, entry *dto.AuditLogEntryDTO) error {
	auditLog := dm.AuditLogFromDTO(entry)
	return r.db.WithContext(ctx).Create(auditLog).Error
}

// LogAsync 异步记录审计日志（非阻塞）
func (r *AuditLogRepo) LogAsync(entry *dto.AuditLogEntryDTO) {
	go func() {
		// 创建独立的context，不依赖原始context的取消状态
		// 使用Background + Timeout来防止goroutine泄漏
		asyncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 如果有超时或错误，仅记录到日志，不影响主流程
		done := make(chan error, 1)
		go func() {
			// 在异步context中执行日志记录
			done <- r.log(asyncCtx, entry)
		}()

		select {
		case err := <-done:
			if err != nil {
				// 记录失败日志，但不影响主业务流程
				log.Printf("异步审计日志记录失败: %v", err)
			}
		case <-asyncCtx.Done():
			// 超时或取消，记录警告但不抛出错误
			log.Printf("异步审计日志记录超时或取消: %v", asyncCtx.Err())
		}
	}()
}

// LogBatch 批量记录审计日志
func (r *AuditLogRepo) LogBatch(ctx context.Context, entries []*dto.AuditLogEntryDTO) error {
	if len(entries) == 0 {
		return nil
	}

	auditLogs := make([]*dm.AuditLogDO, len(entries))
	for i, entry := range entries {
		auditLogs[i] = dm.AuditLogFromDTO(entry)
	}

	return r.db.WithContext(ctx).CreateInBatches(auditLogs, 100).Error
}

// FindByID 根据ID查找审计日志
func (r *AuditLogRepo) FindByID(ctx context.Context, id string) (*dm.AuditLogDO, error) {
	var auditLog dm.AuditLogDO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&auditLog).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("审计日志不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找审计日志失败")
	}

	return &auditLog, nil
}

// FindByUserID 根据用户ID查找审计日志
func (r *AuditLogRepo) FindByUserID(ctx context.Context, userID string, pagination *model.Pagination) ([]*dm.AuditLogDO, error) {
	var auditLogs []*dm.AuditLogDO

	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	query = r.applyPagination(query, pagination)

	err := query.Find(&auditLogs).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "通过用户查找审计日志失败")
	}

	return auditLogs, nil
}

// FindByResource 根据资源查找审计日志
func (r *AuditLogRepo) FindByResource(ctx context.Context, resourceType, resourceID string, pagination *model.Pagination) ([]*dm.AuditLogDO, error) {
	var auditLogs []*dm.AuditLogDO

	query := r.db.WithContext(ctx).Where("resource_type = ? AND resource_id = ?", resourceType, resourceID)

	query = r.applyPagination(query, pagination)

	err := query.Find(&auditLogs).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "通过资源查找审计日志失败")
	}

	return auditLogs, nil
}

// FindByAction 根据操作类型查找审计日志
func (r *AuditLogRepo) FindByAction(ctx context.Context, action string, pagination *model.Pagination) ([]*dm.AuditLogDO, error) {
	var auditLogs []*dm.AuditLogDO

	query := r.db.WithContext(ctx).Where("action = ?", action)

	if pagination != nil {
		query = r.applyPagination(query, pagination)
	}

	err := query.Find(&auditLogs).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "通过操作查找审计日志失败")
	}

	return auditLogs, nil
}

// FindByTimeRange 根据时间范围查找审计日志
func (r *AuditLogRepo) FindByTimeRange(ctx context.Context, start, end time.Time, pagination *model.Pagination) ([]*dm.AuditLogDO, error) {
	var auditLogs []*dm.AuditLogDO

	query := r.db.WithContext(ctx).Where("created_at BETWEEN ? AND ?", start, end)

	if pagination != nil {
		query = r.applyPagination(query, pagination)
	}

	err := query.Find(&auditLogs).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "通过时间范围查找审计日志失败")
	}

	return auditLogs, nil
}

// FindByTenantID 根据租户ID查找审计日志
func (r *AuditLogRepo) FindByTenantID(ctx context.Context, tenantID string, pagination *model.Pagination) ([]*dm.AuditLogDO, error) {
	var auditLogs []*dm.AuditLogDO

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if pagination != nil {
		query = r.applyPagination(query, pagination)
	}

	err := query.Find(&auditLogs).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "通过租户查找审计日志失败")
	}

	return auditLogs, nil
}

// FindFailedOperations 查找失败的操作
func (r *AuditLogRepo) FindFailedOperations(ctx context.Context, pagination *model.Pagination) ([]*dm.AuditLogDO, error) {
	var auditLogs []*dm.AuditLogDO

	query := r.db.WithContext(ctx).Where("success = ?", false)

	if pagination != nil {
		query = r.applyPagination(query, pagination)
	}

	err := query.Find(&auditLogs).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查找失败操作失败")
	}

	return auditLogs, nil
}

// GetActionStats 获取操作统计
func (r *AuditLogRepo) GetActionStats(ctx context.Context, userID string, start, end time.Time) (map[string]int64, error) {
	var stats []struct {
		Action string
		Count  int64
	}

	query := r.db.WithContext(ctx).
		Table("audit_logs").
		Select("action, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", start, end)

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	err := query.Group("action").Scan(&stats).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "获取操作统计失败")
	}

	result := make(map[string]int64)
	for _, stat := range stats {
		result[stat.Action] = stat.Count
	}

	return result, nil
}

// GetErrorStats 获取错误统计
func (r *AuditLogRepo) GetErrorStats(ctx context.Context, start, end time.Time) (map[string]int64, error) {
	var stats []struct {
		ErrorMessage string
		Count        int64
	}

	err := r.db.WithContext(ctx).
		Table("audit_logs").
		Select("error_message, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", start, end).
		Where("success = ?", false).
		Group("error_message").
		Scan(&stats).Error

	if err != nil {
		return nil, errors.WrapBizError(err, "获取错误统计失败")
	}

	result := make(map[string]int64)
	for _, stat := range stats {
		result[stat.ErrorMessage] = stat.Count
	}

	return result, nil
}

// GetOperationStats 获取操作统计（综合）
func (r *AuditLogRepo) GetOperationStats(ctx context.Context, tenantID string, start, end time.Time) (*dto.AuditLogStatsDTO, error) {
	stats := &dto.AuditLogStatsDTO{
		ActionStats:   make(map[string]int64),
		ErrorStats:    make(map[string]int64),
		SuccessStats:  make(map[string]int64),
		ResourceStats: make(map[string]int64),
	}

	query := r.db.WithContext(ctx).Table("audit_logs").Where("created_at BETWEEN ? AND ?", start, end)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// 获取操作统计
	var actionStats []struct {
		Action string
		Count  int64
	}
	err := query.Select("action, COUNT(*) as count").Group("action").Scan(&actionStats).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "获取操作统计失败")
	}
	for _, stat := range actionStats {
		stats.ActionStats[stat.Action] = stat.Count
	}

	// 获取错误统计
	var errorStats []struct {
		ErrorMessage string
		Count        int64
	}
	err = query.Select("error_message, COUNT(*) as count").Where("success = ?", false).Group("error_message").Scan(&errorStats).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "获取错误统计失败")
	}
	for _, stat := range errorStats {
		if stat.ErrorMessage != "" {
			stats.ErrorStats[stat.ErrorMessage] = stat.Count
		}
	}

	// 获取成功/失败统计
	var successStats []struct {
		Success bool
		Count   int64
	}
	err = query.Select("success, COUNT(*) as count").Group("success").Scan(&successStats).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "获取成功/失败统计失败")
	}
	for _, stat := range successStats {
		key := "success"
		if !stat.Success {
			key = "failed"
		}
		stats.SuccessStats[key] = stat.Count
	}

	// 获取资源统计
	var resourceStats []struct {
		ResourceType string
		Count        int64
	}
	err = query.Select("resource_type, COUNT(*) as count").Group("resource_type").Scan(&resourceStats).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "获取资源统计失败")
	}
	for _, stat := range resourceStats {
		stats.ResourceStats[stat.ResourceType] = stat.Count
	}

	return stats, nil
}

// CleanOldLogs 清理旧日志
func (r *AuditLogRepo) CleanOldLogs(ctx context.Context, days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days)

	result := r.db.WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&dm.AuditLogDO{})

	if result.Error != nil {
		return 0, errors.WrapBizError(result.Error, "清理旧日志失败")
	}

	return result.RowsAffected, nil
}

// ==================== 辅助方法 ====================

// applyPagination 应用分页和排序
func (r *AuditLogRepo) applyPagination(query *gorm.DB, pagination *model.Pagination) *gorm.DB {
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

// FindLoginLogs 查找用户的登录日志（按时间倒序）
func (r *AuditLogRepo) FindLoginLogs(ctx context.Context, userID string, pagination *model.Pagination, searchText, action string) ([]*dm.AuditLogDO, error) {
	var auditLogs []*dm.AuditLogDO

	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	// 操作类型过滤
	if action != "" {
		if action == "login_fail" {
			// 登录失败：查询 action='login' 且 success=false 的记录
			query = query.Where("action = ? AND success = ?", "login", false)
		} else {
			query = query.Where("action = ?", action)
		}
	} else {
		// 如果没有指定action，默认只查询登录相关操作
		query = query.Where("action IN ?", []string{constant.ActionLogin, constant.ActionLogout, constant.ActionCreateSession, constant.ActionRefreshToken})
	}

	// 添加搜索条件（IP、设备信息）
	if searchText != "" {
		pat := utils.ILIKEPattern(searchText)
		query = query.Where(
			"(ip_address LIKE ? OR user_agent LIKE ?)",
			pat,
			pat,
		)
	}

	query = r.applyPagination(query, pagination)

	err := query.Find(&auditLogs).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查找登录日志失败")
	}

	return auditLogs, nil
}

// LoginLogFilter 登录日志筛选条件
type LoginLogFilter struct {
	UserID     string    // 用户ID（管理员筛选用）
	Action     string    // 操作类型
	StartTime  time.Time // 开始时间
	EndTime    time.Time // 结束时间
	Success    *bool     // 是否成功
	SearchText string    // 搜索文本（IP、设备等）
}

// FindAllLoginLogs 查找所有登录日志（管理员用，支持分页筛选）
func (r *AuditLogRepo) FindAllLoginLogs(ctx context.Context, tenantID string, pagination *model.Pagination, filter *LoginLogFilter) ([]*dm.AuditLogDO, int64, error) {
	var auditLogs []*dm.AuditLogDO

	query := r.db.WithContext(ctx).
		Where("action IN ?", []string{constant.ActionLogin, constant.ActionLogout, constant.ActionCreateSession, constant.ActionRefreshToken})

	// 租户筛选
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// 用户筛选
	if filter != nil && filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}

	// 操作类型筛选
	if filter != nil && filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}

	// 时间范围筛选
	if filter != nil && !filter.StartTime.IsZero() {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if filter != nil && !filter.EndTime.IsZero() {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	// 成功/失败筛选
	if filter != nil && filter.Success != nil {
		query = query.Where("success = ?", *filter.Success)
	}

	// 获取总数
	var total int64
	if err := query.Model(&dm.AuditLogDO{}).Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "获取日志总数失败")
	}

	// 分页和排序
	query = query.Order("created_at DESC")
	query = r.applyPagination(query, pagination)

	err := query.Find(&auditLogs).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查找登录日志失败")
	}

	return auditLogs, total, nil
}

// CountLoginLogs 统计用户的登录日志数量
func (r *AuditLogRepo) CountLoginLogs(ctx context.Context, userID, action string) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Model(&dm.AuditLogDO{})

	// 操作类型过滤
	if action != "" {
		if action == "login_fail" {
			// 登录失败：统计 action='login' 且 success=false 的记录
			query = query.Where("action = ? AND success = ?", "login", false)
		} else {
			query = query.Where("action = ?", action)
		}
	} else {
		// 如果没有指定action，默认只统计登录相关操作
		query = query.Where("action IN ?", []string{constant.ActionLogin, constant.ActionLogout, constant.ActionCreateSession, constant.ActionRefreshToken})
	}

	err := query.Count(&count).Error

	if err != nil {
		return 0, errors.WrapBizError(err, "统计登录日志失败")
	}

	return count, nil
}
