## 上下文

当前代码生成功能存在以下问题：
1. 角色表的索引配置错误 - `idx_tenant_code` 唯一索引只包含 Code 字段，不包含 TenantID
2. DB 查询使用 LIKE 前缀匹配效率低下
3. Redis key 命名缺乏区分 - 所有实体共用 `code_gen:{prefix}` 前缀

## 目标 / 非目标

**目标：**
1. 修复角色表索引，创建 (tenant_id, code) 联合唯一索引
2. 优化 DB 查询使用 max(code) + where 条件替代 LIKE
3. 重构 Redis key 结构，支持按实体类型和租户隔离

**非目标：**
- 不修改现有 code 格式规范
- 不修改查询接口的返回格式

## 决策

### 1. 数据库索引
- **方案 A**: 修改 GORM 模型定义，添加联合唯一索引
  - Tenant: 已有 uniqueIndex 在 Code 上 ✓
  - Permission: 已有 uniqueIndex:idx_tenant_code 包含 tenant_id 和 code ✓
  - Role: 需要修复为联合索引
- **方案 B**: 使用 goose 创建迁移脚本添加索引
  - 选择方案 B，更清晰且易于管理

### 2. DB 查询优化
- **方案 A**: 使用 `SELECT MAX(code) FROM table WHERE tenant_id = ? AND code LIKE 'P%'`
  - 优点：利用索引
- **方案 B**: 使用 `SELECT code FROM table WHERE tenant_id = ? AND code LIKE 'P%' ORDER BY code DESC LIMIT 1`
  - 优点：更直观
  - 选择方案 B，与现有实现一致，但优化为使用 tenant_id 条件

### 3. Redis Key 结构
- **方案 A**: 使用 Hash 结构
  - key: `code_gen:permission:{tenant_id}`
  - field: `counter`
- **方案 B**: 使用 String 结构，key 包含 tenant_id
  - key: `code_gen:P:{tenant_id}`
  - 选择方案 A，Hash 结构更灵活，支持存储更多元数据

## 风险 / 权衡

- [风险] Redis key 变更后，旧 key 失效 → 迁移期间重新从 DB 初始化
- [风险] 索引添加可能锁表 → 使用 CONCURRENTLY 选项在线添加

## 迁移计划

1. 执行 goose 迁移添加索引
2. 更新 code_gen 包使用 Hash 结构
3. 更新 repo 查询使用 tenant_id 条件
4. 清理旧格式的 Redis key
