## 为什么

前端租户详情页及子页面（编辑、设置等）的返回按钮样式和 hover 态颜色存在问题，影响用户体验。需要优化这些按钮的视觉效果，使其与整体 UI 风格保持一致。

## 变更内容

1. **修复返回按钮样式** - 调整按钮颜色、边框、圆角等
2. **修复返回按钮 hover 态** - 优化 hover 时的颜色变化
3. **统一按钮风格** - 确保与其他页面按钮风格保持一致

## 功能 (Capabilities)

### 新增功能
- `tenant-button-styles`: 租户页面按钮样式修复

### 修改功能
- 无

## 影响

- `ui/app/tenants/[id]/page.tsx` - 租户详情页
- `ui/app/tenants/[id]/edit/page.tsx` - 租户编辑页
- `ui/app/tenants/[id]/settings/page.tsx` - 租户设置页
- `ui/components/ui/button.tsx` - 按钮组件
