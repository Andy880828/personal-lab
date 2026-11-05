# Personal-Page 第六輪：聯絡、作品證據、用語統一

> 2026-11-05 依兩份整頁審視（設計與結構、用語）寫成。實作直接推 `Andy880828/Personal-Page` 的 `main`。
> 使用者已決定：公開 Email `andy0010502@gmail.com`；段落順序與深度敘事不改；精選作品 FORGE、Trim2Sort、syssec；
> 只有使用者知道的事實先留欄位、不編內容（清單見最後一節）。

## 進度

| 階段 | 狀態 | 紀錄 |
|---|---|---|
| A 聯絡方式 | [x] 2026-11-05 | [15a3f1b](https://github.com/Andy880828/Personal-Page/commit/15a3f1b) |
| B 作品：精選與證據欄位 | [x] 2026-11-05 | [15a3f1b](https://github.com/Andy880828/Personal-Page/commit/15a3f1b) |
| C 經歷與關於 | [x] 2026-11-05 | [15a3f1b](https://github.com/Andy880828/Personal-Page/commit/15a3f1b) |
| D 技能、首屏、版面與字級 | [x] 2026-11-05 | [15a3f1b](https://github.com/Andy880828/Personal-Page/commit/15a3f1b) |
| E 用語統一 | [x] 2026-11-05 | [15a3f1b](https://github.com/Andy880828/Personal-Page/commit/15a3f1b) |
| F 可及性與互動 | [x] 2026-11-05 | [15a3f1b](https://github.com/Andy880828/Personal-Page/commit/15a3f1b) |
| G 驗證、push、正式站確認 | [x] 2026-11-05 | 單元 115、Playwright 37 全過；已推 main（[15a3f1b](https://github.com/Andy880828/Personal-Page/commit/15a3f1b)） |

## 審視的主要結論

- 設計：「很好的下潛體驗、弱的求職證據」。10 秒內只看到名字與職稱；最前面四張作品都是 CLI／基礎建設，
  有截圖、有線上展示的作品被收進清單；全站沒有 Email；鍵盤 Tab 會跳過還沒浮出的內容。
- 用語：段落標題用比喻取代資訊；「資料／數據」「視覺化／可視化」混用；五個分類在技能、作品、經歷、meta 各叫不同名字；
  作品卡寫給工程師看（JSON envelope、冪等、KEV/EPSS）；歐化被動句（可以被驗證、可被程式呼叫）。

## A 聯絡方式（P0）

1. 聯絡卡改為「聯絡我」，第一個連結是 Email（`mailto:` + 一鍵複製），再來履歷、GitHub、Facebook。
2. 表單關閉時不再叫人「在 GitHub 開 issue」。
3. 手機上聯絡卡排在興趣與照片之前。
4. 指令列新增 `contact`（Email 與連結）、`resume`；`whoami` 加 Email；noscript 改成連結並附 Email。

## B 作品（P0／P1）

1. `Project` 加 `featured`（排序）、`outcome`（一行成果，有數字）、`role`（個人或團隊、負責什麼）；
   `outcome`／`role` 沒填就不顯示，不編。
2. 「全部」狀態的卡片只放精選：FORGE、Trim2Sort、syssec。前兩個用截圖封面；其餘進「其他作品」清單。
3. 清單列顯示中文名、年份、摘要、原始碼與線上展示連結（FORGE 的線上展示不會再消失）。
4. 大卡片：截圖與文字等高；終端機封面只用在 CLI 作品，固定高度、窄螢幕可橫向捲動，不截斷重疊。
5. 分類列不套 `uppercase`（「Agent 應用」不會變「AGENT 應用」）。

## C 經歷與關於（P1）

1. 經歷：拿掉高中；打工度假、鯨豚解說員、研究助理改成精簡列；台郡留成果欄位（待補）。
2. 經歷種類不用分類色（和技能／作品的顏色衝突），改用中性的點與文字標籤；「開發」→「專案」、「現在」→「至今」。
3. 7 字元短碼降成很淡的等寬小字，放在日期之後（保留沉積層意象，但不搶日期）。
4. 關於：「證照」→「語言成績與證照」，PADI 寫中文名。

## D 技能、首屏、版面與字級（P1／P2）

1. 技能：拿掉右側重複的英文圖例（tier 0 的 SVG 也不再和圖例重疊）。
2. 首屏：名字縮小；加第三顆「聯絡我」（「直接看作品」保留）；桌機座標行加「按 / 開指令列」。代表作連結沒加：作品段第一屏就是三張精選。
3. 段落標題的深度刻度改從 `lib/sections.ts` 讀（不再各段手寫，聯絡段 3,800 m 與選單一致）。
4. 字級：技能面板標題降到和作品卡同級；區塊小標（語言成績與證照、其他作品、不在工作的時候）拉開與內文的差；興趣清單降到內文大小。
5. 1024–1279 寬時儀表不佔內容右欄（只在 xl 以上顯示）。
6. 手機選單底色不透明。

## E 用語統一（P0–P2，依用語審視）

- 對照表：資料（不用數據；台郡正式職稱例外）、視覺化（不用可視化）、分類「資料探勘／資料視覺化／流程數位化／Agent 應用／全端網站」全站一致、
  講者辨識、安全強化、原始碼未公開、線上展示、動態效果、至今、4,000 m 寫法、中英間空格、全形標點。
- 標題：技能「技能：五個方向」、作品「精選作品」、經歷「經歷與學歷」；關於、聯絡保留。
- `roleSecond`：Oceanographer → Oceanography M.S.（和中文「海洋科學出身」一致）。
- 作品文案：syssec、local-transcribe、scout、homelab、andy-claude-kit 改成招募者讀得懂的說法；拿掉「僅供參考」與非亮點的維運細節。
- 指令列：`ctd` 改名 `here`（`ctd` 留作別名）、輸出改中文標籤；help、tier、motion、ascii 的說明與空格修正。
- meta description／OG 改成講清楚做什麼、有什麼作品。
- 文案 lint 範圍擴大到 `src/components/*.tsx` 與 `src/lib/console.ts`，加禁用「可視化」。

## F 可及性與互動（P1／P3）

1. `useReveal` 不再用 `visibility: hidden`（autoAlpha）：還沒浮出的內容也要能 Tab 到。
2. 「減少動態」按鈕的文字跟著狀態換；指令列按鈕的 sr-only 文字依開關切換；Logo 的 aria-label「回到頁首」。

## G 驗證

lint、型別、單元、build、size、Playwright（含新的聯絡與作品測試），`npm run fonts`；commit 帶 `[update-snapshots]`；部署後確認正式站。

## 沒做／和計畫不同

- LinkedIn：沒有網址，沒加。
- FORGE 的「即時計分」、syssec 封面的檢查數字（51 項中 21 pass、9 fail）沿用原文，沒有改成審視建議的版本（要看實際輸出）。
- 「話蛇添足工作室」與「軒齊生態工作室」兩個名字都保留，等確認。

## 待使用者補的事實（不編，補了再加）

| 項目 | 放在哪 |
|---|---|
| 台郡科技的工作內容與 2–3 條成果 | `src/content/experience.ts` 的 highlights |
| 精選作品的成果數字（FORGE 參與人數／期數、Trim2Sort 處理量或使用單位、syssec 掃過的主機數等） | `projects.ts` 的 `outcome` |
| RecipeCube 是幾人專題、你負責哪部分 | `projects.ts` 的 `role` |
| stock-analysis-app 的實際功能（目前 highlights 只有「投資紀錄、顧問式分析」） | `projects.ts` |
| 「話蛇添足工作室」與「軒齊生態工作室」哪個是正式名稱 | `projects.ts` |
| 碩士學制正式名稱（研究所／碩士班）、資展國際課程正式名稱 | `experience.ts` |
| LinkedIn 網址（若有） | `profile.ts` 的 links |
| 視覺化實際用過的工具（取代 Tailwind CSS） | `skills.ts` 的 viz tools |
