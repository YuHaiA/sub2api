//go:build unit

package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newSessionHeaderContext(t *testing.T, headers map[string]string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	c.Request = req
	return c
}

func TestSanitizeSessionID(t *testing.T) {
	longRunes := strings.Repeat("a", maxPersistedSessionIDLength+50)
	multibyte := strings.Repeat("好", maxPersistedSessionIDLength+10)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"whitespace only", "   \t  ", ""},
		{"trims surrounding whitespace", "  sess-123  ", "sess-123"},
		{"plain value", "conv_abc-123.XYZ", "conv_abc-123.XYZ"},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440000"},
		{"reject CR", "sess\r123", ""},
		{"reject LF", "sess\n123", ""},
		{"reject CRLF injection", "sess-1\r\nSet-Cookie: x=y", ""},
		{"reject tab inside", "sess\t123", ""},
		{"reject NUL", "sess\x00123", ""},
		{"reject DEL", "sess\x7f123", ""},
		{"reject invalid UTF-8", string([]byte{'s', 'e', 's', 's', '-', 0xff}), ""},
		{"accepts value at column bound", strings.Repeat("b", maxPersistedSessionIDLength), strings.Repeat("b", maxPersistedSessionIDLength)},
		{"rejects overlong value", longRunes, ""},
		{"rejects overlong multibyte value", multibyte, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeSessionID(tc.in)
			require.Equal(t, tc.want, got, "sanitizeSessionID(%q)", tc.in)
			// Sanitized output must never exceed the DB column bound (rune-counted).
			require.LessOrEqual(t, len([]rune(got)), maxPersistedSessionIDLength)
		})
	}
}

func TestExtractClientSessionID_NilContext(t *testing.T) {
	require.Equal(t, "", ExtractClientSessionID(nil))
}

func TestExtractClientSessionID_NilRequest(t *testing.T) {
	require.Equal(t, "", ExtractClientSessionID(&gin.Context{}))
}

func TestExtractClientSessionID_AbsentReturnsEmpty(t *testing.T) {
	c := newSessionHeaderContext(t, nil)
	require.Equal(t, "", ExtractClientSessionID(c))
}

func TestExtractClientSessionID_SupportedHeaders(t *testing.T) {
	tests := []struct {
		name   string
		header string
		value  string
	}{
		{"session_id", "session_id", "sess-A"},
		{"conversation_id", "conversation_id", "conv-B"},
		{"X-Session-Affinity", openCodeSessionAffinityHeader, "aff-C"},
		{"X-Session-Id", openCodeSessionIDHeader, "sid-D"},
		{"X-OpenCode-Session", openCodeNativeSessionHeader, "oc-E"},
		{"X-Conversation-ID", codeBuddyConversationHeader, "cb-F"},
		{"X-Claude-Code-Session-Id", claudeCodeSessionHeader, "cc-G"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newSessionHeaderContext(t, map[string]string{tc.header: tc.value})
			require.Equal(t, tc.value, ExtractClientSessionID(c))
		})
	}
}

func TestExtractClientSessionID_HeaderPrecedence(t *testing.T) {
	// session_id ranks ahead of conversation_id and the X-* variants.
	c := newSessionHeaderContext(t, map[string]string{
		"session_id":                "primary",
		"conversation_id":           "secondary",
		openCodeSessionIDHeader:     "tertiary",
		codeBuddyConversationHeader: "quaternary",
	})
	require.Equal(t, "primary", ExtractClientSessionID(c))
}

func TestExtractClientSessionID_Sanitizes(t *testing.T) {
	c := newSessionHeaderContext(t, map[string]string{openCodeSessionIDHeader: "  clean-123  "})
	require.Equal(t, "clean-123", ExtractClientSessionID(c))
}

func TestExtractClientSessionID_IgnoresNonSessionHeaders(t *testing.T) {
	// prompt_cache_key, request/message ids, and a Grok conversation header on a
	// non-Grok request are NOT persisted as session_id.
	c := newSessionHeaderContext(t, map[string]string{
		"prompt_cache_key": "cache-key-should-not-persist",
		"X-Request-Id":     "req-should-not-persist",
		"x-grok-conv-id":   "grok-conv-should-not-persist",
	})
	require.Equal(t, "", ExtractClientSessionID(c))
}

func TestExtractClientSessionID_GrokConversationHeader(t *testing.T) {
	c := newSessionHeaderContext(t, map[string]string{
		grokConversationIDHeader: "grok-native-session",
	})
	c.Set("api_key", &APIKey{
		ID:    42,
		Group: &Group{Platform: PlatformGrok},
	})

	require.Equal(t, "grok-native-session", ExtractClientSessionID(c))
}

func TestExtractClientSessionID_GrokConversationHeaderForCompositeRoute(t *testing.T) {
	c := newSessionHeaderContext(t, map[string]string{
		grokConversationIDHeader: "grok-composite-session",
	})
	c.Set("api_key", &APIKey{
		ID:    43,
		Group: &Group{Platform: PlatformComposite},
	})
	c.Request = c.Request.WithContext(WithResolvedTargetPlatform(context.Background(), PlatformGrok))

	require.Equal(t, "grok-composite-session", ExtractClientSessionID(c))
}

func TestExtractClientSessionID_InjectionHeaderDropped(t *testing.T) {
	// A supported header carrying a CRLF payload is rejected, not persisted mangled.
	c := newSessionHeaderContext(t, map[string]string{"session_id": "abc"})
	c.Request.Header.Set("session_id", "abc\r\nX-Injected: 1")
	require.Equal(t, "", ExtractClientSessionID(c))
}

func TestExtractClientRequestCorrelationFromCodexBody(t *testing.T) {
	c := newSessionHeaderContext(t, nil)
	body := []byte(`{
		"client_metadata": {
			"session_id": "session-body",
			"thread_id": "01a00f7a-1ea9-7d71-a006-141f6e7e3555",
			"turn_id": "turn-direct",
			"x-codex-turn-metadata": "{\"session_id\":\"session-embedded\",\"thread_id\":\"thread-embedded\",\"turn_id\":\"turn-embedded\"}"
		},
		"input": [
			{"type":"message","internal_chat_message_metadata_passthrough":{"turn_id":"01a00f7a-20e3-7491-91c5-f2abafa3b212"}}
		]
	}`)

	correlation := ExtractClientRequestCorrelation(c, body)
	require.Equal(t, "session-body", correlation.SessionID)
	require.Equal(t, "01a00f7a-1ea9-7d71-a006-141f6e7e3555", correlation.ThreadID)
	require.Equal(t, "01a00f7a-20e3-7491-91c5-f2abafa3b212", correlation.TurnID)
	require.Equal(t, correlation.ThreadID, ExtractClientSessionID(c, body))
}

func TestExtractClientSessionIDHeaderPrecedesCodexBody(t *testing.T) {
	c := newSessionHeaderContext(t, map[string]string{"session-id": "header-session"})
	body := []byte(`{"client_metadata":{"thread_id":"body-thread"}}`)
	require.Equal(t, "header-session", ExtractClientSessionID(c, body))
}

func TestExtractClientRequestCorrelationFromCodexTurnMetadataHeader(t *testing.T) {
	c := newSessionHeaderContext(t, map[string]string{
		"x-codex-turn-metadata": `{"session_id":"header-session","thread_id":"01a00f7a-1ea9-7d71-a006-141f6e7e3555","turn_id":"01a00f7a-20e3-7491-91c5-f2abafa3b212"}`,
	})

	correlation := ExtractClientRequestCorrelation(c, nil)
	require.Equal(t, "header-session", correlation.SessionID)
	require.Equal(t, "01a00f7a-1ea9-7d71-a006-141f6e7e3555", correlation.ThreadID)
	require.Equal(t, "01a00f7a-20e3-7491-91c5-f2abafa3b212", correlation.TurnID)
	require.Equal(t, correlation.ThreadID, ExtractClientSessionID(c, []byte(`{}`)))
}

func TestExtractClientRequestCorrelationRejectsInvalidMetadataIDs(t *testing.T) {
	c := newSessionHeaderContext(t, nil)
	body := []byte(`{"client_metadata":{"thread_id":"bad\nthread","x-codex-turn-metadata":"not-json"}}`)

	correlation := ExtractClientRequestCorrelation(c, body)
	require.Empty(t, correlation.SessionID)
	require.Empty(t, correlation.ThreadID)
	require.Empty(t, correlation.TurnID)
	require.Empty(t, ExtractClientSessionID(c, body))
}
