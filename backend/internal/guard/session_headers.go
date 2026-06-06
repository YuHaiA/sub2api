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
		if v := strings.TrimSpace(headers.Get(key)); v != "" {
			sessionID = v
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

func EnsureCodexHeaders(headers http.Header, promptCacheKey string) {
	if headers == nil {
		return
	}
	if headers.Get("X-Client-Request-Id") == "" && promptCacheKey != "" {
		headers.Set("X-Client-Request-Id", promptCacheKey)
	}
	if promptCacheKey != "" {
		headers.Set("Thread-Id", promptCacheKey)
		headers.Set("X-Codex-Window-Id", promptCacheKey+":0")
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

func ApplySessionGovernance(headers http.Header, promptCacheKey string) {
	CanonicalizeSessionHeaders(headers)
	EnsureCodexHeaders(headers, promptCacheKey)
	SyncConversationID(headers)
}
