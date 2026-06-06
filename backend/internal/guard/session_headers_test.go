package guard

import (
	"net/http"
	"testing"
)

func TestApplySessionGovernance(t *testing.T) {
	headers := make(http.Header)
	headers.Set("Session-Id", "sess-1")

	ApplySessionGovernance(headers, "cache-1")

	if got := headers.Get("session_id"); got != "sess-1" {
		t.Fatalf("unexpected session_id: %s", got)
	}
	if got := headers.Get("conversation_id"); got != "sess-1" {
		t.Fatalf("unexpected conversation_id: %s", got)
	}
	if got := headers.Get("X-Client-Request-Id"); got != "cache-1" {
		t.Fatalf("unexpected X-Client-Request-Id: %s", got)
	}
	if got := headers.Get("Thread-Id"); got != "cache-1" {
		t.Fatalf("unexpected Thread-Id: %s", got)
	}
	if got := headers.Get("X-Codex-Window-Id"); got != "cache-1:0" {
		t.Fatalf("unexpected X-Codex-Window-Id: %s", got)
	}
}
