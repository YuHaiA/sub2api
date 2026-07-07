# Upstream Sync Log

## 2026-07-07

### Absorbed upstream commit `6cea1c35`

- Source: `Wei-Shaw/sub2api`
- Title: `feat: 适配 OpenAI 新模型 gpt-5.6-sol/terra/luna`
- Scope:
  - Added OpenAI model constants and aliases for `gpt-5.6-sol`, `gpt-5.6-terra`, and `gpt-5.6-luna`
  - Synced fallback pricing and pricing resource data for the new 5.6 models
  - Updated Codex/OpenAI model transform mapping to recognize the new 5.6 variants
  - Updated frontend model whitelist and key usage modal to expose the new models
- Merge notes:
  - Resolved local conflicts by preserving existing fork-specific Codex/OpenAI mappings
  - Kept local `gpt-5.3-codex-spark-*` mappings and merged upstream `gpt-5.5-pro` / `gpt-5.6-*` support

### Absorbed upstream commit `7c2a828c`

- Source: `Wei-Shaw/sub2api`
- Title: `fix(frontend): add compact probe mode to admin account test modal`
- Scope:
  - Synced the admin account test modal so OpenAI account tests can pass `mode: compact`
  - Synced upstream test coverage for Grok default model selection and OpenAI compact probe payloads
- Merge notes:
  - Kept the local fixed `/api/v1` request path instead of importing a missing `buildApiUrl` helper
  - Preserved the existing admin modal structure while merging the upstream request body behavior

### Absorbed upstream commit `1c0ccb47`

- Source: `Wei-Shaw/sub2api`
- Title: `fix: add missing Codex CLI headers for OAuth account test`
- Scope:
  - Synced the OAuth account test path so ChatGPT/Codex test requests send the required Codex CLI headers
  - Added `OpenAI-Beta`, `Originator`, and `User-Agent` headers to the non-compact OAuth test flow
- Merge notes:
  - The upstream patch referenced a local helper that does not exist in this fork snapshot
  - Resolved by applying the equivalent headers directly and preserving the existing `chatgpt-account-id` behavior

### Absorbed upstream commit `cb151e36`

- Source: `Wei-Shaw/sub2api`
- Title: `fix: respect custom User-Agent in OAuth account test`
- Scope:
  - Synced the OAuth account test path so a configured custom OpenAI `user_agent` credential takes precedence
  - Keeps the default Codex CLI `User-Agent` only as a fallback
- Merge notes:
  - Reused the current fork's `account.GetOpenAIUserAgent()` access path
  - Preserved the already merged Codex CLI headers and `chatgpt-account-id` behavior

### Absorbed upstream commit `f881ff7c`

- Source: `Wei-Shaw/sub2api`
- Title: `fix(models): support non-v1 OpenAI models URLs`
- Scope:
  - Synced OpenAI upstream model discovery so model listing works for custom base URLs that do not end with `/v1`
  - Synced upstream unit tests covering those URL shapes
- Merge notes:
  - Applied cleanly with no local conflict
