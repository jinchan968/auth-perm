# 问题修复：新建用户按钮错误

## 问题描述

当在权限管理页面点击"新建用户"按钮时，出现以下错误：

```
http://localhost:3000/api/users/new?tenant_id=xxx
{"code":404,"msg":"用户不存在","error":"NOT_FOUND: 查找账户失败: 账户不存在: new"}
```

## 问题原因

权限管理页面 (`/permissions/page.tsx`) 中的"新建用户"按钮链接到了 `/permissions/users/new`，这导致：

1. Next.js 将 "new" 解析为动态路由参数 `[id]`
2. 前端尝试访问 ID 为 "new" 的用户详情页
3. 前端调用后端 API `GET /api/v1/users/new`
4. 后端尝试查找 account_id 为 "new" 的用户，返回 404 错误

## 解决方案

修改了 `/ui/app/permissions/page.tsx` 文件中"用户列表" Tab 的按钮行为：

**修改前：**
```tsx
<Link href="/permissions/users/new">
  <Button>新建用户</Button>
</Link>
```

**修改后：**
```tsx
<Link href="/permissions/users">
  <Button>管理用户</Button>
</Link>
```

## 原因说明

用户管理功能的设计是：
- **用户列表页面** (`/permissions/users`) - 包含用户列表和"新建用户"对话框
- **用户详情页面** (`/permissions/users/[id]`) - 显示用户详情和角色分配

创建用户的功能是通过用户列表页面中的对话框实现的，而不是一个单独的路由。因此：
- 点击"管理用户"按钮 → 跳转到 `/permissions/users`
- 在用户列表页面点击"新建用户"按钮 → 打开对话框
- 在用户列表页面点击"详情"按钮 → 跳转到 `/permissions/users/[id]`

## 使用流程

### 正确的创建用户流程

1. 访问权限管理页面 (`/permissions`)
2. 切换到"用户列表" Tab
3. 点击"管理用户"按钮
4. 进入用户列表页面 (`/permissions/users`)
5. 点击右上角"新建用户"按钮
6. 在弹出的对话框中填写用户信息
7. 点击"创建"按钮完成创建

### 查看用户详情流程

1. 在用户列表页面 (`/permissions/users`)
2. 点击某个用户行的"详情"按钮
3. 进入用户详情页面 (`/permissions/users/[id]`)
4. 查看用户信息和分配角色

## 注意事项

- Next.js 的动态路由 `[id]` 会匹配任何路径段，包括 "new"
- 如果需要单独的"新建"页面，应该使用固定路由放在动态路由之前，或使用不同的路径结构
- 当前设计使用对话框创建用户更符合用户体验，无需单独页面

## 测试

修复后，测试以下场景：
1. ✓ 在权限管理页面点击"管理用户"可以正常跳转到用户列表
2. ✓ 在用户列表页面点击"新建用户"弹出对话框
3. ✓ 在对话框中创建用户成功
4. ✓ 点击用户行的"详情"按钮可以正常查看用户详情

---

**修复时间**: 2026-02-19
**修复文件**: `/ui/app/permissions/page.tsx`
