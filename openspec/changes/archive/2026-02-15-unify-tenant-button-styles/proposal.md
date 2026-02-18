## 为什么

租户详情页右侧的操作按钮（设置、编辑、删除）与返回按钮在尺寸和样式上存在差异，影响页面整体美观。需要统一这些按钮的风格。

## 变更内容

1. **统一返回按钮与操作按钮样式** - 确保高度、宽度、内边距一致

## 功能 (Capabilities)

### 新增功能
- `unify-button-styles`: 按钮样式统一

### 修改功能
- 无

## 影响

- `ui/app/tenants/[id]/page.tsx` - 租户详情页
- `ui/app/tenants/[id]/edit/page.tsx` - 租户编辑页
- `ui/app/tenants/[id]/settings/page.tsx` - 租户设置页
