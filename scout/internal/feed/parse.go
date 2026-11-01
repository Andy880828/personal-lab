// Package feed 抓取並解析來源：RSS 2.0、Atom 1.0、RSS 1.0 (RDF) 與
// Hacker News Algolia JSON。只用標準庫的 encoding/xml 與 encoding/json。
package feed

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

// excerptMax 是來源摘要保留的字元數。摘要只用來幫排序與 agent 初篩，
// 要讀全文請用 `scout read`。
const excerptMax = 400

// Entry 是解析後、尚未去重的一筆文章。
type Entry struct {
	URL           string
	Title         string
	Excerpt       string
	Author        string
	Published     *time.Time
	DiscussionURL string
	Points        int
	Comments      int
}

// ErrUnknownFormat 代表回應不是可辨識的 feed。
var ErrUnknownFormat = errors.New("unrecognised feed format (expected <rss>, <feed> or <rdf:RDF>)")

type xmlLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Text string `xml:",chardata"`
}

type xmlEntry struct {
	Title       string    `xml:"title"`
	Links       []xmlLink `xml:"link"`
	Description string    `xml:"description"`
	Summary     string    `xml:"summary"`
	Content     string    `xml:"content"`
	Encoded     string    `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	PubDate     string    `xml:"pubDate"`
	Published   string    `xml:"published"`
	Updated     string    `xml:"updated"`
	DCDate      string    `xml:"http://purl.org/dc/elements/1.1/ date"`
	Author      struct {
		Name string `xml:"name"`
		Text string `xml:",chardata"`
	} `xml:"author"`
	Creator  string `xml:"http://purl.org/dc/elements/1.1/ creator"`
	GUID     string `xml:"guid"`
	ID       string `xml:"id"`
	Comments string `xml:"comments"`
}

type xmlDoc struct {
	XMLName xml.Name
	Channel struct {
		Items []xmlEntry `xml:"item"`
	} `xml:"channel"`
	Items   []xmlEntry `xml:"item"`  // RSS 1.0 (RDF) 的 item 在根節點下
	Entries []xmlEntry `xml:"entry"` // Atom
}

// ParseFeed 自動偵測並解析 RSS/Atom/RDF，回傳格式名稱與文章。
// 沒有可用連結的項目會被略過，略過數量透過 skipped 回報（不無聲丟棄）。
func ParseFeed(body []byte) (format string, entries []Entry, skipped int, err error) {
	dec := xml.NewDecoder(bytes.NewReader(body))
	dec.Strict = false
	dec.Entity = xml.HTMLEntity
	dec.CharsetReader = charsetReader
	var doc xmlDoc
	if err := dec.Decode(&doc); err != nil {
		return "", nil, 0, fmt.Errorf("parse xml: %w", err)
	}
	var raw []xmlEntry
	switch strings.ToLower(doc.XMLName.Local) {
	case "rss":
		format, raw = "rss", doc.Channel.Items
	case "feed":
		format, raw = "atom", doc.Entries
	case "rdf":
		format, raw = "rdf", doc.Items
	default:
		return "", nil, 0, ErrUnknownFormat
	}
	for _, x := range raw {
		e := Entry{
			URL:           pickLink(x),
			Title:         cleanText(x.Title, 300),
			Excerpt:       cleanText(firstNonEmpty(x.Description, x.Summary, x.Encoded, x.Content), excerptMax),
			Author:        cleanText(firstNonEmpty(x.Author.Name, x.Creator, x.Author.Text), 100),
			Published:     parseTime(firstNonEmpty(x.Published, x.PubDate, x.DCDate, x.Updated)),
			DiscussionURL: strings.TrimSpace(x.Comments),
		}
		if e.URL == "" {
			skipped++
			continue
		}
		if e.Title == "" {
			e.Title = e.URL
		}
		entries = append(entries, e)
	}
	return format, entries, skipped, nil
}

func pickLink(x xmlEntry) string {
	// Atom：優先 rel="alternate"（或沒有 rel），其次任何有 href 的連結。
	var fallback string
	for _, l := range x.Links {
		href := strings.TrimSpace(l.Href)
		if href == "" {
			href = strings.TrimSpace(l.Text) // RSS 的 <link>文字</link>
		}
		if href == "" {
			continue
		}
		if l.Rel == "" || l.Rel == "alternate" {
			return href
		}
		if fallback == "" && l.Rel != "self" && l.Rel != "replies" {
			fallback = href
		}
	}
	if fallback != "" {
		return fallback
	}
	for _, g := range []string{x.GUID, x.ID} {
		g = strings.TrimSpace(g)
		if strings.HasPrefix(g, "http://") || strings.HasPrefix(g, "https://") {
			return g
		}
	}
	return ""
}

type hnResponse struct {
	Hits []struct {
		ObjectID    string `json:"objectID"`
		Title       string `json:"title"`
		URL         string `json:"url"`
		Author      string `json:"author"`
		Points      int    `json:"points"`
		NumComments int    `json:"num_comments"`
		CreatedAtI  int64  `json:"created_at_i"`
		StoryText   string `json:"story_text"`
	} `json:"hits"`
}

// ParseHN 解析 Algolia HN search API。Ask HN 這類沒有外部連結的貼文，
// 以 HN 討論串本身作為文章網址。
func ParseHN(body []byte) ([]Entry, error) {
	var r hnResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("parse hn json: %w", err)
	}
	var out []Entry
	for _, h := range r.Hits {
		disc := "https://news.ycombinator.com/item?id=" + h.ObjectID
		u := h.URL
		if u == "" {
			u = disc
		}
		var pub *time.Time
		if h.CreatedAtI > 0 {
			t := time.Unix(h.CreatedAtI, 0).UTC()
			pub = &t
		}
		out = append(out, Entry{
			URL: u, Title: cleanText(h.Title, 300), Excerpt: cleanText(h.StoryText, excerptMax),
			Author: h.Author, Published: pub, DiscussionURL: disc,
			Points: h.Points, Comments: h.NumComments,
		})
	}
	return out, nil
}

// Parse 依來源種類分派 parser。
func Parse(kind core.SourceKind, body []byte) (format string, entries []Entry, skipped int, err error) {
	switch kind {
	case core.KindHN:
		e, err := ParseHN(body)
		return "hn-algolia", e, 0, err
	case core.KindFeed:
		return ParseFeed(body)
	}
	return "", nil, 0, fmt.Errorf("unknown source kind %q", kind)
}

var (
	tagRe   = regexp.MustCompile(`(?s)<[^>]*>`)
	spaceRe = regexp.MustCompile(`\s+`)
)

// cleanText 去掉 HTML 標籤與實體、壓縮空白，並在 max 個字元處截斷。
func cleanText(s string, max int) string {
	s = tagRe.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	// &nbsp; 解碼後是 U+00A0，Go regexp 的 \s 不涵蓋它，要先換成一般空白。
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.TrimSpace(spaceRe.ReplaceAllString(s, " "))
	if utf8.RuneCountInString(s) > max {
		r := []rune(s)
		s = strings.TrimSpace(string(r[:max])) + "…"
	}
	return s
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

var timeLayouts = []string{
	time.RFC1123Z, time.RFC1123, time.RFC3339Nano, time.RFC3339,
	"Mon, 2 Jan 2006 15:04:05 -0700", "Mon, 2 Jan 2006 15:04:05 MST",
	"Mon, 02 Jan 2006 15:04 -0700", "2 Jan 2006 15:04:05 -0700", "02 Jan 2006 15:04:05 MST",
	"Mon, 2 Jan 2006 15:04:05 Z", "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02",
}

// parseTime 嘗試常見格式；全部失敗就回傳 nil（未知），不會假裝是現在。
func parseTime(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, l := range timeLayouts {
		if t, err := time.Parse(l, s); err == nil {
			t = t.UTC()
			return &t
		}
	}
	return nil
}

// charsetReader 支援 UTF-8 以外最常見的兩種宣告；其他編碼明確報錯，
// 讓 agent 知道是哪個來源、哪種編碼不支援，而不是解出亂碼。
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(charset)) {
	case "utf-8", "utf8", "us-ascii", "ascii":
		return input, nil
	case "iso-8859-1", "latin1", "latin-1":
		b, err := io.ReadAll(input)
		if err != nil {
			return nil, err
		}
		r := make([]rune, len(b))
		for i, c := range b {
			r[i] = rune(c)
		}
		return strings.NewReader(string(r)), nil
	}
	return nil, fmt.Errorf("unsupported charset %q", charset)
}
