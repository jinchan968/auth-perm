## 上下文

当前前端 API 客户端 (`lib/api/client.ts`) 在接收到 401 Unauthorized 响应时会抛出 `ApiError`，但不会自动清除认证状态或重定向用户到登录页面。用户停留在当前页面，不知道会话已过期。

现有代码已经有响应拦截器机制 (`addResponseInterceptor`)，可以在响应处理前拦截并执行自定义逻辑。

## 目标 / 非目标

**目标：**
- 在 API 客户端全局捕获 401 响应
- 自动清除本地存储的认证信息 (localStorage + Zustand store)
- 自动重定向用户到登录页面 (`/login`)

**非目标：**
- 不修改后端 API
- 不修改现有的 UI 布局
- 不处理其他错误码（如 403、500 等），保持现有行为

## 决策

1. **实现方式**：使用现有的响应拦截器机制
   - 方案 A：直接在 `handleResponse` 中处理 401
   - 方案 B：添加全局响应拦截器 (推荐 - 符合开闭原则，不修改现有代码)

2. **路由跳转方式**：使用 Next.js router
   - 由于 API 客户端是通用模块，需要动态导入 `next/navigation` 的 `useRouter` 来避免 SSR 问题

3. **认证状态清除**：
   - 清除 localStorage 中的 `auth_token` 和 `auth-storage`
   - 调用 authStore 的 `logout()` 方法清除状态

## 风险 / 权衡

- **循环重定向**：如果登录页面本身调用需要认证的 API，可能会导致循环重定向
  - 缓解：排除登录页面、注册页面等公开路由
- **SSR 兼容**：API 客户端可能在 SSR 环境下被调用
  - 缓解：检查 `typeof window` 确保只在客户端执行重定向
