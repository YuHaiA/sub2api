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
- OAuth 刷新底層新增 `RefreshNow` 路徑，沿用既有分布式鎖與 DB 重讀保護，避免與其他刷新路徑競爭。

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
  - 若當前容器已經運行與 `weishaw/sub2api:latest` 相同的 image digest，則直接判定已是最新版並跳過 `load/tag/compose`。
- 部署成功或 no-op 後會執行 Docker 未使用鏡像清理：
  - `docker image prune -a -f`
  - 用於移除無用舊鏡像層，降低磁碟佔用
- 管理後台系統設置頁的部署狀態現已補充最近部署輸出展示：
  - 後端 `DeployState.last_output`
  - 前端系統設置部署區塊可直接查看最近一輪部署輸出

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
  - `frontend/src/api/admin/accounts.ts`
  - `frontend/src/i18n/locales/zh.ts`
  - `frontend/src/i18n/locales/en.ts`

## 已知驗證狀態

- 前端 `pnpm typecheck` 已通過。
- 目前工作環境缺少 `go` 指令，因此尚未在本機執行後端 `go test` / `go build`。
