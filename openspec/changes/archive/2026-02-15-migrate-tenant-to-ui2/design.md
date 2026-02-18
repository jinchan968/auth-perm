# UI 迁移 Design

## 上下文

项目中存在两个前端目录：
- `ui`: 旧版前端项目，包含已实现的租户管理功能
- `ui2`: 新版前端项目，需要包含租户管理功能

需要将租户管理功能从 `ui` 迁移到 `ui2`，并将 `ui2` 替换为正式的 `ui`。

## 目标 / 非目标

**目标：**
- 将 `ui/app/tenants` 相关文件复制到 `ui2/app/tenants`
- 将 `ui/app/api/tenants` 相关文件复制到 `ui2/app/api/tenants`
- 将 `ui/types/tenant.ts` 复制到 `ui2/types/`
- 将 `ui/lib/api/tenant.ts` 复制到 `ui2/lib/api/`
- 将 `ui/components/layout/dashboard-sidebar.tsx` 更新合并到 `ui2`
- 将 `ui` 重命名为 `ui-deprecated`
- 将 `ui2` 重命名为 `ui`

**非目标：**
- 不合并其他未完成的功能
- 不删除任何历史代码

## 决策

### 1. 迁移策略
**决策**：采用文件复制方式迁移

**理由**：
- 直接复制可以保留文件的完整历史
- 避免复杂的合并冲突

### 2. 目录重命名顺序
**决策**：先重命名 ui2 为 ui，再重命名 ui 为 ui-deprecated

**理由**：
- 确保 ui 目录始终存在
- 避免目录不存在导致的构建错误

## 风险 / 权衡

1. **文件覆盖** → 迁移前确认 ui2 中无同名文件需要保留
2. **依赖差异** → 迁移后需验证构建是否正常

## 待定事项

1. 确认是否需要迁移其他共享组件
