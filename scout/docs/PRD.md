# scout PRD

## 問題

使用者擔心被科技發展拋下，但被動滑資訊流會 FOMO、浪費時間，也被演算法決定看什麼。需要一個**主動 pull、來源自選、每天只讀 4–7 篇**的習慣。

靈感來源（Threads 貼文，2026）：作者每天早上用自製 agent 抓取 HN、TechCrunch、The Verge 與十幾位追蹤的作者（Karpathy、Simon Willison、Geoffrey Huntley…），挑出值得讀的幾篇，用 LLM 寫一句話繁中摘要推送到 Slack，並有每日 4–7 篇的網站版。核心主張：自己決定資訊來源，才能把主導權拿回來。

## 使用者

1. **第一使用者是 AI agent**：使用者家中常駐小電腦（Cubi N9）上的 Claude session，負責執行、挑選、摘要、回報。
2. 其次是使用者本人：在 Claude 對話中收到精選、追問、調整來源；偶爾直接下指令。

## 目標

| # | 目標 | 驗收方式 |
|---|---|---|
| G1 | 每天一份 4–7 篇、附一句話繁中摘要的精選，直接出現在對話中 | `digest create` 回傳 `markdown`，agent 原樣呈現 |
| G2 | 來源完全自選且可維護 | `sources add` 先實測抓取；`sources list` 回報健康度 |
| G3 | 不重複、不遺漏 | 候選時間窗從上一份 digest 的終點接續；同一篇文章跨來源合併 |
| G4 | 排序可解釋 | 每個候選附 `reasons`，規則由 `describe` 輸出 |
| G5 | 看得到趨勢，不只是單日新聞 | `trends` 比較兩個時間窗的詞頻 |
| G6 | 低維運成本 | 零外部依賴、單一 binary、JSON 檔儲存 |
| G7 | 可選的網頁閱讀版 | `digest export` 產生靜態網站，交給 Caddy 提供 |

## 非目標

- **不呼叫 LLM API**：挑選與摘要由操作它的 Claude session 完成（不接 LiteLLM、不帶 API key）。
- **不內建通知**：Telegram／Slack／email 等使用者認為有必要時再評估（見 ROADMAP）。
- **不做 GUI、不做 MCP server、不做多使用者**。
- 不做全文 RSS 閱讀器：只存標題、連結、短摘要，要讀全文用 `read` 或開原網址。

## 限制與假設

- 部署在 x86-64 Linux 小電腦，24 小時開機；網路正常但可能偶爾斷線。
- 資料量：每天數百篇候選，60 天保留期內 state.json 約數 MB。
- 使用者主要語言為繁體中文；來源多為英文。
