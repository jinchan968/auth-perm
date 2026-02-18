## 上下文

后端 API 已完成租户隔离改造，所有权限相关接口都需要 tenant_id 参数。前端需要相应修改以传递正确的租户ID。

当前前端代码中：
- API 调用缺少 tenant_id 参数
- 页面使用 `user?.tenant_id` 获取租户ID

## 目标 / 非目标

**目标：**
- 修改所有权限管理页面的 API 调用，传递 tenant_id 参数
- 确保页面从当前登录用户获取 tenant_id

**非目标：**
- 不修改后端 API
- 不修改现有的 UI 布局

## 决策

1. **tenant_id 获取方式**
   - 从 useAuthStore 获取当前用户的 tenant_id
   - 使用 `user?.tenant_id` 或 `currentUser?.tenant_id`

2. **修改清单**

| 页面 | 修改的 API 调用 |
|------|----------------|
| permissions/page.tsx | listPermissions |
| permissions/[id]/page.tsx | getPermission, updatePermission, deletePermission |
| permissions/new/page.tsx | createPermission |
| permissions/roles/page.tsx | listRoles, createRole, updateRole, deleteRole |
| permissions/roles/[id]/page.tsx | getRole, getRolePermissions, assignPermissionsToRole |
| permissions/users/page.tsx | assignRoleToAccount, removeRoleFromAccount |

## 风险 / 权衡

- **用户无租户ID**：如果当前用户没有 tenant_id，API 调用会失败 → 页面需处理空值情况
