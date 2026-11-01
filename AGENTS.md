# personal-lab — repo 規則

跨專案工程習慣見 `~/.claude/AGENTS.md`（由 andy-init 安裝），這裡只寫 repo 層級的規則。

## 結構

這是多個個人專案的集合，每個子目錄是獨立專案，有自己的 `AGENTS.md`，進到子目錄工作時以它為準。

| 目錄 | 內容 |
|---|---|
| `scout/` | 給 agent 操作的科技資訊雷達 CLI（Go、零依賴） |
| `homelab/` | 小電腦基礎服務（Caddy、Uptime Kuma、LiteLLM）的架設 prompt |
| `docs/IDEAS.md` | 歸檔中的專案點子，之後再挑 |

## 規則

- 子專案之間不互相 import；共用的東西先問「屬於誰」，不要建 `common/`。
- 根目錄只放 `README.md`、`AGENTS.md`、`CLAUDE.md`；其餘進 `docs/` 或子專案。
- 祕密（API key、token、密碼）永遠不進 repo；homelab 的實際設定只存在小電腦上。
- CI 依子目錄分開（`.github/workflows/<專案>.yml`，用 `paths` 過濾）。
