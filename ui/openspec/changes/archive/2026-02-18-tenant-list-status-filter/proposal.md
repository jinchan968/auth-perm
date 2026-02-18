# 租户列表状态过滤

## 为什么

在权限和角色管理页面，需要根据租户状态过滤租户列表。默认只显示 active 状态的租户，通过勾选框可以显示全部租户。

## 变更内容

1. 后端：修改租户列表接口，添加可选的 status 参数，不传参时默认过滤 active 状态
2. 前端权限/角色页面：添加"显示全部租户"勾选框，默认只显示 active 租户
3. 前端租户列表页：不加过滤，显示全部租户

## 功能 (Capabilities)

### 修改功能
- `tenant-list-api`: 租户列表接口添加状态过滤参数
- `tenant-filter-checkbox`: 权限/角色页面租户过滤勾选框

## 影响

- **后端 API**: GET /api/v1/tenants
- **前端页面**:
  - ui/app/permissions/page.tsx
  - ui/app/permissions/roles/[id]/page.tsx
  - ui/app/tenants/page.tsx（不变）
