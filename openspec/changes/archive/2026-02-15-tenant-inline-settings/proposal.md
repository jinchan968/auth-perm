## 为什么

当前租户设置需要跳转到独立页面（`/tenants/:id/settings`），用户体验不够流畅。改为在租户详情页就地编辑（inline editing），点击设置后在当前页面展开编辑区域。

## 变更内容

1. **创建 Switch 组件** - 用于功能开关就地编辑
2. **改造详情页布局** - 在详情页添加设置编辑区域
3. **实现就地编辑功能** - 功能和配额设置直接在页面编辑

## 功能 (Capabilities)

### 新增功能
- `tenant-inline-settings`: 租户设置就地编辑

### 修改功能
- 无

## 影响

- `ui/components/ui/switch.tsx` - 新建 Switch 组件
- `ui/app/tenants/[id]/page.tsx` - 租户详情页改造
