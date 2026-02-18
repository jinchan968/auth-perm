# 租户删除作为禁用

## 为什么

当前租户禁用调用 suspend API，将状态设置为 suspended。用户希望改为调用 delete API，将状态设置为 deleted，并将 deleted 状态的租户视为禁用状态。

## 变更内容

1. 后端：修改 Delete 接口，将状态更新为 deleted（而非物理删除）
2. 前端：修改 API 调用，从 suspendTenant 改为 deleteTenant

## 功能 (Capabilities)

### 修改功能
- `tenant-delete`: 删除接口改为更新状态为 deleted

## 影响

- **后端 API**: DELETE /api/v1/tenants/:id
- **前端 API**: ui/lib/api/tenant.ts
