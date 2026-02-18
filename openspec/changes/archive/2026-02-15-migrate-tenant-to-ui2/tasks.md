# UI迁移任务清单

## 1. 迁移租户管理功能

- [x] 1.1 创建 ui2/app/tenants 目录并复制页面文件
- [x] 1.2 创建 ui2/app/api/tenants 目录并复制API文件
- [x] 1.3 复制 ui/types/tenant.ts 到 ui2/types/
- [x] 1.4 创建 ui2/lib/api 目录并复制 tenant.ts
- [x] 1.5 更新 ui2 侧边栏添加租户管理入口

## 2. 目录重命名

- [x] 2.1 重命名 ui2 为 ui
- [x] 2.2 重命名 ui 为 ui-deprecated

## 3. 验证

- [x] 3.1 验证构建是否正常
