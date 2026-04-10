# 架构总览

本文档描述当前仓库的后端 / 前端分层边界、关键数据流，以及做改动时通常需要同步检查的落点

## 1. 系统概览
- 后端：Go + Gin + GORM + Redis。
- 前端：Next.js App Router + React Query + Zustand。
- 关键主题：认证、权限、租户、Todo。
- 启动入口：`./cmd/api/main.go`。

## 2. 后端启动链路
当前后端的主流程可以概括为：

1. `./cmd/api/main.go`
   - 加载 `.env` 与 `config/app.yaml`。
   - 构建顶层 `context.Context`。
   - 调用 `container.BuildContainer(cfg)` 完成依赖装配。
   - 通过依赖注入拿到 `*gin.Engine` 和 `Scheduler`，启动 HTTP 服务与后台调度器。
2. `./internal/container/container.go`
   - 负责数据库、Redis、缓存、CodeGenerator 等基础设施装配。
   - 调用各领域模块的 `Register*Domain` 完成 repo / service 注册。
   - 注册 HTTP handler。
   - 创建 Gin Engine，并在其中调用 `controller/http.RegisterRoutes(...)` 完成路由挂载。
3. `./internal/controller/http/route.go`
   - 负责集中注册所有 `/api/v1` 路由。
   - 统一挂载认证与 API 权限中间件。
   - 按主题拆分为 `RegisterAuthRoutes`、`RegisterPermissionRoutes`、`RegisterTenantRoutes`、`RegisterUserRoutes`、`RegisterTodoRoutes`。

### 架构含义
- `main.go` 只负责启动和生命周期控制，不承担业务逻辑。
- `container` 只做依赖装配，不直接承载业务规则。
- `route.go` 只做路由组织和中间件拼装，不写业务判断。
- 业务规则应沉到领域模块的 service / repo / validator / dto 中。

## 3. 后端分层边界

### 3.1 推荐职责划分
- `./internal/controller/http/`
  - HTTP 路由注册、参数绑定、响应封装。
  - 不写复杂业务规则，不直接操作数据库。
- `./internal/controller/middleware/`
  - 认证、限流、日志、CORS、统一错误处理等横切逻辑。
- `./internal/domain/<module>/`
  - 每个业务域的 repo、service、handler、dto、validator 等。
  - 模块对外通过 service / handler 暴露能力。
- `./internal/infra/`
  - 缓存、CodeGenerator 等基础设施能力。
- `./internal/common/`
  - 错误、常量、通用 DTO、工具、监控等跨域共享内容。

### 3.2 当前领域模块
- `auth`：注册、登录、会话、OAuth、密码、设备、安全日志、2FA。
- `permission`：角色、权限项、权限资源、组织。
- `tenant`：租户管理、租户设置、租户状态切换。
- `todo`：分类、待办项、调度器。

### 3.3 新增后端能力的最低检查清单
如果你新增一个后端接口或服务，通常至少要检查这些位置：
1. 领域模块内是否新增 / 调整 repo、service、dto、validator。
2. 对应领域的 `module.go` 是否完成注册。
3. 若涉及跨域装配，`./internal/container/container.go` 是否需要新增适配器或 handler 注册。
4. `./internal/controller/http/route.go` 是否已注册路由。
5. 是否需要新增或更新中间件、权限资源标识、测试。

## 4. 前端分层边界

### 4.1 目录职责
- `./ui/app/`
  - App Router 页面与路由段。
  - 页面负责组装组件、触发 hook，不应堆积复杂业务逻辑。
- `./ui/components/`
  - 通用 UI 组件、业务组件、表单组件。
- `./ui/hooks/`
  - 面向页面和组件的状态逻辑、行为封装。
- `./ui/lib/api/`
  - 与后端 API 的通信封装，是前端访问接口的首选入口。
- `./ui/lib/services/`
  - 客户端基础服务，如 logger、token-storage、error-handler。
- `./ui/store/`
  - Zustand 等全局状态。
- `./ui/types/`
  - 纯类型定义，避免引入运行时依赖。

### 4.2 当前权限链路
当前权限控制已经形成比较明确的前后端闭环：

1. 后端在 `./internal/controller/http/route.go` 中为 `/api/v1` 挂载 `APIPermissionMiddleware(...)`。
2. 认证后用户可访问 `/api/v1/auth/my-resources` 获取自身资源清单。
3. 前端在 `./ui/lib/api/resource.ts` 中请求资源接口。
4. `./ui/hooks/use-permissions.ts` 将资源拆分为菜单、按钮、API 路径集合。
5. `./ui/components/ui/perm-guard.tsx` 使用这些集合做菜单 / 按钮显隐控制。

### 4.3 新增前端能力的最低检查清单
1. 接口调用是否进入 `ui/lib/api/`，而不是散落在页面组件里。
2. 共享状态或权限判断是否抽到 `ui/hooks/` 或 `ui/store/`。
3. 数据结构是否同步到 `ui/types/`。
4. 若涉及权限显隐，是否同步更新资源 ID、`use-permissions` 使用方式、`PermGuard` 或页面条件渲染。
5. 若接口契约变化，是否回看后端 DTO/VO 和前端 API 封装是否一致。

## 5. 常见改动落点

| 任务类型 | 典型落点 |
| --- | --- |
| 登录 / 注册 / 会话 | `./internal/domain/auth/`、`./internal/controller/http/route.go`、`./ui/lib/api/auth.ts`、`./ui/hooks/use-auth.ts`、`./ui/app/login/`、`./ui/app/register/` |
| 用户管理 | `./internal/domain/auth/handler`、`./internal/domain/auth/service`、`./ui/lib/api/user.ts`、`./ui/types/user.ts`、`./ui/components/profile/` 或相关用户页面 |
| 权限 / 角色 / 资源 | `./internal/domain/permission/`、`./internal/controller/http/route.go`、`./ui/lib/api/permission*.ts`、`./ui/hooks/use-permissions.ts`、`./ui/components/ui/perm-guard.tsx`、`./ui/app/permissions/` |
| 租户 | `./internal/domain/tenant/`、`./ui/lib/api/tenant.ts`、`./ui/types/tenant.ts`、`./ui/app/tenants/`、`./ui/components/tenants/` |
| Todo | `./internal/domain/todo/`、`./ui/lib/api/todo.ts`、`./ui/types/todo.ts`、`./ui/app/todos/` |
| 中间件 / 路由保护 | `./internal/controller/middleware/`、`./internal/controller/http/route.go`、`./ui/middleware.ts` |

## 6. 分层约束
1. 不要把业务逻辑写进 `main.go`、`route.go`、页面组件或临时脚本。
2. 领域 service 不应反向依赖 HTTP controller。
3. 前端页面不要直接拼接原始 fetch 逻辑；优先通过 `ui/lib/api/` 和 hook 间接访问。
4. `ui/types/` 中不放运行时代码。
5. 新增权限或租户相关接口时，要同时检查：
   - 后端路由和中间件；
   - 前端 API 封装与页面权限显隐；
   - 资源标识命名是否统一；
   - 文档是否需要同步更新。

## 7. 测试与验证位置
- Go 测试：`./tests/unit/`、`./tests/integration/`、`./tests/performance/`，以及包内 `_test.go`。
- 前端静态检查：`./ui/` 下的 `pnpm lint`、`pnpm type-check`。
- 非 trivial 改动至少要覆盖：构建、lint/type-check、目标模块测试、关键链路回归。

## 8. 文档同步规则
当以下内容变动时，除了代码本身，还要同步更新文档：
- 目录结构或职责边界变化：更新本文件。
- 构建 / 测试 / 启动方式变化：更新 `DEVELOPMENT.md`。
- 代理协作流程变化：更新 `AGENT_WORKFLOWS.md`。
- 权限、用户、租户专题变化：更新相应专题文档。

