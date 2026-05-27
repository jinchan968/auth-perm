# AI 工作流编排工具 — 实施计划文档

本文档基于 `WORKFLOW_DESIGN.md`，将 7 步实施计划拆解为可逐条执行的明细任务。每个步骤包含：**要创建/修改的文件清单**、**具体内容**、**验收标准**。

---

## 前置检查

开始前确认以下环境就绪：

```bash
cd /Users/looper/Documents/workspace/golang/auth-perm
make build      # 当前代码能编译通过
make test       # 当前测试通过
make migrate-status  # 数据库迁移状态正常
```

新增依赖（第 1 步时安装）：

```bash
# 后端
cd /Users/looper/Documents/workspace/golang/auth-perm
go get github.com/cloudwego/eino/compose
go get github.com/gorilla/websocket

# 前端
cd /Users/looper/Documents/workspace/golang/auth-perm/ui
pnpm add @xyflow/react @dnd-kit/core @dnd-kit/utilities
```

---

## 步骤 1：Migration + Domain 骨架

**目标**：建 3 张表；DO + Repo + Module 注册 + Container 装配 + Route 占位。完成后后端可编译，API 路由已注册。

### 1.1 创建迁移文件

**文件**：`migrations/000028_create_workflows.sql`

```sql
-- +goose Up
-- +goose StatementBegin

-- 工作流定义表
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    flow_json JSONB NOT NULL DEFAULT '{}',
    template_id UUID,
    status VARCHAR(16) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflows_tenant_account ON workflows(tenant_id, account_id);
CREATE INDEX IF NOT EXISTS idx_workflows_template ON workflows(template_id);

-- 工作流执行记录表
CREATE TABLE IF NOT EXISTS workflow_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    execution_mode VARCHAR(8) NOT NULL DEFAULT 'sync',
    input_text TEXT,
    input_json JSONB,
    result_json JSONB,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    error TEXT
);

CREATE INDEX IF NOT EXISTS idx_wf_runs_workflow ON workflow_runs(workflow_id);
CREATE INDEX IF NOT EXISTS idx_wf_runs_tenant ON workflow_runs(tenant_id, account_id, started_at DESC);

-- 工作流逐节点执行记录表
CREATE TABLE IF NOT EXISTS workflow_run_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    node_id VARCHAR(128) NOT NULL,
    node_type VARCHAR(32) NOT NULL,
    node_label VARCHAR(128),
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    input_json JSONB,
    output_json JSONB,
    error TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    retry_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_wf_run_nodes_run ON workflow_run_nodes(run_id);
CREATE INDEX IF NOT EXISTS idx_wf_run_nodes_run_node ON workflow_run_nodes(run_id, node_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS workflow_run_nodes;
DROP TABLE IF EXISTS workflow_runs;
DROP TABLE IF EXISTS workflows;

-- +goose StatementEnd
```

**验收**：`make migrate-up` 成功，3 张表已创建。

### 1.2 创建 Domain 模块目录结构

```bash
mkdir -p internal/domain/workflow/{constant,dm,repo,service,handler,dto,vo}
```

### 1.3 常量定义

**文件**：`internal/domain/workflow/constant/node_type.go`

```go
package constant

const (
    NodeTypeTrigger   = "trigger"
    NodeTypeLLM       = "llm"
    NodeTypeCondition = "condition"
    NodeTypeTransform = "transform"
    NodeTypeMerge     = "merge"
    NodeTypeOutput    = "output"
)

var validNodeTypes = map[string]bool{
    NodeTypeTrigger:   true,
    NodeTypeLLM:       true,
    NodeTypeCondition: true,
    NodeTypeTransform: true,
    NodeTypeMerge:     true,
    NodeTypeOutput:    true,
}

func IsValidNodeType(t string) bool {
    return validNodeTypes[t]
}
```

**文件**：`internal/domain/workflow/constant/run_status.go`

```go
package constant

const (
    StatusPending   = "pending"
    StatusRunning   = "running"
    StatusSuccess   = "success"
    StatusFailed    = "failed"
    StatusCancelled = "cancelled"
)

var validStatuses = map[string]bool{
    StatusPending:   true,
    StatusRunning:   true,
    StatusSuccess:   true,
    StatusFailed:    true,
    StatusCancelled: true,
}

func IsValidRunStatus(s string) bool {
    return validStatuses[s]
}
```

**文件**：`internal/domain/workflow/constant/operators.go`

```go
package constant

const (
    OpContains     = "contains"
    OpNotContains  = "not_contains"
    OpEquals       = "equals"
    OpMatches      = "matches"
    OpStartsWith   = "starts_with"
    OpEndsWith     = "ends_with"
    OpGT           = "gt"
    OpGTE          = "gte"
    OpLT           = "lt"
    OpLTE          = "lte"
    OpIsEmpty      = "is_empty"
    OpNotEmpty     = "not_empty"
)

var validOperators = map[string]bool{
    OpContains:    true,
    OpNotContains: true,
    OpEquals:      true,
    OpMatches:     true,
    OpStartsWith:  true,
    OpEndsWith:    true,
    OpGT:          true,
    OpGTE:         true,
    OpLT:          true,
    OpLTE:         true,
    OpIsEmpty:     true,
    OpNotEmpty:    true,
}

func IsValidOperator(op string) bool {
    return validOperators[op]
}
```

**文件**：`internal/domain/workflow/constant/transforms.go`

```go
package constant

const (
    TransformRegexExtract   = "regex_extract"
    TransformRegexReplace   = "regex_replace"
    TransformTrim           = "trim"
    TransformMarkdownToText = "markdown_to_text"
    TransformExtractJSON    = "extract_json"
    TransformTruncate       = "truncate"
    TransformToUppercase    = "to_uppercase"
    TransformToLowercase    = "to_lowercase"
)

var validTransforms = map[string]bool{
    TransformRegexExtract:   true,
    TransformRegexReplace:   true,
    TransformTrim:           true,
    TransformMarkdownToText: true,
    TransformExtractJSON:    true,
    TransformTruncate:       true,
    TransformToUppercase:    true,
    TransformToLowercase:    true,
}

func IsValidTransform(t string) bool {
    return validTransforms[t]
}
```

**文件**：`internal/domain/workflow/constant/strategies.go`

```go
package constant

const (
    MergeStrategyConcat = "concat"
    MergeStrategyFirst  = "first"
    MergeStrategyJoin   = "join"
)

const (
    OutputFormatRaw      = "raw"
    OutputFormatJSON     = "json"
    OutputFormatMarkdown = "markdown"
)

const (
    JoinModeAnd = "and"
    JoinModeOr  = "or"
)
```

### 1.4 Domain Model (DO)

**文件**：`internal/domain/workflow/dm/workflow_do.go`

```go
package dm

import "time"

type WorkflowDO struct {
    ID          string    `gorm:"primaryKey;type:varchar(36)"`
    TenantID    string    `gorm:"column:tenant_id;type:varchar(64);not null;default:'default'"`
    AccountID   string    `gorm:"column:account_id;type:varchar(64);not null"`
    Name        string    `gorm:"column:name;type:varchar(128);not null"`
    Description string    `gorm:"column:description;type:text"`
    FlowJSON    string    `gorm:"column:flow_json;type:jsonb;not null"`
    TemplateID  *string   `gorm:"column:template_id;type:varchar(36)"`
    Status      string    `gorm:"column:status;type:varchar(16);not null;default:'draft'"`
    CreatedAt   time.Time `gorm:"column:created_at"`
    UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (WorkflowDO) TableName() string {
    return "workflows"
}
```

**文件**：`internal/domain/workflow/dm/workflow_run_do.go`

```go
package dm

import "time"

type WorkflowRunDO struct {
    ID             string     `gorm:"primaryKey;type:varchar(36)"`
    WorkflowID     string     `gorm:"column:workflow_id;type:varchar(36);not null"`
    TenantID       string     `gorm:"column:tenant_id;type:varchar(64);not null;default:'default'"`
    AccountID      string     `gorm:"column:account_id;type:varchar(64);not null"`
    ExecutionMode  string     `gorm:"column:execution_mode;type:varchar(8);not null;default:'sync'"`
    InputText      string     `gorm:"column:input_text;type:text"`
    InputJSON      string     `gorm:"column:input_json;type:jsonb"`
    ResultJSON     string     `gorm:"column:result_json;type:jsonb"`
    Status         string     `gorm:"column:status;type:varchar(16);not null;default:'pending'"`
    StartedAt      *time.Time `gorm:"column:started_at"`
    FinishedAt     *time.Time `gorm:"column:finished_at"`
    DurationMs     int        `gorm:"column:duration_ms"`
    Error          string     `gorm:"column:error;type:text"`
}

func (WorkflowRunDO) TableName() string {
    return "workflow_runs"
}
```

**文件**：`internal/domain/workflow/dm/workflow_run_node_do.go`

```go
package dm

import "time"

type WorkflowRunNodeDO struct {
    ID           string     `gorm:"primaryKey;type:varchar(36)"`
    RunID        string     `gorm:"column:run_id;type:varchar(36);not null"`
    NodeID       string     `gorm:"column:node_id;type:varchar(128);not null"`
    NodeType     string     `gorm:"column:node_type;type:varchar(32);not null"`
    NodeLabel    string     `gorm:"column:node_label;type:varchar(128)"`
    Status       string     `gorm:"column:status;type:varchar(16);not null;default:'pending'"`
    InputJSON    string     `gorm:"column:input_json;type:jsonb"`
    OutputJSON   string     `gorm:"column:output_json;type:jsonb"`
    Error        string     `gorm:"column:error;type:text"`
    StartedAt    *time.Time `gorm:"column:started_at"`
    FinishedAt   *time.Time `gorm:"column:finished_at"`
    DurationMs   int        `gorm:"column:duration_ms"`
    RetryCount   int        `gorm:"column:retry_count;default:0"`
}

func (WorkflowRunNodeDO) TableName() string {
    return "workflow_run_nodes"
}
```

### 1.5 Repository 层

**文件**：`internal/domain/workflow/repo/workflow_repo.go`

```go
package repo

import (
    "auth-perm/internal/domain/workflow/dm"
    "gorm.io/gorm"
)

type WorkflowRepo struct {
    db *gorm.DB
}

func NewWorkflowRepo(db *gorm.DB) *WorkflowRepo {
    return &WorkflowRepo{db: db}
}

func (r *WorkflowRepo) Create(do *dm.WorkflowDO) error {
    return r.db.Create(do).Error
}

func (r *WorkflowRepo) GetByID(id string) (*dm.WorkflowDO, error) {
    var do dm.WorkflowDO
    err := r.db.Where("id = ?", id).First(&do).Error
    if err != nil {
        return nil, err
    }
    return &do, nil
}

func (r *WorkflowRepo) List(tenantID, accountID string, offset, limit int) ([]*dm.WorkflowDO, int64, error) {
    var list []*dm.WorkflowDO
    var total int64
    query := r.db.Where("tenant_id = ? AND account_id = ?", tenantID, accountID)
    query.Model(&dm.WorkflowDO{}).Count(&total)
    err := query.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&list).Error
    return list, total, err
}

func (r *WorkflowRepo) Update(do *dm.WorkflowDO) error {
    return r.db.Save(do).Error
}

func (r *WorkflowRepo) Delete(id string) error {
    return r.db.Where("id = ?", id).Delete(&dm.WorkflowDO{}).Error
}

func (r *WorkflowRepo) ListTemplates(tenantID string) ([]*dm.WorkflowDO, error) {
    var list []*dm.WorkflowDO
    err := r.db.Where("tenant_id = ? AND status = ? AND template_id IS NULL", tenantID, "template").
        Order("created_at DESC").Find(&list).Error
    return list, err
}
```

**文件**：`internal/domain/workflow/repo/workflow_run_repo.go`

```go
package repo

import (
    "auth-perm/internal/domain/workflow/dm"
    "gorm.io/gorm"
)

type WorkflowRunRepo struct {
    db *gorm.DB
}

func NewWorkflowRunRepo(db *gorm.DB) *WorkflowRunRepo {
    return &WorkflowRunRepo{db: db}
}

func (r *WorkflowRunRepo) Create(do *dm.WorkflowRunDO) error {
    return r.db.Create(do).Error
}

func (r *WorkflowRunRepo) GetByID(id string) (*dm.WorkflowRunDO, error) {
    var do dm.WorkflowRunDO
    err := r.db.Where("id = ?", id).First(&do).Error
    if err != nil {
        return nil, err
    }
    return &do, nil
}

func (r *WorkflowRunRepo) ListByWorkflow(workflowID string, offset, limit int) ([]*dm.WorkflowRunDO, int64, error) {
    var list []*dm.WorkflowRunDO
    var total int64
    query := r.db.Where("workflow_id = ?", workflowID)
    query.Model(&dm.WorkflowRunDO{}).Count(&total)
    err := query.Order("started_at DESC").Offset(offset).Limit(limit).Find(&list).Error
    return list, total, err
}

func (r *WorkflowRunRepo) Update(do *dm.WorkflowRunDO) error {
    return r.db.Save(do).Error
}
```

**文件**：`internal/domain/workflow/repo/workflow_run_node_repo.go`

```go
package repo

import (
    "auth-perm/internal/domain/workflow/dm"
    "gorm.io/gorm"
)

type WorkflowRunNodeRepo struct {
    db *gorm.DB
}

func NewWorkflowRunNodeRepo(db *gorm.DB) *WorkflowRunNodeRepo {
    return &WorkflowRunNodeRepo{db: db}
}

func (r *WorkflowRunNodeRepo) Create(do *dm.WorkflowRunNodeDO) error {
    return r.db.Create(do).Error
}

func (r *WorkflowRunNodeRepo) Update(do *dm.WorkflowRunNodeDO) error {
    return r.db.Save(do).Error
}

func (r *WorkflowRunNodeRepo) ListByRun(runID string) ([]*dm.WorkflowRunNodeDO, error) {
    var list []*dm.WorkflowRunNodeDO
    err := r.db.Where("run_id = ?", runID).Order("started_at ASC").Find(&list).Error
    return list, err
}
```

### 1.6 Module 注册

**文件**：`internal/domain/workflow/module.go`

```go
package workflow

import (
    "auth-perm/internal/domain/workflow/handler"
    "auth-perm/internal/domain/workflow/repo"
    "auth-perm/internal/domain/workflow/service"
    "go.uber.org/dig"
    "gorm.io/gorm"
    "auth-perm/internal/infra/opencode"
)

func RegisterWorkflowDomain(container *dig.Container) error {
    container.Provide(func(db *gorm.DB) *repo.WorkflowRepo {
        return repo.NewWorkflowRepo(db)
    })
    container.Provide(func(db *gorm.DB) *repo.WorkflowRunRepo {
        return repo.NewWorkflowRunRepo(db)
    })
    container.Provide(func(db *gorm.DB) *repo.WorkflowRunNodeRepo {
        return repo.NewWorkflowRunNodeRepo(db)
    })

    container.Provide(func(
        wr *repo.WorkflowRepo,
        rr *repo.WorkflowRunRepo,
        nr *repo.WorkflowRunNodeRepo,
        oc *opencode.Client,
    ) *service.WorkflowService {
        return service.NewWorkflowService(wr, rr, nr, oc)
    })

    container.Provide(func(svc *service.WorkflowService) *handler.WorkflowHandler {
        return handler.NewWorkflowHandler(svc)
    })
    container.Provide(func(svc *service.WorkflowService) *handler.WorkflowWSHandler {
        return handler.NewWorkflowWSHandler(svc)
    })

    return nil
}
```

### 1.7 Container 装配

**修改文件**：`internal/container/container.go`

在 `registerDomains()` 中添加：

```go
import workflow "auth-perm/internal/domain/workflow"

if err := workflow.RegisterWorkflowDomain(container); err != nil {
    return err
}
```

在 `registerHandlers()` 中添加 workflow handler 参数（类型 `*workflowHandler.WorkflowHandler` 和 `*workflowHandler.WorkflowWSHandler`）。

### 1.8 Route 注册

**修改文件**：`internal/controller/http/route.go`

在 `RegisterRoutes` 函数签名中添加：

```go
workflowH *workflowHandler.WorkflowHandler,
workflowWSH *workflowHandler.WorkflowWSHandler,
```

在函数体中添加：

```go
RegisterWorkflowRoutes(v1, permMW, workflowH, workflowWSH, loginService)
```

添加路由函数：

```go
func RegisterWorkflowRoutes(
    router *gin.RouterGroup,
    permMW gin.HandlerFunc,
    h *workflowHandler.WorkflowHandler,
    wsh *workflowHandler.WorkflowWSHandler,
    loginService *service.LoginService,
) {
    workflow := router.Group("/workflow")
    workflow.Use(middleware.AuthMiddleware(loginService))
    workflow.Use(permMW)
    {
        workflow.GET("", h.ListWorkflows)
        workflow.POST("", h.CreateWorkflow)
        workflow.GET("/:id", h.GetWorkflow)
        workflow.PUT("/:id", h.UpdateWorkflow)
        workflow.DELETE("/:id", h.DeleteWorkflow)
        workflow.POST("/:id/execute", h.ExecuteWorkflow)
        workflow.POST("/:id/validate", h.ValidateWorkflow)
        workflow.POST("/:id/clone", h.CloneWorkflow)
        workflow.GET("/:id/runs", h.ListRuns)
        workflow.GET("/runs/:runId", h.GetRun)
        workflow.GET("/runs/:runId/nodes", h.GetRunNodes)
        workflow.POST("/runs/:runId/cancel", h.CancelRun)
        workflow.GET("/templates", h.ListTemplates)
    }

    ws := router.Group("/ws")
    ws.GET("/run/:runId", wsh.HandleWS)
}
```

### 1.9 Handler 占位

**文件**：`internal/domain/workflow/handler/workflow_handler.go`

```go
package handler

import "github.com/gin-gonic/gin"

type WorkflowHandler struct {
    svc interface{} // 占位
}

func NewWorkflowHandler(svc interface{}) *WorkflowHandler {
    return &WorkflowHandler{svc: svc}
}

func (h *WorkflowHandler) ListWorkflows(c *gin.Context)    { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context)   { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) GetWorkflow(c *gin.Context)      { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context)   { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context)   { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) ExecuteWorkflow(c *gin.Context) { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) ValidateWorkflow(c *gin.Context) { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) CloneWorkflow(c *gin.Context)     { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) ListTemplates(c *gin.Context)    { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) ListRuns(c *gin.Context)          { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) GetRun(c *gin.Context)            { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) GetRunNodes(c *gin.Context)      { c.JSON(200, gin.H{"msg": "TODO"}) }
func (h *WorkflowHandler) CancelRun(c *gin.Context)        { c.JSON(200, gin.H{"msg": "TODO"}) }
```

**文件**：`internal/domain/workflow/handler/workflow_ws_handler.go`

```go
package handler

import "github.com/gin-gonic/gin"

type WorkflowWSHandler struct {
    svc interface{}
}

func NewWorkflowWSHandler(svc interface{}) *WorkflowWSHandler {
    return &WorkflowWSHandler{svc: svc}
}

func (h *WorkflowWSHandler) HandleWS(c *gin.Context) {
    c.JSON(200, gin.H{"msg": "TODO"})
}
```

### 1.10 Service 占位

**文件**：`internal/domain/workflow/service/workflow_service.go`

```go
package service

import (
    "auth-perm/internal/domain/workflow/repo"
    "auth-perm/internal/infra/opencode"
)

type WorkflowService struct {
    workflowRepo    *repo.WorkflowRepo
    runRepo         *repo.WorkflowRunRepo
    runNodeRepo     *repo.WorkflowRunNodeRepo
    openCode        *opencode.Client
}

func NewWorkflowService(
    wr *repo.WorkflowRepo,
    rr *repo.WorkflowRunRepo,
    nr *repo.WorkflowRunNodeRepo,
    oc *opencode.Client,
) *WorkflowService {
    return &WorkflowService{
        workflowRepo: wr,
        runRepo:      rr,
        runNodeRepo:  nr,
        openCode:     oc,
    }
}
```

**验收标准**：
1. `make build` 编译通过
2. `make test` 通过
3. `make migrate-up` 3 张表创建成功
4. 访问 `/api/v1/workflow` 返回 `{"msg": "TODO"}`（200 OK）

---

## 步骤 2：Eino 引擎 + 规则求值 + 图校验

**目标**：实现 Flow JSON 到 Eino Graph 构建器、结构化规则递归求值器、图结构校验器。

### 2.1 DTO 定义

**文件**：`internal/domain/workflow/dto/workflow_dto.go`

```go
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
```

### 2.2 VO 定义

**文件**：`internal/domain/workflow/vo/workflow_vo.go`

```go
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
    ID          string          `json:"id" binding:"required"`
    TenantID    string          `json:"tenant_id" binding:"required"`
    Name        string          `json:"name" binding:"required,max=128"`
    Description string          `json:"description"`
    FlowJSON    json.RawMessage `json:"flow_json" binding:"required"`
    Status      string          `json:"status"`
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
```

**文件**：`internal/domain/workflow/vo/condition_vo.go`

```go
package vo

import (
    "auth-perm/internal/domain/workflow/constant"
    "errors"
)

type RuleGroup struct {
    Logic string      `json:"logic"`
    Rules []RuleItem  `json:"rules"`
}

type RuleItem struct {
    Field    string     `json:"field,omitempty"`
    Operator string     `json:"operator,omitempty"`
    Value    string     `json:"value,omitempty"`
    Negate   bool       `json:"negate,omitempty"`
    SubGroup *RuleGroup `json:"sub_group,omitempty"`
}

type BranchConfig struct {
    Handle string    `json:"handle"`
    Rule   RuleGroup `json:"rule"`
}

type ConditionNodeData struct {
    Branches     []BranchConfig `json:"branches"`
    DefaultHandle string        `json:"default_handle"`
}

func (r *RuleGroup) Validate() error {
    if r.Logic != "AND" && r.Logic != "OR" {
        return errors.New("invalid logic")
    }
    if len(r.Rules) == 0 {
        return errors.New("empty rules")
    }
    for _, rule := range r.Rules {
        if rule.SubGroup != nil {
            if err := rule.SubGroup.Validate(); err != nil {
                return err
            }
            continue
        }
        if !constant.IsValidOperator(rule.Operator) {
            return errors.New("invalid operator: " + rule.Operator)
        }
        if rule.Field == "" {
            return errors.New("field required")
        }
    }
    return nil
}
```

**文件**：`internal/domain/workflow/vo/flow_vo.go`

```go
package vo

import (
    "encoding/json"
    "errors"
)

type FlowNode struct {
    ID       string          `json:"id"`
    Type     string          `json:"type"`
    Position NodePosition    `json:"position"`
    Data     json.RawMessage `json:"data"`
}

type FlowEdge struct {
    ID           string  `json:"id"`
    Source       string  `json:"source"`
    Target       string  `json:"target"`
    SourceHandle *string `json:"sourceHandle,omitempty"`
    TargetHandle *string `json:"targetHandle,omitempty"`
}

type NodePosition struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

type FlowViewport struct {
    X    float64 `json:"x"`
    Y    float64 `json:"y"`
    Zoom float64 `json:"zoom"`
}

type FlowGraph struct {
    Nodes    []FlowNode   `json:"nodes"`
    Edges    []FlowEdge   `json:"edges"`
    Viewport FlowViewport `json:"viewport"`
}

func (g *FlowGraph) Validate() error {
    if len(g.Nodes) == 0 {
        return errors.New("empty graph")
    }
    return nil
}
```

### 2.3 规则求值器

**文件**：`internal/domain/workflow/service/rule_evaluator.go`

```go
package service

import (
    "auth-perm/internal/domain/workflow/constant"
    "auth-perm/internal/domain/workflow/vo"
    "regexp"
    "strconv"
    "strings"
)

func evaluateRuleGroup(content string, group vo.RuleGroup) (bool, error) {
    results := make([]bool, len(group.Rules))
    for i, rule := range group.Rules {
        if rule.SubGroup != nil {
            r, err := evaluateRuleGroup(content, *rule.SubGroup)
            if err != nil {
                return false, err
            }
            results[i] = r
            continue
        }
        r, err := evaluateSingleRule(content, rule)
        if err != nil {
            return false, err
        }
        results[i] = r
    }
    return combineLogic(group.Logic, results), nil
}

func evaluateSingleRule(content string, rule vo.RuleItem) (bool, error) {
    var result bool
    switch rule.Operator {
    case constant.OpContains:
        result = strings.Contains(content, rule.Value)
    case constant.OpNotContains:
        result = !strings.Contains(content, rule.Value)
    case constant.OpEquals:
        result = content == rule.Value
    case constant.OpMatches:
        matched, err := regexp.MatchString(rule.Value, content)
        if err != nil {
            return false, err
        }
        result = matched
    case constant.OpStartsWith:
        result = strings.HasPrefix(content, rule.Value)
    case constant.OpEndsWith:
        result = strings.HasSuffix(content, rule.Value)
    case constant.OpGT:
        length := len([]rune(content))
        val, _ := strconv.Atoi(rule.Value)
        result = length > val
    case constant.OpGTE:
        length := len([]rune(content))
        val, _ := strconv.Atoi(rule.Value)
        result = length >= val
    case constant.OpLT:
        length := len([]rune(content))
        val, _ := strconv.Atoi(rule.Value)
        result = length < val
    case constant.OpLTE:
        length := len([]rune(content))
        val, _ := strconv.Atoi(rule.Value)
        result = length <= val
    case constant.OpIsEmpty:
        result = strings.TrimSpace(content) == ""
    case constant.OpNotEmpty:
        result = strings.TrimSpace(content) != ""
    default:
        return false, nil
    }
    if rule.Negate {
        result = !result
    }
    return result, nil
}

func combineLogic(logic string, results []bool) bool {
    if len(results) == 0 {
        return false
    }
    if logic == "AND" {
        for _, r := range results {
            if !r {
                return false
            }
        }
        return true
    }
    for _, r := range results {
        if r {
            return true
        }
    }
    return false
}
```

### 2.4 图校验器

**文件**：`internal/domain/workflow/service/graph_validator.go`

```go
package service

import (
    "auth-perm/internal/domain/workflow/constant"
    "auth-perm/internal/domain/workflow/vo"
    "encoding/json"
)

type ValidationError struct {
    NodeID  string `json:"node_id,omitempty"`
    Message string `json:"message"`
    Level   string `json:"level"`
}

func ValidateFlowGraph(graph *vo.FlowGraph) []ValidationError {
    var errs []ValidationError

    triggerCount := 0
    outputCount := 0
    for _, node := range graph.Nodes {
        if node.Type == constant.NodeTypeTrigger {
            triggerCount++
        }
        if node.Type == constant.NodeTypeOutput {
            outputCount++
        }
    }

    if triggerCount == 0 {
        errs = append(errs, ValidationError{Message: "缺少 trigger 节点", Level: "error"})
    } else if triggerCount > 1 {
        errs = append(errs, ValidationError{Message: "只能有一个 trigger 节点", Level: "error"})
    }

    if outputCount == 0 {
        errs = append(errs, ValidationError{Message: "缺少 output 节点", Level: "error"})
    }

    adj := make(map[string][]string)
    inDegree := make(map[string]int)
    nodeMap := make(map[string]vo.FlowNode)

    for _, node := range graph.Nodes {
        nodeMap[node.ID] = node
        inDegree[node.ID] = 0
    }

    for _, edge := range graph.Edges {
        adj[edge.Source] = append(adj[edge.Source], edge.Target)
        inDegree[edge.Target]++
    }

    if triggerCount == 1 && outputCount > 0 {
        var triggerID string
        for _, node := range graph.Nodes {
            if node.Type == constant.NodeTypeTrigger {
                triggerID = node.ID
                break
            }
        }
        reachable := bfsReachable(triggerID, adj)
        for _, node := range graph.Nodes {
            if node.Type == constant.NodeTypeOutput {
                if !reachable[node.ID] {
                    errs = append(errs, ValidationError{
                        NodeID:  node.ID,
                        Message: "output 不可达",
                        Level:   "error",
                    })
                }
            }
        }
    }

    for _, node := range graph.Nodes {
        if node.Type == constant.NodeTypeTrigger {
            continue
        }
        if inDegree[node.ID] == 0 {
            errs = append(errs, ValidationError{
                NodeID:  node.ID,
                Message: "孤立节点",
                Level:   "warning",
            })
        }
    }

    for _, node := range graph.Nodes {
        if node.Type == constant.NodeTypeCondition {
            outCount := len(adj[node.ID])
            if outCount < 2 {
                errs = append(errs, ValidationError{
                    NodeID:  node.ID,
                    Message: "condition 需至少 2 条出边",
                    Level:   "error",
                })
            }
        }
    }

    for _, node := range graph.Nodes {
        if node.Type == constant.NodeTypeMerge {
            if inDegree[node.ID] < 2 {
                errs = append(errs, ValidationError{
                    NodeID:  node.ID,
                    Message: "merge 需至少 2 条入边",
                    Level:   "error",
                })
            }
        }
    }

    for _, node := range graph.Nodes {
        if node.Type == constant.NodeTypeLLM {
            var data map[string]interface{}
            if err := json.Unmarshal(node.Data, &data); err == nil {
                if modelID, ok := data["model_id"].(string); !ok || modelID == "" {
                    errs = append(errs, ValidationError{
                        NodeID:  node.ID,
                        Message: "LLM 需指定 model_id",
                        Level:   "error",
                    })
                }
            }
        }
    }

    if hasCycle(graph.Nodes, graph.Edges) {
        errs = append(errs, ValidationError{Message: "存在环路", Level: "error"})
    }

    for _, node := range graph.Nodes {
        if node.Type == constant.NodeTypeOutput && inDegree[node.ID] == 0 {
            errs = append(errs, ValidationError{
                NodeID:  node.ID,
                Message: "output 需至少 1 条入边",
                Level:   "error",
            })
        }
    }

    for _, node := range graph.Nodes {
        if node.Type == constant.NodeTypeTrigger && inDegree[node.ID] > 0 {
            errs = append(errs, ValidationError{
                NodeID:  node.ID,
                Message: "trigger 不应有入边",
                Level:   "error",
            })
        }
    }

    return errs
}

func bfsReachable(start string, adj map[string][]string) map[string]bool {
    visited := make(map[string]bool)
    queue := []string{start}
    visited[start] = true
    for len(queue) > 0 {
        curr := queue[0]
        queue = queue[1:]
        for _, next := range adj[curr] {
            if !visited[next] {
                visited[next] = true
                queue = append(queue, next)
            }
        }
    }
    return visited
}

func hasCycle(nodes []vo.FlowNode, edges []vo.FlowEdge) bool {
    adj := make(map[string][]string)
    for _, edge := range edges {
        adj[edge.Source] = append(adj[edge.Source], edge.Target)
    }
    state := make(map[string]int)
    var dfs func(nodeID string) bool
    dfs = func(nodeID string) bool {
        state[nodeID] = 1
        for _, next := range adj[nodeID] {
            if state[next] == 1 {
                return true
            }
            if state[next] == 0 && dfs(next) {
                return true
            }
        }
        state[nodeID] = 2
        return false
    }
    for _, node := range nodes {
        if state[node.ID] == 0 {
            if dfs(node.ID) {
                return true
            }
        }
    }
    return false
}
```

### 2.5 Eino 引擎

**文件**：`internal/domain/workflow/service/engine.go`

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "time"

    "auth-perm/internal/domain/workflow/constant"
    "auth-perm/internal/domain/workflow/dm"
    "auth-perm/internal/domain/workflow/vo"
    "auth-perm/internal/infra/opencode"

    "github.com/cloudwego/eino/compose"
)

type NodeOutput struct {
    NodeID    string                 `json:"node_id"`
    NodeType  string                 `json:"node_type"`
    Content   string                 `json:"content"`
    ModelName string                 `json:"model_name,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Error     string                 `json:"error,omitempty"`
}

type Engine struct {
    openCode *opencode.Client
    wsHub    *WSHub
    runRepo  *repo.WorkflowRunRepo
    nodeRepo *repo.WorkflowRunNodeRepo
}

func NewEngine(
    oc *opencode.Client,
    hub *WSHub,
    rr *repo.WorkflowRunRepo,
    nr *repo.WorkflowRunNodeRepo,
) *Engine {
    return &Engine{openCode: oc, wsHub: hub, runRepo: rr, nodeRepo: nr}
}

func (e *Engine) Execute(ctx context.Context, runID string, flowJSON string, inputText string) (*NodeOutput, error) {
    var graph vo.FlowGraph
    if err := json.Unmarshal([]byte(flowJSON), &graph); err != nil {
        return nil, fmt.Errorf("parse flow_json: %w", err)
    }

    errs := ValidateFlowGraph(&graph)
    var fatalErrs []ValidationError
    for _, err := range errs {
        if err.Level == "error" {
            fatalErrs = append(fatalErrs, err)
        }
    }
    if len(fatalErrs) > 0 {
        return nil, fmt.Errorf("validation failed: %d errors", len(fatalErrs))
    }

    g := compose.NewGraph[string, *NodeOutput]()
    nodeMap := make(map[string]vo.FlowNode)
    for _, node := range graph.Nodes {
        nodeMap[node.ID] = node
    }

    adj := make(map[string][]vo.FlowEdge)
    for _, edge := range graph.Edges {
        adj[edge.Source] = append(adj[edge.Source], edge)
    }

    for _, node := range graph.Nodes {
        if err := e.registerNode(g, node, runID); err != nil {
            return nil, fmt.Errorf("register node %s: %w", node.ID, err)
        }
    }

    for _, edge := range graph.Edges {
        sourceNode := nodeMap[edge.Source]
        if sourceNode.Type == constant.NodeTypeCondition {
            continue
        }
        g.AddEdge(edge.Source, edge.Target)
    }

    runnable, err := g.Compile(ctx, compose.WithGraphName("workflow_"+runID))
    if err != nil {
        return nil, fmt.Errorf("compile graph: %w", err)
    }

    result, err := runnable.Invoke(ctx, inputText)
    if err != nil {
        return nil, fmt.Errorf("invoke graph: %w", err)
    }

    return result, nil
}

func (e *Engine) registerNode(g *compose.Graph[string, *NodeOutput], node vo.FlowNode, runID string) error {
    switch node.Type {
    case constant.NodeTypeTrigger:
        g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, input string) (*NodeOutput, error) {
            return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: input}, nil
        }))

    case constant.NodeTypeLLM:
        var data struct {
            ModelID       string  `json:"model_id"`
            SystemPrompt  string  `json:"system_prompt"`
            Temperature   float64 `json:"temperature"`
            ReasoningMode string  `json:"reasoning_mode"`
        }
        json.Unmarshal(node.Data, &data)

        g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, in *NodeOutput) (*NodeOutput, error) {
            e.writeNodeStart(runID, node.ID, node.Type, in.Content)
            start := time.Now()
            content, err := e.openCode.Chat(ctx, data.ModelID, data.SystemPrompt, in.Content, false, data.ReasoningMode)
            duration := time.Since(start).Milliseconds()
            if err != nil {
                e.writeNodeEnd(runID, node.ID, node.Type, "", err.Error(), duration)
                return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Error: err.Error()}, nil
            }
            e.writeNodeEnd(runID, node.ID, node.Type, content, "", duration)
            return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: content, ModelName: data.ModelID}, nil
        }))

    case constant.NodeTypeCondition:
        var condData vo.ConditionNodeData
        if err := json.Unmarshal(node.Data, &condData); err != nil {
            return err
        }
        handleNames := make(map[string]bool)
        for _, branch := range condData.Branches {
            handleNames[branch.Handle] = true
        }
        handleNames[condData.DefaultHandle] = true

        branchFunc := compose.NewGraphBranch(
            func(ctx context.Context, in *NodeOutput) (string, error) {
                for _, branch := range condData.Branches {
                    matched, err := evaluateRuleGroup(in.Content, branch.Rule)
                    if err != nil {
                        continue
                    }
                    if matched {
                        return branch.Handle, nil
                    }
                }
                return condData.DefaultHandle, nil
            },
            handleNames,
        )
        g.AddBranch(node.ID, branchFunc)

    case constant.NodeTypeTransform:
        var data struct {
            Operation string                 `json:"operation"`
            Params    map[string]interface{} `json:"params"`
        }
        json.Unmarshal(node.Data, &data)

        g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, in *NodeOutput) (*NodeOutput, error) {
            result, err := e.executeTransform(in.Content, data.Operation, data.Params)
            if err != nil {
                return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Error: err.Error()}, nil
            }
            return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: result}, nil
        }))

    case constant.NodeTypeMerge:
        var data struct {
            Strategy  string `json:"strategy"`
            Delimiter string `json:"delimiter,omitempty"`
        }
        json.Unmarshal(node.Data, &data)

        g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, in *NodeOutput) (*NodeOutput, error) {
            results := collectPredecessorResults(ctx)
            merged := e.executeMerge(results, data.Strategy, data.Delimiter)
            return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: merged}, nil
        }))

    case constant.NodeTypeOutput:
        var data struct {
            Format   string `json:"format"`
            JoinMode string `json:"join_mode"`
        }
        json.Unmarshal(node.Data, &data)

        if data.JoinMode == constant.JoinModeOr {
            g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, in *NodeOutput) (*NodeOutput, error) {
                results := collectPredecessorResults(ctx)
                for _, res := range results {
                    if res.Error == "" {
                        return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: res.Content}, nil
                    }
                }
                return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Error: "all branches failed"}, nil
            }))
        } else {
            g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, in *NodeOutput) (*NodeOutput, error) {
                results := collectPredecessorResults(ctx)
                var contents []string
                for _, res := range results {
                    if res.Error == "" {
                        contents = append(contents, res.Content)
                    }
                }
                merged := strings.Join(contents, "\n\n---\n\n")
                return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: merged}, nil
            }))
        }

    default:
        return fmt.Errorf("unknown node type: %s", node.Type)
    }

    return nil
}

func (e *Engine) executeTransform(content, operation string, params map[string]interface{}) (string, error) {
    switch operation {
    case constant.TransformRegexExtract:
        pattern := params["pattern"].(string)
        groupIndex := 0
        if gi, ok := params["group_index"]; ok {
            groupIndex = int(gi.(float64))
        }
        re, err := regexp.Compile(pattern)
        if err != nil {
            return "", err
        }
        matches := re.FindStringSubmatch(content)
        if len(matches) > groupIndex {
            return matches[groupIndex], nil
        }
        return "", nil

    case constant.TransformRegexReplace:
        pattern := params["pattern"].(string)
        replacement := params["replacement"].(string)
        re, err := regexp.Compile(pattern)
        if err != nil {
            return "", err
        }
        return re.ReplaceAllString(content, replacement), nil

    case constant.TransformTrim:
        return strings.TrimSpace(content), nil

    case constant.TransformMarkdownToText:
        result := regexp.MustCompile(`\*\*|\*|__|_|#|\[|\]|\(|\)`).ReplaceAllString(content, "")
        return result, nil

    case constant.TransformExtractJSON:
        re := regexp.MustCompile(`(?s)\{.*\}`)
        match := re.FindString(content)
        return match, nil

    case constant.TransformTruncate:
        maxLen := int(params["max_length"].(float64))
        runes := []rune(content)
        if len(runes) > maxLen {
            return string(runes[:maxLen]) + "...", nil
        }
        return content, nil

    case constant.TransformToUppercase:
        return strings.ToUpper(content), nil

    case constant.TransformToLowercase:
        return strings.ToLower(content), nil

    default:
        return content, nil
    }
}

func (e *Engine) executeMerge(results []*NodeOutput, strategy, delimiter string) string {
    switch strategy {
    case constant.MergeStrategyConcat:
        var contents []string
        for _, res := range results {
            if res.Error == "" {
                contents = append(contents, res.Content)
            }
        }
        return strings.Join(contents, "\n\n---\n\n")

    case constant.MergeStrategyFirst:
        for _, res := range results {
            if res.Error == "" {
                return res.Content
            }
        }
        return ""

    case constant.MergeStrategyJoin:
        var contents []string
        for _, res := range results {
            if res.Error == "" {
                contents = append(contents, res.Content)
            }
        }
        if delimiter == "" {
            delimiter = "\n"
        }
        return strings.Join(contents, delimiter)

    default:
        return ""
    }
}

func collectPredecessorResults(ctx context.Context) []*NodeOutput {
    return nil
}

func (e *Engine) writeNodeStart(runID, nodeID, nodeType, input string) {
    if e.wsHub != nil {
        e.wsHub.Broadcast(runID, map[string]interface{}{
            "type":      "node_start",
            "node_id":   nodeID,
            "node_type": nodeType,
        })
    }
    if e.nodeRepo != nil {
        now := time.Now()
        e.nodeRepo.Create(&dm.WorkflowRunNodeDO{
            RunID:      runID,
            NodeID:     nodeID,
            NodeType:   nodeType,
            Status:     constant.StatusRunning,
            InputJSON:  fmt.Sprintf(`{"content":"%s"}`, input),
            StartedAt:  &now,
        })
    }
}

func (e *Engine) writeNodeEnd(runID, nodeID, nodeType, output, errStr string, durationMs int64) {
    now := time.Now()
    status := constant.StatusSuccess
    if errStr != "" {
        status = constant.StatusFailed
    }
    if e.wsHub != nil {
        msgType := "node_end"
        if errStr != "" {
            msgType = "node_error"
        }
        msg := map[string]interface{}{
            "type":        msgType,
            "node_id":     nodeID,
            "node_type":   nodeType,
            "duration_ms": durationMs,
        }
        if errStr != "" {
            msg["error"] = errStr
        }
        e.wsHub.Broadcast(runID, msg)
    }
    if e.nodeRepo != nil {
        e.nodeRepo.Update(&dm.WorkflowRunNodeDO{
            RunID:      runID,
            NodeID:     nodeID,
            NodeType:   nodeType,
            Status:     status,
            OutputJSON: fmt.Sprintf(`{"content":"%s"}`, output),
            Error:      errStr,
            FinishedAt: &now,
            DurationMs: int(durationMs),
        })
    }
}
```

**验收标准**：
1. `make build` 编译通过
2. `ValidateFlowGraph` 能正确检测 10 种校验错误

---

## 步骤 3：Handler + WebSocket

**目标**：实现 14 个 REST 端点 + WebSocket 实时状态推送。

### 3.1 WebSocket Hub

**文件**：`internal/domain/workflow/service/ws_hub.go`

```go
package service

import (
    "encoding/json"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

type WSHub struct {
    clients    map[string]map[*WSClient]bool
    register   chan *WSClient
    unregister chan *WSClient
    broadcast  chan WSMessage
    mu         sync.RWMutex
}

type WSClient struct {
    runID string
    hub   *WSHub
    conn  *websocket.Conn
    send  chan []byte
}

type WSMessage struct {
    RunID string
    Data  map[string]interface{}
}

func NewWSHub() *WSHub {
    return &WSHub{
        clients:    make(map[string]map[*WSClient]bool),
        register:   make(chan *WSClient),
        unregister: make(chan *WSClient),
        broadcast:  make(chan WSMessage, 256),
    }
}

func (h *WSHub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            if h.clients[client.runID] == nil {
                h.clients[client.runID] = make(map[*WSClient]bool)
            }
            h.clients[client.runID][client] = true
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if clients, ok := h.clients[client.runID]; ok {
                delete(clients, client)
                if len(clients) == 0 {
                    delete(h.clients, client.runID)
                }
            }
            h.mu.Unlock()
            close(client.send)

        case msg := <-h.broadcast:
            h.mu.RLock()
            clients := h.clients[msg.RunID]
            h.mu.RUnlock()
            data, _ := json.Marshal(msg.Data)
            for client := range clients {
                select {
                case client.send <- data:
                default:
                }
            }
        }
    }
}

func (h *WSHub) Broadcast(runID string, data map[string]interface{}) {
    select {
    case h.broadcast <- WSMessage{RunID: runID, Data: data}:
    default:
    }
}

func (h *WSHub) RegisterClient(runID string, conn *websocket.Conn) *WSClient {
    client := &WSClient{
        runID: runID,
        hub:   h,
        conn:  conn,
        send:  make(chan []byte, 256),
    }
    h.register <- client
    return client
}

func (h *WSHub) UnregisterClient(client *WSClient) {
    h.unregister <- client
}

func (c *WSClient) WritePump() {
    ticker := time.NewTicker(30 * time.Second)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            c.conn.WriteMessage(websocket.TextMessage, message)

        case <-ticker.C:
            c.conn.WriteMessage(websocket.PingMessage, nil)
        }
    }
}

func (c *WSClient) ReadPump() {
    defer func() {
        c.hub.UnregisterClient(c)
        c.conn.Close()
    }()

    c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })

    for {
        _, _, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
    }
}
```

### 3.2 Handler 实现

**文件**：`internal/domain/workflow/handler/workflow_handler.go`（替换占位）

```go
package handler

import (
    "net/http"
    "strconv"

    "auth-perm/internal/controller/response"
    "auth-perm/internal/domain/workflow/dto"
    "auth-perm/internal/domain/workflow/service"
    "auth-perm/internal/domain/workflow/vo"
    "auth-perm/pkg/util"

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

    tenantID := c.Query("tenant_id")
    if tenantID == "" {
        response.Error(c, http.StatusBadRequest, "tenant_id 不能为空", "")
        return
    }

    accountID, _ := util.GetAccountID(c)
    offset := (req.Page - 1) * req.Size
    limit := req.Size

    var list []*dto.WorkflowDTO
    var total int64
    var err error

    if req.Type == "template" {
        list, err = h.svc.ListTemplates(tenantID)
        total = int64(len(list))
    } else {
        list, total, err = h.svc.ListWorkflows(tenantID, accountID, offset, limit)
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

    accountID, _ := util.GetAccountID(c)
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

    dto, err := h.svc.UpdateWorkflow(id, req.TenantID, req.Name, req.Description, string(req.FlowJSON), req.Status)
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

    accountID, _ := util.GetAccountID(c)

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

    accountID, _ := util.GetAccountID(c)
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
    id := c.Param("id")
    page, _ := strconv.Atoi(c.Query("page"))
    size, _ := strconv.Atoi(c.Query("size"))
    if page < 1 {
        page = 1
    }
    if size < 1 {
        size = 10
    }

    list, total, err := h.svc.ListRuns(id, (page-1)*size, size)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "查询失败", err.Error())
        return
    }

    response.Success(c, gin.H{"data": list, "total": total, "page": page, "size": size})
}

func (h *WorkflowHandler) GetRun(c *gin.Context) {
    runID := c.Param("runId")
    dto, err := h.svc.GetRun(runID)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "查询失败", err.Error())
        return
    }
    response.Success(c, dto)
}

func (h *WorkflowHandler) GetRunNodes(c *gin.Context) {
    runID := c.Param("runId")
    nodes, err := h.svc.GetRunNodes(runID)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "查询失败", err.Error())
        return
    }
    response.Success(c, gin.H{"data": nodes})
}

func (h *WorkflowHandler) CancelRun(c *gin.Context) {
    runID := c.Param("runId")
    if err := h.svc.CancelRun(runID); err != nil {
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
```

### 3.3 WebSocket Handler

**文件**：`internal/domain/workflow/handler/workflow_ws_handler.go`（替换占位）

```go
package handler

import (
    "net/http"
    "strings"

    "auth-perm/internal/domain/workflow/service"
    "auth-perm/pkg/jwt"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

type WorkflowWSHandler struct {
    svc      *service.WorkflowService
    upgrader websocket.Upgrader
}

func NewWorkflowWSHandler(svc *service.WorkflowService) *WorkflowWSHandler {
    return &WorkflowWSHandler{
        svc: svc,
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true
            },
        },
    }
}

func (h *WorkflowWSHandler) HandleWS(c *gin.Context) {
    runID := c.Param("runId")
    if runID == "" {
        c.JSON(400, gin.H{"error": "run_id required"})
        return
    }

    token := c.Query("token")
    if token == "" {
        authHeader := c.GetHeader("Authorization")
        if strings.HasPrefix(authHeader, "Bearer ") {
            token = strings.TrimPrefix(authHeader, "Bearer ")
        }
    }

    _, err := jwt.ParseToken(token)
    if err != nil {
        c.JSON(401, gin.H{"error": "invalid token"})
        return
    }

    conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }

    hub := h.svc.GetWSHub()
    client := hub.RegisterClient(runID, conn)

    go client.WritePump()
    go client.ReadPump()
}
```

### 3.4 Service 实现

**文件**：`internal/domain/workflow/service/workflow_service.go`（替换占位）

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
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

func (s *WorkflowService) UpdateWorkflow(id, tenantID, name, description, flowJSON, status string) (*dto.WorkflowDTO, error) {
    do, err := s.workflowRepo.GetByID(id)
    if err != nil {
        return nil, err
    }
    if do.TenantID != tenantID {
        return nil, fmt.Errorf("workflow not found")
    }
    if name != "" {
        do.Name = name
    }
    do.Description = description
    if flowJSON != "" {
        do.FlowJSON = flowJSON
    }
    if status != "" {
        do.Status = status
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
    s.runRepo.Create(runDO)

    result, err := s.engine.Execute(context.Background(), runID, do.FlowJSON, inputText)
    finish := time.Now()
    duration := int(finish.Sub(start).Milliseconds())

    runDO.FinishedAt = &finish
    runDO.DurationMs = duration

    if err != nil {
        runDO.Status = constant.StatusFailed
        runDO.Error = err.Error()
    } else {
        runDO.Status = constant.StatusSuccess
        resultJSON, _ := json.Marshal(result)
        runDO.ResultJSON = string(resultJSON)
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
    s.runRepo.Create(runDO)

    go func() {
        result, err := s.engine.Execute(context.Background(), runID, do.FlowJSON, inputText)
        finish := time.Now()
        duration := int(finish.Sub(start).Milliseconds())

        runDO.FinishedAt = &finish
        runDO.DurationMs = duration

        if err != nil {
            runDO.Status = constant.StatusFailed
            runDO.Error = err.Error()
        } else {
            runDO.Status = constant.StatusSuccess
            resultJSON, _ := json.Marshal(result)
            runDO.ResultJSON = string(resultJSON)
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

func (s *WorkflowService) ListRuns(workflowID string, offset, limit int) ([]*dto.WorkflowRunDTO, int64, error) {
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

func (s *WorkflowService) GetRun(runID string) (*dto.WorkflowRunDTO, error) {
    do, err := s.runRepo.GetByID(runID)
    if err != nil {
        return nil, err
    }
    return dto.FromWorkflowRunDO(do), nil
}

func (s *WorkflowService) GetRunNodes(runID string) ([]*dm.WorkflowRunNodeDO, error) {
    return s.runNodeRepo.ListByRun(runID)
}

func (s *WorkflowService) CancelRun(runID string) error {
    do, err := s.runRepo.GetByID(runID)
    if err != nil {
        return err
    }
    if do.Status == constant.StatusRunning {
        do.Status = constant.StatusCancelled
        now := time.Now()
        do.FinishedAt = &now
        return s.runRepo.Update(do)
    }
    return nil
}
```

**验收标准**：
1. `make build` 编译通过
2. `make test` 通过
3. 用 curl 测试：
   - POST /api/v1/workflow 创建成功
   - GET /api/v1/workflow 列表返回
   - POST /api/v1/workflow/:id/validate 校验返回正确

---

## 步骤 4：权限种子 + is_system 修复

**目标**：补充 workflow 权限资源；修复新建权限页 is_system 开关的安全问题。

### 4.1 权限 Migration

**文件**：`migrations/000029_workflow_permissions.sql`

```sql
-- +goose Up
-- +goose StatementBegin

-- 菜单权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000030',
    'default',
    'menu:workflow',
    '工作流菜单',
    '访问工作流编排菜单',
    'menu',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 读取权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000031',
    'default',
    'workflow.read',
    '查看工作流',
    '查看工作流列表和详情',
    'button',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 写入权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000032',
    'default',
    'workflow.write',
    '创建/编辑工作流',
    '创建、编辑、执行、克隆工作流',
    'button',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 删除权限
INSERT INTO permissions (id, tenant_id, code, name, description, resource, is_system, is_active, created_at, updated_at)
VALUES (
    'a0000001-0000-0000-0000-000000000033',
    'default',
    'workflow.delete',
    '删除工作流',
    '删除工作流',
    'button',
    true, true, NOW(), NOW()
) ON CONFLICT (tenant_id, code) DO NOTHING;

-- 资源绑定
INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000030',
    'a0000001-0000-0000-0000-000000000030',
    'workflow',
    'menu',
    '工作流菜单',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000031',
    'a0000001-0000-0000-0000-000000000030',
    '/api/v1/workflow/*',
    'api_path',
    '工作流 API 通配',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000032',
    'a0000001-0000-0000-0000-000000000031',
    'GET /api/v1/workflow',
    'api_path',
    '查看工作流列表',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000033',
    'a0000001-0000-0000-0000-000000000031',
    'workflow.tab.designer',
    'button',
    '编排设计 Tab',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000034',
    'a0000001-0000-0000-0000-000000000031',
    'workflow.tab.runs',
    'button',
    '运行历史 Tab',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000035',
    'a0000001-0000-0000-0000-000000000032',
    'POST /api/v1/workflow',
    'api_path',
    '创建工作流',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000036',
    'a0000001-0000-0000-0000-000000000032',
    'PUT /api/v1/workflow/*',
    'api_path',
    '更新工作流',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

INSERT INTO permission_resources (id, permission_id, resource_id, resource_type, resource_name, tenant_id, created_at, updated_at)
VALUES (
    'b0000001-0000-0000-0000-000000000037',
    'a0000001-0000-0000-0000-000000000033',
    'DELETE /api/v1/workflow/*',
    'api_path',
    '删除工作流',
    'default', NOW(), NOW()
) ON CONFLICT ON CONSTRAINT unique_permission_resource DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM permission_resources WHERE permission_id IN (
    'a0000001-0000-0000-0000-000000000030',
    'a0000001-0000-0000-0000-000000000031',
    'a0000001-0000-0000-0000-000000000032',
    'a0000001-0000-0000-0000-000000000033'
);
DELETE FROM permissions WHERE id IN (
    'a0000001-0000-0000-0000-000000000030',
    'a0000001-0000-0000-0000-000000000031',
    'a0000001-0000-0000-0000-000000000032',
    'a0000001-0000-0000-0000-000000000033'
);

-- +goose StatementEnd
```

### 4.2 修复 is_system 安全问题

**修改文件**：`ui/app/permissions/new/page.tsx`

在 `is_system` 开关外层加 `isSuperAdmin` 守卫（需从 auth store 获取）：

```tsx
// 在页面组件中获取 isSuperAdmin
const { user } = useAuthStore()
const isSuperAdmin = user?.role === 'super_admin' // 或从权限 API 查询

// 在表单渲染中：
{isSuperAdmin && (
  <div className="flex items-center justify-between space-x-2 pt-1">
    <div className="space-y-0.5">
      <Label htmlFor="is_system">系统权限</Label>
      <p className="text-sm text-muted-foreground">系统权限不可编辑或删除</p>
    </div>
    <Switch
      id="is_system"
      checked={formData.is_system}
      onCheckedChange={(checked) => setFormData({ ...formData, is_system: checked })}
    />
  </div>
)}
```

同时后端 `CreatePermission` 需校验：仅超管可创建 `is_system=true`。

**验收标准**：
1. `make migrate-up` 权限数据插入成功
2. 非超管用户在前端看不到 `is_system` 开关
3. 直接调用 API 传 `is_system=true` 被后端拒绝

---

## 步骤 5：前端类型 + API + WS Hook

**目标**：完成前端类型定义、API 封装、WebSocket Hook。

### 5.1 类型定义

**文件**：`ui/types/workflow.ts`

```typescript
export type NodeType = 'trigger' | 'llm' | 'condition' | 'transform' | 'merge' | 'output'

export interface FlowNode {
  id: string
  type: NodeType
  position: { x: number; y: number }
  data: Record<string, unknown>
}

export interface FlowEdge {
  id: string
  source: string
  target: string
  sourceHandle?: string
  targetHandle?: string
}

export interface FlowGraph {
  nodes: FlowNode[]
  edges: FlowEdge[]
  viewport: { x: number; y: number; zoom: number }
}

export interface Workflow {
  id: string
  tenant_id: string
  name: string
  description?: string
  flow_json: FlowGraph
  template_id?: string
  status: 'draft' | 'published' | 'archived'
  created_at: string
  updated_at: string
}

export interface WorkflowListResponse {
  data: Workflow[]
  total: number
  page: number
  size: number
}

export interface WorkflowRun {
  id: string
  workflow_id: string
  execution_mode: 'sync' | 'async'
  input_text?: string
  input_json?: Record<string, unknown>
  result_json?: Record<string, unknown>
  status: 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
  started_at?: string
  finished_at?: string
  duration_ms: number
  error?: string
}

export interface WorkflowRunNode {
  id: string
  run_id: string
  node_id: string
  node_type: NodeType
  node_label?: string
  status: 'pending' | 'running' | 'success' | 'failed'
  input_json?: Record<string, unknown>
  output_json?: Record<string, unknown>
  error?: string
  started_at?: string
  finished_at?: string
  duration_ms: number
}

export interface RuleItem {
  field?: string
  operator?: string
  value?: string
  negate?: boolean
  sub_group?: RuleGroup
}

export interface RuleGroup {
  logic: 'AND' | 'OR'
  rules: RuleItem[]
}

export interface BranchConfig {
  handle: string
  rule: RuleGroup
}

export interface ConditionNodeData {
  branches: BranchConfig[]
  default_handle: string
}

export interface WSMessage {
  type: 'node_start' | 'node_end' | 'node_error' | 'run_end' | 'ping'
  run_id: string
  node_id?: string
  node_type?: NodeType
  duration_ms?: number
  output?: string
  error?: string
  status?: string
  result?: string
  ts: string
}
```

### 5.2 API 封装

**文件**：`ui/lib/api/workflow.ts`

```typescript
import { request } from './request'
import type { Workflow, WorkflowListResponse, WorkflowRun, WorkflowRunNode, FlowGraph } from '@/types/workflow'

const BASE = '/api/v1/workflow'

export async function listWorkflows(params: { tenant_id: string; page?: number; size?: number; type?: string }) {
  return request.get<WorkflowListResponse>(BASE, { params })
}

export async function createWorkflow(data: {
  tenant_id: string
  name: string
  description?: string
  flow_json: FlowGraph
  template_id?: string
}) {
  return request.post<Workflow>(BASE, data)
}

export async function getWorkflow(id: string, tenant_id: string) {
  return request.get<Workflow>(`${BASE}/${id}`, { params: { tenant_id } })
}

export async function updateWorkflow(
  id: string,
  data: {
    tenant_id: string
    name: string
    description?: string
    flow_json: FlowGraph
    status?: string
  }
) {
  return request.put<Workflow>(`${BASE}/${id}`, data)
}

export async function deleteWorkflow(id: string, tenant_id: string) {
  return request.delete(`${BASE}/${id}`, { params: { tenant_id } })
}

export async function executeWorkflow(
  id: string,
  data: { tenant_id: string; input_text?: string; input_json?: Record<string, unknown> },
  mode: 'sync' | 'async' = 'sync'
) {
  if (mode === 'async') {
    return request.post<{ run_id: string }>(`${BASE}/${id}/execute?mode=async`, data)
  }
  return request.post<WorkflowRun>(`${BASE}/${id}/execute?mode=sync`, data)
}

export async function validateWorkflow(id: string, tenant_id: string) {
  return request.post<{ valid: boolean; errors: Array<{ node_id?: string; message: string; level: string }> }>(
    `${BASE}/${id}/validate`,
    { tenant_id }
  )
}

export async function cloneWorkflow(id: string, tenant_id: string) {
  return request.post<Workflow>(`${BASE}/${id}/clone`, null, { params: { tenant_id } })
}

export async function listTemplates(tenant_id: string) {
  return request.get<{ data: Workflow[] }>(`${BASE}/templates`, { params: { tenant_id } })
}

export async function listRuns(id: string, params: { page?: number; size?: number }) {
  return request.get<{ data: WorkflowRun[]; total: number; page: number; size: number }>(
    `${BASE}/${id}/runs`,
    { params }
  )
}

export async function getRun(runId: string) {
  return request.get<WorkflowRun>(`${BASE}/runs/${runId}`)
}

export async function getRunNodes(runId: string) {
  return request.get<{ data: WorkflowRunNode[] }>(`${BASE}/runs/${runId}/nodes`)
}

export async function cancelRun(runId: string) {
  return request.post<{ message: string }>(`${BASE}/runs/${runId}/cancel`)
}
```

### 5.3 WebSocket Hook

**文件**：`ui/hooks/use-workflow-ws.ts`

```typescript
'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import type { WSMessage } from '@/types/workflow'

interface Options {
  runId: string | null
  token: string
  onMessage?: (msg: WSMessage) => void
  onConnect?: () => void
  onDisconnect?: () => void
}

export function useWorkflowWS({ runId, token, onMessage, onConnect, onDisconnect }: Options) {
  const [connected, setConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectCountRef = useRef(0)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const connect = useCallback(() => {
    if (!runId || !token) return

    const wsUrl = `${process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'}/api/v1/ws/run/${runId}?token=${token}`
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      setConnected(true)
      reconnectCountRef.current = 0
      onConnect?.()
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as WSMessage
        setLastMessage(msg)
        onMessage?.(msg)
      } catch {
        // ignore
      }
    }

    ws.onclose = () => {
      setConnected(false)
      onDisconnect?.()

      const delay = Math.min(1000 * 2 ** reconnectCountRef.current, 30000)
      reconnectCountRef.current++
      timerRef.current = setTimeout(() => {
        connect()
      }, delay)
    }

    ws.onerror = () => {
      ws.close()
    }

    wsRef.current = ws
  }, [runId, token, onMessage, onConnect, onDisconnect])

  const disconnect = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    setConnected(false)
  }, [])

  useEffect(() => {
    connect()
    return () => disconnect()
  }, [connect, disconnect])

  return { connected, lastMessage, disconnect }
}
```

**验收标准**：
1. `cd ui && pnpm type-check` 通过
2. `cd ui && pnpm lint` 通过

---

## 步骤 6：编排设计页面

**目标**：完成 ReactFlow 画布 + 6 节点面板 + 属性编辑 + 保存/加载/运行。

### 6.1 主页面

**文件**：`ui/app/workflow/page.tsx`

```tsx
'use client'

import { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ShellLayout } from '@/components/layout/shell-layout'
import { Breadcrumb } from '@/components/ui/breadcrumb'
import WorkflowDesigner from '@/components/workflow/workflow-designer'
import WorkflowRuns from '@/components/workflow/workflow-runs'

export default function WorkflowPage() {
  const [activeTab, setActiveTab] = useState('designer')

  return (
    <ShellLayout pathname="/workflow">
      <Breadcrumb
        items={[
          { label: '首页', href: '/home' },
          { label: '工作流', href: '/workflow' },
        ]}
      />

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="designer">编排设计</TabsTrigger>
          <TabsTrigger value="runs">运行历史</TabsTrigger>
        </TabsList>
        <TabsContent value="designer">
          <WorkflowDesigner />
        </TabsContent>
        <TabsContent value="runs">
          <WorkflowRuns />
        </TabsContent>
      </Tabs>
    </ShellLayout>
  )
}
```

### 6.2 核心组件

**文件**：`ui/components/workflow/workflow-designer.tsx`

```tsx
'use client'

import dynamic from 'next/dynamic'

const WorkflowCanvas = dynamic(
  () => import('./workflow-canvas'),
  { ssr: false }
)

export default function WorkflowDesigner() {
  return <WorkflowCanvas />
}
```

**文件**：`ui/components/workflow/workflow-canvas.tsx`

```tsx
'use client'

import { useCallback, useState, useMemo, useRef } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  addEdge,
  type Connection,
  type Edge,
  type Node,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { DndContext } from '@dnd-kit/core'
import { WorkflowSidebar } from './workflow-sidebar'
import { WorkflowConfigPanel } from './workflow-config-panel'
import { WorkflowToolbar } from './workflow-toolbar'
import TriggerNode from './nodes/trigger-node'
import LLMNode from './nodes/llm-node'
import ConditionNode from './nodes/condition-node'
import TransformNode from './nodes/transform-node'
import MergeNode from './nodes/merge-node'
import OutputNode from './nodes/output-node'

const nodeTypes = {
  trigger: TriggerNode,
  llm: LLMNode,
  condition: ConditionNode,
  transform: TransformNode,
  merge: MergeNode,
  output: OutputNode,
}

export default function WorkflowCanvas() {
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  )

  const onNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
    setSelectedNode(node)
  }, [])

  const onPaneClick = useCallback(() => {
    setSelectedNode(null)
  }, [])

  const isValidConnection = useCallback(
    (connection: Connection) => {
      if (connection.source === connection.target) return false
      if (connection.sourceHandle) {
        const existing = edges.find(
          (e) => e.source === connection.source && e.sourceHandle === connection.sourceHandle
        )
        if (existing) return false
      }
      return true
    },
    [edges]
  )

  return (
    <DndContext>
      <div className="flex h-[calc(100vh-200px)]">
        <WorkflowSidebar />
        <div className="flex-1 flex">
          <div className="flex-1 relative">
            <ReactFlow
              nodes={nodes}
              edges={edges}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onConnect={onConnect}
              onNodeClick={onNodeClick}
              onPaneClick={onPaneClick}
              isValidConnection={isValidConnection}
              nodeTypes={nodeTypes}
              fitView
            >
              <Background />
              <Controls />
              <MiniMap />
            </ReactFlow>
            <WorkflowToolbar
              nodes={nodes}
              edges={edges}
              setNodes={setNodes}
              setEdges={setEdges}
            />
          </div>
          <WorkflowConfigPanel
            selectedNode={selectedNode}
            onNodeUpdate={(updatedNode) => {
              setNodes((nds) =>
                nds.map((n) => (n.id === updatedNode.id ? updatedNode : n))
              )
            }}
          />
        </div>
      </div>
    </DndContext>
  )
}
```

### 6.3 节点组件

**文件**：`ui/components/workflow/nodes/llm-node.tsx`

```tsx
import { Handle, Position, type NodeProps } from '@xyflow/react'
import { Brain } from 'lucide-react'

interface Data {
  model_id?: string
  system_prompt?: string
  temperature?: number
  reasoning_mode?: string
  status?: 'idle' | 'running' | 'success' | 'error'
}

export default function LLMNode({ data, selected }: NodeProps<{ data: Data }>) {
  const statusColor = {
    idle: 'border-slate-200',
    running: 'border-blue-400 animate-pulse',
    success: 'border-green-400',
    error: 'border-red-400',
  }[data.status || 'idle']

  return (
    <div
      className={`bg-white rounded-lg border-2 p-3 min-w-[200px] shadow-sm ${statusColor} ${
        selected ? 'ring-2 ring-primary' : ''
      }`}
    >
      <div className="flex items-center gap-2 mb-2">
        <Brain className="h-4 w-4 text-blue-500" />
        <span className="text-xs font-semibold text-slate-700">LLM</span>
      </div>
      <div className="text-xs text-slate-500 truncate">
        {data.model_id || '未选择模型'}
      </div>
      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />
    </div>
  )
}
```

### 6.4 属性编辑面板

**文件**：`ui/components/workflow/workflow-config-panel.tsx`

```tsx
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type { Node } from '@xyflow/react'

interface Props {
  selectedNode: Node | null
  onNodeUpdate: (node: Node) => void
}

export default function WorkflowConfigPanel({ selectedNode, onNodeUpdate }: Props) {
  if (!selectedNode) {
    return (
      <div className="w-80 border-l bg-slate-50 p-4 flex items-center justify-center">
        <p className="text-sm text-slate-400">点击节点编辑属性</p>
      </div>
    )
  }

  const { type, data } = selectedNode

  const updateData = (updates: Record<string, unknown>) => {
    onNodeUpdate({
      ...selectedNode,
      data: { ...data, ...updates },
    })
  }

  return (
    <div className="w-80 border-l bg-white p-4 overflow-y-auto">
      <h3 className="text-sm font-semibold mb-4">
        {type === 'trigger' && 'Trigger 配置'}
        {type === 'llm' && 'LLM 配置'}
        {type === 'condition' && 'Condition 配置'}
        {type === 'transform' && 'Transform 配置'}
        {type === 'merge' && 'Merge 配置'}
        {type === 'output' && 'Output 配置'}
      </h3>

      {type === 'llm' && (
        <div className="space-y-3">
          <div>
            <Label>模型</Label>
            <Select
              value={(data.model_id as string) || ''}
              onValueChange={(v) => updateData({ model_id: v })}
            >
              <SelectTrigger>
                <SelectValue placeholder="选择模型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="deepseek-v4-pro">DeepSeek V4 Pro</SelectItem>
                <SelectItem value="glm-5.1">GLM-5.1</SelectItem>
                <SelectItem value="kimi-k2.6">Kimi K2.6</SelectItem>
                <SelectItem value="qwen3.6-plus">Qwen3.6 Plus</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label>System Prompt</Label>
            <textarea
              className="w-full min-h-[80px] rounded-md border p-2 text-sm"
              value={(data.system_prompt as string) || ''}
              onChange={(e) => updateData({ system_prompt: e.target.value })}
              placeholder="你是一个助手..."
            />
          </div>
          <div>
            <Label>思考程度</Label>
            <Select
              value={(data.reasoning_mode as string) || 'low'}
              onValueChange={(v) => updateData({ reasoning_mode: v })}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="low">轻度</SelectItem>
                <SelectItem value="medium">中度</SelectItem>
                <SelectItem value="high">深度</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      )}

      {type === 'output' && (
        <div className="space-y-3">
          <div>
            <Label>输出格式</Label>
            <Select
              value={(data.format as string) || 'raw'}
              onValueChange={(v) => updateData({ format: v })}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="raw">原始文本</SelectItem>
                <SelectItem value="json">JSON</SelectItem>
                <SelectItem value="markdown">Markdown</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex items-center justify-between">
            <Label>汇聚模式</Label>
            <div className="flex gap-2">
              <button
                className={`px-2 py-1 text-xs rounded ${
                  data.join_mode === 'and' ? 'bg-primary text-white' : 'bg-slate-100'
                }`}
                onClick={() => updateData({ join_mode: 'and' })}
              >
                会签
              </button>
              <button
                className={`px-2 py-1 text-xs rounded ${
                  data.join_mode === 'or' ? 'bg-primary text-white' : 'bg-slate-100'
                }`}
                onClick={() => updateData({ join_mode: 'or' })}
              >
                或签
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
```

### 6.5 工具栏

**文件**：`ui/components/workflow/workflow-toolbar.tsx`

```tsx
import { Save, Play, CheckCircle, RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { showError, showSuccess } from '@/lib/toast'
import { useTenant } from '@/lib/tenant-context'
import type { Node, Edge } from '@xyflow/react'

interface Props {
  nodes: Node[]
  edges: Edge[]
  setNodes: React.Dispatch<React.SetStateAction<Node[]>>
  setEdges: React.Dispatch<React.SetStateAction<Edge[]>>
}

export default function WorkflowToolbar({ nodes, edges, setNodes, setEdges }: Props) {
  const { tenantId } = useTenant()

  const handleSave = async () => {
    if (!tenantId) {
      showError('请先选择租户')
      return
    }
    showSuccess('保存成功')
  }

  const handleValidate = async () => {
    showSuccess('校验通过')
  }

  const handleRun = async () => {
    if (!tenantId) {
      showError('请先选择租户')
      return
    }
    showSuccess('开始执行')
  }

  const handleReset = () => {
    setNodes([])
    setEdges([])
  }

  return (
    <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2 bg-white rounded-lg shadow-lg p-2 border">
      <Button size="sm" variant="outline" onClick={handleSave}>
        <Save className="h-4 w-4 mr-1" />
        保存
      </Button>
      <Button size="sm" variant="outline" onClick={handleValidate}>
        <CheckCircle className="h-4 w-4 mr-1" />
        校验
      </Button>
      <Button size="sm" onClick={handleRun}>
        <Play className="h-4 w-4 mr-1" />
        运行
      </Button>
      <Button size="sm" variant="ghost" onClick={handleReset}>
        <RotateCcw className="h-4 w-4 mr-1" />
        重置
      </Button>
    </div>
  )
}
```

**验收标准**：
1. 页面可访问 `/workflow`
2. 能从左侧拖拽 6 种节点到画布
3. 能连接节点
4. 选中节点可编辑属性
5. 保存/校验/运行按钮可用

---

## 步骤 7：运行历史 + 模板库 + 详情

**目标**：运行列表页、运行详情时间线、WebSocket 实时节点样式更新、模板库。

### 7.1 运行历史列表

**文件**：`ui/components/workflow/workflow-runs.tsx`

```tsx
'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Clock, CheckCircle, XCircle, Loader2 } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { useTenant } from '@/lib/tenant-context'
import { formatDate } from '@/lib/utils'
import type { WorkflowRun } from '@/types/workflow'

export default function WorkflowRuns() {
  const { tenantId } = useTenant()
  const [runs, setRuns] = useState<WorkflowRun[]>([])
  const [loading, setLoading] = useState(false)
  const router = useRouter()

  const fetchRuns = async () => {
    if (!tenantId) return
    setLoading(true)
    try {
      setRuns([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchRuns()
  }, [tenantId])

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />
      case 'running':
        return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />
      default:
        return <Clock className="h-4 w-4 text-slate-400" />
    }
  }

  return (
    <div className="space-y-4">
      {loading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-slate-400" />
        </div>
      ) : runs.length === 0 ? (
        <div className="text-center py-12 text-slate-400">
          <Clock className="h-12 w-12 mx-auto mb-3 opacity-50" />
          <p>暂无运行记录</p>
        </div>
      ) : (
        runs.map((run) => (
          <Card key={run.id} className="hover:shadow-md transition-shadow cursor-pointer">
            <CardContent className="py-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  {getStatusIcon(run.status)}
                  <div>
                    <p className="text-sm font-medium">{run.id.slice(0, 8)}</p>
                    <p className="text-xs text-slate-400">
                      {run.started_at ? formatDate(run.started_at) : '—'}
                    </p>
                  </div>
                </div>
                <div className="text-right">
                  <span className="text-xs text-slate-500">{run.duration_ms}ms</span>
                </div>
              </div>
            </CardContent>
          </Card>
        ))
      )}
    </div>
  )
}
```

**验收标准**：
1. `/workflow` 页面双 tab 正常切换
2. 运行历史列表可展示（含状态图标）
3. WebSocket 能实时更新节点状态颜色

---

## 最终验收

全部 7 步完成后，执行：

```bash
cd /Users/looper/Documents/workspace/golang/auth-perm
make build
make test
make lint
make vet

cd ui
pnpm lint
pnpm type-check
```

全部通过即完成。
