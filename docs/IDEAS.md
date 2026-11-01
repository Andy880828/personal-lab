# 專案點子歸檔

2026-11-01 與 Claude 討論後整理。方向：**self-host + LLM API**，跑在家中常駐小電腦（Cubi N9）。
已做的：`scout`（資訊雷達）、`homelab`（Caddy + Uptime Kuma + LiteLLM）。其餘先歸檔，之後再挑。

前提（使用者確認過）：4060 電腦不常開（不做 GPU 運算類）；Claude Code 雲端額度只用於開發；有 Claude API；已有 Tailscale 與密碼管理器；有魚缸；沒有 ESP32、沒有攝影機、不做陽台觀察。

## 熱門類

| # | 點子 | LLM 負責 | Claude 支援 | 難度 |
|---|---|---|---|---|
| 1 | **文件中樞**：Paperless-ngx + paperless-gpt（輕量版可用 Papra） | 看圖 OCR、自動標題／標籤／分類（收據、保單、合約） | ✅ paperless-gpt 原生支援 Anthropic | ⭐⭐ |
| 2 | **AI 記帳**：Actual Budget + actual-ai | 交易自動分類、查商家、建議新類別 | ✅ | ⭐⭐ |
| 3 | **網頁變動監控**：changedetection.io | 白話規則判斷哪些變動重要（價格、補貨、公告、法規） | ⚠️ 經 LiteLLM 支援，開源版是否已有待確認 | ⭐ |
| 4 | **AI 書籤**：Karakeep | 自動標籤與摘要 | ⚠️ 官方列 OpenAI／Ollama；閒置吃記憶體 | ⭐ |
| 5 | **自架 Claude 聊天介面**：LibreChat（MIT）或 Open WebUI | 家人共用 API key、分帳號 | ✅ LibreChat 多供應商 | ⭐⭐ |
| 6 | **自動化平台**：n8n | 串接上面各服務 | ✅ | ⭐⭐ |
| 7 | **信件分流**：Inbox Zero | Gmail／Outlook 分類、草擬回覆、退訂 | ⚠️ AGPL＋商用限制，個人可用 | ⭐⭐⭐ |

## 冷門類

| # | 點子 | 有意思的地方 | 難度 |
|---|---|---|---|
| 8 | **家庭物品清冊**：Homebox + Homebox Companion | 拍照自動建檔（品牌、型號、序號）；官方只測 OpenAI | ⭐⭐ |
| 9 | **水族紀錄 app**：fork WaterScribe（Flask）或自己寫 | 水質趨勢讓 Claude 寫養護建議、拍試紙照片自動讀值；結合海洋生物背景 | ⭐⭐ |
| 10 | **個人 MCP Hub** | 把自架服務（文件、記帳、水族）包成 MCP，Claude 直接問「去年保險費多少」 | ⭐⭐⭐ |
| 11 | **私人 AI 搜尋**：Perplexica + SearXNG | 自架的 Perplexity；無內建登入，要放在 Caddy 驗證後面 | ⭐⭐ |
| 12 | **AI 日記** | 成熟自架方案幾乎空白，適合自己做：每週回顧心情與主題 | ⭐⭐ |
| 13 | **Home Assistant + 官方 Anthropic 整合** | 未來接 IoT（市售 Zigbee 感測器即可，不一定要 ESP32） | ⭐⭐ |

## 建議組合（之後挑一組）

- **個人資料中樞**：#1 文件 + #2 記帳 + #8 物品清冊，用 #10 MCP Hub 串起來
- **資訊雷達延伸**：scout + #3 網頁監控 + #4 書籤 + #11 AI 搜尋
- **自己動手做**：#9 水族 app + #12 AI 日記

## 已排除

Immich（照片放雲端成本高，且 GPU 機不常開）、BirdNET-Go（不做陽台）、Frigate／攝影機 AI（沒有攝影機）、4060 算力佇列（不做 GPU 運算）、T-Pot 蜜罐（需要獨立大台 VM）。
