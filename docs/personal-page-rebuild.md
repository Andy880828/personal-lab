# Personal-Page 重構追蹤

> 狀態追蹤檔。計畫在下方；每個階段完成後把 `[ ]` 改 `[x]` 並填上 commit / PR 連結。
> 實作在 `Andy880828/Personal-Page` 的分支 `rebuild/deep-sea` 進行，這裡只追蹤。

## 進度

| 階段 | 狀態 | 紀錄 |
|---|---|---|
| 計畫（本檔） | [x] 2026-11-02 | 使用者已確認四項決策 |
| 開放問題回答 | [ ] | 見文末六題 |
| Phase 0 盤點與資產 | [ ] | |
| Phase 1 骨架（DOM 版 + CI） | [ ] | |
| Phase 2 核心 3D 場景 | [ ] | |
| Phase 3 互動元素 | [ ] | |
| Phase 4 收斂與驗證 | [ ] | |
| Phase 5 impeccable 視覺審視 | [ ] | |
| PR 合併、Vercel production 確認 | [ ] | |

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
