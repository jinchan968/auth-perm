## 上下文

当前 ghost variant 使用 `hover:bg-accent/10`，accent 是紫色系，hover 时会产生奇怪的紫色调。应使用更中性的 slat/gray 色系。

## 目标 / 非目标

**目标：**
- ghost 按钮 hover 使用 slat-100/slat-200
- 保持与整体 UI 风格一致

**非目标：**
- 不改变按钮功能

## 决策

1. **ghost hover 改为 slat-100**：使用中性灰色替代紫色 accent

## 风险 / 权衡

- 风险低，仅涉及样式调整
