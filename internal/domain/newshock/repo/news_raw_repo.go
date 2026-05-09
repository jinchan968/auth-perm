// NewsRawRepo 原始新闻数据仓库，提供 RSS 采集新闻的存储和查询。
// 核心方法：FindByContentHash（去重）、GetUnprocessed（获取待处理）、MarkProcessed（标记已处理）。
package repo

import (
	"context"
	stdErr "errors"
	"time"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/newshock/dm"
)

type NewsRawRepo struct {
	db *gorm.DB
}

func NewNewsRawRepo(db *gorm.DB) *NewsRawRepo {
	return &NewsRawRepo{db: db}
}

// Create 创建原始新闻记录（RSS 拉取后写入）
func (r *NewsRawRepo) Create(ctx context.Context, news *dm.NewsRaw) error {
	if err := r.db.WithContext(ctx).Create(news).Error; err != nil {
		return errors.WrapBizError(err, "创建原始新闻失败")
	}
	return nil
}

// FindByID 按 ID 查找原始新闻
func (r *NewsRawRepo) FindByID(ctx context.Context, id string) (*dm.NewsRaw, error) {
	var news dm.NewsRaw
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&news).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("新闻不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找新闻失败")
	}
	return &news, nil
}

// FindByContentHash 按内容哈希查找新闻，用于 RSS 去重（相同标题+链接的新闻不重复写入）
func (r *NewsRawRepo) FindByContentHash(ctx context.Context, hash string) (*dm.NewsRaw, error) {
	var news dm.NewsRaw
	err := r.db.WithContext(ctx).Where("content_hash = ?", hash).First(&news).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.WrapBizError(err, "查找新闻失败")
	}
	return &news, nil
}

// GetUnprocessed 获取未处理的新闻（processed=false），按创建时间升序，供 NewsProcessor 消费
func (r *NewsRawRepo) GetUnprocessed(ctx context.Context, tenantID string, limit int) ([]dm.NewsRaw, error) {
	var news []dm.NewsRaw
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND processed = false", tenantID).
		Order("created_at ASC").
		Limit(limit).
		Find(&news).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询未处理新闻失败")
	}
	return news, nil
}

// MarkProcessed 标记单条新闻为已处理
func (r *NewsRawRepo) MarkProcessed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&dm.NewsRaw{}).Where("id = ?", id).Update("processed", true).Error
}

// BatchMarkProcessed 批量标记新闻为已处理
func (r *NewsRawRepo) BatchMarkProcessed(ctx context.Context, ids []string) error {
	return r.db.WithContext(ctx).Model(&dm.NewsRaw{}).Where("id IN ?", ids).Update("processed", true).Error
}

// CountUnprocessed 统计未处理的新闻数量
func (r *NewsRawRepo) CountUnprocessed(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&dm.NewsRaw{}).
		Where("tenant_id = ? AND processed = false", tenantID).
		Count(&count).Error
	return count, err
}

// CountTotal 统计新闻总数
func (r *NewsRawRepo) CountTotal(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&dm.NewsRaw{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return count, err
}

// GetLatestTime 获取最新新闻时间
func (r *NewsRawRepo) GetLatestTime(ctx context.Context, tenantID string) (*time.Time, error) {
	var news dm.NewsRaw
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		First(&news).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &news.CreatedAt, nil
}

func (r *NewsRawRepo) List(ctx context.Context, tenantID string, page, pageSize int) ([]dm.NewsRaw, int64, error) {
	var news []dm.NewsRaw
	var total int64

	q := r.db.WithContext(ctx).Model(&dm.NewsRaw{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计新闻数量失败")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	err := q.Order("published_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&news).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询新闻列表失败")
	}
	return news, total, nil
}

// DistinctTenantIDs 获取所有有新闻数据的租户 ID 列表，供批量处理遍历
func (r *NewsRawRepo) DistinctTenantIDs(ctx context.Context) ([]string, error) {
	var tenantIDs []string
	err := r.db.WithContext(ctx).Model(&dm.NewsRaw{}).
		Distinct("tenant_id").
		Pluck("tenant_id", &tenantIDs).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询新闻租户列表失败")
	}
	if len(tenantIDs) == 0 {
		tenantIDs = []string{}
	}
	return tenantIDs, nil
}
