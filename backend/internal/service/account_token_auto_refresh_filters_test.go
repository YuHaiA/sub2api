package service

import (
	"testing"
	"time"
)

func TestFilterAccountsByTokenRefreshHealthStatus(t *testing.T) {
	resetAt := time.Now().Add(5 * time.Minute)
	accounts := []Account{
		{ID: 1, Extra: map[string]any{"health_check": map[string]any{"status": "healthy"}}},
		{ID: 2, RateLimitResetAt: &resetAt, Extra: map[string]any{"health_check": map[string]any{"status": "healthy"}}},
		{ID: 3, Extra: map[string]any{"health_check": map[string]any{"status": "banned_or_exhausted", "message": "account banned"}}},
		{ID: 4},
	}

	filtered := filterAccountsByTokenRefreshHealthStatus(accounts, "constrained")
	if len(filtered) != 1 || filtered[0].ID != 2 {
		t.Fatalf("expected only constrained account 2, got %+v", filtered)
	}

	unchecked := filterAccountsByTokenRefreshHealthStatus(accounts, "unchecked")
	if len(unchecked) != 1 || unchecked[0].ID != 4 {
		t.Fatalf("expected only unchecked account 4, got %+v", unchecked)
	}

	unavailable := filterAccountsByTokenRefreshHealthStatus(accounts, "unavailable")
	if len(unavailable) != 1 || unavailable[0].ID != 3 {
		t.Fatalf("expected only unavailable account 3, got %+v", unavailable)
	}
}
