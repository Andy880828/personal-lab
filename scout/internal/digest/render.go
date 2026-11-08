package digest

import (
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

// Markdown 輸出適合直接貼進 Claude 對話或筆記的版本。
func Markdown(d *core.Digest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# 科技精選 %s\n\n", d.Date)
	if d.Note != "" {
		fmt.Fprintf(&b, "> %s\n\n", d.Note)
	}
	for i, it := range d.Items {
		fmt.Fprintf(&b, "%d. **[%s](%s)**\n   %s\n", i+1, mdEscape(it.Title), it.URL, it.Summary)
		meta := "來源：" + strings.Join(it.Sources, "、")
		if it.DiscussionURL != "" {
			meta += fmt.Sprintf(" · [討論](%s)", it.DiscussionURL)
		}
		fmt.Fprintf(&b, "   <sub>%s</sub>\n\n", meta)
	}
	fmt.Fprintf(&b, "<sub>從 %d 篇候選中挑選 · 時間窗 %s – %s</sub>\n", d.CandidatesConsidered,
		d.WindowFrom.Local().Format("01/02 15:04"), d.WindowTo.Local().Format("01/02 15:04"))
	return b.String()
}

// Text 是最精簡的純文字版，之後接 Telegram 之類的通道時可直接使用。
func Text(d *core.Digest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "科技精選 %s\n", d.Date)
	if d.Note != "" {
		fmt.Fprintf(&b, "%s\n", d.Note)
	}
	for i, it := range d.Items {
		fmt.Fprintf(&b, "\n%d. %s\n%s\n%s\n", i+1, it.Title, it.Summary, it.URL)
	}
	return b.String()
}

func mdEscape(s string) string {
	return strings.NewReplacer("[", `\[`, "]", `\]`).Replace(s)
}

// 版面採報紙頭版：報頭、雙線、導讀、頭條＋側欄、細線分欄。自足（內嵌 CSS、無 JS、無外部資源）。
const css = `@font-face{font-family:LatinQuote;src:local("Georgia"),local("Times New Roman"),local("Liberation Serif"),local("DejaVu Serif");unicode-range:U+2018,U+2019}
:root{color-scheme:light dark;--paper:#f6f1e7;--ink:#1a1712;--muted:#6f675a;--hair:#cfc6b4;--accent:#a8281e;
--serif:LatinQuote,"Source Han Serif TC","Noto Serif TC","Noto Serif CJK TC","Songti TC","PMingLiU",Georgia,"Times New Roman",serif;
--sans:system-ui,-apple-system,"PingFang TC","Noto Sans TC","Microsoft JhengHei",sans-serif}
@media(prefers-color-scheme:dark){:root{--paper:#15130f;--ink:#ece5d8;--muted:#9a917f;--hair:#38332a;--accent:#e5674f}}
*{box-sizing:border-box}html{-webkit-text-size-adjust:100%}
body{margin:0;background:var(--paper);color:var(--ink);font:17px/1.8 var(--serif)}
a{color:inherit;text-decoration:none}a:hover{color:var(--accent)}a:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
.sheet{max-width:74rem;margin:0 auto;padding:1.25rem 1.25rem 3rem}
.topline{display:flex;flex-wrap:wrap;justify-content:space-between;gap:.15rem 1.5rem;font:500 .78rem/1.5 var(--sans);letter-spacing:.06em;color:var(--muted);padding-bottom:.6rem;border-bottom:1px solid var(--hair)}
.name{font:900 clamp(3.25rem,11vw,7.25rem)/1.1 var(--serif);letter-spacing:.1em;text-indent:.1em;text-align:center;margin:.9rem 0 .35rem}
.motto{font:500 .78rem/1.5 var(--sans);letter-spacing:.32em;text-indent:.32em;text-align:center;text-transform:uppercase;color:var(--muted);margin:0 0 .9rem}
.rule{border:0;border-top:4px double var(--ink);margin:0}
.label{font:700 .75rem/1.5 var(--sans);letter-spacing:.14em;color:var(--accent)}
.brief{display:grid;gap:.25rem 2rem;padding:1.25rem 0;border-bottom:1px solid var(--ink)}
.brief p{margin:0;font-size:clamp(1.1rem,2.2vw,1.35rem);line-height:1.75;font-style:normal}
.front{display:grid;grid-template-columns:minmax(0,1fr);gap:0}.main{display:contents}.lead{order:1}.side{order:2}.rest{order:3}
article{padding:1.4rem 0;border-bottom:1px solid var(--hair)}
.kicker{font:700 .74rem/1.5 var(--sans);letter-spacing:.08em;color:var(--accent);margin:0 0 .45rem}
h2{font-weight:800;line-height:1.3;margin:0 0 .6rem;text-wrap:balance;overflow-wrap:break-word;font-size:1.35rem}
.lead h2{font-size:clamp(1.9rem,4.4vw,3.15rem);font-weight:900;line-height:1.2;margin-bottom:.9rem}
.deck{margin:0 0 .65rem}.lead .deck{font-size:1.14rem;line-height:1.9}
.lead .deck::first-letter{float:left;font-size:3.5em;line-height:.95;font-weight:900;padding:.04em .12em 0 0;color:var(--accent)}
.why{margin:0 0 .4rem;font:400 .82rem/1.65 var(--sans);color:var(--muted)}.why b{font-weight:700}
.links{font:500 .8rem/1.5 var(--sans)}.links a{color:var(--accent);margin-right:1rem}
.pager{display:flex;justify-content:space-between;gap:1rem;font:600 .86rem/1.5 var(--sans);padding:1.1rem 0;border-bottom:1px solid var(--ink)}
.archive{padding-top:1.75rem}.archive h3{font:900 1.5rem/1.3 var(--serif);margin:0 0 1rem;padding-bottom:.5rem;border-bottom:4px double var(--ink)}
.month{font:700 .78rem/1.5 var(--sans);letter-spacing:.1em;color:var(--accent);margin:1.2rem 0 .25rem}
.days{list-style:none;margin:0;padding:0}.days li{border-bottom:1px solid var(--hair)}
.days a{display:grid;grid-template-columns:4.5rem minmax(0,1fr);gap:.75rem;padding:.6rem 0}
.days time{font:600 .84rem/1.9 var(--sans);color:var(--muted)}.days span{overflow-wrap:break-word}
.colophon{margin-top:2rem;font:400 .78rem/1.7 var(--sans);color:var(--muted)}
@media(min-width:56rem){
.brief{grid-template-columns:8rem minmax(0,1fr)}
.front{grid-template-columns:repeat(12,minmax(0,1fr))}
.main{display:block;grid-column:span 8;padding-right:2rem;border-right:1px solid var(--hair)}.main:only-child{grid-column:1/-1;padding-right:0;border-right:0}
.side{grid-column:span 4;padding-left:2rem}.side article:last-child{border-bottom:0}
.rest{columns:2;column-gap:2.5rem;column-rule:1px solid var(--hair);border-top:1px solid var(--ink)}
.rest article{break-inside:avoid;padding:1.25rem 0}
.days{columns:2;column-gap:3rem}.days li{break-inside:avoid}}
@media print{.pager,.archive{display:none}}`

func page(title, body string) string {
	return `<!doctype html><html lang="zh-Hant"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">` +
		`<title>` + html.EscapeString(title) + `</title><style>` + css + `</style></head><body><div class="sheet">` + body + `</div></body></html>`
}

// safeURL 只放行 http／https，避免來源資料把 javascript: 之類的連結帶進公開頁面。
func safeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	return html.EscapeString(raw)
}

func longDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	weekday := [...]string{"日", "一", "二", "三", "四", "五", "六"}[t.Weekday()]
	return fmt.Sprintf("%d 年 %d 月 %d 日 星期%s", t.Year(), int(t.Month()), t.Day(), weekday)
}

func writeArticle(b *strings.Builder, it core.DigestItem, class string) {
	fmt.Fprintf(b, `<article class="%s"><p class="kicker">%s</p>`, class, html.EscapeString(strings.Join(it.Sources, " · ")))
	title := html.EscapeString(it.Title)
	if link := safeURL(it.URL); link != "" {
		fmt.Fprintf(b, `<h2><a href="%s" rel="noopener">%s</a></h2>`, link, title)
	} else {
		fmt.Fprintf(b, `<h2>%s</h2>`, title)
	}
	fmt.Fprintf(b, `<p class="deck">%s</p>`, html.EscapeString(it.Summary))
	if it.Why != "" {
		fmt.Fprintf(b, `<p class="why"><b>入選理由</b>　%s</p>`, html.EscapeString(it.Why))
	}
	if link := safeURL(it.DiscussionURL); link != "" {
		fmt.Fprintf(b, `<p class="links"><a href="%s" rel="noopener">讀者討論 →</a></p>`, link)
	}
	b.WriteString(`</article>`)
}

// edition 畫出一整期：報頭、導讀、頭條與其餘報導。issue 為 0 時不顯示期數。
func edition(b *strings.Builder, d *core.Digest, issue int) {
	issueText := "每日編選"
	if issue > 0 {
		issueText = fmt.Sprintf("第 %d 期", issue)
	}
	fmt.Fprintf(b, `<header><div class="topline"><span>%s</span><span>%s</span><span>%d 則精選｜%d 篇候選</span></div>`,
		issueText, longDate(d.Date), len(d.Items), d.CandidatesConsidered)
	b.WriteString(`<div class="name"><a href="index.html">科技精選</a></div><p class="motto">Scout Daily</p><hr class="rule"></header><main>`)
	if d.Note != "" {
		fmt.Fprintf(b, `<section class="brief"><span class="label">今日導讀</span><p>%s</p></section>`, html.EscapeString(d.Note))
	}
	// DOM 先放主欄（頭條＋其餘）再放側欄；窄螢幕用 display:contents 與 order 排回重要性順序。
	b.WriteString(`<div class="front"><div class="main">`)
	if len(d.Items) > 0 {
		writeArticle(b, d.Items[0], "lead")
	}
	if len(d.Items) > 3 {
		b.WriteString(`<div class="rest">`)
		for _, it := range d.Items[3:] {
			writeArticle(b, it, "story")
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)
	if len(d.Items) > 1 {
		b.WriteString(`<div class="side">`)
		for _, it := range d.Items[1:min(len(d.Items), 3)] {
			writeArticle(b, it, "story")
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div></main>`)
}

func colophon(b *strings.Builder, d *core.Digest) {
	fmt.Fprintf(b, `<p class="colophon">摘要與導讀由 AI 依原文撰寫，可能有誤，請以原文為準。收錄時間 %s – %s。</p>`,
		d.WindowFrom.Local().Format("01/02 15:04"), d.WindowTo.Local().Format("01/02 15:04"))
}

func pager(b *strings.Builder, newer, older string) {
	b.WriteString(`<nav class="pager"><span>`)
	if older != "" {
		fmt.Fprintf(b, `<a href="%s.html">← 前一期 %s</a>`, html.EscapeString(older), html.EscapeString(older))
	}
	b.WriteString(`</span><span>`)
	if newer != "" {
		fmt.Fprintf(b, `<a href="%s.html">後一期 %s →</a>`, html.EscapeString(newer), html.EscapeString(newer))
	} else {
		b.WriteString(`<a href="index.html#archive">所有期數</a>`)
	}
	b.WriteString(`</span></nav>`)
}

// HTMLPage 輸出單日頁面。newer／older 是相鄰期的日期（沒有時傳空字串），issue 是期數（未知時傳 0）。
func HTMLPage(d *core.Digest, newer, older string, issue int) string {
	var b strings.Builder
	edition(&b, d, issue)
	pager(&b, newer, older)
	colophon(&b, d)
	return page("科技精選 "+d.Date, b.String())
}

// HTMLIndex 輸出首頁：最新一期的完整版面，加上依月份分組的往期。digests 由新到舊。
func HTMLIndex(digests []*core.Digest) string {
	var b strings.Builder
	if len(digests) == 0 {
		return page("科技精選", `<div class="name">科技精選</div>`)
	}
	latest := digests[0]
	edition(&b, latest, len(digests))
	older := ""
	if len(digests) > 1 {
		older = digests[1].Date
	}
	pager(&b, "", older)
	b.WriteString(`<section class="archive" id="archive"><h3>往期</h3>`)
	month := ""
	for _, d := range digests {
		if m := d.Date[:min(len(d.Date), 7)]; m != month {
			if month != "" {
				b.WriteString(`</ol>`)
			}
			month = m
			fmt.Fprintf(&b, `<p class="month">%s</p><ol class="days">`, html.EscapeString(strings.Replace(m, "-", " 年 ", 1)+" 月"))
		}
		first := ""
		if len(d.Items) > 0 {
			first = d.Items[0].Title
		}
		fmt.Fprintf(&b, `<li><a href="%s.html"><time datetime="%s">%s</time><span>%s</span></a></li>`,
			html.EscapeString(d.Date), html.EscapeString(d.Date), html.EscapeString(d.Date[min(len(d.Date), 5):]), html.EscapeString(first))
	}
	b.WriteString(`</ol></section>`)
	colophon(&b, latest)
	return page("科技精選", b.String())
}
