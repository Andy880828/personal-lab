package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
	"github.com/Andy880828/personal-lab/scout/internal/feed"
	"github.com/Andy880828/personal-lab/scout/internal/store"
)

type env struct {
	t    *testing.T
	home string
	now  time.Time
	srv  *httptest.Server
}

type result struct {
	Code int
	Env  struct {
		OK       bool            `json:"ok"`
		Command  string          `json:"command"`
		Schema   string          `json:"schema"`
		Data     json.RawMessage `json:"data"`
		Warnings []string        `json:"warnings"`
		Error    *Error          `json:"error"`
	}
	Stdout string
}

func newEnv(t *testing.T) *env {
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	base := srv.URL
	pub := func(h int) string { return now.Add(-time.Duration(h) * time.Hour).Format(time.RFC1123Z) }
	rss := fmt.Sprintf(`<?xml version="1.0"?><rss version="2.0"><channel>
<item><title>Claude ships a new agent SDK</title><link>%[4]s/article</link><pubDate>%[1]s</pubDate></item>
<item><title>Crypto token pumps</title><link>%[4]s/crypto</link><pubDate>%[2]s</pubDate></item>
<item><title>Ancient post</title><link>%[4]s/ancient</link><pubDate>%[3]s</pubDate></item>
</channel></rss>`, pub(3), pub(4), pub(24*30), base)
	hn := fmt.Sprintf(`{"hits":[{"objectID":"1","title":"Claude ships a new agent SDK","url":"%s/article/?utm_source=hn","points":300,"num_comments":120,"created_at_i":%d}]}`, base, now.Add(-2*time.Hour).Unix())
	mux.HandleFunc("/rss", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(rss)) })
	mux.HandleFunc("/hn", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(hn)) })
	mux.HandleFunc("/broken", func(w http.ResponseWriter, r *http.Request) { http.Error(w, "nope", 503) })
	mux.HandleFunc("/article", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<html><body><article><p>" + strings.Repeat("Article body. ", 40) + "</p></article></body></html>"))
	})
	mux.HandleFunc("/pdf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.4"))
	})
	return &env{t: t, home: t.TempDir(), now: now, srv: srv}
}

func (e *env) run(stdin string, args ...string) result {
	e.t.Helper()
	var out, errb bytes.Buffer
	rt := &Runtime{JSON: true, Home: e.home, Stdout: &out, Stderr: &errb, Stdin: strings.NewReader(stdin),
		Now: func() time.Time { return e.now }, Client: feed.NewClient()}
	var r result
	r.Code = RunWith(context.Background(), rt, args)
	r.Stdout = out.String()
	if err := json.Unmarshal(out.Bytes(), &r.Env); err != nil {
		e.t.Fatalf("%v: stdout is not one JSON envelope: %v\n%s\nstderr: %s", args, err, out.String(), errb.String())
	}
	if r.Env.Schema != core.SchemaVersion {
		e.t.Fatalf("%v: schema %q", args, r.Env.Schema)
	}
	return r
}

func (e *env) data(r result, v any) {
	e.t.Helper()
	if err := json.Unmarshal(r.Env.Data, v); err != nil {
		e.t.Fatalf("decode data: %v\n%s", err, r.Env.Data)
	}
}

// 寫入只含測試伺服器來源的 watchlist（不碰真實網路）。
func (e *env) seed() {
	w := &core.Watchlist{
		Sources: []core.Source{
			{ID: "blog", Name: "Blog", Kind: core.KindFeed, URL: e.srv.URL + "/rss", Weight: 1, Followed: true},
			{ID: "hn", Name: "HN", Kind: core.KindHN, URL: e.srv.URL + "/hn", Weight: 1.2},
			{ID: "broken", Name: "Broken", Kind: core.KindFeed, URL: e.srv.URL + "/broken", Weight: 1},
		},
		Boost: []string{"claude"}, Mute: []string{"crypto"},
	}
	if err := store.New(e.home).SaveWatchlist(w); err != nil {
		e.t.Fatal(err)
	}
}

func TestFullWorkflow(t *testing.T) {
	e := newEnv(t)

	r := e.run("", "status")
	if !r.Env.OK || !strings.Contains(string(r.Env.Data), "scout init") {
		t.Fatalf("status before init should point to init: %s", r.Env.Data)
	}
	if r := e.run("", "candidates"); r.Env.OK || r.Env.Error.Code != "no_data" {
		t.Fatalf("candidates without watchlist: %+v", r.Env.Error)
	}
	e.seed()

	r = e.run("", "fetch")
	var fetched struct {
		Sources []fetchRow     `json:"sources"`
		Totals  map[string]int `json:"totals"`
	}
	e.data(r, &fetched)
	if !r.Env.OK || fetched.Totals["ok"] != 2 || fetched.Totals["failed"] != 1 || fetched.Totals["new_items"] != 2 {
		t.Fatalf("fetch totals %+v rows %+v", fetched.Totals, fetched.Sources)
	}
	if len(r.Env.Warnings) == 0 || !strings.Contains(strings.Join(r.Env.Warnings, " "), "broken") {
		t.Fatalf("failed source must surface as a warning: %v", r.Env.Warnings)
	}
	if fetched.Sources[0].Old != 1 {
		t.Errorf("ancient post should be counted as old_skipped: %+v", fetched.Sources[0])
	}

	r = e.run("", "candidates")
	var cands struct {
		Total       int            `json:"total"`
		MutedHidden int            `json:"muted_hidden"`
		Candidates  []candidateRow `json:"candidates"`
		Basis       string         `json:"window_basis"`
	}
	e.data(r, &cands)
	if cands.Total != 1 || cands.MutedHidden != 1 || cands.Basis != BasisDefault {
		t.Fatalf("candidates %+v", cands)
	}
	top := cands.Candidates[0]
	if len(top.SourceIDs) != 2 || top.DiscussionURL == "" {
		t.Fatalf("HN and blog copies must merge into one candidate: %+v", top)
	}

	bad := `{"items":[{"id":"nope","summary":"x"}]}`
	if r := e.run(bad, "digest", "create", "--input", "-"); r.Code != ExitUsage || r.Env.Error.Code != "not_found" {
		t.Fatalf("unknown id: code=%d err=%+v", r.Code, r.Env.Error)
	}
	sub := fmt.Sprintf(`{"items":[{"id":%q,"summary":"Claude 推出新的 agent SDK。"}],"note":"n"}`, top.ID)
	if r := e.run(sub, "digest", "create", "--input", "-", "--dry-run"); !r.Env.OK || strings.Contains(string(r.Env.Data), `"saved":true`) {
		t.Fatalf("dry run: %s", r.Stdout)
	}
	if _, err := store.New(e.home).LoadDigest("latest"); err == nil {
		t.Fatal("dry run must not save")
	}
	if r := e.run(sub, "digest", "create", "--input", "-"); !r.Env.OK {
		t.Fatalf("create: %+v", r.Env.Error)
	}
	if r := e.run(sub, "digest", "create", "--input", "-"); r.Env.Error == nil || r.Env.Error.Code != "conflict" {
		t.Fatalf("second create same day must conflict: %s", r.Stdout)
	}

	e.now = e.now.Add(time.Hour)
	r = e.run("", "candidates")
	e.data(r, &cands)
	if cands.Total != 0 || cands.Basis != BasisLastDigest {
		t.Fatalf("after digest, window continues from it and picked items are gone: %+v", cands)
	}

	r = e.run("", "status")
	var st struct {
		TodayDigest bool             `json:"today_digest_exists"`
		Failing     []map[string]any `json:"failing_sources"`
	}
	e.data(r, &st)
	if !st.TodayDigest || len(st.Failing) != 1 {
		t.Fatalf("status %s", r.Env.Data)
	}

	out := filepath.Join(t.TempDir(), "site")
	if r := e.run("", "digest", "export", "--out", out); !r.Env.OK {
		t.Fatalf("export %+v", r.Env.Error)
	}
	if b, err := os.ReadFile(filepath.Join(out, "2026-11-01.html")); err != nil || !strings.Contains(string(b), "Claude 推出新的 agent SDK") {
		t.Fatalf("exported page missing summary: %v", err)
	}

	if r := e.run("", "read", top.ID); !r.Env.OK {
		t.Fatalf("read by id should work offline-resolvable ids: %+v", r.Env.Error)
	}
	if r := e.run("", "read", e.srv.URL+"/pdf"); r.Env.Error == nil || r.Env.Error.Code != "unsupported_content" {
		t.Fatalf("pdf: %s", r.Stdout)
	}
	if r := e.run("", "read", e.srv.URL+"/article"); !r.Env.OK || !strings.Contains(string(r.Env.Data), `"container":"article"`) {
		t.Fatalf("article: %s", r.Stdout)
	}

	r = e.run("", "prune", "--older-than", "30m", "--dry-run")
	if !r.Env.OK || !strings.Contains(string(r.Env.Data), `"items_deleted":2`) {
		t.Fatalf("prune dry run: %s", r.Env.Data)
	}
	if st, _ := store.New(e.home).LoadState(); len(st.Items) != 2 {
		t.Fatal("dry-run prune must not delete")
	}
}

func TestSourcesAndKeywords(t *testing.T) {
	e := newEnv(t)
	e.seed()
	if r := e.run("", "sources", "add", "--id", "Bad ID", "--url", e.srv.URL+"/rss"); r.Code != ExitUsage {
		t.Fatalf("invalid id: %s", r.Stdout)
	}
	if r := e.run("", "sources", "add", "--id", "dead", "--url", e.srv.URL+"/broken"); r.Env.OK {
		t.Fatal("add must refuse a source that fails the live test")
	}
	if r := e.run("", "sources", "add", "--id", "blog", "--url", e.srv.URL+"/other"); r.Env.Error == nil || r.Env.Error.Code != "conflict" {
		t.Fatalf("duplicate id: %s", r.Stdout)
	}
	if r := e.run("", "sources", "add", "--id", "blog2", "--url", e.srv.URL+"/rss?x=1", "--weight", "1.5", "--followed"); !r.Env.OK {
		t.Fatalf("valid add: %+v", r.Env.Error)
	}
	if r := e.run("", "sources", "disable", "broken"); !r.Env.OK {
		t.Fatal(r.Stdout)
	}
	if r := e.run("", "sources", "rm", "ghost"); r.Env.Error == nil || r.Env.Error.Code != "not_found" {
		t.Fatalf("rm unknown: %s", r.Stdout)
	}
	r := e.run("", "keywords", "add", "--boost", "Rust,claude", "--mute", "nft")
	if !r.Env.OK || len(r.Env.Warnings) == 0 {
		t.Fatalf("existing keyword should produce a warning: %s", r.Stdout)
	}
	w, _ := store.New(e.home).LoadWatchlist()
	if len(w.Sources) != 4 || !w.Sources[2].Disabled || strings.Join(w.Boost, ",") != "claude,rust" || strings.Join(w.Mute, ",") != "crypto,nft" {
		t.Fatalf("watchlist after edits: %+v", w)
	}
	if r := e.run("", "keywords", "rm", "--mute", "CRYPTO"); !r.Env.OK {
		t.Fatal(r.Stdout)
	}
}

func TestUsageAndDiscovery(t *testing.T) {
	e := newEnv(t)
	if r := e.run("", "fetch", "--bogus"); r.Code != ExitUsage || r.Env.Error.Code != "usage" {
		t.Fatalf("bad flag: %d %s", r.Code, r.Stdout)
	}
	if r := e.run("", "candidates", "--since", "yesterday"); r.Code == ExitOK {
		t.Fatal("bad --since must fail")
	}
	r := e.run("", "describe")
	var d struct {
		Commands   []CommandInfo     `json:"commands"`
		ErrorCodes map[string]string `json:"error_codes"`
		Ranking    map[string]string `json:"ranking"`
	}
	e.data(r, &d)
	// 每個有 handler 的指令都要出現在 manifest，避免新增指令卻忘了 describe。
	want := 0
	var count func(c *Command)
	count = func(c *Command) {
		if c.Setup != nil {
			want++
		}
		for _, s := range c.Sub {
			count(s)
		}
	}
	count(Root())
	if len(d.Commands) != want || want < 18 {
		t.Fatalf("describe lists %d commands, tree has %d", len(d.Commands), want)
	}
	if len(d.ErrorCodes) != len(ErrorCodes) || d.Ranking["window"] == "" {
		t.Fatalf("describe incomplete")
	}
	if r := e.run("", "digest", "create", "--help", "--json"); !r.Env.OK || !strings.Contains(string(r.Env.Data), "--input") && !strings.Contains(string(r.Env.Data), `"input"`) {
		t.Fatalf("per-command help: %s", r.Env.Data)
	}
}

// 每個 error code 都要有 classify 之外的實際產生點或明確對應；
// 這裡確認 classify 對已知 sentinel 給出預期的 code。
func TestClassify(t *testing.T) {
	cases := map[error]string{
		store.ErrNoWatchlist:                   "no_data",
		store.ErrNoDigest:                      "no_data",
		fmt.Errorf("x: %w", store.ErrLocked):   "locked",
		&Error{Code: "conflict", Message: "c"}: "conflict",
		fmt.Errorf("plain"):                    "runtime",
	}
	for err, want := range cases {
		if got := classify(err).Code; got != want {
			t.Errorf("%v → %s, want %s", err, got, want)
		}
		if _, ok := ErrorCodes[classify(err).Code]; !ok {
			t.Errorf("classify produced undocumented code %q", classify(err).Code)
		}
	}
}
