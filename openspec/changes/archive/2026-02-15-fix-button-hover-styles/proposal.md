## 为什么

租户列表中的"查看"按钮使用 ghost variant，其 hover 态使用 accent 颜色（紫色系），视觉效果奇怪。需要统一全局按钮的 hover 颜色，使其更加协调。

## 变更内容

1. **修复 ghost 按钮 hover 颜色** - 从 accent 改为 slat-100
2. **统一全局按钮 hover 风格** - 确保所有按钮 hover 使用一致的色系

## 功能 (Capabilities)

### 新增功能
- `button-hover-styles`: 按钮 hover 样式统一

### 修改功能
- 无

## 影响

- `ui/components/ui/button.tsx` - 按钮组件
