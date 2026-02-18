## 为什么

当前租户的编辑和设置分为两个独立功能：- 编辑：修改基本信息（名称、状态、套餐）- 设置：修改功能开关和配额用户体验上需要分别点击不同按钮，后端也有两个独立接口。需要合并为一个统一的编辑功能。

## 变更内容

1. **后端接口合并** - 将 PUT `/tenants/:id` 和 PUT `/tenants/:id/settings` 合并为单一接口
2. **前端合并** - 将"编辑"和"编辑设置"合并为一个编辑区域
3. **删除冗余代码** - 删除独立的设置编辑页面

## 功能 (Capabilities)

### 新增功能
- `merge-tenant-edit-settings`: 租户编辑与设置合并

### 修改功能
- 无

## 影响

- 后端：`internal/domain/tenant/` - Handler、Service、VO、Param 修改
- 前端：`ui/app/tenants/[id]/page.tsx` - 合并编辑区域
- 前端：删除 `ui/app/tenants/[id]/edit/page.tsx` 和 `ui/app/tenants/[id]/settings/page.tsx`
