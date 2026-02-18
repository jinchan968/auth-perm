# 租户删除作为禁用 - 技术设计

## 上下文

将租户的删除操作改为更新状态为 deleted（deleted 状态视为禁用）。

## 目标 / 非目标

**目标：**
- 删除操作更新状态为 deleted
- deleted 状态视为禁用状态

**非目标：**
- 不实现物理删除

## 决策

### 后端实现
- 修改 Delete 方法：将 Delete 改为 UpdateStatus(deleted)
- 不再使用物理删除

### 前端实现
- 修改前端调用：从 suspendTenant 改为 deleteTenant

## 风险 / 权衡

- 状态一致性：需要确保 deleted 状态在前端正确显示为"已禁用"
