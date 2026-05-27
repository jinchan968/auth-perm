package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"auth-perm/internal/domain/workflow/constant"
	"auth-perm/internal/domain/workflow/dm"
	"auth-perm/internal/domain/workflow/dto"
	"auth-perm/internal/domain/workflow/repo"
	"auth-perm/internal/domain/workflow/vo"
	"auth-perm/internal/infra/opencode"

	"github.com/google/uuid"
)

type WorkflowService struct {
	workflowRepo *repo.WorkflowRepo
	runRepo      *repo.WorkflowRunRepo
	runNodeRepo  *repo.WorkflowRunNodeRepo
	openCode     *opencode.Client
	engine       *Engine
	wsHub        *WSHub
	cancels      sync.Map
}

func NewWorkflowService(
	wr *repo.WorkflowRepo,
	rr *repo.WorkflowRunRepo,
	nr *repo.WorkflowRunNodeRepo,
	oc *opencode.Client,
) *WorkflowService {
	hub := NewWSHub()
	go hub.Run()
	engine := NewEngine(oc, hub, rr, nr)
	return &WorkflowService{
		workflowRepo: wr,
		runRepo:      rr,
		runNodeRepo:  nr,
		openCode:     oc,
		engine:       engine,
		wsHub:        hub,
	}
}

func (s *WorkflowService) GetWSHub() *WSHub {
	return s.wsHub
}

func (s *WorkflowService) CreateWorkflow(tenantID, accountID, name, description, flowJSON string, templateID *string) (*dto.WorkflowDTO, error) {
	do := &dm.WorkflowDO{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		AccountID:   accountID,
		Name:        name,
		Description: description,
		FlowJSON:    flowJSON,
		TemplateID:  templateID,
		Status:      constant.StatusDraft,
	}
	if err := s.workflowRepo.Create(do); err != nil {
		return nil, err
	}
	return dto.FromWorkflowDO(do), nil
}

func (s *WorkflowService) ListWorkflows(tenantID, accountID string, offset, limit int) ([]*dto.WorkflowDTO, int64, error) {
	list, total, err := s.workflowRepo.List(tenantID, accountID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	result := make([]*dto.WorkflowDTO, len(list))
	for i, do := range list {
		result[i] = dto.FromWorkflowDO(do)
	}
	return result, total, nil
}

func (s *WorkflowService) GetWorkflow(id, tenantID string) (*dto.WorkflowDTO, error) {
	do, err := s.workflowRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if do.TenantID != tenantID {
		return nil, fmt.Errorf("workflow not found")
	}
	return dto.FromWorkflowDO(do), nil
}

func (s *WorkflowService) UpdateWorkflow(id, tenantID string, name, description, flowJSON, status *string) (*dto.WorkflowDTO, error) {
	do, err := s.workflowRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if do.TenantID != tenantID {
		return nil, fmt.Errorf("workflow not found")
	}
	if name != nil {
		do.Name = *name
	}
	if description != nil {
		do.Description = *description
	}
	if flowJSON != nil {
		do.FlowJSON = *flowJSON
	}
	if status != nil {
		if *status != constant.StatusDraft && *status != constant.StatusPublished {
			return nil, fmt.Errorf("invalid status: %s", *status)
		}
		if *status == constant.StatusPublished && do.FlowJSON == "" {
			return nil, fmt.Errorf("无法发布：工作流缺少流程定义")
		}
		do.Status = *status
	}
	do.UpdatedAt = time.Now()
	if err := s.workflowRepo.Update(do); err != nil {
		return nil, err
	}
	return dto.FromWorkflowDO(do), nil
}

func (s *WorkflowService) DeleteWorkflow(id, tenantID string) error {
	do, err := s.workflowRepo.GetByID(id)
	if err != nil {
		return err
	}
	if do.TenantID != tenantID {
		return fmt.Errorf("workflow not found")
	}
	return s.workflowRepo.Delete(id)
}

func (s *WorkflowService) ValidateWorkflow(id, tenantID string) ([]ValidationError, error) {
	do, err := s.workflowRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if do.TenantID != tenantID {
		return nil, fmt.Errorf("workflow not found")
	}

	var graph vo.FlowGraph
	if err := json.Unmarshal([]byte(do.FlowJSON), &graph); err != nil {
		return nil, fmt.Errorf("parse flow_json: %w", err)
	}

	return ValidateFlowGraph(&graph), nil
}

func (s *WorkflowService) ExecuteWorkflowSync(workflowID, tenantID, accountID, inputText, inputJSON string) (*dto.WorkflowRunDTO, error) {
	do, err := s.workflowRepo.GetByID(workflowID)
	if err != nil {
		return nil, err
	}
	if do.TenantID != tenantID {
		return nil, fmt.Errorf("workflow not found")
	}

	runID := uuid.New().String()
	start := time.Now()

	runDO := &dm.WorkflowRunDO{
		ID:            runID,
		WorkflowID:    workflowID,
		TenantID:      tenantID,
		AccountID:     accountID,
		ExecutionMode: "sync",
		InputText:     inputText,
		InputJSON:     inputJSON,
		Status:        constant.StatusRunning,
		StartedAt:     &start,
	}
	if err := s.runRepo.Create(runDO); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	result, err := s.engine.Execute(ctx, runID, do.FlowJSON, inputText)
	finish := time.Now()
	duration := int(finish.Sub(start).Milliseconds())

	runDO.FinishedAt = &finish
	runDO.DurationMs = duration

	if err != nil {
		runDO.Status = constant.StatusFailed
		runDO.Error = err.Error()
	} else {
		runDO.Status = constant.StatusSuccess
		resultJSON, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			runDO.Status = constant.StatusFailed
			runDO.Error = marshalErr.Error()
		} else {
			runDO.ResultJSON = string(resultJSON)
		}
	}

	s.runRepo.Update(runDO)
	s.wsHub.Broadcast(runID, map[string]interface{}{
		"type":   "run_end",
		"status": runDO.Status,
	})

	return dto.FromWorkflowRunDO(runDO), nil
}

func (s *WorkflowService) ExecuteWorkflowAsync(workflowID, tenantID, accountID, inputText, inputJSON string) (string, error) {
	do, err := s.workflowRepo.GetByID(workflowID)
	if err != nil {
		return "", err
	}
	if do.TenantID != tenantID {
		return "", fmt.Errorf("workflow not found")
	}

	runID := uuid.New().String()
	start := time.Now()

	runDO := &dm.WorkflowRunDO{
		ID:            runID,
		WorkflowID:    workflowID,
		TenantID:      tenantID,
		AccountID:     accountID,
		ExecutionMode: "async",
		InputText:     inputText,
		InputJSON:     inputJSON,
		Status:        constant.StatusRunning,
		StartedAt:     &start,
	}
	if err := s.runRepo.Create(runDO); err != nil {
		return "", err
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancels.Store(runID, cancel)

	go func() {
		defer s.cancels.Delete(runID)
		result, err := s.engine.Execute(ctx, runID, do.FlowJSON, inputText)
		finish := time.Now()
		duration := int(finish.Sub(start).Milliseconds())

		runDO.FinishedAt = &finish
		runDO.DurationMs = duration

		if ctx.Err() != nil {
			runDO.Status = constant.StatusCancelled
		} else if err != nil {
			runDO.Status = constant.StatusFailed
			runDO.Error = err.Error()
		} else {
			runDO.Status = constant.StatusSuccess
			resultJSON, marshalErr := json.Marshal(result)
			if marshalErr != nil {
				runDO.Status = constant.StatusFailed
				runDO.Error = marshalErr.Error()
			} else {
				runDO.ResultJSON = string(resultJSON)
			}
		}

		s.runRepo.Update(runDO)
		s.wsHub.Broadcast(runID, map[string]interface{}{
			"type":   "run_end",
			"status": runDO.Status,
		})
	}()

	return runID, nil
}

func (s *WorkflowService) CloneWorkflow(id, tenantID, accountID string) (*dto.WorkflowDTO, error) {
	do, err := s.workflowRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if do.TenantID != tenantID {
		return nil, fmt.Errorf("workflow not found")
	}

	newDO := &dm.WorkflowDO{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		AccountID:   accountID,
		Name:        do.Name + " (副本)",
		Description: do.Description,
		FlowJSON:    do.FlowJSON,
		TemplateID:  &do.ID,
		Status:      constant.StatusDraft,
	}
	if err := s.workflowRepo.Create(newDO); err != nil {
		return nil, err
	}
	return dto.FromWorkflowDO(newDO), nil
}

func (s *WorkflowService) ListTemplates(tenantID string) ([]*dto.WorkflowDTO, error) {
	list, err := s.workflowRepo.ListTemplates(tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.WorkflowDTO, len(list))
	for i, do := range list {
		result[i] = dto.FromWorkflowDO(do)
	}
	return result, nil
}

func (s *WorkflowService) ListRuns(workflowID, tenantID string, offset, limit int) ([]*dto.WorkflowRunDTO, int64, error) {
	wd, err := s.workflowRepo.GetByID(workflowID)
	if err != nil {
		return nil, 0, err
	}
	if wd.TenantID != tenantID {
		return nil, 0, fmt.Errorf("workflow not found")
	}
	list, total, err := s.runRepo.ListByWorkflow(workflowID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	result := make([]*dto.WorkflowRunDTO, len(list))
	for i, do := range list {
		result[i] = dto.FromWorkflowRunDO(do)
	}
	return result, total, nil
}

func (s *WorkflowService) GetRun(runID, tenantID string) (*dto.WorkflowRunDTO, error) {
	do, err := s.runRepo.GetByID(runID)
	if err != nil {
		return nil, err
	}
	if do.TenantID != tenantID {
		return nil, fmt.Errorf("run not found")
	}
	return dto.FromWorkflowRunDO(do), nil
}

func (s *WorkflowService) GetRunNodes(runID, tenantID string) ([]*dm.WorkflowRunNodeDO, error) {
	run, err := s.runRepo.GetByID(runID)
	if err != nil {
		return nil, err
	}
	if run.TenantID != tenantID {
		return nil, fmt.Errorf("run not found")
	}
	return s.runNodeRepo.ListByRun(runID)
}

func (s *WorkflowService) CancelRun(runID, tenantID string) error {
	do, err := s.runRepo.GetByID(runID)
	if err != nil {
		return err
	}
	if do.TenantID != tenantID {
		return fmt.Errorf("run not found")
	}
	if do.Status == constant.StatusRunning {
		if cancel, ok := s.cancels.LoadAndDelete(runID); ok {
			cancel.(context.CancelFunc)()
		}
		cancelled, err := s.runRepo.CancelIfRunning(runID)
		if err != nil {
			return err
		}
		if !cancelled {
			return fmt.Errorf("run already finished")
		}
		return nil
	}
	return nil
}
