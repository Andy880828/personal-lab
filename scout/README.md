# scout

**給 AI agent 操作的個人科技資訊雷達 CLI。** 單一靜態 binary、零外部依賴，抓取你自選的來源（HN、科技媒體、AI 實驗室、你追蹤的人），去重、用可解釋的規則排序；挑選和一句話繁中摘要交給 agent，結果直接在 Claude 對話裡回報給你。

靈感來自一篇 Threads 貼文：作者每天早上用自製 agent 主動 pull 資訊、挑出 4–7 篇、用 LLM 寫一句話繁中摘要再推送，重點是「**自己決定資訊來源，主導權回到自己手上**」。scout 把其中確定性的部分（抓取、去重、排序、儲存、輸出）做成穩定的 CLI，判斷與寫作留給 agent。

```text
$ scout candidates
 42.1  b0ecb253f4a7  HackerTube: YouTube built from videos recommended by HN
       Andrej Karpathy, Hacker News 首頁 · https://karpathy.bearblog.dev/hacker-tube
 23.1  f24d72617b00  Jepsen: ParticleDB 1.1.4
       Hacker News 首頁, Lobsters · https://jepsen.io/analyses/particledb-1.1.4
 ...
```

## 和一般 RSS + LLM 摘要工具的差別

| | scout | 一般 RSS 摘要工具（RSSbrew、Matcha…） |
|---|---|---|
| 誰挑、誰寫摘要 | 你的 Claude session（有上下文、能追問） | 工具內建呼叫 LLM API |
| 依賴 | 零（Go 標準庫） | DB、LLM SDK、排程器 |
| 排序 | 公開規則，每一分附理由 | 多為時間序或黑盒 |
| 通知 | 不內建，由 agent 在對話中回報 | Email／Slack／Telegram |

## 安裝

```bash
cd scout
make build            # → bin/scout（CGO_ENABLED=0，可直接複製到其他機器）
make install          # 或裝到 $GOPATH/bin
```

需要 Go 1.24 以上；`go.mod` 指定 `toolchain go1.26.10`，確保標準庫是已修補的版本。

## 讓 agent 使用

把 [`SKILLS.md`](SKILLS.md) 當成 Claude Code skill（它帶 frontmatter），或在對話中請 agent 先讀它。之後直接說：

- 「今天有什麼科技新聞？」→ agent 跑完 fetch → candidates → read → digest create，把精選貼回對話
- 「最近什麼在升溫？」→ `scout trends`，agent 歸納 3–5 個主題
- 「多關注 Rust，少一點加密貨幣」→ agent 提議調整關鍵字，經你同意後套用

## 指令

| 指令 | 用途 | 需要網路 |
|---|---|---|
| `status` | 一次看完：設定、上次抓取、失敗來源、候選數、今天是否已有 digest、下一步 | 否 |
| `init` | 寫入預設 watchlist（15 個來源） | 否 |
| `fetch` | 平行抓取、條件式請求、去重入庫，回報每個來源健康度 | 是 |
| `candidates` | 目前時間窗內未選過的文章，依分數排序並附理由 | 否 |
| `read <id\|url>` | 擷取文章正文給 agent 摘要 | 是 |
| `digest create/show/list/export` | 儲存 agent 的精選、顯示、列表、輸出靜態網站 | 否 |
| `trends` | 本期 vs 前期的標題詞頻，找出升溫主題 | 否 |
| `sources list/add/rm/enable/disable/test` | 管理來源（add 會先實測抓取） | add/test 需要 |
| `keywords list/add/rm` | 管理加分（boost）與排除（mute）關鍵字 | 否 |
| `prune` | 清理舊文章（digest 永遠保留），支援 `--dry-run` | 否 |
| `describe` / `help <cmd>` | 機器可讀的完整說明 | 否 |

每個指令都支援 `--json`，輸出固定的 envelope（`ok`／`data`／`warnings`／`error.code`＋`hint`）。

## 資料位置

全部在 `SCOUT_HOME`（預設 `~/.scout`）：`watchlist.json`（你的來源與關鍵字，建議放進 git）、`state.json`（scout 維護）、`digests/YYYY-MM-DD.json`。

`digest export --out <dir>` 產生自足的 HTML（內嵌 CSS、無 JS、支援深色模式），可以用 homelab 的 Caddy `file_server` 提供。

## 開發

```bash
make test   # go test -race ./...
make vet    # 含 darwin / windows
```

設計與取捨：[`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)；進度與驗證：[`docs/PROGRESS.md`](docs/PROGRESS.md)。
