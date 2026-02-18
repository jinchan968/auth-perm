## 新增需求

### 需求:Redis Key按类型隔离
系统必须使用不同前缀区分不同实体的Code生成器。

#### 场景:租户Code生成使用独立Key
- **当** 生成租户Code时
- **那么** Redis key 格式为 `code_gen:tenant`

#### 场景:权限Code生成按租户隔离
- **当** 为特定租户生成权限Code时
- **那么** Redis key 格式为 `code_gen:permission:{tenant_id}`

#### 场景:角色Code生成按租户隔离
- **当** 为特定租户生成角色Code时
- **那么** Redis key 格式为 `code_gen:role:{tenant_id}`

### 需求:支持Hash结构存储
系统可以使用Hash结构存储Code生成器的计数器。

#### 场景:使用Hash存储租户计数器
- **当** 存储租户Code计数器时
- **那么** 使用 HINCRBY code_gen:tenant counter 1

#### 场景:使用Hash存储权限计数器
- **当** 存储权限Code计数器时
- **那么** 使用 HINCRBY code_gen:permission:{tenant_id} counter 1

#### 场景:使用Hash存储角色计数器
- **当** 存储角色Code计数器时
- **那么** 使用 HINCRBY code_gen:role:{tenant_id} counter 1
