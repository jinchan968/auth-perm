# 角色详情接口优化

## 为什么

当前角色详情接口（GET /api/v1/permissions/roles/:id）在返回时出现错误，原因是查询角色与权限关联关系时返回空结果。现在需要优化该接口，在返回角色详情时同时返回：
1. 租户下全量的权限列表
2. 角色已关联的权限
3. 由后端标记已关联的状态，前端直接根据状态展示勾选

## 变更内容

1. 修改 `GetRole` 方法，在返回角色详情时同时返回权限信息
2. 新增返回字段：全量权限列表，每个权限包含 `is_selected` 标记表示是否已关联
3. 保持向后兼容，现有返回结构不变，仅增加字段

## 功能 (Capabilities)

### 新增功能
- `role-permission-detail`: 角色详情接口优化，返回全量权限列表及关联状态

## 影响

- **后端**: 修改 `permission_service.go` 的 `GetRole` 方法和 `RoleDTO`
- **API**: GET /api/v1/permissions/roles/:id 返回结构变更
