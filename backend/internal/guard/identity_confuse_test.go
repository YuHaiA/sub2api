package guard

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestConfuseCodexMetadataLight(t *testing.T) {
	body := []byte(`{"client_metadata":{"x-codex-installation-id":"dev-123","x-codex-turn-metadata":"{\"turn_id\":\"turn-1\",\"window_id\":\"w-1\"}"}}`)

	got := ConfuseCodexMetadataLight(body, 42)

	inst := gjson.GetBytes(got, "client_metadata.x-codex-installation-id").String()
	if inst == "" || inst == "dev-123" {
		t.Fatalf("expected installation id obfuscated, got %q", inst)
	}

	turnMeta := gjson.GetBytes(got, "client_metadata.x-codex-turn-metadata").String()
	if gjson.Get(turnMeta, "turn_id").String() == "turn-1" {
		t.Fatalf("expected turn_id obfuscated")
	}
	if gjson.Get(turnMeta, "window_id").String() != "w-1" {
		t.Fatalf("expected window_id preserved")
	}
}

func TestConfuseKeyStablePerAccount(t *testing.T) {
	a := ConfuseKey(1, "turn", "abc")
	b := ConfuseKey(1, "turn", "abc")
	c := ConfuseKey(2, "turn", "abc")
	if a != b {
		t.Fatalf("expected stable obfuscation for same account")
	}
	if a == c {
		t.Fatalf("expected different accounts to produce different values")
	}
}
