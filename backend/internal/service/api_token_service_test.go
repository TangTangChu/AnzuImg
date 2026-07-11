package service

import (
	"testing"
	"time"
)

func TestTokenExpiresAt(t *testing.T) {
	createdAt := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	got := tokenExpiresAt(createdAt, 8)
	if got == nil || !got.Equal(createdAt.Add(8*time.Hour)) {
		t.Fatalf("unexpected expiry: %v", got)
	}
	if tokenExpiresAt(createdAt, 0) != nil {
		t.Fatal("ttl=0 must keep tokens non-expiring")
	}
}
