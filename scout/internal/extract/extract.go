// Package extract 把文章 HTML 轉成純文字，讓 agent 讀全文後寫摘要。
//
// 不依賴 HTML parser（golang.org/x/net/html 或 readability 類套件），
// 用「優先取 <article>，其次 <main>、<body>，去掉 script/style/nav 等」
// 的啟發式。對多數部落格與新聞站夠用；版面特殊的網站可能夾雜選單文字，
// agent 應以 warnings 為準判斷可信度。
package extract

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Result 是擷取結果。
type Result struct {
	Title      string `json:"title"`
	Text       string `json:"text"`
	Container  string `json:"container"` // article / main / body / document：取自哪個區塊
	TotalChars int    `json:"total_chars"`
	Returned   int    `json:"returned_chars"`
	Truncated  bool   `json:"truncated"`
}

var (
	dropBlockRe = regexp.MustCompile(`(?is)<(script|style|noscript|svg|nav|footer|aside|form|template|iframe)\b[^>]*>.*?</(?:script|style|noscript|svg|nav|footer|aside|form|template|iframe)>`)
	commentRe   = regexp.MustCompile(`(?s)<!--.*?-->`)
	titleRe     = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	ogTitleRe   = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']+)["']`)
	blockRe     = regexp.MustCompile(`(?i)</?(p|div|br|h[1-6]|li|ul|ol|tr|blockquote|pre|section|article|header|figure|figcaption|table)\b[^>]*>`)
	tagRe       = regexp.MustCompile(`(?s)<[^>]*>`)
	hSpaceRe    = regexp.MustCompile(`[ \t\f\r\x{00a0}]+`)
	nlRe        = regexp.MustCompile(`\n\s*\n+`)
)

// HTML 擷取標題與正文，正文超過 maxChars 時截斷並標示 Truncated。
func HTML(doc string, maxChars int) Result {
	r := Result{}
	if m := ogTitleRe.FindStringSubmatch(doc); m != nil {
		r.Title = clean(m[1])
	} else if m := titleRe.FindStringSubmatch(doc); m != nil {
		r.Title = clean(m[1])
	}
	doc = commentRe.ReplaceAllString(doc, " ")
	doc = dropBlockRe.ReplaceAllString(doc, " ")

	body, container := doc, "document"
	for _, tag := range []string{"article", "main", "body"} {
		if inner, ok := inside(doc, tag); ok && len(strings.TrimSpace(tagRe.ReplaceAllString(inner, ""))) > 200 {
			body, container = inner, tag
			break
		}
	}
	r.Container = container

	text := blockRe.ReplaceAllString(body, "\n")
	text = tagRe.ReplaceAllString(text, " ")
	text = html.UnescapeString(text)
	text = hSpaceRe.ReplaceAllString(text, " ")
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	text = strings.TrimSpace(nlRe.ReplaceAllString(strings.Join(lines, "\n"), "\n\n"))

	r.TotalChars = utf8.RuneCountInString(text)
	if maxChars > 0 && r.TotalChars > maxChars {
		text = string([]rune(text)[:maxChars])
		r.Truncated = true
	}
	r.Text = text
	r.Returned = utf8.RuneCountInString(text)
	return r
}

// inside 回傳第一個 <tag ...> 到最後一個 </tag> 之間的內容（處理巢狀時取最外層）。
func inside(doc, tag string) (string, bool) {
	// 直接在原字串上用不分大小寫的 regexp 找位置；先 ToLower 再取索引
	// 會因為部分 Unicode 字元轉小寫後位元組長度改變而錯位。
	open := regexp.MustCompile(`(?i)<` + tag + `\b[^>]*>`).FindStringIndex(doc)
	if open == nil {
		return "", false
	}
	closes := regexp.MustCompile(`(?i)</`+tag+`\s*>`).FindAllStringIndex(doc, -1)
	if len(closes) == 0 || closes[len(closes)-1][0] < open[1] {
		return doc[open[1]:], true
	}
	return doc[open[1]:closes[len(closes)-1][0]], true
}

func clean(s string) string {
	s = html.UnescapeString(tagRe.ReplaceAllString(s, " "))
	return strings.TrimSpace(hSpaceRe.ReplaceAllString(strings.ReplaceAll(s, "\n", " "), " "))
}
