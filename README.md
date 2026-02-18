# Auth-Perm

基于DDD领域驱动设计的Go身份验证和授权系统，支持多租户、OAuth第三方登录、RBAC权限管理。

## 🚀 特性

- ✅ **DDD架构设计**：领域驱动，清晰的分层架构，胖模型模式
- ✅ **多租户支持**：数据库级租户数据隔离，User全局，Account租户特定
- ✅ **灵活标识系统**：支持手机号或邮箱作为用户唯一标识
- ✅ **RBAC权限管理**：灵活的角色权限控制系统
- ✅ **OAuth集成**：支持GitHub、Google、微信等第三方登录
- ✅ **高性能缓存**：Redis + 本地LRU双层缓存，延迟双删策略
- ✅ **监控报警**：完善的健康检查和性能监控，Redis内存监控
- ✅ **生产就绪**：Docker + K8s + 完整的部署方案
- ✅ **自定义认证**：自定义Token + Session机制（非JWT）
- ✅ **完整审计**：操作审计日志记录
- ✅ **API保护**：中间件级别的路由保护和速率限制
- ✅ **DTO数据传输**：领域服务返回DTO，避免数据泄露
- ✅ **统一路由管理**：路由集中管理，中间件抽离

## 📋 技术栈

- **语言**：Go 1.21+
- **Web框架**：Gin
- **数据库**：PostgreSQL + GORM
- **缓存**：Redis + 本地LRU
- **依赖注入**：Uber Dig
- **定时任务**：robfig/cron
- **认证**：自定义Token + Session
- **日志**：Go标准库 + 文件轮转
- **配置**：YAML + .env
- **数据库迁移**：golang-migrate

## 🏗️ 项目架构

```
auth-perm/
├── cmd/api/main.go                        # 应用入口
├── configs/                               # 配置文件 (app.yaml, .env.sample)
├── internal/
│   ├── container/                         # 依赖注入容器
│   ├── controller/
│   │   ├── http/                          # HTTP处理器和路由
│   │   │   ├── auth_handler.go            # 认证处理器
│   │   │   ├── permission_handler.go      # 权限处理器
│   │   │   └── route.go                   # 路由注册
│   │   └── middleware/                    # HTTP中间件
│   │       ├── cors.go                    # CORS中间件
│   │       ├── logging.go                 # 日志中间件
│   │       ├── ratelimit.go               # 限流中间件
│   │       └── recovery.go                # 恢复中间件
│   ├── domain/
│   │   ├── auth/                          # 认证领域
│   │   │   ├── dm/                        # 领域模型 (Domain Models)
│   │   │   ├── repository/                # 仓储接口
│   │   │   │   └── impl/                  # 仓储实现 (GORM)
│   │   │   ├── service/                   # 领域服务
│   │   │   └── dto/                       # 数据传输对象
│   │   └── permission/                    # 权限领域
│   │       ├── dm/                        # 领域模型
│   │       ├── repository/                # 仓储接口
│   │       │   └── impl/                  # 仓储实现
│   │       ├── service/                   # 领域服务
│   │       └── dto/                       # 数据传输对象
│   ├── infra/
│   │   └── cache/                         # 基础设置 - 缓存 (Redis + LRU)
│   ├── common/                            # 通用代码
│   │   ├── errors/                        # 自定义错误
│   │   ├── monitoring/                    # 监控组件
│   │   └── dto/                           # 通用DTO
│   └── middleware/                        # 业务中间件
│       └── auth.go                        # 认证中间件
├── pkg/                                   # 共享库
├── migrations/                            # 数据库迁移文件
├── docs/                                  # 项目文档
│   ├── architecture/                      # 架构文档
│   ├── database/                          # 数据库设计
│   └── design/                            # 其他设计文档
├── scripts/                               # 构建和部署脚本
└── tests/                                 # 测试 (单元, 集成, E2E)
```

## 🛠️ 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 12+
- Redis 6+
- Docker (可选)

### 本地开发

1. **克隆项目**
```bash
git clone <repo-url>
cd auth-perm
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境变量**
```bash
cp config/.env.sample .env
# 编辑 .env 文件，配置数据库和Redis连接信息
```

4. **数据库迁移**
```bash
# 安装golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 运行所有迁移（按相关性分组）
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" up

# 回滚最后一个迁移版本
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" down 1

# 查看迁移状态
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" version

# 创建新的迁移文件
migrate create -ext sql -dir migrations -seq create_new_feature_table
```

> **golang-migrate 简介**: 一个流行的Go数据库迁移工具，支持多种数据库，提供版本控制和迁移管理功能，在Go社区中被广泛使用。
>
> **迁移文件组织策略**:
> - `000001_init_schema.sql`: 完整的数据库结构初始化（包含所有表、索引、触发器）
>   - 账户相关表：accounts、users、organizations、user_org
>   - 权限相关表：roles、permissions、user_roles、role_permissions
>   - 会话管理表：sessions
>   - 所有必要的索引和触发器
>
> **设计说明**:
> - UUID生成：使用Go代码生成（github.com/google/uuid），不依赖数据库扩展
> - 时间戳管理：GORM自动处理CreatedAt和UpdatedAt字段，无需额外标签或数据库触发器
> - 租户隔离：通过tenant_id字段实现数据库级多租户隔离
> - 软删除：使用deleted_at字段支持软删除和数据恢复
>
> **性能优化**:
> - 零数据库触发器开销，完全使用GORM内置时间戳管理
> - 简化的模型定义，减少冗余标签和代码
> - 利用GORM自动优化，提升数据库操作效率
> - 更少的数据库依赖，提高系统可移植性

5. **启动服务**
```bash
go run cmd/api/main.go
```

### Docker部署

```bash
# 构建镜像
docker build -t auth-perm .

# 使用docker-compose
docker-compose up -d
```

## 📚 API文档

### 认证相关

#### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

#### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123"
}
```

#### OAuth登录
```http
GET /api/v1/auth/oauth/{provider}?redirect_uri=xxx
```

### 用户管理

#### 获取用户信息
```http
GET /api/v1/users/profile
Authorization: Bearer <token>
```

#### 更新用户信息
```http
PUT /api/v1/users/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "nickname": "New Nickname",
  "avatar": "https://example.com/avatar.jpg"
}
```

### 权限管理

#### 创建角色
```http
POST /api/v1/roles
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "管理员",
  "code": "admin",
  "description": "系统管理员角色"
}
```

#### 分配权限
```http
POST /api/v1/roles/{roleId}/permissions
Authorization: Bearer <token>
Content-Type: application/json

{
  "permission_ids": ["perm-1", "perm-2"]
}
```

## 🔧 配置说明

### 应用配置 (configs/app.yaml)

```yaml
server:
  host: localhost
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  user: auth_perm
  password: your_password
  dbname: auth_perm
  sslmode: disable

redis:
  host: localhost
  port: 6379
  db: 0

token:
  secret: your-secret-key-here
  expires_in: 24h
  session_secret: your-session-secret-here
```

### 环境变量 (.env)

```env
# 数据库配置
DB_HOST=localhost
DB_USER=auth_perm
DB_PASSWORD=your_password
DB_NAME=auth_perm

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379

# OAuth配置
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 运行特定包的测试
go test ./internal/auth/service
```

## 📊 监控

### 健康检查
```http
GET /health
```

响应：
```json
{
  "status": "healthy",
  "time": {
    "timestamp": "2024-01-20T12:00:00Z"
  }
}
```

### 系统指标
```http
GET /metrics
```

响应：
```json
{
  "uptime_seconds": 3600,
  "requests_total": 1000,
  "errors_total": 10,
  "db_connections": 5,
  "redis_memory_usage": "128MB"
}
```

## 🏗️ 架构设计详解

### 分层架构
```
┌─────────────────────────────────────┐
│           Controller Layer          │  ← HTTPHandlers, Routes, Middleware
├─────────────────────────────────────┤
│           Domain Layer              │  ← Services (返回DTO)
├─────────────────────────────────────┤
│        Repository Layer             │  ← Interfaces
├─────────────────────────────────────┤
│       Infrastructure Layer          │  ← GORM Implementations
└─────────────────────────────────────┘
```

### 关键设计原则

1. **DTO数据传输**
   - Domain Service 返回 DTO 而不是 Domain Model
   - 避免内部实现泄露
   - API 层独立于领域模型

2. **统一路由管理**
   - 路由注册集中在 `internal/controller/http/route.go`
   - Handler 专注于业务逻辑
   - 中件间抽离到独立模块

3. **灵活标识系统**
   - User 表支持手机号或邮箱作为标识
   - 复合唯一约束：(identifier_type, identifier_value)
   - 支持标识类型切换

4. **多租户设计**
   - User: 全局用户信息（无 tenant_id）
   - Account: 租户特定账户（有 tenant_id）
   - 一个 User 可以对应多个 Account

## 🔒 安全考虑

1. **密码安全**：使用bcrypt进行密码哈希
2. **令牌管理**：自定义Token + Session机制
3. **输入验证**：严格的参数验证和过滤
4. **CORS保护**：配置合适的CORS策略
5. **速率限制**：防止API滥用
6. **审计日志**：记录所有重要操作

## 📈 性能优化

1. **双层缓存**：Redis + 本地LRU缓存
2. **连接池**：数据库和Redis连接池优化
3. **延迟双删**：缓存一致性策略
4. **索引优化**：数据库查询优化
5. **并发控制**：合理的锁策略

## 🚀 部署

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o auth-perm cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/auth-perm .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./auth-perm"]
```

### Kubernetes部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-perm
spec:
  replicas: 3
  selector:
    matchLabels:
      app: auth-perm
  template:
    metadata:
      labels:
        app: auth-perm
    spec:
      containers:
      - name: auth-perm
        image: auth-perm:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: REDIS_HOST
          value: "redis-service"
```

## 📖 文档

### 🏛️ 架构与设计
- [AI编码实践指南](doc/design/ai-coding-practices.md) - AI辅助编码的最佳实践和规范
- [测试标准与规范](doc/design/testing-standards.md) - 单元测试、集成测试的标准和指南
- [仓储重构文档](doc/architecture/repository-refactoring.md) - 仓储层按聚合根重构的设计说明

### 🗄️ 数据库设计
- [实体关系设计](doc/database/entity-relationships.md) - 数据库实体关系图(ERD)和设计说明
- [表结构设计](doc/database/schema-design.md) - 详细的数据库表结构定义和索引策略
- [数据库迁移指南](doc/database/migration-guide.md) - golang-migrate工具使用和迁移最佳实践

### 📚 API文档
- [认证API](#api文档) - 用户注册、登录、OAuth相关接口
- [权限管理API](#api文档) - 角色、权限管理接口
- [监控接口](#api文档) - 健康检查、系统指标接口

### 🚀 部署与运维
- [快速开始](#快速开始) - 本地开发环境搭建
- [Docker部署](#docker部署) - 容器化部署方案
- [Kubernetes部署](#kubernetes部署) - K8s集群部署配置
- [配置说明](#配置说明) - 应用配置和环境变量
- [监控与告警](#监控) - 系统监控和性能指标

### 🧪 开发指南
- [测试指南](#测试) - 测试命令和覆盖率报告
- [编码规范](doc/design/ai-coding-practices.md) - 项目编码标准和最佳实践
- [安全考虑](#安全考虑) - 安全相关的设计和实现要点

## 🤝 贡献

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📝 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🆘 支持

如果您遇到问题或有疑问，请：

1. 查看 [文档](doc/)
2. 搜索 [Issues](../../issues)
3. 创建新的 [Issue](../../issues/new)

## 🎯 核心特性详解

### 🔐 认证系统
- **自定义Token机制**: 采用非JWT的Token+Session方案，支持令牌续期和撤销
- **OAuth集成**: 支持GitHub、Google、微信等第三方OAuth登录
- **会话管理**: Redis缓存会话信息，支持多设备登录控制
- **安全防护**: 密码哈希、输入验证、速率限制、CORS保护

### 🛡️ 权限系统
- **RBAC模型**: 基于角色的访问控制，支持细粒度权限管理
- **资源权限**: 支持通配符权限表达式（如：`users.*`, `posts.read`）
- **权限缓存**: 权限信息缓存，提升性能
- **权限继承**: 支持角色继承和权限传递

### ⚡ 性能优化
- **双层缓存**: Redis + 本地LRU缓存，延迟双删策略保证数据一致性
- **连接池**: 数据库和Redis连接池优化
- **并发控制**: 合理的锁策略和并发处理
- **索引优化**: 数据库查询优化，支持分页查询

### 📊 监控与审计
- **健康检查**: 数据库、Redis等依赖服务健康状态监控
- **性能指标**: 请求量、错误率、响应时间等指标收集
- **审计日志**: 完整的操作审计记录，支持追溯
- **告警机制**: 支持钉钉、企业微信、邮件等告警方式

## 🗺️ 路线图

### ✅ 已完成 (v1.0.0)
- [x] DDD架构设计和实现
- [x] 基础认证功能（注册、登录、会话管理）
- [x] RBAC权限系统
- [x] OAuth第三方登录（GitHub、Google、微信）
- [x] 双层缓存系统（Redis + LRU）
- [x] 监控和健康检查接口
- [x] 审计日志系统
- [x] API中间件保护
- [x] 灵活标识系统（手机号/邮箱）
- [x] 多租户设计（User全局，Account租户特定）
- [x] DTO数据传输模式
- [x] 统一路由管理和中间件抽离

### 🚧 进行中
- [ ] 完整的单元测试和集成测试
- [ ] API文档自动生成（Swagger）
- [ ] 性能基准测试

### 📋 计划中
- [ ] v1.1.0 - 增强监控和报警（接入阿里云日志服务）
- [ ] v1.2.0 - 支持更多OAuth提供商
- [ ] v1.3.0 - 图形化管理界面
- [ ] v2.0.0 - 微服务架构改造

### 💡 未来特性
- [ ] 多租户SaaS模式支持
- [ ] LDAP/AD集成
- [ ] 无密码登录
- [ ] API网关集成
- [ ] 实时权限更新

---

**Auth-Perm** - 让身份验证和授权变得简单而强大！ 🚀

> 遵循 [AI编码实践](doc/design/ai-coding-practices.md) 进行开发，确保代码质量和一致性。