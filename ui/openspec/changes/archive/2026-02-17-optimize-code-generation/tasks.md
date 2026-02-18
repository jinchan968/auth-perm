## 1. 数据库索引优化

- [x] 1.1 检查现有索引状态，确认缺失的索引
- [x] 1.2 创建 goose 迁移脚本添加租户表 code 索引
- [x] 1.3 创建 goose 迁移脚本添加权限表 (tenant_id, code) 联合索引
- [x] 1.4 创建 goose 迁移脚本添加角色表 (tenant_id, code) 联合索引
- [x] 1.5 修复 RoleDO GORM 模型，添加正确的联合索引定义

## 2. DB 查询优化

- [x] 2.1 修改租户 repo，添加按 tenant_id 查询最大 code 的方法
- [x] 2.2 修改权限 repo，使用 tenant_id + code 前缀条件查询最大 code
- [x] 2.3 修改角色 repo，使用 tenant_id + code 前缀条件查询最大 code
- [x] 2.4 更新 code_gen 包，支持 tenant_id 参数

## 3. Redis Key 结构重构

- [x] 3.1 修改 RedisCodeGenerator，支持实体类型前缀
- [x] 3.2 实现 Hash 结构存储计数器
- [x] 3.3 更新 code_gen 包，使用新的 key 格式
- [x] 3.4 添加迁移逻辑，清理旧格式的 Redis key
