## 为什么

权限管理后端 API 中，部分接口缺少租户隔离验证，存在安全漏洞。具体表现为：通过 ID 直接操作资源时，没有验证该资源是否属于当前租户，可能导致跨租户访问。

## 变更内容

修复以下缺少租户隔离的接口：

1. **GetPermission** - 添加 tenant_id 参数验证
2. **UpdatePermission** - 添加 tenant_id 参数验证
3. **DeletePermission** - 添加 tenant_id 参数验证
4. **GetRole** - 添加 tenant_id 参数验证
5. **UpdateRole** - 添加 tenant_id 参数验证
6. **DeleteRole** - 添加 tenant_id 参数验证
7. **GetRolePermissions** - 添加 tenant_id 参数验证
8. **RemovePermissionFromRole** - 添加 tenant_id 参数验证
9. **RemoveRoleFromAccount** - 添加 tenant_id 参数验证

## 功能 (Capabilities)

### 修改功能

- **permission-crud**: 权限 CRUD 接口添加租户隔离
- **role-crud**: 角色 CRUD 接口添加租户隔离
- **role-permission**: 角色权限关联接口添加租户隔离
- **account-role**: 账户角色关联接口添加租户隔离

## 影响

- **Handler 层**：修改 9 个 Handler 函数的参数处理
- **Param 层**：修改对应 Param 结构体，添加 TenantID
- **Service 层**：修改 Service 方法，传递 TenantID 进行验证
- **Repo 层**：确保 Repository 层查询时包含 TenantID 条件
