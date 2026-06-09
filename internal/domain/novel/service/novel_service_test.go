package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"auth-perm/internal/domain/novel/constant"
	"auth-perm/internal/domain/novel/dm"
	"auth-perm/internal/domain/novel/dto"
	"auth-perm/internal/domain/novel/repo"
)

func TestNovelServiceBlocksPublishWhenBlockingConflictOpen(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{
		Title:  "Echo",
		Status: constant.NovelStatusSerial,
	})
	require.NoError(t, err)

	volume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{
		Title: "Vol. 1",
	})
	require.NoError(t, err)

	chapter, err := svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
		VolumeID: volume.ID,
		Title:    "序章",
		Body:     "他第一次听见了回声。",
	})
	require.NoError(t, err)

	_, err = svc.CreateRuleConflict(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateRuleConflictRequest{
		TargetID:   chapter.ID,
		TargetType: "chapter",
		Level:      constant.RuleConflictLevelBlocking,
		Title:      "无代价答案",
	})
	require.NoError(t, err)

	published := constant.ChapterStatusPublished
	_, err = svc.UpdateChapter(ctx, chapter.ID, "account-1", "tenant-1", &dto.UpdateChapterRequest{
		Status: &published,
		Note:   "尝试发布",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "阻断级规则冲突")
}

func TestNovelServiceBlocksCreatePublishedChapterWhenBlockingConflictOpen(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	volume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 1"})
	require.NoError(t, err)
	_, err = svc.CreateRuleConflict(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateRuleConflictRequest{
		TargetID:   novel.ID,
		TargetType: "novel",
		Level:      constant.RuleConflictLevelBlocking,
		Title:      "世界规则未闭合",
	})
	require.NoError(t, err)

	_, err = svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
		VolumeID: volume.ID,
		Title:    "序章",
		Body:     "他第一次听见了回声。",
		Status:   constant.ChapterStatusPublished,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "阻断级规则冲突")
}

func TestNovelServiceHidesDraftNovelFromPublicDetail(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{
		Title:  "Echo",
		Status: constant.NovelStatusDraft,
	})
	require.NoError(t, err)

	_, err = svc.GetNovelDetail(ctx, novel.ID, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "小说不存在")
}

func TestNovelServicePublicListRejectsDraftStatusFilter(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	_, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{
		Title:  "Draft Echo",
		Status: constant.NovelStatusDraft,
	})
	require.NoError(t, err)
	_, err = svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{
		Title:  "Serial Echo",
		Status: constant.NovelStatusSerial,
	})
	require.NoError(t, err)
	_, err = svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{
		Title:  "Completed Echo",
		Status: constant.NovelStatusCompleted,
	})
	require.NoError(t, err)

	result, err := svc.ListPublicNovels(ctx, &repo.QueryParams{
		TenantID: "tenant-1",
		Status:   string(constant.NovelStatusDraft),
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	require.EqualValues(t, 2, result.Total)
	for _, novel := range result.Data {
		require.NotEqual(t, constant.NovelStatusDraft, novel.Status)
	}
}

func TestNovelServiceHidesPublishedChapterWhenNovelDraft(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{
		Title:  "Draft Echo",
		Status: constant.NovelStatusDraft,
	})
	require.NoError(t, err)
	volume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 1"})
	require.NoError(t, err)
	chapter, err := svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
		VolumeID: volume.ID,
		Slug:     "published-in-draft",
		Title:    "草稿小说里的已发布章",
		Body:     "不应该公开。",
		Status:   constant.ChapterStatusPublished,
	})
	require.NoError(t, err)

	_, err = svc.GetPublishedChapterBySlug(ctx, novel.ID, chapter.Slug)
	require.Error(t, err)
	require.Contains(t, err.Error(), "小说不存在")
}

func TestNovelServicePublicDetailHidesDraftStructure(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{
		Title:  "Echo",
		Status: constant.NovelStatusSerial,
	})
	require.NoError(t, err)
	draftVolume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Draft Vol"})
	require.NoError(t, err)
	draftUnit, err := svc.CreateUnit(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateUnitRequest{
		VolumeID: draftVolume.ID,
		Title:    "Draft Unit",
	})
	require.NoError(t, err)
	_, err = svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
		VolumeID: draftVolume.ID,
		UnitID:   &draftUnit.ID,
		Title:    "草稿章",
		Body:     "尚未发布。",
	})
	require.NoError(t, err)
	publishedVolume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Public Vol"})
	require.NoError(t, err)
	publishedUnit, err := svc.CreateUnit(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateUnitRequest{
		VolumeID: publishedVolume.ID,
		Title:    "Public Unit",
	})
	require.NoError(t, err)
	_, err = svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
		VolumeID: publishedVolume.ID,
		UnitID:   &publishedUnit.ID,
		Title:    "公开章",
		Body:     "已经发布。",
		Status:   constant.ChapterStatusPublished,
	})
	require.NoError(t, err)

	detail, err := svc.GetNovelDetail(ctx, novel.ID, false)
	require.NoError(t, err)
	require.Len(t, detail.Volumes, 1)
	require.Len(t, detail.Units, 1)
	require.Len(t, detail.Chapters, 1)
	require.Equal(t, "Public Vol", detail.Volumes[0].Title)
	require.Equal(t, "Public Unit", detail.Units[0].Title)
}

func TestNovelServiceRejectsMovingUnitWithChapters(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	volume1, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 1"})
	require.NoError(t, err)
	volume2, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 2"})
	require.NoError(t, err)
	unit, err := svc.CreateUnit(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateUnitRequest{
		VolumeID: volume1.ID,
		Title:    "Unit 1",
	})
	require.NoError(t, err)
	_, err = svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
		VolumeID: volume1.ID,
		UnitID:   &unit.ID,
		Title:    "序章",
	})
	require.NoError(t, err)

	_, err = svc.UpdateUnit(ctx, unit.ID, "account-1", "tenant-1", &dto.UpdateUnitRequest{
		VolumeID: &volume2.ID,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "已有章节")
}

func TestNovelServiceRejectsVolumeListForOtherAccount(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	_, err = svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 1"})
	require.NoError(t, err)

	_, err = svc.ListManagedVolumes(ctx, novel.ID, "account-2", "tenant-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "小说不存在")
}

func TestNovelServiceManagedDetailReturnsMoreThanOneHundredChapters(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	volume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 1"})
	require.NoError(t, err)
	for i := 0; i < 105; i++ {
		_, err = svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
			VolumeID:  volume.ID,
			Title:     "章节",
			Body:      "正文",
			SortOrder: i + 1,
		})
		require.NoError(t, err)
	}

	detail, err := svc.GetManagedNovelDetail(ctx, novel.ID, "account-1", "tenant-1")
	require.NoError(t, err)
	require.Len(t, detail.Chapters, 105)
}

func TestNovelServicePublishesAfterConflictResolved(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	volume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 1"})
	require.NoError(t, err)
	chapter, err := svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
		VolumeID: volume.ID,
		Title:    "序章",
		Body:     "他第一次听见了回声。",
	})
	require.NoError(t, err)
	conflict, err := svc.CreateRuleConflict(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateRuleConflictRequest{
		TargetID:   chapter.ID,
		TargetType: "chapter",
		Level:      constant.RuleConflictLevelBlocking,
		Title:      "无代价答案",
	})
	require.NoError(t, err)

	_, err = svc.ResolveRuleConflict(ctx, conflict.ID, "account-1", "tenant-1", &dto.ResolveRuleConflictRequest{
		Status:     constant.RuleConflictStatusResolved,
		Resolution: "补充代价说明",
	})
	require.NoError(t, err)

	published := constant.ChapterStatusPublished
	updated, err := svc.UpdateChapter(ctx, chapter.ID, "account-1", "tenant-1", &dto.UpdateChapterRequest{
		Status: &published,
		Note:   "发布",
	})
	require.NoError(t, err)
	require.Equal(t, constant.ChapterStatusPublished, updated.Status)
	require.NotNil(t, updated.PublishedAt)
}

func TestNovelServiceClearsPublishedAtWhenChapterWithdrawn(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	volume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 1"})
	require.NoError(t, err)
	chapter, err := svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
		VolumeID: volume.ID,
		Title:    "序章",
		Body:     "他第一次听见了回声。",
		Status:   constant.ChapterStatusPublished,
	})
	require.NoError(t, err)
	require.NotNil(t, chapter.PublishedAt)

	draft := constant.ChapterStatusDraft
	updated, err := svc.UpdateChapterStatus(ctx, chapter.ID, "account-1", "tenant-1", &dto.UpdateChapterStatusRequest{
		Status: draft,
		Note:   "撤回",
	})
	require.NoError(t, err)
	require.Equal(t, constant.ChapterStatusDraft, updated.Status)
	require.Nil(t, updated.PublishedAt)
	require.Empty(t, updated.Body)
}

func TestNovelServiceBatchUpdatesChapterStatusWithoutBody(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	volume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 1"})
	require.NoError(t, err)
	first, err := svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
		VolumeID: volume.ID,
		Title:    "第一章",
		Body:     "第一章正文。",
	})
	require.NoError(t, err)
	second, err := svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
		VolumeID: volume.ID,
		Title:    "第二章",
		Body:     "第二章正文。",
	})
	require.NoError(t, err)
	published, err := svc.CreateChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateChapterRequest{
		VolumeID: volume.ID,
		Title:    "已发布章",
		Body:     "已发布正文。",
		Status:   constant.ChapterStatusPublished,
	})
	require.NoError(t, err)

	result, err := svc.BatchUpdateChapterStatus(ctx, novel.ID, "account-1", "tenant-1", &dto.BatchUpdateChapterStatusRequest{
		IDs:    []string{first.ID, second.ID, published.ID, first.ID},
		Status: constant.ChapterStatusPublished,
		Note:   "批量发布",
	})
	require.NoError(t, err)
	require.Equal(t, 2, result.Updated)
	require.Equal(t, 1, result.Skipped)
	require.Len(t, result.Chapters, 2)
	for _, chapter := range result.Chapters {
		require.Equal(t, constant.ChapterStatusPublished, chapter.Status)
		require.NotNil(t, chapter.PublishedAt)
		require.Empty(t, chapter.Body)
	}
}

func TestNovelServiceImportsMarkdownChapterWithFrontMatter(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	volume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 1"})
	require.NoError(t, err)

	chapter, err := svc.ImportMarkdownChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.ImportMarkdownChapterRequest{
		VolumeID: volume.ID,
		Markdown: `---
title: 序章
slug: prologue
number: "00"
summary: 第一次回声。
status: draft
sort_order: 7
---

林澈第一次听见了回声。

它没有回答，只是反问。`,
	})
	require.NoError(t, err)
	require.Equal(t, "序章", chapter.Title)
	require.Equal(t, "prologue", chapter.Slug)
	require.Equal(t, "00", chapter.Number)
	require.Equal(t, "第一次回声。", chapter.Summary)
	require.Equal(t, constant.ChapterStatusDraft, chapter.Status)
	require.Equal(t, 7, chapter.SortOrder)
	require.Contains(t, chapter.Body, "林澈第一次听见了回声。")
	require.NotContains(t, chapter.Body, "title:")
}

func TestNovelServiceImportsMarkdownChapterWithCRLFFrontMatter(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	volume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 1"})
	require.NoError(t, err)

	chapter, err := svc.ImportMarkdownChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.ImportMarkdownChapterRequest{
		VolumeID: volume.ID,
		Markdown: "---\r\ntitle: 雨夜\r\nslug: rainy-night\r\nstatus: draft\r\n---\r\n\r\n正文第一段。",
	})
	require.NoError(t, err)
	require.Equal(t, "雨夜", chapter.Title)
	require.Equal(t, "rainy-night", chapter.Slug)
	require.Equal(t, "正文第一段。", chapter.Body)
	require.NotContains(t, chapter.Body, "title:")
}

func TestNovelServiceImportsMarkdownChapterWithHeadingTitle(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	volume, err := svc.CreateVolume(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateVolumeRequest{Title: "Vol. 1"})
	require.NoError(t, err)

	chapter, err := svc.ImportMarkdownChapter(ctx, novel.ID, "account-1", "tenant-1", &dto.ImportMarkdownChapterRequest{
		VolumeID: volume.ID,
		Markdown: "# 第1章：试探\n\n他输入了第一个问题。",
	})
	require.NoError(t, err)
	require.Equal(t, "第1章：试探", chapter.Title)
	require.Equal(t, "他输入了第一个问题。", chapter.Body)
}

func TestNovelServiceImportsMarkdownBundleWithVolumeUnitChapters(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)

	result, err := svc.ImportMarkdownBundle(ctx, novel.ID, "account-1", "tenant-1", &dto.ImportMarkdownBundleRequest{
		Files: []dto.MarkdownBundleFile{
			{
				Path: "第一卷-旧城/01-初遇/001-序章.md",
				Content: `---
title: 序章
slug: prologue
---

林澈第一次听见了回声。`,
			},
			{
				Path:    "第一卷-旧城/01-初遇/002-试探.md",
				Content: "# 试探\n\n他输入了第一个问题。",
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, result.Imported)
	require.Equal(t, 2, result.Created)
	require.Equal(t, 0, result.Updated)
	require.Equal(t, 0, result.Skipped)

	detail, err := svc.GetManagedNovelDetail(ctx, novel.ID, "account-1", "tenant-1")
	require.NoError(t, err)
	require.Len(t, detail.Volumes, 1)
	require.Len(t, detail.Units, 1)
	require.Len(t, detail.Chapters, 2)
	require.Equal(t, "第一卷-旧城", detail.Volumes[0].Title)
	require.Equal(t, "初遇", detail.Units[0].Title)
	for _, chapter := range detail.Chapters {
		require.NotNil(t, chapter.UnitID)
		require.Equal(t, detail.Units[0].ID, *chapter.UnitID)
	}
}

func TestNovelServiceInspectsMarkdownBundlePaths(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	result, err := svc.InspectMarkdownBundle(ctx, &dto.InspectMarkdownBundleRequest{
		Files: []dto.MarkdownBundleFile{
			{
				Path: "第一卷-旧城/01-初遇/001-序章.md",
				Content: `---
title: 序章
slug: prologue
number: "000"
sort_order: 9
---

正文。`,
			},
			{
				Path:    "第二卷-群星/002-抵达.md",
				Content: "# 抵达\n\n他们来到群星之下。",
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, result.Total)
	require.Equal(t, 2, result.Valid)
	require.Equal(t, 0, result.Skipped)
	require.Equal(t, []string{"第一卷-旧城", "第二卷-群星"}, result.Volumes)
	require.Equal(t, []string{"第一卷-旧城/初遇"}, result.Units)

	first := result.Items[0]
	require.Equal(t, "第一卷-旧城", first.VolumeTitle)
	require.Equal(t, "初遇", first.UnitTitle)
	require.Equal(t, "序章", first.Title)
	require.Equal(t, "prologue", first.Slug)
	require.Equal(t, "000", first.Number)
	require.Equal(t, 9, first.SortOrder)

	second := result.Items[1]
	require.Equal(t, "第二卷-群星", second.VolumeTitle)
	require.Empty(t, second.UnitTitle)
	require.Equal(t, "抵达", second.Title)
	require.Equal(t, "002", second.Number)
}

func TestNovelServiceUpdatesMarkdownBundleChapterBySlug(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	req := &dto.ImportMarkdownBundleRequest{
		Files: []dto.MarkdownBundleFile{
			{
				Path: "第一卷/01-初遇/001-序章.md",
				Content: `---
title: 序章
slug: prologue
---

旧内容。`,
			},
		},
	}
	result, err := svc.ImportMarkdownBundle(ctx, novel.ID, "account-1", "tenant-1", req)
	require.NoError(t, err)
	require.Equal(t, 1, result.Created)
	chapterID := result.Items[0].ChapterID

	req.Files[0].Content = `---
title: 序章
slug: prologue
---

新内容。`
	result, err = svc.ImportMarkdownBundle(ctx, novel.ID, "account-1", "tenant-1", req)
	require.NoError(t, err)
	require.Equal(t, 0, result.Created)
	require.Equal(t, 1, result.Updated)
	require.Equal(t, chapterID, result.Items[0].ChapterID)
	require.Contains(t, result.Items[0].Chapter.Body, "新内容")

	versions, err := svc.ListChapterVersions(ctx, chapterID)
	require.NoError(t, err)
	require.Len(t, versions, 1)
	require.Contains(t, versions[0].Body, "旧内容")
}

func TestNovelServiceRollsBackMarkdownBundleWhenWriteFails(t *testing.T) {
	svc := newTestNovelService(t)
	ctx := context.Background()

	novel, err := svc.CreateNovel(ctx, "tenant-1", "account-1", &dto.CreateNovelRequest{Title: "Echo"})
	require.NoError(t, err)
	_, err = svc.CreateRuleConflict(ctx, novel.ID, "account-1", "tenant-1", &dto.CreateRuleConflictRequest{
		TargetID:   novel.ID,
		TargetType: "novel",
		Level:      constant.RuleConflictLevelBlocking,
		Title:      "规则未闭合",
	})
	require.NoError(t, err)

	_, err = svc.ImportMarkdownBundle(ctx, novel.ID, "account-1", "tenant-1", &dto.ImportMarkdownBundleRequest{
		Files: []dto.MarkdownBundleFile{
			{
				Path:    "第一卷/001-草稿.md",
				Content: "# 草稿\n\n这章本来会先写入。",
			},
			{
				Path: "第一卷/002-发布.md",
				Content: `---
title: 发布
status: published
---

这章会触发阻断冲突。`,
			},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "阻断级规则冲突")

	detail, err := svc.GetManagedNovelDetail(ctx, novel.ID, "account-1", "tenant-1")
	require.NoError(t, err)
	require.Empty(t, detail.Volumes)
	require.Empty(t, detail.Chapters)
}

func newTestNovelService(t *testing.T) *NovelService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&dm.NovelDO{},
		&dm.NovelVolumeDO{},
		&dm.NovelUnitDO{},
		&dm.NovelChapterDO{},
		&dm.NovelChapterVersionDO{},
		&dm.NovelCodexEntryDO{},
		&dm.NovelRuleConflictDO{},
	))

	return NewNovelService(repo.NewNovelRepo(db))
}
