# Personal-Page 重構追蹤

> 狀態追蹤檔。計畫在下方；每個階段完成後把 `[ ]` 改 `[x]` 並填上 commit / PR 連結。
> 實作在 `Andy880828/Personal-Page` 的分支 `rebuild/deep-sea` 進行，這裡只追蹤。

## 進度

| 階段 | 狀態 | 紀錄 |
|---|---|---|
| 計畫（本檔） | [x] 2026-11-02 | 使用者已確認四項決策 |
| 開放問題回答 | [x] 依預設 | 使用者未回覆，全部採預設；需要確認的事項見 Personal-Page `docs/REVIEW.md`「沒有採納」 |
| Phase 0 盤點與資產 | [x] 2026-11-02 | [a1f9a76](https://github.com/Andy880828/Personal-Page/commit/a1f9a76) |
| Phase 1 骨架（DOM 版 + CI） | [x] 2026-11-02 | [a1f9a76](https://github.com/Andy880828/Personal-Page/commit/a1f9a76)：首屏 JS 90 KB gzip |
| Phase 2 核心 3D 場景 | [x] 2026-11-02 | [fd923ce](https://github.com/Andy880828/Personal-Page/commit/fd923ce) |
| Phase 3 互動元素 | [x] 2026-11-02 | [fd923ce](https://github.com/Andy880828/Personal-Page/commit/fd923ce) |
| Phase 4 收斂與驗證 | [x] 2026-11-02 | [4bec028](https://github.com/Andy880828/Personal-Page/commit/4bec028)：Lighthouse 桌機 99/100/100/100、Playwright 21 項 |
| Phase 5 impeccable 視覺審視 | [x] 2026-11-02 | [3419ef2](https://github.com/Andy880828/Personal-Page/commit/3419ef2)：偵測器 8 → 0 項，修好 tier 1 不渲染 3D |
| PR 合併、Vercel production 確認 | [x] 2026-11-02 | [Andy880828/Personal-Page#1](https://github.com/Andy880828/Personal-Page/pull/1) 已合併（71305317）；`personal-page-gules.vercel.app` 已是新版，瀏覽器載入無錯誤 |
| 第二輪：修閃爍、科學層、指令列、雙螺旋、互動 | [x] 2026-11-02 | [Andy880828/Personal-Page#2](https://github.com/Andy880828/Personal-Page/pull/2) 已合併 |
| 第三輪：正式站 impeccable 審視修正 | [x] 2026-11-02 | [Andy880828/Personal-Page#3](https://github.com/Andy880828/Personal-Page/pull/3) 已合併；偵測器 0 項；紀錄在 Personal-Page `docs/REVIEW.md` |
| 第四輪：side-tab 修正 | [x] 2026-11-02 | [Andy880828/Personal-Page#4](https://github.com/Andy880828/Personal-Page/pull/4) 已合併；正式站偵測器只剩 tier 0 SVG 雙螺旋的 advisory |
| 第五輪計畫：海洋 × 代碼深度融合、分層生物、內容轉向 | [x] 2026-11-02 計畫完成、開放問題已回答 | 見下方「第五輪計畫」；尚未實作，等 Opus 派工；Sketchfab 模型由 Andy 以 glTF 格式下載到 `raw/models/` |
| 第五輪實作 Phase A 內容轉向與去 AI 味 | [x] 2026-11-02 | [c734820](https://github.com/Andy880828/Personal-Page/commit/c734820)（分支 `round5/ocean-code`）；文案 lint `tests/unit/copy.test.ts` |
| 第五輪實作 Phase B 頂端列、儀表、拿掉探測器 | [x] 2026-11-02 | [c734820](https://github.com/Andy880828/Personal-Page/commit/c734820)；nav 像素測試在 Phase E 補上 |
| 第五輪實作 Phase C 模型素材管線 | [x] 2026-11-02 | [8e9764a](https://github.com/Andy880828/Personal-Page/commit/8e9764a)：release raw-models-v1 的 10 個模型 → 合計 688 KB；攝影測量模型烘頂點色後減面；1BNA 資料 [1e2405f](https://github.com/Andy880828/Personal-Page/commit/1e2405f)；size-limit 模型兩條在 Phase E 補上 |
| 第五輪實作 Phase D 分層代表生物 | [x] 2026-11-02 | [c734820](https://github.com/Andy880828/Personal-Page/commit/c734820)；另修海床岩塊碎裂 [912404a](https://github.com/Andy880828/Personal-Page/commit/912404a)；管水母模型為 CC-BY-NC-SA，Andy 確認保留 |
| 第五輪實作 Phase E 海洋 × 代碼意象 | [x] 2026-11-02 | [cf92152](https://github.com/Andy880828/Personal-Page/commit/cf92152)：六個意象全做；size-limit 加模型兩條、頂端列像素測試、`ascii on` 冒煙測試；Playwright 互動測試本機全過（視覺基準未更新） |
| 第五輪實作 Phase F 驗證、impeccable、PR | [ ] 進行中 | 依 Andy 指示直接推 main（[2351766](https://github.com/Andy880828/Personal-Page/commit/2351766)），[Andy880828/Personal-Page#5](https://github.com/Andy880828/Personal-Page/pull/5) 因此自動標為已合併；閃爍量測腳本 `scripts/flicker.mjs` 進 repo，各位置都不比 main 差到 1.5 以上；視覺基準由 CI 產生中；合併後跑正式站 impeccable |

### 第二、三輪（使用者回饋後）

- **閃爍**：正式站海面每幀閃動。原因是海面高光用 `pow(…,420)` 的逐像素閃點加上快速雜訊，Bloom 再放大；另外排除了天空球 `xyww` 深度技巧、後製 MSAA、`backdrop-filter`、效能降級重建幾何這幾個只在真 GPU 上會閃的構造。量測方法：`?slow=0.02` 放慢場景時間後連續截圖，相鄰幀差異由 10+ 降到 <1。
- **科學層**：`src/content/science.ts`（壓力、溫度剖面、光衰減、分層帶、每段 field note 與來源）+ 右側遙測儀表（log 深度軸的溫度／光剖面）。
- **Coding**：站內指令列（`/`）、`j`/`k`。
- **技能段**：泡泡換成 DNA 雙螺旋。
- **互動**：點擊驚散魚群、聲納、深層帶游標探照燈、水母 hover。
- **正式站 impeccable 第二輪**：Nielsen 25/36；P1 全修（儀表蓋內容、停用表單、指令列與儀表相撞、手機溢出）。

### 和計畫不同的地方

| 計畫 | 實際 | 原因 |
|---|---|---|
| TypeScript 7 | TypeScript 6.0.3 | typescript-eslint 只支援到 TS 6.0 |
| react-router 8 | 自寫 40 行 History API 小路由 | 只有三個「捲到某段」的網址，省約 30 KB 首屏 JS |
| GSAP、Lenis 在首屏 | 首屏後才載入；首頁標題改純 CSS 逐字浮出 | 首屏 JS 原本 155 KB，超過 120 KB 預算 |
| 技能氣泡用 `MeshTransmissionMaterial` | three 內建 `MeshPhysicalMaterial`（transmission + iridescence） | 五顆各自一個 FBO pass 太貴；內建透射一次處理全部 |
| 段落標題上方放深度小標 | 深度刻度放在標題同一列右端 | impeccable craft-floor 明文禁止標題上方小標 |
| leva、vite-plugin-glsl | 沒用 | 除錯只需要 r3f-perf；GLSL 用 Vite 內建 `?raw` |
| 軟體算繪當 tier 1 | 軟體算繪（SwiftShader、llvmpipe）當 tier 0 | 只有幾 fps 且卡住捲動；也讓 CI 的 Lighthouse 量到首屏實況 |
| GPU 偵測在主執行緒 | Web Worker + OffscreenCanvas | 主執行緒建 WebGL context 是 TBT 的最大來源 |
| Lighthouse 手機版另跑一次 | 沒加 | 桌機 assert 已涵蓋；CI 無 GPU，手機數字參考價值低 |
| 視覺基準圖在本機產生 | 由 CI 產生（commit 訊息加 `[update-snapshots]`），本機預設不比對 | 不同機器字型光柵化有次像素差異，中文換行會跟著變 |

---

## Context

- 現況：Vue 3.5 + Vite 5 + Tailwind 3 的三頁 SPA（`/`、`/about`、`/project`），部署在 Vercel（`personal-page-gules.vercel.app`）。沒有 3D、沒有測試、沒有 CI；`src/assets/image` 9.9 MB（`bg-3.jpg` 3.8 MB、`hero-bg.png` 2.7 MB、`個人.png` 1.9 MB）；Font Awesome 由 CDN 載入 CSS+JS 兩份；EmailJS public key 寫死在 `App.vue`；`output.css` 編譯產物進 repo；`index.html` 的 `<html lang="">` 空白。
- 內容過時：About 的經歷停在「台郡科技 數據分析工程師 2025/04–2026/03」；作品只有 2024 年的四個。2025–2026 的新方向（給 AI agent 用的 CLI、自架 homelab、資料分析）完全沒出現。
- 使用者決定（2026-11-02 確認）：
  1. **原地重構 Personal-Page repo**，Vercel 部署流程不變。
  2. **改用 React 19 + react-three-fiber v9**（放棄 Vue，換取 drei / postprocessing / 21st.dev / React Bits 生態）。
  3. 視覺方向：**深海 × 數據**。海洋身分升級成即時 3D，往下捲像潛得更深。
  4. 作品：移除 AquariumFishVending、BingoBingo；保留 RecipeCube、Trim2Sort；新增 scout、homelab、syssec、andy-claude-kit、stock-analysis-app、local-transcribe、fitness_challenge、nature-footprint-studio。
- 硬性要求：非常吸睛、大量 3D 元素；**可驗證**（測試、視覺回歸、效能預算、CI）；**大小適中**（有預算、有量測）；最後用 **impeccable** 做視覺審視。

## 技術棧（版本為 2026-11-02 npm latest，實作時以 `npm view` 為準）

| 層 | 選擇 | 理由 |
|---|---|---|
| 建置 | Vite 8 + `@vitejs/plugin-react` 6 + TypeScript 7 | 現有 Vercel 設定直接沿用 |
| UI | React 19.3、`react-router` 8（單頁 + hash/section 導航，三條 route 保留） | |
| 樣式 | Tailwind 4（`@tailwindcss/vite`，CSS-first `@theme`） | 不再有 `tailwind.config.js` 與 `output.css` |
| 3D | `three` 0.187、`@react-three/fiber` 9.8、`@react-three/drei` 10.7、`@react-three/postprocessing` 3.2 | pmndrs 全家桶 |
| 2D shader 背景 | `@paper-design/shaders-react` 0.0.82（Apache-2.0，**鎖版本** `0.0.x` 會 breaking） | MeshGradient / GrainGradient / Water 類背景，超輕量，不開 3D 的段落用它 |
| 動畫 | `gsap` 3.15（+ ScrollTrigger、SplitText，現已全免費）、`@gsap/react`、`lenis` 1.3 | Awwwards 系標配；Lenis 用 `gsap.ticker` 驅動，一個 scroll source |
| 狀態 | `zustand` 5（深度、章節、reduced-motion、GPU tier） | |
| 開發輔助 | `leva`、`r3f-perf`（只在 `?debug` 載入） | |
| 圖片 | `vite-imagetools`（產 AVIF/WebP + 尺寸） | 解決 9.9 MB 問題 |
| 測試 | `vitest` 5 + `@react-three/test-renderer` 9；`@playwright/test` 1.64（視覺回歸）；`size-limit` 14；`@lhci/cli` 0.15 | 見「驗證」 |
| 圖示 | `lucide-react` 或 `@icons-pack/react-simple-icons`（tree-shake） | 移除 Font Awesome CDN |
| 聯絡 | 保留 `@emailjs/browser`，key 改讀 `import.meta.env.VITE_EMAILJS_*`（public key 本來就是公開的，但不要寫死） | |

**不用**：Nuxt/Next（純靜態作品集不需要 SSR）、`@react-three/rapier`（沒有物理需求，省 1 MB wasm）、Vue Bits（換 React 後改參考 React Bits，MIT+Commons Clause，個人站可用）。

### 參考資源（給實作者找靈感與元件，不是全部都要用）
- Paper Shaders：<https://shaders.paper.design>、<https://github.com/paper-design/shaders>（vanilla 與 React 版都有）
- 21st.dev 上 MIT 的 R3F hero：「Living Fluid Hero」、「Prism Hero」（MeshTransmissionMaterial 折射水晶）、「Hero Particles WEBGL」；ThreeUI（threeui.com，MIT，109 個 three.js 元件，shadcn registry）
- React Bits <https://reactbits.dev>（文字動畫：SplitText、ScrollReveal；背景：Particles、Aurora、Liquid）
- Codrops 2026 案例：「The Architecture Behind Trionn」（GSAP+Three+Lenis 一個 tick）、「From Shader Uniforms to Clip-Path Wipes」、「Arnaud Rocca's Portfolio」
- drei：`MeshTransmissionMaterial`、`Caustics`、`Float`、`Sparkles`、`Environment`、`Text3D`、`ScrollControls`（**不用** ScrollControls，改 Lenis+ScrollTrigger 驅動 camera，避免兩套 scroll）
- 參考 shader：three.js examples `webgl_shaders_ocean`、Simondev boids、「Underwater caustics」GLSL 片段（自己寫，約 60 行）

## 資訊架構與敘事（單頁長捲動，深度從 0 m → 4000 m）

路由仍提供 `/`、`/about`、`/project`（對應捲到該段 + 可分享），404 導回 `/`。全站繁中為主、英文副標（姓名/職稱雙語）。

| 深度帶 | 段落 | 3D / 視覺 | 內容來源 |
|---|---|---|---|
| 0 m 海面 | **Hero** | 全螢幕 R3F 場景：俯視水面（Gerstner/噪聲波 shader + 折射 + 太陽光斑），滑鼠造成漣漪；姓名 `林敬勛 / Ching-Hsun Lin` 以 3D 文字或 SplitText 浮出；副標改為使用者確認的現職（**待確認**，見開放問題）；「往下潛」提示 | 舊 Hero |
| 0–200 m 透光帶 | **簡介 + 技能** | 相機下潛，光線變成體積光束（god rays，postprocessing `GodRays` 或自製 radial blur）；技能做成**漂浮的玻璃／氣泡群**（`MeshTransmissionMaterial`，hover 放大、點擊展開：前端 / 後端 / 資料庫 / 雲端與 agent / 生物研究五組）。旁邊游過魚群（GPU boids 或 instanced mesh 跟隨曲線，200–500 隻） | 舊 MainView 技能輪播 + AboutView 五組技能，合併去重 |
| 200–1000 m | **作品集** | 每個作品是一張「深海探測器」卡片：3D 傾斜（pointer tilt）+ 玻璃質感 + 專案截圖；背景粒子像海雪（marine snow）往上飄表示下潛。篩選：`agent 工具 / 自架 / 資料 / 網站 / 生物` | `content/projects.ts`（下節） |
| 1000–4000 m | **經歷時間軸** | 深度計 HUD：左側固定「深度 × 年份」數字隨捲動跳動；時間軸節點是發光生物（bioluminescence，`Sparkles` + bloom） | 舊 AboutView 學經歷，補 2026 |
| 海底 | **關於我 / 興趣 / 聯絡** | 相機落到海床，沉船或熱泉（低多邊形 procedural 幾何即可，不載 glTF 大模型）；興趣標籤；聯絡表單（EmailJS）與 GitHub / Facebook / Resume 連結；「感謝你的造訪」 | 舊 AboutView + App.vue modal |

**作品資料**（`src/content/projects.ts`，每筆：slug、名稱、一句話、標籤、年份、repo、demo、截圖、分類、狀態）：

| 分類 | 專案 | 備註 |
|---|---|---|
| agent 工具 | **scout**（personal-lab 子專案，Go，零依賴科技資訊雷達 CLI） | 連 `Andy880828/personal-lab/tree/main/scout` |
| agent 工具 | **syssec**（Go，agent 當資安管理員用的 CLI，查最新漏洞並判斷伺服器是否受影響） | 2026-11 還在推，確認 README |
| agent 工具 | **andy-claude-kit**（私人 Claude marketplace / skills） | |
| 自架 | **homelab**（Caddy + Uptime Kuma + LiteLLM 的 agent 架設 prompt；網域 chlinandy.com） | 連 personal-lab |
| 資料 | **stock-analysis-app**（Python，投資紀錄與顧問 APP） | |
| 資料 | **local-transcribe**（Python，本地轉錄與聲源辨識，需 GPU） | |
| 資料／生物 | **Trim2Sort**（Python GUI，Sanger + NGS 序列分析，eDNA 資料流） | 保留，更新描述 |
| 網站 | **RecipeCube**（Vue + ASP.NET Core API 食材管理） | 保留；demo 網址 chlinandys.com 已停用 → 改 repo only |
| 網站 | **fitness_challenge**（TypeScript，demo fitness-challenge-seven.vercel.app） | |
| 網站 | **nature-footprint-studio**（Vue，軒齊生態工作室，demo nature-footprint-studio.vercel.app） | 照片有版權，只放連結與自己截圖 |

實作時先用 `gh api repos/Andy880828/<repo>` 確認每個 repo 的 `private` 旗標與 README，**私有 repo 不放 GitHub 連結**，只放說明。

## 專案結構（新）

```
Personal-Page/
├─ index.html                 lang="zh-Hant"，meta description / OG / theme-color
├─ vite.config.ts             react + tailwind + imagetools + glsl；manualChunks: three / r3f / gsap 分開
├─ vercel.json                SPA rewrite 保留；加 Cache-Control for /assets
├─ lighthouserc.json / .size-limit.json / playwright.config.ts / vitest.config.ts
├─ .github/workflows/ci.yml   lint → typecheck → unit → build → size-limit → playwright → lhci
├─ docs/PLAN.md               本計畫複本；docs/DESIGN.md（色票、字體、動效規範）
├─ public/                    favicon.svg、og.png（build 時由截圖產生）、fonts/
└─ src/
   ├─ main.tsx / App.tsx / router.tsx
   ├─ content/                projects.ts、experience.ts、skills.ts、profile.ts（所有文案集中，方便之後只改資料）
   ├─ store/                  useSiteStore（depth、section、reducedMotion、gpuTier、debug）
   ├─ lib/                    lenis.ts（gsap.ticker 綁定）、scroll.ts（ScrollTrigger ↔ depth）、gpu.ts（detect-gpu 或自製 tier）、seed.ts（確定性隨機，測試用）
   ├─ scene/                  R3F：SceneRoot（單一 <Canvas>，全站共用，fixed 背景層）、Water、Caustics、Boids、MarineSnow、SkillBubbles、ProjectProbe、Seafloor、Lights、Effects（postprocessing）、shaders/*.glsl
   ├─ sections/               Hero、Intro、Skills、Projects、Experience、Contact（DOM 層，疊在 Canvas 上）
   ├─ components/             Nav（深度指示 + 章節）、DepthHUD、ProjectCard、TiltCard、SplitHeading、Loader（進度 = drei useProgress）、ReducedMotionToggle
   ├─ hooks/                  useSectionDepth、usePointer、useInView
   └─ styles/                 index.css（Tailwind @theme：深海色票、字體）
```

關鍵設計：**單一 Canvas**（`position: fixed`，`eventSource` 指到 document，`events` 用 drei 的 pointer）；所有 3D 物件依 `depth` 由 ScrollTrigger 更新 zustand，再由 `useFrame` 讀取並 lerp，不讓 React 每 frame 重渲染。DOM 段落放在 Canvas 上層、`pointer-events: none` 除了互動元素。

## 效能與大小預算（size-limit 與 Lighthouse CI 強制）

| 項目 | 預算 |
|---|---|
| 首屏 JS（含 React + router + GSAP + Lenis，不含 three） | ≤ 120 KB gzip |
| three + r3f + drei chunk（lazy，Hero 出現後才載） | ≤ 250 KB gzip |
| postprocessing chunk（lazy，僅 tier ≥ 2 載入） | ≤ 80 KB gzip |
| 全部 JS 總和 | ≤ 500 KB gzip |
| 圖片總和（AVIF/WebP，imagetools 輸出） | ≤ 1.5 MB；單張 ≤ 200 KB；舊 PNG/JPG 原檔移到 `raw/`（gitignore）或直接刪 |
| 字體 | ≤ 2 個 woff2 子集（繁中用 Noto Sans TC 子集化或 system-ui 回退；英文一個 display 字體）≤ 300 KB |
| Lighthouse（desktop）| Performance ≥ 85、Accessibility ≥ 95、Best Practices ≥ 95、SEO ≥ 95 |
| 執行期 | desktop 60 fps；mobile tier 1：關 postprocessing、boids 數量減半、DPR ≤ 1.5；`prefers-reduced-motion`：停 Lenis 慣性、場景改靜態一幀、所有 GSAP 動畫改瞬時 |

GPU 分級（`lib/gpu.ts`）：以 `WEBGL_debug_renderer_info` + `devicePixelRatio` + `hardwareConcurrency` 粗分 0–3；tier 0（無 WebGL）顯示 Paper Shaders 2D 背景 + 靜態 OG 圖，網站內容完整可讀。這也是視覺回歸測試用的「確定性模式」入口。

## 驗證（「可驗證」的具體作法）

1. **靜態**：`eslint`（typescript-eslint + react-hooks + jsx-a11y）、`tsc --noEmit`、`prettier --check`。
2. **單元（vitest）**：`content/*` schema 驗證（zod：每個 project 必有 slug/name/tags，連結為 https）；`lib/scroll.ts` 深度對映（section → depth 區間）；`lib/gpu.ts` 分級；`@react-three/test-renderer` 測 `SceneRoot` 在 tier 0/2 各掛載哪些物件、`Boids` 實例數隨 tier 變化。
3. **視覺回歸（Playwright）**：`?test=1` 進入確定性模式（`seed.ts` 固定亂數、`clock` 固定 t=0、關 Lenis 慣性、關 postprocessing 雜訊、DPR=1）。對 Hero / Skills / Projects / Experience / Contact 五個捲動位置在 1440×900 與 390×844 各截一張，`toHaveScreenshot({ maxDiffPixelRatio: 0.02 })`，baseline 進 repo。Chromium 用 `--use-gl=angle --use-angle=swiftshader` 確保 CI 有 WebGL。
4. **互動煙霧測試（Playwright）**：導航到 `/about`、`/project` 會捲到正確段落；技能氣泡點擊展開；作品篩選；聯絡表單驗證（不真的送 EmailJS，攔截請求）；鍵盤可達所有互動元素；`prefers-reduced-motion: reduce` 時 Canvas 不持續 render（`requestAnimationFrame` 計數 < 5/秒）。
5. **大小**：`size-limit` 檢查上表 chunk；`vite-bundle-visualizer` 產報告進 CI artifact。
6. **Lighthouse CI**：`lhci autorun` 對 `vite preview`，assert 上表分數；mobile 另跑一次只記錄不 assert（第一版）。
7. **CI**：`.github/workflows/ci.yml`，PR 與 main 都跑；Vercel preview 連結貼到 PR。

## 實作階段（給 Opus 派工；每階段結尾要能 `npm run build && npm run test` 綠燈並 commit）

### Phase 0 — 盤點與資產（0.5 天）
- 新分支 `rebuild/deep-sea`。把舊 `src/` 改名 `legacy/`（最後一階段刪除），抽出可重用素材：`M105090005.jpg`（大頭照）、`個人.png`、作品截圖、`diver-1.png`、三個魚/珊瑚 SVG（可轉成 3D 場景的 sprite 或剪影）。
- 用 `gh api` 抓 10 個作品 repo 的描述、語言、private 旗標、homepage，寫進 `src/content/projects.ts` 草稿；對沒有截圖的（scout、homelab、syssec、andy-claude-kit、stock-analysis-app、local-transcribe）用 Playwright 對 README 或 demo 截圖，或做 1200×750 的程式碼風格封面（深海色票）。
- 寫 `docs/DESIGN.md`：色票（海面 `#dff3ff` → 透光帶 `#0b4f7a` → 深海 `#02101f` → 海底 `#010409`；強調色生物光 `#4ff0d2` / `#ffb86b`）、字體、動效時長曲線、深度對映表。

### Phase 1 — 骨架（1 天）
- `npm create vite@latest -- --template react-ts` 重建，裝上表依賴，Tailwind 4、路由、Lenis+GSAP ticker、zustand store、`content/` 全部文案、六個 DOM 段落（無 3D，先用 Paper Shaders `MeshGradient` 當背景）。
- ESLint/TS/Prettier/Vitest/size-limit/CI 全部就位；此時就要有第一個 Playwright 截圖 baseline。
- 驗收：純 DOM 版已經是可讀、可導航、RWD、a11y 通過的完整網站（這就是 tier 0 回退版）。

### Phase 2 — 核心 3D 場景（2–3 天）
- `SceneRoot`：fixed Canvas、`dpr=[1, tier>=2 ? 2 : 1.5]`、`frameloop="demand"` 在 reduced-motion、`Preload`、`AdaptiveDpr`。
- `Water`：俯視水面 shader（頂點位移 + 法線 + 菲涅爾 + 太陽高光），pointer 漣漪以 uniform 陣列（≤ 8 個）傳入。
- 相機路徑：`depth` 0 → 1 對應 camera y 從 +5 到 −60，pitch 從俯視到平視；`Lights` 隨深度變暗、霧 `fogExp2` 顏色跟深度帶。
- `Caustics` 投影（drei `Caustics` 或自製 projected texture）與 god rays（tier ≥ 2）。
- `MarineSnow`：instanced points 2000 顆（tier 1：800），用 shader 隨 depth 位移。
- 驗收：fps 量測寫成 Playwright 測試（`?test=1&perf=1` 回報 60 幀平均 frame time < 20 ms on CI swiftshader 只做 smoke）。

### Phase 3 — 互動元素（2 天）
- `Boids`：GPU（drei `useFBO` + ping-pong）或 CPU 簡化版 instanced（300 隻，tier 1：120），跟隨滑鼠微避讓；魚形用低多邊形程序幾何或舊 SVG 擠出。
- `SkillBubbles`：五顆 `MeshTransmissionMaterial` 球（tier<2 用 `MeshPhysicalMaterial` 替代），`Float` 漂浮，hover 放大，點擊時 DOM 面板展開對應技能文字。
- `ProjectCard` 3D tilt + 玻璃；作品段落背景海雪加速。
- `DepthHUD` 與 `Experience` 發光節點（`Sparkles` + Bloom，tier ≥ 2）。
- `Seafloor` 程序幾何 + 熱泉粒子；聯絡表單。
- Hero 文字：SplitText 逐字浮出，載入完成（`useProgress` 100%）才播。

### Phase 4 — 收斂與驗證（1 天）
- 所有預算跑綠：size-limit、lhci、Playwright baseline 更新並人工檢視。
- 行動版實機（iOS Safari、Android Chrome）檢查：觸控漣漪、無橫向捲動、Canvas 不吃觸控捲動。
- SEO/OG：`index.html` meta、`og.png` 由 Playwright 對 Hero 截圖產生、`robots.txt`、`sitemap.xml`（三條 route）。
- 刪 `legacy/`、`output.css`、Font Awesome CDN、`dist/`（已在 gitignore 但 repo 裡有，`git rm -r --cached dist`）。
- README 重寫（繁中：技術棧、`npm run` 指令、預算表、如何更新 `content/`）。

### Phase 5 — impeccable 視覺審視（0.5 天）
- 安裝：`/plugin marketplace add pbakaus/impeccable` → `/plugin` 安裝；或 `npx skills add pbakaus/impeccable`。
- 依序跑：`/impeccable critique`（UX 整體）→ `/impeccable audit`（27 條 anti-pattern 檢查：漸層濫用、Inter 萬用、卡片套卡片、間距擠壓）→ 針對報告跑 `typeset`、`colorize`、`animate`、`polish`、`adapt`（RWD）、`harden`（邊界狀態：長文、無圖、慢網）。
- 每輪 impeccable 修改後重跑 Playwright 截圖並更新 baseline，commit message 註明是哪個 impeccable 指令的結果。
- 最後產出 `docs/REVIEW.md`：critique 原文、採納／不採納的理由。

### 交付
- PR `rebuild/deep-sea` → `main`，描述附：Vercel preview、Lighthouse 分數截圖、size-limit 表、五張視覺 baseline。
- 合併後確認 Vercel production 更新，`personal-page-gules.vercel.app` 可用；若要綁 `chlinandy.com` 子網域（例如 `me.chlinandy.com`）是另一個任務。

## 開放問題（實作前請 Andy 回答，否則依預設）

1. **現職／副標**：台郡的經歷到 2026/03，現在 Hero 副標要寫什麼？預設：「Full-stack & Data Engineer / Oceanographer」，經歷時間軸最後一筆開放結尾。
2. **履歷連結** `rxresu.me/andy880828/literary-polite-sparrow` 還有效？預設保留。
3. **Facebook 連結**要不要留？預設留（舊站有）。
4. **私有 repo**（andy-claude-kit、stock-analysis-app、local-transcribe、syssec 若為私有）：只放說明不放連結，可接受？
5. 是否要**英文版切換**？預設不做，雙語只在標題。
6. 大頭照與「個人.png」合照要不要換新？預設沿用並壓縮。

## 風險與對策

- **Paper Shaders 0.0.x 會 break** → `package.json` 鎖確切版本，不用 `^`。
- **CI 沒 GPU 導致截圖不穩** → swiftshader + 確定性模式 + 2% 容差；真的不穩就把 3D 截圖降為「Canvas 有畫出非純色像素」的 smoke，DOM 層另截。
- **three 0.187 與 drei 10.7 版本耦合** → 一次裝齊、不混版；`npm ls three` 必須只有一份。
- **行動版耗電/發熱** → tier 1 預設關 postprocessing、boids 減半、`AdaptiveDpr`、頁面不可見時 `frameloop` 停。
- **React 19 + R3F 9 的 StrictMode 雙掛載** → shader uniforms 用 `useMemo` 穩定引用，資源在 `useEffect` cleanup 釋放。
- **Tailwind 4 無 config** → 色票與字體全在 `styles/index.css` 的 `@theme`，impeccable 的 `colorize/typeset` 也改這裡。

---

# 第五輪計畫：海洋 × 代碼深度融合、分層生物、內容轉向

> 給 Opus 派工用。2026-11-02 依 Andy 第三次回饋寫成；**只是計畫，尚未動工**。
> 實作在 `Andy880828/Personal-Page` 新分支 `round5/ocean-code`，完成後開 PR → `main`，再回這裡填進度表。
> 每個 Phase 結尾都要 `npm run lint && npm run typecheck && npm run test && npm run build && npm run size` 綠燈並 commit；視覺基準一樣用 commit 訊息 `[update-snapshots]` 由 CI 產生。

## 0. Andy 的回饋 → 對應作法（總表）

| # | 回饋 | 作法 | Phase |
|---|---|---|---|
| 1 | 頂端列變色前最下方有一條黑線 | 海面狀態不再用 `border-b` + 短漸層當背景，改成一片**比列高、以 mask 淡出**的遮罩；加一個抓像素的 Playwright 測試 | B |
| 2 | CTD 概念好但太刻意、field note 太直白；要以意象為主，把 coding 跟海洋雜揉 | 刪掉五段 field note 與所有解釋性標籤；科學只留在**畫面行為與數字**裡；代碼感改由「字元海雪、等深線掃描、魚群即資料、沉積層即 git log、真實 DNA 資料」這些意象承擔（§3） | A、E |
| 3 | 魚群不錯；要加隆頭魚與珊瑚；每一層都要有代表生物；火山可以，多面體看不出是什麼 | 建一張**依深度帶的生物表** `fauna.ts`（§4），用 CC0／CC-BY 現成模型；`ProjectProbe` 整個移除 | C、D、B |
| 4 | 主要方向改為資料探勘與數據可視化、工作流程數位化、Agent 應用 | 技能五組、作品分類、個人簡介、經歷最後一筆全部重寫（§2） | A |
| 5 | DNA 太假，找現成的模 | 用 RCSB 真實 B-DNA 結構 **1BNA** 的原子座標產生模型（資料驅動，和 eDNA 背景呼應）；備案 Sketchfab CC-BY 模型 | C、E |
| 6 | 作品區的光球拿掉 | `ProjectProbe.tsx`（線框二十面體 + 軌道環 + 聲納球）整個刪除 | B |
| 7 | 拿掉「用 React Three Fiber 打造」 | `Contact.tsx` 頁尾只留 `© 林敬勛 · 照片版權所有` | A |
| 8 | AI 相關說明太 AI 味，都拿掉 | 寫一條**文案 lint 單元測試**禁用字表（§2.4），所有內容檔過關才算完成 | A |
| 9 | 強調深入融合海洋與代碼 | §3 的六個核心意象，每個都有「海洋原型 → 代碼原型 → 畫面行為」三欄，並且都可以被測試 | E |
| 10 | 變成計畫給 Opus | 本節 | — |

## 1. 廣搜結果：別人怎麼把 code 跟海洋放在一起

找得到的案例大多停在「深度計 + 終端機」這一層，沒有人真的把海洋元素做成**會像資料一樣行為**的東西，這就是我們要做的差異。

| 來源 | 做了什麼 | 可借用 |
|---|---|---|
| [DeepRealm](https://github.com/DChitale/DeepRealm) | 捲動對應深度計、潛艇式 HUD、three.js 海洋、過分層門檻時背景切換 | 分層門檻當「事件」來用（§3.6 的掃描帶就是從這裡延伸） |
| [DeepSea Portfolio](https://devpost.com/software/deepsea-portfolio) | 非 WebGL；分層捲動、可輸入指令的終端機、跟游標的魚群 | 我們已有指令列；魚群跟游標改成「魚群即資料」 |
| jessuiii/Portfolio（第二輪就找到） | 深海下潛 + 深度計 + 即時終端 | 同上，證明這個方向可行但大家都停在這 |
| [Codrops 2026-09 Shape-aware ASCII renderer in three.js](https://tympanus.net/codrops/2026/09/04/beyond-the-luminance-ramp-a-shape-aware-ascii-renderer-in-three-js/) | 不只用亮度選字元，還用形狀（邊緣方向）選字元的 ASCII 後製 | §3.6 分層掃描帶、§3.1 字元海雪的字元圖集作法 |
| [Medium：ASCII animated hero in three.js](https://medium.com/@egorshesternin/how-to-add-a-cool-ascii-animated-hero-section-to-your-website-step-by-step-guide-d4070e45e2c8) | 亮度→字元的後製 pass，字元集、反相、單色選項 | 最簡單的實作參考（約 60 行 GLSL） |
| [three.js `webgl_shaders_ocean`](https://github.com/mrdoob/three.js/blob/master/examples/webgl_shaders_ocean.html) | 官方水面 + Sky | 已用過，不再引入 |
| [Manya et al. 2013，Procedural contour lines](https://otik.uk.zcu.cz/handle/11025/10597?locale=en) | 用高度取模在 shader 裡畫等高線、依坡度補償線寬、淡出 | §3.2 海床等深線 |
| [Esri：WebGL 洋流 streamline 動畫](https://developers.arcgis.com/javascript/latest/sample-code/layers-imagery-afrenderer-blending/) | 洋流當粒子流線畫 | §3.3 魚群分群時的流線感（參考，不引入套件） |
| [Codrops 2026-03 Corentin Bernadou's portfolio](https://tympanus.net/codrops/2026/03/05/inside-corentin-bernadous-portfolio-swiss-inspired-layouts-webgl-geometry-and-thoughtful-motion/) | three + GLSL + GSAP + Lenis 的節制用法 | 動效節奏參考 |

結論：**代碼感不靠標籤和說明文字，靠海洋物件的行為像資料**。下面每個意象都要能一句話說出「這個海洋東西在做什麼程式的事」。

## 2. 內容轉向（Phase A，純文字，先做、先驗）

### 2.1 技能五組（`src/content/skills.ts`，保持五組 → 對應雙螺旋五段）

| id | 名稱 | 一句話（草稿，Opus 可潤但不得出現 §2.4 禁用字） | strengths（草稿） | tools |
|---|---|---|---|---|
| `mining` | 資料探勘與分析 / Data Mining | 從雜訊裡找訊號：一條溪的 eDNA 定序、一家工廠的產線紀錄、我自己的投資紀錄，做法是同一套。 | 資料清理與特徵整理、統計檢定與時間序列、定序資料流（eDNA、BLAST）、把結論寫成可以被重算的筆記 | Python、Pandas、R、SQL、BLAST+、Cutadapt、USEARCH |
| `viz` | 數據可視化 / Data Visualization | 讓圖表先回答問題，再談好看；這個網站本身就是一張把深度換算成壓力、水溫、光的圖。 | 圖表選型與尺度（線性、對數）、儀表板設計、互動式與 3D 視覺化、配色與可讀性 | Python 繪圖（Matplotlib）、Three.js、Tailwind CSS（只列站上已有的工具，不新增） |
| `workflow` | 工作流程數位化 / Workflow Digitization | 把靠人記、靠 Excel 傳的流程做成系統：表單、排程、報表、通知，而且出錯看得到。 | 需求盤點與資料模型、表單與審核流程、排程與背景工作、自架服務與監控 | ASP.NET Core、FastAPI、Supabase、SQL Server、PostgreSQL、Caddy、Linux、Vercel |
| `agent` | Agent 應用 / Agent Applications | 讓程式替人跑完整段流程（查漏洞、整理資訊、轉逐字稿），每一步的輸出都能被檢查。 | 可被程式呼叫的 CLI（JSON 輸出、穩定錯誤碼）、流程拆解與權限邊界、本地執行與資料不外流、結果驗證 | Go、Python、Claude Code、LiteLLM、Linux |
| `web` | 網站全端 / Web | 從切版到 API 到資料庫，一個人把網站做到能上線、能維護。 | 響應式介面、元件化與狀態管理、RESTful API、關聯式資料庫設計、前端測試 | Vue 3、Nuxt 4、React、TypeScript、Tailwind CSS、Vite、C#、Express.js、Figma |

原本的「網頁前端／後端／資料庫」併成 `web`；「資料與生物研究」併進 `mining`（eDNA 就是資料探勘）。五組顏色沿用現有五色（`#4ff0d2 #7cc8ff #a99bff #ffb86b #ff7eb6`），順序改為 mining → viz → workflow → agent → web（主方向在前）。

### 2.2 作品分類與歸屬（`src/content/types.ts`、`projects.ts`、`Projects.tsx` 的 FILTERS、`ProjectCard.tsx` 的 ACCENT、`lib/console.ts`、`tests/unit/content.test.ts` 一起改）

`ProjectCategory = 'mining' | 'viz' | 'workflow' | 'agent' | 'web'`；`categoryLabels`：資料探勘、數據可視化、流程數位化、Agent 應用、網站。

| 作品 | 分類 | summary 改寫方向（去 AI 味） |
|---|---|---|
| syssec | agent、workflow | 「主機安全稽核與漏洞情資 CLI：單一執行檔，查最新漏洞並比對這台機器裝了什麼，輸出可以被程式讀的報告。」 |
| scout | agent、mining | 「個人科技資訊雷達：抓自選來源、去重、用公開規則排序，摘要再交給人或程式接手。」（不寫「交給 Claude 寫」） |
| local-transcribe | workflow、agent | 保留現在的描述，已經夠具體 |
| homelab | workflow | 「家中小電腦的 Caddy、Uptime Kuma、LiteLLM：一份從盤點到驗證的架設手冊，照著做就能重建。」 |
| andy-claude-kit | agent | 「一行指令在任何電腦重建我的開發環境：工程規則、語言 rules、唯讀審查流程。」 |
| stock-analysis-app | mining、viz | 「投資紀錄與回顧 App：每一筆進出都留紀錄，再用圖表回頭看決策對不對。」（highlight 維持兩條） |
| Trim2Sort | mining、workflow | 保留 |
| fitness_challenge | viz、web | 保留（InBody 數據與難度係數就是可視化與公平計分） |
| nature-footprint-studio | web | 保留 |
| RecipeCube | web、workflow | 保留 |

作品段 `lead` 改為：「現在主要做三件事：從資料裡找答案、把流程做成系統、讓程式替人跑完一段工作。網站與 eDNA 資料流是更早的作品，也留在這裡。」

### 2.3 個人簡介與經歷

- `profile.ts` `intro[1]`：「後來轉進軟體：從 Vue、ASP.NET 的全端網站，到台郡科技的數據分析；現在專注在資料探勘與數據可視化、工作流程數位化，以及把 agent 放進實際流程裡。」`role` 改 `Data & Workflow Engineer`（Andy 已確認），`roleZh` 改「資料與流程工程師，海洋科學出身」。
- `experience.ts` 最後一筆 `title`：「資料探勘、流程數位化與 Agent 應用」，`note`：「stock-analysis-app、scout、syssec、homelab」。
- `README.md`、`index.html` meta description、`docs/DESIGN.md` 同步改措辭。

### 2.4 文案 lint（可驗證的「去 AI 味」）

新增 `tests/unit/copy.test.ts`：讀 `src/content/*.ts`、`src/sections/*.tsx`、`index.html`、`README.md` 的字串，禁止出現：`AI`、`LLM`、`大語言模型`、`智慧`、`賦能`、`打造`、`探索`、`解鎖`、`無縫`、`強大`、`革命`、`旅程`、`沉浸`、`field note`、`CTD`（UI 字串；程式註解與 `science.ts` 的來源標題不算）。另外禁止 emoji。失敗訊息要印出檔名與那一行。
寫法原則：動詞具體（抓、去重、比對、輸出）、說給誰用、說怎麼驗證；不形容感受。

### 2.5 頁尾
`Contact.tsx` 頁尾改成 `© {profile.name} · 照片版權所有`。

## 3. 海洋 × 代碼：六個核心意象（Phase E）

每一個都要有「海洋原型 → 代碼原型 → 畫面行為」，而且不加任何解釋文字；使用者看得懂就看，看不懂也只是覺得好看。

| # | 海洋原型 | 代碼原型 | 畫面行為 | 落點 |
|---|---|---|---|---|
| 3.1 | 海雪（marine snow） | log stream／token 流 | 200 m 以下，海雪粒子逐漸從圓點變成**等寬字元**（`0 1 . : ; = + { } < > / \ \| - _ *`），往下飄像 `tail -f`；越深字元比例越高，1000 m 以下全部是字。每顆粒子出生時固定一個字元、固定大小，**不逐幀換字**（避免閃爍）。 | `MarineSnow.tsx`：點精靈改用 canvas 產生的 16×16 字元圖集 + 每實例 uv offset 屬性；tier 1 粒子減半 |
| 3.2 | 海床地形、等深線海圖 | 資料的等高線／heatmap 掃描 | 海床在游標探照燈範圍內顯示**等深線**（`fract(height/spacing)` + `fwidth` 抗鋸齒，每五條一條粗線），燈移開就回到岩石；全域另留 8% 的淡線。就像拿掃描器掃過資料。 | `Seafloor.tsx` 地形 shader 加 uniform `uCursor`、`uRadius`；參考 Manya 2013 的坡度補償 |
| 3.3 | 魚群 | 資料集與查詢結果 | 中層帶的燈籠魚群，每條魚帶一個作品分類（五色其一，顏色只在發光點上，很淡）。按作品篩選時，**符合分類的魚聚成一團、其他散開**，群體重新成形像一張散點圖被 filter；篩選回「全部」就散回去。點擊驚散保留。 | `Fish.tsx`/`boids.ts`：每隻魚加 `tag`；boids 多一個依 `projectFilter` 的目標點與排斥項；用 instance color |
| 3.4 | 沉積岩芯、地層 | `git log --graph` | 經歷段改成**由新到舊往下**（跟 `git log` 一樣新的在上），背景是一條垂直的地層帶，每筆經歷是一層不同色帶，左側一條細線 + 節點像 commit graph，節點旁一個 7 字元的短碼（由起訖年月 hash 出來，像 commit sha）。不寫「這是 git」。 | `Experience.tsx` 排序反轉、`YEAR` 儀表方向反轉；`experience.ts` 註解改「由新到舊」；新元件 `Strata.tsx`（純 DOM/SVG，不進 3D） |
| 3.5 | DNA | 真實資料驅動的模型、序列字串 | 技能雙螺旋改用 **RCSB 1BNA**（B-DNA 十二聚體，Drew & Dickerson 1981）的真實原子座標：建置時從 `.pdb` 取主鏈磷原子與鹼基中心，沿軸堆三份得到 36 bp，五個技能段分配到這條序列上。技能段左側的 ACGT 假序列改成 1BNA 的真序列 `CGCGAATTCGCG` 重複三次，當前技能段的那段字母亮起（跟 3D 一致）。 | `scripts/build-dna.mjs`（Node，讀 `raw/1BNA.pdb` → 寫 `src/assets/dna-1bna.json`，只存座標與原子種類，<20 KB）；`SkillHelix.tsx` 改讀 json 排 instanced 球與棍；`Skills.tsx` 的 `pseudoSequence` 換真序列 |
| 3.6 | 聲納、分層門檻 | ASCII 渲染、掃描 | 捲過 200 m 與 1000 m 門檻時，一條約 18% 高的水平帶由上往下掃過畫面（600 ms），**帶內**的畫面以字元渲染（亮度→字元，字元格 8 px），帶外原樣；像聲納回波把世界讀成字再還原。另加指令 `ascii on|off` 可以整頁維持字元渲染（彩蛋，列在 help）。tier ≥ 2 且非減少動態才有。 | `Effects.tsx` 新增自製 `AsciiBand` Effect（postprocessing `Effect` 子類，fragment 約 50 行，字元圖集與 3.1 共用）；`lib/console.ts` 加 `ascii` |

輔助意象（小，順手做）：
- 指令列提示字元改成 `~/ocean/<section> $`，深度就是「目前目錄」；`ls` 無參數列出目前段落可 `cat` 的東西。
- 首頁座標行保留；`SectionHeader` 右端的深度刻度保留（這兩個是好的、不解釋的意象）。
- 遙測儀表（`DepthHUD`）保留深度計、溫光剖面曲線與三個讀數，但**拿掉所有文字標籤與說明**，只留數值與單位（`412 atm`、`4.1 °C`、`1e-9`）；剖面圖不畫軸名。
- `science.ts` 的 NOTES 與 SOURCES 不刪，改由指令列 `cat zone` 顯示（想看的人才看）；頁尾加一個小字連結「資料來源」到 README 的科學來源段。

**閃爍守門**：3.1、3.6 是這一輪最可能重蹈第二輪閃爍的地方。規則：粒子的字元與大小在出生時決定、掃描帶只在門檻事件時出現、`ascii on` 的字元選擇只看亮度量化（8 階）且 uv 取樣用 cell 中心。完成後跑第二輪留下的幀差腳本（`?slow=0.02` 連續截圖，相鄰幀差 <1），結果寫進 PR。

## 4. 分層代表生物（Phase C、D）

### 4.1 生物表 `src/content/fauna.ts`（單一事實來源，單元測試驗證）

| 段落（深度） | 分層 | 代表生物 | 來源與授權 | 面數（原→目標） | 作法 |
|---|---|---|---|---|---|
| Hero 0–10 m | 海面 | 無；水面下隱約的魚影 | — | — | 現有魚群在水面下 2–6 m 游過，只看得到影子 |
| About 10–100 m | 透光帶、礁區 | **珊瑚礁**：鹿角珊瑚、扇珊瑚、海葵；**隆頭魚** 2–3 隻個體 + 現有小型魚群 | 珊瑚：[Smithsonian 3D corals](https://3d.si.edu/corals)（CC0，glTF；需逐個模型頁確認）、備案 [Fan Coral Cluster Low](https://sketchfab.com/3d-models/none-ee315b6063f8472ea4788a75098a2cb5)（CC-BY，2.3k）、[Low-Poly Anemone](https://sketchfab.com/3d-models/none-ef8834fa2ed84432bf87b33c08e0ab6d)（CC-BY，1.1k）、[Acropora cervicornis](https://sketchfab.com/3d-models/none-87457b4ee3bd4828abfaba66b71790bd)（CC-BY，32k）。隆頭魚：ffishAsia-and-flyingfish 的 CC0 攝影測量 [Snooty Wrasse](https://sketchfab.com/3d-models/none-a1fa2be4f363454baadf35fafa403e6b)（1.19M）、[Red Naped Wrasse](https://sketchfab.com/3d-models/none-30a07294d36c4539b23f398e522dd910)（590k）、[Bambooleaf Wrasse](https://sketchfab.com/3d-models/none-ce4d7f7fc1d84015bf3c7de637963c67)（782k） | 珊瑚各 ≤ 6k；隆頭魚 → 8k | 隆頭魚是靜態掃描，用 `onBeforeCompile` 在頂點 shader 沿身體做正弦擺動；礁放在畫面左下的岩坡上，跟 About 的「高雄、西子灣」呼應 |
| Skills 100–200 m | 透光帶下緣 | 小型魚群變稀、扇珊瑚（深水柳珊瑚）、第一隻水母 | 同上 | — | 魚群密度隨深度線性降到 0；DNA 在右側 |
| Projects 200–1000 m | 中層帶 | **燈籠魚群**（3.3 的資料魚群）、**管水母**、現有水母 | 燈籠魚：[Quaternius Lowpoly Animated Fish](https://quaternius.itch.io/lowpoly-animated-fish)（CC0，FBX/OBJ/Blend，有游泳動畫）任選一隻小魚染成深色 + 發光點；備案 [Khronos BarramundiFish.glb](https://github.com/KhronosGroup/glTF-Sample-Assets/blob/main/Models/BarramundiFish/README.md)（CC0，可直接抓）。管水母：[Marrus/Deep-Sea Jellyfish (Siphonophora)](https://sketchfab.com/3d-models/none-4a9d9b8553b14ccda32033cbb6c2c9b7)（CC-BY，36k）或 [Nanomia bijuga, MBARI](https://sketchfab.com/3d-models/none-bf0b8d2378914a66a2e1f94f63d63543)（CC-BY，124k） | 魚 ≤ 2k；管水母 → 8k | `ProjectProbe` 移除後，這一段的焦點就是魚群分群 |
| Experience 1000–3800 m | 深層帶 | **鮟鱇魚** 1 隻（游標靠近時誘餌亮）、**大鰭後肛魚（Magnapinna）** 1 隻遠處、字元海雪最密 | 鮟鱇：[Weird deepsea anglerfish](https://sketchfab.com/3d-models/none-3085eefe0bb348d89902cd51e7e91f05)（CC-BY，2k）或 [Angler Fish, Inktober](https://sketchfab.com/3d-models/none-f7192d3520c348b2a7bc31f8e178abd7)（CC-BY，564）；Magnapinna：[Dotflare](https://sketchfab.com/3d-models/none-be2b90cc7e2f436bb7a689a45aa256c1)（CC-BY，8k） | 不用減 | 鮟鱇的誘餌是場景裡唯一的暖色點光，取代被拿掉的光球 |
| Contact 3800–4000 m | 海床、熱泉 | 現有程序式黑煙囪保留；加**管蟲叢**（Riftia）、**大王具足蟲** 1 隻、海百合 | 管蟲：Sketchfab 這次沒搜到可商用的，用程序式（圓柱束 + 紅色端點）；具足蟲：[ffishAsia CC0 Giant Isopod](https://sketchfab.com/3d-models/none-3979c291d1f9454c90851efe291eab60)（582k → 6k）；黑煙囪若想換現成：[Black Smoker](https://sketchfab.com/3d-models/none-7210f54bb30943ea80347472ce00b64f)（CC-BY，119k） | 見左 | 等深線掃描（3.2）在這裡最明顯 |

`tests/unit/fauna.test.ts`：每個段落至少一種生物；每種生物的 `depthRange` 必須落在它所屬分層（`science.ts` 的 ZONES）內；每個模型檔在 `public/models/` 存在且 ≤ 200 KB；`public/models/LICENSES.md` 列到每一個 CC-BY 的作者與連結。

### 4.2 素材管線 `scripts/models/`

1. **取得**：能直接抓的（RCSB `https://files.rcsb.org/download/1BNA.pdb`、Khronos GitHub、Smithsonian 下載頁、Quaternius itch.io）由 Opus 用 curl 放到 `raw/models/`（gitignore）；**Sketchfab 需要登入下載**，列在 §6 請 Andy 先抓好放進 `raw/models/`，拿不到的一律用程序式備案，不等。
2. **轉檔**：FBX → glTF 用 `fbx2gltf`（npm `fbx2gltf`，Linux 可跑）；OBJ 用 `obj2gltf`。
3. **減面與壓縮**：`@gltf-transform/cli`：`dedup → weld → simplify --ratio <r> --error 0.001 → resize --width 1024 → webp --quality 80 → meshopt`。攝影測量模型目標 ≤ 8k 面、單張貼圖 1024 WebP；低多邊形模型不減面只壓縮。
4. **載入**：three 的 `GLTFLoader` + `MeshoptDecoder`（`three/examples/jsm/libs/meshopt_decoder.module.js`），包成 `hooks/useModel.ts`，依段落鄰近才載（前後各一段），載完快取。不用 Draco（wasm 太大）、不用 KTX2（transcoder 太大）。
5. **預算**：`.size-limit.json` 加兩條：「模型（meshopt GLB）」合計 ≤ 800 KB、「模型解碼器」≤ 40 KB gzip；每個 GLB ≤ 200 KB。全站首次潛到底的總下載量目標 ≤ 3 MB。
6. **確定性**：`?test=1` 下模型的動畫時間固定、擺動停在 t=0，視覺基準才穩。

## 5. 頂端列、儀表、移除探測器（Phase B）

- **頂端列**（`Nav.tsx`）：先在正式站用 Playwright 放大截圖確認那條線的成因（最可能是 `border-b` 的 1 px 加上 `from-deep/80 via-deep/40 to-transparent` 在 56 px 內收尾太快，交界處形成暗帶）。改法：海面狀態 `header` 本身無背景無邊框；用 `::before` 畫一片高 160 px、`mask-image: linear-gradient(black, transparent)` 的 `bg-deep/70` 遮罩，不隨 header 高度收尾；選單膠囊保留 `bg-deep/75`。非海面狀態維持 `bg-deep/85`，邊線改 `border-foam/8` → 拿掉，用 1 px 的 `shadow` 取代或乾脆不要。
  驗證：`tests/e2e/nav.spec.ts` 在 hero 狀態截圖，取 header 底緣下方 0–6 px 每列的平均亮度，相鄰列差異 < 12/255（不能有突然變暗的一列）。
- **遙測儀表**（`DepthHUD.tsx`）：拿掉文字標籤與剖面圖軸名，只留深度計、曲線、三個「數值 + 單位」；寬度縮到 9 rem，`main` 的右側保留量跟著縮。
- **移除 `ProjectProbe.tsx`**：從 `SceneRoot.tsx` 拿掉；`useSiteStore` 的 `pulse/firePulse` 若只有它在用就一起刪；相關單元測試更新。聲納 ping 改為 3.6 的掃描帶負責「回波」意象。
- **刪 `ScienceNote.tsx`** 與五段的引用；`data-testid="science-note"` 的 e2e 斷言移除。

## 6. Andy 已回答（2026-11-02）與要先做的事

| # | 事項 | 決定 |
|---|---|---|
| 1 | Sketchfab 模型下載 | Andy 會下載。**格式選 Sketchfab 下載選單裡的 glTF**（自動轉檔的 zip，內含 `.gltf` + `.bin` + 貼圖），不要 USDZ；原始格式（OBJ／FBX／Blend）只有在 glTF 選項不存在時才拿。每個模型解壓到 Personal-Page 的 `raw/models/<slug>/`（gitignore），同資料夾放一個 `source.txt` 寫模型網址、作者、授權。攝影測量模型的 zip 可能上百 MB，沒關係，減面後才進 `public/models/`。 |
| 2 | 新技能組的工具清單 | 不補，只列站上已經有的工具；§2.1 標「待 Andy 補」的項目全部拿掉 |
| 3 | 職稱 | 改成 `Data & Workflow Engineer`，`roleZh`「資料與流程工程師，海洋科學出身」 |
| 4 | stock-analysis-app | 維持現在兩條 highlight，只改 summary 措辭 |
| 5 | 經歷由新到舊 | 改 |
| 6 | 本站列為作品 | 不列（預設） |

## 7. 驗證與審視（Phase F）

1. 既有：lint、typecheck、Vitest（新增 `copy.test.ts`、`fauna.test.ts`、helix 讀 1BNA json 的測試、boids 分群測試：給定 filter，符合 tag 的魚與目標點平均距離要小於其他魚）、Playwright（新增 nav 像素測試、`ascii on` 後 canvas 非純色、篩選後魚群分群的 smoke）、size-limit（新增模型兩條）、Lighthouse 桌機 ≥ 99/100/100/100 維持。
2. **閃爍量測**：第二輪的幀差腳本在 `?slow=0.02` 跑海面、200 m、1000 m、海床四個位置，相鄰幀差 < 1；結果表貼 PR。
3. **視覺基準**：commit 加 `[update-snapshots]` 由 CI 產生，人工看 artifact。
4. **impeccable 第三輪（正式站）**：合併上線後對 `personal-page-gules.vercel.app` 跑 critique + 偵測器（兩個獨立 subagent，同第二輪作法），P1 全修、P2 記錄在 `docs/REVIEW.md`；特別檢查：字元海雪與正文的對比、掃描帶在 reduced-motion 關閉、手機無橫向溢出、HUD 不蓋內容。
5. PR 描述附：改寫前後文案對照表、生物表、模型大小表、閃爍量測表、Lighthouse 分數。
6. 合併後回本檔把 §0 的進度表 `[ ]` 改 `[x]`，填 PR 連結。

## 8. 風險

| 風險 | 對策 |
|---|---|
| 攝影測量模型減面後貼圖接縫、法線破 | `simplify` 時 `--lock-border`；不行就改用低多邊形備案，不硬撐 |
| 模型拖慢首屏或 tier 1 | 全部 lazy、tier 1 只載每層一種、tier 0 完全不載 |
| 字元海雪／ASCII 帶再度閃爍 | §3 的閃爍守門三條規則 + 幀差量測是合併門檻 |
| 魚群分群讓 boids 失真（擠成一球） | 目標吸引力上限 + 分離項加權；分群只在 Projects 段 weight > 0.5 時啟用 |
| 內容改寫走味 | §2.4 lint 只擋禁用字，語氣由 Opus 依 §2 草稿寫，PR 對照表給 Andy 看 |
| Sketchfab 素材拿不到 | 每一種都有程序式或可直接抓的備案（§4.1），不阻塞 |
