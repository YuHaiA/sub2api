package guard

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const MaxGPTReasoningSignatureLen = 32 * 1024 * 1024

type GPTReasoningSignatureInfo struct {
	DecodedLen    int
	CiphertextLen int
}

func SanitizeReasoning(_ string, body []byte) []byte {
	input := gjson.GetBytes(body, "input")
	if !input.Exists() || !input.IsArray() {
		return body
	}

	updated := body
	for index, item := range input.Array() {
		if strings.TrimSpace(item.Get("type").String()) != "reasoning" {
			continue
		}

		encPath := fmt.Sprintf("input.%d.encrypted_content", index)
		enc := gjson.GetBytes(updated, encPath)
		if !enc.Exists() {
			continue
		}
		if validateEncryptedContent(enc) == "" {
			continue
		}

		updated, _ = sjson.DeleteBytes(updated, encPath)
		updated, _ = sjson.SetBytes(updated, fmt.Sprintf("input.%d.summary", index), []byte("[]"))
		updated, _ = sjson.SetBytes(updated, fmt.Sprintf("input.%d.content", index), nil)
	}

	return updated
}

func validateEncryptedContent(enc gjson.Result) string {
	switch enc.Type {
	case gjson.String:
		raw := enc.String()
		if raw != strings.TrimSpace(raw) {
			return "encrypted_content has leading or trailing whitespace"
		}
		if _, err := InspectGPTReasoningSignature(raw); err != nil {
			return err.Error()
		}
		return ""
	case gjson.Null:
		return "encrypted_content is null"
	default:
		return fmt.Sprintf("encrypted_content must be a string, got %s", enc.Type.String())
	}
}

func InspectGPTReasoningSignature(raw string) (*GPTReasoningSignatureInfo, error) {
	sig := strings.TrimSpace(raw)
	if sig == "" {
		return nil, fmt.Errorf("empty GPT reasoning signature")
	}
	if len(sig) > MaxGPTReasoningSignatureLen {
		return nil, fmt.Errorf("GPT reasoning signature exceeds maximum length (%d bytes)", MaxGPTReasoningSignatureLen)
	}
	if index, r, ok := firstInvalidGPTReasoningSignatureChar(sig); ok {
		return nil, fmt.Errorf("invalid GPT reasoning signature: contains non-base64url character U+%04X at byte %d", r, index)
	}
	if !strings.HasPrefix(sig, "gAAAA") {
		return nil, fmt.Errorf("invalid GPT reasoning signature: expected gAAAA prefix")
	}

	decoded, err := decodeGPTReasoningSignature(sig)
	if err != nil {
		return nil, err
	}
	if len(decoded) < 73 {
		return nil, fmt.Errorf("invalid GPT reasoning signature: decoded payload too short (%d bytes)", len(decoded))
	}
	if decoded[0] != 0x80 {
		return nil, fmt.Errorf("invalid GPT reasoning signature: expected version 0x80, got 0x%02x", decoded[0])
	}

	ciphertextLen := len(decoded) - 57
	if ciphertextLen <= 0 || ciphertextLen%16 != 0 {
		return nil, fmt.Errorf("invalid GPT reasoning signature: ciphertext length %d is not a positive AES block multiple", ciphertextLen)
	}

	return &GPTReasoningSignatureInfo{
		DecodedLen:    len(decoded),
		CiphertextLen: ciphertextLen,
	}, nil
}

func decodeGPTReasoningSignature(sig string) ([]byte, error) {
	if decoded, err := base64.RawURLEncoding.DecodeString(sig); err == nil {
		return decoded, nil
	}
	if decoded, err := base64.URLEncoding.DecodeString(sig); err == nil {
		return decoded, nil
	}
	return nil, fmt.Errorf("invalid GPT reasoning signature: base64url decode failed")
}

func firstInvalidGPTReasoningSignatureChar(sig string) (int, rune, bool) {
	for index, r := range sig {
		switch {
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_' || r == '=':
		default:
			return index, r, true
		}
	}
	return 0, 0, false
}
