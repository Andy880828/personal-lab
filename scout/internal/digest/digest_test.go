package digest

import (
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
	h := HTMLPage(d)
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
