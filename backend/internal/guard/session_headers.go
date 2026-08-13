package guard

import (
	"net/http"
	"strings"
)

func CanonicalizeSessionHeaders(headers http.Header) {
	if headers == nil {
		return
	}

	var sessionID string
	for _, key := range []string{"session_id", "Session_id", "Session-Id", "Session_ID"} {
		if value := strings.TrimSpace(headers.Get(key)); value != "" {
			sessionID = value
		}
	}
	if sessionID == "" {
		return
	}

	delete(headers, "Session-Id")
	delete(headers, "Session_id")
	delete(headers, "session_id")
	delete(headers, "Session_ID")
	headers.Set("session_id", sessionID)
}

// EnsureCodexHeaders fills missing Codex correlation headers without
// overwriting IDs produced by the upstream fingerprint convergence layer.
func EnsureCodexHeaders(headers http.Header, sessionKey string) {
	if headers == nil || strings.TrimSpace(sessionKey) == "" {
		return
	}
	if strings.TrimSpace(headers.Get("X-Client-Request-Id")) == "" {
		headers.Set("X-Client-Request-Id", sessionKey)
	}
	if strings.TrimSpace(headers.Get("Thread-Id")) == "" {
		headers.Set("Thread-Id", sessionKey)
	}
	if strings.TrimSpace(headers.Get("X-Codex-Window-Id")) == "" {
		headers.Set("X-Codex-Window-Id", sessionKey+":0")
	}
}

func SyncConversationID(headers http.Header) {
	if headers == nil {
		return
	}

	sessionID := strings.TrimSpace(headers.Get("session_id"))
	if sessionID == "" {
		return
	}
	if strings.TrimSpace(headers.Get("conversation_id")) == "" &&
		strings.TrimSpace(headers.Get("Conversation_id")) == "" {
		headers.Set("conversation_id", sessionID)
	}
}

func ApplySessionGovernance(headers http.Header) {
	CanonicalizeSessionHeaders(headers)
	sessionKey := strings.TrimSpace(headers.Get("Thread-Id"))
	if sessionKey == "" {
		sessionKey = strings.TrimSpace(headers.Get("session_id"))
	}
	EnsureCodexHeaders(headers, sessionKey)
	SyncConversationID(headers)
}
