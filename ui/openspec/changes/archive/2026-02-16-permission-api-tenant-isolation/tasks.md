## 1. Permission CRUD 租户隔离

- [x] 1.1 修改 GetPermission Handler - 添加 tenant_id query 参数
- [x] 1.2 修改 UpdatePermission Handler - 添加 tenant_id 请求体参数
- [x] 1.3 修改 DeletePermission Handler - 添加 tenant_id query 参数
- [x] 1.4 修改对应的 Param 结构体
- [x] 1.5 修改 Service 层添加租户验证逻辑
- [x] 1.6 修改 Repository 层查询添加 tenant_id 条件

## 2. Role CRUD 租户隔离

- [x] 2.1 修改 GetRole Handler - 添加 tenant_id query 参数
- [x] 2.2 修改 UpdateRole Handler - 添加 tenant_id 请求体参数
- [x] 2.3 修改 DeleteRole Handler - 添加 tenant_id query 参数
- [x] 2.4 修改对应的 Param 结构体
- [x] 2.5 修改 Service 层添加租户验证逻辑
- [x] 2.6 修改 Repository 层查询添加 tenant_id 条件

## 3. Role-Permission 租户隔离

- [x] 3.1 修改 GetRolePermissions Handler - 添加 tenant_id query 参数
- [x] 3.2 修改 RemovePermissionFromRole Handler - 添加 tenant_id 请求体参数
- [x] 3.3 修改对应的 Param 结构体
- [x] 3.4 修改 Service 层添加租户验证逻辑

## 4. Account-Role 租户隔离

- [x] 4.1 修改 RemoveRoleFromAccount Handler - 添加 tenant_id 请求体参数
- [x] 4.2 修改对应的 Param 结构体
- [x] 4.3 修改 Service 层添加租户验证逻辑

## 5. 前端 API 同步

- [x] 5.1 更新 lib/api/permission.ts - 添加 tenant_id 参数
- [x] 5.2 更新 lib/api/role.ts - 添加 tenant_id 参数
- [x] 5.3 测试前端调用是否正常
