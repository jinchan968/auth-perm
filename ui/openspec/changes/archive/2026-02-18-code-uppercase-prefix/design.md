# Code 大写前缀 - 技术设计

## 上下文

修改权限和角色创建时的 code 生成逻辑，使其使用大写前缀。

## 目标 / 非目标

**目标：**
- 权限创建时 code 使用大写前缀
- 角色创建时 code 使用大写前缀

**非目标：**
- 不修改现有权限/角色的 code
- 不修改其他业务逻辑

## 决策

### 实现方案
- 在生成 code 时使用 strings.ToUpper() 转换
- 保持原有的格式规则不变（如下划线分隔）

### 受影响的文件
- internal/domain/permission/service/permission_service.go
- internal/domain/permission/service/role_service.go

## 风险 / 权衡

- 兼容性：仅影响新创建的资源，不影响现有数据
