# 租户管理 Design

## 上下文

当前系统已经通过 `accounts` 表的 `tenant_id` 字段支持多租户，但：
1. 没有独立的租户表来存储租户的配置信息
2. 缺乏租户级别的管理功能（创建、配置、限制）
3. 租户信息分散在各个表中，缺乏统一管理

需要建立独立的租户领域，提供完整的租户管理能力。

## 目标 / 非目标

**目标：**
- 创建租户领域，包含租户实体
- 实现租户 CRUD 操作
- 实现租户配置管理（功能开关、配额限制）
- 支持租户管理员

**非目标：**
- 计费系统（后期可能扩展）
- 多租户数据隔离的具体实现细节
- 租户迁移工具

## 决策

### 1. 租户表设计
**决策**：创建独立的 `tenants` 表存储租户核心信息

```go
// TenantDO 租户领域对象
type TenantDO struct {
    ID          string    // 租户ID (UUID)
    Name        string    // 租户名称
    Code        string    // 租户代码（唯一）
    Status      string    // 租户状态：active, suspended, deleted
    Plan        string    // 套餐：free, basic, pro, enterprise
    ExpireAt    *time.Time // 过期时间
    Settings    dto.TenantSettings      // 租户设置
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**理由**：
- 独立表便于扩展租户级别的配置
- 与 accounts 表的 tenant_id 关联

### 2. 租户配置设计
**决策**：使用 JSON 字段存储租户配置

**理由**：
- 避免频繁修改表结构
- 不同套餐可以有不同配置项

### 3. 领域分层
**决策**：采用与其他领域相同的分层结构
- dm: 领域模型
- dto: 数据传输对象
- param: 参数定义
- repo: 数据仓储
- service: 业务服务
- handler: HTTP 处理

**理由**：保持与现有代码风格一致

### 4. 依赖关系
**决策**：租户领域作为基础领域，不依赖其他领域

**理由**：
- 租户是最基础的实体
- 其他领域（auth, permission）依赖租户

## 风险 / 权衡

1. **数据一致性** → 初始阶段通过外键约束保证 accounts.tenant_id 与 tenants.id 的关联
2. **配置灵活性** → 使用 JSON 字段可能导致查询不便，但可以通过冗余字段优化
3. **套餐扩展性** → 当前设计支持后期扩展计费功能

## 待定事项

1. 租户配额的具体字段设计（用户数、API调用次数等）: 这些字段暂时指定，但是不做业务处理
2. 是否需要租户级别的角色继承：不需要
3. dto.TenantSettings 对象具体字段需要进一步定义，并且实现「Scan,Value」方法来pg数据库的 JSON 字段存储和读取
