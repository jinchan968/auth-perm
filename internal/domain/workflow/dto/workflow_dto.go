package dto

import (
	"auth-perm/internal/domain/workflow/dm"
	"encoding/json"
)

type WorkflowDTO struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	AccountID   string          `json:"account_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	FlowJSON    json.RawMessage `json:"flow_json"`
	TemplateID  *string         `json:"template_id,omitempty"`
	Status      string          `json:"status"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

func FromWorkflowDO(do *dm.WorkflowDO) *WorkflowDTO {
	return &WorkflowDTO{
		ID:          do.ID,
		TenantID:    do.TenantID,
		AccountID:   do.AccountID,
		Name:        do.Name,
		Description: do.Description,
		FlowJSON:    json.RawMessage(do.FlowJSON),
		TemplateID:  do.TemplateID,
		Status:      do.Status,
		CreatedAt:   do.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   do.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type WorkflowRunDTO struct {
	ID            string          `json:"id"`
	WorkflowID    string          `json:"workflow_id"`
	TenantID      string          `json:"tenant_id"`
	AccountID     string          `json:"account_id"`
	ExecutionMode string          `json:"execution_mode"`
	InputText     string          `json:"input_text,omitempty"`
	InputJSON     json.RawMessage `json:"input_json,omitempty"`
	ResultJSON    json.RawMessage `json:"result_json,omitempty"`
	Status        string          `json:"status"`
	StartedAt     *string         `json:"started_at,omitempty"`
	FinishedAt    *string         `json:"finished_at,omitempty"`
	DurationMs    int             `json:"duration_ms"`
	Error         string          `json:"error,omitempty"`
}

func FromWorkflowRunDO(do *dm.WorkflowRunDO) *WorkflowRunDTO {
	dto := &WorkflowRunDTO{
		ID:            do.ID,
		WorkflowID:    do.WorkflowID,
		TenantID:      do.TenantID,
		AccountID:     do.AccountID,
		ExecutionMode: do.ExecutionMode,
		InputText:     do.InputText,
		InputJSON:     json.RawMessage(do.InputJSON),
		ResultJSON:    json.RawMessage(do.ResultJSON),
		Status:        do.Status,
		DurationMs:    do.DurationMs,
		Error:         do.Error,
	}
	if do.StartedAt != nil {
		t := do.StartedAt.Format("2006-01-02T15:04:05Z07:00")
		dto.StartedAt = &t
	}
	if do.FinishedAt != nil {
		t := do.FinishedAt.Format("2006-01-02T15:04:05Z07:00")
		dto.FinishedAt = &t
	}
	return dto
}
