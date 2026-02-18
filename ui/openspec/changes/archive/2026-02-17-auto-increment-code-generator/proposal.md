## 为什么

在多租户系统中，租户、权限、角色都需要 code 字段来唯一标识。当前没有统一的方法来生成自增的 code，需要手动指定或者使用 UUID，不够友好。需要一个公共方法来自动生成格式为"前缀+6位数字"的自增码。

## 变更内容

1. 创建公共的 Code 生成器服务
2. 支持 Redis 分布式锁（如果 Redis 不可用则降级到内存缓存）
3. 提供通用方法：`GenerateCode(currentMaxCode, prefix)` 返回下一个自增码

## 功能 (Capabilities)

### 新增功能
- auto-increment-code: 自动生成自增 code 的公共方法

### 修改功能
- 无

## 影响

- 新增服务：`internal/infra/code_gen/`
- 可被租户、权限、角色等服务复用
