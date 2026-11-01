# scout PROGRESS

## §0 接手速覽

- **狀態**：v0.1 完成，尚未部署到 Cubi N9。
- **驗證基準**：`make test` → 7 個套件、27 個測試全過（含 `-race`）；`make vet` 三平台通過；`go list -m all` 只有本模組（零依賴）。
- **真實來源實測（2026-11-01，雲端開發環境）**：15 個預設來源、14 個啟用，13 個成功、1 個失敗（GitHub releases.atom 被開發環境 proxy 擋，403）；首次 fetch 1.4 秒、241 篇入庫、2,300 篇超過 14 天被略過；第二次 fetch 10 個來源回 304；state.json 168 KB；`candidates --since 3d --limit 0` 10 ms。
- **下一步**：在 Cubi N9 上 `make build`、`scout init`、確認 claude-code-releases 能抓到，跑第一份真實 digest，依使用者回饋調整權重與關鍵字。
- **動手前只要讀**：`AGENTS.md` 的「已踩過的陷阱」。

## 進度

| 項目 | 狀態 | 驗收 |
|---|---|---|
| feed 解析（RSS/Atom/RDF/HN） | ✅ 完成並驗證 | `TestParse*`；真實 13 個來源 entries 皆 > 0 |
| 條件式 GET、單一來源失敗隔離 | ✅ 完成並驗證 | `TestFetchConditionalGet`、`TestFetchErrors`；真實第二次 fetch 10× not_modified |
| 跨來源去重 | ✅ 完成並驗證 | `TestCanonical`、`TestMerge`；真實 Karpathy+HN、Jepsen HN+Lobsters 合併 |
| 舊文略過（14 天） | ✅ 完成並驗證 | 真實 state 從 2,541 篇降到 241 篇，候選數不變（161） |
| 排序與理由、時間窗 | ✅ 完成並驗證 | `TestScoreComponents`、`TestInWindow`、`TestBoostCapped` |
| read 正文擷取 | ✅ 完成並驗證 | `TestPrefersArticleAndDropsNoise` 等；真實 karpathy.bearblog.dev 取得 `<main>` 2,178 字 |
| digest 建立／驗證／顯示／匯出 | ✅ 完成並驗證 | `TestFullWorkflow`；真實資料建立 3 篇 digest、匯出 2 個 HTML |
| trends | ✅ 完成並驗證 | `TestCompute`、`TestTerms`；真實資料出現 `anthropic` 0→6、`data centers` |
| 來源／關鍵字管理 | ✅ 完成並驗證 | `TestSourcesAndKeywords` |
| describe／help／錯誤分類 | ✅ 完成並驗證 | `TestUsageAndDiscovery`、`TestClassify` |
| 回歸測試有效性（探針） | ✅ 完成並驗證 | 重新注入 AutoClose 與 nbsp 兩個 bug，`TestParseRSSLinkText` 皆在正確位置轉紅，移除後轉綠 |
| Cubi N9 部署與每日實際使用 | ⬜ 未開始 | 第一份真實 digest 由使用者確認品質 |

## 已知未覆蓋

| 沒驗什麼 | 為什麼 | 替代覆蓋 | 它不能證明什麼 |
|---|---|---|---|
| GitHub releases.atom | 開發環境 proxy 回 403 | Atom 解析有單元測試 | 不能證明 GitHub 在使用者網路下不會限流 |
| 由 agent 完整跑一次每日流程（挑選品質） | 需要使用者在 Cubi N9 上的 session | CLI 流程由整合測試與手動實測覆蓋 | 不能證明 SKILLS.md 的編輯準則足以產出好的精選 |
| 非 UTF-8／latin-1 以外編碼的 feed | 預設來源皆為 UTF-8 | 明確報錯並帶編碼名稱（測試涵蓋 big5） | Big5 feed 目前無法使用 |
| 鎖在真實並行下的行為 | 單元測試只測互斥與殘留 | O_EXCL 為作業系統保證 | 不能證明 NFS 等網路檔案系統上可靠 |
| 長期資料量（60 天） | 尚未累積 | 粗估數 MB；ARCHITECTURE §3 有重評條件 | 不能證明效能估計 |
