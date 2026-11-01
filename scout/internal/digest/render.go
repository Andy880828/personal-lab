package digest

import (
	"fmt"
	"html"
	"strings"

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

const css = `:root{color-scheme:light dark;--fg:#1f2328;--muted:#656d76;--bg:#fff;--card:#f6f8fa;--accent:#0969da}
@media(prefers-color-scheme:dark){:root{--fg:#e6edf3;--muted:#8d96a0;--bg:#0d1117;--card:#161b22;--accent:#4493f8}}
*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--fg);font:17px/1.75 system-ui,-apple-system,"PingFang TC","Noto Sans TC","Microsoft JhengHei",sans-serif}
main{max-width:42rem;margin:0 auto;padding:2.5rem 1.25rem 4rem}h1{font-size:1.6rem;margin:0 0 .25rem}
.sub{color:var(--muted);font-size:.85rem;margin:0 0 2rem}.note{border-left:3px solid var(--accent);padding:.25rem .9rem;margin:0 0 2rem;color:var(--muted)}
article{background:var(--card);border-radius:12px;padding:1.1rem 1.25rem;margin:0 0 1rem}article h2{font-size:1.05rem;line-height:1.5;margin:0 0 .4rem}
a{color:var(--accent);text-decoration:none}a:hover{text-decoration:underline}article p{margin:0 0 .5rem}.meta{color:var(--muted);font-size:.8rem}
ol.days{list-style:none;padding:0}ol.days li{padding:.6rem 0;border-bottom:1px solid var(--card)}nav{margin-top:2.5rem;font-size:.9rem}`

func page(title, body string) string {
	return `<!doctype html><html lang="zh-Hant"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">` +
		`<title>` + html.EscapeString(title) + `</title><style>` + css + `</style></head><body><main>` + body + `</main></body></html>`
}

// HTMLPage 輸出單日頁面：自足（內嵌 CSS、無 JS、無外部資源），可直接用 Caddy file_server 提供。
func HTMLPage(d *core.Digest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<h1>科技精選 %s</h1><p class=\"sub\">從 %d 篇候選中挑選 %d 篇</p>", d.Date, d.CandidatesConsidered, len(d.Items))
	if d.Note != "" {
		fmt.Fprintf(&b, "<p class=\"note\">%s</p>", html.EscapeString(d.Note))
	}
	for _, it := range d.Items {
		fmt.Fprintf(&b, "<article><h2><a href=\"%s\">%s</a></h2><p>%s</p><div class=\"meta\">%s",
			html.EscapeString(it.URL), html.EscapeString(it.Title), html.EscapeString(it.Summary), html.EscapeString(strings.Join(it.Sources, "、")))
		if it.DiscussionURL != "" {
			fmt.Fprintf(&b, " · <a href=\"%s\">討論</a>", html.EscapeString(it.DiscussionURL))
		}
		b.WriteString("</div></article>")
	}
	b.WriteString(`<nav><a href="index.html">← 所有日期</a></nav>`)
	return page("科技精選 "+d.Date, b.String())
}

// HTMLIndex 輸出日期索引頁，digests 由新到舊。
func HTMLIndex(digests []*core.Digest) string {
	var b strings.Builder
	b.WriteString(`<h1>科技精選</h1><p class="sub">每日由 agent 從自選來源挑選</p><ol class="days">`)
	for _, d := range digests {
		first := ""
		if len(d.Items) > 0 {
			first = " — " + html.EscapeString(d.Items[0].Title)
		}
		fmt.Fprintf(&b, "<li><a href=\"%s.html\">%s</a> <span class=\"meta\">%d 篇%s</span></li>", d.Date, d.Date, len(d.Items), first)
	}
	b.WriteString("</ol>")
	return page("科技精選", b.String())
}
