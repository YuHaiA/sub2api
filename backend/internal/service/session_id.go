package service

import (
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// maxPersistedSessionIDLength bounds the persisted client session identifier to the
// usage_logs.session_id column width (VARCHAR(255)). Longer values are rejected so
// distinct identifiers can never alias through truncation.
const maxPersistedSessionIDLength = 255

// clientSessionIDHeaders extends the OpenAI-compatible sticky-session signals with
// native protocol identifiers that are safe to persist but must not alter OpenAI
// scheduling behavior.
var clientSessionIDHeaders = append(
	append([]string{"session-id"}, explicitOpenAIHeaderSessionNames...),
	claudeCodeSessionHeader,
)

type ClientRequestCorrelation struct {
	SessionID string
	ThreadID  string
	TurnID    string
}

// ExtractClientSessionID resolves the explicit client-provided session identifier from
// request headers or Codex request metadata for usage-log correlation and returns it sanitized. It is
// protocol-agnostic and shared by every gateway handler so all supported protocols
// record session_id through one seam. Returns "" when no valid identifier is present.
//
// Optional request bodies are inspected only when no supported session header exists.
// This value feeds only usage_logs.session_id persistence. It does NOT affect sticky
// routing, account selection, request_id semantics, or upstream prompt caching, which
// keep their own (intentionally broader) session-signal resolution.
func ExtractClientSessionID(c *gin.Context, bodies ...[]byte) string {
	if c == nil || c.Request == nil {
		return ""
	}
	for _, header := range clientSessionIDHeaders {
		if sessionID := sanitizeSessionID(c.GetHeader(header)); sessionID != "" {
			return sessionID
		}
	}
	if isGrokRequestContext(c) {
		if sessionID := sanitizeSessionID(c.GetHeader(grokConversationIDHeader)); sessionID != "" {
			return sessionID
		}
	}
	if len(bodies) > 0 {
		correlation := ExtractClientRequestCorrelation(c, bodies[0])
		if correlation.ThreadID != "" {
			return correlation.ThreadID
		}
		if correlation.SessionID != "" {
			return correlation.SessionID
		}
		return correlation.TurnID
	}
	return ""
}

func ExtractClientRequestCorrelation(c *gin.Context, body []byte) ClientRequestCorrelation {
	correlation := ClientRequestCorrelation{}
	if c != nil && c.Request != nil {
		for _, header := range clientSessionIDHeaders {
			if correlation.SessionID = sanitizeSessionID(c.GetHeader(header)); correlation.SessionID != "" {
				break
			}
		}
		correlation.ThreadID = sanitizeSessionID(c.GetHeader("thread-id"))
		applyClientCorrelationMetadata(&correlation, gjson.Parse(c.GetHeader("x-codex-turn-metadata")))
	}
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return correlation
	}

	clientMetadata := gjson.GetBytes(body, "client_metadata")
	applyClientCorrelationMetadata(&correlation, clientMetadata)
	applyClientCorrelationMetadata(&correlation, parseEmbeddedClientCorrelationMetadata(clientMetadata.Get("x-codex-turn-metadata")))

	input := gjson.GetBytes(body, "input")
	if input.IsArray() {
		input.ForEach(func(_, item gjson.Result) bool {
			turnID := item.Get("internal_chat_message_metadata_passthrough.turn_id")
			if sanitized := sanitizeSessionID(turnID.String()); sanitized != "" {
				correlation.TurnID = sanitized
			}
			return true
		})
	}
	return correlation
}

func applyClientCorrelationMetadata(correlation *ClientRequestCorrelation, metadata gjson.Result) {
	if correlation == nil || !metadata.IsObject() {
		return
	}
	if correlation.SessionID == "" {
		correlation.SessionID = sanitizeSessionID(metadata.Get("session_id").String())
	}
	if correlation.ThreadID == "" {
		correlation.ThreadID = sanitizeSessionID(metadata.Get("thread_id").String())
	}
	if correlation.TurnID == "" {
		correlation.TurnID = sanitizeSessionID(metadata.Get("turn_id").String())
	}
}

func parseEmbeddedClientCorrelationMetadata(value gjson.Result) gjson.Result {
	if value.IsObject() {
		return value
	}
	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return gjson.Result{}
	}
	parsed := gjson.Parse(raw)
	if !parsed.IsObject() {
		return gjson.Result{}
	}
	return parsed
}

// sanitizeSessionID normalizes a raw client-supplied session identifier for safe
// persistence: it trims surrounding whitespace, rejects the value outright if it
// contains any control character (CR/LF/tab/NUL/…) so a log- or header-injection style
// payload cannot slip into stored correlation data, and rejects values longer than
// the DB column bound. Absent or invalid input yields "".
func sanitizeSessionID(raw string) string {
	if !utf8.ValidString(raw) {
		return ""
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	count := 0
	for _, r := range trimmed {
		if r < 0x20 || r == 0x7f {
			// An explicit correlation id never legitimately contains control
			// characters; drop the whole value rather than persist a mangled or
			// partially-injected identifier.
			return ""
		}
		count++
		if count > maxPersistedSessionIDLength {
			return ""
		}
	}
	return trimmed
}
