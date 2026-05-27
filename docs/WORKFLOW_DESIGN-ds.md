# AI 工作流编排工具 — 设计文档（DS 视角）

## 概述

在 AI 矩阵推理基础上，构建可视化工作流编排工具。与 glm 版的核心差异：**6 节点类型**（新增 transform、merge）、**结构化规则构建器**替代 DSL 文本、**异步执行模式**、**工作流模板库**、**独立节点执行记录表**。

## 与 glm 版的关键差异

| 维度 | glm 版 | ds 版 |
|------|--------|-------|
| 节点数 | 4 种 | 6 种（+transform +merge） |
| 条件规则 | govaluate 文本 DSL | 结构化 JSON 规则构建器 |
| 执行模式 | 同步执行 | 同步 + 异步（fire-and-forget） |
| 数据库表 | 2 张（workflows, workflow_runs） | 3 张（+workflow_run_nodes 逐节点记录） |
| 模板 | 无 | 内置 workflow 模板库 |
| 输入类型 | 自由文本 | 自由文本 + 结构化 JSON Schema |
| OR-join | channel 竞争 | 超时+channel 双重竞争 |

## 节点类型（6 种）

| 节点 | 入边 | 出边 | 配置 | 说明 |
|------|------|------|------|------|
| **trigger** | 无 | 1 | `input_schema?: Record<string, string>` | 工作流入口，支持简单文本或结构化字段 |
| **llm** | 1 | 1 | `model_id`, `system_prompt`, `temperature`, `reasoning_mode` | 复用 OpenCode client |
| **condition** | 1 | N | `rules: Rule[]`, `default_handle` | 结构化规则，按序求值 |
| **transform** | 1 | 1 | `operation: Operation`, `params` | 文本转换：regex 提取/替换、trim、markdown→text |
| **merge** | N | 1 | `strategy: "concat"\|"first"\|"join"` | 合并多上游结果（拼接/取首/用分隔符拼接） |
| **output** | N | 无 | `format: "raw"\|"json"\|"markdown"`, `join_mode: "and"\|"or"` | 汇聚并格式化最终输出 |

## 结构化条件规则

不使用自由文本 DSL，改为结构化 JSON 规则。前端用表单构建，后端直接求值。

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

支持的运算符：

| 运算符 | field 类型 | value 类型 | 说明 |
|--------|-----------|-----------|------|
| `contains` | content（文本） | string | 包含子串 |
| `not_contains` | content | string | 不包含 |
| `equals` | content | string | 完全相等 |
| `matches` | content | string（正则） | 正则匹配 |
| `starts_with` | content | string | 以...开头 |
| `ends_with` | content | string | 以...结尾 |
| `gt` / `gte` / `lt` / `lte` | length（数值） | number | 长度比较 |
| `is_empty` / `not_empty` | content | — | 内容空/非空 |
| `model_name_equals` | model_name | string | 比较来源模型名（merge 场景） |

逻辑组合：`AND` / `OR`，支持嵌套规则组。

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

### condition 分支路由

```json
{
  "branches": [
    {
      "handle": "long_and_clean",
      "rule": {
        "logic": "AND",
        "rules": [
          { "field": "length", "operator": "gte", "value": 100 },
          { "field": "content", "operator": "not_contains", "value": "error" }
        ]
      }
    },
    {
      "handle": "has_error",
      "rule": {
        "logic": "OR",
        "rules": [{ "field": "content", "operator": "contains", "value": "error" }]
      }
    }
  ],
  "default_handle": "fallback"
}
```

按序求值，首个满足的 branch 路由到对应 handle。

### 前端规则构建器

condition 节点属性面板提供表单式构建：
- 下拉选逻辑（AND/OR）
- 每行：字段下拉 + 运算符下拉 + 值输入 + 删除按钮
- 支持添加规则组（嵌套）
- 实时预览生成的等效表达式文字
- 每 branch 一个 handle 名输入框

## 新的节点详解

### transform 节点

对上游 LLM 输出做后处理：

```json
{
  "operation": "regex_replace",
  "params": {
    "pattern": "<[^>]+>",
    "replacement": ""
  }
}
```

| operation | 说明 | params |
|-----------|------|--------|
| `regex_extract` | 提取匹配组 | `pattern`, `group_index?` |
| `regex_replace` | 正则替换 | `pattern`, `replacement` |
| `trim` | 去首尾空白 | — |
| `markdown_to_text` | 去掉 Markdown 标记 | — |
| `extract_json` | 提取文本中的 JSON 块 | — |
| `truncate` | 截断到指定长度 | `max_length` |
| `to_uppercase` / `to_lowercase` | 大小写转换 | — |

### merge 节点

汇聚多个上游 LLM 节点的输出：

| strategy | 行为 |
|----------|------|
| `concat` | 将各输出拼接为一段文本（用 `\n\n---\n\n` 分隔） |
| `first` | 取第一个完成的输出（OR-join，配合 channel 竞争） |
| `join` | 用自定义分隔符拼接 |

merge 节点自身是 AND-join（等待所有上游完成），`first` 策略通过上游抢 channel 实现 OR-join 语义。

## 执行模式

| 模式 | 触发方式 | 行为 |
|------|---------|------|
| **同步** | POST `/execute?mode=sync` | 等待全流程完成，返回最终结果（HTTP 超时 120s） |
| **异步** | POST `/execute?mode=async` | 立即返回 `run_id`，后台执行，前端轮询或 WS 获取状态 |
| **WS 实时** | WebSocket `/ws/run/:runId` | 连接后接收实时节点状态推送，无需轮询 |

异步模式下前端展示"运行中"状态，轮询 `GET /runs/:runId` 获取进度。

## 数据库设计

### workflows 表（同 glm）

```sql
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    flow_json JSONB NOT NULL,
    template_id UUID,                       -- 若从模板创建，引用源模板
    status VARCHAR(16) NOT NULL DEFAULT 'draft',  -- draft / published / archived
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### workflow_runs 表（扩展）

```sql
CREATE TABLE IF NOT EXISTS workflow_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
    account_id VARCHAR(64) NOT NULL,
    execution_mode VARCHAR(8) NOT NULL DEFAULT 'sync',   -- sync / async
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
CREATE INDEX IF NOT EXISTS idx_wf_runs_tenant_status ON workflow_runs(tenant_id, status, started_at DESC);
```

### workflow_run_nodes 表（新增，ds 版特有）

逐节点存储执行明细，支持运行详情时间线展示和调试。

```sql
CREATE TABLE IF NOT EXISTS workflow_run_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    node_id VARCHAR(128) NOT NULL,           -- 画布节点 ID（如 "llm_1"）
    node_type VARCHAR(32) NOT NULL,          -- input / llm / condition / transform / merge / output
    node_label VARCHAR(128),                 -- 用户可见标签
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    input_json JSONB,                        -- 节点输入
    output_json JSONB,                       -- 节点输出
    error TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    retry_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_wf_run_nodes_run ON workflow_run_nodes(run_id);
```

## 工作流模板

示例预置模板：

| 模板名 | 结构 | 用途 |
|--------|------|------|
| **AI 矩阵推理** | trigger → 5×llm → merge → output | 5 模型并行回答，等价 AI 矩阵功能 |
| **质量卫士** | trigger → llm → condition（检查质量）→ output（达标）/ llm（重写） | 输出不合格时自动重写 |
| **多模型对比** | trigger → 3×llm → merge(concat) → output | 多模型并行回答，拼接对比 |
| **两阶段分析** | trigger → llm（分析）→ llm（总结）→ output | 先分析再总结的串行管道 |

模板存储在 `workflows` 表中（`template_id IS NULL` 且 `status = 'template'`）。用户从模板创建时生成新 workflow，`template_id` 指向源。

## 流程引擎（不使用 Eino）

ds 版不使用 Eino，改用**自研轻量引擎**。理由：依赖更少，OR-join 实现更自然，节点扩展更灵活。

### 核心结构

```go
type Engine struct {
    openCode *opencode.Client
}

type Execution struct {
    RunID       string
    inputText   string
    inputJSON   map[string]any
    nodeResults map[string]*NodeOutput      // node_id → result
    mu          sync.RWMutex
    ctx         context.Context
    cancel      context.CancelFunc
    wsHub       *Hub                          // WebSocket 推送
    db          *gorm.DB
}

type NodeOutput struct {
    NodeID    string          `json:"node_id"`
    NodeType  string          `json:"node_type"`
    Content   string          `json:"content"`
    ModelName string          `json:"model_name,omitempty"`
    Metadata  map[string]any  `json:"metadata"`
}

type FlowGraph struct {
    Nodes []FlowNode
    Edges []FlowEdge
}

type FlowNode struct {
    ID       string          `json:"id"`
    Type     string          `json:"type"`
    Position xy              `json:"position"`
    Data     json.RawMessage `json:"data"`
}

type FlowEdge struct {
    ID           string  `json:"id"`
    Source       string  `json:"source"`
    Target       string  `json:"target"`
    SourceHandle *string `json:"sourceHandle"`
    TargetHandle *string `json:"targetHandle"`
}
```

### 执行算法（拓扑排序 + 并行调度）

```
1. 解析 flow_json → FlowGraph
2. 拓扑排序，确定执行层级（同一层级的节点可并行）
3. 按层级执行：
   a. 获取当前层级所有就绪节点（前驱全部完成）
   b. 并发执行当前层级节点
   c. 收集结果，写入 nodeResults + node_statuses
4. OR-join 特殊处理：当 or-output 的任一前驱完成时，取消其余运行中分支
5. 执行完成，汇总 output 节点结果，写入 workflow_runs
```

### OR-join 实现

```go
func (e *Execution) runNodeOrJoin(node FlowNode, predecessors []string) *NodeOutput {
    resultCh := make(chan *NodeOutput, len(predecessors))
    timeoutCh := time.After(120 * time.Second)

    for _, predID := range predecessors {
        if res, ok := e.nodeResults[predID]; ok && res.Error == nil {
            select {
            case resultCh <- res:
            default:
            }
        }
    }

    select {
    case result := <-resultCh:
        return result
    case <-timeoutCh:
        return &NodeOutput{Error: "or-join timeout"}
    case <-e.ctx.Done():
        return &NodeOutput{Error: "cancelled"}
    }
}
```

### 条件求值

```go
func evaluateRule(content string, rule Rule) (bool, error) {
    switch rule.Operator {
    case "contains":
        result := strings.Contains(content, rule.Value.(string))
        if rule.Negate { result = !result }
        return result, nil
    case "gt":
        val, _ := strconv.Atoi(rule.Value.(string))
        result := len([]rune(content)) > val
        if rule.Negate { result = !result }
        return result, nil
    case "matches":
        matched, err := regexp.MatchString(rule.Value.(string), content)
        if err != nil { return false, err }
        if rule.Negate { matched = !matched }
        return matched, nil
    // ... 其余运算符
    }
}

func evaluateRuleGroup(content string, group RuleGroup) (bool, error) {
    results := make([]bool, len(group.Rules))
    for i, rule := range group.Rules {
        r, err := evaluateRule(content, rule)
        if err != nil { return false, err }
        results[i] = r
    }
    return combineLogic(group.Logic, results), nil
}
```

## 图验证层

保存/执行前校验：

| 验证项 | 说明 |
|--------|------|
| 连通性 | trigger 到 output 存在可达路径 |
| 孤立节点 | 无入边且非 trigger 的节点 |
| condition 分支 | condition 节点至少 2 条出边 |
| merge 输入 | merge 节点至少 2 条入边 |
| 模型必填 | llm 节点必须指定 model_id |
| 环路检测 | 无有向环（DFS 检测） |
| output 连通 | output 有入边 |
| trigger 无入边 | trigger 节点入度为 0 |

## 后端模块结构

```
internal/domain/workflow/
├── module.go
├── constant/
│   ├── node_type.go          # NodeType + IsValidNodeType
│   ├── run_status.go         # Status + IsValidStatus
│   ├── operators.go          # ConditionOperator 枚举 + 白名单校验
│   ├── transforms.go         # TransformOperation 枚举
│   └── strategies.go         # MergeStrategy / OutputFormat
├── dm/
│   ├── workflow_do.go
│   ├── workflow_run_do.go
│   └── workflow_run_node_do.go   # 新增
├── repo/
│   ├── workflow_repo.go
│   ├── workflow_run_repo.go
│   ├── workflow_run_node_repo.go # 新增
│   └── workflow_template_repo.go # 新增
├── service/
│   ├── workflow_service.go       # CRUD
│   ├── engine.go                 # 自研图执行引擎
│   ├── graph_validator.go        # 图结构校验
│   ├── rule_evaluator.go         # 结构化规则求值
│   ├── node_executor.go          # 各节点类型执行器
│   └── ws_hub.go
├── handler/
│   ├── workflow_handler.go       # REST
│   ├── template_handler.go       # 模板 CRUD
│   └── ws_handler.go
├── dto/
│   ├── workflow_dto.go
│   └── workflow_run_dto.go
└── vo/
    ├── workflow_vo.go
    ├── condition_vo.go           # 规则校验（递归校验 RuleGroup）
    └── flow_vo.go                # flow_json 结构校验
```

## 后端 API（扩展版）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/workflow` | 工作流列表（含模板过滤 `?type=template`） |
| POST | `/api/v1/workflow` | 创建工作流 |
| GET | `/api/v1/workflow/:id` | 工作流详情 |
| PUT | `/api/v1/workflow/:id` | 更新 |
| DELETE | `/api/v1/workflow/:id` | 删除 |
| POST | `/api/v1/workflow/:id/execute` | 执行（`?mode=sync\|async`） |
| POST | `/api/v1/workflow/:id/validate` | 校验图结构（不执行） |
| GET | `/api/v1/workflow/:id/runs` | 运行历史 |
| GET | `/api/v1/workflow/runs/:runId` | 运行详情 |
| GET | `/api/v1/workflow/runs/:runId/nodes` | 节点级执行明细 |
| POST | `/api/v1/workflow/runs/:runId/cancel` | 取消运行 |
| GET | `/api/v1/workflow/templates` | 模板列表 |
| POST | `/api/v1/workflow/:id/clone` | 克隆工作流 |
| WS | `/api/v1/ws/run/:runId` | WebSocket 实时状态 |

## 前端结构（扩展版）

```
ui/
├── app/workflow/
│   ├── page.tsx                    # 双 tab：[编排设计] [运行历史]
│   ├── [id]/
│   │   ├── page.tsx                # 工作流详情/执行
│   │   └── runs/
│   │       └── [runId]/
│   │           └── page.tsx        # 运行详情（节点时间线）
│   └── templates/
│       └── page.tsx                # 模板库
├── components/workflow/
│   ├── workflow-canvas.tsx         # ReactFlow 主画布
│   ├── workflow-sidebar.tsx        # 节点面板（6 节点 + 模板快捷入口）
│   ├── workflow-config-panel.tsx   # 右侧属性面板（按节点类型渲染不同表单）
│   ├── workflow-toolbar.tsx        # 底部工具栏（保存/校验/运行/撤销/重做）
│   ├── nodes/
│   │   ├── trigger-node.tsx        # 入口节点
│   │   ├── llm-node.tsx            # LLM 节点
│   │   ├── condition-node.tsx      # 条件节点（结构化规则表单）
│   │   ├── transform-node.tsx      # 转换节点
│   │   ├── merge-node.tsx          # 合并节点
│   │   └── output-node.tsx         # 输出节点
│   ├── rules/
│   │   └── rule-builder.tsx        # 结构化规则构建器（递归组件）
│   └── run-detail/
│       ├── run-timeline.tsx        # 运行时间线（节点级）
│       ├── run-status-badge.tsx    # 状态徽章
│       └── node-output-viewer.tsx  # 节点输出查看器
├── lib/api/workflow.ts             # API
├── hooks/
│   ├── use-workflow-ws.ts          # WS hook
│   └── use-workflow-validation.ts  # 图校验 hook
└── types/
    ├── workflow.ts                 # 工作流类型
    ├── nodes.ts                    # 节点联合类型
    └── rules.ts                    # 规则构建器类型
```

## 规则构建器 UI（condition 节点配置面板）

```
┌─ Condition Node 配置 ────────────────────────────┐
│                                                    │
│  Branch 1: "long_clean"                            │
│  ┌─ 规则组 (AND) ───────────────────────────┐      │
│  │  [content] [contains ▼] [       ] [×]    │      │
│  │  [length]  [gte ▼]      [ 100   ] [×]    │      │
│  │  [+ 添加规则] [添加规则组]                  │      │
│  └─────────────────────────────────────────┘      │
│                                                    │
│  Branch 2: "has_error"                             │
│  ┌─ 规则组 (OR) ────────────────────────────┐      │
│  │  [content] [contains ▼] [error  ] [×]    │      │
│  │  [+ 添加规则]                              │      │
│  └─────────────────────────────────────────┘      │
│                                                    │
│  [+ 添加分支]                                      │
│                                                    │
│  默认分支: [fallback        ▼]                     │
│  预览: 当 (content含"error") → has_error           │
│        否则 → fallback                             │
└────────────────────────────────────────────────────┘
```

## 新增依赖

| 层 | 包 | 用途 | 与 glm 版差异 |
|----|----|------|--------------|
| 后端 | — | 自研引擎，无 Eino 依赖 | 替代 Eino |
| 后端 | `github.com/gorilla/websocket` | WebSocket | 相同 |
| 后端 | — | 无线 govaluate，改自研求值 | 替代 govaluate |
| 前端 | `@xyflow/react` v12 | 画布 | 相同 |
| 前端 | `@dnd-kit/core` | 拖拽 | 相同 |

## 实施步骤（7 步）

| # | 步骤 | 内容 | 范围 |
|---|------|------|------|
| 1 | Migration + Domain 骨架 | 建 3 张表（workflows, workflow_runs, workflow_run_nodes）；DO + Repo + Module + Container + Route 占位 | 后端 |
| 2 | 图引擎 + 规则求值 + 校验 | engine.go（拓扑排序+并行调度+OR-join）；rule_evaluator.go（结构化规则递归求值）；graph_validator.go（8 项校验） | 后端 |
| 3 | Handler + WS + 模板 API | CRUD 14 端点；ws_hub + ws_handler；template_handler；cancel handler | 后端 |
| 4 | 权限种子 + is_system 修复 | migration 补充 workflow 权限资源；新建权限页 is_system 加超管守卫 | 前后端 |
| 5 | 前端类型 + API + WS | types/workflow.ts + types/nodes.ts + types/rules.ts；lib/api/workflow.ts；hooks/use-workflow-ws.ts | 前端 |
| 6 | 编排设计页面 | ReactFlow 画布 + 6 节点面板 + 6 节点组件 + 规则构建器 + 属性面板 + 校验 + 保存/加载/运行 | 前端 |
| 7 | 运行历史 + 模板库 + 详情 | 运行列表 + 节点时间线 + WS 实时渲染 + 模板库浏览/使用 + 克隆 | 前端 |
