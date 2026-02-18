## 1. 实现 401 重定向功能

- [x] 1.1 在 lib/api/client.ts 中创建 401 响应拦截器函数 handleUnauthorized
- [x] 1.2 实现清除 localStorage 认证信息的逻辑
- [x] 1.3 实现清除 Zustand auth store 的逻辑（调用 logout）
- [x] 1.4 实现动态导入 Next.js router 并重定向到 /login
- [x] 1.5 添加 SSR 检查，确保只在客户端执行
- [x] 1.6 添加公开路由排除逻辑（/login, /register 等）
- [x] 1.7 将拦截器注册到 apiClient

## 2. 测试验证

- [ ] 2.1 验证 401 响应时正确清除 localStorage
- [ ] 2.2 验证 401 响应时清除 auth store
- [ ] 2.3 验证 401 响应时重定向到 /login
- [ ] 2.4 验证公开路由不触发重定向
- [ ] 2.5 验证 SSR 环境不报错
