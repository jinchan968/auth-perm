## 1. 扩展缓存接口

- [x] 1.1 扩展 auth 的 CacheService，添加批量失效方法
- [x] 1.2 删除 permission/dto 下的接口定义，直接依赖 auth

## 2. 修改缓存 TTL

- [x] 2.1 在 `constant/values.go` 中添加新的 TTL 常量 `CacheTTLPermission`（10分钟）
- [x] 2.2 修改 `permission_service.go` 中缓存 TTL 从 `CacheTTLLong` 改为新常量

## 3. 添加缓存失效逻辑

- [x] 3.1 在 `permission_service.go` 的 `DeleteRole` 方法中添加缓存失效调用
- [x] 3.2 在 `AssignPermissionToRole` 方法中添加缓存失效调用
- [x] 3.3 在 `RemovePermissionFromRole` 方法中添加缓存失效调用
- [x] 3.4 在 `AssignRoleToAccount` 方法中添加缓存失效调用（已有）
- [x] 3.5 在 `RemoveRoleFromAccount` 方法中添加缓存失效调用（已有）
- [x] 3.6 在 `CreateRole` 方法中不需要缓存失效（新角色无用户）
- [x] 3.7 在 `UpdateRole` 方法中添加缓存失效调用

## 4. 集成延迟双删

- [x] 4.1 在 auth 的 CacheService 中集成 DoubleDeleteCache
- [x] 4.2 权限缓存删除使用延迟双删策略

## 5. 测试验证

- [x] 5.1 构建项目验证代码正确性
- [x] 5.2 代码审查确认延迟双删已集成
