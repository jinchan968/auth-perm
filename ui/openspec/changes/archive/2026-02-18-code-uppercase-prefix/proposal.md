# Code 字段大写前缀

## 为什么

当前新增权限和新增角色时，后端自动生成的 code 字段使用小写前缀，不符合项目规范。code 字段应当使用大写前缀以提高可读性和一致性。

## 变更内容

1. 修改权限创建时 code 生成逻辑，使用大写前缀
2. 修改角色创建时 code 生成逻辑，使用大写前缀

## 功能 (Capabilities)

### 新增功能
- `code-uppercase-prefix`: code 字段大写前缀功能

### 修改功能
（无）

## 影响

- **后端服务**: internal/domain/permission/service/permission_service.go
- **后端服务**: internal/domain/permission/service/role_service.go
