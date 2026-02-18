## 为什么

当前权限列表页面使用当前登录用户的 tenant_id 进行过滤，但用户无法切换查看其他租户的权限。在多租户管理场景下，管理员需要能够查看和管理不同租户的权限，因此需要添加租户过滤下拉框。

## 变更内容

- 在权限列表页面顶部添加租户下拉选择器
- 调用租户列表 API (`listTenants`) 获取可选租户列表
- 租户下拉支持按名称搜索过滤
- 选择租户后自动更新权限列表的 tenant_id 过滤条件

## 功能 (Capabilities)

### 新增功能
- **permission-list-tenant-filter**: 权限列表租户过滤功能

### 修改功能
（无）

## 影响

- 前端权限列表页面 (`app/permissions/page.tsx`)
- 使用现有的 `lib/api/tenant.ts` 中的 `listTenants` API
- 需要添加 UI 组件（Select 下拉框）
