package guard

import (
	"encoding/base64"
	"testing"

	"github.com/tidwall/gjson"
)

func TestSanitizeReasoningRemovesInvalidEncryptedContent(t *testing.T) {
	body := []byte(`{"input":[{"type":"reasoning","encrypted_content":"bad","summary":[{"type":"summary_text","text":"x"}],"content":"x"},{"type":"input_text","text":"hello"}]}`)
	got := SanitizeReasoning(body)

	if gjson.GetBytes(got, "input.0.encrypted_content").Exists() {
		t.Fatal("expected encrypted_content removed")
	}
	if gjson.GetBytes(got, "input.0.summary").Raw != "[]" {
		t.Fatalf("expected summary reset, got %s", gjson.GetBytes(got, "input.0.summary").Raw)
	}
	if gjson.GetBytes(got, "input.1.text").String() != "hello" {
		t.Fatal("unexpected non-reasoning mutation")
	}
}

func TestSanitizeReasoningPreservesValidEncryptedContent(t *testing.T) {
	payload := make([]byte, 73)
	payload[0] = 0x80
	signature := base64.RawURLEncoding.EncodeToString(payload)
	body := []byte(`{"input":[{"type":"reasoning","encrypted_content":"` + signature + `","summary":[]}]}`)

	got := SanitizeReasoning(body)
	if gjson.GetBytes(got, "input.0.encrypted_content").String() != signature {
		t.Fatal("expected valid encrypted_content preserved")
	}
}

func TestInspectGPTReasoningSignatureRejectsInvalidPrefix(t *testing.T) {
	if _, err := InspectGPTReasoningSignature("abc"); err == nil {
		t.Fatal("expected invalid signature error")
	}
}
