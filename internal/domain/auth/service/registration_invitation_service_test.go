package service

import "testing"

func TestBuildInviteURLUsesOriginOnly(t *testing.T) {
	got := buildInviteURL("https://example.com/admin/invitations?tab=active", "invite-code")
	want := "https://example.com/register?invite_code=invite-code"
	if got != want {
		t.Fatalf("buildInviteURL() = %q, want %q", got, want)
	}
}

func TestBuildInviteURLFallsBackToRelativePath(t *testing.T) {
	got := buildInviteURL("/admin/invitations", "invite-code")
	want := "/register?invite_code=invite-code"
	if got != want {
		t.Fatalf("buildInviteURL() = %q, want %q", got, want)
	}
}

func TestResolveInvitationTenantScopeRejectsCrossTenantForRegularUser(t *testing.T) {
	_, err := resolveInvitationTenantScope("tenant-b", "tenant-a", false)
	if err == nil {
		t.Fatal("expected cross-tenant operation to be rejected")
	}
}

func TestResolveInvitationTenantScopeAllowsSuperAdminRequestedTenant(t *testing.T) {
	got, err := resolveInvitationTenantScope("tenant-b", "tenant-a", true)
	if err != nil {
		t.Fatalf("resolveInvitationTenantScope() returned error: %v", err)
	}
	if got != "tenant-b" {
		t.Fatalf("resolveInvitationTenantScope() = %q, want %q", got, "tenant-b")
	}
}
