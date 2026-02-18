## 为什么

当前租户、权限、角色的创建接口需要前端手动输入 code 字段，这种方式不够友好且容易出错。需要改为后端自动生成 code，前端只需输入业务相关的字段（如名称、描述等）。

## 变更内容

1. 后端：改造租户创建服务，使用 CodeGenerator 自动生成 code
2. 后端：改造权限创建服务，使用 CodeGenerator 自动生成 code
3. 后端：改造角色创建服务，使用 CodeGenerator 自动生成 code
4. 前端：改造租户创建表单，去掉 code 输入框
5. 前端：改造权限创建表单，去掉 code 输入框
6. 前端：改造角色创建表单，去掉 code 输入框

## 功能 (Capabilities)

### 新增功能
- auto-generate-code: 在创建实体时自动生成 code

### 修改功能
- tenant-management: 租户创建时自动生成 code
- permission-management: 权限创建时自动生成 code
- role-management: 角色创建时自动生成 code

## 影响

- 后端服务层需要注入 CodeGenerator
- 前端表单组件需要移除 code 输入框
- 后续其他实体创建也应使用此模式
