# 租户下拉状态同步 - 技术设计

## 上下文

前端使用租户下拉组件切换租户时，存在状态同步问题：UI 显示切换到新租户，但 API 请求中携带的 `tenant_id` 仍然是旧值。

## 目标 / 非目标

**目标：**
- 确保租户下拉切换时，状态变量同步更新
- 确保所有 API 请求使用最新的 tenant_id

**非目标：**
- 不修改后端逻辑（后端已处理租户验证）
- 不修改租户下拉组件的基础样式

## 决策

### 问题定位

在 React/Next.js 中，租户下拉组件可能使用了：
1. 本地组件状态 (`useState`)
2. 全局状态管理 (`useContext`, `zustand`, `redux`)
3. URL 参数 (`useSearchParams`)

当状态更新时，需要确保：
1. 下拉组件的显示值与实际使用的值一致
2. 其他使用 tenant_id 的组件或 hooks 收到通知

### 解决方案

使用 React Context 或全局状态管理，确保 tenant_id 在应用级别统一管理：
- 创建一个 `TenantContext` 管理当前租户
- 所有需要 tenant_id 的组件从 Context 获取
- 切换租户时更新 Context 值

## 风险 / 权衡

- 需要确保状态更新的响应性
- 需要遍历所有使用 tenant_id 的地方
