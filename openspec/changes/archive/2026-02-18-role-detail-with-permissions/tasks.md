# 角色详情接口优化 - 任务清单

## 1. DTO 修改

- [x] 1.1 在 PermissionDTO 中添加 `is_selected` 字段

## 2. Service 层修改

- [x] 2.1 修改 GetRole 方法，查询租户全量权限列表
- [x] 2.2 在 GetRole 方法中标记已关联权限的 is_selected 状态

## 3. 验证

- [x] 3.1 编译验证
- [ ] 3.2 接口测试验证
