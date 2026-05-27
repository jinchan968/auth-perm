package handler

import (
	"net/http"
	"strconv"

	"auth-perm/internal/common/dto/response"
	"auth-perm/internal/controller/util"
	"auth-perm/internal/domain/workflow/constant"
	"auth-perm/internal/domain/workflow/dto"
	"auth-perm/internal/domain/workflow/service"
	"auth-perm/internal/domain/workflow/vo"

	"github.com/gin-gonic/gin"
)

type WorkflowHandler struct {
	svc *service.WorkflowService
}

func NewWorkflowHandler(svc *service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{svc: svc}
}

func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	var req vo.ListWorkflowsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	accountID, err := util.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	offset := (req.Page - 1) * req.Size
	limit := req.Size

	var list []*dto.WorkflowDTO
	var total int64

	if req.Type == constant.TypeTemplate {
		list, err = h.svc.ListTemplates(req.TenantID)
		total = int64(len(list))
	} else {
		list, total, err = h.svc.ListWorkflows(req.TenantID, accountID, offset, limit)
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败", err.Error())
		return
	}

	response.Success(c, vo.WorkflowListResponse{
		Data:  h.toResponseList(list),
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	})
}

func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var req vo.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	accountID, err := util.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	dto, err := h.svc.CreateWorkflow(req.TenantID, accountID, req.Name, req.Description, string(req.FlowJSON), req.TemplateID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败", err.Error())
		return
	}

	response.Success(c, h.toResponse(dto))
}

func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "tenant_id 不能为空", "")
		return
	}

	dto, err := h.svc.GetWorkflow(id, tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败", err.Error())
		return
	}

	response.Success(c, h.toResponse(dto))
}

func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	id := c.Param("id")
	var req vo.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	var flowJSON *string
	if req.FlowJSON != nil {
		s := string(*req.FlowJSON)
		flowJSON = &s
	}
	dto, err := h.svc.UpdateWorkflow(id, req.TenantID, req.Name, req.Description, flowJSON, req.Status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败", err.Error())
		return
	}

	response.Success(c, h.toResponse(dto))
}

func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "tenant_id 不能为空", "")
		return
	}

	if err := h.svc.DeleteWorkflow(id, tenantID); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *WorkflowHandler) ExecuteWorkflow(c *gin.Context) {
	id := c.Param("id")
	mode := c.Query("mode")
	if mode == "" {
		mode = "sync"
	}

	var req vo.ExecuteWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	accountID, err := util.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}

	if mode == "async" {
		runID, err := h.svc.ExecuteWorkflowAsync(id, req.TenantID, accountID, req.InputText, string(req.InputJSON))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "执行失败", err.Error())
			return
		}
		response.Success(c, gin.H{"run_id": runID})
		return
	}

	result, err := h.svc.ExecuteWorkflowSync(id, req.TenantID, accountID, req.InputText, string(req.InputJSON))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "执行失败", err.Error())
		return
	}

	response.Success(c, result)
}

func (h *WorkflowHandler) ValidateWorkflow(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "tenant_id 不能为空", "")
		return
	}

	errs, err := h.svc.ValidateWorkflow(id, tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "校验失败", err.Error())
		return
	}

	response.Success(c, gin.H{"valid": len(errs) == 0, "errors": errs})
}

func (h *WorkflowHandler) CloneWorkflow(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "tenant_id 不能为空", "")
		return
	}

	accountID, err := util.GetAccountID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "未认证", err.Error())
		return
	}
	dto, err := h.svc.CloneWorkflow(id, tenantID, accountID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "克隆失败", err.Error())
		return
	}

	response.Success(c, h.toResponse(dto))
}

func (h *WorkflowHandler) ListTemplates(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "tenant_id 不能为空", "")
		return
	}

	list, err := h.svc.ListTemplates(tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败", err.Error())
		return
	}

	response.Success(c, gin.H{"data": h.toResponseList(list)})
}

func (h *WorkflowHandler) ListRuns(c *gin.Context) {
	workflowID := c.Param("workflowId")
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "tenant_id 不能为空", "")
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	list, total, err := h.svc.ListRuns(workflowID, tenantID, (page-1)*size, size)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败", err.Error())
		return
	}

	response.Success(c, gin.H{"data": list, "total": total, "page": page, "size": size})
}

func (h *WorkflowHandler) GetRun(c *gin.Context) {
	runID := c.Param("runId")
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "tenant_id 不能为空", "")
		return
	}
	dto, err := h.svc.GetRun(runID, tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败", err.Error())
		return
	}
	response.Success(c, dto)
}

func (h *WorkflowHandler) GetRunNodes(c *gin.Context) {
	runID := c.Param("runId")
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "tenant_id 不能为空", "")
		return
	}
	nodes, err := h.svc.GetRunNodes(runID, tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败", err.Error())
		return
	}
	response.Success(c, gin.H{"data": nodes})
}

func (h *WorkflowHandler) CancelRun(c *gin.Context) {
	runID := c.Param("runId")
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		response.Error(c, http.StatusBadRequest, "tenant_id 不能为空", "")
		return
	}
	if err := h.svc.CancelRun(runID, tenantID); err != nil {
		response.Error(c, http.StatusInternalServerError, "取消失败", err.Error())
		return
	}
	response.Success(c, gin.H{"message": "已取消"})
}

func (h *WorkflowHandler) toResponse(d *dto.WorkflowDTO) *vo.WorkflowResponse {
	return &vo.WorkflowResponse{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		FlowJSON:    d.FlowJSON,
		TemplateID:  d.TemplateID,
		Status:      d.Status,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func (h *WorkflowHandler) toResponseList(list []*dto.WorkflowDTO) []*vo.WorkflowResponse {
	result := make([]*vo.WorkflowResponse, len(list))
	for i, d := range list {
		result[i] = h.toResponse(d)
	}
	return result
}
