# scout — 專案規則

跨專案工程習慣見 `~/.claude/AGENTS.md`（由 andy-init 安裝），這裡只寫本專案特有的規則與陷阱。
使用方式（給操作 scout 的 agent）在 `SKILLS.md`；設計理由在 `docs/ARCHITECTURE.md`，被指引才讀。

## 指令

```bash
make build        # bin/scout
make test         # go test -race ./...（全部離線，用 httptest）
make vet          # linux + darwin + windows
```

改完程式必跑 `make test && make vet`；動到 fetch／parse 再用真實來源跑一次：
`SCOUT_HOME=$(mktemp -d) ./bin/scout init && ./bin/scout fetch`，確認每個來源的 entries 數不是 0。

## 架構硬規則

- **零外部依賴**：只用 Go 標準庫。要加依賴前先寫下「標準庫為什麼做不到」並取得使用者同意。
- **scout 不呼叫 LLM、不通知人**：判斷、摘要、回報都是 agent 的工作。不要加 LLM SDK、Telegram、email、MCP server；要接通知管道時由使用者明確提出，並在 `docs/ROADMAP.md` 重評。
- **單一決定點**：
  - error code 只在 `internal/cli/cli.go` 的 `ErrorCodes`；新增 code 要同時讓 `classify` 或某個 handler 實際產生它（`TestClassify` 會檢查未文件化的 code）。
  - 排序規則與常數只在 `internal/rank`，`rank.Rules()` 是 describe 的來源；改常數就是改語意，要同步更新 `rank_test.go` 的期望值。
  - 候選規則只在 `collectCandidates`（`internal/cli/shared.go`），status／fetch／candidates／digest 共用，不要另寫一份。
  - 新指令只要掛進 `Root()`，describe 與 help 會自動涵蓋（`TestUsageAndDiscovery` 會比對數量）。
- **資料誠實**：略過的項目一律計數回報（`skipped`、`invalid`、`old_skipped`、`muted_hidden`、`truncated`）；未知的發布時間是 `nil`，不是「現在」；抓取失敗進 `warnings` 與 `consecutive_failures`。
- 所有 stdout 在 `--json` 模式下必須是**單一** JSON envelope；人類訊息走 stderr。

## 已踩過的陷阱（勿重蹈）

- `xml.Decoder.AutoClose = xml.HTMLAutoClose` 會把 RSS 的 `<link>url</link>` 當成空元素，**所有 RSS 2.0 解析出 0 筆**（Atom 正常所以難察覺）。回歸測試：`TestParseRSSLinkText`。
- `&nbsp;` 解碼成 U+00A0，Go regexp 的 `\s` 不涵蓋，要先替換。回歸測試同上。
- 不要先 `strings.ToLower` 再拿索引切原字串：部分 Unicode 轉小寫後位元組長度會變，切片錯位。用 `(?i)` regexp 直接在原字串找位置（`TestUnicodeBeforeArticleDoesNotShiftOffsets`）。
- Atom `type="html"` 的 `&lt;b&gt;` 是真的標籤要去掉；字面上的 `<` 要寫成 `&amp;lt;`。
- 官方部落格 feed（OpenAI、Hugging Face）會列出上千篇完整歷史；超過 14 天的新項目不入庫（`feed.MaxEntryAgeDays`），以 `old_skipped` 回報。
- 這個雲端開發環境的 proxy 會擋 `github.com/.../releases.atom`（403），在使用者機器上正常；不要因此刪掉該來源。

## 文件

- `README.md`（人看）、`SKILLS.md`（agent 操作手冊，英文以求精確）、`docs/PRD.md`、`docs/ARCHITECTURE.md`、`docs/ROADMAP.md`、`docs/PROGRESS.md`（開頭 §0 接手速覽）。
- 改指令、flag、error code 或排序規則時，同步檢查 `SKILLS.md` 對應段落。
