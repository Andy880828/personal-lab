package trends

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

func TestTerms(t *testing.T) {
	got := Terms("Introducing Claude Code: the new GPT-5 and o3 models for AI agents")
	for _, want := range []string{"claude", "code", "claude code", "gpt-5", "o3", "models", "agents"} {
		if !slices.Contains(got, want) {
			t.Errorf("missing %q in %v", want, got)
		}
	}
	for _, bad := range []string{"the", "new", "introducing", "and", "for", "ai"} {
		if slices.Contains(got, bad) {
			t.Errorf("stopword/short %q should be dropped: %v", bad, got)
		}
	}
}

func TestCompute(t *testing.T) {
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	items := map[string]*core.Item{}
	add := func(id, title string, daysAgo int) {
		p := now.AddDate(0, 0, -daysAgo).Add(-time.Hour)
		items[id] = &core.Item{ID: id, Title: title, Published: &p, FirstSeen: now}
	}
	for i := 0; i < 4; i++ {
		add(fmt.Sprintf("c%d", i), "Rust compiler news", 1)
	}
	add("p1", "Rust compiler news", 9)
	add("p2", "Kubernetes release", 9)
	add("p3", "Kubernetes release", 10)
	add("z", "台灣 新聞", 1)
	add("only1", "Singleton topic", 1)

	r := Compute(items, 7, 10, now)
	if r.ItemsCurrent != 5 || r.ItemsPrev != 3 || r.SkippedCJK != 1 {
		t.Fatalf("counts cur=%d prev=%d cjk=%d", r.ItemsCurrent, r.ItemsPrev, r.SkippedCJK)
	}
	if len(r.Rising) == 0 || r.Rising[0].Delta != 3 {
		t.Fatalf("rising %+v", r.Rising)
	}
	for _, term := range r.Top {
		if term.Term == "singleton" || term.Term == "kubernetes" {
			t.Errorf("%q should not appear (count below %d in current window)", term.Term, MinCount)
		}
	}
	if len(r.Rising[0].Examples) > 3 {
		t.Errorf("examples capped at 3")
	}
}
