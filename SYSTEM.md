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
  - 只處理可刷新的 OAuth 帳號
  - 依設定的 batch size 分批刷新
  - 批次間加入短暫間隔，避免瞬時大量請求
- OAuth 刷新底層新增 `RefreshNow` 路徑，沿用既有分布式鎖與 DB 重讀保護，避免與其他刷新路徑競爭。

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
