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
- 用途：保存伺服器上的 `docker-compose.yml`、`.env`、`bin/deploy-from-package.sh`、資料目錄與運維腳本
- 特性：它不是正式 Git 倉庫主源，只是遠端運行與部署輔助目錄
- 注意：此目錄下還存在一份 `sub2api-src/` 快照，但它不應被視為正式開發主源

### 3. 容器運行層

- 目前主應用容器：`sub2api`
- 目前運行鏡像：`weishaw/sub2api:latest`
- 真正對外提供服務的是容器，不是本地 repo，也不是部署目錄裡的源碼快照

## 維護規則

- 業務功能改動：優先修改本地 / Git 倉庫，然後提交推送。
- 伺服器運維腳本如果在遠端熱修：
  - 若屬於應長期保留的能力，必須同步回本地 repo。
  - 若只是一次性排障操作，可只留在伺服器，不必回寫 repo。
- 發布與服務器更新流程：
  - GitHub Releases 頁面預期保留版本 tag（如 `v0.1.134`）與固定部署入口 `docker-deploy` 兩類。
  - 服務器更新只使用固定 `docker-deploy` release 的部署包。
  - 普通可部署修正不需要新建版本 tag；應提交後推送到 `origin/main`，由 `Deploy Package` workflow 更新 `docker-deploy/sub2api-docker-image.tar`。
  - 只有使用者明確要求版本化發布時，才建立新的 `v*` tag / version release。
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
  - 舊 backup image 自動保留最近 2 個
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
