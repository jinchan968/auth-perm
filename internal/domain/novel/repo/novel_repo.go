package repo

import (
	"context"
	stdErr "errors"
	"time"

	"gorm.io/gorm"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/common/utils"
	"auth-perm/internal/domain/novel/constant"
	"auth-perm/internal/domain/novel/dm"
)

type QueryParams struct {
	TenantID  string
	AccountID string
	NovelID   string
	Kind      constant.CodexKind
	Status    string
	Statuses  []constant.NovelStatus
	Keyword   string
	Page      int
	PageSize  int
}

type NovelRepo struct {
	db *gorm.DB
}

func NewNovelRepo(db *gorm.DB) *NovelRepo {
	return &NovelRepo{db: db}
}

func (r *NovelRepo) GetDB() *gorm.DB { return r.db }

func (r *NovelRepo) WithTransaction(ctx context.Context, fn func(*NovelRepo) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&NovelRepo{db: tx})
	})
}

func (r *NovelRepo) CreateNovel(ctx context.Context, novel *dm.NovelDO) error {
	if err := r.db.WithContext(ctx).Create(novel).Error; err != nil {
		return errors.WrapBizError(err, "创建小说失败")
	}
	return nil
}

func (r *NovelRepo) UpdateNovel(ctx context.Context, novel *dm.NovelDO) error {
	if err := r.db.WithContext(ctx).Save(novel).Error; err != nil {
		return errors.WrapBizError(err, "更新小说失败")
	}
	return nil
}

func (r *NovelRepo) FindNovelByID(ctx context.Context, id string) (*dm.NovelDO, error) {
	var novel dm.NovelDO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&novel).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("小说不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找小说失败")
	}
	return &novel, nil
}

func (r *NovelRepo) FindPublicNovelByID(ctx context.Context, id string) (*dm.NovelDO, error) {
	var novel dm.NovelDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND status IN ?", id, []constant.NovelStatus{constant.NovelStatusSerial, constant.NovelStatusCompleted}).
		First(&novel).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("小说不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找小说失败")
	}
	return &novel, nil
}

func (r *NovelRepo) FindNovelByIDAndAccount(ctx context.Context, id, accountID, tenantID string) (*dm.NovelDO, error) {
	var novel dm.NovelDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		First(&novel).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("小说不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找小说失败")
	}
	return &novel, nil
}

func (r *NovelRepo) ListNovels(ctx context.Context, p *QueryParams) ([]*dm.NovelDO, int64, error) {
	normalizePagination(p)
	q := r.db.WithContext(ctx).Model(&dm.NovelDO{}).Where("tenant_id = ?", p.TenantID)
	if p.AccountID != "" {
		q = q.Where("account_id = ?", p.AccountID)
	}
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	} else if len(p.Statuses) > 0 {
		q = q.Where("status IN ?", p.Statuses)
	}
	if p.Keyword != "" {
		pat := utils.ILIKEPattern(p.Keyword)
		q = q.Where("title ILIKE ? OR subtitle ILIKE ? OR description ILIKE ?", pat, pat, pat)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计小说失败")
	}

	var novels []*dm.NovelDO
	err := q.Order("updated_at DESC").
		Offset((p.Page - 1) * p.PageSize).
		Limit(p.PageSize).
		Find(&novels).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询小说列表失败")
	}
	return novels, total, nil
}

func (r *NovelRepo) DeleteNovel(ctx context.Context, id, accountID, tenantID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		Delete(&dm.NovelDO{})
	if result.Error != nil {
		return errors.WrapBizError(result.Error, "删除小说失败")
	}
	if result.RowsAffected == 0 {
		return errors.NewNotFoundErrorF("小说不存在: %s", id)
	}
	return nil
}

func (r *NovelRepo) CreateVolume(ctx context.Context, volume *dm.NovelVolumeDO) error {
	if err := r.db.WithContext(ctx).Create(volume).Error; err != nil {
		return errors.WrapBizError(err, "创建分卷失败")
	}
	return nil
}

func (r *NovelRepo) UpdateVolume(ctx context.Context, volume *dm.NovelVolumeDO) error {
	if err := r.db.WithContext(ctx).Save(volume).Error; err != nil {
		return errors.WrapBizError(err, "更新分卷失败")
	}
	return nil
}

func (r *NovelRepo) FindVolumeByIDAndAccount(ctx context.Context, id, accountID, tenantID string) (*dm.NovelVolumeDO, error) {
	var volume dm.NovelVolumeDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		First(&volume).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("分卷不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找分卷失败")
	}
	return &volume, nil
}

func (r *NovelRepo) FindVolumeByTitle(ctx context.Context, novelID, title, accountID, tenantID string) (*dm.NovelVolumeDO, error) {
	var volume dm.NovelVolumeDO
	err := r.db.WithContext(ctx).
		Where("novel_id = ? AND title = ? AND account_id = ? AND tenant_id = ?", novelID, title, accountID, tenantID).
		First(&volume).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("分卷不存在: %s", title)
		}
		return nil, errors.WrapBizError(err, "查找分卷失败")
	}
	return &volume, nil
}

func (r *NovelRepo) ListVolumes(ctx context.Context, novelID string) ([]*dm.NovelVolumeDO, error) {
	var volumes []*dm.NovelVolumeDO
	err := r.db.WithContext(ctx).
		Where("novel_id = ?", novelID).
		Order("sort_order ASC, created_at ASC").
		Find(&volumes).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询分卷列表失败")
	}
	return volumes, nil
}

func (r *NovelRepo) CreateUnit(ctx context.Context, unit *dm.NovelUnitDO) error {
	if err := r.db.WithContext(ctx).Create(unit).Error; err != nil {
		return errors.WrapBizError(err, "创建单元失败")
	}
	return nil
}

func (r *NovelRepo) UpdateUnit(ctx context.Context, unit *dm.NovelUnitDO) error {
	if err := r.db.WithContext(ctx).Save(unit).Error; err != nil {
		return errors.WrapBizError(err, "更新单元失败")
	}
	return nil
}

func (r *NovelRepo) FindUnitByIDAndAccount(ctx context.Context, id, accountID, tenantID string) (*dm.NovelUnitDO, error) {
	var unit dm.NovelUnitDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		First(&unit).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("单元不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找单元失败")
	}
	return &unit, nil
}

func (r *NovelRepo) FindUnitByTitle(ctx context.Context, novelID, volumeID, title, accountID, tenantID string) (*dm.NovelUnitDO, error) {
	var unit dm.NovelUnitDO
	err := r.db.WithContext(ctx).
		Where("novel_id = ? AND volume_id = ? AND title = ? AND account_id = ? AND tenant_id = ?", novelID, volumeID, title, accountID, tenantID).
		First(&unit).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("单元不存在: %s", title)
		}
		return nil, errors.WrapBizError(err, "查找单元失败")
	}
	return &unit, nil
}

func (r *NovelRepo) ListUnits(ctx context.Context, novelID string) ([]*dm.NovelUnitDO, error) {
	var units []*dm.NovelUnitDO
	err := r.db.WithContext(ctx).
		Where("novel_id = ?", novelID).
		Order("volume_id ASC, sort_order ASC, created_at ASC").
		Find(&units).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询单元列表失败")
	}
	return units, nil
}

func (r *NovelRepo) CountChaptersByUnit(ctx context.Context, unitID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&dm.NovelChapterDO{}).Where("unit_id = ?", unitID).Count(&count).Error; err != nil {
		return 0, errors.WrapBizError(err, "统计单元章节失败")
	}
	return count, nil
}

func (r *NovelRepo) CreateChapter(ctx context.Context, chapter *dm.NovelChapterDO) error {
	if err := r.db.WithContext(ctx).Create(chapter).Error; err != nil {
		return errors.WrapBizError(err, "创建章节失败")
	}
	return nil
}

func (r *NovelRepo) UpdateChapterWithVersion(ctx context.Context, chapter *dm.NovelChapterDO, version *dm.NovelChapterVersionDO) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if version != nil {
			if err := tx.Create(version).Error; err != nil {
				return errors.WrapBizError(err, "创建章节版本失败")
			}
		}
		if err := tx.Save(chapter).Error; err != nil {
			return errors.WrapBizError(err, "更新章节失败")
		}
		return nil
	})
}

func (r *NovelRepo) UpdateChapterStatus(ctx context.Context, id, accountID, tenantID string, status constant.ChapterStatus, publishedAt *time.Time, updatedAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&dm.NovelChapterDO{}).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		Updates(map[string]any{
			"status":       status,
			"published_at": publishedAt,
			"updated_at":   updatedAt,
		})
	if result.Error != nil {
		return errors.WrapBizError(result.Error, "更新章节状态失败")
	}
	if result.RowsAffected == 0 {
		return errors.NewNotFoundErrorF("章节不存在: %s", id)
	}
	return nil
}

func (r *NovelRepo) BatchUpdateChapterStatus(ctx context.Context, ids []string, accountID, tenantID string, status constant.ChapterStatus, publishedAt *time.Time, updatedAt time.Time) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result := r.db.WithContext(ctx).
		Model(&dm.NovelChapterDO{}).
		Where("id IN ? AND account_id = ? AND tenant_id = ?", ids, accountID, tenantID).
		Updates(map[string]any{
			"status":       status,
			"published_at": publishedAt,
			"updated_at":   updatedAt,
		})
	if result.Error != nil {
		return 0, errors.WrapBizError(result.Error, "批量更新章节状态失败")
	}
	return result.RowsAffected, nil
}

func (r *NovelRepo) FindChapterByID(ctx context.Context, id string) (*dm.NovelChapterDO, error) {
	var chapter dm.NovelChapterDO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&chapter).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("章节不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找章节失败")
	}
	return &chapter, nil
}

func (r *NovelRepo) FindChapterByIDAndAccount(ctx context.Context, id, accountID, tenantID string) (*dm.NovelChapterDO, error) {
	var chapter dm.NovelChapterDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		First(&chapter).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("章节不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找章节失败")
	}
	return &chapter, nil
}

func (r *NovelRepo) ListChaptersByIDsAndAccount(ctx context.Context, ids []string, accountID, tenantID string) ([]*dm.NovelChapterDO, error) {
	if len(ids) == 0 {
		return []*dm.NovelChapterDO{}, nil
	}
	var chapters []*dm.NovelChapterDO
	err := r.db.WithContext(ctx).
		Select("id", "tenant_id", "account_id", "novel_id", "volume_id", "unit_id", "slug", "number", "title", "summary", "status", "word_count", "reading_minutes", "sort_order", "published_at", "created_at", "updated_at").
		Where("id IN ? AND account_id = ? AND tenant_id = ?", ids, accountID, tenantID).
		Find(&chapters).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询章节失败")
	}
	return chapters, nil
}

func (r *NovelRepo) FindChapterBySlugAndAccount(ctx context.Context, novelID, slug, accountID, tenantID string) (*dm.NovelChapterDO, error) {
	var chapter dm.NovelChapterDO
	err := r.db.WithContext(ctx).
		Where("novel_id = ? AND slug = ? AND account_id = ? AND tenant_id = ?", novelID, slug, accountID, tenantID).
		First(&chapter).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("章节不存在: %s", slug)
		}
		return nil, errors.WrapBizError(err, "查找章节失败")
	}
	return &chapter, nil
}

func (r *NovelRepo) FindPublishedChapterBySlug(ctx context.Context, novelID, slug string) (*dm.NovelChapterDO, error) {
	var chapter dm.NovelChapterDO
	err := r.db.WithContext(ctx).
		Where("novel_id = ? AND slug = ? AND status = ?", novelID, slug, constant.ChapterStatusPublished).
		First(&chapter).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("章节不存在: %s", slug)
		}
		return nil, errors.WrapBizError(err, "查找章节失败")
	}
	return &chapter, nil
}

func (r *NovelRepo) ListChapters(ctx context.Context, p *QueryParams) ([]*dm.NovelChapterDO, int64, error) {
	normalizePagination(p)
	q := r.db.WithContext(ctx).Model(&dm.NovelChapterDO{}).Where("novel_id = ?", p.NovelID)
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}
	if p.Keyword != "" {
		pat := utils.ILIKEPattern(p.Keyword)
		q = q.Where("title ILIKE ? OR summary ILIKE ?", pat, pat)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计章节失败")
	}

	var chapters []*dm.NovelChapterDO
	err := q.Order("sort_order ASC, created_at ASC").
		Offset((p.Page - 1) * p.PageSize).
		Limit(p.PageSize).
		Find(&chapters).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询章节列表失败")
	}
	return chapters, total, nil
}

func (r *NovelRepo) ListAllChaptersByNovel(ctx context.Context, novelID string, status string) ([]*dm.NovelChapterDO, error) {
	q := r.db.WithContext(ctx).Model(&dm.NovelChapterDO{}).Where("novel_id = ?", novelID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var chapters []*dm.NovelChapterDO
	err := q.Order("sort_order ASC, created_at ASC").Find(&chapters).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询章节列表失败")
	}
	return chapters, nil
}

func (r *NovelRepo) ListChapterVersions(ctx context.Context, chapterID string) ([]*dm.NovelChapterVersionDO, error) {
	var versions []*dm.NovelChapterVersionDO
	err := r.db.WithContext(ctx).
		Where("chapter_id = ?", chapterID).
		Order("created_at DESC").
		Find(&versions).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询章节版本失败")
	}
	return versions, nil
}

func (r *NovelRepo) UpsertCodexEntry(ctx context.Context, entry *dm.NovelCodexEntryDO) error {
	if err := r.db.WithContext(ctx).Save(entry).Error; err != nil {
		return errors.WrapBizError(err, "保存资料条目失败")
	}
	return nil
}

func (r *NovelRepo) FindCodexEntryByIDAndAccount(ctx context.Context, id, accountID, tenantID string) (*dm.NovelCodexEntryDO, error) {
	var entry dm.NovelCodexEntryDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		First(&entry).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("资料条目不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找资料条目失败")
	}
	return &entry, nil
}

func (r *NovelRepo) ListCodexEntries(ctx context.Context, p *QueryParams) ([]*dm.NovelCodexEntryDO, int64, error) {
	normalizePagination(p)
	q := r.db.WithContext(ctx).Model(&dm.NovelCodexEntryDO{}).Where("novel_id = ?", p.NovelID)
	if p.Kind != "" {
		q = q.Where("kind = ?", p.Kind)
	}
	if p.Keyword != "" {
		pat := utils.ILIKEPattern(p.Keyword)
		q = q.Where("title ILIKE ? OR summary ILIKE ? OR evidence ILIKE ?", pat, pat, pat)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errors.WrapBizError(err, "统计资料条目失败")
	}

	var entries []*dm.NovelCodexEntryDO
	err := q.Order("kind ASC, sort_order ASC, created_at ASC").
		Offset((p.Page - 1) * p.PageSize).
		Limit(p.PageSize).
		Find(&entries).Error
	if err != nil {
		return nil, 0, errors.WrapBizError(err, "查询资料条目失败")
	}
	return entries, total, nil
}

func (r *NovelRepo) CreateRuleConflict(ctx context.Context, conflict *dm.NovelRuleConflictDO) error {
	if err := r.db.WithContext(ctx).Create(conflict).Error; err != nil {
		return errors.WrapBizError(err, "创建规则冲突失败")
	}
	return nil
}

func (r *NovelRepo) UpdateRuleConflict(ctx context.Context, conflict *dm.NovelRuleConflictDO) error {
	if err := r.db.WithContext(ctx).Save(conflict).Error; err != nil {
		return errors.WrapBizError(err, "更新规则冲突失败")
	}
	return nil
}

func (r *NovelRepo) FindRuleConflictByIDAndAccount(ctx context.Context, id, accountID, tenantID string) (*dm.NovelRuleConflictDO, error) {
	var conflict dm.NovelRuleConflictDO
	err := r.db.WithContext(ctx).
		Where("id = ? AND account_id = ? AND tenant_id = ?", id, accountID, tenantID).
		First(&conflict).Error
	if err != nil {
		if stdErr.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundErrorF("规则冲突不存在: %s", id)
		}
		return nil, errors.WrapBizError(err, "查找规则冲突失败")
	}
	return &conflict, nil
}

func (r *NovelRepo) ListRuleConflicts(ctx context.Context, novelID string) ([]*dm.NovelRuleConflictDO, error) {
	var conflicts []*dm.NovelRuleConflictDO
	err := r.db.WithContext(ctx).
		Where("novel_id = ?", novelID).
		Order("created_at DESC").
		Find(&conflicts).Error
	if err != nil {
		return nil, errors.WrapBizError(err, "查询规则冲突失败")
	}
	return conflicts, nil
}

func normalizePagination(p *QueryParams) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}
