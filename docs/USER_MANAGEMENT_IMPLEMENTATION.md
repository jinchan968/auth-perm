# 用户管理功能实现说明

## 概述

本次实现了完整的用户管理功能，包括后端 API 和前端页面，支持用户列表查看、创建用户、状态管理和角色分配。

## 后端实现

### 1. 数据层 (Repository)

#### account_repo.go
- **SearchAccountsWithCount**: 新增方法，支持带总数的账户搜索
  - 支持租户过滤
  - 支持关键词搜索（username/email/phone）
  - 支持状态过滤
  - 支持账户类型过滤
  - 返回分页数据和总数

#### user_repo.go
- **FindByIDs**: 新增批量查询用户方法
  - 通过 ID 列表批量加载用户信息
  - 用于性能优化，避免 N+1 查询

### 2. DTO 层

#### account_query_dto.go
- **UserWithAccountDTO**: 新增组合 DTO
  - 包含账户信息（account_id, tenant_id, account_type, account_status等）
  - 包含用户信息（user_id, username, nickname, email, phone等）
  - 用于前后端数据传输

### 3. 服务层 (Service)

#### auth_service.go
新增三个用户管理方法：

- **ListAccountsByTenant**: 列出租户下的所有用户（含账户信息）
  - 调用 `SearchAccountsWithCount` 获取账户列表
  - 批量加载关联的用户信息
  - 组合返回 `UserWithAccountDTO` 列表

- **GetAccountWithUser**: 获取单个账户详情（含用户信息）
  - 查询账户信息
  - 查询关联用户信息
  - 返回组合后的详情

- **UpdateAccountStatus**: 更新账户状态
  - 支持租户归属校验（防越权）
  - 更新账户状态（active/inactive/suspended）
  - 自动清除相关缓存

### 4. 控制器层 (Handler)

#### user_handler.go (新文件)
实现 5 个接口：

1. **ListUsers** - `GET /api/v1/users`
   - 查询参数：tenant_id（必填）, keyword, status, account_type, page, page_size
   - 返回：用户列表（含账户信息）+ 分页信息

2. **GetUser** - `GET /api/v1/users/:id`
   - 查询参数：tenant_id（必填）
   - 返回：用户详细信息（含账户信息）

3. **UpdateUserStatus** - `PATCH /api/v1/users/:id/status`
   - 请求体：tenant_id, status
   - 功能：更新账户状态

4. **CreateUser** - `POST /api/v1/users`
   - 请求体：与注册接口相同（identifier_type, email/phone, username, password等）
   - 功能：管理员创建用户（复用 RegisterService）

5. **GetUserAccounts** - `GET /api/v1/users/:id/accounts`
   - 功能：获取用户在所有租户下的账户（暂未完全实现）

#### user_vo.go (新文件)
- **UserWithAccountResponse**: 用户和账户组合响应 VO
- **UpdateUserStatusRequest**: 更新状态请求 VO

### 5. 路由注册

#### route.go
- 新增 `RegisterUserRoutes` 函数
- 挂载到 `/api/v1/users` 路径
- 使用 `AuthMiddleware` 进行认证

### 6. 依赖注入

#### container.go
- 注册 `UserHandler` 到 DI 容器
- 注入 `AuthService` 和 `RegisterService`

## 前端实现

### 1. 类型定义

#### types/user.ts (新文件)
定义了以下类型：
- `AccountStatus`: 账户状态类型
- `AccountType`: 账户类型
- `UserStatus`: 用户状态类型
- `AccountListItem`: 用户列表项
- `AccountListResponse`: 列表响应
- `CreateUserRequest`: 创建用户请求
- `UpdateUserStatusRequest`: 更新状态请求

### 2. API 客户端

#### lib/api/user.ts (新文件)
实现了以下 API 调用：
- `listUsers`: 获取用户列表
- `getUser`: 获取用户详情
- `updateUserStatus`: 更新用户状态
- `createUser`: 创建用户
- `getUserAccounts`: 获取用户账户列表

### 3. 用户列表页面

#### app/permissions/users/page.tsx (重写)
功能特性：
- ✅ 使用 `useTenant` 进行租户选择（与其他页面保持一致）
- ✅ 用户列表展示（表格形式）
- ✅ 关键词搜索（username/email/phone）
- ✅ 状态筛选和管理（下拉选择器）
- ✅ 分页支持
- ✅ 创建用户对话框
  - 标识符类型选择（email/phone）
  - 表单验证
  - 密码确认
- ✅ 跳转到用户详情页面

### 4. 用户详情页面

#### app/permissions/users/[id]/page.tsx (新文件)
功能特性：
- ✅ 显示用户基本信息
- ✅ 显示账户信息
- ✅ 角色分配功能
  - 从当前租户加载可用角色列表
  - 显示已分配的角色
  - 多选复选框界面
  - 保存角色分配
- ✅ 租户隔离（角色和账户在同一租户下）

## 权限控制

### 当前实现
- 所有用户管理接口使用 `AuthMiddleware` 进行认证
- 通过 `tenant_id` 参数进行租户隔离
- 在 Service 层校验账户的 `tenant_id` 防止越权

### 未来增强
- 可以添加 `AdminPermissionMiddleware` 限制为管理员专用
- 可以在 Permission 系统中添加细粒度的用户管理权限

## API 路由

所有用户管理接口都在 `/api/v1/users` 路径下：

```
GET    /api/v1/users                    - 获取用户列表
POST   /api/v1/users                    - 创建用户
GET    /api/v1/users/:id                - 获取用户详情
PATCH  /api/v1/users/:id/status         - 更新用户状态
GET    /api/v1/users/:id/accounts       - 获取用户账户列表
```

## 使用示例

### 1. 获取用户列表

```bash
curl -X GET "http://localhost:8080/api/v1/users?tenant_id=xxx&keyword=test&page=1&page_size=10" \
  -H "Cookie: auth_token=xxx"
```

### 2. 创建用户

```bash
curl -X POST "http://localhost:8080/api/v1/users" \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=xxx" \
  -d '{
    "identifier_type": "email",
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123",
    "confirm_password": "password123",
    "tenant_id": "xxx"
  }'
```

### 3. 更新用户状态

```bash
curl -X PATCH "http://localhost:8080/api/v1/users/{account_id}/status" \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=xxx" \
  -d '{
    "tenant_id": "xxx",
    "status": "inactive"
  }'
```

## 前端使用

### 1. 访问用户列表
导航到：`/permissions/users`

### 2. 创建用户
点击"新建用户"按钮，填写表单

### 3. 查看用户详情和分配角色
在用户列表中点击"详情"按钮

## 测试

### 后端编译
```bash
cd /Users/looper/Documents/workspace/golang/auth-perm
go build -o bin/auth-perm cmd/api/main.go
```

### 前端构建
```bash
cd /Users/looper/Documents/workspace/golang/auth-perm/ui
npm run build
```

### 运行服务
```bash
# 后端
cd /Users/looper/Documents/workspace/golang/auth-perm
./bin/auth-perm

# 前端（开发模式）
cd /Users/looper/Documents/workspace/golang/auth-perm/ui
npm run dev
```

## 架构说明

### User vs Account 的关系
- **User 表**: 存储用户基础信息（无租户概念）
- **Account 表**: 存储账户信息（绑定租户）
- 一个 User 可以对应多个 Account（跨租户）
- 本实现以 Account 为主体，JOIN User 获取用户信息

### 租户隔离
- 所有查询都必须传入 `tenant_id`
- 后端在 Service 层校验 `account.tenant_id == request.tenant_id`
- 前端使用 `useTenant` Hook 统一管理租户上下文

### 角色分配
- 角色和账户都有 `tenant_id` 属性
- 只能在同一租户下分配角色
- 使用已有的 `assignRoleToAccount` API
- 前端确保 `tenant_id` 一致性

## 文件清单

### 后端新增/修改文件
```
internal/domain/auth/repo/
  ├── account_repo.go          (新增 SearchAccountsWithCount)
  └── user_repo.go             (新增 FindByIDs)

internal/domain/auth/dto/
  └── account_query_dto.go     (新增 UserWithAccountDTO)

internal/domain/auth/service/
  └── auth_service.go          (新增 3 个用户管理方法)

internal/controller/http/
  ├── user_handler.go          (新文件)
  └── route.go                 (新增 RegisterUserRoutes)

internal/controller/vo/
  └── user_vo.go               (新文件)

internal/container/
  └── container.go             (注册 UserHandler)
```

### 前端新增/修改文件
```
ui/types/
  └── user.ts                  (新文件)

ui/lib/api/
  └── user.ts                  (新文件)

ui/app/permissions/users/
  ├── page.tsx                 (重写)
  └── [id]/
      └── page.tsx             (新文件)
```

## 总结

本次实现完成了：
1. ✅ 后端完整的用户管理 API（列表、详情、创建、状态更新）
2. ✅ 前端用户列表页面（搜索、分页、创建、状态管理）
3. ✅ 前端用户详情页面（基本信息展示、角色分配）
4. ✅ 租户隔离和权限控制
5. ✅ 与现有系统的集成（复用 RegisterService、角色分配 API）
6. ✅ 代码编译测试通过

所有功能按照计划实现，保持了与现有代码风格和架构的一致性。
