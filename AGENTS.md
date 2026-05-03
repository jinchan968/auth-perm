# [auth-perm] Agent Guide

`AGENTS.md` 只保留团队共享的最小规则、任务入口与文档索引；长说明统一下沉到 `docs/`。

## 快速入口
- [文档总览](./docs/README.md) — 文档地图、维护约定、专题索引
- [架构总览](./docs/ARCHITECTURE.md) — 分层边界、数据流、改动落点
- [开发指南](./docs/DEVELOPMENT.md) — 构建、测试、lint、联调流程
- [代理协作规范](./docs/AGENT_WORKFLOWS.md) — 参考 Claude folder / harness agent 的任务拆分与交接方式
- [外部参考资料](./docs/REFERENCES.md) — Claude folder 与 harness agent 相关外部链接索引
- [权限系统实现说明](./docs/PERMISSION_SYSTEM_IMPLEMENTATION.md)
- [用户管理实现说明](./docs/USER_MANAGEMENT_IMPLEMENTATION.md)
- [用户管理快速开始](./docs/USER_MANAGEMENT_QUICKSTART.md)

## 项目速览
- 后端：Go + Gin + GORM + Redis，入口 `./cmd/api/main.go`
- 依赖装配：`./internal/container/`
- HTTP 路由：`./internal/controller/http/route.go`
- 领域模块：`./internal/domain/{auth,permission,tenant,todo}`
- 前端：Next.js App Router，主目录为 `./ui/`
- 前端权限控制：`./ui/hooks/use-permissions.ts` + `./ui/components/ui/perm-guard.tsx`

## 必守规则
1. 请使用中文交流。
2. 在脚本、文档、说明中统一使用 `./相对路径` 表示项目位置，即当前仓库根目录下的相对路径；避免写绝对路径，本地偏好见 `AGENTS.local.md`。
3. 修改代码前，先定位所属层级与相邻文件；不要只改一个点而忽略依赖注入、路由注册、类型定义、测试或前后端契约同步。
4. 不要把业务逻辑塞进路由、Handler、页面组件或临时脚本；按既有分层落位。
5. 后端新增能力通常至少检查：领域模块注册、`container` 装配、路由注册、DTO/VO、测试。
6. 前端新增能力通常至少检查：`ui/components/ui`、`ui/components`、`ui/lib/api`、`ui/hooks`、`ui/types`、页面/组件、权限显隐；新增页面或交互前优先复用已有共享组件，无合适抽象再新增。
7. 认证、权限、租户相关接口一旦变化，必须同步检查后端接口、前端调用、权限资源标识和相关文档。
8. 使用结构化日志思维，禁止新增 `console.log` / `print()` / `fmt.Println()` 这类调试式输出。
9. 单文件尽量不超过 2000 行；命名遵循 PascalCase（类型）、camelCase（函数）、kebab-case（文件名）。
10. `ui/types/` 仅放纯类型定义，避免引入运行时依赖。
11. 前端通用交互组件（错误提示、确认弹窗、Toast 通知、加载/空状态等）须封装到 `./ui/components/ui/` 下统一复用；全局样式（配色、动画、间距、圆角等）统一走 Tailwind 主题配置与 `./ui/components/ui/` 现有组件体系，禁止在页面或业务组件中硬编码颜色值或独立定义动画。
12. 后端领域常量（枚举值、资源类型、状态码等）统一放到对应领域的 `./internal/domain/<module>/constant/` 目录下；禁止在 `dto`、`repo`、`service`、`handler` 中硬编码魔法字符串，引用方通过 `import constant` 包使用。
13. 通过依赖注入装配的组件（`*CacheService`、`*PermissionResourceRepo`、`*SessionRepo` 等由 `container` 提供的实例）必然非 nil，禁止在业务代码中对其判空。
14. Commit message 使用英文 Conventional Commits 格式。

## 任务路由
- **后端任务**：先读 `docs/ARCHITECTURE.md`，重点看 `./internal/container/`、`./internal/controller/http/`、对应领域模块。
- **前端任务**：先读 `docs/ARCHITECTURE.md` 与 `docs/DEVELOPMENT.md`，重点看 `./ui/components/ui/`、`./ui/components/`、`./ui/app/`、`./ui/hooks/`、`./ui/lib/api/`；优先复用已有组件再考虑新增抽象。
- **全栈任务**：先明确 API 契约，再按“后端接口 → 前端 API 封装 → Hook/Store → 页面/组件 → 文档/测试”顺序推进。
- **文档/规范任务**：优先更新 `AGENTS.md` 的索引，再把长内容写入 `docs/` 下对应专题。

## 已验证命令
```bash
make build
make test
make lint
make vet
cd ui && pnpm lint
cd ui && pnpm type-check
```

> 说明：初始草案里提到 `make lint-arch`，但当前仓库 `Makefile` 中未看到该目标；使用前请先确认本地是否已补充。

## 文档维护约定
- `AGENTS.md` 只放高频、稳定、必须遵守的共享规则，控制在短文本范围内。
- 详细规范放到 `docs/`；新增长文档时，记得回到本文件补一条索引。
- 一次性问题排查文档可放 `docs/` 根目录；长期有效的规范优先写入 `ARCHITECTURE`、`DEVELOPMENT`、`AGENT_WORKFLOWS` 或 `docs/README.md`。
