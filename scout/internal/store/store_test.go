package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

func TestLockIsExclusiveAndStaleLocksExpire(t *testing.T) {
	s := New(t.TempDir())
	unlock, err := s.Lock()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Lock(); !errors.Is(err, ErrLocked) {
		t.Fatalf("second lock: want ErrLocked, got %v", err)
	}
	unlock()
	unlock2, err := s.Lock()
	if err != nil {
		t.Fatalf("relock after unlock: %v", err)
	}
	unlock2()

	p := filepath.Join(s.Home, "scout.lock")
	os.WriteFile(p, []byte("pid=1"), 0o644)
	old := time.Now().Add(-lockStaleAfter - time.Minute)
	os.Chtimes(p, old, old)
	unlock3, err := s.Lock()
	if err != nil {
		t.Fatalf("stale lock should be taken over: %v", err)
	}
	unlock3()
}

func TestRoundTripAndDigests(t *testing.T) {
	s := New(t.TempDir())
	if _, err := s.LoadWatchlist(); !errors.Is(err, ErrNoWatchlist) {
		t.Fatalf("want ErrNoWatchlist, got %v", err)
	}
	st, err := s.LoadState()
	if err != nil || st.Items == nil || st.Sources == nil {
		t.Fatalf("empty state must be usable: %+v %v", st, err)
	}
	w := DefaultWatchlist()
	if err := s.SaveWatchlist(w); err != nil {
		t.Fatal(err)
	}
	got, _ := s.LoadWatchlist()
	if len(got.Sources) != len(w.Sources) {
		t.Fatalf("watchlist round trip lost sources")
	}
	if _, err := s.LoadDigest("latest"); !errors.Is(err, ErrNoDigest) {
		t.Fatalf("want ErrNoDigest, got %v", err)
	}
	for _, d := range []string{"2026-10-30", "2026-11-01", "2026-10-31"} {
		if err := s.SaveDigest(&core.Digest{Date: d}); err != nil {
			t.Fatal(err)
		}
	}
	latest, err := s.LoadDigest("latest")
	if err != nil || latest.Date != "2026-11-01" {
		t.Fatalf("latest=%v err=%v", latest, err)
	}
	if _, err := s.LoadDigest("../etc/passwd"); err == nil {
		t.Fatal("path-like date must be rejected")
	}
	if err := s.SaveDigest(&core.Digest{Date: "bad"}); err == nil {
		t.Fatal("invalid date must not be saved")
	}
}

func TestDefaultWatchlistIsValid(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range DefaultWatchlist().Sources {
		if seen[s.ID] || s.URL == "" || s.Weight <= 0 {
			t.Errorf("bad default source %+v", s)
		}
		if _, ok := core.SourceKinds[s.Kind]; !ok {
			t.Errorf("unknown kind in %s", s.ID)
		}
		seen[s.ID] = true
	}
}
