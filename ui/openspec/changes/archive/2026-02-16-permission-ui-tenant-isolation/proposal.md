## 为什么

后端 API 已添加租户隔离，所有权限相关接口都需要传入 tenant_id。前端页面调用这些接口时需要传递租户ID参数，确保多租户环境下的数据隔离。

## 变更内容

修改前端权限管理页面中的所有 API 调用，确保每个接口都传递 tenant_id 参数：
- 权限列表页 - listPermissions 添加 tenant_id
- 权限详情页 - getPermission, updatePermission, deletePermission 添加 tenant_id
- 角色列表页 - listRoles 添加 tenant_id
- 角色详情页 - getRole, updateRole, deleteRole 添加 tenant_id
- 角色权限分配 - getRolePermissions, assignPermissionsToRole, removePermissionFromRole 添加 tenant_id
- 用户角色分配 - assignRoleToAccount, removeRoleFromAccount 添加 tenant_id

## 功能 (Capabilities)

### 修改功能

- **permission-ui**: 修改权限管理前端页面，添加 tenant_id 参数传递

## 影响

- **前端**: 修改 app/permissions/ 目录下的所有页面组件
- **API**: 所有 API 调用都需要从当前用户获取 tenant_id
