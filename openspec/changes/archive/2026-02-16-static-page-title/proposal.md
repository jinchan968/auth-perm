## 为什么

当前页面头部标题会随页面变化（如"租户管理"、"租户详情"等），用户希望左上角标题固定为"Auth-Perm"，与其他页面保持一致。

## 变更内容

1. 租户列表页头部标题改为"Auth-Perm"
2. 租户详情页头部标题改为"Auth-Perm"
3. 新建租户页头部标题改为"Auth-Perm"

## 功能 (Capabilities)

### 新增功能
- 无

### 修改功能
- 无

## 影响

- 前端：`ui/app/tenants/page.tsx` - 租户列表页
- 前端：`ui/app/tenants/[id]/page.tsx` - 租户详情页
- 前端：`ui/app/tenants/new/page.tsx` - 新建租户页
