# AI 工作流编排工具 — 设计文档

## 概述

在现有 AI 矩阵推理的基础上，构建一个可视化的 AI 工作流编排工具。用户通过拖拽节点来设计和执行 LLM 工作流，支持并行模型调用、条件分支、或签/会签汇聚等模式。

后端使用字节跳动 Eino 框架作为流程引擎，前端使用 @xyflow/react 实现拖拽画布。

## 决策记录

| 决策项 | 选择 |
|--------|------|
| 流程引擎 | Eino compose.Graph |
| 画布 | @xyflow/react v12 + @dnd-kit/core |
| 条件判断 | govaluate 表达式 DSL |
| 存储 | PostgreSQL JSONB |
| 实时状态 | WebSocket（gorilla/websocket） |
| 工作流输入 | 自由文本输入框 |
| Multi-LLM 表达 | N 个独立 LLM 节点，用户手动连线扇出/汇聚 |
| 或签 | MVP 支持 OR-join + AND-join |

## 系统架构

```
前端 Flow JSON ─→ 后端解析 ─→ 动态构建 Eino Graph ─→ 编译执行 ─→ 结果返回
                                                        │
                                                   Callback → WebSocket 推送节点状态
```

后端复用已有的 `./internal/infra/opencode/` 客户端，LLM 节点内部直接调用 `Client.Chat()`。

## 节点类型（4 种）

| 节点 | 画布行为 | 配置数据 | 入边 | 出边 |
|------|---------|---------|------|------|
| **input** | 工作流入口，用户在此填入文本 | `placeholder: string` | 无 | 1 个 |
| **llm** | 调用一个 LLM，复用 OpenCode client | `model_id`, `system_prompt`, `reasoning_mode` | 1 个 | 1 个 |
| **condition** | DSL 表达式判断，路由到不同分支 | `expression`（见下方 DSL），每分支一个 handle | 1 个 | N 个 |
| **output** | 收集输出作为最终结果 | 无，可选指定 join_mode: `and` / `or` | 1-N 个 | 无 |

连接语义：
- **1 source → N target**：扇出（并行分离），N 个下游节点同时执行
- **N source → 1 target**：汇聚，由 output 节点指定 `and`（会签，等所有上游完成）或 `or`（或签，取第一个完成的结果）

## 条件 DSL

使用 `github.com/Knetic/govaluate` 作为表达式引擎，注册自定义函数绑定上个节点输出。

```
output.contains("error")
output.length > 100 AND NOT output.contains("retry")
output.matches("^[0-9]+$")
output.startsWith("OK:") AND output.length < 500
```

支持的函数：
- `contains(str)` — 文本包含
- `not_contains(str)` — 文本不包含
- `equals(str)` — 文本相等
- `matches("regex")` — 正则匹配
- `startsWith(str)` — 文本以...开头
- `endsWith(str)` — 文本以...结尾
- `length` — 文本字数（数字属性）

逻辑运算符：`AND`, `OR`, `NOT`，支持括号分组。

condition 节点配置每分支的 handle ID（如 `true`/`false` 或自定义名），表达式求值为 `true` 时走对应分支。

### condition 分支路由

condition 节点的 `data` 结构：

```json
{
  "rules": [
    {
      "expression": "output.contains(\"error\") AND output.length > 100",
      "handle": "branch_1"
    },
    {
      "expression": "output.contains(\"error\")",
      "handle": "branch_2"
    }
  ],
  "default_handle": "else"
}
```

按顺序求值，命中第一条为 true 的规则即路由到对应 handle；无命中走 default_handle。

## 汇聚模式（OR-join / AND-join）

### AND-join（会签）

使用 Eino 原生 `AllPredecessor` 触发模式。output 节点的 `join_mode: "and"` 时，等待所有上游 LLM/condition 节点完成，合并所有输出为最终结果。

### OR-join（或签）

output 节点的 `join_mode: "or"` 时：每个上游节点完成后向共享 channel 写入结果；output 节点取第一个成功结果，通过 `context.WithCancel` 取消其余分支，返回该结果。实现了 channel 竞争 + context 取消的组合模式。

## 数据库设计

```sql
-- 工作流定义
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    flow_json JSONB NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflows_tenant_account ON workflows(tenant_id, account_id);

-- 工作流执行记录
CREATE TABLE IF NOT EXISTS workflow_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    input_text TEXT,
    result_json JSONB,
    node_statuses JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(16) NOT NULL DEFAULT 'running',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    finished_at TIMESTAMP WITH TIME ZONE,
    error TEXT
);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow ON workflow_runs(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_tenant ON workflow_runs(tenant_id, account_id, started_at DESC);
```

### flow_json 结构

React Flow `toObject()` 的输出直接存储：

```json
{
  "nodes": [
    {
      "id": "input_1",
      "type": "input",
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
        "reasoning_mode": "low"
      }
    },
    {
      "id": "cond_1",
      "type": "condition",
      "position": { "x": 700, "y": 200 },
      "data": {
        "rules": [
          { "expression": "output.contains(\"error\")", "handle": "error" }
        ],
        "default_handle": "ok"
      }
    },
    {
      "id": "output_1",
      "type": "output",
      "position": { "x": 1000, "y": 200 },
      "data": { "join_mode": "and" }
    }
  ],
  "edges": [
    { "id": "e1", "source": "input_1", "target": "llm_1" },
    { "id": "e2", "source": "llm_1", "target": "cond_1" },
    { "id": "e3", "source": "cond_1", "sourceHandle": "ok", "target": "output_1" }
  ],
  "viewport": { "x": 0, "y": 0, "zoom": 1 }
}
```

## WebSocket 实时状态

### 消息协议

Eino callback 在 `OnStart`/`OnEnd`/`OnError` 中推送：

```json
{"type": "node_start",  "run_id": "xxx", "node_id": "llm_1", "node_type": "llm"}
{"type": "node_end",   "run_id": "xxx", "node_id": "llm_1", "output": "...", "duration_ms": 2300}
{"type": "node_error",  "run_id": "xxx", "node_id": "cond_1", "error": "rule eval failed"}
{"type": "run_end",     "run_id": "xxx", "status": "success", "result": "..."}
```

### 前端渲染

根据消息类型更新对应节点的 data 中的 status：

```ts
// node.data.status 取值
// 'idle' | 'running'（蓝色脉冲边框）| 'success'（绿色边框）| 'error'（红色边框）
```

### ws_hub.go 设计

```go
type Hub struct {
    clients    map[string]*Client  // run_id → client
    register   chan *Client
    unregister chan *Client
}

type Client struct {
    runID string
    conn  *websocket.Conn
    send  chan []byte
}
```

Client 连接时带上 `runId` 参数，Hub 按 run 分组广播。

## 后端模块结构

```
internal/domain/workflow/
├── module.go                    # DI 注册（RegisterWorkflowDomain）
├── constant/
│   ├── node_type.go             # NodeTypeInput/LLM/Condition/Output
│   ├── run_status.go            # StatusRunning/Success/Failed
│   └── join_mode.go             # JoinModeAnd/JoinModeOr
├── dm/
│   ├── workflow_do.go           # Workflow GORM DO
│   └── workflow_run_do.go       # WorkflowRun GORM DO
├── repo/
│   ├── workflow_repo.go         # Workflow CRUD（List/Create/Get/Update/Delete）
│   └── workflow_run_repo.go     # WorkflowRun CRUD（Create/Get/List）
├── service/
│   ├── workflow_service.go      # 工作流 CRUD 逻辑
│   ├── engine.go                # Flow JSON → Eino Graph 构建器 + 执行
│   ├── condition_eval.go        # govaluate 表达式求值器 + 自定义函数注册
│   └── ws_hub.go                # WebSocket Hub（run_id → conn 映射，广播）
├── handler/
│   ├── workflow_handler.go      # REST handlers（CRUD + Execute）
│   └── ws_handler.go            # WebSocket upgrade handler
├── dto/
│   └── workflow_dto.go          # DTO（dm → dto 转换）
└── vo/
    ├── workflow_vo.go           # 请求/响应 VO
    ├── node_type_vo.go          # 节点类型校验（白名单 map + IsValidNodeType）
    └── condition_vo.go          # 条件规则校验
```

## Eino 图构建引擎（engine.go）

### 构建流程

```
flow_json → parse nodes → 按类型构建 Eino 节点 → parse edges → 构建边/分支 → compile → execute
```

### 节点映射

| React Flow type | Eino 构建 |
|-----------------|-----------|
| `input` | `AddLambdaNode` — passthrough（输出 = inputText） |
| `llm` | `AddLambdaNode` — 内部调用 `opencode.Client.Chat()` |
| `condition` | `AddLambdaNode`（条件求值）+ `AddBranch`（按 handle 路由） |
| `output` | `AddLambdaNode` — 收集上游结果，按 join_mode 处理 |

### 条件节点实现

```go
// 为每个 condition 节点添加 branch
branchFunc := compose.NewGraphBranch(
    func(ctx context.Context, in *NodeOutput) (string, error) {
        for _, rule := range nodeConfig.Rules {
            matched, err := evalExpression(rule.Expression, in.Content)
            if err != nil { return rule.DefaultHandle, nil }
            if matched { return rule.Handle, nil }
        }
        return nodeConfig.DefaultHandle, nil
    },
    allHandleNames, // 所有可能的 handle ID
)
g.AddBranch(conditionNodeID, branchFunc)
```

### OR-join 实现要点

```go
// OR-join output 节点
type orJoinNode struct {
    results chan *NodeOutput
}

func (n *orJoinNode) Invoke(ctx context.Context, input *NodeOutput) (*NodeOutput, error) {
    select {
    case result := <-n.results:
        return result, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

调用时：使用 `context.WithCancel`，每个上游节点完成后写入 channel；OR-join 节点取第一个；defer cancel() 取消其他分支。

## 后端 API

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/v1/journal/workflows` | ListWorkflows | 工作流列表 |
| POST | `/api/v1/journal/workflows` | CreateWorkflow | 创建工作流 |
| GET | `/api/v1/journal/workflows/:id` | GetWorkflow | 获取工作流详情（含 flow_json） |
| PUT | `/api/v1/journal/workflows/:id` | UpdateWorkflow | 更新（名称、描述、flow_json） |
| DELETE | `/api/v1/journal/workflows/:id` | DeleteWorkflow | 删除 |
| POST | `/api/v1/journal/workflows/:id/execute` | ExecuteWorkflow | 执行工作流 |
| GET | `/api/v1/journal/workflows/:id/runs` | ListRuns | 运行历史列表 |
| GET | `/api/v1/journal/workflows/:id/runs/:runId` | GetRun | 运行详情（含 node_statuses） |
| WS | `/api/v1/ws/workflow-runs/:runId` | WSHandler | 实时节点状态 |

## 前端结构

```
ui/
├── app/workflow/
│   └── page.tsx                      # 双 tab 页面：[编排设计] [运行历史]
├── components/workflow/
│   ├── workflow-canvas.tsx           # ReactFlow 画布主组件（DndContext + ReactFlow）
│   ├── workflow-sidebar.tsx          # 左侧节点拖拽面板（4 种节点）
│   ├── workflow-config-panel.tsx     # 右侧节点属性编辑面板
│   ├── nodes/
│   │   ├── input-node.tsx            # 输入节点
│   │   ├── llm-node.tsx              # LLM 节点（模型选择 + system_prompt + reasoning_mode）
│   │   ├── condition-node.tsx        # 条件节点（多规则 + 默认分支）
│   │   └── output-node.tsx           # 输出节点（join_mode 选择）
│   └── run-detail-panel.tsx          # 运行详情（节点状态时间线 + 输出结果）
├── lib/api/workflow.ts               # API 封装
├── hooks/
│   └── use-workflow-ws.ts            # WebSocket hook（连接 + 自动重连 + 消息处理）
└── types/workflow.ts                 # 类型定义
```

### 技术要点

- ReactFlow 用 `next/dynamic` 导入，`ssr: false`
- 拖拽用 `@dnd-kit/core`（`useDraggable` + `useDroppable`），放下时 `screenToFlowPosition` 定位
- 节点属性编辑：选中节点 → 右侧面板显示表单 → 修改 `node.data`
- 连线验证（`isValidConnection`）：禁止自环、禁止 condition 单源连接、禁止同一节点重复连线
- 保存前全量验证：条件节点需 2+ 输出边、LLM 节点需指定模型、无孤立节点
- WebSocket 断开自动重连（exponential backoff）

### 页面结构

```
┌──────────────────────────────────────────────────────┐
│  [编排设计]  [运行历史]                                 │
├──────────────────┬──────────────────┬────────────────┤
│  左侧节点面板     │  中间画布         │  右侧属性面板   │
│  (workflow-      │  (workflow-      │  (workflow-    │
│   sidebar)       │   canvas)        │   config-panel) │
│                  │                  │                │
│  📥 输入节点      │  ┌──┐   ┌──┐    │  模型: [select]│
│  🧠 LLM 节点     │  │in│──▶│ll│    │  Prompt: [...] │
│  🔀 条件节点      │  └──┘   └──┘    │  思考: [select] │
│  📤 输出节点      │           │     │                │
│                  │           ▼     │                │
│                  │  ┌──┐   ┌──┐    │                │
│                  │  │ll│◀──│co│    │                │
│                  │  └──┘   └──┘    │                │
│                  │           │     │                │
│                  │           ▼     │                │
│                  │         ┌──┐    │                │
│                  │         │ou│    │                │
│                  │         └──┘    │                │
├──────────────────┴──────────────────┴────────────────┤
│  [保存] [发布] [▶ 运行]                               │
└──────────────────────────────────────────────────────┘
```

## 权限配置

| code | name | resource 类型 | 说明 |
|------|------|--------------|------|
| `menu:workflow` | 工作流菜单 | menu | 侧边栏菜单可见 |
| `workflow.read` | 查看工作流 | button | 列表/详情查看 |
| `workflow.write` | 创建/编辑工作流 | button | 创建、编辑、执行 |
| `workflow.delete` | 删除工作流 | button | 删除操作 |

权限资源绑定：

```
menu:workflow       →  resource_id: "workflow",             resource_type: "menu"
menu:workflow       →  resource_id: "/api/v1/workflow/*",    resource_type: "api_path"
workflow.read       →  resource_id: "GET /api/v1/workflow",  resource_type: "api_path"
workflow.read       →  resource_id: "workflow.tab.designer", resource_type: "button"
workflow.read       →  resource_id: "workflow.tab.runs",     resource_type: "button"
workflow.write      →  resource_id: "POST /api/v1/workflow", resource_type: "api_path"
workflow.write      →  resource_id: "PUT /api/v1/workflow",  resource_type: "api_path"
workflow.delete     →  resource_id: "DELETE /api/v1/workflow", resource_type: "api_path"
```

## 新增依赖

| 层 | 包 | 用途 |
|----|----|------|
| 后端 | `github.com/cloudwego/eino` | 图编排（compose.Graph） |
| 后端 | `github.com/Knetic/govaluate` | 表达式 DSL 求值 |
| 后端 | `github.com/gorilla/websocket` | WebSocket |
| 前端 | `@xyflow/react` v12 | ReactFlow 画布 |
| 前端 | `@dnd-kit/core` + `@dnd-kit/utilities` | 节点拖拽 |

## 实施步骤

| # | 步骤 | 内容 | 范围 |
|---|------|------|------|
| 1 | Migration + Domain 骨架 | 建 `workflows`/`workflow_runs` 表；DO + Repo + Module 注册 + Container 装配 + Route 占位 | 后端 |
| 2 | 流程引擎核心 | `engine.go`（JSON→Eino Graph）；`condition_eval.go`（govaluate 表达式）；OR-join channel 竞争 | 后端 |
| 3 | Handler + WebSocket | CRUD 8 端点 + Execute；`ws_hub.go` + `ws_handler.go`（实时广播节点状态） | 后端 |
| 4 | 权限种子 + is_system 修复 | migration 补充 workflow 权限；新建权限页 `is_system` 开关加超管守卫 | 前后端 |
| 5 | 前端类型 + API + WS Hook | `types/workflow.ts`；`lib/api/workflow.ts`；`hooks/use-workflow-ws.ts` | 前端 |
| 6 | 编排设计页面 | ReactFlow 画布 + 节点面板 + 4 种节点组件 + 属性编辑面板 + 保存/加载/运行 | 前端 |
| 7 | 运行历史 + 实时状态渲染 | 运行列表页 + 运行详情时间线 + WebSocket 实时节点颜色更新 | 前端 |
