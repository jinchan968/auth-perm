## 为什么

当前代码生成功能存在以下问题：
1. 数据库索引不完整 - 租户、权限、角色的 code 字段缺少适当的索引，影响查询性能
2. DB 查询使用 LIKE 前缀匹配效率低下 - 应改用 max(code) + where 条件
3. Redis key 命名缺乏区分 - 不同实体共用前缀，无法按租户隔离

## 变更内容

1. **数据库索引优化**：检查并添加缺失的索引
   - 租户表：code 字段索引
   - 权限表：tenant_id + code 联合索引
   - 角色表：tenant_id + code 联合索引
2. **DB 查询优化**：改用 max(code) + where tenant_id 条件替代 LIKE 前缀查询
3. **Redis Key 结构优化**：使用 Hash 结构，按实体类型和租户ID区分

## 功能 (Capabilities)

### 新增功能
- `code-index-optimization`: 数据库索引优化和查询改进
- `redis-key-refactor`: Redis key 结构重构，支持租户隔离

### 修改功能
- 无

## 影响

- 数据库：新增索引脚本（goose SQL）
- 后端代码：修改 code_gen 包、repo 查询方法
- Redis：key 命名规范变更（需要清理旧 key）
