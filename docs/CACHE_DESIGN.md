# 权限缓存设计

本文档说明权限系统的双层缓存架构、失效策略，以及可能出现的数据不一致场景和排查方法。

## 缓存架构总览

权限系统使用两层 Redis 缓存，服务于不同的查询场景：

```
请求进入 APIPermissionMiddleware
  │
  ├─ [Layer 1] auth:permission:{accountID} → []permissionCode
  │   用途: CheckPermission (检查账户是否拥有指定权限 code)
  │   TTL:  10 分钟
  │   写入: getAccountPermissions() 缓存未命中时
  │
  └─ [Layer 2] auth:resource:{accountID}:{resourceType} → []resourceID  ← 新增
      用途: GetAccountResources (获取账户可访问的 API 路径/菜单/按钮等)
      TTL:  10 分钟
      写入: GetAccountResources() 缓存未命中时
```

### 缓存 Key 格式

| Key 模式 | 值类型 | 示例 |
|----------|--------|------|
| `auth:permission:{accountID}` | JSON `[]string` | `["user:read","tenant:write"]` |
| `auth:resource:{accountID}:api_path` | JSON `[]string` | `["/api/v1/users/*","/api/v1/tenants"]` |
| `auth:resource:{accountID}:menu` | JSON `[]string` | `["tenants","permissions","todos"]` |
| `auth:resource:{accountID}:button` | JSON `[]string` | `["tenants.show_all"]` |
| `auth:resource:{accountID}:field` | JSON `[]string` | — |

### 删除策略

所有权限/资源缓存的删除统一使用 **延迟双删（Double-Delete）**：
1. 立即 `DEL key`
2. 500ms 后通过后台 worker 再次 `DEL key`（最多 3 次指数退避重试）

---

## 缓存失效矩阵

| # | 变更方法 | 触发时机 | 失效的缓存层 | 失效范围 | 失效方式 |
|---|---------|---------|-------------|---------|---------|
| 1 | `CreatePermission` | 新建权限 code | 无 | — | 新建未分配，无需失效 |
| 2 | `UpdatePermission` | 启用/停用权限 | Layer 1 + Layer 2 | 持有该权限的所有账户 | `FindAccountIDsByPermissionID` → 批量失效 |
| 3 | `DeletePermission` | 删除权限（仅在无关联角色时允许） | Layer 1 + Layer 2 | 持有该权限的所有账户 | `FindAccountIDsByPermissionID` → 批量失效 |
| 4 | `CreateRole` | 新建角色 | 无 | — | 新建未分配，无需失效 |
| 5 | `UpdateRole` | 角色属性变更 | Layer 1 + Layer 2 | 拥有该角色的所有账户 | `FindAccountIDsByRoleIDs` → 批量失效 |
| 6 | `DeleteRole` | 删除角色 | Layer 1 + Layer 2 | 拥有该角色的所有账户 | `FindAccountIDsByRoleIDs` → 批量失效 |
| 7 | `AssignPermissionToRole` | 角色获得新权限 | Layer 1 + Layer 2 | 拥有该角色的所有账户 | `FindAccountIDsByRoleIDs` → 批量失效 |
| 8 | `RemovePermissionFromRole` | 角色失去权限 | Layer 1 + Layer 2 | 拥有该角色的所有账户 | `FindAccountIDsByRoleIDs` → 批量失效 |
| 9 | `AssignRoleToAccount` | 账户获得新角色 | Layer 1 + Layer 2 | 单个账户 | `DeletePermissions(accountID)` |
| 10 | `RemoveRoleFromAccount` | 账户失去角色 | Layer 1 + Layer 2 | 单个账户 | `DeletePermissions(accountID)` |
| 11 | `BindPermissionResources` | 权限绑定新资源 | Layer 2 | 持有该权限的所有账户 | `FindAccountIDsByPermissionID` → 批量失效 |
| 12 | `UnbindPermissionResources` | 权限解绑资源 | Layer 2 | 持有该权限的所有账户 | `FindAccountIDsByPermissionID` → 批量失效 |
| 13 | `DeletePermissionResource` | 删除单个资源关联 | Layer 2 | 持有该权限的所有账户 | 先 `FindByID` 获取 permissionID → 同上 |
| 14 | `UpdatePermissionResource` | 更新资源名称 | Layer 2 | 持有该权限的所有账户 | 先 `FindByID` 获取 permissionID → 同上 |
| 15 | `UpdateAccountStatus` | 禁用/暂停/注销账户 | Layer 1 + Layer 2 + Account cache | 单个账户 | 三重失效 |

---

## 数据不一致场景

### 场景 1：TTL 窗口期（缓存未过期但 DB 已变更）

**触发条件**
任何缓存失效未能执行时（如 Redis 暂时不可达、双删第二次未命中），缓存将在 TTL（10分钟）内保持旧值。

**不一致窗口**
最大 10 分钟（CacheTTLPermission）。

**影响**
- 用户可能短暂保留已撤销的权限（如角色被移除后仍可访问 API）。
- 用户可能短暂缺少新授予的权限（如角色被分配后 API 仍返回 403）。

**缓解措施**
- TTL 作为安全网上限（10 分钟）。
- 所有变更点均主动调用失效方法，双删机制最大化命中率。

**排查方法**
```bash
# 检查 Redis 中某账户的缓存内容
redis-cli GET "auth:permission:{accountID}"
redis-cli GET "auth:resource:{accountID}:api_path"

# 检查 TTL（-1 表示永不过期，-2 表示 key 不存在）
redis-cli TTL "auth:permission:{accountID}"

# 手动清除
redis-cli DEL "auth:permission:{accountID}"
redis-cli DEL "auth:resource:{accountID}:api_path"
redis-cli DEL "auth:resource:{accountID}:menu"
redis-cli DEL "auth:resource:{accountID}:button"
redis-cli DEL "auth:resource:{accountID}:field"
```

---

### 场景 2：双删失败（延迟二次删除异常）

**触发条件**
- 双删的延迟 worker 队列满或 goroutine panic
- Redis 连接在延迟执行期间中断

**不一致窗口**
最大 10 分钟（TTL）。

**影响**
与场景 1 相同。

**缓解措施**
- TTL 10 分钟兜底。
- 双删 worker 有 3 次指数退避重试（`internal/infra/cache/double_delete.go`）。

**自查方法**
检查应用日志中是否有 `Failed to execute delayed delete` 或类似的缓存删除失败日志。

---

### 场景 3：部分失效遗漏（权限→角色→账户链路查询有漏）

**触发条件**
`FindAccountIDsByPermissionID` 使用单次 SQL JOIN（`account_roles` + `role_permissions`）查询受影响的账户。如果账户或角色在两次独立查询之间发生变化，可能遗漏。

**不一致窗口**
无。单次 JOIN 保证原子性。

**影响**
无。

**缓解措施**
使用单次 JOIN 查询，避免多次查询间的竞态窗口。代码位置：`./internal/domain/permission/repo/permission_repo.go:FindAccountIDsByPermissionID`。

---

### 场景 4：并发读写（管理员改权限的同时用户请求）

**触发条件**
```
时间线:  管理员修改权限-资源映射  →  A 用户读取缓存（旧值） →  管理员提交 →  管理员清缓存
                                                                    ↑
                                                            A 用户已持有旧数据
```

**不一致窗口**
仅影响当前正在进行的那一次请求。

**影响**
极低。单次请求使用旧权限数据，在 handle 层面可能被 403 或不一致地放行。下一次请求命中新缓存。

**缓解措施**
无需额外措施。这是读-写并发中的天然窗口，TTL 和主动失效已将窗口缩到最小。

---

### 场景 5：DeletePermissionResource 先查后删的时序窗口

**触发条件**
`DeletePermissionResource(id)` 的执行顺序：
1. `FindByID(id)` 获取 `permissionID`
2. 通过 `permissionID` 查受影响账户
3. `DeleteByID(id)` 删除记录
4. 失效受影响账户的缓存

如果步骤 2 和 3 之间，有另一个请求修改了同一记录，会出现竞争。

**不一致窗口**
极低。`DeletePermissionResource` 在同一服务、同一 goroutine 内顺序执行。

**缓解措施**
操作在同一请求上下文中顺序执行，实践中无并发风险。若未来拆分微服务需加分布式锁。

---

### 场景 6：UnbindPermissionResources 批量解绑中间态

**触发条件**
`UnbindPermissionResources` 从 `params.PermissionID` 直接获取 permissionID，在循环删除每个 resource 之前一次性查完所有受影响账户。

**不一致窗口**
如果批量删除过程中有新的账户获得该权限（通过角色分配），其缓存不会被清除。

**影响**
低。新获得权限的账户在 10 分钟内可能无法访问对应资源（缓存未命中时会回源 DB，实际无影响）。

**缓解措施**
资源缓存未命中时回源 DB 查询，不存在"假阴性"问题。

---

### 场景 7：账户禁用后权限缓存未及时失效

**触发条件**
管理员禁用账户 → `UpdateAccountStatus` 更新 `accounts.status` → 清除 `auth:account:{id}` 缓存 → 但未清除 `auth:permission:{id}` 和 `auth:resource:{id}:*`。

原实现中仅清除 account 缓存，**遗漏**了权限/资源缓存的失效。

**不一致窗口**
最大 10 分钟。

**影响**
- 中间件层（`APIPermissionMiddleware`）：资源缓存仍然有效 → 请求通过中间件
- 服务层（后续 handler/service）：`FindAccountByID` 从 DB 查到禁用状态 → 业务拒绝
- 最终结果：中间件多放行了一次，但业务层拦截。安全无泄漏但响应路径多走了一层。

**缓解措施**
已在 `UpdateAccountStatus` 中追加 `DeletePermissions(accountID)` + `DeleteAccountResources(accountID)`。

---

### 场景 8：Redis 内存压力

**触发条件**
每个账户现在可能拥有最多 5 个缓存 key（1 个 permission + 4 个 resource type）。在大规模部署（10 万+ 账户）下 Redis 内存消耗增加。

**影响**
- 每账户 ~5 keys，约 500 bytes（假设每 key 的资源列表平均 10 个路径）
- 10 万账户 ≈ 50 MB 额外内存

**缓解措施**
- TTL 10 分钟自动过期清理。
- 资源缓存 key 按 resourceType 拆分而非全存一个 key，避免大 key 的序列化开销。
- 若内存压力显著，可考虑合并为单 key 或缩短 TTL。

**监控建议**
```bash
# 统计资源缓存 key 数量
redis-cli --scan --pattern "auth:resource:*" | wc -l

# 估算单个 key 大小
redis-cli DEBUG OBJECT "auth:resource:{accountID}:api_path"
```

---

## 排查指南

### 现象 → 缓存层 → 排查步骤

| 现象 | 可能涉及的缓存层 | 排查步骤 |
|------|-----------------|---------|
| 用户权限变更后仍能看到旧菜单 | Layer 2 (menu) | 1. 手动清 `auth:resource:{accountID}:menu` 2. 重新登录 3. 若恢复则确认是缓存问题 |
| API 返回 403 但用户应有权访问 | Layer 2 (api_path) | 1. 检查 `auth:resource:{accountID}:api_path` 2. 验证 DB 中 account → roles → permissions → resources 链路 |
| API 放行但用户应被拒绝 | Layer 2 (api_path) | 1. 检查资源缓存是否包含不该有的路径 2. 手动清缓存重新请求 3. 检查失效逻辑是否被触发 |
| 账户禁用后仍可访问 API | Layer 1 + Layer 2 | 1. 检查 `UpdateAccountStatus` 是否调用了失效 2. 手动清该账户所有 key 3. 检查 account cache 与 permission cache 是否同时失效 |
| 所有用户权限异常（大面积） | 全局 | 1. 检查 Redis 连接状态 2. 检查双删 worker 是否有异常日志 3. 考虑 `FLUSHDB` 后观察恢复情况 |

### 手动清除某账户全部缓存

```bash
ACCOUNT_ID="your-account-uuid"
redis-cli DEL \
  "auth:permission:${ACCOUNT_ID}" \
  "auth:resource:${ACCOUNT_ID}:api_path" \
  "auth:resource:${ACCOUNT_ID}:menu" \
  "auth:resource:${ACCOUNT_ID}:button" \
  "auth:resource:${ACCOUNT_ID}:field"
```

### 验证缓存状态

```bash
# 查看某账户当前缓存的资源列表
redis-cli GET "auth:resource:${ACCOUNT_ID}:api_path" | python3 -m json.tool

# 查看剩余 TTL
redis-cli TTL "auth:resource:${ACCOUNT_ID}:api_path"

# 监听缓存删除事件（调试用）
redis-cli --csv PSUBSCRIBE "__keyevent@0__:del" | grep "auth:resource:"
```

---

## 维护检查清单

新增权限相关 API 或缓存变更时，检查以下项：

- [ ] 新缓存 key 是否加了常量定义（`auth/constant/cache.go`）？
- [ ] 新缓存方法是否走 `DoubleDeleteCache` 删除（`auth/service/cache_service.go`）？
- [ ] 是否添加了对应的缓存写入逻辑？
- [ ] 所有修改该数据的变更点是否调用了缓存失效？
- [ ] 失效范围是否正确（单账户 vs 批量）？
- [ ] 如果新增 cache key 格式，本文档的「缓存架构总览」和「失效矩阵」是否已更新？
- [ ] 是否存在可能导致部分失效的 N+1 查询？建议使用单次 JOIN 查询受影响账户。
