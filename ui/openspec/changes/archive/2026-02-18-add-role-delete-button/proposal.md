# 角色详情页增加删除按钮

## 为什么

当前角色详情页（/permissions/roles/[id]）仅提供编辑权限功能，缺少删除角色的操作入口。用户需要能够直接在角色详情页删除不需要的角色。

## 变更内容

1. 在角色详情页增加删除按钮
2. 删除前显示确认对话框
3. 删除成功后跳转到角色列表页

## 功能 (Capabilities)

### 新增功能
- `role-delete`: 角色删除功能

### 修改功能
（无）

## 影响

- **前端页面**: ui/app/permissions/roles/[id]/page.tsx
- **后端API**: 已存在 DELETE /api/v1/permissions/roles/:id 接口
