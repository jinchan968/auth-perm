package service

import (
	"context"
	"regexp"
	"time"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/journal/constant"
	"auth-perm/internal/domain/journal/dm"
	"auth-perm/internal/domain/journal/dto"
	"auth-perm/internal/domain/journal/repo"
	"auth-perm/internal/domain/journal/vo"
)

var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// JournalService 札记业务服务
type JournalService struct {
	entryRepo *repo.JournalRepo
	tagRepo   *repo.TagRepo
}

func NewJournalService(entryRepo *repo.JournalRepo, tagRepo *repo.TagRepo) *JournalService {
	return &JournalService{entryRepo: entryRepo, tagRepo: tagRepo}
}

// --------- Tag ---------

// ListTags 获取账户标签列表
func (s *JournalService) ListTags(ctx context.Context, accountID, tenantID string) (*dto.TagListResult, error) {
	tags, err := s.tagRepo.ListByAccount(ctx, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.TagDTO, 0, len(tags))
	for _, t := range tags {
		result = append(result, dto.FromTagDO(t))
	}
	return &dto.TagListResult{Data: result}, nil
}

// CreateTag 创建标签
func (s *JournalService) CreateTag(ctx context.Context, accountID, tenantID, name, color string) (*dto.TagDTO, error) {
	if name == "" {
		return nil, errors.NewValidationError("标签名称不能为空")
	}
	if color != "" && !hexColorRe.MatchString(color) {
		return nil, errors.NewValidationError("颜色格式不合法，请使用 #RRGGBB 格式")
	}
	existing, err := s.tagRepo.FindByNameAndAccount(ctx, name, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewValidationError("标签名称已存在")
	}
	tag := dm.NewTag(tenantID, accountID, name, color)
	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, err
	}
	return dto.FromTagDO(tag), nil
}

// UpdateTag 更新标签
func (s *JournalService) UpdateTag(ctx context.Context, id, accountID, tenantID, name, color string) (*dto.TagDTO, error) {
	tag, err := s.tagRepo.FindByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if name != "" {
		existing, err := s.tagRepo.FindByNameAndAccount(ctx, name, accountID, tenantID)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.NewValidationError("标签名称已存在")
		}
		tag.Name = name
	}
	if color != "" {
		if !hexColorRe.MatchString(color) {
			return nil, errors.NewValidationError("颜色格式不合法，请使用 #RRGGBB 格式")
		}
		tag.Color = color
	}
	tag.UpdatedAt = time.Now()
	if err := s.tagRepo.Update(ctx, tag); err != nil {
		return nil, err
	}
	return dto.FromTagDO(tag), nil
}

// DeleteTag 软删除标签
func (s *JournalService) DeleteTag(ctx context.Context, id, accountID, tenantID string) error {
	return s.tagRepo.SoftDelete(ctx, id, accountID, tenantID)
}

// --------- Entry ---------

// deduplicateTagIDs 对 tag ID 列表去重
func deduplicateTagIDs(ids []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

// CreateEntry 创建札记主条目
func (s *JournalService) CreateEntry(ctx context.Context, p *vo.CreateEntryParams) (*dto.EntryDTO, error) {
	if p.Content == "" {
		return nil, errors.NewValidationError("札记内容不能为空")
	}
	if len([]rune(p.Content)) > 800 {
		return nil, errors.NewValidationError("札记内容不能超过800字")
	}
	if !constant.IsValidPeriod(p.Period) {
		return nil, errors.NewValidationError("时段不合法")
	}
	if p.Weather != nil && !constant.IsValidWeather(*p.Weather) {
		return nil, errors.NewValidationError("天气不合法")
	}

	// 去重后再校验
	dedupedTagIDs := deduplicateTagIDs(p.TagIDs)

	// 验证所有标签存在且属于当前账户
	if len(dedupedTagIDs) > 0 {
		validTags, err := s.tagRepo.FindByIDs(ctx, dedupedTagIDs, p.AccountID, p.TenantID)
		if err != nil {
			return nil, err
		}
		if len(validTags) != len(dedupedTagIDs) {
			return nil, errors.NewValidationError("部分标签不存在或无权访问")
		}
	}

	entry := dm.NewJournalEntry(p.TenantID, p.AccountID, p.Title, p.Content, p.Weather, p.Location, p.Period, p.EntryDate)

	if err := s.entryRepo.Create(ctx, entry); err != nil {
		return nil, err
	}

	// 关联标签
	if len(dedupedTagIDs) > 0 {
		if err := s.entryRepo.ReplaceTags(ctx, entry.ID, dedupedTagIDs); err != nil {
			return nil, err
		}
		tags, err := s.tagRepo.FindByIDs(ctx, dedupedTagIDs, p.AccountID, p.TenantID)
		if err != nil {
			return nil, errors.WrapBizError(err, "加载标签失败")
		}
		entry.Tags = tags
	}

	return dto.FromEntryDO(entry), nil
}

// GetEntry 获取札记详情（含修正和标签）
func (s *JournalService) GetEntry(ctx context.Context, id, accountID, tenantID string) (*dto.EntryDTO, error) {
	entry, err := s.entryRepo.FindByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	return dto.FromEntryDO(entry), nil
}

// ListEntries 按日期范围查询札记列表
func (s *JournalService) ListEntries(ctx context.Context, p *repo.JournalQueryParams) (*dto.EntryListResult, error) {
	entries, total, err := s.entryRepo.ListByDateRange(ctx, p)
	if err != nil {
		return nil, err
	}
	return &dto.EntryListResult{
		Data:     dto.FromEntryDOList(entries),
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}

// AddCorrection 为札记追加修正
func (s *JournalService) AddCorrection(ctx context.Context, entryID, accountID, tenantID, content string) (*dto.EntryDTO, error) {
	if content == "" {
		return nil, errors.NewValidationError("修正内容不能为空")
	}
	if len([]rune(content)) > 800 {
		return nil, errors.NewValidationError("修正内容不能超过800字")
	}

	// 验证主条目存在且属于当前账户
	entry, err := s.entryRepo.FindByIDAndAccount(ctx, entryID, accountID, tenantID)
	if err != nil {
		return nil, err
	}

	correction := dm.NewCorrection(tenantID, accountID, entry.ID, content, entry.EntryDate)
	if err := s.entryRepo.Create(ctx, correction); err != nil {
		return nil, err
	}

	// 重新加载以获取最新数据
	return s.GetEntry(ctx, entryID, accountID, tenantID)
}

// UpdateTags 更新札记标签（随时可增减）
func (s *JournalService) UpdateTags(ctx context.Context, entryID, accountID, tenantID string, tagIDs []string) (*dto.EntryDTO, error) {
	// 验证主条目存在且属于当前账户
	_, err := s.entryRepo.FindByIDAndAccount(ctx, entryID, accountID, tenantID)
	if err != nil {
		return nil, err
	}

	// 去重后再校验
	deduped := deduplicateTagIDs(tagIDs)

	// 验证所有标签存在且属于当前账户
	if len(deduped) > 0 {
		validTags, err := s.tagRepo.FindByIDs(ctx, deduped, accountID, tenantID)
		if err != nil {
			return nil, err
		}
		if len(validTags) != len(deduped) {
			return nil, errors.NewValidationError("部分标签不存在或无权访问")
		}
	}

	if err := s.entryRepo.ReplaceTags(ctx, entryID, deduped); err != nil {
		return nil, err
	}

	return s.GetEntry(ctx, entryID, accountID, tenantID)
}

// DeleteEntry 软删除札记
func (s *JournalService) DeleteEntry(ctx context.Context, id, accountID, tenantID string) error {
	return s.entryRepo.SoftDelete(ctx, id, accountID, tenantID)
}
