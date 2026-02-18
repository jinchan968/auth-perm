## 为什么

当前权限系统存在缓存失效机制不完善的问题：权限缓存 TTL 设置为 24 小时过长，当角色权限发生变更时，用户可能需要等待最多 24 小时才能获得最新的权限。这会导致以下问题：

1. **权限变更延迟**：管理员修改用户角色后，用户无法立即获得/失去相应权限
2. **安全风险**：回收恶意用户权限后，其仍可能在 24 小时内继续访问
3. **数据不一致**：不同用户看到不同的权限状态

## 变更内容

1. **优化缓存 TTL**：将权限缓存 TTL 从 24 小时缩短至 5-10 分钟
2. **完善失效机制**：在权限变更（角色创建/修改/删除、权限分配/回收）时主动失效缓存
3. **实现延迟双删**：使用项目已有的 DoubleDeleteCache 策略确保缓存一致性

## 功能 (Capabilities)

### 新增功能
- `permission-cache-invalidation`: 权限缓存失效机制
  - 缩短缓存 TTL
  - 权限变更时主动失效缓存
  - 支持批量失效和模式失效

## 影响

- `internal/domain/permission/service/permission_service.go`: 修改缓存失效调用
- `internal/domain/permission/dto/permission_interface.go`: 扩展 CacheService 接口
- `internal/domain/permission/repo/permission_repo.go`: 在权限变更操作中添加缓存失效
