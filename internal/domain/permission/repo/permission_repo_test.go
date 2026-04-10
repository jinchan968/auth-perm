package repo

import (
	"context"
	"testing"

	"auth-perm/internal/domain/permission/dm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPermissionRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&dm.AccountRoleDO{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_account_roles_account_role ON account_roles (account_id, role_id)`).Error; err != nil {
		t.Fatalf("create unique index failed: %v", err)
	}

	return db
}

func TestAssignRoleToAccount_Idempotent(t *testing.T) {
	db := newPermissionRepoTestDB(t)
	repo := NewPermissionRepo(db)
	ctx := context.Background()

	const (
		accountID = "account-1"
		roleID    = "role-1"
		tenantID  = "tenant-1"
	)

	if err := repo.AssignRoleToAccount(ctx, accountID, roleID, tenantID); err != nil {
		t.Fatalf("first assign failed: %v", err)
	}
	if err := repo.AssignRoleToAccount(ctx, accountID, roleID, tenantID); err != nil {
		t.Fatalf("second assign should be idempotent, got error: %v", err)
	}

	roleIDs, err := repo.GetAccountRoles(ctx, accountID)
	if err != nil {
		t.Fatalf("get account roles failed: %v", err)
	}
	if len(roleIDs) != 1 || roleIDs[0] != roleID {
		t.Fatalf("expected one role %q, got %#v", roleID, roleIDs)
	}
}

func TestSyncAccountRoles_ReplacesExistingRoles(t *testing.T) {
	db := newPermissionRepoTestDB(t)
	repo := NewPermissionRepo(db)
	ctx := context.Background()

	const (
		accountID = "account-1"
		tenantID  = "tenant-1"
	)

	if err := repo.AssignRoleToAccount(ctx, accountID, "role-1", tenantID); err != nil {
		t.Fatalf("seed role-1 failed: %v", err)
	}

	if err := repo.SyncAccountRoles(ctx, accountID, []string{"role-1", "role-2"}, tenantID); err != nil {
		t.Fatalf("sync add role-2 failed: %v", err)
	}

	roleIDs, err := repo.GetAccountRoles(ctx, accountID)
	if err != nil {
		t.Fatalf("get roles after first sync failed: %v", err)
	}
	if len(roleIDs) != 2 {
		t.Fatalf("expected 2 roles after first sync, got %#v", roleIDs)
	}

	if err := repo.SyncAccountRoles(ctx, accountID, []string{"role-2"}, tenantID); err != nil {
		t.Fatalf("sync remove role-1 failed: %v", err)
	}

	roleIDs, err = repo.GetAccountRoles(ctx, accountID)
	if err != nil {
		t.Fatalf("get roles after second sync failed: %v", err)
	}
	if len(roleIDs) != 1 || roleIDs[0] != "role-2" {
		t.Fatalf("expected only role-2 after second sync, got %#v", roleIDs)
	}

	if err := repo.SyncAccountRoles(ctx, accountID, []string{}, tenantID); err != nil {
		t.Fatalf("sync clear roles failed: %v", err)
	}

	roleIDs, err = repo.GetAccountRoles(ctx, accountID)
	if err != nil {
		t.Fatalf("get roles after clear failed: %v", err)
	}
	if len(roleIDs) != 0 {
		t.Fatalf("expected no roles after clear, got %#v", roleIDs)
	}
}
