package digest

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

func TestValidateListsEveryProblem(t *testing.T) {
	items := map[string]*core.Item{"a": {ID: "a"}}
	err := Validate(&Submission{Items: []Pick{
		{ID: "a", Summary: "ok"},
		{ID: "a", Summary: "dup"},
		{ID: "ghost", Summary: ""},
		{ID: "", Summary: strings.Repeat("長", MaxSummaryRunes+1)},
	}}, items)
	ve, ok := IsValidation(err)
	if !ok {
		t.Fatalf("want ValidationError, got %v", err)
	}
	if len(ve.Problems) != 5 || len(ve.UnknownIDs) != 1 || ve.UnknownIDs[0] != "ghost" {
		t.Fatalf("problems=%v unknown=%v", ve.Problems, ve.UnknownIDs)
	}
	if err := Validate(&Submission{}, items); err == nil {
		t.Error("empty submission must fail")
	}
	many := &Submission{}
	for i := 0; i <= MaxItems; i++ {
		many.Items = append(many.Items, Pick{ID: "a", Summary: "x"})
	}
	if err := Validate(many, items); err == nil || !strings.Contains(err.Error(), "maximum") {
		t.Errorf("over max must fail: %v", err)
	}
}

func TestParseSubmissionRejectsUnknownFields(t *testing.T) {
	if _, err := ParseSubmission(strings.NewReader(`{"items":[{"id":"a","sumary":"typo"}]}`)); err == nil {
		t.Fatal("misspelled field must be rejected, not silently ignored")
	}
}

func TestRenderEscapes(t *testing.T) {
	d := &core.Digest{Date: "2026-11-01", WindowFrom: time.Now(), WindowTo: time.Now(), Items: []core.DigestItem{
		{Title: `<script>alert(1)</script> [x]`, URL: `https://e.com/?a="b"`, Summary: "摘要 <b>", Sources: []string{"HN"}},
	}}
	h := HTMLPage(d, "", "", 0)
	if strings.Contains(h, "<script>alert") || strings.Contains(h, `?a="b"`) || strings.Contains(h, "摘要 <b>") {
		t.Fatalf("html output not escaped: %s", h)
	}
	if md := Markdown(d); !strings.Contains(md, `\[x\]`) {
		t.Errorf("markdown link text brackets not escaped: %s", md)
	}
	if idx := HTMLIndex([]*core.Digest{d}); strings.Contains(idx, "<script>alert") {
		t.Errorf("index not escaped")
	}
}

func TestHTMLFrontPage(t *testing.T) {
	mk := func(date string, n int) *core.Digest {
		d := &core.Digest{Date: date, Note: "導讀<b>", CandidatesConsidered: 30}
		for i := 0; i < n; i++ {
			d.Items = append(d.Items, core.DigestItem{Title: fmt.Sprintf("T%d", i), URL: "https://e.example/a", Summary: fmt.Sprintf("摘要%d", i), Why: "理由", Sources: []string{"S"}})
		}
		return d
	}
	// 所有則數的版面結構都要平衡（side／rest 的 div 有開有關）
	for n := 0; n <= 7; n++ {
		h := HTMLPage(mk("2026-11-02", n), "", "", 0)
		if o, c := strings.Count(h, "<div"), strings.Count(h, "</div>"); o != c {
			t.Errorf("n=%d: <div %d != </div> %d", n, o, c)
		}
	}
	idx := HTMLIndex([]*core.Digest{mk("2026-11-02", 6), mk("2026-10-31", 5)})
	for _, want := range []string{"摘要5", "導讀&lt;b&gt;", "第 2 期", "2026 年 11 月 2 日 星期一", "2026 年 10 月", `href="2026-10-31.html"`, "入選理由"} {
		if !strings.Contains(idx, want) {
			t.Errorf("index missing %q", want)
		}
	}
	p := HTMLPage(mk("2026-11-01", 3), "2026-11-02", "2026-10-31", 4)
	if !strings.Contains(p, `href="2026-11-02.html"`) || !strings.Contains(p, `href="2026-10-31.html"`) {
		t.Error("pager links missing")
	}
	bad := mk("2026-11-01", 1)
	bad.Items[0].URL = "javascript:alert(1)"
	bad.Items[0].DiscussionURL = "data:text/html,x"
	if h := HTMLPage(bad, "", "", 0); strings.Contains(h, "javascript:") || strings.Contains(h, "data:text") {
		t.Error("unsafe URL rendered as link")
	}
}
