# Auth-Perm

基于 DDD 思路构建的 Go 身份认证与权限系统，支持多租户、OAuth、RBAC、权限资源控制，并配套 Next.js 管理端。

## 🚀 特性

- ✅ **DDD 分层**：按启动入口、依赖装配、路由层、领域模块、基础设施分层组织
- ✅ **多租户支持**：用户全局、账户租户化，支持租户管理与状态控制
- ✅ **灵活身份标识**：支持手机号或邮箱作为用户唯一标识
- ✅ **RBAC 权限模型**：角色、权限项、权限资源、组织关系协同工作
- ✅ **OAuth 集成**：支持 GitHub、Google、微信登录
- ✅ **双层缓存**：Redis + 本地缓存
- ✅ **安全能力**：自定义 Token + Session、限流、审计、安全日志、2FA
- ✅ **前后端一体化**：Go API + Next.js App Router 管理界面
- ✅ **统一路由管理**：HTTP 路由集中在同一入口注册

## 📋 技术栈

### 后端
- Go 1.21+
- Gin
- GORM + PostgreSQL
- Redis
- Uber Dig
- Goose（数据库迁移）

### 前端
- Next.js 14 App Router
- React 18
- TanStack Query
- Zustand
- Tailwind CSS / shadcn/ui

## 📚 文档入口

- [团队规则入口](./AGENTS.md)
- [文档总览](./docs/README.md)
- [架构总览](./docs/ARCHITECTURE.md)
- [开发指南](./docs/DEVELOPMENT.md)
- [代理协作规范](./docs/AGENT_WORKFLOWS.md)
- [权限系统实现说明](./docs/PERMISSION_SYSTEM_IMPLEMENTATION.md)
- [用户管理实现说明](./docs/USER_MANAGEMENT_IMPLEMENTATION.md)
- [用户管理快速开始](./docs/USER_MANAGEMENT_QUICKSTART.md)

## 🏗️ 当前项目结构

```text
auth-perm/
├── AGENTS.md                              # 团队共享规则入口
├── AGENTS.local.md                        # 本地偏好与隐私相关补充
├── cmd/api/main.go                        # 应用入口
├── config/                                # app.yaml / .env.sample
├── docs/                                  # 架构、开发、专题文档
├── internal/
│   ├── common/                            # 错误、常量、DTO、工具、监控
│   ├── container/                         # 依赖装配
│   ├── controller/
│   │   ├── http/route.go                  # HTTP 路由注册
│   │   └── middleware/                    # 中间件
│   ├── domain/
│   │   ├── auth/                          # 认证、会话、OAuth、2FA
│   │   ├── permission/                    # 权限、角色、组织、资源
│   │   ├── tenant/                        # 租户管理
│   │   └── todo/                          # Todo 与调度器
│   └── infra/                             # 缓存、CodeGenerator 等
├── migrations/                            # 数据库迁移
├── tests/                                 # 单元、集成、性能测试
└── ui/                                    # Next.js 前端
```

更完整的分层边界、依赖关系与改动落点，请查看 [`./docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md)。

## 🛠️ 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 12+
- Redis 6+
- Node.js / pnpm（如需检查或开发前端）

### 后端本地开发

1. **克隆项目**
```bash
git clone <repo-url>
cd auth-perm
```

2. **安装依赖**
```bash
go mod download
```

3. **准备配置**
```bash
cp config/.env.sample .env
# 按需修改 .env 与 config/app.yaml
```

4. **安装 Goose 并执行迁移**
```bash
make migrate-install

export DB_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable'
make migrate-up
```

5. **启动服务**
```bash
make run
```

### 常用验证命令

```bash
make build
make test
make lint
make vet
cd ui && pnpm lint
cd ui && pnpm type-check
```

> 说明：以上命令与当前 `Makefile`、`ui/package.json`、`AGENTS.md` 中的入口保持一致。

## 🔧 配置说明

### 应用配置
主配置文件：[`./config/app.yaml`](./config/app.yaml)

关键配置包括：
- `server`：服务地址、端口、运行模式
- `database`：PostgreSQL 连接与连接池
- `redis`：Redis 连接参数
- `cache`：缓存类型（`redis` / `memory`）
- `token`：Token / Session 相关配置
- `log`：日志级别与日志文件
- `tenant`：多租户开关与默认租户
- `oauth`：GitHub / Google / WeChat 配置
- `smtp`：邮件发送配置
- `monitoring`：健康检查、指标、告警阈值

### 环境变量模板
环境变量模板文件：[`./config/.env.sample`](./config/.env.sample)

常见变量示例：

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=auth_perm
DB_PASSWORD=your_secure_password_here
DB_NAME=auth_perm
REDIS_HOST=localhost
REDIS_PORT=6379
TOKEN_SECRET=your_very_long_and_complex_secret_key_here_at_least_32_characters
SESSION_SECRET=your_very_long_and_complex_session_secret_here_at_least_32_characters
```

## 📚 API 概览

当前后端路由以 [`./internal/controller/http/route.go`](./internal/controller/http/route.go) 为准，README 只保留高频入口，避免与实现脱节。

### 认证相关
- `POST /api/v1/auth/public/register` — 用户注册
- `POST /api/v1/auth/public/login` — 用户登录
- `POST /api/v1/auth/public/forgot-password` — 忘记密码
- `GET /api/v1/auth/profile` — 获取当前用户资料（需认证）
- `PATCH /api/v1/auth/profile` — 更新当前用户资料（需认证）
- `GET /api/v1/auth/my-resources` — 获取当前用户资源清单（前端权限显隐使用）

### 其他资源分组
- `/api/v1/permissions/...` — 权限、角色、权限资源
- `/api/v1/organizations/...` — 组织管理
- `/api/v1/tenants/...` — 租户管理
- `/api/v1/users/...` — 用户管理
- `/api/v1/todos/...` — Todo 与分类管理

若需精确到具体 handler、方法与鉴权链路，请直接查看：
- [`./internal/controller/http/route.go`](./internal/controller/http/route.go)
- 对应领域下的 `handler/` 与 `service/`
- `docs/` 下的专题文档

## 🧪 测试与质量检查

### 后端
```bash
make test
make test-unit
make test-integration
make vet
make lint
make benchmark
```

### 前端
```bash
cd ui && pnpm lint
cd ui && pnpm type-check
```

## 📊 监控

### 健康检查
```http
GET /health
```

示例响应：

```json
{
  "status": "healthy"
}
```

> 当前代码中已明确注册 `/health`。监控实现中虽然存在指标相关代码，但 README 不再把 `/metrics` 写成已公开接口，因为在当前已读路由中未看到公开注册。

## 🚀 部署说明

当前仓库根目录未看到已提交的 `Dockerfile`、`docker-compose` 或 Kubernetes 清单，因此 README 不再保留旧的示例部署片段，以免与仓库现状不一致。

如果后续补齐部署制品，建议同步更新：
- 本 README
- [`./docs/DEVELOPMENT.md`](./docs/DEVELOPMENT.md)
- [`./AGENTS.md`](./AGENTS.md)

## 🧭 核心设计要点

1. **统一路由管理**
   - 路由集中注册于 `internal/controller/http/route.go`
   - 路由层只负责组织接口与中间件，不承载复杂业务逻辑

2. **依赖装配集中化**
   - `internal/container/` 负责数据库、Redis、缓存、领域模块、handler、Gin Engine 的装配

3. **前后端权限闭环**
   - 后端提供当前用户资源清单
   - 前端通过 `ui/hooks/use-permissions.ts` 与 `ui/components/ui/perm-guard.tsx` 做菜单 / 按钮权限显隐

4. **文档与规则分层**
   - `AGENTS.md` 只保留最小共享规则
   - 长说明统一下沉到 `docs/`

## 📖 代码与文档索引

### 文档
- [文档总览](./docs/README.md)
- [架构总览](./docs/ARCHITECTURE.md)
- [开发指南](./docs/DEVELOPMENT.md)
- [代理协作规范](./docs/AGENT_WORKFLOWS.md)
- [权限系统实现说明](./docs/PERMISSION_SYSTEM_IMPLEMENTATION.md)
- [用户管理实现说明](./docs/USER_MANAGEMENT_IMPLEMENTATION.md)
- [用户管理快速开始](./docs/USER_MANAGEMENT_QUICKSTART.md)
- [实现总结](./docs/IMPLEMENTATION_SUMMARY.md)
- [新建用户按钮错误排查](./docs/FIX_NEW_USER_BUTTON_ERROR.md)

### 关键代码入口
- [应用入口](./cmd/api/main.go)
- [依赖装配](./internal/container/container.go)
- [HTTP 路由注册](./internal/controller/http/route.go)
- [前端权限 Hook](./ui/hooks/use-permissions.ts)
- [前端权限守卫](./ui/components/ui/perm-guard.tsx)

## 🤝 贡献

1. 创建特性分支
2. 完成功能或修复
3. 运行相应的构建、lint、type-check、测试命令
4. 同步更新必要文档
5. 提交合并请求

## 🆘 支持

如果你在本地开发或阅读源码时遇到问题，建议按以下顺序排查：

1. 先看 [文档总览](./docs/README.md)
2. 再看 [架构总览](./docs/ARCHITECTURE.md) 与 [开发指南](./docs/DEVELOPMENT.md)
3. 若是权限、用户、租户问题，查看对应专题文档
4. 若你所在的代码托管平台启用了 Issue / PR 流程，再按团队约定提交问题

## 🗺️ 路线图

### ✅ 已完成
- [x] DDD 架构与领域模块拆分
- [x] 基础认证功能（注册、登录、会话管理）
- [x] RBAC 权限系统
- [x] OAuth 第三方登录（GitHub、Google、微信）
- [x] 双层缓存系统（Redis + LRU）
- [x] 健康检查与监控基础能力
- [x] 审计日志与安全日志
- [x] API 中间件保护
- [x] 多租户设计
- [x] 前端管理界面基础功能

### 🚧 进行中
- [ ] 更完整的单元测试和集成测试
- [ ] 更明确的 API 文档沉淀
- [ ] 更系统的性能基准测试

### 📋 计划中
- [ ] 增强监控与告警集成
- [ ] 支持更多 OAuth 提供商
- [ ] 持续完善管理界面与权限配置体验

---

**Auth-Perm** - 让身份验证、授权与权限资源管理更清晰。
