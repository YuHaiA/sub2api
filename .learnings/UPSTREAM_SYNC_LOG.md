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

### Absorbed upstream batch `2026-07-07-a`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `87dd5f5d` `fix(openai): 切组后剥离失配的 previous_response_id，修复跨组会话鉴权失败`
  - `9ecfc4e` `docs: add Sub2API admin skill`
  - `029b6d61` `feat(usage): 聚合统计拆分缓存创建与命中 token`
  - `0760cda9` `feat(i18n): 添加缓存命中/创建/命中率文案`
  - `7386f38c` `test(usage): API契约测试补充缓存创建/命中token字段`
  - `cb4f0015` `docs: use Codex skill path in examples`
  - `154e0ed6` `fix: force Content-Type to application/json on non-streaming responses`
  - `f5cecea5` `fix(ui): 放开 Select 下拉选项区高度，避免选项被截断`
  - `a4362963` `fix: validate OpenAI sticky session groups`
  - `606bfb42` `docs: update Sub2API admin skill`
- Scope:
  - Synced OpenAI sticky-session validation and previous-response cleanup behavior
  - Synced usage cache token aggregation, API contract coverage, and cache token i18n labels
  - Added upstream Sub2API admin skill documentation/scripts and Select dropdown height fix
- Merge notes:
  - Batch applied cleanly with no manual conflict resolution

### Absorbed upstream batch `2026-07-07-b`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `16bc8769` `fix(usage): sync 5h ResetsAt to SessionWindowEnd and zero expired window`
  - `9a0e4398` `fix(openai): 跨组会话失配保护移到生效的 WSv2 路径并补测`
  - `1a86c6ce` `fix: enforce exclusive group access for api keys`
  - `217f8599` `fix(openai): /responses 传输层错误转 failover + 持久故障临时摘除账号`
  - `f20e6bf7` `feat(ops): 新增 account_temp_unscheduled_count 告警指标`
  - `af19d443` `feat(proxies): 代理有效期与失败回退`
- Scope:
  - Synced usage session-window reset handling and related API contract coverage
  - Synced OpenAI WSv2 cross-group session protection and transport-error failover handling
  - Added exclusive API-key group access enforcement and temporary unscheduled account ops metric
  - Added proxy expiry, proxy fallback, migration, backend services, and frontend proxy UI support
- Merge notes:
  - Resolved `backend/internal/handler/dto/types.go` by merging upstream proxy fallback origin fields with local account health status fields
  - Preserved the local `credentials_status,omitempty` response tag

### Absorbed upstream batch `2026-07-07-c`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `0aad6030` `chore: sync VERSION to 0.1.135 [skip ci]`
  - `d251487d` `fix(openai): propagate prompt cache key for chat completions`
  - `acbcb50d` `chore: update README`
  - `b7cfe246` `chore: update README`
  - `be017445` `chore: update sponsors`
  - `329414ea` `feat(admin): /admin/users 新增按用户 API Key 所在分组过滤`
  - `a67b10f4` `fix(gateway): anchor responses fallback to input`
  - `da30c599` `fix(openai): fail over image server errors`
  - `63d95b4e` `chore: updeta sponsors`
  - `c10598df` `fix idempotency response utf8 truncation`
- Scope:
  - Synced version/readme/sponsor updates, including upstream README_CN removal
  - Synced prompt cache key propagation for OpenAI chat completions and idempotency UTF-8 truncation fix
  - Added admin user filtering by API-key group and related backend/frontend tests
  - Synced OpenAI responses fallback anchoring and image server-error failover coverage
- Merge notes:
  - Batch applied cleanly with no manual conflict resolution
