# scout ROADMAP

## v0.1（目前）

✅ 抓取（RSS/Atom/RDF/HN）、條件式 GET、跨來源去重、可解釋排序、首尾相接的時間窗、全文擷取、digest 建立／顯示／列表／靜態網站、標題趨勢、來源與關鍵字管理、prune、describe。

## 候選功能（每項附重評觸發條件）

| 功能 | 說明 | 觸發條件 |
|---|---|---|
| Telegram 推送 | `digest show --format text` 已是可直接送出的純文字；屆時加一個獨立的 `notify` 子指令或交給 agent 呼叫 Bot API | 使用者明確提出需要離開對話也收到精選 |
| 偏好學習 | `scout feedback <id> --like/--dislike`，依回饋調整來源權重 | 使用者連續兩週以上手動調整同一類關鍵字／權重 |
| 標題相似度去重 | 同一則新聞不同網址（媒體轉述）合併 | agent 在 digest 中連續 3 天以上遇到同事件多網址重複候選 |
| OPML 匯入／匯出 | 從其他閱讀器搬移來源 | 使用者要匯入超過 10 個來源 |
| GitHub Trending／Releases 專用來源 | 目前用 releases.atom 即可；Trending 無官方 API | 使用者要求追蹤 trending repo |
| 中文標題斷詞 | trends 目前略過 CJK 標題 | `skipped_cjk_titles` 長期佔本期文章 30% 以上 |
| 每週回顧 digest | 從一週 digest 與 trends 產生週報 | 使用者要求週報 |
| SQLite 儲存 | 見 ARCHITECTURE §3 | state.json > 20 MB 或 candidates > 500 ms |
