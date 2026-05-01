package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/controller/util"
	"auth-perm/internal/domain/todo/constant"
	"auth-perm/internal/domain/todo/repo"
	"auth-perm/internal/domain/todo/service"
	"auth-perm/internal/domain/todo/vo"
)

// TodoHandler HTTP 处理器
type TodoHandler struct {
	svc *service.TodoService
}

func NewTodoHandler(svc *service.TodoService) *TodoHandler {
	return &TodoHandler{svc: svc}
}

// ---------- 分类 ----------

// ListCategories GET /api/v1/todos/categories
func (h *TodoHandler) ListCategories(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.ListCategories(c.Request.Context(), auth.AccountID, tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取分类列表失败", err.Error())
		return
	}
	response.Success(c, result)
}

// CreateCategory POST /api/v1/todos/categories
func (h *TodoHandler) CreateCategory(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	var req vo.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.CreateCategory(c.Request.Context(), auth.AccountID, tenantID, req.Name, req.Color, req.Icon)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "创建分类失败", err.Error())
		return
	}
	response.Success(c, result)
}

// UpdateCategory PUT /api/v1/todos/categories/:id
func (h *TodoHandler) UpdateCategory(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	id := c.Param("id")
	var req vo.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.UpdateCategory(c.Request.Context(), id, auth.AccountID, tenantID, req.Name, req.Color, req.Icon)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "更新分类失败", err.Error())
		return
	}
	response.Success(c, result)
}

// DeleteCategory DELETE /api/v1/todos/categories/:id
func (h *TodoHandler) DeleteCategory(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	id := c.Param("id")
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteCategory(c.Request.Context(), id, auth.AccountID, tenantID); err != nil {
		response.Error(c, http.StatusBadRequest, "删除分类失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

// ---------- 待办 ----------

// ListTodos GET /api/v1/todos
func (h *TodoHandler) ListTodos(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	params := &repo.TodoQueryParams{
		TenantID:   tenantID,
		AccountID:  auth.AccountID,
		Status:     c.Query("status"),
		Priority:   c.Query("priority"),
		CategoryID: c.Query("category_id"),
		Keyword:    c.Query("keyword"),
		Page:       page,
		PageSize:   pageSize,
	}

	result, err := h.svc.ListTodos(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取待办列表失败", err.Error())
		return
	}
	response.Success(c, result)
}

// CreateTodo POST /api/v1/todos
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	var req vo.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if req.Priority != "" && !validateTodoPriority(req.Priority) {
		respondValidationError(c, "优先级不合法")
		return
	}
	params := &service.CreateTodoParams{
		TenantID:    tenantID,
		AccountID:   auth.AccountID,
		CategoryID:  req.CategoryID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Deadline:    req.Deadline,
	}
	if params.Priority == "" {
		params.Priority = constant.TodoPriorityMedium
	}

	result, err := h.svc.CreateTodo(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "创建待办失败", err.Error())
		return
	}
	response.Success(c, result)
}

// GetTodo GET /api/v1/todos/:id
func (h *TodoHandler) GetTodo(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	id := c.Param("id")
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	result, err := h.svc.GetTodo(c.Request.Context(), id, auth.AccountID, tenantID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "待办不存在", err.Error())
		return
	}
	response.Success(c, result)
}

// UpdateTodo PUT /api/v1/todos/:id
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	id := c.Param("id")
	var req vo.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	var catID *string
	if req.ClearCategory {
		empty := ""
		catID = &empty
	} else {
		catID = req.CategoryID
	}

	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if req.Priority != nil && !validateTodoPriority(*req.Priority) {
		respondValidationError(c, "优先级不合法")
		return
	}
	params := &service.UpdateTodoParams{
		ID:            id,
		TenantID:      tenantID,
		AccountID:     auth.AccountID,
		CategoryID:    catID,
		Title:         req.Title,
		Description:   req.Description,
		Priority:      req.Priority,
		Deadline:      req.Deadline,
		ClearDeadline: req.ClearDeadline,
	}

	result, err := h.svc.UpdateTodo(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "更新待办失败", err.Error())
		return
	}
	response.Success(c, result)
}

// UpdateTodoStatus PATCH /api/v1/todos/:id/status
func (h *TodoHandler) UpdateTodoStatus(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	id := c.Param("id")
	var req vo.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if !validateTodoStatus(req.Status) {
		respondValidationError(c, "状态不合法")
		return
	}
	result, err := h.svc.UpdateTodoStatus(c.Request.Context(), id, auth.AccountID, tenantID, req.Status)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "更新状态失败", err.Error())
		return
	}
	response.Success(c, result)
}

// UpdateTodoPriority PATCH /api/v1/todos/:id/priority
func (h *TodoHandler) UpdateTodoPriority(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	id := c.Param("id")
	var req vo.UpdatePriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if !validateTodoPriority(req.Priority) {
		respondValidationError(c, "优先级不合法")
		return
	}
	result, err := h.svc.UpdateTodoPriority(c.Request.Context(), id, auth.AccountID, tenantID, req.Priority)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "更新优先级失败", err.Error())
		return
	}
	response.Success(c, result)
}

// DeleteTodo DELETE /api/v1/todos/:id
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	auth, err := util.GetAuthInfo(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	id := c.Param("id")
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteTodo(c.Request.Context(), id, auth.AccountID, tenantID); err != nil {
		response.Error(c, http.StatusBadRequest, "删除待办失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

// requireTenantID 获取 tenant_id 并校验必填
func requireTenantID(c *gin.Context) (string, bool) {
	tenantID := getTenantID(c)
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "租户ID不能为空", "")
		return "", false
	}
	return tenantID, true
}

// getTenantID 从 context 中提取 tenant_id（优先 query，其次 middleware 注入）
func getTenantID(c *gin.Context) string {
	if tid := c.Query("tenant_id"); tid != "" {
		return tid
	}
	if v, exists := c.Get("tenant_id"); exists {
		if tid, ok := v.(string); ok {
			return tid
		}
	}
	return ""
}

// respondValidationError 统一验证错误响应（422）
func respondValidationError(c *gin.Context, msg string) {
	response.Error(c, http.StatusUnprocessableEntity, msg, "")
}

// validateTodoStatus 校验状态值
func validateTodoStatus(status constant.TodoStatus) bool {
	switch status {
	case constant.TodoStatusPending, constant.TodoStatusInProgress, constant.TodoStatusCompleted, constant.TodoStatusCancelled:
		return true
	default:
		return false
	}
}

// validateTodoPriority 校验优先级值
func validateTodoPriority(priority constant.TodoPriority) bool {
	switch priority {
	case constant.TodoPriorityLow, constant.TodoPriorityMedium, constant.TodoPriorityHigh, constant.TodoPriorityUrgent:
		return true
	default:
		return false
	}
}
