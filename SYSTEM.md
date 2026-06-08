# SYSTEM

## 項目目標

`sub2api-src` 是一套前後端一體的管理系統，主要負責 AI 平台帳號管理、路由調度、健康檢查、OAuth 憑證刷新、定時任務與後台運維。

## 主要結構

- `backend/`
  - `internal/service/`：核心業務邏輯與後台任務。
  - `internal/handler/admin/`：管理端 HTTP Handler。
  - `internal/server/routes/`：管理端與公開路由註冊。
  - `cmd/server/`：依賴注入與啟動入口。
- `frontend/`
  - `src/views/admin/`：後台頁面。
  - `src/api/admin/`：管理端 API 封裝。
  - `src/i18n/locales/`：多語系文案。

## 三層狀態邊界

這個項目後續維護時，必須明確區分以下三層，避免把「Git 倉庫」、「伺服器部署目錄」和「實際運行容器」混為一談：

### 1. 本地 / Git 倉庫層

- 來源：當前工作區 `C:\Users\admin\Desktop\sub2api-deploy\sub2api-src`
- 對外同步：GitHub `main`
- 這一層是功能開發、頁面改動、後端邏輯、部署腳本來源文件的唯一正式來源。

### 2. 伺服器部署目錄層

- 位置：`/home/ec2-user/sub2api-deploy`
- 對外域名：
  - 舊域名：`https://mysuby.duckdns.org/`
  - Cloudflare 新域名：`https://tupai.cyou/`，`https://www.tupai.cyou/`
- DNS 解析：`mysuby.duckdns.org -> 3.22.185.140`（AWS EC2）
- Cloudflare DNS：
  - `tupai.cyou` 使用 A 記錄代理到源站 `3.22.185.140`
  - `www.tupai.cyou` 使用 CNAME 指向 `tupai.cyou`
  - 公共 DNS 會解析到 Cloudflare 邊緣 IP，源站 IP 不直接暴露在瀏覽器解析結果中
- 源站 Nginx：
  - 實際配置：`/etc/nginx/conf.d/sub2api.conf`
  - `server_name` 已包含 `mysuby.duckdns.org tupai.cyou www.tupai.cyou`
  - 80 端口對三個域名統一 `301` 到 `https://$host$request_uri`
  - 443 端口沿用既有反向代理到 `127.0.0.1:8080`
  - 源站證書已切換到 `/etc/letsencrypt/live/tupai.cyou/fullchain.pem`
  - 該 Let’s Encrypt 證書 SAN 包含 `mysuby.duckdns.org`、`tupai.cyou`、`www.tupai.cyou`，可支援 Cloudflare `Full (strict)` 或源站直連校驗
- SSH：`ec2-user@mysuby.duckdns.org`；本機需使用已授權私鑰連線，私鑰路徑不寫入項目文檔。
- 用途：保存伺服器上的 `docker-compose.yml`、`.env`、`bin/deploy-from-package.sh`、資料目錄與運維腳本
- 特性：它不是正式 Git 倉庫主源，只是遠端運行與部署輔助目錄
- 注意：此目錄下還存在一份 `sub2api-src/` 快照，但它不應被視為正式開發主源

### 3. 容器運行層

- 目前主應用容器：`sub2api`
- 目前運行鏡像：`sub2api:rollback`
- 真正對外提供服務的是容器，不是本地 repo，也不是部署目錄裡的源碼快照

## 維護規則

- 業務功能改動：優先修改本地 / Git 倉庫，然後提交推送。
- 伺服器運維腳本如果在遠端熱修：
  - 若屬於應長期保留的能力，必須同步回本地 repo。
  - 若只是一次性排障操作，可只留在伺服器，不必回寫 repo。
- 發布與服務器更新流程：
  - GitHub Releases / 服務器更新只使用固定部署入口 `docker-deploy`。
  - 普通可部署修正不得新建版本 tag；應提交後推送到 `origin/main`，由 `Deploy Package` workflow 更新 `docker-deploy/sub2api-docker-image.tar`。
  - 舊的 `v*` tag / version release 視為歷史遺留；後續不要再新增版本 tag。
- 之後描述問題時必須明確說是哪一層：
  - `Git 倉庫`
  - `伺服器部署目錄`
  - `容器運行狀態`

## 刪除語義

- `accounts` 以前使用 Ent 的 `SoftDeleteMixin`，列表默认只显示 `deleted_at is null` 的有效账号。
- 服务器已执行一次恢复，将已软删除账号重新设为可见。
- 自本次改动起，账号删除改为**物理删除**：
  - 会先删 `account_groups`
  - 再通过 `mixins.SkipSoftDelete(ctx)` 真正删除 `accounts` 行
- 因此前端账号页看到的数量与数据库原始总行数不再因为历史软删除记录而长期背离。

## 本次變更

### 首 Token 延遲排查與線上調度優化

- 排查對象：
  - 管理端使用記錄中 `2026/06/08 08:46:26` 的 `首 Token` 約 `4.79s`、總耗時約 `12.88s` 的請求。
- 查證結論：
  - `first_token_ms` 不是前端渲染或 Nginx 靜態資源壓縮指標，而是後端網關從接收請求到收到第一個上游流式 token 的 TTFT。
  - 目標請求為 `gpt-5.5`、`/v1/responses`、stream 請求，`reasoning_effort=high`。
  - 同時間段 `gpt-5.5` 請求首 Token 在不同帳號間差異明顯，慢帳號平均 TTFT 明顯高於快帳號。
  - 服務器資料庫中 `settings.openai_advanced_scheduler_enabled` 原為 `false`，因此線上未使用包含 TTFT / 錯誤率 / 負載 / 排隊數的高級 OpenAI 帳號調度。
- 已執行的線上變更：
  - 將服務器 `settings.openai_advanced_scheduler_enabled` 更新為 `true`。
  - 該設定由後端以約 5 秒快取讀取，無需重啟容器即可對新請求生效。
- 變更前後觀察：
  - 開啟前窗口：`58` 個樣本，平均 TTFT 約 `4966ms`，P50 約 `3770ms`，P90 約 `6427ms`，最大約 `22505ms`。
  - 開啟後窗口：`19` 個樣本，平均 TTFT 約 `3406ms`，P50 約 `3232ms`，P90 約 `5682ms`，最大約 `6209ms`。
  - 開啟後候選帳號切換到較快帳號池，短窗口內首 Token 延遲已有下降。
- 後續建議：
  - 若仍出現長 TTFT，優先檢查該筆請求的 `reasoning_effort`、模型、帳號 ID、是否觸發 429 / 測活背景任務。
  - 若要更精準定位，可後續新增拆分指標，例如 gateway preflight、upstream connect、upstream first token，避免單一 `first_token_ms` 把本地處理與上游等待混在一起。

### 自動刷新 Token 功能

- 在測活管理頁整合新的 `Token` 自動刷新 tab。
- 新增後端設定模型 `account_token_auto_refresh_config`，使用 DB setting 存放：
  - 是否啟用
  - 刷新週期數值
  - 刷新週期單位（小時 / 天）
  - 每批刷新數量
  - 上次執行時間與上次執行統計
- 新增管理端設定接口：
  - `GET /api/v1/admin/settings/account-token-auto-refresh`
  - `PUT /api/v1/admin/settings/account-token-auto-refresh`
- 定時執行整合到既有 `ScheduledTestRunnerService`：
  - 每分鐘巡檢一次是否到達刷新時機
  - 處理所有可刷新的 OAuth 帳號，不因異常狀態自動排除；僅跳過明確停用帳號
  - 依設定的 batch size 分批刷新
  - 批次間加入短暫間隔，避免瞬時大量請求
- 補充手動觸發入口：
  - `POST /api/v1/admin/settings/account-token-auto-refresh/run`
  - 前端 token tab 可直接手動执行一轮刷新
  - 手动执行现改为后台异步任务，避免 HTTP 约 31 秒超时导致 `context canceled`
  - 页面通过既有轮询在任务完成后自动更新最近刷新统计
- 刷新範圍現已支援：
  - `全部账号`
  - `指定分组`
  - 設定會一併作用於定時刷新與手動刷新
- 背景刷新進度會持續回寫到設定：
  - `running`
  - `current_total`
  - `current_success`
  - `current_failed`
  - 前端會透過輪詢直接顯示目前進度
- 刷新范围说明：
  - 数据库总账号数可包含软删除记录
  - 后台账号列表默认展示 `deleted_at is null` 的有效账号
  - token 刷新针对“支持 refresh token 的 OAuth 账号”，不等于数据库原始总行数
- OAuth 刷新底層新增 `RefreshNow` 路徑，沿用既有分布式鎖與 DB 重讀保護，避免與其他刷新路徑競爭。

### 測活分批執行

- 手動測活與排程自動測活都改為分批執行：
  - 每批 10 個帳號
  - 批次間停 2 秒
  - 批內保留有限並發，避免大量帳號時瞬間壓太多請求到上游
- 此改動不改變測活篩選邏輯，只降低大批量測活的瞬時壓力。
- 測活狀態現已補充背景進度欄位：
  - `running`
  - `current_total`
  - `current_success`
  - `current_failed`
- 手動測活入口已改為背景任務，前端會透過輪詢顯示進度，不再等整輪完成才返回。
- 測活成功後的恢復語義：
  - `error` 账号會自動恢復為 `active`
  - `disabled` 账号也會自動恢復為 `active`
  - 適用於你這裡把異常/停用帳號也納入測活的場景
- 測活頁與 Token 刷新頁在任務執行中會把輪詢頻率提升到 `2s`，閒置時維持 `15s`。
- 頁面新增 `已完成 / 未完成` 進度卡塊，讓每完成一批後的變化能直接看到。
- 後端背景任務現加入共享並發限流：
  - 測活與 Token 刷新共用背景任務槽位
  - 預設最多同時 `4` 個高成本背景工作
  - 目標是避免背景任務把前台頁面查詢、切頁等一般請求一起拖慢
- Token 刷新批次執行已改為固定 worker pool，而不是每批直接按帳號數全量並發。
- 測活與 Token 刷新現再進一步共用一把「全局維護任務鎖」：
  - 同一時刻只允許一個大任務執行
  - `自動測活 / 手動測活 / 自動刷新 / 手動刷新` 共用同一條維護任務隊列
  - 目前策略：
    - 正在執行 1 個
    - 待執行 1 個
    - 同類型新任務會自動合併，不重複排隊
  - 目標是避免測活與刷新互相打架，同時又不讓後來的操作白點。

### 服務器部署腳本同步

- 伺服器實際使用的包部署腳本是：
  - `/home/ec2-user/sub2api-deploy/bin/deploy-from-package.sh`
- 專案內對應來源腳本是：
  - `deploy/host-agent/deploy-from-package.sh`
- 為避免「伺服器已手改、本地 repo 未同步」造成後續覆蓋，本地腳本已同步補上：
  - 部署後健康檢查等待
  - 舊 backup image 自動保留最近 1 個
  - 部署完成後輸出 image/container 結果摘要
- 部署腳本現在具備 no-op 短路：
  - 先下載並 `docker load` 最新 release 包，再比較 `LOADED_IMAGE` 與當前容器 image digest。
  - 若兩者一致，則判定已是最新版並跳過 `tag/compose`，避免把新 release 包錯誤擋掉。
- 部署腳本已進一步加入 release asset `ETag` 緩存：
  - 若當前容器 image digest 已匹配，且遠端 release 包 `ETag` 與本地緩存一致，則直接跳過下載與部署。
  - 緩存文件位置：`/home/ec2-user/sub2api-deploy/.deploy-state/archive.etag`
- 部署成功或 no-op 後會執行 Docker 未使用鏡像清理：
  - `docker image prune -a -f`
  - 用於移除無用舊鏡像層，降低磁碟佔用
- 管理後台系統設置頁的部署狀態現已補充最近部署輸出展示：
  - 後端 `DeployState.last_output`
  - 前端系統設置部署區塊可直接查看最近一輪部署輸出
- 部署狀態另補充鏡像版本對比字段：
  - `requested_image_id`
  - `running_image_id`
  - 前端以簡潔文案提示是否已切換到目標版本

## 涉及文件

- 後端
  - `backend/internal/service/account_token_auto_refresh_config.go`
  - `backend/internal/service/account_token_auto_refresh_runner.go`
  - `backend/internal/service/oauth_refresh_api.go`
  - `backend/internal/service/scheduled_test_runner_service.go`
  - `backend/internal/service/token_refresh_service.go`
  - `backend/internal/service/domain_constants.go`
  - `backend/internal/service/wire.go`
  - `backend/internal/handler/admin/setting_handler.go`
  - `backend/internal/server/routes/admin.go`
  - `backend/cmd/server/wire_gen.go`
- 前端
  - `frontend/src/views/admin/AccountHealthView.vue`
  - `frontend/src/components/admin/account-health/AccountTokenAutoRefreshPanel.vue`
  - `frontend/src/api/admin/accounts.ts`
  - `frontend/src/i18n/locales/zh.ts`
  - `frontend/src/i18n/locales/en.ts`

## 已知驗證狀態

- 前端 `pnpm typecheck` 已通過。
- 目前工作環境缺少 `go` 指令，因此尚未在本機執行後端 `go test` / `go build`。

## 本次吸收上游更新

- 已安全吸收上游 `Wei-Shaw/sub2api` 的 `gpt-5.5` 支援更新。
- 修改內容：
  - `backend/internal/pkg/openai/constants.go`
  - `backend/internal/service/billing_service.go`
  - `backend/internal/service/openai_codex_transform.go`
  - `backend/internal/service/pricing_service.go`
  - `frontend/src/composables/useModelWhitelist.ts`
  - `frontend/src/components/keys/UseKeyModal.vue`
- 修改前後差異：
  - 修改前：本倉庫未內建 `gpt-5.5`，預設模型列表、Codex 正規化、白名單與 OpenCode 配置均無此型號。
  - 修改後：新增 `gpt-5.5` 顯示與識別，並讓定價回退沿用 `gpt-5.4`，降低新模型接入時的失敗風險。
- 影響範圍：
  - 只擴充模型支援與計費回退邏輯，不涉及資料庫、部署流程或權限模型。

## 本次續吸收上游修補

- 已追加吸收上游 `codex-spark` / `compact` 相關的低風險修補。
- 修改內容：
  - 顯式 `gpt-5.3-codex-spark` 不再被組級默認映射覆蓋
  - Spark 模型保留自身歸一化結果，不再直接降為普通 `gpt-5.3-codex`
  - 對 Spark 圖片輸入新增後端請求校驗，直接返回 `400 invalid_request_error`
  - Codex OAuth transform 會補充 Spark 不支援圖片能力的說明文案
  - Spark 計費回退沿用既有 `gpt-5.3-codex` fallback
- 影響範圍：
  - 只收斂模型映射與輸入驗證行為，避免 Spark 在圖片場景被錯誤透傳到上游。

## 本次前端表格優化

- 已優化管理端账号表格視覺與列寬：
  - 名稱列改為固定寬度並對長內容截斷
  - 账号表主要列補充明確寬度，降低內容互相擠壓
  - 通用 `DataTable` 增加更穩定的邊框、陰影、表頭與 sticky 列陰影
  - 操作列改為橫向低高度按鈕組，減少右側固定列的視覺擁擠
- 影響範圍：
  - 僅調整前端展示樣式與列寬，不改動 API、資料結構或帳號操作流程。

## 本次前端工具列修正

- 已修正管理端账号頁頂部操作按鈕被擠壓成直排的問題。
- 修改內容：
  - `frontend/src/views/admin/AccountsView.vue`
  - `frontend/src/components/admin/account/AccountTableActions.vue`
- 修改前後差異：
  - 修改前：操作按鈕與篩選器共用左右擠壓布局，視窗寬度不足時「自動刷新」「列設置」等按鈕文字會被壓成直排。
  - 修改後：頂部控制區改為上下兩行，第一行只放操作按鈕，第二行放搜尋與篩選；按鈕增加 `shrink-0` 與 `whitespace-nowrap`，保證文字不被壓縮換行。
- 影響範圍：
  - 僅調整前端布局與樣式，不改動 API、資料結構、帳號操作流程或權限邏輯。

## 本次前端工具列二次收斂

- 已進一步優化管理端账号頁頂部控制區的視覺密度。
- 修改內容：
  - `frontend/src/views/admin/AccountsView.vue`
  - `frontend/src/components/admin/account/AccountTableActions.vue`
  - `frontend/src/components/admin/account/AccountTableFilters.vue`
- 修改前後差異：
  - 修改前：操作按鈕雖不再直排，但高度、圓角、字級與左右 padding 仍偏大，第一行看起來擁擠。
  - 修改後：账号頁工具按鈕局部收斂為 `h-8`、小字級、較小圓角與緊湊間距；篩選器局部縮小高度與固定寬度，形成更穩定的兩行工具區。
- 影響範圍：
  - 僅影響账号管理頁的局部展示樣式，不改全站通用控件預設尺寸，也不改業務流程。

## 本次前端工具列分組修正

- 已確認線上服務器正在運行最新提交，但第一行操作按鈕仍因所有功能平鋪在同一層而顯得擁擠。
- 修改內容：
  - `frontend/src/components/admin/account/AccountTableActions.vue`
  - `frontend/src/views/admin/AccountsView.vue`
- 修改前後差異：
  - 修改前：刷新、自動刷新、配置、同步、導入導出、清理與新增等按鈕共用同一條 flex 流，視覺上缺少功能分段。
  - 修改後：工具列拆成兩個低高度分組容器，刷新/配置類控制與資料操作類控制分開排列；按鈕高度進一步收斂為 `h-7`，減少第一行壓迫感。
- 影響範圍：
  - 僅影響账号管理頁頂部操作區的展示，不改 API、資料流或帳號操作邏輯。

## 本次 GitHub Release 發布收斂

- 已修正 GitHub Actions 會同時生成版本 release 與固定部署 release 的問題。
- 修改內容：
  - `.github/workflows/deploy-package.yml`
  - `.github/workflows/release.yml`
- 修改前後差異：
  - 修改前：推送 `v*` tag 時，`deploy-package.yml` 會發布固定 `docker-deploy` release，並額外把版本化部署包掛到 `v*` release；`release.yml` 也會在 `v*` tag 上自動建立正式版本 release。
  - 修改後：`deploy-package.yml` 只在 `main` 推送或手動執行時更新固定 `docker-deploy` release；不再生成版本化部署資產。`release.yml` 不再由 `v*` tag 自動觸發，只保留手動發布入口。
- 影響範圍：
  - 後續自動部署包只會更新 `docker-deploy` 這個固定 release，避免 GitHub Releases 頁面同時出現 `v0.x.x` 與 `docker-deploy` 兩個條目。
  - 已存在的舊 `v*` release/tag 不會被本次代碼修改自動刪除；若需要清理，需要單獨對 GitHub release/tag 執行刪除操作。

## 本次 GitHub Actions 失敗修正

- 已針對 `main` 推送後的 CI / Security Scan 紅叉補修編譯與前端 lint 問題。
- 修改內容：
  - `backend/internal/handler/dto/types.go`
  - `backend/internal/service/openai_gateway_service.go`
  - `frontend/src/components/admin/account-health/AccountHealthAutoCheckPanel.vue`
  - `frontend/src/components/admin/account-health/AccountTokenAutoRefreshPanel.vue`
  - `frontend/src/views/admin/AccountHealthView.vue`
- 修改前後差異：
  - 修改前：`account_handler.go` 會寫入 `dto.Account` 的健康狀態欄位，但 DTO 未定義對應字段，導致後端編譯失敗。
  - 修改前：`openai_gateway_service.go` 合併後殘留重複 `isCodexImageGenerationBridgeEnabled` 與錯誤的 `GetAPIKeyFromContext`/重複變數宣告，導致後端編譯失敗。
  - 修改前：健康檢查與 Token 自動刷新面板直接 `v-model` 修改 props，觸發 `vue/no-mutating-props`。
  - 修改後：DTO 補齊健康狀態輸出欄位；OpenAI gateway 刪除重複/錯誤宣告；前端子組件改為 emit 更新事件，由父頁統一更新 reactive 狀態。
- 影響範圍：
  - 僅修正編譯與 lint 阻塞，不改變健康檢查、Token 自動刷新或 OpenAI gateway 既有業務語義。
  - 本機已通過前端 `lint:check` 與 `typecheck`；本機環境缺少 `go` / `gofmt`，後端編譯需依 GitHub Actions 最終驗證。

## 本次後端 CI 二次失敗修正

- 已針對最新 `main` 後端 CI 失敗補齊合併後遺漏的 handler 與依賴注入接線。
- 修改內容：
  - `backend/internal/handler/admin/setting_auto_account_handler.go`
  - `backend/internal/handler/admin/setting_handler.go`
  - `backend/internal/handler/admin/system_handler.go`
  - `backend/internal/handler/admin/system_handler_test.go`
  - `backend/internal/handler/wire.go`
  - `backend/cmd/server/wire_gen.go`
  - 多個 `NewAccountHandler` 單元測試呼叫點
- 修改前後差異：
  - 修改前：路由已指向 `SettingHandler` / `SystemHandler`，但對應的帳號健康自動檢查、Token 自動刷新與部署配置查詢/更新/狀態/觸發 handler 方法缺失，導致後端編譯失敗。
  - 修改前：`NewAccountHandler` 新增 `settingService` 參數後，`wire_gen.go` 與測試仍使用舊參數位置；`ProvideTokenRefreshService`、`ProvideUpdateService`、`ProvideScheduledTestRunnerService` 生成呼叫也缺少新版依賴注入參數。
  - 修改後：新增 `setting_auto_account_handler.go` 承接 `/admin/settings/account-health-auto-check`、`/admin/settings/account-token-auto-refresh` 與手動刷新路由；`SystemHandler` 補上部署配置、部署狀態與觸發部署路由；依賴注入與測試呼叫點全部對齊新版建構子。
- 影響範圍：
  - 僅修正後端編譯與路由接線，不改資料庫 schema、不改 API path。
  - 本機仍缺少 `go` / `gofmt`，後端實際編譯、測試與 golangci-lint 需由 GitHub Actions 驗證。

## 本次前端工具列回退與刷新按鈕修正

- 已根據反饋將账号頁第一行工具列退回上一版分組排列。
- 已追加修正自動刷新與列設定按鈕的桌面寬度。
- 修改內容：
  - `frontend/src/components/admin/account/AccountTableActions.vue`
  - `frontend/src/views/admin/AccountsView.vue`
- 修改前後差異：
  - 修改前：低頻操作被收進「更多」下拉，與使用者期望的上一版直出操作不一致。
  - 修改後：恢復同步、匯入、匯出、去重、錯誤透傳、TLS 指紋等按鈕直出；僅將刷新類純圖示按鈕寬度從 `w-7` 提升到 `w-8`，修正刷新按鈕寬度不足。
  - 追加修正：自動刷新與列設定按鈕不再套用固定寬度 icon class；手機維持 32px icon 寬，桌面改回依文字內容自動寬度並帶左右內邊距。
- 影響範圍：
  - 僅調整账号管理頁第一行工具列展示，不改動功能、API 與資料流。

## 本次前端刷新白屏修正

- 已定位線上管理頁刷新偶發白屏的後端原因：
  - `index.html` 會注入帶 CSP nonce 的公開配置腳本。
  - 舊邏輯在 `If-None-Match` 命中時返回 `304 Not Modified`，瀏覽器會復用本地舊 HTML body，但套用本次回應的新 CSP header。
  - 這會導致 HTML 裡的舊 nonce 與新 CSP nonce 不一致，內聯配置腳本可能被瀏覽器阻擋，刷新時出現白屏或初始化異常。
- 修改內容：
  - `backend/internal/web/embed_on.go`
  - `backend/internal/web/embed_test.go`
- 修改前後差異：
  - 修改前：HTML 命中 ETag 時可能返回 `304`。
  - 修改後：HTML 即使帶 `If-None-Match` 也會返回 `200` 並重新替換本次 nonce，同時 `Cache-Control` 改為 `no-store`，避免 nonce HTML 被瀏覽器重用。
- 影響範圍：
  - 僅影響嵌入式前端 `index.html` 的快取策略。
  - 靜態 JS/CSS 資源、API 行為、資料結構與權限邏輯不變。

## 本次服務器 Nginx 靜態資源快取止血

- 背景：
  - 服務器部署文檔中的 Nginx 注意事項目前只提示 `underscores_in_headers on;`，不包含完整的靜態資源壓縮與瀏覽器快取配置。
  - `deploy/Caddyfile` 有 `encode` 與 `/assets/*` 長快取示例，但 Docker 部署本身只是把應用暴露到 `8080`；如果服務器前置代理實際使用 Nginx 或直接訪問 `8080`，Caddyfile 中的壓縮與快取策略不會自動生效。
  - 2026-06-07 外部檢查確認 `https://mysuby.duckdns.org/` 前置代理為 `nginx/1.28.3`。
  - 線上真實資源如 `/assets/index-krrpSYhQ.js`、`/assets/vendor-vue-Oirx9HYx.js`、`/assets/index-DdNla0ru.css` 返回 `Content-Length`，但未返回 `Content-Encoding` 或 `Cache-Control`，因此瀏覽器重新請求時會重新下載大 JS/CSS。
- 修改內容：
  - 服務器 Nginx 站點配置：`/etc/nginx/conf.d/sub2api.conf`
  - 本地只更新 `SYSTEM.md` 記錄；沒有保留後端代碼改動，避免與 Nginx 職責重複。
- 修改前後差異：
  - 修改前：Nginx 只有一個 `location /` 反代到 `127.0.0.1:8080`，未啟用 gzip，也沒有 `/assets/` 專用長快取。
  - 修改後：Nginx 對 JS/CSS 等文本資源啟用 gzip，並為 `/assets/` 返回 `Cache-Control: public, max-age=31536000, immutable`。
- 架構與行為影響：
  - 僅影響服務器 Nginx 反向代理層的靜態資源壓縮與瀏覽器快取 header，不修改 API、資料庫、權限或前端打包流程。
  - `index.html` 仍維持 `no-store`，避免 CSP nonce HTML 被瀏覽器復用。
- 驗證狀態：
  - 使用本機已授權私鑰成功連線 `ec2-user@mysuby.duckdns.org`；主機名為 `ip-172-31-45-171.us-east-2.compute.internal`。
- 服務器 Nginx 熱修：
  - 實際站點配置：`/etc/nginx/conf.d/sub2api.conf`
  - Nginx 服務為 `active`，Caddy 為 `inactive`；容器 `sub2api` 運行鏡像為 `sub2api:rollback`。
  - 熱修過程曾短暫生成兩個精確配置備份，使用者要求不要保留備份後已刪除：
    - `/etc/nginx/conf.d/sub2api.conf.bak.20260607143646`
    - `/etc/nginx/conf.d/sub2api.conf.bak.20260607143808`
  - 已加入：
    - `underscores_in_headers on;`
    - `gzip on; gzip_comp_level 6; gzip_min_length 256; gzip_vary on;`
    - `gzip_types text/plain text/css text/javascript application/javascript application/json application/xml image/svg+xml;`
    - `location /assets/` 專用反代，啟用 `proxy_buffering on` 並返回 `Cache-Control: public, max-age=31536000, immutable`
  - 已執行 `sudo nginx -t`，配置測試成功；已 `sudo systemctl reload nginx`。
  - 外部驗證：
    - `https://mysuby.duckdns.org/assets/index-krrpSYhQ.js`
    - `https://mysuby.duckdns.org/assets/index-DdNla0ru.css`
    - 均已返回 `Content-Encoding: gzip` 與 `Cache-Control: public, max-age=31536000, immutable`。
  - 首頁 `https://mysuby.duckdns.org/` 仍返回 `Cache-Control: no-store`，符合 CSP nonce HTML 防白屏策略。

## 本次吸收上游生圖橋接開關

- 已安全吸收上游與 `Codex` 圖片生成橋接直接相關的最小改動。
- 修改內容：
  - `backend/internal/config/config.go`
  - `backend/internal/service/codex_image_generation_bridge.go`
  - `backend/internal/service/openai_codex_transform.go`
  - `backend/internal/service/openai_gateway_service.go`
  - `frontend/src/components/account/EditAccountModal.vue`
  - `frontend/src/i18n/locales/zh.ts`
  - `frontend/src/i18n/locales/en.ts`
- 修改前後差異：
  - 修改前：本地已包含 `gpt-image-1` 系列定價資料，但沒有顯式的 `Codex` 圖片生成橋接全局開關，也沒有帳號級覆蓋開關。
  - 修改後：新增 `gateway.codex_image_generation_bridge_enabled` 全局配置，並支持帳號級 `extra.codex_image_generation_bridge` 覆蓋；當橋接開關啟用且請求已帶 `image_generation` 工具時，`Codex /responses` 會補充圖片橋接提示指令，降低客戶端因本地不暴露 `image_gen` 命名空間而誤判不可生圖的情況。
- 影響範圍：
  - 僅影響 OpenAI / Codex `/responses` 圖片橋接開關與帳號編輯 UI，不包含上游後續的大批量圖片並發治理、Channel 級圖片治理與圖片輸出計費整流。

## 本次試合併上游 v0.1.133

- 已在隔離分支 `absorb-upstream-v0.1.133` 合併遠端 `upstream/main`，上游版本點為 `v0.1.133` 附近。
- 修改內容：
  - 大量吸收上游後端、前端、資料庫 migration、部署配置、依賴與文檔更新。
  - 主要新增/變更範圍包含 user-platform quota、channel monitor、risk control、OpenAI embeddings/images gateway、OAuth/支付/用量統計修復與前端管理頁更新。
  - 衝突檔案以 `upstream/main` 版本為主解決；`SYSTEM.md` 按工作區規範保留並補充本紀錄。
- 修改前後差異：
  - 修改前：本地 `main` 落後上游數百個提交，僅有本地部署、健康檢查與 Codex 圖片橋接等分支改動。
  - 修改後：工作樹基於上游最新大版本變更，並保留專案級 `SYSTEM.md` 維護文檔；本地若曾在衝突檔案中覆蓋上游邏輯，本次暫以新版上游實作為準。
- 架構影響：
  - 引入多批 migration，正式部署前必須先備份資料庫並在測試環境驗證升級路徑。
  - 後端依賴注入、服務層、repository、handler 與前端路由/頁面均有大幅變更，屬於中大型上游吸收，不應未驗證直接部署生產。
- 待驗證事項：
  - 後端 `go test ./...` 或至少 `go test ./cmd/server ./internal/service ./internal/handler/...`。
  - 前端 `pnpm --dir frontend run typecheck` 與 `pnpm --dir frontend run build`。
  - Docker image build 與一次乾跑容器啟動。

## 本次後台導航卡頓止血

- 已針對後台左側導航切換卡頓做低風險前端修正。
- 修改內容：
  - `frontend/src/App.vue`
  - `frontend/src/utils/usageLoadQueue.ts`
- 修改前後差異：
  - 修改前：後台路由每次切換都銷毀並重建頁面實例，已載入過的管理頁返回時仍會重新執行 `onMounted`、重新拉資料並重渲染表格。
  - 修改後：`/admin` 下的路由使用 `KeepAlive` 快取最多 12 個頁面實例，已訪問頁面切回時保留組件狀態與表格狀態，減少重建成本。
  - 修改前：帳號頁多個 `AccountUsageCell` 掛載時會同步發起所有 `/admin/accounts/{id}/usage` 請求。
  - 修改後：usage 請求通過全局小隊列限制為最多 4 個並發，避免切入帳號頁或恢復頁面時瞬間打滿瀏覽器與後端連線。
- 影響範圍：
  - 僅影響前端後台頁面切換體感與帳號 usage 請求節流，不修改後端 API、資料庫 schema、權限或部署流程。
- 注意事項：
  - `KeepAlive` 會保留頁面資料狀態，切回頁面時不一定即時重新拉最新資料；需要最新資料時仍可點頁面內刷新按鈕。

## 本次最新版本渠道限制 503 修正

- 背景：
  - 線上最新 `docker-deploy` 版本在使用 API 密鑰請求 `/responses` 時會返回 503 provider error。
  - 日誌顯示實際錯誤為 `no available accounts supporting model: gpt-5.5 (channel pricing restriction)`，不是上游 provider 真正不可用。
  - 線上資料庫中 group 2 關聯渠道 `codex` 的 `restrict_models=true`，但 `channel_model_pricing` 沒有任何模型定價行；最新程式碼把空定價列表解讀為 deny-all，導致選帳號前直接阻擋。
- 修改內容：
  - `backend/internal/service/channel_service.go`
  - `backend/internal/service/gateway_channel_restriction_test.go`
- 修改前後差異：
  - 修改前：`ChannelService.IsModelRestricted` 只要 `restrict_models=true` 且找不到匹配 pricing，就返回 restricted；當 pricing allowlist 為空時會禁止所有模型。
  - 修改後：新增 `pricingAllowlistByGroup` 快取與 `hasPricingAllowlistForGroupPlatform` O(1) 判定，只有當當前分組/平台實際配置了定價 allowlist 時才執行模型限制；空 pricing 不再等於 deny-all。
  - 新增回歸測試覆蓋 `checkChannelPricingRestriction` 與 `isUpstreamModelRestrictedByChannel` 在空 pricing allowlist 下不阻擋 `gpt-5.5`；相關渠道限制測試統一以 `gpt-5.5` 作為請求模型，不在 allowlist 的對照模型使用 `gpt-5.4-mini`。
- 影響範圍：
  - 僅改變渠道模型限制判定的空列表語義。
  - 已配置 pricing allowlist 的渠道仍維持原有限制行為；未配置 pricing 的渠道會回退到全局定價/帳號能力判定，不會在選帳號前被渠道定價擋掉。
- 待驗證事項：
  - 已嘗試執行 `go test -tags unit ./internal/service -run 'TestBillingModelForRestriction|TestResolveAccountUpstreamModel|TestCheckChannelPricingRestriction|TestIsUpstreamModelRestrictedByChannel'`；本機因 `proxy.golang.org` 依賴下載逾時未能完成，需以 GitHub Actions 完整構建結果作為發布驗證。
  - 部署前仍需在非生產環境驗證最新映像可正常處理 group 2 / `gpt-5.5` 的 `/responses` 請求。

## 0.1.134 發布準備

- 修改內容：
  - `backend/cmd/server/VERSION`
- 修改前後差異：
  - 版本號由 `0.1.133` 升至 `0.1.134`，用於本次渠道空 pricing allowlist 503 修正發布。
- 影響範圍：
  - 僅影響構建版本標識與發布 tag 對應；不改變運行邏輯。
- 驗證狀態：
  - 已執行 `gofmt` 格式化本次 Go 改動。
  - 本地 Go 測試受依賴下載網路逾時阻塞，推送後需查看 CI / deploy-package workflow 結果。

## 账号健康页 i18n 热修

- 背景：
  - 部署 `0.1.134` 后，后台账号健康页面显示 `admin.accounts.*` / `admin.accountHealth.*` 原始 key，页面文案不可读。
  - 根因是 `AccountHealthView`、`AccountHealthAutoCheckPanel`、`AccountTokenAutoRefreshPanel` 引用了新增 i18n key，但 `zh.ts` / `en.ts` 未补齐对应翻译。
- 修改内容：
  - `frontend/src/i18n/locales/zh.ts`
  - `frontend/src/i18n/locales/en.ts`
- 修改前后差异：
  - 修改前：账号健康页标题、摘要卡片、自动检查面板、Token 刷新面板均可能裸露翻译 key。
  - 修改后：补齐 `admin.accountHealth`、`admin.accounts.healthSummary`、`admin.accounts.tokenRefresh`、队列状态、健康检查与删除异常账号相关文案。
- 影响范围：
  - 仅影响前端国际化文案，不修改 API、权限、数据结构或后端调度逻辑。
- 验证状态：
  - 已执行 `pnpm --dir frontend build`，`vue-tsc -b && vite build` 通过。

## 本次后台导航切换错位修正

- 背景：
  - 后台页面切换时，中间内容区域偶发与左侧当前导航不一致，表现为标题/模块已经切到新页面，但表格或主体内容仍像上一页，体感上还有卡顿。
  - 左侧导航里的账号健康入口显示 `nav.accountHealth`，说明导航文案缺少对应 i18n key。
- 修改内容：
  - `frontend/src/App.vue`
  - `frontend/src/i18n/locales/zh.ts`
  - `frontend/src/i18n/locales/en.ts`
- 修改前后差异：
  - 修改前：`/admin` 路由统一进入 `KeepAlive`，但 route view component 没有显式 `key`，多个后台页在缓存和快速切换时可能出现实例复用边界不清，造成内容区域与当前路由观感不一致。
  - 修改后：路由视图统一使用 `route.path` 作为组件 `key`，确保每个后台页面都有独立缓存实例；同一路径 query 变化不强制重建整页。
  - 修改前：侧边栏使用 `t('nav.accountHealth')`，但中英文 `nav` 均未定义该 key。
  - 修改后：补齐 `nav.accountHealth` 中英文文案，左侧导航显示为账号健康 / Account Health。
- 影响范围：
  - 仅影响前端路由视图缓存边界与侧边栏文案。
  - 不修改 API、权限、后端逻辑、数据结构或部署流程。
- 验证状态：
  - 已执行 `pnpm --dir frontend run typecheck`，通过。
  - 已执行 `pnpm --dir frontend run build`，通过；仅保留既有 chunk size / dynamic import 构建警告。
  - 已启动本地前端 `http://127.0.0.1:5173/` 并用浏览器访问 `/admin/account-health`；未登录状态会按预期跳转登录页，因此未能直接进入后台侧栏做人工点击验证。

## 本次设置页固定部署入口恢复

- 背景：
  - 后台系统设置页原有的服务器更新 / 部署入口在上游合并后的 `SettingsView.vue` 中不再显示。
  - 排查确认后端路由仍存在：`/admin/system/deploy-config`、`/admin/system/deploy-status`、`/admin/system/deploy`、`/admin/system/update`。
  - 前端 API 封装 `frontend/src/api/admin/system.ts` 也仍保留部署配置、状态查询与触发部署方法；缺失的是设置页 UI tab 挂载。
- 修改内容：
  - `frontend/src/views/admin/SettingsView.vue`
  - `frontend/src/i18n/locales/zh.ts`
  - `frontend/src/i18n/locales/en.ts`
  - `SYSTEM.md`
- 修改前后差异：
  - 修改前：设置页 tab 只有通用、登录条款、功能开关、安全认证、用户默认值、网关服务、支付、邮件、备份，没有部署更新入口。
  - 修改后：新增 `部署更新 / Deploy` tab，恢复固定 `docker-deploy` 部署包的配置、状态查看、演练与立即更新操作。
  - 页面文案明确说明服务器更新只使用固定 `docker-deploy` 包，不创建版本 tag。
- 影响范围：
  - 仅恢复前端设置页入口并调用既有后端 API。
  - 不新增后端路由、不修改部署执行逻辑、不创建版本 tag。
- 验证状态：
  - 已执行 `pnpm --dir frontend run typecheck`，通过。
  - 已执行 `pnpm --dir frontend run build`，通过；仅保留既有 chunk size / dynamic import 构建警告。

## 本次固定部署参数修正

- 背景：
  - 服务器实际 `docker-compose.yml` 中 `sub2api` 服务使用的运行镜像 tag 是 `sub2api:rollback`。
  - 设置页恢复部署入口后，默认 `Runtime Image` 仍显示 `weishaw/sub2api:latest`，会导致后台更新时把固定包 tag 到错误镜像，容器可能继续使用旧 `sub2api:rollback`。
  - 因为本次先通过宿主机脚本直接部署，后台数据库中的 `system_deploy_state` 仍停留在 `pending`。
- 修改内容：
  - `backend/internal/service/update_deploy.go`
  - `frontend/src/views/admin/SettingsView.vue`
  - `SYSTEM.md`
- 修改前后差异：
  - 修改前：后端部署配置默认值与前端默认表单均使用 `weishaw/sub2api:latest`。
  - 修改后：默认运行镜像统一改为服务器实际使用的 `sub2api:rollback`；固定包仍从 `docker-deploy/sub2api-docker-image.tar` 拉取，加载镜像仍为 `sub2api-gha:docker-deploy`。
  - 服务器数据库已手动校正 `system_deploy_config.default_image=sub2api:rollback`，并将 `system_deploy_state.status` 从 `pending` 校正为 `succeeded`。
- 影响范围：
  - 仅修正固定部署默认参数与文档，不改变部署包地址、不创建版本 tag。
- 验证状态：
  - 已执行 `pnpm --dir frontend run typecheck`，通过。
  - 已执行 `pnpm --dir frontend run build`，通过；仅保留既有 chunk size / dynamic import 构建警告。
  - 本机缺少 `gofmt` / `go`，未能运行 Go 格式化与后端测试；本次 Go 改动只调整字符串默认值。

## 本次固定部署已最新提示修正

- 背景：
  - 宿主部署脚本已经具备镜像 digest / release ETag 的 no-op 短路，但后端状态与前端提示没有结构化字段，只能依赖最近输出中的文字判断。
  - 宿主脚本默认 `IMAGE_TAG` 仍是历史 `weishaw/sub2api:latest`，与当前服务器实际运行镜像 `sub2api:rollback` 不一致。
- 修改内容：
  - `deploy/host-agent/deploy-from-package.sh`
  - `deploy/host-agent/sub2api_host_deploy_agent.py`
  - `backend/internal/service/update_deploy.go`
  - `frontend/src/api/admin/system.ts`
  - `frontend/src/views/admin/SettingsView.vue`
- 修改前后差异：
  - 修改前：已是最新时脚本会跳过部署，但 agent / 后端 / 前端没有统一的 `already_up_to_date` 状态；后台只能显示普通成功消息。
  - 修改后：agent 在输出包含 `already up to date` 时返回 `already_up_to_date=true`；后端保存到 `DeployState` 并回传到 `DeployResult`；设置页会用明确提示告知“当前运行镜像已是最新，已跳过更新”。
  - 宿主脚本默认 `IMAGE_TAG` 调整为 `sub2api:rollback`，与后端默认配置和服务器实际 compose 镜像保持一致。
- 影响范围：
  - 仅影响固定部署入口的状态展示与宿主脚本默认参数。
  - 不改变 release 包地址、docker load / tag / compose up 的实际更新流程，也不会创建版本 tag。
- 待验证事项：
  - 需要在具备 Docker / 宿主 agent 的服务器环境点击后台“立即更新”，确认已最新时返回 no-op 提示且不重启容器。

## 本次固定部署中文状态提示补强

- 背景：
  - 后台设置页已能识别 `already_up_to_date`，但状态区仍会直接显示后端英文 `succeeded` / `Already up to date`，容易让操作者误判是否真的更新完成。
  - 使用者确认需要中文提示，并关心旧镜像是否会自动清理。
- 修改内容：
  - `frontend/src/views/admin/SettingsView.vue`
  - `SYSTEM.md`
- 修改前后差异：
  - 修改前：部署状态显示原始英文状态值，最近消息直接显示后端返回内容。
  - 修改后：部署状态在中文界面显示为“成功 / 失败 / 执行中 / 等待中 / 空闲”；已是最新时显示“已是最新镜像，无需更新”，并保留绿色说明“当前运行镜像已是最新，已跳过重复部署”。
  - 部署页说明新增旧镜像清理提示：更新成功或确认已是最新后，会自动执行未使用镜像清理，并保留最近 1 个备份镜像。
- 影响范围：
  - 仅影响后台设置页部署区的展示文案，不改变后端 API 或宿主部署脚本行为。

## 本次固定部署发布流程错误记录

- 错误现象：
  - 修复已提交并推送到功能分支 `fix/latest-provider-503` 后，后台“立即更新”仍显示旧输出：`backup images within keep limit: 0/2`。
  - 这说明服务器拉到的 `docker-deploy/sub2api-docker-image.tar` 仍是旧包，而不是新代码没有生效。
- 根因：
  - `.github/workflows/deploy-package.yml` 只在 `main` push 或手动 `workflow_dispatch` 时更新固定 `docker-deploy` release asset。
  - 仅推送功能分支不会刷新后台部署入口下载的固定镜像包。
- 正确流程：
  - 若使用后台部署入口验证服务端更新，必须确保改动已进入 `origin/main`。
  - 等 `Deploy Package` workflow 成功完成并覆盖 `docker-deploy/sub2api-docker-image.tar` 后，再到后台点击“立即更新”。
  - 如果后台更新仍显示旧日志，先检查 GitHub Actions 与 `docker-deploy` release 的目标 commit，不要先假设宿主脚本或容器运行逻辑失败。
- 影响范围：
  - 这是发布流程问题，不是部署脚本的运行时 bug。

## 本次宿主部署脚本同步修正

- 背景：
  - `docker-deploy` 镜像包更新后，后台部署页中文提示与镜像 digest 已更新，但最近部署输出仍显示 `backup images within keep limit: 0/2`。
  - 排查服务器发现宿主机真实执行脚本 `/home/ec2-user/sub2api-deploy/bin/deploy-from-package.sh` 仍是旧版：
    - `IMAGE_TAG="${IMAGE_TAG:-weishaw/sub2api:latest}"`
    - `KEEP_BACKUPS="${KEEP_BACKUPS:-2}"`
  - 根因是 host deploy agent 调用的是宿主机 root-owned 脚本，Docker 镜像更新不会自动覆盖宿主机 `bin/deploy-from-package.sh`。
- 服务器操作：
  - 已将仓库脚本 `deploy/host-agent/deploy-from-package.sh` 上传到服务器 `/tmp/deploy-from-package.sh.new`。
  - 使用 `sudo install -o root -g root -m 0755` 覆盖 `/home/ec2-user/sub2api-deploy/bin/deploy-from-package.sh`。
  - 已重启 `sub2api-host-deploy-agent`，服务状态为 `active`。
- 验证结果：
  - 服务器脚本当前为：
    - `IMAGE_TAG="${IMAGE_TAG:-sub2api:rollback}"`
    - `KEEP_BACKUPS="${KEEP_BACKUPS:-1}"`
  - 直接执行宿主脚本 no-op 路径后输出 `backup images within keep limit: 0/1`，确认新保留策略生效。
- 后续规则：
  - 以后凡是修改 `deploy/host-agent/deploy-from-package.sh` 或 host agent 行为，除了进入 `origin/main` 并等待 `Deploy Package` 完成，还必须确认是否需要同步宿主机 `/home/ec2-user/sub2api-deploy/bin/deploy-from-package.sh` 与 systemd agent。
  - 不能假设 Docker 镜像更新会自动更新宿主机部署脚本。

## 本次固定 Release Tag 指向修正

- 背景：
  - GitHub Release 页面显示 `docker-deploy` tag 仍指向旧提交 `97706cd`，但资源文件显示最近 1 分钟已更新，后台运行镜像 digest 也已变化。
  - 这会造成“GitHub 页面提交号”和“服务器镜像 digest / 资源文件时间”看起来对不上。
- 根因：
  - `softprops/action-gh-release` 覆盖 release asset 时不会保证已存在的 `docker-deploy` tag 被移动到当前 `main` commit。
  - GitHub 页面资源旁边的 `sha256:...` 是 tar 资源文件本身的 SHA256，不是 Docker image ID；后台显示的 `sha256:...` 是容器运行镜像 ID，两者正常不会相同。
- 修改内容：
  - `.github/workflows/deploy-package.yml`
- 修改前后差异：
  - 修改前：workflow 只覆盖 `sub2api-docker-image.tar` 与 `.sha256`，release tag 可能停留在旧 commit。
  - 修改后：发布固定包前先执行 `git tag -f docker-deploy "$GITHUB_SHA"` 并强推 `refs/tags/docker-deploy`，确保 Release 页面 tag、固定资源包和 `main` commit 对齐。
- 后续规则：
  - 判断固定部署包是否对应最新代码时，以 `Deploy Package` workflow 的 head SHA、`docker-deploy` tag 指向和资源更新时间共同确认。
  - 不要拿 release asset 文件 SHA 与 Docker image ID 做相等比较。

## 本次备份镜像保留逻辑修正

- 背景：
  - 部署脚本设置了 `KEEP_BACKUPS=1`，但服务器 `docker images` 只剩 `sub2api:rollback`，没有 `sub2api:backup-*`。
  - 根因是脚本在 `backup_current_image` 创建备份 tag 后，又执行 `docker image prune -a -f`；`-a` 会删除所有未被容器使用的镜像，包括刚创建的 backup tag。
- 修改内容：
  - `deploy/host-agent/deploy-from-package.sh`
- 修改前后差异：
  - 修改前：`prune_unused_images` 执行 `docker image prune -a -f`，会把未被容器引用的备份镜像一并清掉。
  - 修改后：`prune_unused_images` 改为 `docker image prune -f`，只清理 dangling 镜像层；备份镜像由 `prune_old_backups` 按 `KEEP_BACKUPS=1` 单独控制。
- 后续规则：
  - 如果目标是保留 tagged backup image，禁止使用 `docker image prune -a` 作为常规收尾清理。

## 本次终端操作问题记录

- PowerShell 与远端 shell 混用问题：
  - 在 PowerShell 中执行 `ssh "... $tmp ..."` 或 `ssh "... $(date ...) ..."` 时，`$tmp` / `$(...)` 可能会先被本机 PowerShell 展开，导致远端命令异常。
  - 本次表现：
    - `curl: (2) no URL specified`
    - `Cannot bind parameter 'Date'. Cannot convert value "+%Y%m%d%H%M%S"`
  - 后续规则：
    - 远端 shell 变量尽量避免放在 PowerShell 双引号字符串里。
    - 需要远端展开变量时，优先用单引号包住整段远端命令，或在本机先生成安全值再传入远端命令。
- Windows 本机命令差异：
  - 本机默认 shell 是 PowerShell，不保证有 Linux `grep` / `bash`。
  - 本次表现：
    - `grep` 在本机不可用。
    - `bash -n` 触发 WSL 但 `/bin/bash` 不存在。
  - 后续规则：
    - 本机搜索文件继续优先用 `rg` / `Select-String`。
    - shell 脚本语法检查优先在服务器 Linux 环境执行 `bash -n`，不要假设 Windows 本机可跑 bash。
- GitHub API 查询限制：
  - 未认证 GitHub REST API 容易触发 rate limit。
  - 本次表现：
    - `API rate limit exceeded`
  - 后续规则：
    - 查询 workflow 状态时优先减少轮询次数。
    - 能用 `git ls-remote` 验证 tag / branch 指向时，不依赖 GitHub API。
- 宿主机 root-owned 文件同步：
  - `/home/ec2-user/sub2api-deploy/bin/deploy-from-package.sh` 是 root-owned，`scp` 不能直接覆盖。
  - 本次表现：
    - `Permission denied`
    - `chmod: Operation not permitted`
  - 后续规则：
    - 先上传到 `/tmp/*.new`。
    - 在服务器上执行 `bash -n /tmp/*.new` 验证。
    - 再用 `sudo install -o root -g root -m 0755 /tmp/*.new <target>` 原子替换。

## 本次账号物理删除与批量状态删除修正

- 背景：
  - 后台文档已声明账号删除应为物理删除，但实际 `accountRepository.Delete` 仍调用普通 Ent `Delete()`。
  - `accounts` 使用 `SoftDeleteMixin`，普通 `Delete()` 会被 hook 改写为 `UPDATE deleted_at = NOW()`，因此仍是软删除。
  - 测活页的“删除异常账号”原本只删除健康检查状态为 `unavailable` 的账号，无法覆盖其它页面测活后写入的账号状态。
- 修改内容：
  - `backend/internal/repository/account_repo.go`
  - `backend/internal/handler/admin/account_handler.go`
  - `frontend/src/api/admin/accounts.ts`
  - `frontend/src/views/admin/AccountHealthView.vue`
  - `frontend/src/components/admin/account-health/AccountHealthAutoCheckPanel.vue`
  - `frontend/src/i18n/locales/zh.ts`
  - `frontend/src/i18n/locales/en.ts`
- 修改前后差异：
  - 修改前：单账号删除、去重删除、异常账号批量删除最终都会走普通 Ent 删除，实际会软删除 `accounts.deleted_at`。
  - 修改后：账号删除在删除 `account_groups`、`scheduled_test_plans` 后，通过 `mixins.SkipSoftDelete(ctx)` 执行真实 `DELETE FROM accounts` 语义。
  - 修改前：`POST /admin/accounts/delete-unhealthy` 空 payload 只删除健康检查 `unavailable`。
  - 修改后：空 payload 仍保持兼容；新增 `account_statuses` 与 `health_statuses` 可选数组，可按账号状态和健康检查状态批量物理删除。
- 支持的批量删除状态：
  - 账号状态：`disabled` / `inactive`、`error`、`rate_limited`、`temp_unschedulable`、`unschedulable`。
  - 健康检查状态：`unavailable`、`constrained`、`unchecked`，接口也接受 `healthy` 但前端未默认展示为清理选项。
- 前端行为：
  - 测活管理页新增“批量删除范围”勾选区，默认勾选停用、错误、不可用。
  - 删除确认框明确提示“物理删除数据库记录，不可撤销”，并列出当前匹配状态。
- 验证记录：
  - `frontend` 已执行 `pnpm run typecheck` 通过。
  - 本机 Windows 环境没有 `go/gofmt`，后端格式化与 Go 测试无法在本机直接执行；后续如需 Go 验证，应在 Linux/CI 或具备 Go 工具链的环境运行。
- 后续规则：
  - 只要任务要求“不要软删除”，必须检查 Ent soft-delete hook 是否需要 `mixins.SkipSoftDelete(ctx)`。
  - 批量删除类按钮必须让用户明确知道删除条件与是否不可撤销，不能只写“删除异常”这种模糊状态。

## 本次账号管理表格与账号健康卡片样式修正

- 背景：
  - 账号管理表的名称列在长邮箱 / 长域名账号名场景下会撑得过宽，挤压平台、状态、操作等后续列。
  - 账号健康页摘要卡片使用大面积等高色块，数字较少时留白过大，右侧设置面板与左侧卡片高度比例不协调。
- 修改内容：
  - `frontend/src/components/common/types.ts`
  - `frontend/src/components/common/DataTable.vue`
  - `frontend/src/views/admin/AccountsView.vue`
  - `frontend/src/components/admin/account-health/AccountHealthAutoCheckPanel.vue`
- 修改前后差异：
  - 修改前：`DataTable` 的列定义只支持 `class`，无法稳定给单列设置宽度；账号名文本没有明确的最大宽度约束。
  - 修改后：通用列定义新增 `width` / `minWidth` / `maxWidth`，表头与单元格会应用这些尺寸；账号名列固定为紧凑宽度，并对账号名与邮箱做 `truncate`，完整内容保留在 `title`。
  - 修改前：账号健康统计卡片最小高度较高、数字字号偏大、提示文字底部留白明显。
  - 修改后：统计卡片统一使用 `health-stat-card` 紧凑样式，降低高度、收敛数字字号与间距，右侧控制面板不再被左侧大卡片明显拉高。
- 影响范围：
  - 仅影响前端展示样式与通用表格列尺寸能力。
  - 不修改账号 API、分页、排序、删除、测活或任何后端逻辑。
- 待验证事项：
  - 需执行前端 `typecheck` / `build` 验证 Vue 模板与新增列字段类型。

## 本次健康检查间隔单位与使用记录布局修正

- 背景：
  - 测活管理页的自动健康检查间隔只显示分钟，实际配置较长周期时不直观，容易误以为只能设置很短的分钟级频率。
  - 使用记录页筛选项较多，原本使用一整条 flex 换行布局，宽屏下筛选区和操作按钮区比例不稳定，视觉上偏散。
- 修改内容：
  - `frontend/src/views/admin/AccountHealthView.vue`
  - `frontend/src/components/admin/account-health/AccountHealthAutoCheckPanel.vue`
  - `frontend/src/components/admin/usage/UsageFilters.vue`
  - `frontend/src/i18n/locales/zh.ts`
  - `frontend/src/i18n/locales/en.ts`
  - `.learnings/ERRORS.md`
- 修改前后差异：
  - 修改前：自动健康检查间隔只输入分钟值，并直接把输入数字传给后端 `interval_minutes`。
  - 修改后：前端新增分钟 / 小时单位选择；读取配置时会把可整除 60 的分钟数显示为小时，保存时仍换算为后端既有的 `interval_minutes`，不改 API 结构。
  - 修改前：使用记录筛选条和操作按钮在同一 flex 区域中自然换行，筛选项过多时布局不够均衡。
  - 修改后：筛选项改为响应式 grid，操作按钮单独放在带顶部分隔线的工具列，宽屏更平衡，窄屏仍可自然换行。
- 影响范围：
  - 仅影响前端展示、输入换算与中文 / 英文文案。
  - 不修改后端接口、数据库字段、调度逻辑或使用记录查询参数。
- 流程记录：
  - 新增 `.learnings/ERRORS.md`，记录 `findstr` 卡住时应立即改用 `rg` + 指定文件片段读取，避免后续继续在低效搜索上耗时。
- 待验证事项：
  - 需执行前端 `pnpm run typecheck` 验证 Vue 模板与类型。

## 本次吸收上游更新（v0.1.134）

- 背景：
  - 本地 `main` 之前已吸收到上游 `f18451e5` 一带的内容，本次继续吸收 `upstream/main` 至 `635ad81c`。
  - 由于 GitHub 网络不稳定，本次上游抓取先经历 shallow fetch、`shallow.lock` 残留与 HTTPS 连接重置，后续已补充到 `.learnings/ERRORS.md`。
- 吸收范围：
  - 上游区间：`f18451e5..635ad81c`
  - 吸收分支：`absorb-upstream-v0.1.134`
- 本次上游重点内容：
  - OpenAI / Responses / Codex / Claude Code 兼容链路继续增强，包括流事件校验、sticky account、failed 透传、图片限流冷却、Codex/Claude 模拟一致性等。
  - 运维与观测侧新增失败请求展示、用户错误请求视图、TTFT 样本权重修正、错误日志归因补强。
  - 多实例定时任务引入 leader lock，避免订阅过期提醒、支付订单过期扫表等任务在多实例下重复执行。
  - 管理端 / 用量页继续演进，新增失败请求相关界面与测试，搜索与筛选能力增强。
- 本地保留的自定义能力：
  - 账号物理删除与可按账号状态/健康状态批量删除。
  - 账号健康检查页与 Token 自动刷新页。
  - 固定 docker deploy / host-agent 相关的本地部署体系与文档记录。
- 冲突处理：
  - `backend/cmd/server/wire_gen.go`
    - 保留上游新增的 `leader lock` 依赖注入。
    - 同时保留本地 `ScheduledTestRunnerService` 对 `settingService / accountRepo / tokenRefreshService` 的依赖，以维持自动测活与 token 刷新能力。
  - `frontend/src/components/common/Select.vue`
    - 合并保留本地的 `size` 属性与上游的 `clearable` 属性。
- 验证记录：
  - `frontend` 已执行 `pnpm run typecheck` 通过。
  - 本机 Windows 环境仍无 `go`，因此无法在本地执行 `go test`、`go build` 或重跑 `wire` 生成；后端需依赖 CI 或 Linux/Go 工具链环境继续验证。

## 本次吸收外部防风控补丁（增量吸收，非整包覆盖）

- 背景：
  - 使用者提供下载目录中的 `v133.zip` / `v134.zip` 防风控补丁，希望吸收有效代码，但不要整包替换当前项目。
  - 经比对，补丁中相当一部分能力在当前主链已存在正式实现，例如：
    - OAuth/Codex 会话隔离
    - `invalid_encrypted_content` 的恢复重试
    - 浏览器 UA 改写与 TLS 指纹相关能力
  - 因此本次采用“只吸收低侵入、可独立落地的 helper 与接线”，避免把旧补丁版主逻辑倒灌回当前主链。
- 本次实际吸收内容：
  - 新增 `backend/internal/guard/reasoning_sanitizer.go`
    - 在请求发往上游前，预清洗结构无效的 `reasoning.encrypted_content`
    - 目标是减少可避免的 `400 invalid_encrypted_content`
  - 新增 `backend/internal/guard/session_headers.go`
    - 规范化 `session_id` 头部变体
    - 补齐 `X-Client-Request-Id`、`Thread-Id`、`X-Codex-Window-Id`
    - 在缺失时同步 `conversation_id`
  - 接入位置：
    - `backend/internal/service/openai_gateway_service.go`
      - OAuth 请求在转发前执行 reasoning 预清洗
      - OAuth HTTP 转发请求头执行 session governance
    - `backend/internal/service/openai_ws_forwarder.go`
      - OAuth WS 建连头执行 session governance
  - 新增单元测试：
    - `backend/internal/guard/reasoning_sanitizer_test.go`
    - `backend/internal/guard/session_headers_test.go`
- 明确未吸收的补丁部分：
  - `identity_confuse.go`
  - `codex_reasoning_replay.go`
  - 对旧版 `openai_gateway_service.go` / `openai_ws_forwarder.go` 的整段响应恢复与 replay 注入逻辑
- 未吸收原因：
  - 与当前主链已有的会话隔离、重试、自愈逻辑重叠较多。
  - 侵入面较大，直接嫁接的回归风险高于当前收益。
- 验证记录：
  - 本机无 `go` 工具链，无法在当前环境运行 Go 单测或编译验证。
  - 需后续在 CI 或 Linux/Go 环境继续验证。

## 本次吸收外部防风控补丁第二层（轻量 identity confuse）

- 背景：
  - 使用者允许在已保存备份与第一层低风险 helper 的基础上，继续吸收一小部分身份隔离逻辑。
  - 目标是继续增强 OAuth / Codex 元数据的账号隔离，但仍避免引入补丁里那套重型的响应恢复、prompt cache 混淆与 replay cache。
- 本次实际吸收内容：
  - 新增 `backend/internal/guard/identity_confuse.go`
    - 提供 `ConfuseKey`
    - 提供 `ConfuseCodexMetadataLight`
  - 轻量混淆范围仅限：
    - `client_metadata.x-codex-installation-id`
    - `client_metadata.x-codex-turn-metadata` 中的 `turn_id`
  - 明确保留不改动：
    - `prompt_cache_key`
    - `window_id`
    - 响应体恢复逻辑
    - header 级二次混淆
- 接入位置：
  - `backend/internal/service/openai_gateway_service.go`
    - OAuth 请求在发往上游前，对上述 Codex 元数据执行轻量混淆
- 新增测试：
  - `backend/internal/guard/identity_confuse_test.go`
- 设计原因：
  - 这一步只碰上游可见但客户端通常不依赖回读的元数据字段，风险明显低于直接混淆 `prompt_cache_key` 或对响应体做恢复替换。
  - 继续避免与当前主链已有的会话隔离、`invalid_encrypted_content` 恢复、自愈和重试链路发生大面积耦合。
- 验证记录：
  - 本机无 `go` 工具链，无法在当前环境运行 Go 单测或编译验证。
  - 需后续在 CI 或 Linux/Go 环境继续验证。

## 本次 GitHub Actions 失敗修復（OpenAI messages bridge 編譯錯）

- 背景：
  - commit `c4ce65b9 feat: enhance account maintenance tools` 的 GitHub Actions 失敗。
  - GitHub check annotations 顯示失敗集中在後端相關任務：
    - `test`
    - `golangci-lint`
    - `backend-security`
    - `publish-deploy-package`
  - 前端 workflow 已通過。
- 根因：
  - `backend/internal/service/openai_gateway_service.go` 在 `buildUpstreamRequest` 內只於 OAuth 分支中使用 `:=` 宣告 `compatMessagesBridge`。
  - 後續 `guard.ApplySessionGovernance` 在 OAuth 分支外引用該變數，造成 Go 編譯錯：
    - `undefined: compatMessagesBridge`
  - 因為 unit test、lint、govulncheck 與 Docker build 都需要編譯後端，所以同一個 scope 錯誤連帶造成多個 workflow 失敗。
- 本次修改：
  - 在透傳 header 後、OAuth 分支前先宣告 `compatMessagesBridge := false`。
  - OAuth 分支內改為賦值 `compatMessagesBridge = ...`。
  - 保持原行為不變：
    - 非 OAuth 账号默认不是 messages bridge。
    - OAuth messages bridge 仍跳过 session governance。
- 修改文件：
  - `backend/internal/service/openai_gateway_service.go`
  - `SYSTEM.md`
- 影響範圍：
  - 僅修正變數作用域，不新增 API、schema 或配置。
  - 預期修復後端編譯失敗，進而解除 CI test/lint/security/deploy 的同源失敗。
- 驗證記錄：
  - 已透過 GitHub check annotations 確認原失敗訊息為 `undefined: compatMessagesBridge`。
  - 當前 Windows 本機無 `go` 工具鏈；Docker Desktop 啟動前不可用，需待 Docker 可用後執行後端編譯/測試或交由 CI 驗證。

## 本次 GitHub Actions 第二輪修復（測試 constructor 參數同步）

- 背景：
  - 推送 `85333744 fix(openai): repair messages bridge build` 後，遠端 CI 中：
    - `Security Scan` 通過
    - `Deploy Package` 通過
    - `frontend` 通過
    - `test` / `golangci-lint` 仍失敗
- 根因：
  - GitHub check annotations 顯示 Go 編譯仍有測試呼叫參數不足：
    - `not enough arguments in call to NewUpdateService`
    - `not enough arguments in call to NewAccountHandler`
  - `NewAccountHandler` 已包含 `tokenCacheInvalidator` 參數，兩個多行測試 helper 少補最後一個 `nil`。
  - `NewUpdateService` 已包含 `settingRepo` 參數，單測仍按舊簽名傳入。
- 本次修改：
  - `backend/internal/handler/admin/account_handler_passthrough_test.go`
    - `NewAccountHandler` 測試 helper 補最後一個 `nil`。
  - `backend/internal/handler/admin/account_data_handler_test.go`
    - `NewAccountHandler` 測試 helper 補最後一個 `nil`。
  - `backend/internal/service/update_service_test.go`
    - `NewUpdateService` 測試補 `settingRepo` 的 `nil` 參數。
- 影響範圍：
  - 僅同步測試構造器呼叫，不改正式業務邏輯。
  - 預期修復 CI 單測與 golangci-lint 的 typecheck 失敗。
- 驗證記錄：
  - 本機仍無 Go 工具鏈，將以遠端 GitHub Actions 作為主要驗證來源。

## 本次 GitHub Actions 第三輪修復（lint / gofmt 清理）

- 背景：
  - 推送測試 constructor 修復後，遠端 CI 中 `Security Scan`、`Deploy Package`、`frontend` 通過。
  - `CI` 仍失敗，annotations 顯示剩餘問題集中於 golangci-lint 與 Go 測試編譯階段。
- 根因：
  - `backend/internal/service/update_deploy.go` 保留了三個未使用 helper：
    - `parseDeployImageID`
    - `parseDeployRunningImageID`
    - `parseDeployResultField`
  - `backend/internal/handler/admin/account_handler.go` 對嵌入字段使用 `item.Account.Health...`，觸發 staticcheck `QF1008`。
  - 三個 Go 文件存在 gofmt 對齊問題：
    - `backend/internal/service/account_test_service.go`
    - `backend/internal/handler/dto/types.go`
    - `backend/internal/handler/admin/account_handler.go`
- 本次修改：
  - 移除未使用 deploy helper。
  - 將 `item.Account.HealthStatus / HealthResultStatus / HealthMessage` 改為直接訪問嵌入字段。
  - 手動調整 CI 指出的 gofmt 對齊區塊。
- 影響範圍：
  - 僅清理 lint / 格式問題，不改 API、schema、部署流程或前端行為。
- 驗證記錄：
  - 本機可執行 `git diff --check` 且結果乾淨。
  - 本機無 `gofmt` / Go 工具鏈，最終 gofmt 與單測仍以遠端 GitHub Actions 為準。

## 本次 GitHub Actions 第四輪修復（官方 gofmt 與剩餘 staticcheck）

- 背景：
  - 第三輪提交後，遠端 CI 中 `Security Scan`、`Deploy Package`、`frontend` 繼續通過。
  - `golangci-lint` 仍指出剩餘 staticcheck selector 與 gofmt 格式問題。
- 根因：
  - `AccountWithConcurrency` 嵌入了 `*dto.Account`，仍有部分健康欄位透過 `item.Account...` 訪問，觸發 `QF1008`。
  - 手動空格對齊未完全等同官方 `gofmt`。
- 本次修改：
  - 將剩餘 `item.Account.HealthLatencyMs`、`item.Account.HealthLastCheckedAt` 以及列表構建處的健康欄位訪問改為直接訪問嵌入字段。
  - 使用 Go 官方格式化接口對以下文件執行格式化：
    - `backend/internal/handler/admin/account_handler.go`
    - `backend/internal/handler/dto/types.go`
    - `backend/internal/service/account_test_service.go`
  - 清除 PowerShell 寫檔引入的 UTF-8 BOM，保留無 BOM UTF-8。
- 影響範圍：
  - 僅格式化與 lint 清理，不改功能邏輯。
- 驗證記錄：
  - 本機 `git diff --check` 通過。
  - 最終 Go 單測與 golangci-lint 仍由遠端 GitHub Actions 驗證。

## 本次分页选项与测活状态筛选修正

- 背景：
  - 账号列表分页下拉仍只显示旧档位，未稳定出现 `100 / 500 / 1000`。
  - 测活页手动筛选的“检查状态”下拉显示的是账号运行状态（正常、停用、错误、限流中等），与健康状态筛选期望不一致。
- 根因：
  - `frontend/src/utils/tablePreferences.ts` 在读到服务端注入的 `table_page_size_options` 后，会完全使用服务端配置；如果服务器数据库仍保存旧配置，新内建分页档位会被覆盖。
  - `frontend/src/views/admin/AccountHealthView.vue` 手动测活 filters 发送的是 `status`，对应账号状态；后端实际也支持 `health_status` 用于健康状态筛选。
  - `frontend/src/components/admin/account-health/AccountHealthAutoCheckPanel.vue` 使用原生 `<select>` 渲染状态筛选，视觉上与项目内统一 `Select` 组件不一致。
- 本次修改：
  - `frontend/src/utils/tablePreferences.ts`
    - 将分页选项改为“内建默认档位 + 服务端配置”合并去重排序，确保旧服务端配置也会显示 `100 / 500 / 1000`。
  - `frontend/src/utils/__tests__/tablePreferences.spec.ts`
    - 更新分页偏好测试，覆盖旧配置与新内建档位合并行为。
  - `frontend/src/views/admin/AccountHealthView.vue`
    - 手动测活筛选从 `status` 改为发送 `health_status`。
  - `frontend/src/components/admin/account-health/AccountHealthAutoCheckPanel.vue`
    - 状态下拉改用项目统一 `Select` 组件。
    - 选项改为：全部健康状态、健康、受限、不可用、未检查。
- 影响范围：
  - 不改后端 API；复用既有 `health_status` filter。
  - 分页组件在存在旧服务端配置时会自动补齐新档位，不需要用户手动进设置页保存一次。
- 验证记录：
  - `pnpm --dir frontend run typecheck` 通过。
  - `pnpm --dir frontend exec vitest run src/utils/__tests__/tablePreferences.spec.ts` 通过，6 个测试全部通过。

## 本次测活页批量删除布局调整

- 背景：
  - 账号健康检查页右侧自动测活配置栏同时包含“批量删除范围”，导致右栏高度过高。
- 本次修改：
  - `frontend/src/components/admin/account-health/AccountHealthAutoCheckPanel.vue`
    - 将“批量删除范围”卡片从右侧配置栏移动到左侧健康统计卡片下方。
    - 将“删除异常账号”按钮随批量删除卡片一起移动到左侧，保持删除范围与删除动作相邻。
    - 右侧配置栏底部仅保留“立即检测”和“保存配置”两个按钮，降低右栏高度。
- 影响范围：
  - 仅调整页面布局，不改变删除 API、测活 API、筛选参数或状态逻辑。
- 验证记录：
  - `pnpm --dir frontend run typecheck` 通过。

## 本次账号管理批量工具与测活筛选增强

- 背景：
  - 后端已经存在 `POST /api/v1/admin/accounts/deduplicate` 去重接口，但账号管理页面没有入口。
  - 分页底层已允许最大 `1000`，但前端与系统默认设置仍只提供到 `100`。
  - 手动账号健康检查后端已支持 `filters`，但测活管理页只传模型 ID，无法按分组或账号状态缩小范围。
- 修改内容：
  - `frontend/src/views/admin/AccountsView.vue`
    - 在账号管理「更多操作」的数据操作区新增「去除重复账号」按钮。
    - 点击后会二次确认，并按当前列表筛选条件调用既有去重接口。
    - 去重完成后清空当前选择、刷新列表，并显示删除数量、命中重复组数与保留账号数。
  - `frontend/src/views/admin/AccountHealthView.vue`
  - `frontend/src/components/admin/account-health/AccountHealthAutoCheckPanel.vue`
    - 手动健康检查新增可选「检查分组」与「检查状态」。
    - 分组与状态为空时不传过滤条件，保持默认检查全部账号。
    - 自动健康检查配置仍沿用原逻辑，不受手动筛选影响。
  - `frontend/src/utils/tablePreferences.ts`
  - `frontend/src/stores/app.ts`
  - `frontend/src/views/admin/SettingsView.vue`
  - `backend/internal/service/setting_service.go`
    - 默认分页可选条数扩展为 `[10,20,50,100,500,1000]`。
    - 设置页 placeholder 与 fallback 同步到新默认值。
  - `frontend/src/i18n/locales/zh.ts`
  - `frontend/src/i18n/locales/en.ts`
    - 补充去重、测活筛选和分页设置相关文案。
  - `frontend/src/utils/__tests__/tablePreferences.spec.ts`
    - 更新内建分页默认值测试期望。
- 修改前后差异：
  - 修改前：账号去重只能通过已有 API 或其它非页面入口触发；分页常规选择到 `100` 为止；手动测活总是检查全部符合后端默认范围的账号。
  - 修改后：账号页可直接按当前筛选去重；分页组件默认显示 `100 / 500 / 1000`；手动测活可以按指定分组和账号状态执行，不选则默认全部。
- 影响范围：
  - 前端新增入口使用既有后端接口，不新增数据库 schema 或新 API。
  - 去重仍由后端按既有规则执行：按平台、类型、账号名归组，保留每组 ID 最小的账号，删除其它重复项。
  - 分页默认值变化会影响未配置自定义分页选项的新部署或 fallback 场景；已由后台设置保存的自定义值仍以服务端配置为准。
- 待验证事项：
  - 需执行前端 typecheck 与相关单元测试。
  - 本机若无 Go 工具链，后端默认设置改动无法在本地跑 Go 测试，需依赖 CI 或 Linux/Go 环境验证。

## 本次吸收外部防风控补丁第三层（Cloudflare challenge 渐进冷却）

- 背景：
  - 使用者同意继续按低风险思路推进，不直接照搬补丁版 `cloudflare_backoff`，而是改写成兼容当前主链 `runtime block` 的版本。
  - 当前主链已经有：
    - `httputil.IsCloudflareChallengeResponse`
    - `BlockAccountScheduling`
    - OpenAI OAuth 429 / 403 / 临时不可调度等运行时阻断能力
  - 因此本次不新增独立大系统，只在现有 fastpath 上加一层 challenge 冷却。
- 本次实际吸收内容：
  - 在 `backend/internal/service/openai_account_runtime_block_fastpath.go` 增加 Cloudflare challenge 渐进冷却状态：
    - 10 秒
    - 30 秒
    - 90 秒
    - 120 秒封顶
  - 同一账号若连续命中 challenge，会逐级升高；若超过一段空闲窗口未再触发，则回到首档。
  - 仅对 `PlatformOpenAI + OAuth` 生效，不作用于 API Key 账号。
- 接入方式：
  - 复用 `httputil.IsCloudflareChallengeResponse` 做检测。
  - 复用 `BlockAccountScheduling` 做临时调度阻断。
  - 不改客户端返回格式，不新增外部配置项，不影响现有 429/image cooldown 链路。
- 修改文件：
  - `backend/internal/service/openai_account_runtime_block_fastpath.go`
  - `backend/internal/service/openai_account_runtime_block_fastpath_test.go`
  - `backend/internal/service/openai_gateway_service.go`
- 新增测试覆盖：
  - OAuth 账号命中 challenge 后会被 runtime block
  - 连续命中 challenge 的冷却时长会递增
  - 空闲一段时间后冷却级别会重置
- 验证记录：
  - 本机无 `go` 工具链，无法在当前环境运行 Go 单测或编译验证。
  - 需后续在 CI 或 Linux/Go 环境继续验证。

## 本次吸收外部防风控补丁第三层（Cloudflare 质询渐进冷却）

- 背景：
  - 使用者同意继续按“低侵入、与当前主链兼容”的方式推进，不直接搬运补丁中的整套 Cloudflare backoff 代码。
  - 当前主链已存在：
    - `httputil.IsCloudflareChallengeResponse`
    - `BlockAccountScheduling`
    - OpenAI OAuth 账号 runtime block 快路径
  - 因此本次直接挂接在现有 runtime block 链路上，避免新起一套并行状态机。
- 本次实际吸收内容：
  - 在 `backend/internal/service/openai_account_runtime_block_fastpath.go`
    - 增加 Cloudflare challenge 检测后的渐进冷却
    - 冷却梯度：`10s -> 30s -> 90s -> 120s`
    - 若长时间未再次触发，则重新从 `10s` 起步
  - 在 `backend/internal/service/openai_gateway_service.go`
    - 新增 `openaiCloudflareChallengeState` 状态表
  - 新增测试：
    - `backend/internal/service/openai_account_runtime_block_fastpath_test.go`
- 生效范围：
  - 仅对 `PlatformOpenAI + OAuth` 账号生效
  - 仅在识别到 `Cloudflare challenge` 的上游响应时触发
  - 不改变客户端错误返回格式，仅影响账号短时调度冷却
- 明确保留不改动：
  - API Key 账号不进入该冷却逻辑
  - 不改现有 429 fallback、image cooldown、temp_unschedulable 规则
  - 不引入补丁原版中独立的全新 backoff 配置结构
- 设计原因：
  - 目标是降低同一 OAuth 账号在短时间内反复撞到 Cloudflare challenge 的概率，同时不重写现有 rate limit / runtime block 架构。
- 验证记录：
  - 本机无 `go` 工具链，无法在当前环境运行 Go 单测或编译验证。
  - 需后续在 CI 或 Linux/Go 环境继续验证。
