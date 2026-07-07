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

### Absorbed upstream batch `2026-07-07-d`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `2c45f91d` `fix openai failover model body replacement`
  - `46bd7968` `fix: reuse OpenAI failover error body`
  - `914c059f` `fix: avoid double-writing error frame on non-stream upstream errors`
  - `6c886316` `fix(gateway): prevent double-write on error passthrough responses`
  - `20f3f204` `fix(gateway): complete MarkResponseCommitted coverage for all platforms`
  - `bf28a009` `fix(bedrock): filter unsupported top-level fields and fix beta token cleanup`
  - `12962bab` `refactor(bedrock): merge header filtering into ApplyBedrockCCCompat`
  - `72c11216` `fix(frontend): bedrock_cc_compat toggle not persisting on reload`
  - `448936d9` `fix(ci): fix gofmt, errcheck, and test for supported context-management beta token`
  - `30c00a91` `优化账号分组调度索引`
  - `2c27548b` `优化调度日志循环开销`
  - `d662c973` `feat: claude-fable-5`
  - `0acf00c4` `Add admin compliance acknowledgement gate`
- Scope:
  - Synced OpenAI failover body reuse, non-stream double-write protection, and response-commit coverage
  - Synced Bedrock compatibility cleanup, scheduler index/log optimizations, and Claude Fable 5 support
  - Added admin compliance acknowledgement gate with backend middleware/service, legal docs, frontend dialog, and store
- Merge notes:
  - Resolved `frontend/src/App.vue` by keeping local admin route keep-alive behavior and adding upstream admin compliance store initialization

### Absorbed upstream batch `2026-07-07-e`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `fa8f1749` `fix: treat invalid_refresh_token as non-retryable / 将 invalid_refresh_token 判定为不可重试`
  - `e34ad2b1` `chore: sync VERSION to 0.1.136 [skip ci]`
  - `ad135854` `fix(docker): ship docs/legal in build context for admin-compliance gate`
  - `727ac3f6` `fix: add app_session_terminated to non-retryable refresh errors / 将 app_session_terminated 添加到不可重试的刷新错误中`
  - `0da1fe28` `fix(openai): prevent false image billing on text-only /v1/responses requests`
  - `6bf6722b` `chore: gofmt test file`
  - `65559ac5` `fix(antigravity): merge system role messages`
  - `f8c80bf0` `fix(auth): apply promo codes to oauth signups`
  - `b256f911` `fix(gateway): intercept max_tokens=1 haiku probes for streaming requests too`
  - `e4c255a7` `fix：account expiry autopause index`
  - `b62b573f` `feat(openai): cyber_policy 硬阻断全链路透传、审计与计费`
  - `f6e0ebc6` `fix: preserve Anthropic window cooldowns`
  - `edfd5e37` `fix(apicompat): default tool strict to false`
  - `c1c28ac7` `fix(gateway): 解压 zstd 上游响应体`
  - `c70c6a26` `feat(渠道监控): 检测间隔支持正负随机抖动配置`
  - `8ce7b9a8` `feat: configure Claude OAuth system prompt blocks`
  - `25a9762a` `feat: show account id in account list`
- Scope:
  - Synced refresh-token non-retryable error classification, Docker legal-doc packaging, OpenAI false image billing fix, and Antigravity system-role merge
  - Added promo-code OAuth signup handling, streaming Haiku probe interception, account expiry autopause index, zstd response decompression, and channel monitor jitter
  - Added OpenAI cyber policy passthrough/blocking/audit/billing flow and Claude OAuth system prompt block settings
  - Exposed account IDs in the admin account list
- Merge notes:
  - Resolved `frontend/src/views/admin/AccountsView.vue` by preserving local column width settings and adding the upstream `id` column after `name`

### Absorbed upstream batch `2026-07-07-f`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `ab9987b2` `fix(gateway): fail over on non-JSON 2xx responses`
  - `b63b4116` `fix: remove unused billing attribution helper`
  - `bbd97024` `fix(frontend): bump form-data to >=4.0.6 via pnpm override`
  - `8b698ff4` `fix account list parameter limit`
  - `b0579c48` `fix: move user wait queue accounting off hot path`
  - `74199b6a` `fix: reduce token refresh retry amplification`
  - `9b270f11` `refactor: inline token refresh retry reason prefix`
  - `34e66ec0` `fix: outbox scheduler snapshot coalesce`
  - `3ef70b04` `fix: safely coalesce scheduler outbox events`
  - `60cf89ae` `fix: recover scheduler outbox invalid dedup index`
  - `b3ec6288` `fix: release scheduler outbox dedup on claim`
  - `1fdbe52f` `chore(migrations): renumber scheduler outbox dedup migrations 151/152 -> 152/153`
  - `cb14935e` `fix: cleanup consumed scheduler outbox rows`
  - `31dc8913` `chore(outbox-cleanup): add 10s grace to defend against id-vs-commit race`
  - `acaffe29` `fix(account-repo): refresh candidates SQL excluded healthy accounts; fix CI build`
  - `f069c9ae` `fix(outbox-dedup): buildSchedulerGroupPayload typed-nil broke dedup_key consistency`
  - `b88f8e4c` `fix(openai-probe): /responses 能力探测增加工具调用校验`
  - `b8169492` `feat(openai-quota): query + reset rate-limit credits for OpenAI accounts`
  - `56c62c59` `fix(auth): include client ip in acl denial message`
  - `b8a482e1` `fix(ci): unblock main after recent merges`
- Scope:
  - Synced gateway non-JSON failover, account list parameter limit, wait-queue accounting, and token refresh retry reductions
  - Synced scheduler outbox dedup/coalescing/cleanup migrations and repository/service coverage
  - Added OpenAI responses probe tool-call validation and OpenAI quota query/reset support
  - Synced frontend dependency override and CI fixes
- Merge notes:
  - Batch applied cleanly with no manual conflict resolution

### Absorbed upstream batch `2026-07-07-g`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `6c2db4f4` `fix(gemini): clean unsupported tool schema fields`
  - `c906bf00` `feat(billing): add DeepSeek V4 Pro / Flash fallback pricing`
  - `27e26a3a` `chore: fix gofmt alignment`
  - `5a593a51` `test(billing): tighten DeepSeek V4 fallback assertions; clarify branch comments`
  - `f597d98b` `test(openai): use unpriced model in usage test`
  - `a4ce7339` `feat(billing): add GLM / Kimi / MiniMax fallback pricing for Chinese LLM providers`
  - `c90089c8` `fix(billing): address Copilot review feedback`
  - `4f5f2788` `fix(billing): add kimi-for-coding fallback pricing`
  - `262fe123` `feat(billing): 为 doubao-embedding-vision 添加图文差别兜底定价`
  - `142d8c36` `fix(gateway): normalize DeepSeek reasoning_effort 'max' to 'xhigh'`
  - `34b1e56e` `test: add 'max' → 'xhigh' test cases for reasoning effort normalization`
  - `6baf00d7` `fix(gateway): protocol-aware thinking-block filtering for Anthropic-compatible upstreams`
  - `efbf6d20` `fix(test): update FilterThinkingBlocksForRetry call to use mappedModel param`
  - `56c6325d` `fix(gateway): rewrite thinking.type=enabled to adaptive for MiniMax M-series`
  - `5c528397` `doc(thinking-protocol): clarify mappedModel vs originalModel semantics per call path`
  - `a05d9e87` `feat(billing): 国产模型 thinking-enabled 自动填充 reasoning_effort 默认值`
  - `6c7203d8` `fix(gateway): preserve SSE event:error body so ops logs reflect real upstream errors`
  - `4a5665da` `chore: sync VERSION to 0.1.137 [skip ci]`
  - `abc203a3` `chore: update pnpm action setup`
  - `369f53a7` `chore: force node24 for cla action`
- Scope:
  - Synced Gemini schema cleanup and OpenAI/Gemini thinking-block protocol compatibility
  - Added DeepSeek/GLM/Kimi/MiniMax/Doubao fallback pricing and reasoning_effort normalization behavior
  - Synced SSE error body preservation and CI workflow updates
- Merge notes:
  - Batch applied cleanly with no manual conflict resolution

### Absorbed upstream batch `2026-07-07-h`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `8c4a43cf` `fix(gemini): satisfy schema cleanup test lint`
  - `31640363` `fix(deploy): add :Z SELinux labels to bind mounts`
  - `bab8a9a9` `fix(openai): log /v1/chat/completions upstream endpoint for chat-only API-key accounts`
  - `df51edfb` `Preserve OAuth instructions while keeping developer input`
  - `952be871` `fix(frontend): refresh custom page document title`
- Scope:
  - Synced Gemini lint fix, deploy SELinux bind-mount labels, and OpenAI chat endpoint logging
  - Synced OAuth instruction preservation in Codex transform flow
  - Synced custom-page document title refresh behavior
- Merge notes:
  - Resolved `frontend/src/App.vue` by keeping local admin keep-alive helpers and adding upstream admin settings store plus dynamic document title refresh

### Absorbed upstream batch `2026-07-07-i`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `51d72290` `fix(usage): 显示缓存 Token 明细`
  - `89cfe24a` `fix(openai): normalize glm reasoning effort`
  - `e3e31bd4` `fix(gateway): auto mode recognize Claude Code IDE clients via any cc_entrypoint`
  - `510adf70` `feat(scheduling): add opt-in "prefer soonest reset" account selection`
  - `2dc1387b` `fix(promo): allow clearing promo code expiry on edit`
  - `d3dfa28f` `Update CC Switch OpenAI default model`
  - `0fa604ba` `feat: apply affiliate rebate to subscription payments`
  - `ecedc7c8` `fix(auth): enforce email bind suffix whitelist`
  - `40e1cc14` `fix(gateway): filter anthropic-beta on the Vertex Anthropic path (#3358)`
  - `efffd5d7` `test(gateway): Vertex anthropic-beta filtering`
  - `6cfb7898` `fix(claude-mimicry): drop the cch sign to match new Claude Code CLI`
  - `5cb8cdd3` `test(claude-code): detection recognizes the new-CLI billing block (no cch)`
  - `b0d5592a` `fix(images): 识别 response.incomplete + 记录软失败上游响应`
  - `f4b51b0f` `fix(lint): check WriteString return value in summarizeOpenAIImagesNoOutputBody`
  - `69366878` `fix(lint): check WriteString return in summarizeOpenAIImagesNoOutputBody` (`already present / empty`)
- Scope:
  - Synced usage cache token details, GLM reasoning normalization, Claude Code IDE recognition, and prefer-soonest-reset scheduling
  - Synced promo expiry clearing, CC Switch default model, affiliate rebate payment handling, and email bind suffix whitelist enforcement
  - Synced Vertex Anthropic beta filtering, Claude Code mimicry update, and OpenAI image incomplete soft-failure logging
- Merge notes:
  - `69366878` was skipped as empty because the equivalent WriteString lint fix was already present after `f4b51b0f`

### Absorbed upstream batch `2026-07-07-j`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `7c2fee6c` `fix(billing): dedup fallback pricing warn to stop per-request log spam (#3394)`
  - `6239e395` `i18n(channel): explain case-insensitive matching in pricing conflict messages (#3394)`
  - `d430343f` `chore: sync VERSION to 0.1.138 [skip ci]`
  - `e5f38a6f` `chore: update sponsors`
  - `9f5b57fc` `fix(billing): 防止余额计费持续透支`
  - `c6f375d3` `fix(payment): 订阅订单应用充值汇率换算`
  - `85a3b122` `chore: update sponsors`
  - `32df33a1` `feat: add codex personal access token auth`
- Scope:
  - Synced fallback pricing warning deduplication, pricing conflict i18n, version/sponsor updates, and balance overdraft prevention
  - Synced subscription exchange-rate payment conversion and Codex personal access token authentication support
- Merge notes:
  - Resolved `backend/internal/service/account_test_service.go` by preserving local Codex CLI headers/custom User-Agent behavior and adding upstream `setOpenAIChatGPTAccountHeaders`

### Absorbed upstream batch `2026-07-08-a`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `147c1879` `fix(payment): support plural subscription validity units`
  - `9491de0a` `fix(images): pass content-moderation refusals through as 400 instead of retrying`
  - `28e7adef` `fix(keys): add CLAUDE_CODE_ATTRIBUTION_HEADER=0 to Claude Code terminal templates`
  - `ae5e980d` `fix(gateway): enforce codex_cli_only restriction on /v1/chat/completions`
  - `9707dedc` `fix(ops): prevent monitoring trend cards from growing unbounded`
  - `82576e0a` `fix(auth): stop swallowing email auth identity create error via shadowed err`
  - `65fa7289` `fix(openai): fail over on chat transport errors`
  - `a2cf297d` `feat: 新增 OpenAI quota headroom 调度权重`
  - `30adee43` `feat(admin/accounts): confirm before OpenAI weekly limit reset`
  - `a1560ae7` `chore: update sponsors`
  - `063454ae` `fix(admin/usage): populate cache creation/read token breakdown in stats`
  - `4567f658` `test(admin/usage): update sqlmock rows for cache breakdown columns`
  - `dbdbfb11` `fix: avoid default codex instructions for chat bridge`
  - `00d68ff6` `feat(openai): add GPT-5.5 codex instructions and use as latest fallback`
  - `0a97a5f4` `fix(token-refresh): treat refresh_token_invalidated as non-retryable`
  - `2b49d662` `fix(openai): dedupe passthrough function call args`
  - `55242ffa` `fix(admin): 订单金额币种符号读取 currency 字段`
  - `650c50e3` `fix(antigravity): add project fallback for standard tier`
  - `01127820` `fix(gateway-openai): codex spark 剥离 image_generation 工具，修复 502`
  - `cc7612bd` `Detect OpenAI overloaded error codes`
  - `5bd9368a` `fix claude oauth token exchange payload`
  - `65ad7df4` `fix(payment): 修复后端返回空supported_types时支付提供商卡片消失的问题`
  - `29122e30` `fix(apicompat): avoid doubling tool_call arguments from single-chunk upstreams`
  - `2a58a57a` `fix(frontend): use configured API base for direct requests`
- Scope:
  - Synced payment/order/currency fixes, admin usage cache stats, OpenAI quota headroom scheduling, and weekly limit reset confirmation
  - Synced Codex-only chat restriction, GPT-5.5 Codex instructions, token refresh handling, OpenAI chat failover, overloaded error detection, and Codex Spark image tool stripping
  - Synced frontend direct request URL construction through configured API base
- Merge notes:
  - Resolved `frontend/src/components/admin/account/AccountTestModal.vue` by preserving local compact test request body and switching the URL to `buildApiUrl`
