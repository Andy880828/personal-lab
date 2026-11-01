# scout 架構

## 流程

```
 watchlist.json ──▶ fetch ──(平行、條件式 GET)──▶ feed.Parse ──▶ feed.Merge ──▶ state.json
                                                   RSS/Atom/RDF/HN    URL 正規化去重
                                                                      跨來源 sighting
 state.json ──▶ collectCandidates ──▶ rank.Score（附 reasons）──▶ candidates（agent 讀）
                  時間窗：上一份 digest 終點起                            │
                                                                         ▼
                                         agent：read 全文 → 挑 4–7 篇 → 寫一句話繁中摘要
                                                                         │
 digests/YYYY-MM-DD.json ◀── digest create（驗證 ID、字數、重複）◀────────┘
        │
        ├──▶ markdown → agent 貼回 Claude 對話
        └──▶ digest export → 靜態 HTML → Caddy file_server
```

## 模組

| 套件 | 責任 |
|---|---|
| `core` | 資料型別與列舉（來源種類、抓取狀態），SchemaVersion |
| `store` | SCOUT_HOME 檔案、原子寫入、寫入鎖、預設 watchlist |
| `feed` | HTTP（條件式請求、大小上限）、解析、URL 正規化、合併 |
| `rank` | 評分規則與時間窗判斷（唯一決定點） |
| `trends` | 兩個時間窗的標題詞頻比較 |
| `extract` | HTML → 正文純文字（啟發式，無 parser 依賴） |
| `digest` | 提交驗證、Markdown／純文字／HTML 輸出 |
| `cli` | 指令樹、envelope、錯誤分類、describe |

## 設計決策

### 1. 判斷交給 agent，scout 只做確定性工作

原文的流程是「agent 抓 → 挑 → LLM 摘要 → 推送」。拆開後，抓取、去重、排序、儲存、輸出都是確定性的，適合做成可測試的 CLI；挑選與摘要需要語意判斷，而操作它的 Claude session 本來就具備，而且還有對話上下文（知道使用者最近在意什麼、可以追問）。所以 scout 不內嵌 LLM 呼叫，也不需要 API key 或 LiteLLM。

### 2. 參考但不依賴的現有工具

| 工具 | 借鏡 | 不採用 |
|---|---|---|
| Threads 原文模板 | 來源組合（HN／媒體／追蹤的人）、每日 4–7 篇、一句話繁中摘要、網站版 | Slack 推送（改由對話回報） |
| RSSbrew | 多來源聚合 + 過濾規則 + digest | Django + DB、內建 LLM 呼叫 |
| Matcha | 每日 markdown digest、單一 binary | 內建 OpenAI 呼叫、無去重與排序理由 |
| FreshRSS / Miniflux | 條件式 GET、來源健康度 | 完整閱讀器 UI、資料庫 |
| changedetection.io | 「讓 LLM 判斷哪些變化重要」的分工 | 監控任意網頁的範圍 |
| Hacker News 排名 | 分數 × 時間衰減 | 不公開的懲罰因子 |
| syssec（同作者） | agent-native CLI：envelope、error code + hint、describe、status + next_steps、SKILLS.md | SQLite（資料量小，JSON 檔即可，換來零依賴） |

### 3. 儲存：JSON 檔而非 SQLite

資料量是「每天數百篇、保留 60 天」，查詢都是全量掃描後排序，JSON 檔加原子寫入（暫存檔 + rename）已足夠，換來零依賴。並行寫入用 `scout.lock`（O_EXCL，10 分鐘視為殘留）。重評條件：state.json 超過 20 MB，或 `candidates` 超過 500 ms。

### 4. 去重：網址正規化

ID = sha256(host 去 www + path 去尾斜線 + 排序後且去追蹤參數的 query) 前 12 碼，不含 scheme。同一篇文章從 HN 和原部落格進來會合併成一筆，`sightings` 記錄所有來源，「被幾個來源報導」本身就是排序訊號。標題相似度去重刻意不做（誤合併的代價高於重複），重評條件見 ROADMAP。

### 5. 時間窗：首尾相接

候選起點 = 上一份 digest 的 `window_to`；沒有 digest 時為 24 小時前。條件是 `first_seen ≥ 起點` 且（發布時間未知，或發布時間 ≥ 起點 − 72 小時）。前者保證不重複，後者接住「昨天發、今天才被 feed 列出」的文章，同時擋掉第一次抓取時 feed 裡的舊文。

### 6. 排序：公開的加總規則

來源權重、追蹤的人、多來源報導、HN 分數與留言（取 log）、興趣關鍵字（有上限），再乘上 36 小時半衰期的時間衰減。每一項都寫進 `reasons`。分數只是預設排序，SKILLS.md 要求 agent 以多樣性與重要性做最後判斷。

### 7. 全文擷取：啟發式而非 readability

優先 `<article>`，其次 `<main>`、`<body>`，先移除 script／style／nav／footer／aside。不引入 HTML parser 依賴；可信度透過 `container`、`truncated`、字數過少等 warnings 交給 agent 判斷。

## 不做的事

- **MCP server**：多一層常駐程序與另一套 schema；「CLI + 穩定 JSON」已足夠讓 agent 操作。
- **內建排程**：`fetch` 交給 cron／systemd timer；digest 需要 agent，由使用者的 Claude session 觸發。
- **`--jq`**：agent 可自行 pipe 給 jq／python。
