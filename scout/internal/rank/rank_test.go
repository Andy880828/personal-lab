package rank

import (
	"strings"
	"testing"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

func TestMatch(t *testing.T) {
	cases := []struct {
		text, kw string
		want     bool
	}{
		{"new ai model", "ai", true},
		{"she said hello", "ai", false},
		{"ai-powered tool", "ai", true},
		{"open source llm", "open source", true},
		{"claude code 2.0", "Claude", true},
		{"台灣 ai 新創", "台灣", true},
		{"mcp", "mcp", true},
		{"", "x", false},
		{"anything", "  ", false},
	}
	for _, c := range cases {
		if got := Match(c.text, c.kw); got != c.want {
			t.Errorf("Match(%q,%q)=%v want %v", c.text, c.kw, got, c.want)
		}
	}
}

func TestScoreComponents(t *testing.T) {
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	pub := now
	it := &core.Item{ID: "x", Title: "Claude agents", FirstSeen: now, Published: &pub, Sightings: []core.Sighting{
		{SourceID: "simon"}, {SourceID: "hn", Points: 99, Comments: 9},
	}}
	srcs := map[string]core.Source{"simon": {ID: "simon", Weight: 1.5, Followed: true}, "hn": {ID: "hn", Weight: 1.2}}
	s := Score(it, srcs, []string{"claude", "agent"}, nil, now)
	// 15 (weight) + 15 (followed) + 8 (coverage) + 5*log10(100)=10 + 2*log10(10)=2 + 6 (claude; "agent" ≠ "agents") = 56
	if s.Score != 56 {
		t.Fatalf("score=%v reasons=%v", s.Score, s.Reasons)
	}
	joined := strings.Join(s.Reasons, "|")
	for _, want := range []string{"followed author via simon", "covered by 2 sources", "hn 99 points", "keywords claude"} {
		if !strings.Contains(joined, want) {
			t.Errorf("reasons missing %q: %v", want, s.Reasons)
		}
	}

	old := now.Add(-time.Duration(HalfLifeHours) * time.Hour)
	it.Published = &old
	if s2 := Score(it, srcs, []string{"claude"}, nil, now); s2.Score != 28 {
		t.Errorf("one half-life should halve the score: %v", s2.Score)
	}
	if s3 := Score(it, srcs, nil, []string{"Claude"}, now); s3.Muted != "Claude" {
		t.Errorf("mute keyword not detected")
	}
}

func TestBoostCapped(t *testing.T) {
	now := time.Now()
	it := &core.Item{Title: "a b c d e", FirstSeen: now, Sightings: []core.Sighting{{SourceID: "s"}}}
	s := Score(it, map[string]core.Source{"s": {Weight: 1}}, []string{"a", "b", "c", "d", "e"}, nil, now)
	if s.Score != SourceWeightPoints+BoostCap {
		t.Fatalf("boost must cap at %v: %v", BoostCap, s.Score)
	}
}

func TestInWindow(t *testing.T) {
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	before := func(h float64) *time.Time { t := from.Add(-time.Duration(h * float64(time.Hour))); return &t }
	cases := []struct {
		name string
		it   core.Item
		want bool
	}{
		{"seen before window", core.Item{FirstSeen: from.Add(-time.Minute)}, false},
		{"seen in window, unknown publish", core.Item{FirstSeen: from.Add(time.Hour)}, true},
		{"late discovery within 72h", core.Item{FirstSeen: from.Add(time.Hour), Published: before(48)}, true},
		{"old article on first fetch", core.Item{FirstSeen: from.Add(time.Hour), Published: before(LateDiscoveryHours + 1)}, false},
	}
	for _, c := range cases {
		if got := InWindow(&c.it, from); got != c.want {
			t.Errorf("%s: got %v", c.name, got)
		}
	}
}
