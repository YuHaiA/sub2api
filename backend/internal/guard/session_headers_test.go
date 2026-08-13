package guard

import (
	"net/http"
	"testing"
)

func TestApplySessionGovernance(t *testing.T) {
	headers := make(http.Header)
	headers.Set("Session-Id", "sess-1")

	ApplySessionGovernance(headers)

	for name, want := range map[string]string{
		"session_id":          "sess-1",
		"conversation_id":     "sess-1",
		"X-Client-Request-Id": "sess-1",
		"Thread-Id":           "sess-1",
		"X-Codex-Window-Id":   "sess-1:0",
	} {
		if got := headers.Get(name); got != want {
			t.Fatalf("unexpected %s: got %q want %q", name, got, want)
		}
	}
}

func TestApplySessionGovernancePreservesConvergedHeaders(t *testing.T) {
	headers := make(http.Header)
	headers.Set("session_id", "isolated-session")
	headers.Set("conversation_id", "isolated-conversation")
	headers.Set("X-Client-Request-Id", "converged-request")
	headers.Set("Thread-Id", "converged-thread")
	headers.Set("X-Codex-Window-Id", "converged-window")

	ApplySessionGovernance(headers)

	for name, want := range map[string]string{
		"conversation_id":     "isolated-conversation",
		"X-Client-Request-Id": "converged-request",
		"Thread-Id":           "converged-thread",
		"X-Codex-Window-Id":   "converged-window",
	} {
		if got := headers.Get(name); got != want {
			t.Fatalf("unexpected %s: got %q want %q", name, got, want)
		}
	}
}
