# AI 工作流编排工具 — 设计文档

## 概述

在现有 AI 矩阵推理基础上，构建一个可视化的 AI 工作流编排工具。用户通过拖拽节点来设计和执行 LLM 工作流，支持并行模型调用、条件分支、或签/会签汇聚、文本转换、多路合并等模式。

后端使用字节跳动 **Eino** 框架作为流程引擎，前端使用 `@xyflow/react` v12 实现拖拽画布。本方案综合 glm 版（Eino 引擎、简洁架构）与 ds 版（丰富节点、结构化规则、完善可观测性）的优势。

## 系统架构

```
前端 Flow JSON ─→ 后端解析 ─→ 动态构建 Eino Graph ─→ 编译执行 ─→ 结果返回
                                                        │
                                                   Callback → WebSocket 推送节点状态
                                                        │
                                                   写入 workflow_run_nodes（逐节点记录）
```

后端复用已有的 `./internal/infra/opencode/` 客户端，LLM 节点内部直接调用 `Client.Chat()`。

## 决策记录

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 流程引擎 | Eino compose.Graph | 字节官方框架，callback/branch/interrupt 原生支持 |
| 画布 | @xyflow/react v12 + @dnd-kit/core | v12 SSR 兼容，拖拽体验成熟 |
| 节点类型 | 6 种 | trigger + llm + condition + transform + merge + output，覆盖完整编排场景 |
| 条件规则 | 结构化 JSON 规则构建器 | 前端表单式配置，零语法错误，支持嵌套 AND/OR |
| 存储 | PostgreSQL JSONB | 复用现有基础设施 |
| 实时状态 | WebSocket（gorilla/websocket） | 低延迟推送节点状态 |
| 工作流输入 | 自由文本 + 结构化 JSON Schema | 默认自由文本，高级场景支持结构化字段 |
| Multi-LLM | N 个独立 LLM 节点 + merge 节点 | 用户手动连线，语义清晰 |
| 执行模式 | 同步 + 异步 + WebSocket 实时 | 同步用于简单场景，异步防超时，WS 用于实时监控 |
| OR-join | channel 竞争 + context 取消 | 第一成功结果即返回，取消其余分支 |

## 节点类型（6 种）

| 节点 | 图标 | 入边 | 出边 | 配置数据 | 说明 |
|------|------|------|------|---------|------|
| **trigger** | 📥 | 无 | 1 | `input_schema?: Record<string, string>` | 工作流入口，默认自由文本输入，可选结构化字段 |
| **llm** | 🧠 | 1 | 1 | `model_id`, `system_prompt`, `temperature`, `reasoning_mode` | 调用单个 LLM，复用 OpenCode client |
| **condition** | 🔀 | 1 | N | `rules: Rule[]`, `default_handle` | 结构化规则判断，按序求值，命中即路由 |
| **transform** | 🔧 | 1 | 1 | `operation: Operation`, `params` | 文本转换：regex 提取/替换、trim、markdown→text、截断 |
| **merge** | 🔗 | N | 1 | `strategy: "concat"\|"first"\|"join"` | 合并多上游结果（拼接/取首/自定义分隔符） |
| **output** | 📤 | N | 无 | `format: "raw"\|"json"\|"markdown"`, `join_mode: "and"\|"or"` | 汇聚并格式化最终输出 |

### 连接语义

- **1 source → N target**：扇出（并行分离），N 个下游节点并发执行
- **N source → 1 target**：汇聚，由 output 节点的 `join_mode` 决定：
  - `and`（会签）：等待所有上游完成，合并输出
  - `or`（或签）：第一个成功结果即返回，其余分支取消

## 条件规则 — 结构化 JSON 规则构建器

不使用自由文本 DSL，改用结构化 JSON，前端用表单组件构建，后端直接递归求值。

### 规则结构

```json
{
  "logic": "AND",
  "rules": [
    {
      "field": "content",
      "operator": "contains",
      "value": "error",
      "negate": false
    },
    {
      "field": "length",
      "operator": "gt",
      "value": 200,
      "negate": false
    }
  ]
}
```

### 支持的运算符

| 运算符 | 适用 field | value 类型 | 说明 |
|--------|-----------|-----------|------|
| `contains` | content | string | 包含子串 |
| `not_contains` | content | string | 不包含 |
| `equals` | content | string | 完全相等 |
| `matches` | content | string（正则） | 正则匹配 |
| `starts_with` | content | string | 以...开头 |
| `ends_with` | content | string | 以...结尾 |
| `gt` / `gte` / `lt` / `lte` | length | number | 长度比较 |
| `is_empty` / `not_empty` | content | — | 内容空/非空 |

### 嵌套规则组

```json
{
  "logic": "OR",
  "rules": [
    { "field": "content", "operator": "contains", "value": "pass" },
    {
      "logic": "AND",
      "rules": [
        { "field": "length", "operator": "gt", "value": 100 },
        { "field": "content", "operator": "not_contains", "value": "error" }
      ]
    }
  ]
}
```

### condition 分支配置

```json
{
  "branches": [
    {
      "handle": "has_error",
      "rule": {
        "logic": "OR",
        "rules": [{ "field": "content", "operator": "contains", "value": "error" }]
      }
    },
    {
      "handle": "too_short",
      "rule": {
        "logic": "AND",
        "rules": [{ "field": "length", "operator": "lt", "value": 50 }]
      }
    }
  ],
  "default_handle": "ok"
}
```

按顺序求值，首个满足的 branch 路由到对应 handle；无命中走 `default_handle`。

### 前端规则构建器 UI

condition 节点属性面板提供递归表单式构建器：
- 顶层：逻辑选择（AND/OR）
- 每行：字段下拉 + 运算符下拉 + 值输入 + 删除按钮
- 支持「添加规则组」嵌套
- 每 branch 一个 handle 名输入框
- 实时预览：底部显示等效自然语言描述

## transform 节点详解

对上游输出做后处理：

| operation | params | 说明 |
|-----------|--------|------|
| `regex_extract` | `pattern`, `group_index?: 0` | 提取正则匹配组 |
| `regex_replace` | `pattern`, `replacement` | 正则替换 |
| `trim` | — | 去首尾空白 |
| `markdown_to_text` | — | 去除 Markdown 标记 |
| `extract_json` | — | 提取文本中的第一个 JSON 块 |
| `truncate` | `max_length: number` | 截断到指定字符数 |
| `to_uppercase` / `to_lowercase` | — | 大小写转换 |

## merge 节点详解

汇聚多个上游 LLM 节点输出：

| strategy | 行为 |
|----------|------|
| `concat` | 用 `\n\n---\n\n` 分隔拼接各输出 |
| `first` | 取第一个完成的输出（OR-join 语义） |
| `join` | 用自定义 `delimiter` 拼接各输出 |

## 执行模式

| 模式 | 触发方式 | 超时 | 适用场景 |
|------|---------|------|---------|
| **同步** | POST `/execute?mode=sync` | 120s HTTP timeout | 简单流程、即时返回 |
| **异步** | POST `/execute?mode=async` | 无 HTTP 超时，后台执行 | 复杂并行流程、防浏览器超时 |
| **WS 实时** | WebSocket `/ws/run/:runId` | 持续推送 | 实时监控节点状态 |

异步模式返回 `{ run_id }`，前端轮询 `GET /runs/:runId` 或 WS 订阅获取进度。

## 数据库设计

### workflows 表

```sql
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    flow_json JSONB NOT NULL,
    template_id UUID,                          -- 若从模板创建，引用源模板
    status VARCHAR(16) NOT NULL DEFAULT 'draft', -- draft / published / archived
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflows_tenant_account ON workflows(tenant_id, account_id);
```

### workflow_runs 表

```sql
CREATE TABLE IF NOT EXISTS workflow_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    execution_mode VARCHAR(8) NOT NULL DEFAULT 'sync',  -- sync / async
    input_text TEXT,
    input_json JSONB,                                    -- 结构化输入
    result_json JSONB,                                   -- 最终输出
    status VARCHAR(16) NOT NULL DEFAULT 'pending',       -- pending / running / success / failed / cancelled
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    error TEXT
);

CREATE INDEX IF NOT EXISTS idx_wf_runs_workflow ON workflow_runs(workflow_id);
CREATE INDEX IF NOT EXISTS idx_wf_runs_tenant ON workflow_runs(tenant_id, account_id, started_at DESC);
```

### workflow_run_nodes 表（逐节点执行记录）

```sql
CREATE TABLE IF NOT EXISTS workflow_run_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    node_id VARCHAR(128) NOT NULL,             -- 画布节点 ID（如 "llm_1"）
    node_type VARCHAR(32) NOT NULL,            -- trigger / llm / condition / transform / merge / output
    node_label VARCHAR(128),                   -- 用户可见标签
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    input_json JSONB,                          -- 节点输入
    output_json JSONB,                         -- 节点输出
    error TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    retry_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_wf_run_nodes_run ON workflow_run_nodes(run_id);
CREATE INDEX IF NOT EXISTS idx_wf_run_nodes_run_node ON workflow_run_nodes(run_id, node_id);
```

### flow_json 结构

React Flow `toObject()` 输出直接存储：

```json
{
  "nodes": [
    {
      "id": "input_1",
      "type": "trigger",
      "position": { "x": 100, "y": 200 },
      "data": { "placeholder": "输入问题..." }
    },
    {
      "id": "llm_1",
      "type": "llm",
      "position": { "x": 400, "y": 100 },
      "data": {
        "model_id": "deepseek-v4-pro",
        "system_prompt": "你是一个助手",
        "temperature": 0.7,
        "reasoning_mode": "low"
      }
    },
    {
      "id": "cond_1",
      "type": "condition",
      "position": { "x": 700, "y": 200 },
      "data": {
        "branches": [
          { "handle": "has_error", "rule": { "logic": "OR", "rules": [...] } }
        ],
        "default_handle": "ok"
      }
    },
    {
      "id": "merge_1",
      "type": "merge",
      "position": { "x": 1000, "y": 200 },
      "data": { "strategy": "concat" }
    },
    {
      "id": "output_1",
      "type": "output",
      "position": { "x": 1300, "y": 200 },
      "data": { "format": "raw", "join_mode": "and" }
    }
  ],
  "edges": [
    { "id": "e1", "source": "input_1", "target": "llm_1" },
    { "id": "e2", "source": "llm_1", "target": "cond_1" },
    { "id": "e3", "source": "cond_1", "sourceHandle": "ok", "target": "merge_1" }
  ],
  "viewport": { "x": 0, "y": 0, "zoom": 1 }
}
```

## Eino 图构建引擎

### 核心映射

| React Flow type | Eino 构建方式 | 说明 |
|-----------------|--------------|------|
| `trigger` | `AddLambdaNode` — passthrough | 输出 = inputText / inputJSON |
| `llm` | `AddLambdaNode` — 调用 `opencode.Client.Chat()` | 复用现有 client |
| `condition` | `AddLambdaNode`（规则求值）+ `AddBranch`（按 handle 路由） | 结构化规则递归求值 |
| `transform` | `AddLambdaNode` — 按 operation 执行转换 | 纯文本处理 |
| `merge` | `AddLambdaNode` — 按 strategy 合并上游结果 | 多入边 = 等待全部完成 |
| `output` | `AddLambdaNode` — 格式化并返回最终结果 | join_mode 控制汇聚逻辑 |

### 构建流程

```
flow_json
  → 解析 nodes + edges
  → 按类型注册 Eino LambdaNode（每个 node_id 唯一对应一个 Lambda）
  → 解析 edges：
      - 普通边 → g.AddEdge(source, target)
      - condition 出边 → g.AddBranch(conditionID, branchFunc)
  → g.Compile(ctx)
  → runnable.Invoke(ctx, input) 或 runnable.Stream(ctx, input)
```

### condition 节点 Eino 实现

```go
branchFunc := compose.NewGraphBranch(
    func(ctx context.Context, in *NodeOutput) (string, error) {
        for _, branch := range nodeConfig.Branches {
            matched, err := evaluateRuleGroup(in.Content, branch.Rule)
            if err != nil { continue }
            if matched { return branch.Handle, nil }
        }
        return nodeConfig.DefaultHandle, nil
    },
    allHandleNames, // map[string]bool{"has_error": true, "ok": true, ...}
)
g.AddBranch(conditionNodeID, branchFunc)
```

### OR-join（或签）实现

Eino 默认 `AllPredecessor`（会签）。OR-join 通过自定义 Lambda 实现：

```go
// output 节点 join_mode = "or" 时
orJoinLambda := compose.InvokableLambda(func(ctx context.Context, in *NodeOutput) (*NodeOutput, error) {
    results := getParallelResults(ctx) // 从 context 读取各上游已写入的结果
    for _, res := range results {
        if res.Error == "" { return res, nil }
    }
    return nil, errors.New("all branches failed")
})

// 每个上游 LLM 节点完成后，向 context 写入结果
// 使用 context.WithValue 传递共享 results map
// 第一个成功结果被 orJoinLambda 消费后，其余分支通过 context cancellation 优雅停止
```

### 回调与 WebSocket 推送

```go
handler := callbacks.NewHandlerBuilder().
    OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
        wsHub.Broadcast(runID, map[string]any{
            "type": "node_start",
            "node_id": info.Name,
            "node_type": info.Type,
        })
        writeNodeStatus(runID, info.Name, "running")
        return ctx
    }).
    OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
        wsHub.Broadcast(runID, map[string]any{
            "type": "node_end",
            "node_id": info.Name,
            "duration_ms": getDuration(ctx, info.Name),
        })
        writeNodeStatus(runID, info.Name, "success")
        return ctx
    }).
    OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
        wsHub.Broadcast(runID, map[string]any{
            "type": "node_error",
            "node_id": info.Name,
            "error": err.Error(),
        })
        writeNodeStatus(runID, info.Name, "error")
        return ctx
    }).
    Build()

runnable.Invoke(ctx, input, compose.WithCallbacks(handler))
```

### 规则求值器

```go
func evaluateRuleGroup(content string, group RuleGroup) (bool, error) {
    results := make([]bool, len(group.Rules))
    for i, rule := range group.Rules {
        if rule.SubGroup != nil {
            r, err := evaluateRuleGroup(content, *rule.SubGroup)
            if err != nil { return false, err }
            results[i] = r
            continue
        }
        r, err := evaluateSingleRule(content, rule)
        if err != nil { return false, err }
        results[i] = r
    }
    return combineLogic(group.Logic, results), nil
}

func evaluateSingleRule(content string, rule SingleRule) (bool, error) {
    fieldValue := extractField(content, rule.Field) // content 或 length
    switch rule.Operator {
    case "contains":
        result := strings.Contains(fieldValue, rule.Value)
        if rule.Negate { result = !result }
        return result, nil
    case "gt":
        length := len([]rune(content))
        val, _ := strconv.Atoi(rule.Value)
        result := length > val
        if rule.Negate { result = !result }
        return result, nil
    // ... 其余运算符
    }
}
```

## WebSocket 实时状态

### 消息协议

```json
// 节点开始
{"type": "node_start", "run_id": "xxx", "node_id": "llm_1", "node_type": "llm", "ts": "2026-01-15T10:00:00Z"}

// 节点完成
{"type": "node_end", "run_id": "xxx", "node_id": "llm_1", "duration_ms": 2300, "ts": "..."}

// 节点错误
{"type": "node_error", "run_id": "xxx", "node_id": "cond_1", "error": "rule eval failed", "ts": "..."}

// 运行结束
{"type": "run_end", "run_id": "xxx", "status": "success", "result": "...", "ts": "..."}

// 心跳
{"type": "ping"}
```

### ws_hub.go 设计

```go
type Hub struct {
    clients    map[string]map[*Client]bool  // run_id → clients set
    register   chan *Client
    unregister chan *Client
    broadcast  chan Message
    mu         sync.RWMutex
}

type Client struct {
    runID string
    conn  *websocket.Conn
    send  chan []byte
}
```

Client 连接时 URL 参数 `?runId=xxx`，Hub 按 run 分组广播。支持自动心跳（30s ping/pong）和断线重连。

## 图验证层

保存或执行前校验：

| # | 验证项 | 说明 | 严重度 |
|---|--------|------|--------|
| 1 | trigger 唯一性 | 有且仅有一个 trigger 节点 | error |
| 2 | output 存在性 | 至少一个 output 节点 | error |
| 3 | trigger 可达 output | 从 trigger 到 output 存在有向路径 | error |
| 4 | 孤立节点 | 无入边且非 trigger 的节点 | warning |
| 5 | condition 分支数 | condition 节点至少 2 条出边 | error |
| 6 | merge 入边数 | merge 节点至少 2 条入边 | error |
| 7 | LLM 模型必填 | llm 节点必须指定 model_id | error |
| 8 | 环路检测 | 无有向环（DFS） | error |
| 9 | output 入边 | output 节点至少 1 条入边 | error |
| 10 | trigger 无入边 | trigger 节点入度为 0 | error |

校验失败返回结构化错误列表（含 node_id + message），前端高亮定位问题节点。

## 工作流模板

预置 4 个模板，存储在 `workflows` 表中（`template_id IS NULL` 且 `status = 'template'`）：

| 模板名 | 结构 | 用途 |
|--------|------|------|
| **AI 矩阵推理** | trigger → 5×llm → merge(concat) → output | 5 模型并行回答 |
| **质量卫士** | trigger → llm → condition → output（达标）/ llm（重写）→ output | 质量不合格自动重写 |
| **多模型对比** | trigger → 3×llm → merge(concat) → output | 多模型回答拼接对比 |
| **两阶段分析** | trigger → llm（分析）→ llm（总结）→ output | 先分析再总结串行管道 |

用户从模板创建时生成新 workflow，`template_id` 指向源模板。

## 后端模块结构

```
internal/domain/workflow/
├── module.go                         # DI 注册 RegisterWorkflowDomain
├── constant/
│   ├── node_type.go                  # NodeType 枚举 + IsValidNodeType
│   ├── run_status.go                 # RunStatus 枚举 + IsValidRunStatus
│   ├── operators.go                  # ConditionOperator 枚举 + 白名单校验
│   ├── transforms.go                 # TransformOperation 枚举
│   └── strategies.go                 # MergeStrategy / OutputFormat / JoinMode
├── dm/
│   ├── workflow_do.go                # Workflow GORM DO
│   ├── workflow_run_do.go            # WorkflowRun GORM DO
│   └── workflow_run_node_do.go       # WorkflowRunNode GORM DO
├── repo/
│   ├── workflow_repo.go              # Workflow CRUD
│   ├── workflow_run_repo.go          # WorkflowRun CRUD
│   ├── workflow_run_node_repo.go     # WorkflowRunNode CRUD
│   └── workflow_template_repo.go     # 模板查询
├── service/
│   ├── workflow_service.go           # 工作流 CRUD 逻辑
│   ├── engine.go                     # Flow JSON → Eino Graph 构建器 + 编译 + 执行
│   ├── graph_validator.go            # 图结构 10 项校验
│   ├── rule_evaluator.go             # 结构化规则递归求值
│   ├── node_executor.go              # 各节点类型执行器（llm/transform/merge 等）
│   └── ws_hub.go                     # WebSocket Hub
├── handler/
│   ├── workflow_handler.go            # REST handlers（CRUD + Execute + Validate）
│   ├── template_handler.go            # 模板查询
│   └── ws_handler.go                  # WebSocket upgrade handler
├── dto/
│   ├── workflow_dto.go               # DTO 转换
│   └── workflow_run_dto.go           # 运行记录 DTO
└── vo/
    ├── workflow_vo.go                # 请求/响应 VO
    ├── condition_vo.go               # 条件规则校验（递归校验 RuleGroup）
    └── flow_vo.go                    # flow_json 结构校验
```

## 后端 API

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/v1/workflow` | ListWorkflows | 工作流列表（支持 `?type=template` 过滤） |
| POST | `/api/v1/workflow` | CreateWorkflow | 创建工作流 |
| GET | `/api/v1/workflow/:id` | GetWorkflow | 详情（含 flow_json） |
| PUT | `/api/v1/workflow/:id` | UpdateWorkflow | 更新 |
| DELETE | `/api/v1/workflow/:id` | DeleteWorkflow | 删除 |
| POST | `/api/v1/workflow/:id/execute` | ExecuteWorkflow | 执行（`?mode=sync\|async`） |
| POST | `/api/v1/workflow/:id/validate` | ValidateWorkflow | 校验图结构（不执行） |
| POST | `/api/v1/workflow/:id/clone` | CloneWorkflow | 克隆工作流 |
| GET | `/api/v1/workflow/templates` | ListTemplates | 模板列表 |
| GET | `/api/v1/workflow/:id/runs` | ListRuns | 运行历史 |
| GET | `/api/v1/workflow/runs/:runId` | GetRun | 运行详情 |
| GET | `/api/v1/workflow/runs/:runId/nodes` | GetRunNodes | 节点级执行明细 |
| POST | `/api/v1/workflow/runs/:runId/cancel` | CancelRun | 取消运行 |
| WS | `/api/v1/ws/run/:runId` | WSHandler | WebSocket 实时状态 |

## 前端结构

```
ui/
├── app/workflow/
│   ├── page.tsx                        # 双 tab 页面：[编排设计] [运行历史]
│   ├── [id]/
│   │   ├── page.tsx                    # 工作流详情/编辑
│   │   └── runs/
│   │       └── [runId]/
│   │           └── page.tsx            # 运行详情（节点时间线）
│   └── templates/
│       └── page.tsx                    # 模板库
├── components/workflow/
│   ├── workflow-canvas.tsx             # ReactFlow 主画布（DndContext + ReactFlow）
│   ├── workflow-sidebar.tsx            # 左侧节点面板（6 节点 + 模板快捷入口）
│   ├── workflow-config-panel.tsx       # 右侧属性面板（按节点类型渲染不同表单）
│   ├── workflow-toolbar.tsx            # 底部工具栏（保存/校验/运行/撤销/重做）
│   ├── workflow-validation-panel.tsx   # 校验结果面板（高亮问题节点）
│   ├── nodes/
│   │   ├── trigger-node.tsx            # 入口节点
│   │   ├── llm-node.tsx                # LLM 节点
│   │   ├── condition-node.tsx          # 条件节点
│   │   ├── transform-node.tsx          # 转换节点
│   │   ├── merge-node.tsx              # 合并节点
│   │   └── output-node.tsx             # 输出节点
│   ├── rules/
│   │   └── rule-builder.tsx            # 结构化规则构建器（递归组件）
│   └── run-detail/
│       ├── run-timeline.tsx            # 运行时间线（节点级）
│       ├── run-status-badge.tsx        # 状态徽章
│       └── node-output-viewer.tsx      # 节点输出查看器
├── lib/api/workflow.ts                 # API 封装
├── hooks/
│   ├── use-workflow-ws.ts              # WebSocket hook（连接 + 自动重连 + 消息处理）
│   └── use-workflow-validation.ts      # 图校验 hook
└── types/
    ├── workflow.ts                     # 工作流类型
    ├── nodes.ts                        # 节点联合类型
    └── rules.ts                        # 规则构建器类型
```

### 技术要点

- ReactFlow 用 `next/dynamic` 导入，`ssr: false`
- 拖拽用 `@dnd-kit/core`（`useDraggable` + `useDroppable`），放下时 `screenToFlowPosition` 定位
- 节点属性编辑：选中节点 → 右侧面板显示表单 → 修改 `node.data`
- 连线验证（`isValidConnection`）：禁止自环、禁止 condition 单源连接、禁止同一 sourceHandle 重复连线
- 保存前调用 `POST /validate` 全量校验，错误返回结构化列表，前端高亮问题节点
- WebSocket 断开自动重连（指数退避）
- 撤销/重做：捕获 `onNodesChange`/`onEdgesChange` 事件，维护 history stack

### 页面布局

```
┌──────────────────────────────────────────────────────────────┐
│  [编排设计]        [运行历史]                                   │
├──────────────────┬──────────────────┬────────────────────────┤
│  左侧节点面板     │  中间画布         │  右侧属性面板           │
│  (workflow-      │  (workflow-      │  (workflow-config-      │
│   sidebar)       │   canvas)        │   panel)               │
│                  │                  │                        │
│  📥 trigger       │  ┌──┐   ┌──┐    │  类型: condition        │
│  🧠 llm          │  │in│──▶│ll│    │  ─────────────────────  │
│  🔀 condition    │  └──┘   └──┘    │  Branch: "has_error"    │
│  🔧 transform    │           │     │  [规则组 AND]           │
│  🔗 merge        │           ▼     │  ┌──────────────────┐   │
│  📤 output       │  ┌──┐   ┌──┐    │  │ [content] [contains│   │
│                  │  │ll│◀──│co│    │  │ ▼] [ "error" ] [×]│   │
│  ──────────────  │  └──┘   └──┘    │  └──────────────────┘   │
│  模板库           │           │     │  [+ 添加规则]           │
│  ┌────────────┐  │           ▼     │  默认: [fallback ▼]     │
│  │ AI矩阵推理 │  │         ┌──┐    │                         │
│  │ 质量卫士   │  │         │me│    │                         │
│  │ ...       │  │         └──┘    │                         │
│  └────────────┘  │           │     │                         │
│                  │           ▼     │                         │
│                  │         ┌──┐    │                         │
│                  │         │ou│    │                         │
│                  │         └──┘    │                         │
├──────────────────┴──────────────────┴────────────────────────┤
│  [保存] [校验] [发布] [▶ 运行]  [撤销] [重做]                  │
└──────────────────────────────────────────────────────────────┘
```

## 权限配置

| code | name | resource 类型 | 说明 |
|------|------|--------------|------|
| `menu:workflow` | 工作流菜单 | menu | 侧边栏菜单可见 |
| `workflow.read` | 查看工作流 | button | 列表/详情查看 |
| `workflow.write` | 创建/编辑工作流 | button | 创建、编辑、执行、克隆 |
| `workflow.delete` | 删除工作流 | button | 删除操作 |

前端权限控制：
- 侧边栏菜单 `workflow` → `menu:workflow`
- tab 按钮 `workflow.tab.designer` → `workflow.read`
- tab 按钮 `workflow.tab.runs` → `workflow.read`
- 保存/发布/运行按钮 → `workflow.write`
- 删除按钮 → `workflow.delete`

后端路由：统一挂载 `AuthMiddleware` + `APIPermissionMiddleware`。

## 新增依赖

| 层 | 包 | 用途 |
|----|----|------|
| 后端 | `github.com/cloudwego/eino/compose` | 图编排引擎 |
| 后端 | `github.com/gorilla/websocket` | WebSocket |
| 前端 | `@xyflow/react` v12 | ReactFlow 画布 |
| 前端 | `@dnd-kit/core` + `@dnd-kit/utilities` | 节点拖拽 |

## 实施步骤

| # | 步骤 | 内容 | 范围 |
|---|------|------|------|
| 1 | Migration + Domain 骨架 | 建 3 张表；DO + Repo + Module 注册 + Container 装配 + Route 占位 | 后端 |
| 2 | Eino 引擎 + 规则求值 + 图校验 | `engine.go`（Flow JSON → Eino Graph）；`rule_evaluator.go`（结构化规则递归求值）；`graph_validator.go`（10 项校验） | 后端 |
| 3 | Handler + WebSocket | CRUD 14 端点 + Execute + Validate；`ws_hub.go` + `ws_handler.go`（实时广播节点状态） | 后端 |
| 4 | 权限种子 + is_system 修复 | migration 补充 workflow 权限资源；新建权限页 `is_system` 开关加超管守卫 | 前后端 |
| 5 | 前端类型 + API + WS Hook | `types/workflow.ts` + `types/nodes.ts` + `types/rules.ts`；`lib/api/workflow.ts`；`hooks/use-workflow-ws.ts` | 前端 |
| 6 | 编排设计页面 | ReactFlow 画布 + 6 节点面板 + 6 节点组件 + 规则构建器 + 属性面板 + 校验 + 保存/加载/运行/撤销重做 | 前端 |
| 7 | 运行历史 + 模板库 + 详情 | 运行列表 + 节点时间线 + WS 实时渲染 + 模板库浏览/使用 + 克隆 | 前端 |

## 注意事项

1. **Eino 与现有代码库集成**：Eino 的 Go 模块依赖需检查与现有 GORM/Gin 版本的兼容性。如冲突，优先升级 GORM/Gin 到 Eino 兼容版本。
2. **OpenCode client 复用**：LLM 节点直接使用 `internal/infra/opencode/client.go` 的 `Chat()` 方法，无需额外封装。注意 120s 超时与 WebSocket 推送的协调。
3. **WebSocket 认证**：WS 连接建立时需校验 JWT token（复用 `AuthMiddleware` 逻辑），拒绝非法连接。
4. **并发安全**：workflow_run_nodes 的逐节点写入需考虑并发（多个 LLM 节点同时完成时并发写 DB），使用事务或乐观锁。
5. **flow_json 版本兼容**：后续若扩展节点类型或 data 结构，需在 flow_json 中存储 `version` 字段，解析时做版本适配。
6. **条件规则求值性能**：规则求值在内存中完成，不涉及 DB 或外部 API，性能开销可忽略。
7. **模板复制**：从模板创建工作流时，`flow_json` 需深拷贝，避免修改模板影响已创建的工作流。
