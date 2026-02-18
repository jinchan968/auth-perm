## 上下文

当前系统后端已提供完整的权限管理API（权限CRUD、角色管理、角色权限分配、用户角色分配），但前端尚未实现对应的管理页面。用户需要通过可视化界面管理租户下的权限资源。

技术栈：
- Next.js App Router
- TypeScript
- shadcn/ui 组件库
- Zustand 状态管理
- 后端 Go + Gin + GORM

## 目标 / 非目标

**目标：**
- 实现权限列表页面，展示租户下所有权限
- 实现权限创建/编辑/删除功能
- 实现角色管理页面，展示租户下所有角色
- 实现角色权限分配功能
- 实现用户角色分配功能
- 保持与现有租户管理页面一致的Dashboard布局风格

**非目标：**
- 修改后端API（后端已完备）
- 实现权限检查功能（用户登录后的权限验证）
- 实现权限缓存管理

## 决策

1. **页面路由设计**
   - `/permissions` - 权限列表页
   - `/permissions/[id]` - 权限详情/编辑页
   - `/permissions/roles` - 角色管理页
   - `/permissions/users` - 用户角色分配页

2. **布局方案**
   - 复用现有的 DashboardLayout（header + sidebar + main）
   - 使用 AvatarDropdown 和 DashboardSidebar 组件
   - 保持与租户管理页面一致的样式

3. **状态管理**
   - 使用本地 useState 管理表单状态
   - 使用已有的 auth-store 获取当前用户和租户信息

4. **API调用**
   - 使用已有的 API 客户端模式
   - 创建 `lib/api/permission.ts` 和 `lib/api/role.ts`

## 风险 / 权衡

- **多租户隔离**：需要确保所有API调用都传入正确的 tenant_id，当前租户信息从 auth-store 获取
- **权限与角色的关联**：角色权限分配需要处理多对多关系，UI上使用 CheckboxGroup 或 Switch 组件
- **系统权限保护**：后端标记为 is_system=true 的权限不可修改，前端需友好提示
