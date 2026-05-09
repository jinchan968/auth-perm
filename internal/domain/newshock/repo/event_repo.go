// EventRepo 事件数据仓库，提供市场事件的 CRUD、搜索和统计查询。
// 包含按主题统计事件数量和重要度总和的方法，供 ScoringService 评分使用。
package repo

import (
	"context"
	stdErr "errors"
	"time"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/newshock/dm"
)

// EventQueryParams 事件列表查询参数
type EventQueryParams struct {
	TenantID   string
	ThemeID    string
	Channel    string
	Importance int
	Keyword    string
	OrderBy    string
	Page       int
	PageSize   int
}

type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{db: db}
}

// Create 创建新事件记录
func (r *EventRepo) Create(ctx context.Context, event *dm.Event) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return errors.WrapBizError(err, "创建事件失败")
	}
	return nil
}

// FindByID 按 ID 查找事件，不存在返回 NotFoundError
func (r *EventRepo) FindByID(ctx context.Context, id string) (*dm.Event, error) {
	var event dm.Event
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&event).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("事件不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找事件失败")
	}
	return &event, nil
}

// FindByIDAndTenantID 按 ID + 租户 ID 查找事件，确保租户隔离
func (r *EventRepo) FindByIDAndTenantID(ctx context.Context, id, tenantID string) (*dm.Event, error) {
	var event dm.Event
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&event).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("事件不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找事件失败")
	}
	return &event, nil
}

// List 分页查询事件列表，支持按主题、渠道、重要度、关键词筛选
func (r *EventRepo) List(ctx context.Context, params EventQueryParams) ([]dm.Event, int64, error) {
	var events []dm.Event
	var total int64

	q := r.db.WithContext(ctx).Model(&dm.Event{})

	if params.TenantID != "" {
		q = q.Where("tenant_id = ?", params.TenantID)
	}
	if params.ThemeID != "" {
		q = q.Where("theme_id = ?", params.ThemeID)
	}
	if params.Channel != "" {
		q = q.Where("channel = ?", params.Channel)
	}
	if params.Importance > 0 {
		q = q.Where("importance >= ?", params.Importance)
	}
	if params.Keyword != "" {
		pat := ilikePattern(params.Keyword)
		q = q.Where("title ILIKE ? OR summary ILIKE ?", pat, pat)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计事件数量失败")
	}

	order := sanitizeOrderBy(params.OrderBy, eventOrderAllowlist, "created_at DESC")

	page, pageSize := params.Page, params.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	err := q.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&events).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询事件列表失败")
	}
	return events, total, nil
}

// Update 更新事件记录（全量更新）
func (r *EventRepo) Update(ctx context.Context, event *dm.Event) error {
	if err := r.db.WithContext(ctx).Save(event).Error; err != nil {
		return errors.WrapBizError(err, "更新事件失败")
	}
	return nil
}

// Delete 按 ID 删除事件
func (r *EventRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&dm.Event{}).Error; err != nil {
		return errors.WrapBizError(err, "删除事件失败")
	}
	return nil
}

// Search 按关键词搜索事件（模糊匹配 title 和 summary）
func (r *EventRepo) Search(ctx context.Context, tenantID, keyword string, limit int) ([]dm.Event, error) {
	var events []dm.Event
	pat := ilikePattern(keyword)
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND (title ILIKE ? OR summary ILIKE ?)", tenantID, pat, pat).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "搜索事件失败")
	}
	return events, nil
}

// GetRecentEvents 获取最新 N 条事件，供 Dashboard 首页展示
func (r *EventRepo) GetRecentEvents(ctx context.Context, tenantID string, limit int) ([]dm.Event, error) {
	var events []dm.Event
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询最新事件失败")
	}
	return events, nil
}

// GetByThemeID 获取某主题下的事件列表，供主题详情页展示
func (r *EventRepo) GetByThemeID(ctx context.Context, themeID string, limit int) ([]dm.Event, error) {
	var events []dm.Event
	err := r.db.WithContext(ctx).
		Where("theme_id = ?", themeID).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询主题事件失败")
	}
	return events, nil
}

// GetRecentHighImpact returns recent events with importance >= minImportance
func (r *EventRepo) GetRecentHighImpact(ctx context.Context, tenantID string, minImportance int, limit int) ([]dm.Event, error) {
	var events []dm.Event
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND importance >= ?", tenantID, minImportance).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询高重要性事件失败")
	}
	return events, nil
}

// CountByTenant 统计租户下的事件总数，供 StatsService 展示
func (r *EventRepo) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&dm.Event{}).Where("tenant_id = ?", tenantID).Count(&count).Error; err != nil {
		return 0, errors.WrapBizError(err, "统计事件数量失败")
	}
	return count, nil
}

// FindByIDs 批量按 ID 查询事件，用于关联加载
func (r *EventRepo) FindByIDs(ctx context.Context, ids []string) ([]dm.Event, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var events []dm.Event
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&events).Error; err != nil {
		return nil, errors.WrapBizError(err, "批量查询事件失败")
	}
	return events, nil
}

// CountByThemeSince 统计某主题在指定时间后的事件数量
func (r *EventRepo) CountByThemeSince(ctx context.Context, themeID string, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&dm.Event{}).
		Where("theme_id = ? AND created_at >= ?", themeID, since).
		Count(&count).Error
	return count, err
}

// SumImportanceByThemeSince 统计某主题在指定时间后的事件重要度总和
func (r *EventRepo) SumImportanceByThemeSince(ctx context.Context, themeID string, since time.Time) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&dm.Event{}).
		Where("theme_id = ? AND created_at >= ?", themeID, since).
		Select("COALESCE(SUM(importance), 0)").
		Scan(&total).Error
	return total, err
}

// ClearByThemeID 清除某主题下所有事件的 theme_id 关联（删除主题时调用）
func (r *EventRepo) ClearByThemeID(ctx context.Context, themeID string) error {
	return r.db.WithContext(ctx).Model(&dm.Event{}).Where("theme_id = ?", themeID).Update("theme_id", "").Error
}
