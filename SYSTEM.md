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
