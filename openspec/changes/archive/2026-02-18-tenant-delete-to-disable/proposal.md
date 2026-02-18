# 租户删除改为禁用

## 为什么

当前租户的删除操作使用硬删除（物理删除），不符合数据安全最佳实践。将删除操作改为禁用操作（更新状态为 suspended），可以保留数据便于审计和恢复。

## 变更内容

1. 后端：将租户 Delete 接口改为更新状态为 suspended
2. 后端：租户查询时默认排除已禁用的租户
3. 前端：将租户列表页的"删除"按钮改为"禁用"按钮
4. 前端：添加"启用"按钮用于恢复已禁用的租户

## 功能 (Capabilities)

### 新增功能
- `tenant-disable`: 租户禁用功能

### 修改功能
- `tenant-delete`: 删除改为禁用

## 影响

- **后端 API**: DELETE /api/v1/tenants/:id → 更新状态为 suspended
- **后端 Service**: internal/domain/tenant/service/tenant_service.go
- **后端 Repo**: internal/domain/tenant/repo/tenant_repo.go
- **前端页面**: ui/app/tenants/page.tsx
