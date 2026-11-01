# homelab

家中常駐小電腦（Cubi N9）的基礎服務：**Caddy**（反向代理＋HTTPS）、**Uptime Kuma**（監控）、**LiteLLM**（Claude API 閘道，虛擬金鑰＋預算）。

這個目錄**不放實際設定檔**：設定與祕密只存在小電腦的 `/opt/homelab`（本機 git repo），由那台機器上的 agent 依 prompt 架設與維護。

| 檔案 | 用途 |
|---|---|
| [`docs/SETUP-PROMPT.md`](docs/SETUP-PROMPT.md) | 貼給小電腦上 Claude session 的架設 prompt（盤點 → 詢問 → 架設 → 驗證 → 文件） |

## 和其他專案的關係

- `scout` 不使用 LiteLLM：它不呼叫 LLM，由 Claude session 直接操作。scout 的靜態網頁之後交給 Caddy 的 `file_server` 提供（prompt 已預留區塊）。
- 之後的自架 LLM 應用（文件歸檔、記帳分類等，見 [`../docs/IDEAS.md`](../docs/IDEAS.md)）各自向 LiteLLM 申請虛擬金鑰。

## 前置：DNS 與註冊商搬到 Cloudflare（架設前完成）

現況（2026-11-01 查詢）：註冊商 **GoDaddy**、**網域 2026-11-17 到期**、未啟用 DNSSEC。
**Route 53 hosted zone 已停用，但註冊局仍指向 AWS nameserver → 網域目前完全無法解析（SERVFAIL）**，要盡快完成步驟 1–2。

| # | 步驟 | 狀態 | 注意 |
|---|---|---|---|
| 1 | Cloudflare 免費方案新增 `chlinandys.com`，補回需要的記錄 | ⬜ | 指向 tailnet／內網 IP 的記錄設 **DNS only** |
| 2 | GoDaddy 關閉 Domain Lock，nameserver 改成 Cloudflare 指定的兩組 | ⬜ | 確認 WHOIS 聯絡信箱收得到信 |
| 3 | Cloudflare 顯示 Active | ⬜ | 驗證：`dig NS chlinandys.com` 回 Cloudflare |
| 4 | GoDaddy 取得 EPP 授權碼 | ⬜ | 轉移鎖保持關閉 |
| 5 | Cloudflare → Transfer Domains，貼 EPP 碼並付款（加一年） | ⬜ | **11/12 前未開始就先在 GoDaddy 續約** |
| 6 | 到 GoDaddy 待轉移頁面手動同意 | ⬜ | 否則最長等 5 天；驗證：RDAP 註冊商變成 Cloudflare |
| — | Route 53 hosted zone | ✅ 已停用 | |
