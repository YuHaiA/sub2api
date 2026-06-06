# 錯誤記錄

## 2026-05-31：避免用 `findstr` 長時間硬找文件片段

- 現象：用 `findstr` 跟蹤/搜尋文件片段時容易卡住或耗時過長，導致排查節奏變慢。
- 正確流程：優先使用 `rg` 定位，再用 `Get-Content`、`Select-Object` 或具體文件路徑讀取上下文。
- 執行規則：拿到結構後再改，不要在低效搜尋上硬耗。

## 2026-06-06：吸收上游時避免把 shallow fetch 卡死在工作樹

- 現象：對公開 GitHub 倉庫做 `git fetch --deepen` 時，HTTPS 連線可能被重置或超時，並留下 `.git/shallow.lock`，後續 fetch 全部被鎖死。
- 本次表現：
  - `RPC failed; curl 28 Recv failure: Connection was reset`
  - `Failed to connect to github.com port 443`
  - `Unable to create ... .git/shallow.lock: File exists`
- 正確流程：
  - 先用 `git ls-remote` 驗證上游 head 是否可達。
  - 再用 `git fetch --depth=1` 或小步 `--deepen=20/30` 補歷史，不要一口氣 deepen 很大。
  - 若超時後殘留的是本倉庫對應的 fetch 進程，先只清理那組 PID，再移除 `.git/shallow.lock`。
- 執行規則：只終止與當前倉庫相關的卡死 `git fetch/index-pack`，不要誤殺其他專案或工具自己的 Git 進程。
