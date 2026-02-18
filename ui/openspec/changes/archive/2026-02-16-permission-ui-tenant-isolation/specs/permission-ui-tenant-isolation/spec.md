# Permission UI Tenant Isolation Spec

## Overview

This specification defines the implementation details for adding tenant_id to all API calls in the permissions management pages.

## Current State Analysis

After analyzing the codebase, the following pages have been identified:

### Already Implemented (tenant_id correctly passed)

| Page | API Function | Status |
|------|-------------|--------|
| `app/permissions/page.tsx` | listPermissions | PASSES tenant_id |
| `app/permissions/new/page.tsx` | createPermission | PASSES tenant_id |
| `app/permissions/roles/page.tsx` | listRoles, createRole | PASSES tenant_id |
| `app/permissions/roles/[id]/page.tsx` | listPermissions, assignPermissionsToRole | PASSES tenant_id |
| `app/permissions/users/page.tsx` | listRoles, assignRoleToAccount | PASSES tenant_id |

### Needs Implementation (tenant_id NOT passed)

| Page | API Function | Current Call | Required Call |
|------|-------------|--------------|---------------|
| `app/permissions/[id]/page.tsx` | getPermission | `getPermission(id)` | `getPermission(id, tenantId)` |
| `app/permissions/[id]/page.tsx` | updatePermission | `updatePermission(id, request)` | `updatePermission(id, tenantId, request)` |
| `app/permissions/[id]/page.tsx` | deletePermission | `deletePermission(id)` | `deletePermission(id, tenantId)` |
| `app/permissions/roles/page.tsx` | updateRole | `updateRole(id, {...})` | `updateRole(id, tenantId, {...})` |
| `app/permissions/roles/page.tsx` | deleteRole | `deleteRole(id)` | `deleteRole(id, tenantId)` |
| `app/permissions/roles/[id]/page.tsx` | getRole | `getRole(roleId)` | `getRole(roleId, tenantId)` |
| `app/permissions/roles/[id]/page.tsx` | getRolePermissions | `getRolePermissions(roleId)` | `getRolePermissions(roleId, tenantId)` |

## Implementation Details

### 1. app/permissions/[id]/page.tsx

**tenantId variable exists**: Line 27 - `const tenantId = user?.tenant_id || ''`

**Changes needed:**

1. **Line 47** - `getPermission(id)`:
   ```typescript
   // Current (WRONG):
   const data = await getPermission(id)

   // Fixed:
   const data = await getPermission(id, tenantId)
   ```

2. **Line 91** - `updatePermission(id, request)`:
   ```typescript
   // Current (WRONG):
   const updated = await updatePermission(id, request)

   // Fixed:
   const updated = await updatePermission(id, tenantId, request)
   ```

3. **Line 106** - `deletePermission(id)`:
   ```typescript
   // Current (WRONG):
   await deletePermission(id)

   // Fixed:
   await deletePermission(id, tenantId)
   ```

### 2. app/permissions/roles/page.tsx

**tenantId variable exists**: Line 38 - `const tenantId = user?.tenant_id || ''`

**Changes needed:**

1. **Line 97** - `updateRole(editingRole.id, {...})`:
   ```typescript
   // Current (WRONG):
   await updateRole(editingRole.id, {
     id: editingRole.id,
     name: formData.name,
     description: formData.description,
   })

   // Fixed:
   await updateRole(editingRole.id, tenantId, {
     id: editingRole.id,
     name: formData.name,
     description: formData.description,
   })
   ```

2. **Line 123** - `deleteRole(id)`:
   ```typescript
   // Current (WRONG):
   await deleteRole(id)

   // Fixed:
   await deleteRole(id, tenantId)
   ```

### 3. app/permissions/roles/[id]/page.tsx

**tenantId variable exists**: Line 31 - `const tenantId = user?.tenant_id || ''`

**Changes needed:**

1. **Line 38** - `getRole(roleId)`:
   ```typescript
   // Current (WRONG):
   getRole(roleId)

   // Fixed:
   getRole(roleId, tenantId)
   ```

2. **Line 40** - `getRolePermissions(roleId)`:
   ```typescript
   // Current (WRONG):
   getRolePermissions(roleId)

   // Fixed:
   getRolePermissions(roleId, tenantId)
   ```

### 4. app/permissions/users/page.tsx

**tenantId variable exists**: Line 33 - `const tenantId = currentUser?.tenant_id || ''`

**Analysis:**
- `getUserRoles(currentUser.id)` - Line 61: The API function `getUserRoles` does NOT support tenant_id parameter (see `/lib/api/role.ts` line 100). No change required for this call.
- `removeRoleFromAccount` function is not used in this page, so no changes needed.

## API Function Signatures (from lib/api/)

### permission.ts
```typescript
getPermission(id: string, tenantId: string): Promise<Permission>
updatePermission(id: string, tenantId: string, request: UpdatePermissionRequest): Promise<Permission>
deletePermission(id: string, tenantId: string): Promise<void>
```

### role.ts
```typescript
getRole(id: string, tenantId: string): Promise<Role>
updateRole(id: string, tenantId: string, request: UpdateRoleRequest): Promise<Role>
deleteRole(id: string, tenantId: string): Promise<void>
getRolePermissions(roleId: string, tenantId: string): Promise<PermissionListItem[]>
getUserRoles(accountId: string): Promise<RoleListItem[]>  // Note: no tenant_id support
removeRoleFromAccount(accountId: string, roleId: string, tenantId: string): Promise<void>
```

## Tenant ID Source

All pages use `useAuthStore` to get the current user's tenant_id:

```typescript
import { useAuthStore } from '@/store/auth-store'

// Inside component:
const { user } = useAuthStore()
const tenantId = user?.tenant_id || ''
```

The tenantId variable is already defined in all pages, so no changes needed for the tenant ID source.

## Edge Cases

1. **Empty tenant_id**: If `user?.tenant_id` is empty/falsy, the code should not proceed with API calls. All pages already handle this by checking `if (!tenantId) return` or similar guards.

2. **Loading state**: While data is being fetched with tenant_id, loading states should be displayed.

3. **Error handling**: API errors should be caught and displayed to the user.

## Testing Checklist

- [ ] Verify permissions page loads and displays permissions for the correct tenant
- [ ] Verify permission detail page loads correct permission data
- [ ] Verify permission update works with tenant isolation
- [ ] Verify permission deletion works with tenant isolation
- [ ] Verify roles page loads and displays roles for the correct tenant
- [ ] Verify role update works with tenant isolation
- [ ] Verify role deletion works with tenant isolation
- [ ] Verify role permissions page loads correct role and permissions
- [ ] Verify assigning permissions to role works with tenant isolation
- [ ] Verify user roles page loads roles for the correct tenant
- [ ] Verify assigning roles to user works with tenant isolation
