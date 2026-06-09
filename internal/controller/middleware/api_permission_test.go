package middleware

import "testing"

func TestMatchAPIPathWithPathParams(t *testing.T) {
	allowed := []string{"POST /api/v1/auth/invitations/:id/invalidate"}

	if !matchAPIPath("/api/v1/auth/invitations/abc-123/invalidate", "POST", allowed) {
		t.Fatal("expected parameterized invitation path to match")
	}
	if matchAPIPath("/api/v1/auth/invitations/abc-123/invalidate", "GET", allowed) {
		t.Fatal("expected different method not to match")
	}
	if matchAPIPath("/api/v1/auth/invitations/abc-123", "POST", allowed) {
		t.Fatal("expected partial path not to match")
	}
}

func TestMatchAPIPathKeepsLegacyPrefixMatching(t *testing.T) {
	allowed := []string{"/api/v1/journal"}

	if !matchAPIPath("/api/v1/journal/entry-1", "DELETE", allowed) {
		t.Fatal("expected legacy pure-path resource to match sub paths")
	}
}
