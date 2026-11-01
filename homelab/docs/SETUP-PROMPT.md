# Cubi N9 基礎服務架設 Prompt

**用法**：在 Cubi N9 上開一個 Claude Code session，把下方 `---` 之間的內容整段貼給它。
它會先盤點、再問你幾個決定，才開始動手。過程中它會請你提供 Anthropic API key、在瀏覽器建立 Uptime Kuma 管理員帳號，這兩件事刻意留給你自己做。

---

你是在我家中常駐小電腦（Cubi N9，24 小時開機的 Linux 伺服器）上工作的系統管理 agent。請幫我架設三個基礎服務，並把架構整理成可長期維護的形式：

1. **Caddy**：反向代理與自動 HTTPS，所有網頁服務的單一入口
2. **Uptime Kuma**：服務監控，之後我自己寫的排程程式也會用它的 Push 監控回報
3. **LiteLLM Proxy**（搭配 PostgreSQL）：所有自架應用共用的 Claude API 閘道，每個應用一把虛擬金鑰並設預算上限

背景：我已經有 Tailscale（這台機器在我的 tailnet 裡）和密碼管理器；這台機器**已經裝好 Docker，也已經有一個 PostgreSQL 在跑**；我的網域是 `chlinandys.com`，DNS 代管在 **AWS Route 53**。這些服務**只給我自己在家裡網路與 Tailscale 內使用，不對公網開放**。之後還會在這台機器上跑 `scout`（我的科技新聞 CLI，產生靜態網頁），以及其他自架服務。

## 硬規則

1. **先盤點、後動手**。Phase 0 全部唯讀；盤點完先回報並問我決定，得到答覆才進 Phase 1。
2. **不碰既有服務**：發現已在使用 80/443 的程式、既有 Docker 容器、既有反向代理時，回報並詢問，不要停用、覆蓋或刪除。
3. **不對公網開放**：不設定路由器 port forwarding、不開 Tailscale Funnel、不使用 Cloudflare Tunnel 這類對外通道。
4. **防火牆變更要先確認**：啟用或修改 ufw／nftables 之前，先確定 SSH 與 Tailscale 的放行規則已在，並把完整規則給我確認；這類錯誤會把我鎖在門外。
5. **祕密不出現在對話與 git**：API key、master key、資料庫密碼只寫進權限 600 的 `.env`；回報時遮罩（例如 `sk-…a1b2`）。Anthropic API key 由我親自填進檔案，或我貼給你後你寫入檔案，之後不要再印出來。
6. **版本要釘住**：所有 image 用明確的版本 tag，不用 `latest` 或 `main-latest`；選版本前查該專案的 release 頁與安全公告（LiteLLM 在 2026 年有過安全事件，請特別確認你選的版本已修補）。
7. **不可逆操作**（刪 volume、刪資料庫、`docker system prune`）一律先列出具體名稱讓我確認。
8. **不確定就查官方文件**：Caddy 的 Tailscale 憑證整合、LiteLLM 的設定鍵名與健康檢查路徑、Uptime Kuma 2.x 的 image tag，都以當下官方文件為準，不要憑記憶。查不到或和本 prompt 描述不同時，照文件做並告訴我差異。

## Phase 0：盤點（唯讀）

收集並整理成一張表回報：

- OS 版本、kernel、CPU、RAM、磁碟（可用空間、掛載點）
- Docker 與 Compose plugin 是否已安裝、版本；目前有哪些容器與 volume
- 正在監聽的 port（`ss -tlnp`），特別是 80、443、4000、3001、5432
- Tailscale：`tailscale status`、本機的 tailnet IP、MagicDNS 名稱（`*.ts.net`）、是否已啟用 HTTPS 憑證功能
- 防火牆：ufw／nftables 現況
- 區網 IP 與網段
- 是否已有 `/opt/homelab` 或類似目錄

接著問我以下問題（附上你根據盤點結果的建議）：

**Q1. HTTPS 憑證方案**

| 方案 | 網址長相 | 需要 | 取捨 |
|---|---|---|---|
| A. Tailscale 憑證 | `https://cubi.<tailnet>.ts.net`、`https://cubi.<tailnet>.ts.net:8443` | tailnet 開啟 HTTPS；Caddy 能存取 tailscaled socket | 不用自己的網域；每台機器只有一個主機名，服務只能用不同 port 區分 |
| B. 自有網域 + DNS-01 wildcard | `https://status.home.chlinandys.com`、`https://llm.home.chlinandys.com` | 用 xcaddy 編進 Route 53 DNS 模組（`caddy-dns/route53`）；一組權限最小的 AWS IAM 憑證 | 網址最好記、每個服務一個子網域；DNS 記錄指向 tailnet IP（100.x），外網解析得到但連不上 |
| C. Caddy 內部 CA（`tls internal`） | 任意內部名稱 | 每台裝置手動信任 Caddy 根憑證 | 不依賴外部；手機等裝置設定麻煩 |

**我傾向 B**（`*.home.chlinandys.com`）。請確認：
- `home.chlinandys.com` 底下目前沒有會衝突的記錄（用 `dig` 查，不要動既有記錄）。
- IAM 政策只允許這個 hosted zone 的 `route53:ChangeResourceRecordSets`、`route53:ListResourceRecordSets`，加上 `route53:GetChange` 與 `route53:ListHostedZonesByName`；把政策 JSON 給我，由我在 AWS console 建立 IAM user 和 access key，再填進 `.env`。
- wildcard 記錄 `*.home.chlinandys.com` 指向這台機器的 tailnet IP，由我手動在 Route 53 建立，或我同意後由你用同一組憑證建立（那需要把 ChangeResourceRecordSets 用在 A 記錄上，政策相同）。
- 我的手機與筆電透過 Tailscale 連線時能解析並連上（tailnet IP 只有在 tailnet 內可達）。

B 卡住時退回 A。

**Q2. 既有 PostgreSQL 的用法**。盤點它的版本、是容器還是系統服務、資料在哪、目前有哪些資料庫與使用者、是否有備份。我的預設是**沿用它**：為 LiteLLM 建立專屬的資料庫與權限最小的使用者（只擁有自己的資料庫），不碰既有資料庫；LiteLLM 透過 Docker network 連線（必要時把既有容器加入 `homelab` network，或用 host 位址）。如果版本太舊、它是別的專案私有的、或連線方式會迫使我改動既有服務，就回報並建議改開獨立的 PostgreSQL 容器，由我決定。

**Q3.** 每個服務的子網域或 port 安排（給出你的建議表）。

**Q4.** 防火牆要不要啟用、要放行哪些來源（建議：SSH、tailscale0 介面、區網網段；其他拒絕）。

## Phase 1：目錄結構

全部放在 `/opt/homelab`，做成本機 git repo（`.env` 與資料目錄列入 `.gitignore`），讓設定變更可追溯：

```
/opt/homelab/
├── README.md              # 總覽：服務、網址、port、操作方式
├── AGENTS.md              # 給未來 agent 的規則與陷阱（新增服務流程、禁止事項）
├── CLAUDE.md              # 只有一行 @AGENTS.md
├── compose.yaml           # 單一 Compose 專案；共用 network: homelab
├── .env                   # 祕密（chmod 600，不進 git）
├── .env.example           # 同樣的鍵、假值（進 git）
├── caddy/
│   ├── Caddyfile
│   └── Dockerfile         # 方案 B：xcaddy 加 caddy-dns/route53
├── litellm/
│   └── config.yaml
├── scripts/
│   └── backup.sh
├── docs/
│   ├── ARCHITECTURE.md    # 架構圖、資料流、為什麼這樣設計
│   └── PROGRESS.md        # §0 接手速覽 + 驗證紀錄
└── data/                  # bind mount 的資料（不進 git）
    ├── caddy/ uptime-kuma/      # 沿用既有 PostgreSQL 時不需要 postgres/
```

文件一律繁體中文，技術名詞保留英文。

## Phase 2：Caddy

- 只有 Caddy 對外發布 port（80、443，加上方案 A 的額外 port）；其他服務只在 `homelab` Docker network 內，**不發布到 host**。
- 依 Q1 設定 TLS。方案 A：讓 Caddy 從 tailscaled 取得 `*.ts.net` 憑證（掛載 tailscaled socket，必要時設定 tailscaled 允許該使用者取憑證，以官方文件為準）。方案 B：用 xcaddy 編譯含 `caddy-dns/route53` 的 image（Dockerfile 裡釘住 Caddy 與模組版本），AWS 憑證以環境變數（`AWS_ACCESS_KEY_ID`、`AWS_SECRET_ACCESS_KEY`、`AWS_REGION`）放 `.env`；用一張 wildcard 憑證 `*.home.chlinandys.com` 涵蓋所有服務，避免每個子網域都出現在憑證透明度紀錄。
- Caddyfile 用 snippet 抽出共用設定（壓縮、安全標頭、存取紀錄），每個服務一個 site block，方便之後新增。
- 預留一個**註解掉**的 `scout` 靜態網站區塊：`file_server`，root 指向 `/srv/scout-site`（bind mount 到 host 的 `/opt/homelab/data/scout-site`），等我部署 scout 後再啟用。

## Phase 3：Uptime Kuma

- 用 2.x 的明確版本 tag；資料 bind mount 到 `data/uptime-kuma`。
- 透過 Caddy 提供（例如 `status.…`）。
- 管理員帳號由我在瀏覽器建立；準備好後告訴我網址，等我建立完成再繼續。
- Uptime Kuma 沒有建立監控項目的 REST API，所以請在 `README.md` 列出建議的監控清單，我照著在介面上新增：
  - 每個 Caddy 網址（HTTPS，順便監看憑證到期）
  - LiteLLM 健康檢查（從 Kuma 容器內打 `http://litellm:4000` 的 liveliness 路徑，以文件為準）
  - PostgreSQL（TCP，位址依 Q2 的連線方式）
  - 一個 Push 監控範例，並說明之後我的排程程式怎麼回報（`curl` Push URL）
- 通知管道先不設；我目前透過 Claude session 溝通，之後再決定要不要接 Telegram。

## Phase 4：LiteLLM（使用 PostgreSQL）

- PostgreSQL：依 Q2。沿用既有實例時，建立 `litellm` 資料庫與同名使用者（隨機密碼放 `.env`），只授權該資料庫，**不修改既有資料庫、使用者與設定檔**；若需要改 `pg_hba.conf` 或 `listen_addresses` 才連得到，先回報。改開獨立容器時：明確版本的官方 image、資料 bind mount、不發布 port。
- LiteLLM：明確版本的官方 image，連到 PostgreSQL（虛擬金鑰與花費紀錄需要資料庫）。
  - `.env`：`ANTHROPIC_API_KEY`（我提供）、`LITELLM_MASTER_KEY`（隨機產生，`sk-` 開頭）、`LITELLM_SALT_KEY`（隨機產生，用來加密存進資料庫的憑證，**設定後不可更改**，要在文件中強調）、`DATABASE_URL`。
  - 如果該版本支援關閉 telemetry，就關閉。
- `config.yaml` 的 model_list 至少包含三個 Claude 模型，對外名稱與 Anthropic model ID 相同：
  - `claude-opus-5-5`
  - `claude-sonnet-5-5`
  - `claude-haiku-5-5`

  先確認該 LiteLLM 版本的價格表認得這三個模型；**認不得就在 config 明確填入每 token 價格**，否則花費會被記成 0，預算上限形同虛設。每百萬 token 的價格（輸入／輸出）：Opus 5.5 $4／$20、Sonnet 5.5 $2／$10、Haiku 5.5 $0.10／$0.50（填入時換算成每 token，例如 `4e-6`）。
- 透過 Caddy 提供（例如 `llm.…`）；管理 UI 也走這個入口。
- 建立一把測試用虛擬金鑰（預算例如 $1、週期 30 天，只允許 `claude-haiku-5-5`），用它實際打一次 `/v1/chat/completions`，確認：
  1. 回應正常
  2. 該金鑰的花費有增加（查 key info）
  3. 用錯的金鑰會被拒絕
- 在 `README.md` 寫下「新應用要用 Claude 時怎麼申請虛擬金鑰」的步驟（指令範例，金鑰遮罩），包含預算與允許模型的設定。

## Phase 5：備份與更新

- `scripts/backup.sh`：`pg_dump` LiteLLM 資料庫（只備份 `litellm`，既有資料庫的備份另外問我）、打包 `data/caddy`、`data/uptime-kuma` 與 `.env`，存到 `/opt/homelab/backups/`，保留最近 7 份。**先手動跑一次並驗證可以還原**（至少確認 dump 檔能被 `pg_restore --list` 讀取）。要不要設成每日排程、要不要另存一份到別台機器，問我。
- 更新流程寫進 `README.md`：改 tag → `docker compose pull` → `docker compose up -d` → 檢查健康狀態 → commit。**不要**裝 Watchtower 之類的自動更新。

## Phase 6：驗證（全部實測，結果寫進 `docs/PROGRESS.md`）

| 項目 | 怎麼驗 |
|---|---|
| HTTPS | 從這台機器和透過 Tailscale 各 `curl -v` 每個網址：憑證有效、主機名正確、HTTP 自動轉 HTTPS |
| 不對公網開放 | `ss -tlnp` 只看到預期的 port；Postgres、LiteLLM、Kuma 沒有發布到 host |
| LiteLLM | Phase 4 的三項測試結果（附 token 數與記錄到的花費） |
| 重開機存活 | `docker compose restart` 後全部恢復；我同意的話再實際 reboot 一次 |
| 備份 | 備份檔存在、大小合理、dump 可被讀取 |
| 安全掃描（選做） | 如果這台有裝 `syssec`，跑 `syssec scan --json`，把對外 port（NET-002）與新出現的問題列給我 |

沒驗到的項目放在「已知未覆蓋」小節：寫明沒驗什麼、為什麼、它不能證明什麼。

## Phase 7：文件與交接

- `README.md`：服務表（名稱、網址、內部位址、image 版本、資料位置）、日常操作（啟動、停止、看 log、更新、備份還原）、新增服務的步驟（compose 服務 → Caddy site block → Kuma 監控 → 需要時申請 LiteLLM 虛擬金鑰）、Kuma 建議監控清單、LiteLLM 虛擬金鑰申請方式。
- `AGENTS.md`：給之後在這台機器工作的 agent 的硬規則（本 prompt 的硬規則 1–8 精簡版）與實際踩到的陷阱。
- `docs/ARCHITECTURE.md`：架構圖（文字或 Mermaid）、網路拓撲（tailnet／區網／Docker network）、TLS 方案與選擇理由、各服務資料流。
- `docs/PROGRESS.md`：開頭 `§0 接手速覽`，然後是 Phase 6 的驗證紀錄。
- 全部 commit 到 `/opt/homelab` 的本機 git repo（commit 訊息格式：`<type>: <模組中文名>. <描述>`）。

完成後用以下格式回報給我：

1. 服務與網址總表
2. 我需要親自做的事（例如：建立 Kuma 帳號、在 Kuma 加監控、把 API key 填進 `.env`、決定備份排程）
3. 驗證結果摘要（通過幾項、有哪些未覆蓋）
4. 和本 prompt 不一致的地方、你做的判斷及理由

---
