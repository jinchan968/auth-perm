package handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/controller/util"
	"auth-perm/internal/domain/novel/constant"
	"auth-perm/internal/domain/novel/dto"
	"auth-perm/internal/domain/novel/repo"
	"auth-perm/internal/domain/novel/service"
)

type NovelHandler struct {
	svc *service.NovelService
}

const (
	maxMarkdownFileBytes           int64 = 2 * 1024 * 1024
	maxMarkdownBundleBytes         int64 = 30 * 1024 * 1024
	maxMarkdownBundleFiles               = 1000
	maxMarkdownBundleExpandedBytes       = 50 * 1024 * 1024
)

func NewNovelHandler(svc *service.NovelService) *NovelHandler {
	return &NovelHandler{svc: svc}
}

func (h *NovelHandler) ListPublicNovels(c *gin.Context) {
	page, pageSize := pagination(c)
	result, err := h.svc.ListPublicNovels(c.Request.Context(), &repo.QueryParams{
		TenantID: c.DefaultQuery("tenant_id", "default"),
		Status:   c.Query("status"),
		Keyword:  c.Query("keyword"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.ErrorWithMessage(c, err, "获取小说列表失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) GetPublicNovel(c *gin.Context) {
	result, err := h.svc.GetNovelDetail(c.Request.Context(), c.Param("id"), false)
	if err != nil {
		response.ErrorWithMessage(c, err, "获取小说详情失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) GetPublicChapterBySlug(c *gin.Context) {
	result, err := h.svc.GetPublishedChapterBySlug(c.Request.Context(), c.Param("id"), c.Param("slug"))
	if err != nil {
		response.ErrorWithMessage(c, err, "获取章节失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) ListNovels(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	page, pageSize := pagination(c)
	result, err := h.svc.ListNovels(c.Request.Context(), &repo.QueryParams{
		TenantID:  tenantID,
		AccountID: auth.AccountID,
		Status:    c.Query("status"),
		Keyword:   c.Query("keyword"),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		response.ErrorWithMessage(c, err, "获取小说列表失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) CreateNovel(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.CreateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.CreateNovel(c.Request.Context(), tenantID, auth.AccountID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "创建小说失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) GetNovel(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.GetManagedNovelDetail(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID)
	if err != nil {
		response.ErrorWithMessage(c, err, "获取小说详情失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) UpdateNovel(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.UpdateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.UpdateNovel(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "更新小说失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) DeleteNovel(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteNovel(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID); err != nil {
		response.ErrorWithMessage(c, err, "删除小说失败")
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *NovelHandler) ListVolumes(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.ListManagedVolumes(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID)
	if err != nil {
		response.ErrorWithMessage(c, err, "获取分卷列表失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) CreateVolume(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.CreateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.CreateVolume(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "创建分卷失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) UpdateVolume(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.UpdateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.UpdateVolume(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "更新分卷失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) ListUnits(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.ListManagedUnits(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID)
	if err != nil {
		response.ErrorWithMessage(c, err, "获取单元列表失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) CreateUnit(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.CreateUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.CreateUnit(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "创建单元失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) UpdateUnit(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.UpdateUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.UpdateUnit(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "更新单元失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) ListChapters(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if err := h.svc.EnsureNovelAccess(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID); err != nil {
		response.ErrorWithMessage(c, err, "无权访问小说")
		return
	}
	page, pageSize := pagination(c)
	result, err := h.svc.ListChapters(c.Request.Context(), &repo.QueryParams{
		NovelID:  c.Param("id"),
		Status:   c.Query("status"),
		Keyword:  c.Query("keyword"),
		Page:     page,
		PageSize: pageSize,
	}, c.Query("include_body") == "true")
	if err != nil {
		response.ErrorWithMessage(c, err, "获取章节列表失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) CreateChapter(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.CreateChapter(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "创建章节失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) ImportMarkdownChapter(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	req, err := bindMarkdownImportRequest(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	result, err := h.svc.ImportMarkdownChapter(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, req)
	if err != nil {
		response.ErrorWithMessage(c, err, "导入 Markdown 章节失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) ImportMarkdownBundle(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	req, err := bindMarkdownBundleImportRequest(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	taskID := h.svc.ImportMarkdownBundleAsync(c.Param("id"), auth.AccountID, tenantID, req)
	response.Success(c, gin.H{"task_id": taskID, "status": "pending"})
}

func (h *NovelHandler) GetImportTask(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	task, err := h.svc.GetImportTask(c.Request.Context(), c.Param("taskId"), auth.AccountID, tenantID)
	if err != nil {
		response.ErrorWithMessage(c, err, "获取任务状态失败")
		return
	}
	response.Success(c, task)
}

func (h *NovelHandler) InspectMarkdownBundle(c *gin.Context) {
	req, err := bindMarkdownBundleImportRequest(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.InspectMarkdownBundle(c.Request.Context(), &dto.InspectMarkdownBundleRequest{Files: req.Files})
	if err != nil {
		response.ErrorWithMessage(c, err, "识别 Markdown 文件树失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) GetChapter(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.GetManagedChapter(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID)
	if err != nil {
		response.ErrorWithMessage(c, err, "获取章节失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) UpdateChapter(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.UpdateChapter(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "更新章节失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) UpdateChapterStatus(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.UpdateChapterStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.UpdateChapterStatus(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "更新章节状态失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) BatchUpdateChapterStatus(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.BatchUpdateChapterStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.BatchUpdateChapterStatus(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "批量更新章节状态失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) ListChapterVersions(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if _, err := h.svc.GetManagedChapter(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID); err != nil {
		response.ErrorWithMessage(c, err, "无权访问章节")
		return
	}
	result, err := h.svc.ListChapterVersions(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorWithMessage(c, err, "获取章节版本失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) ListCodexEntries(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if err := h.svc.EnsureNovelAccess(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID); err != nil {
		response.ErrorWithMessage(c, err, "无权访问小说")
		return
	}
	page, pageSize := pagination(c)
	result, err := h.svc.ListCodexEntries(c.Request.Context(), &repo.QueryParams{
		NovelID:  c.Param("id"),
		Kind:     constant.CodexKind(c.Query("kind")),
		Keyword:  c.Query("keyword"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.ErrorWithMessage(c, err, "获取资料库失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) CreateCodexEntry(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.UpsertCodexEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.UpsertCodexEntry(c.Request.Context(), c.Param("id"), "", auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "保存资料条目失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) UpdateCodexEntry(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.UpsertCodexEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.UpsertCodexEntry(c.Request.Context(), c.Param("id"), c.Param("entryId"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "保存资料条目失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) ListRuleConflicts(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if err := h.svc.EnsureNovelAccess(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID); err != nil {
		response.ErrorWithMessage(c, err, "无权访问小说")
		return
	}
	result, err := h.svc.ListRuleConflicts(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorWithMessage(c, err, "获取规则冲突失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) CreateRuleConflict(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.CreateRuleConflictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.CreateRuleConflict(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "创建规则冲突失败")
		return
	}
	response.Success(c, result)
}

func (h *NovelHandler) ResolveRuleConflict(c *gin.Context) {
	auth, ok := authInfo(c)
	if !ok {
		return
	}
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	var req dto.ResolveRuleConflictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	result, err := h.svc.ResolveRuleConflict(c.Request.Context(), c.Param("id"), auth.AccountID, tenantID, &req)
	if err != nil {
		response.ErrorWithMessage(c, err, "处理规则冲突失败")
		return
	}
	response.Success(c, result)
}

func authInfo(c *gin.Context) (*util.AuthInfo, bool) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return nil, false
	}
	return auth, true
}

func requireTenantID(c *gin.Context) (string, bool) {
	tenantID, err := util.GetTenantID(c)
	if err != nil || tenantID == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return "", false
	}
	return tenantID, true
}

func pagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func bindMarkdownImportRequest(c *gin.Context) (*dto.ImportMarkdownChapterRequest, error) {
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		return bindMultipartMarkdownImportRequest(c)
	}

	var req dto.ImportMarkdownChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func bindMultipartMarkdownImportRequest(c *gin.Context) (*dto.ImportMarkdownChapterRequest, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, err
	}
	opened, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer opened.Close()

	content, err := readLimited(opened, maxMarkdownFileBytes, "Markdown 文件不能超过 2MB")
	if err != nil {
		return nil, err
	}

	sortOrder, _ := strconv.Atoi(c.PostForm("sort_order"))
	return &dto.ImportMarkdownChapterRequest{
		VolumeID:  c.PostForm("volume_id"),
		UnitID:    optionalStringPtr(c.PostForm("unit_id")),
		Markdown:  string(content),
		Slug:      c.PostForm("slug"),
		Number:    c.PostForm("number"),
		Title:     c.PostForm("title"),
		Summary:   c.PostForm("summary"),
		Status:    constant.ChapterStatus(c.PostForm("status")),
		SortOrder: sortOrder,
	}, nil
}

func bindMarkdownBundleImportRequest(c *gin.Context) (*dto.ImportMarkdownBundleRequest, error) {
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		return bindMultipartMarkdownBundleImportRequest(c)
	}

	var req dto.ImportMarkdownBundleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func bindMultipartMarkdownBundleImportRequest(c *gin.Context) (*dto.ImportMarkdownBundleRequest, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, err
	}
	opened, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer opened.Close()

	content, err := readLimited(opened, maxMarkdownBundleBytes, "Markdown zip 文件不能超过 30MB")
	if err != nil {
		return nil, err
	}
	files, err := readMarkdownBundleZip(content)
	if err != nil {
		return nil, err
	}
	return &dto.ImportMarkdownBundleRequest{Files: files}, nil
}

func readMarkdownBundleZip(content []byte) ([]dto.MarkdownBundleFile, error) {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, err
	}

	files := make([]dto.MarkdownBundleFile, 0, len(reader.File))
	var expandedBytes int64
	for _, zipped := range reader.File {
		if zipped.FileInfo().IsDir() || !strings.HasSuffix(strings.ToLower(strings.TrimSpace(zipped.Name)), ".md") {
			continue
		}
		if len(files) >= maxMarkdownBundleFiles {
			return nil, fmt.Errorf("Markdown zip 文件数量不能超过 %d", maxMarkdownBundleFiles)
		}
		if int64(zipped.UncompressedSize64) > maxMarkdownFileBytes {
			return nil, fmt.Errorf("zip 内单个 Markdown 文件不能超过 2MB")
		}
		if expandedBytes+int64(zipped.UncompressedSize64) > maxMarkdownBundleExpandedBytes {
			return nil, fmt.Errorf("Markdown zip 解压后总大小不能超过 50MB")
		}
		zippedFile, err := zipped.Open()
		if err != nil {
			return nil, err
		}
		body, readErr := readLimited(zippedFile, maxMarkdownFileBytes, "zip 内单个 Markdown 文件不能超过 2MB")
		closeErr := zippedFile.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		expandedBytes += int64(len(body))
		if expandedBytes > maxMarkdownBundleExpandedBytes {
			return nil, fmt.Errorf("Markdown zip 解压后总大小不能超过 50MB")
		}
		files = append(files, dto.MarkdownBundleFile{
			Path:    zipped.Name,
			Content: string(body),
		})
	}
	return files, nil
}

func optionalStringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func readLimited(reader io.Reader, limit int64, message string) ([]byte, error) {
	content, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > limit {
		return nil, fmt.Errorf("%s", message)
	}
	return content, nil
}
