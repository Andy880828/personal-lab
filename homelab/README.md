# homelab

家中常駐小電腦（Cubi N9）的基礎服務：**Caddy**（反向代理＋HTTPS）、**Uptime Kuma**（監控）、**LiteLLM**（Claude API 閘道，虛擬金鑰＋預算）。

這個目錄**不放實際設定檔**：設定與祕密只存在小電腦的 `/opt/homelab`（本機 git repo），由那台機器上的 agent 依 prompt 架設與維護。

| 檔案 | 用途 |
|---|---|
| [`docs/SETUP-PROMPT.md`](docs/SETUP-PROMPT.md) | 貼給小電腦上 Claude session 的架設 prompt（盤點 → 詢問 → 架設 → 驗證 → 文件） |

## 和其他專案的關係

- `scout` 不使用 LiteLLM：它不呼叫 LLM，由 Claude session 直接操作。scout 的靜態網頁之後交給 Caddy 的 `file_server` 提供（prompt 已預留區塊）。
- 之後的自架 LLM 應用（文件歸檔、記帳分類等，見 [`../docs/IDEAS.md`](../docs/IDEAS.md)）各自向 LiteLLM 申請虛擬金鑰。

## 前置：DNS 搬到 Cloudflare（架設前完成）

現況（2026-11-01 查詢）：註冊商 **GoDaddy**、DNS 代管 **Route 53**、未啟用 DNSSEC、apex 沒有 A／MX／TXT 記錄，**網域 2026-11-17 到期**。

| # | 步驟 | 注意 |
|---|---|---|
| 0 | 到 GoDaddy 確認續約狀態 | RDAP 顯示 `client renew prohibited`；先確認不會在 11/17 失效 |
| 1 | 匯出 Route 53 的完整記錄（`aws route53 list-resource-record-sets` 或 console） | 公開 DNS 查不到子網域，以匯出結果為準 |
| 2 | Cloudflare 免費方案新增 `chlinandys.com`，逐筆核對記錄 | 指向 tailnet／內網 IP 的記錄設 **DNS only** |
| 3 | GoDaddy 解除鎖定，把 nameserver 改成 Cloudflare 給的兩組 | 未啟用 DNSSEC，不必先撤 DS 記錄 |
| 4 | Cloudflare 顯示 Active、`dig NS chlinandys.com` 回 Cloudflare 後，觀察 48 小時再刪 Route 53 hosted zone | 刪 zone 是不可逆操作 |
| 5 | （選擇）把註冊商轉到 Cloudflare Registrar | 成本價續約；需先用 Cloudflare DNS；轉移會加一年。若 11/12 前來不及完成，先在 GoDaddy 續約再轉 |
