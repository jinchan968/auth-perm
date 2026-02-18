## 上下文

当前权限管理 API 中，部分接口通过 ID 直接操作资源时缺少租户隔离验证。攻击者可能通过猜测其他租户的 ID 来访问或操作跨租户资源。

现有已实现租户隔离的接口：
- ListPermissions - query 参数 tenant_id
- ListRoles - query 参数 tenant_id
- CreatePermission - 请求体 tenant_id
- CreateRole - 请求体 tenant_id
- AssignPermissionToRole - 请求体 tenant_id
- AssignRoleToAccount - 请求体 tenant_id

## 目标 / 非目标

**目标：**
- 为 9 个缺少租户隔离的接口添加 tenant_id 验证
- 确保所有通过 ID 直接操作资源的接口都包含租户验证

**非目标：**
- 不修改数据库结构
- 不修改已实现租户隔离的接口
- 不添加新的业务逻辑

## 决策

### 1. 参数传递方式

对于 GET/DELETE 操作（通过 URL 参数）：
- 使用 query 参数传递 tenant_id

对于 POST/PUT 操作（通过请求体）：
- 从请求体中获取 tenant_id

### 2. 验证方式

在 Handler 层解析参数后，Service 层进行租户验证：
- 查询数据库时确保 WHERE 条件包含 tenant_id
- 如果资源不属于当前租户，返回 403 错误

### 3. 修改清单

| 接口 | 文件 | 修改内容 |
|------|------|----------|
| GetPermission | permission_handler.go | 添加 tenant_id query 参数 |
| UpdatePermission | permission_handler.go | 添加 tenant_id 请求体参数 |
| DeletePermission | permission_handler.go | 添加 tenant_id query 参数 |
| GetRole | permission_handler.go | 添加 tenant_id query 参数 |
| UpdateRole | permission_handler.go | 添加 tenant_id 请求体参数 |
| DeleteRole | permission_handler.go | 添加 tenant_id query 参数 |
| GetRolePermissions | permission_handler.go | 添加 tenant_id query 参数 |
| RemovePermissionFromRole | permission_handler.go | 添加 tenant_id 请求体参数 |
| RemoveRoleFromAccount | permission_handler.go | 添加 tenant_id 请求体参数 |

## 风险 / 权衡

- **兼容性风险**：修改接口参数可能影响现有前端调用 → 需同步更新前端 API
- **性能影响**：增加租户验证可能略微增加查询时间 → 影响可忽略

## 迁移计划

1. 后端先修改接口参数（向后兼容）
2. 同步修改前端 API 调用
3. 部署后端
4. 验证功能正常
