# Tasks: permission-ui-tenant-isolation

## 实现任务清单

### 任务 1: 修改 permissions/[id]/page.tsx 传递 tenant_id

**文件**: `app/permissions/[id]/page.tsx`

**修改内容**:
- Line 47: `getPermission(id)` → `getPermission(id, tenantId)`
- Line 91: `updatePermission(id, request)` → `updatePermission(id, tenantId, request)`
- Line 106: `deletePermission(id)` → `deletePermission(id, tenantId)`

**验收标准**: 权限详情、编辑、删除操作都正确传递 tenant_id

---

### 任务 2: 修改 permissions/roles/page.tsx 传递 tenant_id

**文件**: `app/permissions/roles/page.tsx`

**修改内容**:
- Line 97: `updateRole(editingRole.id, {...})` → `updateRole(editingRole.id, tenantId, {...})`
- Line 123: `deleteRole(id)` → `deleteRole(id, tenantId)`

**验收标准**: 角色编辑、删除操作都正确传递 tenant_id

---

### 任务 3: 修改 permissions/roles/[id]/page.tsx 传递 tenant_id

**文件**: `app/permissions/roles/[id]/page.tsx`

**修改内容**:
- Line 38: `getRole(roleId)` → `getRole(roleId, tenantId)`
- Line 40: `getRolePermissions(roleId)` → `getRolePermissions(roleId, tenantId)`

**验收标准**: 角色详情和角色权限列表加载时正确传递 tenant_id

---

## 已验证正确的页面（无需修改）

| 页面 | API 函数 | 状态 |
|------|---------|------|
| `app/permissions/page.tsx` | listPermissions | ✅ 已传递 tenant_id |
| `app/permissions/new/page.tsx` | createPermission | ✅ 已传递 tenant_id |
| `app/permissions/roles/page.tsx` | listRoles, createRole | ✅ 已传递 tenant_id |
| `app/permissions/roles/[id]/page.tsx` | listPermissions, assignPermissionsToRole | ✅ 已传递 tenant_id |
| `app/permissions/users/page.tsx` | listRoles, assignRoleToAccount | ✅ 已传递 tenant_id |

## 任务状态

- [ ] 任务 1: 修改 permissions/[id]/page.tsx
- [ ] 任务 2: 修改 permissions/roles/page.tsx
- [ ] 任务 3: 修改 permissions/roles/[id]/page.tsx
