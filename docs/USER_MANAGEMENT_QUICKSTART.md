# 用户管理功能 - 快速启动指南

## 功能概述

✅ **后端功能**
- 用户列表查询（支持搜索、筛选、分页）
- 用户详情查看
- 创建用户（管理员功能）
- 更新用户状态（启用/停用/暂停）
- 租户隔离和权限控制

✅ **前端功能**
- 用户列表页面（`/permissions/users`）
- 用户详情页面（`/permissions/users/[id]`）
- 角色分配功能
- 创建用户对话框
- 状态管理

## 快速启动

### 1. 启动后端服务

```bash
cd /Users/looper/Documents/workspace/golang/auth-perm

# 编译（如果还没编译）
go build -o bin/auth-perm cmd/api/main.go

# 启动服务
./bin/auth-perm
```

后端将在 `http://localhost:8080` 启动

### 2. 启动前端服务

```bash
cd /Users/looper/Documents/workspace/golang/auth-perm/ui

# 安装依赖（如果还没安装）
npm install

# 启动开发服务器
npm run dev
```

前端将在 `http://localhost:3000` 启动

### 3. 访问用户管理页面

1. 打开浏览器访问：`http://localhost:3000/login`
2. 登录系统
3. 导航到：**权限管理 -> 用户列表** 或直接访问 `http://localhost:3000/permissions/users`

## 主要功能使用

### 查看用户列表
1. 选择租户（页面顶部的租户选择器）
2. 查看当前租户下的所有用户
3. 使用搜索框搜索用户（username/email/phone）
4. 使用分页浏览更多用户

### 创建新用户
1. 点击右上角的"新建用户"按钮
2. 填写表单：
   - 选择标识符类型（邮箱或手机号）
   - 填写邮箱/手机号
   - 填写用户名
   - 填写密码（至少6位）
   - 确认密码
   - 可选填写昵称
3. 点击"创建"按钮
4. 新用户将自动添加到当前租户

### 管理用户状态
在用户列表中，每个用户都有状态下拉选择器：
- **活跃**：用户可以正常登录
- **停用**：用户无法登录
- **暂停**：临时暂停用户

直接在下拉框中选择新状态即可更新

### 分配角色
1. 点击用户列表中的"详情"按钮
2. 进入用户详情页面
3. 在"角色分配"区域查看可用角色
4. 点击角色卡片进行选择/取消选择
5. 点击"保存角色分配"按钮保存更改

## API 接口

所有接口都需要认证（Cookie: auth_token）

### 1. 获取用户列表
```http
GET /api/v1/users?tenant_id={tenant_id}&keyword={keyword}&page=1&page_size=10
```

### 2. 获取用户详情
```http
GET /api/v1/users/{account_id}?tenant_id={tenant_id}
```

### 3. 创建用户
```http
POST /api/v1/users
Content-Type: application/json

{
  "identifier_type": "email",
  "email": "user@example.com",
  "username": "newuser",
  "password": "password123",
  "confirm_password": "password123",
  "tenant_id": "{tenant_id}",
  "nickname": "New User"
}
```

### 4. 更新用户状态
```http
PATCH /api/v1/users/{account_id}/status
Content-Type: application/json

{
  "tenant_id": "{tenant_id}",
  "status": "inactive"
}
```

## 架构说明

### 数据模型
- **User 表**：用户基本信息（username, email, phone, password_hash等）
- **Account 表**：账户信息（tenant_id, account_type, status, last_login_at等）
- 关系：一个 User 可以有多个 Account（跨租户）

### 租户隔离
- 所有操作都基于 `tenant_id` 进行隔离
- 前端使用 `useTenant` Hook 管理租户上下文
- 后端在 Service 层校验租户归属，防止越权

### 权限控制
- 当前：所有接口使用 `AuthMiddleware`（需要登录）
- 租户隔离：通过 `tenant_id` 参数限制访问范围
- 未来可增强：添加 `AdminPermissionMiddleware` 限制为管理员

## 技术栈

### 后端
- **语言**：Go 1.21+
- **框架**：Gin
- **数据库**：PostgreSQL
- **缓存**：Redis
- **架构**：领域驱动设计（DDD）

### 前端
- **框架**：Next.js 14
- **语言**：TypeScript
- **UI 库**：Tailwind CSS + shadcn/ui
- **状态管理**：Zustand
- **数据获取**：自定义 API Client

## 目录结构

### 后端
```
internal/
├── controller/
│   ├── http/
│   │   └── user_handler.go      # 用户管理 HTTP 处理器
│   └── vo/
│       └── user_vo.go            # 用户管理 VO
├── domain/
│   └── auth/
│       ├── service/
│       │   └── auth_service.go  # 新增用户管理方法
│       ├── repo/
│       │   ├── user_repo.go     # 新增 FindByIDs
│       │   └── account_repo.go  # 新增 SearchAccountsWithCount
│       └── dto/
│           └── account_query_dto.go  # 新增 UserWithAccountDTO
└── container/
    └── container.go              # DI 容器注册
```

### 前端
```
ui/
├── types/
│   └── user.ts                   # 用户类型定义
├── lib/
│   └── api/
│       └── user.ts               # 用户 API 客户端
└── app/
    └── permissions/
        └── users/
            ├── page.tsx          # 用户列表页面
            └── [id]/
                └── page.tsx      # 用户详情页面
```

## 常见问题

### Q: 为什么创建用户需要密码？
A: 用户管理的"创建用户"功能复用了注册服务，管理员创建的用户可以立即使用密码登录。

### Q: 用户和账户有什么区别？
A: User 是跨租户的用户实体，Account 是租户内的账户实体。一个用户可以在多个租户下有不同的账户。

### Q: 如何给用户分配角色？
A: 进入用户详情页面，在"角色分配"区域选择角色并保存。注意角色和账户必须在同一租户下。

### Q: 状态更新会立即生效吗？
A: 是的，状态更新后会清除相关缓存，用户的登录状态会立即受影响。

## 下一步增强

可考虑的功能增强：
1. 批量操作（批量启用/停用用户）
2. 用户导入/导出功能
3. 高级筛选（按角色、按创建时间等）
4. 用户活动日志查看
5. 密码重置功能（管理员为用户重置密码）
6. 用户资料编辑（管理员修改用户信息）

## 支持

如有问题，请查看：
- 详细实现文档：`USER_MANAGEMENT_IMPLEMENTATION.md`
- API 文档：Swagger UI（如果已配置）
- 日志文件：`logs/` 目录

---

**实现完成时间**: 2026-02-19
**版本**: 1.0.0
