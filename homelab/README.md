# homelab

家中常駐小電腦（Cubi N9）的基礎服務：**Caddy**（反向代理＋HTTPS）、**Uptime Kuma**（監控）、**LiteLLM**（Claude API 閘道，虛擬金鑰＋預算）。

這個目錄**不放實際設定檔**：設定與祕密只存在小電腦的 `/opt/homelab`（本機 git repo），由那台機器上的 agent 依 prompt 架設與維護。

| 檔案 | 用途 |
|---|---|
| [`docs/SETUP-PROMPT.md`](docs/SETUP-PROMPT.md) | 貼給小電腦上 Claude session 的架設 prompt（盤點 → 詢問 → 架設 → 驗證 → 文件） |

## 和其他專案的關係

- `scout` 不使用 LiteLLM：它不呼叫 LLM，由 Claude session 直接操作。scout 的靜態網頁之後交給 Caddy 的 `file_server` 提供（prompt 已預留區塊）。
- 之後的自架 LLM 應用（文件歸檔、記帳分類等，見 [`../docs/IDEAS.md`](../docs/IDEAS.md)）各自向 LiteLLM 申請虛擬金鑰。

## 網域

`chlinandy.com`：2026-11-01 在 **Cloudflare Registrar** 註冊，DNS 也由 Cloudflare 代管，2027-11-01 到期（建議開啟自動續約）。
驗證（2026-11-01）：RDAP 註冊商 Cloudflare、nameserver `clara`／`curt.ns.cloudflare.com`，公共 DNS 解析正常；DNSSEC 未啟用（可在 Cloudflare 後台一鍵開啟）。

舊網域 `chlinandys.com`（GoDaddy，2026-11-17 到期，Route 53 已停用）不再使用。
