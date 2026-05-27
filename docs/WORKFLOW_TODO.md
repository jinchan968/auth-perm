# AI 工作流编排 — TODO

> 基于 code review 产出的待办事项，按优先级排列。

---

## P0 — 功能缺口

### 1. `collectPredecessorResults` 空实现

**现状：** `engine.go` 始终返回 `nil`，merge/output 节点无法收集多路输入。

**影响：** Merge 节点产出空字符串；Output 节点的 or-join 始终报 "all branches failed"；多分支并行工作流的汇聚功能完全不可用。

**方案：** 需改造引擎架构，在 run 级别维护一个 `map[nodeID]*NodeOutput` 缓存。每个节点执行完成后写入缓存，merge/output 节点在闭包中通过前驱列表从缓存中读取结果。Eino 的 `compose.Graph` 不直接暴露前驱上下文，需要在 `WithState` 或自定义 `InvokeState` 中传递缓存。

---

### 2. 前端拖拽放置未实现

**现状：** `workflow-canvas.tsx` 的 `onDragEnd` 是空 stub，侧边栏节点拖到画布无效果。

**方案：** 在 `onDragEnd` 中根据 `event.over` 判断放置目标，用 `event.active.data.current` 获取节点类型，生成新节点 ID，计算放置坐标后调用 `setNodes` 追加。

---

### 3. Toolbar 按钮未接 API

**现状：** `workflow-toolbar.tsx` 的 Save / Validate / Run 按钮只弹 toast，未调用后端接口。

**方案：**
- Save → `updateWorkflow` 或 `createWorkflow`
- Validate → `validateWorkflow`
- Run → `executeWorkflow`（先 async，接 WS 监听）
- Reset → 清空画布前加确认对话框

---

## P1 — 安全与健壮性

### 4. Repo 层缺少租户隔离

**现状：** `workflow_repo.GetByID` / `Delete` / `Update` 不带 `tenant_id` 过滤，依赖 service 层事后校验。存在 TOCTOU 竞态窗口。

**方案：** Repo 的 `GetByID`、`Delete`、`Update` 方法增加 `tenantID` 参数，SQL 查询加 `WHERE tenant_id = ?`。Service 层的 post-fetch 校验可保留作为防御性编程，但不应是唯一屏障。

---

### 5. WebSocket 连接安全性

**现状：** Origin 校验仅阻止空 origin 和不匹配 host，生产环境过于宽松。

**方案：** 使用配置的允许 origin 白名单；或在 WS 握手时校验 JWT token（已通过 query param 传递，但未验证签名）。

---

### 6. Hub 无优雅关闭

**现状：** `WSHub.Run()` 是死循环 goroutine，无 `context`、无 `done` channel，进程退出时无法优雅关闭。

**方案：** `NewWSHub` 接受 `context.Context`，`Run` 中 select `<-ctx.Done()` 退出循环，关闭所有 client 连接。

---

## P2 — 代码质量

### 7. 重复的租户鉴权模式

**现状：** 每个 service 方法都重复 "按 ID 查 → 比对 tenantID → 返回 not found" 的模式（共 11 处）。

**方案：** 抽取 `authorizeTenant(do TenantID, requestedTenantID) error` 辅助函数，或在 repo 层直接过滤。

---

### 8. Async 执行 goroutine 中 runDO 并发写入

**现状：** `ExecuteWorkflowAsync` 中 `runDO` 被 goroutine 闭包捕获并修改，虽然当前只有一个 goroutine 写入，但模式脆弱。

**方案：** 在 goroutine 内部创建 `WorkflowRunDO` 的局部副本，或用 `sync.Mutex` 保护写入。

---

## P3 — 前端体验

### 9. WorkflowRuns 组件缺轮询

**现状：** 运行列表只在切换工作流时加载一次，正在执行的 run 不会自动刷新。

**方案：** 增加 `setInterval` 轮询，或在 WS 消息 `run_end` 时触发 `fetchRuns`。

---

### 10. ConfigPanel 缺少更多节点配置

**现状：** Condition 节点配置区是空桩（"敬请期待"）；Transform 缺少 `params` 配置（如 regex 的 pattern）；Merge 缺少 delimiter 配置。

**方案：** 逐个补全：Condition 需要规则构建器 UI；Transform 根据 operation 动态渲染参数表单；Merge 在 strategy=join 时显示 delimiter 输入框。

---

### 11. 节点缺少错误状态展示

**现状：** 节点有 `error` 边框颜色，但没有展示具体错误信息的 tooltip 或面板。

**方案：** 在节点组件中，当 `status === 'error'` 时显示错误摘要（可 hover 展开详情），或在 ConfigPanel 中展示选中节点的错误信息。

---

## 已修复

| 日期 | 修复内容 |
|------|----------|
| 2026-05-27 | `ExecuteWorkflowSync` 添加 10 分钟超时 |
| 2026-05-27 | `json.Marshal` 错误不再静默丢弃（engine、service、hub 三处） |
| 2026-05-27 | `UpdateWorkflow` 改为指针类型，支持清空字段 |
| 2026-05-27 | Hub Broadcast/Run 添加日志（json 错误 + 通道满） |
| 2026-05-27 | WebSocket hook：intentionalCloseRef 区分主动/异常断开，ref 稳定回调 |
| 2026-05-27 | workflow-runs：useCallback 包裹 fetchRuns，修复 useEffect 依赖 |
| 2026-05-27 | ConfigPanel：原生 textarea 替换为 Textarea 组件 |
| 2026-05-27 | 6 个节点组件：statusColor 提取为共享 node-utils.ts |
| 2026-05-27 | 删除 engine.go 中未使用的 `adj` 变量 |
| 2026-05-27 | `UpdateWorkflow` 发布时校验 flowJSON 非空 |
