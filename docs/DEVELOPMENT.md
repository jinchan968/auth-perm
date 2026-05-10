# 开发指南

本文档汇总当前仓库已验证的构建 / 测试 / lint 命令，以及后端、前端、全栈任务的推荐开发流程

## 1. 已验证命令

### 1.1 后端（仓库根目录）
```bash
make build
make test
make lint
make vet
```

### 1.2 前端（`./ui`）
```bash
cd ui && pnpm lint
cd ui && pnpm type-check
```

### 1.3 补充说明
- 当前 `Makefile` 中已经确认存在：`build`、`test`、`lint`、`vet`、`test-unit`、`test-integration`、`benchmark` 等目标。
- 初始草案中提到的 `make lint-arch` 目前未在仓库 `Makefile` 中看到；若本地另有补充，请在使用前确认。
- `README.md` 中仍保留了部分早期命令和路径描述，和当前仓库状态不完全一致时，以本文件和 `Makefile` / `ui/package.json` 为准。

## 2. 环境与目录认知
开始改动前，建议先确认你在改哪一层：

- 后端 API 入口：`./cmd/api/main.go`
- 后端 Worker 入口：`./cmd/worker/main.go`
- 后端依赖装配：`./internal/container/`（`BuildAPIContainer` / `BuildWorkerContainer`）
- 后端路由注册：`./internal/controller/http/route.go`
- 领域模块：`./internal/domain/{auth,permission,tenant,todo,journal,newshock,cache}`
- 前端页面：`./ui/app/`
- 前端 API：`./ui/lib/api/`
- 前端权限：`./ui/hooks/use-permissions.ts`、`./ui/components/ui/perm-guard.tsx`

## 3. 推荐开发流程

### 3.1 后端任务
1. 先读 `ARCHITECTURE.md`，确认应该落在哪个领域模块。
2. 顺着“路由 → handler → service → repo → dto/validator”链路阅读相邻文件。
3. 若新增能力，检查：
   - 领域 `module.go` 是否已注册；
   - `container` 是否需要新增依赖装配；
   - `route.go` 是否已注册接口；
   - DTO/VO/错误处理是否补齐；
   - 测试是否覆盖关键路径。
4. 开发完成后，至少运行与改动匹配的构建 / 测试 / vet / lint。

### 3.2 前端任务
1. 先确认页面、组件、hook、API 封装、类型定义各自的职责边界。
2. 新增页面或交互前，先检查 `ui/components/ui/` 与 `ui/components/` 是否已有可复用组件；优先复用或组合现有组件，无合适抽象再新增。
3. 优先在 `ui/lib/api/` 增加或修改接口封装，而不是在页面里直接发请求。
4. 若有共享交互或权限逻辑，优先抽到 `ui/hooks/`、`ui/store/` 或业务组件层。
5. 若接口字段变化，同步修改 `ui/types/`。
6. 开发完成后，至少运行 `pnpm lint` 和 `pnpm type-check`。

### 3.3 全栈任务
按下面顺序推进，最不容易漏项：

1. 明确 API 契约。
2. 修改后端接口与 DTO / VO。
3. 更新前端 API 封装。
4. 更新 hook / store / 页面 / 组件。
5. 检查权限资源标识、租户上下文、鉴权流程是否同步。
6. 回写文档与测试。

## 4. 提交前检查清单

### 4.1 通用
- [ ] 没有新增 `console.log` / `print()` / `fmt.Println()` 之类调试输出。
- [ ] 修改点没有绕开既有分层，业务逻辑仍在合适的 service / hook / 组件层。
- [ ] 若新增文件，命名符合 PascalCase / camelCase / kebab-case 约定。

### 4.2 后端
- [ ] 依赖注入是否完整：`module.go` / `container.go`。
- [ ] 路由是否注册：`./internal/controller/http/route.go`。
- [ ] 中间件、权限、租户约束是否被正确复用。
- [ ] 目标模块测试或回归验证已运行。

### 4.3 前端
- [ ] `ui/lib/api`、`ui/hooks`、`ui/types`、页面 / 组件是否同步。
- [ ] 已先检查 `ui/components/ui/` 与 `ui/components/`，优先复用已有共享组件；新增组件有明确复用边界。
- [ ] 权限显隐是否同步更新到 `use-permissions` / `PermGuard` 的使用点。
- [ ] `pnpm lint` 和 `pnpm type-check` 已通过。

### 4.4 认证 / 权限 / 租户变更
- [ ] 后端 API 变动已同步到前端调用。
- [ ] 资源标识、菜单 ID、按钮 ID 命名保持一致。
- [ ] 相关专题文档已更新。

## 5. 测试建议

### 5.1 后端
- 单元测试优先放在目标包附近，或 `tests/unit/`。
- 需要跨模块验证时，使用 `tests/integration/`。
- 关注性能 / 竞态时，可查看 `tests/performance/` 与 `Makefile` 中的 `benchmark`、`race` 目标。

### 5.2 前端
- 当前仓库上下文中已确认的是 lint 与 type-check；如果后续补齐组件测试 / E2E，请把命令同步写回本文件和 `AGENTS.md`。

## 6. 常见易漏项
1. 只改了 handler，没有同步 `module.go` 或 `container.go`。
2. 后端接口字段改了，但 `ui/lib/api/*` 和 `ui/types/*` 没同步。
3. 权限接口变更后，漏改前端 `PermGuard` 使用的资源 ID。
4. 文档还写着旧路径，比如 `configs/`，但当前仓库实际是 `config/`。
5. 只跑了单边验证，没有覆盖全栈链路。

## 7. 什么时候必须补文档
以下情况应同步更新文档：
- 新增或调整代理规则：更新 `AGENTS.md` 和 / 或 `AGENT_WORKFLOWS.md`。
- 调整分层边界或目录职责：更新 `ARCHITECTURE.md`。
- 改变构建、测试、启动或联调命令：更新本文件。
- 变更用户、权限、租户专题的流程或契约：更新对应专题文档。

