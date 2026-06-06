package guard

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestSanitizeReasoning_RemovesInvalidEncryptedContent(t *testing.T) {
	body := []byte(`{"input":[{"type":"reasoning","encrypted_content":"bad","summary":[{"type":"summary_text","text":"x"}],"content":"x"},{"type":"input_text","text":"hello"}]}`)
	got := SanitizeReasoning("codex", body)

	if gjson.GetBytes(got, "input.0.encrypted_content").Exists() {
		t.Fatalf("expected encrypted_content removed")
	}
	if gjson.GetBytes(got, "input.0.summary").Raw != "[]" {
		t.Fatalf("expected summary reset, got %s", gjson.GetBytes(got, "input.0.summary").Raw)
	}
	if gjson.GetBytes(got, "input.1.text").String() != "hello" {
		t.Fatalf("unexpected non-reasoning mutation")
	}
}

func TestInspectGPTReasoningSignature_RejectsInvalidPrefix(t *testing.T) {
	if _, err := InspectGPTReasoningSignature("abc"); err == nil {
		t.Fatalf("expected invalid signature error")
	}
}
