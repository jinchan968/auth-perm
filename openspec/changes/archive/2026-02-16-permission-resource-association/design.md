## 上下文

当前权限系统支持通过权限资源关联表来管理权限与具体资源的映射关系。

## 权限资源关联机制

### 1. 数据模型

**PermissionResourceDO** - 权限资源关联表：
```
- ID: 关联ID (UUID)
- PermissionID: 关联的权限ID
- ResourceID: 资源标识 (如 API 路径 /api/users, 菜单 key "menu:users")
- ResourceType: 资源类型
  - api_path: API 路径
  - menu: 菜单
  - button: 按钮
- ResourceName: 资源名称
- TenantID: 租户ID
- CreatedAt/UpdatedAt/DeletedAt
```

### 2. API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /permissions/items/:permissionId/resources | 获取权限关联的资源列表 |
| POST | /permissions/items/:permissionId/resources | 创建权限资源关联 |
| POST | /permissions/items/:permissionId/resources/batch | 批量创建 |
| PUT | /permissions/items/:permissionId/resources/:resourceId | 更新资源关联 |
| DELETE | /permissions/items/:permissionId/resources/:resourceId | 删除资源关联 |
| POST | /permissions/items/:permissionId/resources/bind | 绑定资源 |
| POST | /permissions/items/:permissionId/resources/unbind | 解绑资源 |

### 3. 使用场景

1. **API 权限控制**: 将权限关联到具体的 API 路径
2. **菜单权限**: 将权限关联到菜单项
3. **按钮权限**: 将权限关联到页面按钮

### 4. 示例

创建权限并关联资源：
```json
// 1. 先创建权限
POST /api/v1/permissions/items
{
  "tenant_id": "xxx",
  "code": "user.view",
  "name": "查看用户"
}

// 2. 关联资源 (API 路径)
POST /api/v1/permissions/items/{permissionId}/resources
{
  "resource_id": "/api/v1/users",
  "resource_type": "api_path",
  "resource_name": "用户列表API"
}

// 3. 关联菜单
POST /api/v1/permissions/items/{permissionId}/resources
{
  "resource_id": "menu:users",
  "resource_type": "menu",
  "resource_name": "用户管理菜单"
}
```

## 结论

当前系统已具备完整的权限-资源关联机制，通过 `permission_resources` 表实现多对多关系。
