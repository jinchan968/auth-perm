package vo

import "encoding/json"

type CreateWorkflowRequest struct {
	TenantID    string          `json:"tenant_id" binding:"required"`
	Name        string          `json:"name" binding:"required,max=128"`
	Description string          `json:"description"`
	FlowJSON    json.RawMessage `json:"flow_json" binding:"required"`
	TemplateID  *string         `json:"template_id"`
}

type UpdateWorkflowRequest struct {
	TenantID    string           `json:"tenant_id" binding:"required"`
	Name        *string          `json:"name" binding:"omitempty,max=128"`
	Description *string          `json:"description"`
	FlowJSON    *json.RawMessage `json:"flow_json"`
	Status      *string          `json:"status"`
}

type ExecuteWorkflowRequest struct {
	TenantID  string          `json:"tenant_id" binding:"required"`
	InputText string          `json:"input_text"`
	InputJSON json.RawMessage `json:"input_json"`
}

type ListWorkflowsRequest struct {
	TenantID string `form:"tenant_id" binding:"required"`
	Page     int    `form:"page,default=1"`
	Size     int    `form:"size,default=10"`
	Type     string `form:"type"`
}

type WorkflowResponse struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	FlowJSON    json.RawMessage `json:"flow_json"`
	TemplateID  *string         `json:"template_id,omitempty"`
	Status      string          `json:"status"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type WorkflowListResponse struct {
	Data  []*WorkflowResponse `json:"data"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Size  int                 `json:"size"`
}
