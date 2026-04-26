# 权限资源体系实施总结

## 📋 实施概述

本次实施完成了完整的前后端权限控制体系，包括：
- **后端**：超管配置、资源清单接口、API 权限拦截中间件
- **前端**：权限 Hook、PermGuard 组件、菜单/按钮权控
- **数据库**：权限资源种子数据

---

## ✅ 已完成的功能

### 1. 配置层 (Config)

**文件**: `config/config.go`

- ✅ 新增 `ServerConfig.SuperAdmin` 字段
- ✅ 绑定 `SUPER_ADMIN` 环境变量
- ✅ `.env` 已配置 `SUPER_ADMIN=looper`

### 2. Session 缓存 Username

**改动文件**:
- `internal/domain/auth/dto/session_dto.go` — 新增 `Username` 字段
- `internal/domain/auth/param/register_params.go` — `CreateSessionParams` 新增 `Username` 字段
- `internal/domain/auth/service/cache_service.go` — Redis 缓存写入 `username`
- `internal/domain/auth/service/login_service.go` — 
  - `ValidateSession` 从 Redis 解析 `username`
  - `CreateSession` 设置 `sessionDTO.Username`
  - `CreateSessionAndToken` 传递 `username`
- `internal/controller/middleware/auth.go` — 注入 `username` 到 gin context

**数据流**:
```
登录 → userDTO.Username → sessionDTO.Username → Redis: {username: "looper"}
请求 → ValidateSession → 从 Redis 读 username → c.Set("username", "looper")
```

### 3. SQL 迁移 — 权限资源种子数据

**文件**: `migrations/000007_seed_permissions_resources.sql`

**插入的权限**:
1. `menu:tenants` — 租户管理菜单
2. `menu:permissions` — 权限管理菜单
3. `menu:todos` — 待办事项菜单
4. `tab:perm.list` — 权限列表 Tab
5. `tab:perm.roles` — 角色列表 Tab
6. `tab:perm.users` — 用户列表 Tab
7. `button:tenants.show_all` — 显示全部租户按钮

**资源类型**:
- `menu` — 菜单资源 (如 `tenants`, `permissions`, `todos`)
- `button` — 按钮资源 (如 `perm.tab.list`, `tenants.show_all`)
- `api_path` — API 路径资源 (如 `/api/v1/tenants/*`, `/api/v1/permissions/*`)

**幂等性**: 使用 `ON CONFLICT DO NOTHING`，可重复执行

### 4. 后端服务层

#### 4.1 权限资源查询

**文件**: `internal/domain/permission/service/permission_service_ext.go`

- ✅ `GetAccountResourcesDetailed(accountID)` — 返回详细资源列表（含 resource_id, resource_type, resource_name）
- ✅ `GetAllResourcesForSuperAdmin()` — 超管专用，返回全部资源

**文件**: `internal/domain/permission/repo/permission_resource_repo.go`

- ✅ `FindAllResources()` — 查询所有权限资源（用于超管）

#### 4.2 资源清单接口

**Handler**: `internal/domain/auth/handler/resource_handler.go`

- ✅ `GetMyResources` — `GET /api/v1/auth/my-resources`
  - 超管（username == `SUPER_ADMIN`）：返回全量资源
  - 普通用户：根据权限返回资源
  - 无权限校验（在白名单中）

**路由注册**:
- `internal/controller/http/route.go` — 注册到 `/auth/my-resources`
- `internal/container/container.go` — DI 容器注册 `ResourceHandler`

### 5. API 权限拦截中间件

**文件**: `internal/controller/middleware/api_permission.go`

**拦截流程**:
1. **白名单**：登录、注册、资源清单、个人操作等 → 直接放行
2. **超管**：`c.Get("username") == cfg.Server.SuperAdmin` → 直接放行
3. **普通用户**：查询 `GetAccountResources(accountID, "api_path")` → 匹配当前路径

**路径匹配规则**:
- **精确匹配**：`/api/v1/tenants` 精确匹配
- **通配符匹配**：`/api/v1/tenants/*` 匹配所有子路径
- **路径参数匹配**：`/api/v1/permissions/roles` 匹配 `/api/v1/permissions/roles/:id`

**白名单路由**:
```
/health
/api/v1/auth/public/*       (登录、注册、忘记密码)
/api/v1/auth/validate
/api/v1/auth/refresh
/api/v1/auth/logout
/api/v1/auth/my-resources   (资源清单)
/api/v1/auth/profile        (个人资料)
/api/v1/auth/sessions
/api/v1/auth/devices
/api/v1/auth/security/logs
/api/v1/auth/2fa/*
/api/v1/auth/oauth/*
```

**挂载位置**: `route.go` — 全局挂载到 `router.Use()`，在所有路由注册之后

### 6. 前端权限控制

#### 6.1 API 客户端

**文件**: `ui/lib/api/resource.ts`

- ✅ `getMyResources()` — 调用 `GET /auth/my-resources`
- ✅ `ResourceItem` 类型定义

#### 6.2 权限 Hook

**文件**: `ui/hooks/use-permissions.ts`

- ✅ `usePermissionsStore` — Zustand store，存储资源列表
- ✅ `usePermissions()` — Hook，自动加载权限
- ✅ `hasMenu(menuId)` — 检查菜单权限
- ✅ `hasButton(buttonId)` — 检查按钮权限
- ✅ 403 视为无权限，不报错

**数据结构**:
```typescript
{
  resources: ResourceItem[]
  menus: Set<string>        // 如 "tenants", "permissions", "todos"
  buttons: Set<string>      // 如 "perm.tab.list", "tenants.show_all"
  apiPaths: Set<string>     // 如 "/api/v1/tenants/*"
  loading: boolean
  loaded: boolean
}
```

#### 6.3 PermGuard 组件

**文件**: `ui/components/ui/perm-guard.tsx`

**用法**:
```tsx
<PermGuard menu="tenants">
  <MenuItem>租户管理</MenuItem>
</PermGuard>

<PermGuard button="perm.tab.list">
  <Button>权限列表</Button>
</PermGuard>
```

**特性**:
- 无权限时返回 `null`（不渲染）
- 支持 `fallback` 自定义无权限显示
- 加载中不渲染（避免闪烁）

#### 6.4 应用到页面

**文件**: `ui/components/layout/dashboard-sidebar.tsx`

- ✅ 租户管理菜单 — `<PermGuard menu="tenants">`
- ✅ 权限管理菜单 — `<PermGuard menu="permissions">`
- ✅ 待办事项菜单 — `<PermGuard menu="todos">`

**文件**: `ui/app/permissions/page.tsx`

- ✅ 权限列表 Tab — `<PermGuard button="perm.tab.list">`
- ✅ 角色列表 Tab — `<PermGuard button="perm.tab.roles">`
- ✅ 用户列表 Tab — `<PermGuard button="perm.tab.users">`

---

## 🚀 部署步骤

### 1. 运行数据库迁移

```bash
cd ./

# 执行迁移
goose -dir migrations postgres "host=localhost port=5432 user=auth_perm password=your_secure_password_here dbname=auth_perm sslmode=disable" up
```

### 2. 确认环境变量

检查 `.env` 文件：
```bash
# 确认已配置超管
SUPER_ADMIN=looper
```

### 3. 重启后端服务

```bash
# 停止旧服务
pkill -f auth-perm

# 重新编译运行
make run
# 或
go run cmd/api/main.go
```

### 4. 重启前端服务

```bash
cd ui
pnpm dev
```

---

## 🧪 测试场景

### 场景 1：超管访问

1. 使用 username=`looper` 的用户登录
2. 验证：
   - 左侧菜单显示所有菜单（租户、权限、待办）
   - 权限管理页显示所有 Tab（权限列表、角色列表、用户列表）
   - API 请求全部通过，无 403

### 场景 2：无权限用户

1. 创建新用户，不分配任何角色/权限
2. 登录验证：
   - 左侧菜单仅显示首页、仪表盘（无权限的菜单不显示）
   - 访问 `/tenants` 会被后端 API 拦截（403）
   - 前端不显示无权限的按钮和 Tab

### 场景 3：部分权限用户

1. 创建角色，分配部分权限（如仅 `menu:todos`）
2. 将角色分配给用户
3. 登录验证：
   - 左侧菜单仅显示：首页、仪表盘、待办事项
   - 访问 `/api/v1/todos/*` 通过
   - 访问 `/api/v1/tenants/*` 被拦截（403）

---

## 📊 数据流图

### 登录时写入 Username

```
用户登录
  ↓
Login() 查询 user → userDTO.Username = "looper"
  ↓
CreateSessionAndToken(userDTO)
  ↓
CreateSession(params.Username = "looper")
  ↓
sessionDTO.Username = "looper"
  ↓
SetSession() → Redis: {username: "looper", ...}
```

### 每次请求权限校验

```
前端请求 /api/v1/tenants
  ↓
AuthMiddleware
  ↓ ValidateSession() → 从 Redis 读 sessionMap["username"]
  ↓ c.Set("username", "looper")
  ↓
APIPermissionMiddleware
  ↓ 检查白名单 → ❌ 不在白名单
  ↓ 检查超管 → username == "looper" == SUPER_ADMIN → ✅ 放行
  ↓
Controller 处理业务
```

### 前端权限控制

```
登录成功
  ↓
usePermissions() 自动调用 getMyResources()
  ↓ GET /api/v1/auth/my-resources
  ↓ 后端返回资源列表
  ↓
存入 Zustand store:
  - menus: Set{"tenants", "permissions", "todos"}
  - buttons: Set{"perm.tab.list", "perm.tab.roles", ...}
  ↓
PermGuard 组件读取 store
  ↓ hasMenu("tenants") → true → 渲染菜单
  ↓ hasMenu("xxx") → false → 不渲染
```

---

## ⚠️ 注意事项

1. **超管判断基于 username**
   - 优点：零 DB 查询（从 Redis 缓存读取）
   - 注意：旧 session 需重新登录才能获取 username

2. **API 中间件挂载位置**
   - 挂载在 v1 路由组上（`v1.Use()`），先于子路由的 `AuthMiddleware` 执行
   - 当 `account_id` 不存在时（未认证请求），中间件直接放行，由后续 `AuthMiddleware` 处理认证
   - 白名单路由（登录/注册等）无需认证也无需权限校验

3. **前端 403 处理**
   - `getMyResources()` 返回 403 时不报错
   - 视为无任何权限，返回空数组

4. **权限缓存**
   - 权限变更后需清除缓存
   - 角色/权限修改时调用 `cache.DeletePermissions(accountID)`
   - 前端登出时自动清除权限缓存（`usePermissionsStore.clear()`）

5. **白名单维护**
   - 精确匹配（`/api/v1/auth/validate`）和前缀匹配（`/api/v1/auth/public/`）分离
   - 新增无需权限的接口需加入 `isWhitelisted()` 白名单
   - 否则会被全局中间件拦截

---

## 🔧 故障排查

### 超管仍然被拦截

1. 检查 `.env` 中 `SUPER_ADMIN=looper` 是否配置
2. 重启后端确保配置生效
3. 重新登录，确保 session 中包含 username
4. 查看日志确认 `c.Get("username")` 返回值

### 前端菜单不显示

1. 打开浏览器控制台，查看 `getMyResources()` 返回值
2. 检查 `usePermissions()` store 中的 `menus` 集合
3. 确认权限种子数据已正确插入（执行迁移脚本）
4. 检查角色是否正确分配权限

### API 403 错误

1. 检查中间件白名单是否包含该接口
2. 检查用户是否分配了对应的 API 资源权限
3. 查看 `permission_resources` 表中是否有对应的 `api_path` 记录
4. 确认路径匹配规则（精确/通配符/参数）

---

## 📝 后续优化建议

1. **权限缓存优化**
   - 资源清单 API 响应加 Redis 缓存（10分钟）
   - 前端 Hook 加 stale-while-revalidate

2. **前端全局 403 拦截**
   - API client 统一处理 403
   - 显示 toast 提示「权限不足」

3. **动态路由守卫**
   - Next.js middleware 中检查页面权限
   - 无权限时重定向到 403 页面

4. **权限测试套件**
   - 超管场景自动化测试
   - 部分权限场景测试
   - API 拦截测试

---

## ✨ 技术亮点

1. **零额外 DB 查询** — username 通过 Redis 缓存传递，超管判断无需查库
2. **统一权限模型** — 菜单、按钮、API 路径统一管理，一套数据驱动前后端
3. **幂等迁移** — SQL 种子数据使用 `ON CONFLICT DO NOTHING`，可重复执行
4. **通配符匹配** — API 路径支持 `*` 通配符和路径参数，灵活控制
5. **组件化权控** — `PermGuard` 组件封装权限逻辑，使用简单直观

---

## 📚 相关文件清单

### 后端 (Go)

**配置**:
- `config/config.go`
- `.env`

**数据模型**:
- `internal/domain/auth/dto/session_dto.go`
- `internal/domain/auth/param/register_params.go`
- `internal/domain/permission/repo/permission_resource_repo.go`

**服务层**:
- `internal/domain/auth/service/cache_service.go`
- `internal/domain/auth/service/login_service.go`
- `internal/domain/permission/service/permission_service_ext.go`

**Handler**:
- `internal/domain/auth/handler/resource_handler.go`

**中间件**:
- `internal/controller/middleware/auth.go`
- `internal/controller/middleware/api_permission.go`

**路由**:
- `internal/controller/http/route.go`

**容器**:
- `internal/container/container.go`

**迁移**:
- `migrations/000007_seed_permissions_resources.sql`

### 前端 (TypeScript/React)

**API**:
- `ui/lib/api/resource.ts`

**Hooks**:
- `ui/hooks/use-permissions.ts`

**组件**:
- `ui/components/ui/perm-guard.tsx`
- `ui/components/layout/dashboard-sidebar.tsx`

**页面**:
- `ui/app/permissions/page.tsx`

---

**实施完成时间**: 2026-03-01
**实施人**: GitHub Copilot + User
