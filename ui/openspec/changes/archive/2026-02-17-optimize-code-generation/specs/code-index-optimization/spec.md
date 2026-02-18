## 新增需求

### 需求:优化Code查询性能
系统必须在查询最大Code时使用高效的方式。

#### 场景:租户表使用索引查询最大Code
- **当** 调用租户服务创建新租户时
- **那么** 后端使用 `SELECT MAX(code) FROM tenants WHERE code LIKE 'T%'` 或 `ORDER BY code DESC LIMIT 1` 查询

#### 场景:权限表按租户查询最大Code
- **当** 调用权限服务创建新权限时
- **那么** 后端使用 `SELECT MAX(code) FROM permissions WHERE tenant_id = ? AND code LIKE 'P%'` 查询

#### 场景:角色表按租户查询最大Code
- **当** 调用角色服务创建新角色时
- **那么** 后端使用 `SELECT MAX(code) FROM roles WHERE tenant_id = ? AND code LIKE 'R%'` 查询

### 需求:数据库索引完整性
系统必须确保Code字段查询有适当的索引支持。

#### 场景:租户表Code索引存在
- **当** 检查租户表索引时
- **那么** 存在code字段的B-tree索引

#### 场景:权限表联合索引存在
- **当** 检查权限表索引时
- **那么** 存在(tenant_id, code)的联合索引

#### 场景:角色表联合索引存在
- **当** 检查角色表索引时
- **那么** 存在(tenant_id, code)的联合索引
