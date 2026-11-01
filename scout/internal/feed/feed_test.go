package feed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

func read(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// 回歸測試：xml.HTMLAutoClose 會把 <link> 當成空元素，曾讓所有 RSS 2.0
// feed 解析出 0 筆（Atom 與 HN 正常，所以不易察覺）。
func TestParseRSSLinkText(t *testing.T) {
	format, entries, skipped, err := ParseFeed(read(t, "rss.xml"))
	if err != nil || format != "rss" {
		t.Fatalf("format=%q err=%v", format, err)
	}
	if len(entries) != 2 || skipped != 1 {
		t.Fatalf("entries=%d skipped=%d, want 2 and 1", len(entries), skipped)
	}
	e := entries[0]
	if e.URL != "https://example.com/posts/1?utm_source=rss" {
		t.Errorf("url=%q", e.URL)
	}
	if e.Title != "First & best post" {
		t.Errorf("title=%q (entities should be decoded)", e.Title)
	}
	if e.Excerpt != "Hello world" || e.Author != "Alice" {
		t.Errorf("excerpt=%q author=%q", e.Excerpt, e.Author)
	}
	if e.Published == nil || !e.Published.Equal(time.Date(2026, 10, 31, 10, 0, 0, 0, time.UTC)) {
		t.Errorf("published=%v", e.Published)
	}
	if entries[1].URL != "https://example.com/posts/2" {
		t.Errorf("guid fallback url=%q", entries[1].URL)
	}
	if entries[1].Published != nil {
		t.Errorf("unparseable date must stay nil (unknown), got %v", entries[1].Published)
	}
}

func TestParseAtomPrefersAlternate(t *testing.T) {
	format, entries, _, err := ParseFeed(read(t, "atom.xml"))
	if err != nil || format != "atom" || len(entries) != 2 {
		t.Fatalf("format=%q n=%d err=%v", format, len(entries), err)
	}
	if entries[0].URL != "https://example.org/a/1/" {
		t.Errorf("want rel=alternate link, got %q", entries[0].URL)
	}
	// type="html"：&lt;b&gt; 是真的標籤要去掉，&amp;lt; 才是字面上的 "<"。
	if entries[0].Title != "Atom <entry> bold" || entries[0].Author != "Bob" {
		t.Errorf("title=%q author=%q", entries[0].Title, entries[0].Author)
	}
	if !entries[0].Published.Equal(time.Date(2026, 10, 30, 8, 0, 0, 0, time.UTC)) {
		t.Errorf("published should prefer <published> over <updated>: %v", entries[0].Published)
	}
	if entries[1].URL != "https://example.org/a/2" || entries[1].Published == nil {
		t.Errorf("bare link entry: %+v", entries[1])
	}
}

func TestParseRDFAndUnknown(t *testing.T) {
	format, entries, _, err := ParseFeed(read(t, "rdf.xml"))
	if err != nil || format != "rdf" || len(entries) != 1 || entries[0].URL != "https://arxiv.org/abs/1" {
		t.Fatalf("rdf: format=%q entries=%+v err=%v", format, entries, err)
	}
	if _, _, _, err := ParseFeed([]byte("<html><body>nope</body></html>")); err != ErrUnknownFormat {
		t.Fatalf("html page: want ErrUnknownFormat, got %v", err)
	}
	if _, _, _, err := ParseFeed([]byte(`<?xml version="1.0" encoding="big5"?><rss></rss>`)); err == nil || !strings.Contains(err.Error(), "big5") {
		t.Fatalf("unsupported charset must be reported by name, got %v", err)
	}
}

func TestParseHN(t *testing.T) {
	entries, err := ParseHN(read(t, "hn.json"))
	if err != nil || len(entries) != 2 {
		t.Fatalf("n=%d err=%v", len(entries), err)
	}
	if entries[0].Points != 250 || entries[0].Comments != 80 || entries[0].DiscussionURL != "https://news.ycombinator.com/item?id=100" {
		t.Errorf("story: %+v", entries[0])
	}
	if entries[1].URL != "https://news.ycombinator.com/item?id=101" || entries[1].Excerpt != "Question & body" {
		t.Errorf("ask hn should link to discussion: %+v", entries[1])
	}
}

func TestCanonical(t *testing.T) {
	_, a, _ := Canonical("https://www.Example.com/post/?utm_source=x&b=2&a=1#frag")
	_, b, _ := Canonical("http://example.com/post?a=1&b=2&fbclid=zzz")
	if a != b {
		t.Errorf("same article must get the same id: %s vs %s", a, b)
	}
	_, c, _ := Canonical("https://example.com/post?a=2&b=2")
	if a == c {
		t.Error("different meaningful query must give a different id")
	}
	disp, _, _ := Canonical("https://Example.com/x/?utm_medium=y")
	if disp != "https://example.com/x" {
		t.Errorf("display=%q", disp)
	}
	for _, bad := range []string{"", "ftp://x.com/a", "/relative", "javascript:alert(1)"} {
		if _, _, ok := Canonical(bad); ok {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestMerge(t *testing.T) {
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	recent := now.Add(-2 * time.Hour)
	old := now.AddDate(0, 0, -(MaxEntryAgeDays + 1))
	st := &core.State{Items: map[string]*core.Item{}}
	ms := Merge(st, core.Source{ID: "a"}, []Entry{
		{URL: "https://x.com/1", Title: "One", Published: &recent},
		{URL: "https://x.com/old", Title: "Old", Published: &old},
		{URL: "not a url", Title: "Bad"},
	}, now)
	if ms.New != 1 || ms.Old != 1 || ms.Invalid != 1 {
		t.Fatalf("stats %+v", ms)
	}
	ms = Merge(st, core.Source{ID: "hn"}, []Entry{{URL: "https://www.x.com/1/?utm_source=hn", Title: "One", Points: 99, DiscussionURL: "https://news.ycombinator.com/item?id=1"}}, now)
	if ms.Seen != 1 || len(st.Items) != 1 {
		t.Fatalf("cross-source duplicate must merge: %+v, items=%d", ms, len(st.Items))
	}
	for _, it := range st.Items {
		if len(it.Sightings) != 2 || it.Sightings[1].Points != 99 {
			t.Fatalf("sightings %+v", it.Sightings)
		}
	}
	// 同一來源再看到一次：更新分數，不新增 sighting。
	Merge(st, core.Source{ID: "hn"}, []Entry{{URL: "https://x.com/1", Points: 150}}, now)
	for _, it := range st.Items {
		if len(it.Sightings) != 2 || it.Sightings[1].Points != 150 {
			t.Fatalf("re-sighting should update in place: %+v", it.Sightings)
		}
	}
}

func TestFetchConditionalGet(t *testing.T) {
	body := read(t, "rss.xml")
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.Header.Get("If-None-Match") == `"v1"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		w.Write(body)
	}))
	defer srv.Close()
	c := NewClient()
	src := core.Source{ID: "s", Kind: core.KindFeed, URL: srv.URL}
	r1 := c.FetchOne(context.Background(), src, nil)
	if r1.Status != core.FetchOK || len(r1.Entries) != 2 || r1.ETag != `"v1"` {
		t.Fatalf("first fetch %+v", r1)
	}
	r2 := c.FetchOne(context.Background(), src, &core.SourceState{ETag: r1.ETag})
	if r2.Status != core.FetchNotModified || r2.ETag != `"v1"` {
		t.Fatalf("second fetch should be 304 and keep the etag: %+v", r2)
	}
}

func TestFetchErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/500" {
			http.Error(w, "boom", 500)
			return
		}
		w.Write([]byte("<html>not a feed</html>"))
	}))
	defer srv.Close()
	c := NewClient()
	results := c.FetchAll(context.Background(), []core.Source{
		{ID: "e500", Kind: core.KindFeed, URL: srv.URL + "/500"},
		{ID: "html", Kind: core.KindFeed, URL: srv.URL + "/page"},
	}, nil)
	if results[0].Status != core.FetchError || !strings.Contains(results[0].Err.Error(), "500") {
		t.Errorf("500: %+v", results[0])
	}
	if results[1].Status != core.FetchError || results[1].Err != ErrUnknownFormat {
		t.Errorf("html: %+v", results[1])
	}
}
