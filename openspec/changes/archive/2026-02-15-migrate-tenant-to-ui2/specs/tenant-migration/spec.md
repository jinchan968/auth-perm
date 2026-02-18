# tenant-migration 规范

## 新增需求

### 需求:租户管理功能必须迁移到ui2
系统必须将租户管理相关文件复制到ui2目录。

#### 场景:迁移租户页面
- **当** 执行迁移操作
- **那么** ui/app/tenants 目录下的文件被复制到 ui2/app/tenants

#### 场景:迁移租户API
- **当** 执行迁移操作
- **那么** ui/app/api/tenants 目录下的文件被复制到 ui2/app/api/tenants

#### 场景:迁移租户类型
- **当** 执行迁移操作
- **那么** ui/types/tenant.ts 被复制到 ui2/types/

#### 场景:迁移租户API客户端
- **当** 执行迁移操作
- **那么** ui/lib/api/tenant.ts 被复制到 ui2/lib/api/

#### 场景:迁移侧边栏组件
- **当** 执行迁移操作
- **那么** 侧边栏组件中的租户管理入口被添加到ui2
