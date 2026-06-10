package service

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"

	"auth-perm/internal/common/errors"
	"auth-perm/internal/domain/novel/constant"
	"auth-perm/internal/domain/novel/dm"
	"auth-perm/internal/domain/novel/dto"
	"auth-perm/internal/domain/novel/repo"
	"auth-perm/internal/domain/novel/vo"
)

type novelImportTask struct {
	mu       sync.Mutex
	TaskID   string
	NovelID  string
	Status   constant.ImportTaskStatus
	Progress *vo.ImportTaskProgressVO
	Result   *vo.MarkdownBundleImportResultVO
	Err      string
	CreatedAt time.Time
}

type NovelService struct {
	novelRepo   *repo.NovelRepo
	importTasks sync.Map
	taskTTL     time.Duration
}

func NewNovelService(novelRepo *repo.NovelRepo) *NovelService {
	s := &NovelService{
		novelRepo: novelRepo,
		taskTTL:   30 * time.Minute,
	}
	go s.cleanupExpiredTasks()
	return s
}

func (s *NovelService) ListNovels(ctx context.Context, p *repo.QueryParams) (*vo.ListResult[vo.NovelVO], error) {
	novels, total, err := s.novelRepo.ListNovels(ctx, p)
	if err != nil {
		return nil, err
	}
	data := make([]*vo.NovelVO, 0, len(novels))
	for _, novel := range novels {
		data = append(data, vo.FromNovelDO(novel))
	}
	return &vo.ListResult[vo.NovelVO]{Data: data, Total: total, Page: p.Page, Size: p.PageSize}, nil
}

func (s *NovelService) ListPublicNovels(ctx context.Context, p *repo.QueryParams) (*vo.ListResult[vo.NovelVO], error) {
	switch constant.NovelStatus(p.Status) {
	case constant.NovelStatusSerial, constant.NovelStatusCompleted:
		p.Statuses = nil
	case "":
		p.Statuses = []constant.NovelStatus{constant.NovelStatusSerial, constant.NovelStatusCompleted}
	default:
		p.Status = ""
		p.Statuses = []constant.NovelStatus{constant.NovelStatusSerial, constant.NovelStatusCompleted}
	}
	return s.ListNovels(ctx, p)
}

func (s *NovelService) GetNovelDetail(ctx context.Context, novelID string, includeDraft bool) (*vo.NovelDetailVO, error) {
	var novel *dm.NovelDO
	var err error
	if includeDraft {
		novel, err = s.novelRepo.FindNovelByID(ctx, novelID)
	} else {
		novel, err = s.novelRepo.FindPublicNovelByID(ctx, novelID)
	}
	if err != nil {
		return nil, err
	}
	status := string(constant.ChapterStatusPublished)
	if includeDraft {
		status = ""
	}
	chapters, err := s.novelRepo.ListAllChaptersByNovel(ctx, novelID, status)
	if err != nil {
		return nil, err
	}
	volumes, err := s.novelRepo.ListVolumes(ctx, novelID)
	if err != nil {
		return nil, err
	}
	units, err := s.novelRepo.ListUnits(ctx, novelID)
	if err != nil {
		return nil, err
	}
	if !includeDraft {
		volumes = filterVolumesByChapters(volumes, chapters)
		units = filterUnitsByChapters(units, chapters)
	}
	result := &vo.NovelDetailVO{
		Novel:    vo.FromNovelDO(novel),
		Volumes:  make([]*vo.VolumeVO, 0, len(volumes)),
		Units:    make([]*vo.UnitVO, 0, len(units)),
		Chapters: make([]*vo.ChapterVO, 0, len(chapters)),
	}
	for _, volume := range volumes {
		result.Volumes = append(result.Volumes, vo.FromVolumeDO(volume))
	}
	for _, unit := range units {
		result.Units = append(result.Units, vo.FromUnitDO(unit))
	}
	for _, chapter := range chapters {
		result.Chapters = append(result.Chapters, vo.FromChapterDO(chapter, false))
	}
	return result, nil
}

func (s *NovelService) GetManagedNovelDetail(ctx context.Context, novelID, accountID, tenantID string) (*vo.NovelDetailVO, error) {
	if _, err := s.novelRepo.FindNovelByIDAndAccount(ctx, novelID, accountID, tenantID); err != nil {
		return nil, err
	}
	return s.GetNovelDetail(ctx, novelID, true)
}

func (s *NovelService) EnsureNovelAccess(ctx context.Context, novelID, accountID, tenantID string) error {
	_, err := s.novelRepo.FindNovelByIDAndAccount(ctx, novelID, accountID, tenantID)
	return err
}

func (s *NovelService) CreateNovel(ctx context.Context, tenantID, accountID string, req *dto.CreateNovelRequest) (*vo.NovelVO, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.NewValidationError("小说标题不能为空")
	}
	if !vo.IsValidNovelStatus(req.Status) {
		return nil, errors.NewValidationError("小说状态不合法")
	}
	novel := dm.NewNovel(tenantID, accountID, strings.TrimSpace(req.Title))
	novel.Subtitle = req.Subtitle
	novel.Description = req.Description
	novel.CoverURL = req.CoverURL
	novel.Tags = req.Tags
	if novel.Tags == nil {
		novel.Tags = []string{}
	}
	if req.Status != "" {
		novel.Status = req.Status
	}
	if err := s.novelRepo.CreateNovel(ctx, novel); err != nil {
		return nil, err
	}
	return vo.FromNovelDO(novel), nil
}

func (s *NovelService) UpdateNovel(ctx context.Context, id, accountID, tenantID string, req *dto.UpdateNovelRequest) (*vo.NovelVO, error) {
	novel, err := s.novelRepo.FindNovelByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return nil, errors.NewValidationError("小说标题不能为空")
		}
		novel.Title = strings.TrimSpace(*req.Title)
	}
	if req.Subtitle != nil {
		novel.Subtitle = *req.Subtitle
	}
	if req.Description != nil {
		novel.Description = *req.Description
	}
	if req.CoverURL != nil {
		novel.CoverURL = *req.CoverURL
	}
	if req.Status != nil {
		if !vo.IsValidNovelStatus(*req.Status) {
			return nil, errors.NewValidationError("小说状态不合法")
		}
		novel.Status = *req.Status
	}
	if req.Tags != nil {
		novel.Tags = req.Tags
	}
	novel.UpdatedAt = time.Now()
	if err := s.novelRepo.UpdateNovel(ctx, novel); err != nil {
		return nil, err
	}
	return vo.FromNovelDO(novel), nil
}

func (s *NovelService) DeleteNovel(ctx context.Context, id, accountID, tenantID string) error {
	return s.novelRepo.DeleteNovel(ctx, id, accountID, tenantID)
}

func (s *NovelService) CreateVolume(ctx context.Context, novelID, accountID, tenantID string, req *dto.CreateVolumeRequest) (*vo.VolumeVO, error) {
	if _, err := s.novelRepo.FindNovelByIDAndAccount(ctx, novelID, accountID, tenantID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.NewValidationError("分卷标题不能为空")
	}
	volume := dm.NewVolume(tenantID, accountID, novelID, strings.TrimSpace(req.Title), req.SortOrder)
	volume.Subtitle = req.Subtitle
	volume.Description = req.Description
	if err := s.novelRepo.CreateVolume(ctx, volume); err != nil {
		return nil, err
	}
	return vo.FromVolumeDO(volume), nil
}

func (s *NovelService) ListVolumes(ctx context.Context, novelID string) ([]*vo.VolumeVO, error) {
	volumes, err := s.novelRepo.ListVolumes(ctx, novelID)
	if err != nil {
		return nil, err
	}
	data := make([]*vo.VolumeVO, 0, len(volumes))
	for _, volume := range volumes {
		data = append(data, vo.FromVolumeDO(volume))
	}
	return data, nil
}

func (s *NovelService) ListManagedVolumes(ctx context.Context, novelID, accountID, tenantID string) ([]*vo.VolumeVO, error) {
	if err := s.EnsureNovelAccess(ctx, novelID, accountID, tenantID); err != nil {
		return nil, err
	}
	return s.ListVolumes(ctx, novelID)
}

func (s *NovelService) UpdateVolume(ctx context.Context, id, accountID, tenantID string, req *dto.UpdateVolumeRequest) (*vo.VolumeVO, error) {
	volume, err := s.novelRepo.FindVolumeByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return nil, errors.NewValidationError("分卷标题不能为空")
		}
		volume.Title = strings.TrimSpace(*req.Title)
	}
	if req.Subtitle != nil {
		volume.Subtitle = *req.Subtitle
	}
	if req.Description != nil {
		volume.Description = *req.Description
	}
	if req.SortOrder != nil {
		volume.SortOrder = *req.SortOrder
	}
	volume.UpdatedAt = time.Now()
	if err := s.novelRepo.UpdateVolume(ctx, volume); err != nil {
		return nil, err
	}
	return vo.FromVolumeDO(volume), nil
}

func (s *NovelService) CreateUnit(ctx context.Context, novelID, accountID, tenantID string, req *dto.CreateUnitRequest) (*vo.UnitVO, error) {
	if _, err := s.novelRepo.FindNovelByIDAndAccount(ctx, novelID, accountID, tenantID); err != nil {
		return nil, err
	}
	volume, err := s.novelRepo.FindVolumeByIDAndAccount(ctx, req.VolumeID, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if volume.NovelID != novelID {
		return nil, errors.NewValidationError("分卷不属于该小说")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.NewValidationError("单元标题不能为空")
	}
	unit := dm.NewUnit(tenantID, accountID, novelID, req.VolumeID, strings.TrimSpace(req.Title), req.SortOrder)
	unit.Subtitle = req.Subtitle
	unit.Description = req.Description
	if err := s.novelRepo.CreateUnit(ctx, unit); err != nil {
		return nil, err
	}
	return vo.FromUnitDO(unit), nil
}

func (s *NovelService) ListUnits(ctx context.Context, novelID string) ([]*vo.UnitVO, error) {
	units, err := s.novelRepo.ListUnits(ctx, novelID)
	if err != nil {
		return nil, err
	}
	data := make([]*vo.UnitVO, 0, len(units))
	for _, unit := range units {
		data = append(data, vo.FromUnitDO(unit))
	}
	return data, nil
}

func (s *NovelService) ListManagedUnits(ctx context.Context, novelID, accountID, tenantID string) ([]*vo.UnitVO, error) {
	if err := s.EnsureNovelAccess(ctx, novelID, accountID, tenantID); err != nil {
		return nil, err
	}
	return s.ListUnits(ctx, novelID)
}

func (s *NovelService) UpdateUnit(ctx context.Context, id, accountID, tenantID string, req *dto.UpdateUnitRequest) (*vo.UnitVO, error) {
	unit, err := s.novelRepo.FindUnitByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if req.VolumeID != nil {
		volume, err := s.novelRepo.FindVolumeByIDAndAccount(ctx, *req.VolumeID, accountID, tenantID)
		if err != nil {
			return nil, err
		}
		if volume.NovelID != unit.NovelID {
			return nil, errors.NewValidationError("分卷不属于该小说")
		}
		if *req.VolumeID != unit.VolumeID {
			count, err := s.novelRepo.CountChaptersByUnit(ctx, unit.ID)
			if err != nil {
				return nil, err
			}
			if count > 0 {
				return nil, errors.NewValidationError("单元下已有章节，不能移动到其他分卷")
			}
		}
		unit.VolumeID = *req.VolumeID
	}
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return nil, errors.NewValidationError("单元标题不能为空")
		}
		unit.Title = strings.TrimSpace(*req.Title)
	}
	if req.Subtitle != nil {
		unit.Subtitle = *req.Subtitle
	}
	if req.Description != nil {
		unit.Description = *req.Description
	}
	if req.SortOrder != nil {
		unit.SortOrder = *req.SortOrder
	}
	unit.UpdatedAt = time.Now()
	if err := s.novelRepo.UpdateUnit(ctx, unit); err != nil {
		return nil, err
	}
	return vo.FromUnitDO(unit), nil
}

func (s *NovelService) CreateChapter(ctx context.Context, novelID, accountID, tenantID string, req *dto.CreateChapterRequest) (*vo.ChapterVO, error) {
	if _, err := s.novelRepo.FindNovelByIDAndAccount(ctx, novelID, accountID, tenantID); err != nil {
		return nil, err
	}
	volume, err := s.novelRepo.FindVolumeByIDAndAccount(ctx, req.VolumeID, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if volume.NovelID != novelID {
		return nil, errors.NewValidationError("分卷不属于该小说")
	}
	unitID, err := s.resolveChapterUnit(ctx, novelID, req.VolumeID, req.UnitID, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.NewValidationError("章节标题不能为空")
	}
	if !vo.IsValidChapterStatus(req.Status) {
		return nil, errors.NewValidationError("章节状态不合法")
	}
	chapter := dm.NewChapter(tenantID, accountID, novelID, req.VolumeID, strings.TrimSpace(req.Title), req.SortOrder)
	chapter.UnitID = unitID
	chapter.Slug = normalizeSlug(req.Slug, chapter.ID)
	chapter.Number = req.Number
	chapter.Summary = req.Summary
	chapter.Body = req.Body
	if req.Status != "" {
		chapter.Status = req.Status
	}
	applyChapterMetrics(chapter)
	if chapter.Status == constant.ChapterStatusPublished {
		if err := s.ensureNoBlockingConflicts(ctx, chapter.NovelID); err != nil {
			return nil, err
		}
		now := time.Now()
		chapter.PublishedAt = &now
	}
	if err := s.novelRepo.CreateChapter(ctx, chapter); err != nil {
		return nil, err
	}
	return vo.FromChapterDO(chapter, true), nil
}

func (s *NovelService) ImportMarkdownChapter(ctx context.Context, novelID, accountID, tenantID string, req *dto.ImportMarkdownChapterRequest) (*vo.ChapterVO, error) {
	parsed, err := parseMarkdownChapter(req.Markdown)
	if err != nil {
		return nil, err
	}

	title := firstNonEmpty(req.Title, parsed.Title)
	if title == "" {
		return nil, errors.NewValidationError("Markdown 章节标题不能为空")
	}

	status := req.Status
	if status == "" {
		status = parsed.Status
	}
	sortOrder := req.SortOrder
	if sortOrder == 0 {
		sortOrder = parsed.SortOrder
	}

	return s.CreateChapter(ctx, novelID, accountID, tenantID, &dto.CreateChapterRequest{
		VolumeID:  req.VolumeID,
		UnitID:    req.UnitID,
		Slug:      firstNonEmpty(req.Slug, parsed.Slug),
		Number:    firstNonEmpty(req.Number, parsed.Number),
		Title:     title,
		Summary:   firstNonEmpty(req.Summary, parsed.Summary),
		Body:      parsed.Body,
		Status:    status,
		SortOrder: sortOrder,
	})
}

func (s *NovelService) ImportMarkdownBundle(ctx context.Context, novelID, accountID, tenantID string, req *dto.ImportMarkdownBundleRequest) (*vo.MarkdownBundleImportResultVO, error) {
	return s.ImportMarkdownBundleWithProgress(ctx, novelID, accountID, tenantID, req, nil)
}

func (s *NovelService) ImportMarkdownBundleWithProgress(ctx context.Context, novelID, accountID, tenantID string, req *dto.ImportMarkdownBundleRequest, onProgress func(processed int)) (*vo.MarkdownBundleImportResultVO, error) {
	if _, err := s.novelRepo.FindNovelByIDAndAccount(ctx, novelID, accountID, tenantID); err != nil {
		return nil, err
	}
	if len(req.Files) == 0 {
		return nil, errors.NewValidationError("Markdown 文件列表不能为空")
	}

	files := append([]dto.MarkdownBundleFile(nil), req.Files...)
	sort.SliceStable(files, func(i, j int) bool {
		return normalizeBundlePath(files[i].Path) < normalizeBundlePath(files[j].Path)
	})

	result := &vo.MarkdownBundleImportResultVO{
		Items: make([]*vo.MarkdownBundleImportItemVO, 0, len(files)),
	}
	err := s.novelRepo.WithTransaction(ctx, func(txRepo *repo.NovelRepo) error {
		txSvc := NewNovelService(txRepo)
		for index, file := range files {
			item, err := txSvc.importMarkdownBundleFile(ctx, novelID, accountID, tenantID, file, index+1)
			if err != nil {
				return err
			}
			result.Items = append(result.Items, item)
			if item.Skipped {
				result.Skipped++
				continue
			}
			result.Imported++
			switch item.Action {
			case "created":
				result.Created++
			case "updated":
				result.Updated++
			}
			if onProgress != nil {
				onProgress(index + 1)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *NovelService) ImportMarkdownBundleAsync(novelID, accountID, tenantID string, req *dto.ImportMarkdownBundleRequest) string {
	taskID := uuid.New().String()
	task := &novelImportTask{
		TaskID:    taskID,
		NovelID:   novelID,
		Status:    constant.ImportTaskStatusPending,
		Progress:  &vo.ImportTaskProgressVO{Total: len(req.Files)},
		CreatedAt: time.Now(),
	}
	s.importTasks.Store(taskID, task)

	go func() {
		task.mu.Lock()
		task.Status = constant.ImportTaskStatusProcessing
		task.mu.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		result, err := s.ImportMarkdownBundleWithProgress(ctx, novelID, accountID, tenantID, req, func(processed int) {
			task.mu.Lock()
			task.Progress.Processed = processed
			task.mu.Unlock()
		})
		task.mu.Lock()
		defer task.mu.Unlock()
		if err != nil {
			task.Status = constant.ImportTaskStatusFailed
			task.Err = err.Error()
		} else {
			task.Status = constant.ImportTaskStatusSuccess
			task.Result = result
			task.Progress.Processed = task.Progress.Total
		}
	}()

	return taskID
}

func (s *NovelService) GetImportTask(ctx context.Context, taskID, accountID, tenantID string) (*vo.ImportTaskVO, error) {
	raw, ok := s.importTasks.Load(taskID)
	if !ok {
		return nil, errors.NewNotFoundError("导入任务不存在")
	}
	task := raw.(*novelImportTask)

	if task.NovelID != "" {
		novel, err := s.novelRepo.FindNovelByIDAndAccount(ctx, task.NovelID, accountID, tenantID)
		if err != nil || novel == nil {
			return nil, errors.NewNotFoundError("导入任务不存在")
		}
	}

	task.mu.Lock()
	defer task.mu.Unlock()
	return &vo.ImportTaskVO{
		TaskID:   task.TaskID,
		NovelID:  task.NovelID,
		Status:   task.Status,
		Progress: task.Progress,
		Result:   task.Result,
		Error:    task.Err,
	}, nil
}

func (s *NovelService) cleanupExpiredTasks() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.importTasks.Range(func(key, value any) bool {
			task := value.(*novelImportTask)
			task.mu.Lock()
			age := time.Since(task.CreatedAt)
			isTerminal := task.Status == constant.ImportTaskStatusSuccess || task.Status == constant.ImportTaskStatusFailed
			task.mu.Unlock()
			if isTerminal && age > s.taskTTL {
				s.importTasks.Delete(key)
			}
			return true
		})
	}
}

func (s *NovelService) InspectMarkdownBundle(ctx context.Context, req *dto.InspectMarkdownBundleRequest) (*vo.MarkdownBundleInspectResultVO, error) {
	if len(req.Files) == 0 {
		return nil, errors.NewValidationError("Markdown 文件列表不能为空")
	}

	files := append([]dto.MarkdownBundleFile(nil), req.Files...)
	sort.SliceStable(files, func(i, j int) bool {
		return normalizeBundlePath(files[i].Path) < normalizeBundlePath(files[j].Path)
	})

	result := &vo.MarkdownBundleInspectResultVO{
		Items:    make([]*vo.MarkdownBundleInspectItemVO, 0, len(files)),
		Strategy: "path-convention",
	}
	volumeSet := make(map[string]struct{})
	unitSet := make(map[string]struct{})

	for index, file := range files {
		item := inspectMarkdownBundleFile(file, index+1)
		result.Items = append(result.Items, item)
		result.Total++
		if item.Skipped {
			result.Skipped++
			continue
		}
		result.Valid++
		if item.VolumeTitle != "" {
			volumeSet[item.VolumeTitle] = struct{}{}
		}
		if item.UnitTitle != "" {
			unitSet[item.VolumeTitle+"/"+item.UnitTitle] = struct{}{}
		}
	}

	result.Volumes = sortedKeys(volumeSet)
	result.Units = sortedKeys(unitSet)
	return result, nil
}

func (s *NovelService) importMarkdownBundleFile(ctx context.Context, novelID, accountID, tenantID string, file dto.MarkdownBundleFile, order int) (*vo.MarkdownBundleImportItemVO, error) {
	normalizedPath := normalizeBundlePath(file.Path)
	item := &vo.MarkdownBundleImportItemVO{
		Path:   normalizedPath,
		Action: "skipped",
	}
	if normalizedPath == "" {
		item.Skipped = true
		item.Reason = "文件路径不能为空"
		return item, nil
	}
	if !strings.EqualFold(path.Ext(normalizedPath), ".md") {
		item.Skipped = true
		item.Reason = "仅支持 Markdown 文件"
		return item, nil
	}

	parsed, err := parseMarkdownChapter(file.Content)
	if err != nil {
		item.Skipped = true
		item.Reason = err.Error()
		return item, nil
	}

	bundlePath := parseBundlePath(normalizedPath, order)
	title := firstNonEmpty(parsed.Title, bundlePath.ChapterTitle)
	if title == "" {
		item.Skipped = true
		item.Reason = "章节标题不能为空"
		return item, nil
	}

	volume, err := s.getOrCreateVolume(ctx, novelID, accountID, tenantID, bundlePath.VolumeTitle, bundlePath.VolumeOrder)
	if err != nil {
		return nil, err
	}
	item.VolumeID = volume.ID

	var unitID *string
	if bundlePath.UnitTitle != "" {
		unit, err := s.getOrCreateUnit(ctx, novelID, volume.ID, accountID, tenantID, bundlePath.UnitTitle, bundlePath.UnitOrder)
		if err != nil {
			return nil, err
		}
		unitID = &unit.ID
		item.UnitID = unitID
	}

	slug := firstNonEmpty(parsed.Slug, slugFromBundlePath(normalizedPath))
	number := firstNonEmpty(parsed.Number, bundlePath.Number)
	sortOrder := parsed.SortOrder
	if sortOrder == 0 {
		sortOrder = bundlePath.ChapterOrder
	}

	chapter, err := s.novelRepo.FindChapterBySlugAndAccount(ctx, novelID, slug, accountID, tenantID)
	if err != nil && !errors.IsNotFoundError(err) {
		return nil, err
	}
	if chapter == nil {
		created, err := s.CreateChapter(ctx, novelID, accountID, tenantID, &dto.CreateChapterRequest{
			VolumeID:  volume.ID,
			UnitID:    unitID,
			Slug:      slug,
			Number:    number,
			Title:     title,
			Summary:   parsed.Summary,
			Body:      parsed.Body,
			Status:    parsed.Status,
			SortOrder: sortOrder,
		})
		if err != nil {
			return nil, err
		}
		item.Action = "created"
		item.ChapterID = created.ID
		item.Slug = created.Slug
		item.Chapter = created
		return item, nil
	}

	status := parsed.Status
	updateReq := &dto.UpdateChapterRequest{
		VolumeID:  &volume.ID,
		UnitID:    unitIDForUpdate(unitID),
		Slug:      &slug,
		Number:    &number,
		Title:     &title,
		Summary:   &parsed.Summary,
		Body:      &parsed.Body,
		SortOrder: &sortOrder,
		Note:      "Markdown 批量导入覆盖",
	}
	if status != "" {
		updateReq.Status = &status
	}
	updated, err := s.UpdateChapter(ctx, chapter.ID, accountID, tenantID, updateReq)
	if err != nil {
		return nil, err
	}
	item.Action = "updated"
	item.ChapterID = updated.ID
	item.Slug = updated.Slug
	item.Chapter = updated
	return item, nil
}

func inspectMarkdownBundleFile(file dto.MarkdownBundleFile, order int) *vo.MarkdownBundleInspectItemVO {
	normalizedPath := normalizeBundlePath(file.Path)
	item := &vo.MarkdownBundleInspectItemVO{Path: normalizedPath}
	if normalizedPath == "" {
		item.Skipped = true
		item.Reason = "文件路径不能为空"
		return item
	}
	if !strings.EqualFold(path.Ext(normalizedPath), ".md") {
		item.Skipped = true
		item.Reason = "仅支持 Markdown 文件"
		return item
	}

	parsed, err := parseMarkdownChapter(file.Content)
	if err != nil {
		item.Skipped = true
		item.Reason = err.Error()
		return item
	}

	bundlePath := parseBundlePath(normalizedPath, order)
	title := firstNonEmpty(parsed.Title, bundlePath.ChapterTitle)
	if title == "" {
		item.Skipped = true
		item.Reason = "章节标题不能为空"
		return item
	}

	sortOrder := parsed.SortOrder
	if sortOrder == 0 {
		sortOrder = bundlePath.ChapterOrder
	}

	item.VolumeTitle = bundlePath.VolumeTitle
	item.UnitTitle = bundlePath.UnitTitle
	item.ChapterTitle = bundlePath.ChapterTitle
	item.Title = title
	item.Slug = firstNonEmpty(parsed.Slug, slugFromBundlePath(normalizedPath))
	item.Number = firstNonEmpty(parsed.Number, bundlePath.Number)
	item.Summary = parsed.Summary
	item.Status = parsed.Status
	item.SortOrder = sortOrder
	item.WordCount = countContentRunes(parsed.Body)
	return item
}

func (s *NovelService) ListChapters(ctx context.Context, p *repo.QueryParams, includeBody bool) (*vo.ListResult[vo.ChapterVO], error) {
	chapters, total, err := s.novelRepo.ListChapters(ctx, p)
	if err != nil {
		return nil, err
	}
	data := make([]*vo.ChapterVO, 0, len(chapters))
	for _, chapter := range chapters {
		data = append(data, vo.FromChapterDO(chapter, includeBody))
	}
	return &vo.ListResult[vo.ChapterVO]{Data: data, Total: total, Page: p.Page, Size: p.PageSize}, nil
}

func (s *NovelService) GetChapter(ctx context.Context, id string, includeDraft bool) (*vo.ChapterVO, error) {
	chapter, err := s.novelRepo.FindChapterByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !includeDraft && chapter.Status != constant.ChapterStatusPublished {
		return nil, errors.NewNotFoundErrorF("章节不存在: %s", id)
	}
	return vo.FromChapterDO(chapter, true), nil
}

func (s *NovelService) GetManagedChapter(ctx context.Context, id, accountID, tenantID string) (*vo.ChapterVO, error) {
	chapter, err := s.novelRepo.FindChapterByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	return vo.FromChapterDO(chapter, true), nil
}

func (s *NovelService) GetPublishedChapterBySlug(ctx context.Context, novelID, slug string) (*vo.ChapterVO, error) {
	if _, err := s.novelRepo.FindPublicNovelByID(ctx, novelID); err != nil {
		return nil, err
	}
	chapter, err := s.novelRepo.FindPublishedChapterBySlug(ctx, novelID, slug)
	if err != nil {
		return nil, err
	}
	return vo.FromChapterDO(chapter, true), nil
}

func (s *NovelService) UpdateChapter(ctx context.Context, id, accountID, tenantID string, req *dto.UpdateChapterRequest) (*vo.ChapterVO, error) {
	chapter, err := s.novelRepo.FindChapterByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	previousChapter := *chapter
	if req.VolumeID != nil {
		volume, err := s.novelRepo.FindVolumeByIDAndAccount(ctx, *req.VolumeID, accountID, tenantID)
		if err != nil {
			return nil, err
		}
		if volume.NovelID != chapter.NovelID {
			return nil, errors.NewValidationError("分卷不属于该小说")
		}
		chapter.VolumeID = *req.VolumeID
		if req.UnitID == nil {
			chapter.UnitID = nil
		}
	}
	if req.UnitID != nil {
		unitID, err := s.resolveChapterUnit(ctx, chapter.NovelID, chapter.VolumeID, req.UnitID, accountID, tenantID)
		if err != nil {
			return nil, err
		}
		chapter.UnitID = unitID
	}
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return nil, errors.NewValidationError("章节标题不能为空")
		}
		chapter.Title = strings.TrimSpace(*req.Title)
	}
	if req.Slug != nil {
		chapter.Slug = normalizeSlug(*req.Slug, chapter.ID)
	}
	if req.Number != nil {
		chapter.Number = *req.Number
	}
	if req.Summary != nil {
		chapter.Summary = *req.Summary
	}
	if req.Body != nil {
		chapter.Body = *req.Body
	}
	if req.Status != nil {
		if !vo.IsValidChapterStatus(*req.Status) {
			return nil, errors.NewValidationError("章节状态不合法")
		}
		if *req.Status == constant.ChapterStatusPublished {
			if err := s.ensureNoBlockingConflicts(ctx, chapter.NovelID); err != nil {
				return nil, err
			}
			now := time.Now()
			chapter.PublishedAt = &now
		} else {
			chapter.PublishedAt = nil
		}
		chapter.Status = *req.Status
	}
	if req.SortOrder != nil {
		chapter.SortOrder = *req.SortOrder
	}
	chapter.UpdatedAt = time.Now()
	applyChapterMetrics(chapter)
	version := dm.NewChapterVersion(&previousChapter, nextVersionLabel(chapter), req.Note)
	if err := s.novelRepo.UpdateChapterWithVersion(ctx, chapter, version); err != nil {
		return nil, err
	}
	return vo.FromChapterDO(chapter, true), nil
}

func (s *NovelService) UpdateChapterStatus(ctx context.Context, id, accountID, tenantID string, req *dto.UpdateChapterStatusRequest) (*vo.ChapterVO, error) {
	chapter, err := s.novelRepo.FindChapterByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if !vo.IsValidChapterStatus(req.Status) {
		return nil, errors.NewValidationError("章节状态不合法")
	}

	var publishedAt *time.Time
	if req.Status == constant.ChapterStatusPublished {
		if err := s.ensureNoBlockingConflicts(ctx, chapter.NovelID); err != nil {
			return nil, err
		}
		now := time.Now()
		publishedAt = &now
	}
	updatedAt := time.Now()
	if err := s.novelRepo.UpdateChapterStatus(ctx, id, accountID, tenantID, req.Status, publishedAt, updatedAt); err != nil {
		return nil, err
	}

	chapter.Status = req.Status
	chapter.PublishedAt = publishedAt
	chapter.UpdatedAt = updatedAt
	return vo.FromChapterDO(chapter, false), nil
}

func (s *NovelService) BatchUpdateChapterStatus(ctx context.Context, novelID, accountID, tenantID string, req *dto.BatchUpdateChapterStatusRequest) (*vo.BatchChapterStatusUpdateResultVO, error) {
	ids := normalizeUniqueIDs(req.IDs)
	if len(ids) == 0 {
		return nil, errors.NewValidationError("章节 ID 不能为空")
	}
	if !vo.IsValidChapterStatus(req.Status) {
		return nil, errors.NewValidationError("章节状态不合法")
	}

	chapters, err := s.novelRepo.ListChaptersByIDsAndAccount(ctx, ids, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if len(chapters) != len(ids) {
		return nil, errors.NewValidationError("存在无权访问或不存在的章节")
	}

	targetIDs := make([]string, 0, len(chapters))
	targetChapters := make([]*dm.NovelChapterDO, 0, len(chapters))
	for _, chapter := range chapters {
		if chapter.NovelID != novelID {
			return nil, errors.NewValidationError("存在不属于该小说的章节")
		}
		if chapter.Status == req.Status {
			continue
		}
		targetIDs = append(targetIDs, chapter.ID)
		targetChapters = append(targetChapters, chapter)
	}
	if len(targetIDs) == 0 {
		return &vo.BatchChapterStatusUpdateResultVO{
			Updated:  0,
			Skipped:  len(chapters),
			Chapters: []*vo.ChapterVO{},
		}, nil
	}

	var publishedAt *time.Time
	if req.Status == constant.ChapterStatusPublished {
		if err := s.ensureNoBlockingConflicts(ctx, novelID); err != nil {
			return nil, err
		}
		now := time.Now()
		publishedAt = &now
	}
	updatedAt := time.Now()
	updatedRows, err := s.novelRepo.BatchUpdateChapterStatus(ctx, targetIDs, accountID, tenantID, req.Status, publishedAt, updatedAt)
	if err != nil {
		return nil, err
	}

	resultChapters := make([]*vo.ChapterVO, 0, len(targetChapters))
	for _, chapter := range targetChapters {
		chapter.Status = req.Status
		chapter.PublishedAt = publishedAt
		chapter.UpdatedAt = updatedAt
		resultChapters = append(resultChapters, vo.FromChapterDO(chapter, false))
	}
	return &vo.BatchChapterStatusUpdateResultVO{
		Updated:  int(updatedRows),
		Skipped:  len(chapters) - int(updatedRows),
		Chapters: resultChapters,
	}, nil
}

func (s *NovelService) ListChapterVersions(ctx context.Context, chapterID string) ([]*vo.ChapterVersionVO, error) {
	versions, err := s.novelRepo.ListChapterVersions(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	data := make([]*vo.ChapterVersionVO, 0, len(versions))
	for _, version := range versions {
		data = append(data, vo.FromChapterVersionDO(version, true))
	}
	return data, nil
}

func (s *NovelService) UpsertCodexEntry(ctx context.Context, novelID, entryID, accountID, tenantID string, req *dto.UpsertCodexEntryRequest) (*vo.CodexEntryVO, error) {
	if _, err := s.novelRepo.FindNovelByIDAndAccount(ctx, novelID, accountID, tenantID); err != nil {
		return nil, err
	}
	if !vo.IsValidCodexKind(req.Kind) {
		return nil, errors.NewValidationError("资料类型不合法")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.NewValidationError("资料标题不能为空")
	}
	var entry *dm.NovelCodexEntryDO
	var err error
	if entryID == "" {
		entry = dm.NewCodexEntry(tenantID, accountID, novelID, req.Kind, strings.TrimSpace(req.Title))
	} else {
		entry, err = s.novelRepo.FindCodexEntryByIDAndAccount(ctx, entryID, accountID, tenantID)
		if err != nil {
			return nil, err
		}
		if entry.NovelID != novelID {
			return nil, errors.NewValidationError("资料条目不属于该小说")
		}
		entry.Kind = req.Kind
		entry.Title = strings.TrimSpace(req.Title)
		entry.UpdatedAt = time.Now()
	}
	entry.Summary = req.Summary
	entry.Aliases = req.Aliases
	entry.Properties = req.Properties
	entry.Relations = req.Relations
	entry.Evidence = req.Evidence
	entry.SortOrder = req.SortOrder
	if entry.Aliases == nil {
		entry.Aliases = []string{}
	}
	if entry.Properties == nil {
		entry.Properties = map[string]string{}
	}
	if entry.Relations == nil {
		entry.Relations = []string{}
	}
	if err := s.novelRepo.UpsertCodexEntry(ctx, entry); err != nil {
		return nil, err
	}
	return vo.FromCodexEntryDO(entry), nil
}

func (s *NovelService) ListCodexEntries(ctx context.Context, p *repo.QueryParams) (*vo.ListResult[vo.CodexEntryVO], error) {
	if p.Kind != "" && !vo.IsValidCodexKind(p.Kind) {
		return nil, errors.NewValidationError("资料类型不合法")
	}
	entries, total, err := s.novelRepo.ListCodexEntries(ctx, p)
	if err != nil {
		return nil, err
	}
	data := make([]*vo.CodexEntryVO, 0, len(entries))
	for _, entry := range entries {
		data = append(data, vo.FromCodexEntryDO(entry))
	}
	return &vo.ListResult[vo.CodexEntryVO]{Data: data, Total: total, Page: p.Page, Size: p.PageSize}, nil
}

func (s *NovelService) CreateRuleConflict(ctx context.Context, novelID, accountID, tenantID string, req *dto.CreateRuleConflictRequest) (*vo.RuleConflictVO, error) {
	if _, err := s.novelRepo.FindNovelByIDAndAccount(ctx, novelID, accountID, tenantID); err != nil {
		return nil, err
	}
	if !vo.IsValidRuleConflictLevel(req.Level) {
		return nil, errors.NewValidationError("冲突等级不合法")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.NewValidationError("冲突标题不能为空")
	}
	conflict := dm.NewRuleConflict(tenantID, accountID, novelID, req.TargetID, req.TargetType, req.Level, strings.TrimSpace(req.Title))
	conflict.Detail = req.Detail
	if err := s.novelRepo.CreateRuleConflict(ctx, conflict); err != nil {
		return nil, err
	}
	return vo.FromRuleConflictDO(conflict), nil
}

func (s *NovelService) ResolveRuleConflict(ctx context.Context, id, accountID, tenantID string, req *dto.ResolveRuleConflictRequest) (*vo.RuleConflictVO, error) {
	if !vo.IsValidRuleConflictStatus(req.Status) {
		return nil, errors.NewValidationError("冲突状态不合法")
	}
	conflict, err := s.novelRepo.FindRuleConflictByIDAndAccount(ctx, id, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	conflict.Status = req.Status
	conflict.Resolution = req.Resolution
	conflict.UpdatedAt = time.Now()
	if err := s.novelRepo.UpdateRuleConflict(ctx, conflict); err != nil {
		return nil, err
	}
	return vo.FromRuleConflictDO(conflict), nil
}

func (s *NovelService) ListRuleConflicts(ctx context.Context, novelID string) ([]*vo.RuleConflictVO, error) {
	conflicts, err := s.novelRepo.ListRuleConflicts(ctx, novelID)
	if err != nil {
		return nil, err
	}
	data := make([]*vo.RuleConflictVO, 0, len(conflicts))
	for _, conflict := range conflicts {
		data = append(data, vo.FromRuleConflictDO(conflict))
	}
	return data, nil
}

func filterVolumesByChapters(volumes []*dm.NovelVolumeDO, chapters []*dm.NovelChapterDO) []*dm.NovelVolumeDO {
	usedVolumeIDs := make(map[string]struct{}, len(chapters))
	for _, chapter := range chapters {
		usedVolumeIDs[chapter.VolumeID] = struct{}{}
	}
	filtered := make([]*dm.NovelVolumeDO, 0, len(volumes))
	for _, volume := range volumes {
		if _, ok := usedVolumeIDs[volume.ID]; ok {
			filtered = append(filtered, volume)
		}
	}
	return filtered
}

func filterUnitsByChapters(units []*dm.NovelUnitDO, chapters []*dm.NovelChapterDO) []*dm.NovelUnitDO {
	usedUnitIDs := make(map[string]struct{}, len(chapters))
	for _, chapter := range chapters {
		if chapter.UnitID != nil {
			usedUnitIDs[*chapter.UnitID] = struct{}{}
		}
	}
	filtered := make([]*dm.NovelUnitDO, 0, len(units))
	for _, unit := range units {
		if _, ok := usedUnitIDs[unit.ID]; ok {
			filtered = append(filtered, unit)
		}
	}
	return filtered
}

func (s *NovelService) ensureNoBlockingConflicts(ctx context.Context, novelID string) error {
	conflicts, err := s.novelRepo.ListRuleConflicts(ctx, novelID)
	if err != nil {
		return err
	}
	for _, conflict := range conflicts {
		if conflict.Level == constant.RuleConflictLevelBlocking && conflict.Status == constant.RuleConflictStatusOpen {
			return errors.NewValidationError("存在未处理的阻断级规则冲突，无法发布章节")
		}
	}
	return nil
}

func (s *NovelService) resolveChapterUnit(ctx context.Context, novelID, volumeID string, unitID *string, accountID, tenantID string) (*string, error) {
	if unitID == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*unitID)
	if trimmed == "" {
		return nil, nil
	}
	unit, err := s.novelRepo.FindUnitByIDAndAccount(ctx, trimmed, accountID, tenantID)
	if err != nil {
		return nil, err
	}
	if unit.NovelID != novelID {
		return nil, errors.NewValidationError("单元不属于该小说")
	}
	if unit.VolumeID != volumeID {
		return nil, errors.NewValidationError("单元不属于该分卷")
	}
	return &unit.ID, nil
}

func (s *NovelService) getOrCreateVolume(ctx context.Context, novelID, accountID, tenantID, title string, sortOrder int) (*dm.NovelVolumeDO, error) {
	title = firstNonEmpty(title, "默认卷")
	volume, err := s.novelRepo.FindVolumeByTitle(ctx, novelID, title, accountID, tenantID)
	if err == nil {
		return volume, nil
	}
	if !errors.IsNotFoundError(err) {
		return nil, err
	}
	volume = dm.NewVolume(tenantID, accountID, novelID, title, sortOrder)
	if err := s.novelRepo.CreateVolume(ctx, volume); err != nil {
		existingVolume, findErr := s.novelRepo.FindVolumeByTitle(ctx, novelID, title, accountID, tenantID)
		if findErr == nil {
			return existingVolume, nil
		}
		return nil, err
	}
	return volume, nil
}

func (s *NovelService) getOrCreateUnit(ctx context.Context, novelID, volumeID, accountID, tenantID, title string, sortOrder int) (*dm.NovelUnitDO, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.NewValidationError("单元标题不能为空")
	}
	unit, err := s.novelRepo.FindUnitByTitle(ctx, novelID, volumeID, title, accountID, tenantID)
	if err == nil {
		return unit, nil
	}
	if !errors.IsNotFoundError(err) {
		return nil, err
	}
	unit = dm.NewUnit(tenantID, accountID, novelID, volumeID, title, sortOrder)
	if err := s.novelRepo.CreateUnit(ctx, unit); err != nil {
		existingUnit, findErr := s.novelRepo.FindUnitByTitle(ctx, novelID, volumeID, title, accountID, tenantID)
		if findErr == nil {
			return existingUnit, nil
		}
		return nil, err
	}
	return unit, nil
}

func applyChapterMetrics(chapter *dm.NovelChapterDO) {
	chapter.WordCount = countContentRunes(chapter.Body)
	minutes := chapter.WordCount / 420
	if chapter.WordCount%420 != 0 {
		minutes++
	}
	if minutes < 1 {
		minutes = 1
	}
	chapter.ReadingMinutes = minutes
}

func countContentRunes(value string) int {
	count := 0
	for _, r := range value {
		if !unicode.IsSpace(r) {
			count++
		}
	}
	return count
}

func normalizeSlug(value, fallback string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "-")
	if value == "" {
		return fallback
	}
	return value
}

func nextVersionLabel(chapter *dm.NovelChapterDO) string {
	return fmt.Sprintf("snapshot-%s", time.Now().Format("20060102150405"))
}

type parsedMarkdownChapter struct {
	Slug      string
	Number    string
	Title     string
	Summary   string
	Status    constant.ChapterStatus
	SortOrder int
	Body      string
}

type bundlePathInfo struct {
	VolumeTitle  string
	UnitTitle    string
	ChapterTitle string
	Number       string
	VolumeOrder  int
	UnitOrder    int
	ChapterOrder int
}

func parseMarkdownChapter(markdown string) (*parsedMarkdownChapter, error) {
	markdown = normalizeMarkdownLineEndings(markdown)
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return nil, errors.NewValidationError("Markdown 内容不能为空")
	}

	metadata := map[string]string{}
	body := markdown
	if strings.HasPrefix(markdown, "---\n") || markdown == "---" {
		rest := strings.TrimPrefix(markdown, "---")
		rest = strings.TrimPrefix(rest, "\r\n")
		rest = strings.TrimPrefix(rest, "\n")
		if end := strings.Index(rest, "\n---"); end >= 0 {
			frontMatter := rest[:end]
			body = strings.TrimSpace(rest[end+len("\n---"):])
			metadata = parseFrontMatter(frontMatter)
		}
	}

	title := strings.TrimSpace(metadata["title"])
	if title == "" {
		title = extractFirstHeading(body)
		if title != "" {
			body = removeFirstHeading(body)
		}
	}

	status := constant.ChapterStatus(metadata["status"])
	if status != "" && !vo.IsValidChapterStatus(status) {
		return nil, errors.NewValidationError("Markdown front matter 中的章节状态不合法")
	}

	sortOrder := 0
	if metadata["sort_order"] != "" {
		parsedSortOrder, err := strconv.Atoi(metadata["sort_order"])
		if err != nil {
			return nil, errors.NewValidationError("Markdown front matter 中的 sort_order 不合法")
		}
		sortOrder = parsedSortOrder
	}

	return &parsedMarkdownChapter{
		Slug:      metadata["slug"],
		Number:    metadata["number"],
		Title:     title,
		Summary:   metadata["summary"],
		Status:    status,
		SortOrder: sortOrder,
		Body:      strings.TrimSpace(body),
	}, nil
}

func normalizeMarkdownLineEndings(markdown string) string {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	return strings.ReplaceAll(markdown, "\r", "\n")
}

func parseBundlePath(value string, fallbackOrder int) *bundlePathInfo {
	parts := strings.Split(value, "/")
	filename := parts[len(parts)-1]
	baseName := strings.TrimSuffix(filename, path.Ext(filename))
	number, chapterTitle, chapterOrder := splitOrderedTitle(baseName, fallbackOrder)
	info := &bundlePathInfo{
		VolumeTitle:  "默认卷",
		ChapterTitle: chapterTitle,
		Number:       number,
		VolumeOrder:  fallbackOrder,
		UnitOrder:    fallbackOrder,
		ChapterOrder: chapterOrder,
	}
	if len(parts) >= 2 {
		_, volumeTitle, volumeOrder := splitOrderedTitle(parts[0], fallbackOrder)
		info.VolumeTitle = firstNonEmpty(volumeTitle, "默认卷")
		info.VolumeOrder = volumeOrder
	}
	if len(parts) >= 3 {
		_, unitTitle, unitOrder := splitOrderedTitle(parts[1], fallbackOrder)
		info.UnitTitle = unitTitle
		info.UnitOrder = unitOrder
	}
	return info
}

func splitOrderedTitle(value string, fallbackOrder int) (string, string, int) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", fallbackOrder
	}
	runes := []rune(value)
	digitEnd := 0
	for digitEnd < len(runes) && unicode.IsDigit(runes[digitEnd]) {
		digitEnd++
	}
	if digitEnd == 0 {
		return "", value, fallbackOrder
	}
	number := string(runes[:digitEnd])
	order, err := strconv.Atoi(number)
	if err != nil {
		order = fallbackOrder
	}
	title := strings.TrimLeft(string(runes[digitEnd:]), " -_.")
	if title == "" {
		title = value
	}
	return number, strings.TrimSpace(title), order
}

func normalizeBundlePath(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	if value == "" {
		return ""
	}
	cleaned := path.Clean("/" + value)
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "." {
		return ""
	}
	return cleaned
}

func slugFromBundlePath(value string) string {
	value = strings.TrimSuffix(value, path.Ext(value))
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-", "_", "-")
	return normalizeSlug(replacer.Replace(value), value)
}

func unitIDForUpdate(unitID *string) *string {
	if unitID != nil {
		return unitID
	}
	empty := ""
	return &empty
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func normalizeUniqueIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		id := strings.TrimSpace(value)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func parseFrontMatter(value string) map[string]string {
	result := map[string]string{}
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		result[strings.TrimSpace(key)] = trimFrontMatterValue(val)
	}
	return result
}

func trimFrontMatterValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"'")
	return value
}

func extractFirstHeading(markdown string) string {
	for _, line := range strings.Split(markdown, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func removeFirstHeading(markdown string) string {
	lines := strings.Split(markdown, "\n")
	removed := false
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if !removed && strings.HasPrefix(strings.TrimSpace(line), "# ") {
			removed = true
			continue
		}
		result = append(result, line)
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
