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

### Absorbed upstream batch `2026-07-08-b`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `c9f42e1f` `fix(lint): gofmt token_refresh_service.go after refresh_token_invalidated addition`
  - `39be1ec9` `feat: add grok subscription support`
- Scope:
  - Synced token refresh lint cleanup
  - Added Grok subscription/OAuth/quota/token refresh/gateway support across backend, admin routes, frontend account setup, and model metadata
- Merge notes:
  - Resolved Wire conflicts by preserving local `settingService` injection and adding upstream `grokOAuthService` injection in both `backend/internal/service/wire.go` and generated `backend/cmd/server/wire_gen.go`

### Absorbed upstream batch `2026-07-08-c`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `b3a07aea` `fix: align grok oauth exchange with xai`
  - `b2e2c7e6` `fix: harden grok oauth gateway paths`
  - `f29ccc7d` `fix: reduce grok oauth account-risk paths`
- Scope:
  - Synced Grok OAuth exchange alignment, hardened gateway paths, and reduced account-risk handling
- Merge notes:
  - Resolved `backend/internal/service/openai_account_runtime_block_fastpath_test.go` by preserving local OpenAI Cloudflare cooldown tests and adding upstream Grok 429 failover coverage

### Absorbed upstream batch `2026-07-08-d`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `e7a4f3f4` `test: fix grok oauth validation suite`
  - `1b9645ca` `fix: allow grok scheduler capability routing`
  - `0d286421` `feat: add grok quota probe parity`
  - `939905b8` `fix: refine grok quota pause behavior`
  - `8a87a658` `test: cover grok readiness paths`
  - `2a804958` `docs: clarify grok public route scope`
  - `720db898` `test: harden grok quota readiness`
  - `815bc6c9` `feat(risk-control): record matched keyword in keyword-block logs`
  - `8a7269f5` `fix: sanitize verbose OpenAI response failed events`
  - `40c82527` `fix(apicompat): 规范化 custom 工具 schema`
  - `ac6e36f9` `feat(cli): sub2api-admin 支持 SUB2API_JWT 认证回退`
  - `44f502ba` `fix: address grok review feedback`
  - `98feeccb` `docs: note admin account wizard requirement in source-compile install`
  - `fcd3bc12` `fix: return 404 model_not_found instead of 503 when no account supports the model`
  - `f93a6c50` `fix: repair CI build & lint regressions on main`
  - `819fda34` `feat(codex-detect): codex_cli_only 检测加固 + 引擎指纹统一信号列表 + 账号级 app-server`
  - `9a0fbcc8` `chore: update sponsors`
  - `c2754222` `chore: sync VERSION to 0.1.139 [skip ci]`
  - `7a38c662` `Bridge OpenAI count_tokens to responses input_tokens`
  - `e5f7836b` `fix(openai): set tool_choice auto for Codex image bridge`
  - `88ca0c1d` `fix(payment): 显示订阅 CNY 换算实付金额`
  - `bad87ff5` `feat(ops): add api key filter to system logs`
  - `b244f850` `feat(keys): add column settings`
  - `b1403e8b` `fix(payment): keep subscription price as direct pay amount`
  - `da810c3b` `fix(keys): reactivate exhausted keys set to unlimited`
- Scope:
  - Synced Grok quota/readiness hardening, risk-control matched keyword logging, custom tool schema normalization, and admin CLI JWT fallback
  - Synced model-not-found 404 handling, codex_cli_only detection hardening, OpenAI count_tokens bridge, Codex image bridge tool_choice, and ops API-key log filtering
  - Synced payment display fixes and user API key column settings/reactivation behavior
- Merge notes:
  - Batch applied cleanly with no manual conflict resolution

### Absorbed upstream batch `2026-07-08-e`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `82553c4d` `fix(openai): preserve quota platform in usage billing`
  - `d86e8325` `fix: allow five platform quota updates`
  - `709cf618` `修复 OpenAI GPT-5.5 的 Codex 指令选择`
  - `7cbf82ed` `修复 OpenAI 上下文窗口错误误触发账号切换`
  - `4a7148e2` `fix: support grok cli compatibility routes`
- Scope:
  - Synced OpenAI quota-platform billing preservation, five-platform quota settings, GPT-5.5 Codex instruction selection, and context-window error handling
  - Added Grok CLI compatibility routes and related account-test coverage
- Merge notes:
  - Resolved `frontend/src/components/admin/account/__tests__/AccountTestModal.spec.ts` by preserving local OpenAI compact probe coverage and adding upstream Grok account modal test support

### Absorbed upstream batch `2026-07-08-f`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `10e623f6` `fix: allow grok messages compatibility`
  - `438510d2` `fix: sanitize grok codex responses payloads`
  - `a0a3d0c3` `fix: show privacy setting result accurately`
  - `260fda19` `feat: fix OAuth email completion flow`
  - `7316d830` `fix(payment): 区分退款 pending 并收敛匿名查单`
  - `93a3bf30` `Fix refund pending finalization gaps`
  - `901958ba` `feat(openai-ws): add http_bridge ingress mode and account ws selector`
  - `0476b5c9` `fix(openai-ws): allow http_bridge accounts in ws ingress selection`
  - `56543213` `fix(frontend): enable ws mode editing for openai setup-token accounts`
  - `906be3f7` `fix(openai-ws): honor explicit http_bridge ingress mode`
  - `185f9c99` `fix(auth-signup): 平台配额快照脱离注册事务 + grok 补入 CHECK 约束`
  - `bf4f0067` `fix: 修复用户列表使用时间排序冲突`
  - `6c46c2cb` `test: update platform quota contract for grok`
  - `345d5c6b` `chore: update sponsors`
  - `915c60b1` `feat(group): 订阅分组新增可选的高峰时段倍率，以支持智谱等coding plan的高峰时段`
  - `89b2d63e` `chore: sync VERSION to 0.1.140 [skip ci]`
  - `cafc95c3` `feat: align user usage analytics with admin`
  - `93032611` `fix: 修复订阅支付金额显示错误`
  - `dc1bc154` `chore: sync VERSION to 0.1.141 [skip ci]`
  - `6e547765` `fix: localize English UI strings`
  - `73de2ea7` `fix(openai): preserve encrypted reasoning across turns on codex OAuth path`
  - `b4f38b09` `fix: preserve zh locale in i18n cleanup`
  - `41bb15d3` `fix: remove remaining English fallbacks`
  - `1158cf31` `fix: require localized interval validation`
  - `03727ac3` `fix(subscription): 修复订阅撤销软删除失效`
  - `b28a2233` `fix(openai): 保留 GPT-5.5 Pro Codex 模型名`
- Scope:
  - Synced Grok messages/codex payload compatibility, OpenAI WS `http_bridge` ingress mode, OAuth email completion, refund/payment fixes, platform quota signup behavior, and group peak-rate multiplier
  - Synced user usage analytics alignment, localization cleanup, encrypted reasoning preservation, subscription revoke fix, and GPT-5.5 Pro Codex model-name retention
- Merge notes:
  - Resolved `backend/internal/service/billing_service.go` and `backend/internal/service/openai_model_alias.go` by preserving local GPT-5.6 sol/terra/luna pricing/alias logic while adding upstream GPT-5.5 Pro retention behavior

### Absorbed upstream batch `2026-07-08-g`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `59e9356c` `feat: 抹除 Anthropic OAuth 请求中客户端 dateline 隐写指纹`
  - `db041423` `feat: 适配 sonnet5`
  - `605b6f66` `fix(test): resolve oauth passthrough unit build errors`
  - `3b5d812f` `fix: route grok media endpoints`
  - `2fe756e` `fix: recognize grok media models`
  - `c3e86060` `fix: include official grok media model ids`
  - `e28b11ad` `fix(test): align compact passthrough expectation`
  - `bdf7ead1` `feat(spark-shadow): OpenAI Spark 链接型影子账号`
- Scope:
  - Synced Anthropic OAuth dateline fingerprint stripping, Sonnet 5 model adaptation, OAuth passthrough test build fixes, Grok media endpoint/model ID support, and compact passthrough expectation updates
  - Added OpenAI Spark linked shadow accounts, including schema migrations, parent account enrichment, credential passthrough, Spark quota isolation, scheduler routing, admin UI affordances, and related backend/frontend tests
- Merge notes:
  - Resolved `backend/internal/handler/admin/account_handler.go` by enriching shadow parent account data before list ETag calculation while preserving local `healthStatus` ETag input
  - Resolved `backend/internal/service/account_test_service.go` by applying upstream shadow credential resolution while preserving local compact Codex probe headers, custom OpenAI user agent, and ChatGPT account header injection
  - Resolved `backend/internal/service/openai_codex_transform.go` by keeping fork-specific GPT-5.6 sol/terra/luna and GPT-5.3 Codex Spark mappings while absorbing upstream Spark shadow support
  - Resolved `frontend/src/views/admin/AccountsView.vue` by using upstream `accountDisplayEmail(row)` for shadow parent email display while preserving local compact truncation styling
- Validation:
  - Not run locally: this Windows environment currently has neither `go` nor `make` on PATH, so backend batch validation must be run in the server/CI environment

### Absorbed upstream batch `2026-07-08-h`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `96c9166b` `fix(usage): 用户使用记录默认显示推理强度列`
  - `de64b026` `fix: resolve build version from release tag`
  - `e998f804` `fix: ci`
  - `fd004bdd` `fix(account-repo): Clone query before Count to prevent state pollution`
  - `ebd83dc0` `fix: 触发ci`
  - `a34d4967` `feat: add LLM media tester`
  - `42e471f5` `fix: harden grok media routing`
  - `c9fb221a` `fix: satisfy grok media lint`
  - `8c2d9b9a` `chore(openai): remove gpt-5.3-codex from OpenAI default model list`
  - `f77cf6b4` `Revert "feat: add LLM media tester"`
- Scope:
  - Synced usage record column defaults, release-tag version resolution, CI fixes, account repository query cloning before Count, and Grok media routing/lint hardening
  - Followed upstream by absorbing the LLM media tester addition and its later revert in order, leaving the final tree aligned with upstream
  - Removed `gpt-5.3-codex` from the OpenAI default model list while preserving fork-specific Codex/Spark mappings outside the default-list change
- Merge notes:
  - Batch applied cleanly with no manual conflict resolution
- Validation:
  - Not run locally: this Windows environment currently has neither `go` nor `make` on PATH

### Absorbed upstream batch `2026-07-08-q`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `6ae5fc31` `fix(admin): gate scheduler score calculation`
  - `a4f942d8` `fix(deps): 升级 AWS SDK 修复 govulncheck 报告的 GO-2026-5764`
  - `d9e514f9` `refactor(i18n): 拆分 zh/en 语言包为域模块（深比较等价校验通过）`
  - `084d26cb` `refactor(service): 纯移动拆分 gateway_service.go（7294→1289行）`
  - `50043b11` `refactor(service): 纯移动拆分 setting_service.go（5471→263行）`
  - `4d23ad4b` `refactor(service): 纯移动拆分 openai_gateway_service.go（4872→1095行）`
  - `db3bd997` `refactor(repository): 纯移动拆分 usage_log_repo.go（4701→212行）`
  - `d0f66933` `refactor(service): 纯移动拆分 openai_ws_forwarder.go（4675→399行）`
  - `2a4c28e8` `refactor(service): 纯移动拆分 antigravity_gateway_service.go（4664→639行）`
  - `f013bc11` `refactor(service): 纯移动拆分 admin_service.go（4409→642行）`
  - `bb5d2e84` `refactor(handler): 纯移动拆分 setting_handler.go（3957→468行）`
  - `d0bcaa8b` `test(i18n): 新增语言包 spread 组装键冲突守卫测试`
  - `a56eb5b4` `fix(compact): body-signal 提升上移到 handler 层并对齐 path-based 链路`
- Scope:
  - Synced scheduler score opt-in gate, AWS SDK eventstream vulnerability fix, and compact body-signal handler-layer routing alignment
  - Synced large pure-move file splits for i18n, gateway, settings, OpenAI gateway, usage logs, OpenAI WS forwarder, Antigravity gateway, admin service, and setting handler
  - Added upstream i18n spread key-collision guard test
- Merge notes:
  - Resolved i18n split conflicts by migrating local account health/token refresh/bulk delete/import group/deploy tab/table page-size placeholder keys into the new `en/zh` domain modules; verified old full-object and new spread-object equality with a local Node deep-flatten comparison (`missing=0`, `changed=0` for both languages)
  - Resolved `setting_service.go`, `openai_gateway_service.go`, and `openai_ws_forwarder.go` pure-move conflicts by adopting the slim split files after confirming moved logic exists in the new companion files
  - Preserved fork-specific OpenAI guard behavior by migrating `guard.SanitizeReasoning`, `guard.ConfuseCodexMetadataLight`, and `guard.ApplySessionGovernance` into `backend/internal/service/openai_gateway_forward.go`
- Validation:
  - `git diff --check` passed during conflict resolution
  - Local Node i18n object comparison passed for `en` and `zh`
  - Not run locally: this Windows environment currently has neither `go` nor `make` on PATH

### Absorbed upstream batch `2026-07-08-o`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `a42e9e3f` `feat: add Grok image pricing controls`
  - `a23a2635` `feat(subscription): preview CNY pay amount in plan editor`
  - `6b4c3c6e` `test: cover Grok video media unit billing`
  - `95717a7e` `test: clarify Grok video media billing name`
  - `de31e573` `docs: document Grok video media billing unit`
  - `4b321142` `fix: update codex version gate wording`
  - `ec7b2064` `feat: support API key account header override`
  - `31b6e0d9` `fix: address header override audit findings`
  - `5089c303` `优化 Redis SCAN 清理架构`
  - `3c43fdec` `docs: add batch image PR readiness notes`
  - `89edba80` `fix: restrict batch image groups to gemini`
  - `a564051f` `chore: renew expired xlsx audit exceptions to 2026-10-06`
  - `7c8e1320` `style: replace interface{} with any per golangci-lint gofmt rewrite rules`
  - `7650cce5` `fix(redis-index): harden index-driven cleanup introduced in #3762`
  - `2fb212b7` `fix(openai): 区分 responses compact 入站端点`
  - `75e30894` `fix: normalize inbound endpoints from raw request path`
  - `ddd63a84` `test: cover endpoint wildcard edge cases`
  - `aff14816` `fix: address batch image ci failures`
  - `c34db70a` `fix: bridge grok composer image inputs`
  - `9d5f1b73` `chore: sync VERSION to 0.1.146 [skip ci]`
- Scope:
  - Synced Grok image pricing controls, Grok video media billing tests/docs, Codex version gate wording, API key account header override support, and header override audit hardening
  - Synced Redis active-index cleanup architecture/hardening, user message queue Redis cleanup support, compact responses inbound endpoint distinction, endpoint normalization/wildcard coverage, batch image CI fixes, and VERSION 0.1.146
  - Synced batch image PR readiness notes, Gemini-only batch image group restriction, xlsx audit exception renewal, `interface{}` to `any` style cleanup, and Grok composer image input bridge
- Merge notes:
  - Resolved `backend/internal/repository/concurrency_cache.go` by keeping upstream index-driven account/user startup cleanup, expired-index hardening, and one-time legacy wait sweep while preserving the fork's API key slot startup cleanup as a narrow `concurrency:api_key:*` scan using the single-key cleanup script
  - Resolved `backend/internal/repository/concurrency_cache_integration_test.go` by covering both fork API key cleanup and upstream protection for unindexed account slots/wait keys
  - Resolved `.github/audit-exceptions.yml` by preserving the already-renewed `xlsx` exception expiry date `2026-10-06` instead of reverting to `2026-08-06`
  - Resolved `deploy/Dockerfile` by keeping upstream's lockfile-compatible `pnpm@9.15.9` pin for reproducible Docker builds
- Validation:
  - `git diff --check` passed during conflict resolution
  - Not run locally: this Windows environment currently has neither `go` nor `make` on PATH

### Absorbed upstream batch `2026-07-08-p`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `83d7b0e4` `fix: isolate duplicate item integration test`
  - `438f17be` `fix(openai): avoid compact usage loss from json sse heuristic`
  - `867616fc` `fix(websearch): 剥离历史消息中上游不接受的 web search 块`
  - `40c563c4` `fix(gateway): 记录请求体解析失败的真实原因，不再吞错`
  - `156e1733` `chore: update sponsors`
  - `e979990b` `add force priority fast policy action`
  - `17b6481f` `chore: update sponsors`
  - `fd64d07e` `fix(codex): 剥离续链 function_call item 的非法 item_* id`
  - `c10ae9f9` `chore: remove batch image draft reports`
  - `3b209935` `fix: route text-only grok video requests`
  - `a02700c1` `fix(image): 识别 Codex namespace image_gen 工具声明以拦截生图请求`
  - `5c0e580f` `fix(messages): /v1/messages 入站支持不兼容 Responses API 的 OpenAI 上游`
  - `69d02d20` `fix: 补齐 failover 路径的 appendOpsUpstreamError 日志`
  - `80a229bc` `fix(batch-image): 修复审计发现的计费死锁、状态机与队列原子性缺陷`
  - `b5604afd` `test(batch-image): 集成测试适配 ReplaceBatchImageItemsForJob 的 indexing 状态守卫`
  - `dad92488` `fix(messages): 修复 /v1/messages CC 回退的错误处理旁路并补齐流式测试`
  - `c7e44a83` `fix(frontend): 路由切换后保持侧边栏滚动位置`
  - `a86c534c` `test: 补充侧边栏滚动位置持久化的前端测试`
  - `2dd2be99` `fix(compact): 识别 /v1/responses body 中的 compaction_trigger 信号`
  - `d754be0d` `refactor(gateway): 抽取 CC forwarder 公共管线并拆分两大 service 文件`
- Scope:
  - Synced compact usage preservation, compact body signal routing, web-search block filtering, request body parse error logging, force-priority fast policy action, and Codex function_call/image namespace handling
  - Synced Grok text-only video routing, OpenAI `/v1/messages` chat fallback support, failover upstream-error logging, frontend sidebar scroll position persistence, and sponsor/report maintenance
  - Synced batch image audit hardening around billing deadlocks, state machine/queue atomicity, and indexing-status integration tests
  - Synced gateway CC forwarder common pipeline extraction and service file split
- Merge notes:
  - No manual conflicts in this batch; all commits cherry-picked cleanly or auto-merged by Git
- Validation:
  - Not run locally: this Windows environment currently has neither `go` nor `make` on PATH

### Absorbed upstream batch `2026-07-08-i`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `aac3261c` `fix: convert grok image edit uploads`
  - `71c16b72` `fix(test): correct passthrough failover assertions`
  - `1034f576` `fix: 高峰倍率全链路透传、计费术语修正与边界处理`
  - `a0eaf12a` `fix: test`
  - `f5b29612` `fix: Handle invalid arguments correctly for Gemini reasoning models`
  - `be26e6ff` `fix: test`
  - `60da9ba1` `fix(frontend): refine grok icon, palette and i18n`
  - `45be32b2` `fix: remove space lines`
  - `e236bff1` `fix: aggregate user model stats by requested model`
  - `98be187d` `chore: update sponsors`
  - `7dc7cfce` `chore: sync VERSION to 0.1.142 [skip ci]`
  - `0435417f` `fix: enable grok media generation groups`
  - `9934bd25` `fix: default grok group media generation`
  - `0ff93aca` `feat: add usage IP geolocation lookup`
  - `b26dcc3d` `feat(subscription): 支持恢复已撤销订阅`
  - `a5781fe3` `fix Claude Code stream keepalive stalls`
  - `7869b7fe` `fix(anthropic): 支持 API Key Bearer 认证方式`
  - `c797159b` `fix(openai): skip Codex image bridge injections for /responses/compact`
  - `c4128580` `fix: handle OpenAI OAuth count_tokens scope errors`
  - `dfb36e45` `show reset credit expiration`
- Scope:
  - Synced Grok image edit upload conversion, media generation group defaults, frontend Grok icon/palette/i18n polish, and requested-model usage aggregation
  - Synced peak-rate full-path propagation and billing terminology fixes, Gemini reasoning invalid-argument handling, OpenAI compact response image-bridge skip, and count_tokens scope-error handling
  - Added usage IP geolocation lookup, revoked subscription restoration, reset credit expiration display, Claude Code stream keepalive stability, and Anthropic API Key Bearer auth support
- Merge notes:
  - Batch applied cleanly with no manual conflict resolution
- Validation:
  - Not run locally: this Windows environment currently has neither `go` nor `make` on PATH

### Absorbed upstream batch `2026-07-08-j`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `5fe4ee1d` `style(test): gofmt oauth passthrough tests`
  - `d0b8760e` `fix: preserve OpenAI plan type from inactive workspaces`
  - `df59b8b9` `fix: persist OpenAI subscription expiration`
  - `5436d0bf` `feat(admin): add group column settings`
  - `99a8d8ad` `fix: normalize grok imagine image alias`
  - `089a7b7f` `feat(keys): add api key concurrency stats`
  - `11a3da65` `fix(group): harden peak-rate config handling and label peak windows with server timezone`
  - `fa70a721` `test(keys): update api key contract concurrency field`
  - `41af8e8e` `chore: update sponsors`
  - `0b8e5eec` `chore: sync VERSION to 0.1.143 [skip ci]`
  - `37243632` `修复邀请码普通兑换错误`
  - `3f2ef604` `fix: optimize ops realtime account stats`
  - `72ccd1b1` `fix: batch group capacity summaries`
  - `a5638a4e` `fix: match Codex session imports by chatgpt_user_id before shared account id`
  - `36d5f4e4` `feat: make setup migration timeout configurable`
  - `2b462b07` `fix: preserve configured Grok OAuth concurrency`
  - `d0a1443a` `fix(antigravity): allow oauth 401 auto recovery`
  - `c620ad6a` `fix: align group capacity SQL with target schema`
  - `b2687b22` `fix(frontend): edit grok oauth model mapping`
  - `f3a3a08` `优化并发槽位清理`
- Scope:
  - Synced OpenAI plan/subscription persistence, group column settings, API key concurrency stats, peak-rate server-timezone labeling, and Grok image alias/model-mapping fixes
  - Synced redeem code ordinary redemption fix, ops realtime account stats optimization, batched group capacity summaries, configurable setup migration timeout, Antigravity OAuth 401 auto-recovery, and concurrency slot cleanup
  - Synced Codex session import matching by `chatgpt_user_id` before shared account id
- Merge notes:
  - Batch applied cleanly with no manual conflict resolution
- Validation:
  - Not run locally: this Windows environment currently has neither `go` nor `make` on PATH

### Absorbed upstream batch `2026-07-08-k`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `a1b2b32e` `fix: prevent silent usage_logs drops under queue overflow (#3656)`
  - `f385cdce` `feat: add Codex image tool strip policy`
  - `ebbdc703` `feat(usage): 错误请求对齐用量明细(UI/排序/筛选/列设置)`
  - `df2cedee` `fix: normalize antigravity gemini 3.1 pro routing`
  - `e5dc1f59` `添加 token_expired 刷新不可重试判断`
  - `a994fbd7` `feat: add batch image MVP`
  - `b3f79697` `feat(anthropic): treat 7d_oi (Fable) window 429 as model-level rate limit`
  - `be297b90` `refactor(frontend): merge Codex image bridge and tool policy into one four-state control`
  - `6bd248fd` `fix(admin): avoid merging Codex access-only imports`
  - `4dd3aee5` `fix(openai): use mapped billing model for responses`
- Scope:
  - Synced usage log queue overflow handling, Codex image tool strip policy, error-request usage detail alignment, Antigravity Gemini 3.1 Pro routing normalization, and non-retry handling for `token_expired`
  - Added batch image MVP foundation, including Ent schema, migrations, services, queue/download limiter, handler wiring, cleanup/worker runtime, and documentation
  - Synced Fable `7d_oi` model-level rate-limit handling, frontend Codex image control refactor, Codex access-only import protection, and mapped billing model use for OpenAI responses
- Merge notes:
  - Resolved `frontend/src/components/admin/usage/UsageFilters.vue` by keeping local compact grid classes while adding upstream `errors` mode visibility rules and error-specific filters
  - Resolved `frontend/src/components/common/DataTable.vue` by preserving local `table-row`/slate hover styling while adding upstream clickable row cursor and `rowClick` emission
  - Resolved `backend/cmd/server/wire_gen.go` by preserving existing scheduler/account/concurrency initialization order, adding batch image repository/queue/download limiter providers, and removing a stale duplicate `tokenRefreshService` initialization in favor of the current `grokOAuthService` signature
- Validation:
  - Not run locally: this Windows environment currently has neither `go` nor `make` on PATH

### Absorbed upstream batch `2026-07-08-l`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `dec709fb` `chore: update sponsors`
  - `cbe0c466` `chore: update readme`
  - `b650bdd6` `chore: sync VERSION to 0.1.144 [skip ci]`
  - `20008264` `feat: 点击侧边栏 Logo/站点名返回首页`
  - `498f010e` `fix(部署): 统一 Docker 部署 URL 安全默认值为开发友好模式`
  - `cbfeab96` `fix(antigravity): default gateway forward base URL to the production endpoint`
  - `e76e0499` `fix: sanitize payment response NUL bytes`
  - `b23475ac` `fix(antigravity): refresh server-invalidated tokens`
  - `41cdd438` `fix(gateway): honor Anthropic custom models list`
  - `aee9a7ba` `fix(usage): add UTF-8 BOM to CSV export`
- Scope:
  - Synced sponsors/readme/version updates, sidebar logo/site-name home navigation, Docker URL safe defaults, Antigravity gateway forward base URL default, payment response NUL-byte sanitization, Anthropic custom model list handling, and UTF-8 BOM CSV export
  - Added Antigravity server-invalidated token refresh recovery handling
- Merge notes:
  - Resolved `backend/internal/service/token_refresh_service.go` by preserving local successful-refresh account recovery and health snapshot restoration while adding upstream Antigravity force-refresh marker cleanup on success/non-retryable outcomes
- Validation:
  - Not run locally: this Windows environment currently has neither `go` nor `make` on PATH

### Absorbed upstream batch `2026-07-08-m`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `f26ca566` `feat: add OpenAI advanced scheduler controls`
  - `728bb1bc` `feat(frontend): 支持账号数据拖拽和批量导入`
  - `83455a3f` `fix(frontend): harden account data batch import`
  - `2d2b6f03` `Update revokeDesc to allow subscription recovery`
  - `ee90246a` `Modify revokeDesc for subscription termination`
  - `6752cdc8` `fix(i18n): 用量页费用明细提示将面向用户的"成本"措辞统一为"费用"`
  - `b408edf9` `fix(payment): convert subscription CNY pay amount`
  - `0fd2e921` `fix(scheduler): 修复 OpenAI 高级调度器审计发现的正确性与性能问题`
  - `8fab6369` `feat: complete batch image workflow`
  - `5553d836` `fix: localize antigravity image mapping labels`
- Scope:
  - Synced OpenAI advanced scheduler controls and audit fixes, account data drag-and-drop/batch import hardening, subscription revoke wording updates, user-facing usage fee terminology, subscription CNY payment conversion, and full batch image workflow
  - Synced Antigravity image mapping label localization
- Merge notes:
  - Resolved `frontend/src/components/admin/account/ImportDataModal.vue` by preserving local import target group selection while absorbing upstream multi-file drag/drop import, per-file validation, ignored-file warnings, and import refresh behavior
  - Resolved `frontend/src/composables/useModelWhitelist.ts` by preserving already-absorbed Antigravity Gemini 3.1 `Pro-Agent` routing semantics while localizing the passthrough label to `透传`
- Validation:
  - Not run locally: this Windows environment currently has neither `go` nor `make` on PATH

### Absorbed upstream batch `2026-07-08-n`

- Source: `Wei-Shaw/sub2api`
- Commits:
  - `a3c5cc45` `test: add batch image QA reports`
  - `d8e96f0f` `fix: bound batch image settlement retries`
  - `202c6989` `fix: preserve immutable initial migration`
  - `d73fa8ea` `fix: keep image generation migration immutable`
  - `d56e94b8` `feat(payment): 订阅 CNY 换算改为独立汇率配置的显式 opt-in`
  - `ba1bb0a3` `Revert "fix(payment): 订阅确认页显示换算后 CNY 金额 + 邀请返利按 USD price 计算"`
  - `bf76168b` `feat: add custom easypay payment methods`
  - `0dc6e56a` `fix: harden easypay custom method validation`
  - `a5a2fea0` `Polish EasyPay custom method UI`
  - `b197ba61` `test: align antigravity mapping preset label`
  - `22ec77b5` `fix: match built-in payment methods exactly`
  - `27cb485d` `fix: share built-in payment method matching`
  - `616cf17d` `fix: hide batch image entry without allowed key`
  - `0b729496` `fix: center batch image empty state`
  - `76bb7b03` `chore: sync VERSION to 0.1.145 [skip ci]`
- Scope:
  - Synced batch image QA reports, settlement retry bound, immutable migration preservation, batch image access/empty-state fixes, and VERSION 0.1.145
  - Synced subscription CNY independent exchange-rate opt-in, reverted the older balance-multiplier based subscription conversion, and added custom EasyPay payment methods with validation/UI polish and shared built-in payment matching
- Merge notes:
  - Resolved payment revert conflicts by preserving the new `subscription_usd_to_cny_rate` opt-in conversion path while ensuring the reverted balance recharge multiplier conversion remains removed
  - Resolved affected backend/frontend payment tests to cover configured-rate conversion, disabled-rate direct price behavior, non-CNY direct price behavior, and fee calculation after conversion
- Validation:
  - Not run locally: this Windows environment currently has neither `go` nor `make` on PATH

### Absorbed upstream batch `2026-07-11-a`

- Source: `Wei-Shaw/sub2api`
- Upstream head at fetch: `e316ebf52838`
- Commits:
  - `5fcbe7e3` `Add response_format compatibility mapping`
  - `e2326a79` `Format response format types`
  - `2a3dcb49` `fix(scheduler): 空 model_mapping 的 OpenAI OAuth 账号不再吸收全部模型`
  - `9b75c7b7` `fix(scheduler): 改用异族厂商前缀黑名单，兼容渠道级模型映射`
  - `fa01aec8` `fix: 前缀全部加连字符，与 thinking_protocol.go 风格对齐`
  - `75fb3c41` `fix(apicompat): responses→chat 桥支持 custom 工具，修复 Codex exec 丢失`
  - `27e29f05` `feat(apicompat): 添加 tool_search 支持，增强工具调用处理逻辑`
  - `13e773ef` `feat: Codex 客户端模型清单（manifest）透传接口`
  - `3866da50` `fix(ratelimit): Anthropic 无 reset 头的 429 也进入兜底冷却`
  - `5aba53d5` `fix(ops): 记录固化 200 SSE 流上的就地错误，修复流内限流不进错误看板`
- Scope:
  - Synced Responses API compatibility for `response_format`, custom/freeform tools, `tool_search`, namespace-style tool bridging, string tool declarations, and Codex manifest passthrough.
  - Synced OpenAI OAuth scheduler model-support hardening so empty mappings no longer absorb all models and cross-provider prefix filtering supports channel-level mappings.
  - Synced Anthropic fallback cooldown behavior for 429 responses without reset headers and ops capture for in-stream SSE errors on finalized 200 responses.
- Merge notes:
  - Resolved `backend/internal/service/openai_gateway_responses_chat_fallback.go` by preserving local `readCCUpstreamJSONResponse`, `newStreamHeaderWriter`, and `scanCCStream` pipeline helpers while wiring upstream `customTools` and `toolSearch` flags into non-stream and streaming Responses conversion.
- Validation:
  - `git diff --check` passed during conflict resolution.
  - Not run locally: this Windows environment currently has neither `go`, `gofmt`, nor `make` on PATH.
