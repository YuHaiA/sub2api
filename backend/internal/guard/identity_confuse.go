package guard

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func ConfuseKey(accountID int64, kind string, value string) string {
	name := fmt.Sprintf(
		"sub2api:identity-confuse:%s:account_%d:%s",
		strings.TrimSpace(kind),
		accountID,
		strings.TrimSpace(value),
	)
	hash := sha256.Sum256([]byte(name))
	return fmt.Sprintf("%x", hash[:16])
}

// ConfuseCodexMetadataLight applies low-risk per-account obfuscation to
// Codex-specific metadata fields without changing prompt_cache_key or requiring
// response restoration.
func ConfuseCodexMetadataLight(body []byte, accountID int64) []byte {
	if len(body) == 0 || accountID <= 0 {
		return body
	}

	updated := body
	if instID := strings.TrimSpace(gjson.GetBytes(updated, "client_metadata.x-codex-installation-id").String()); instID != "" {
		obfuscated := ConfuseKey(accountID, "installation", instID)
		updated, _ = sjson.SetBytes(updated, "client_metadata.x-codex-installation-id", obfuscated)
	}

	if rawMeta := strings.TrimSpace(gjson.GetBytes(updated, "client_metadata.x-codex-turn-metadata").String()); rawMeta != "" {
		updatedMeta := rawMeta
		if turnID := strings.TrimSpace(gjson.Get(rawMeta, "turn_id").String()); turnID != "" {
			obfuscated := ConfuseKey(accountID, "turn", turnID)
			updatedMeta, _ = sjson.Set(updatedMeta, "turn_id", obfuscated)
		}
		if updatedMeta != rawMeta {
			updated, _ = sjson.SetBytes(updated, "client_metadata.x-codex-turn-metadata", updatedMeta)
		}
	}

	return updated
}
