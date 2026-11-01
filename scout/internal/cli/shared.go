package cli

import (
	"errors"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
	"github.com/Andy880828/personal-lab/scout/internal/rank"
	"github.com/Andy880828/personal-lab/scout/internal/store"
)

// Window basis 說明候選時間窗的起點從哪來，回報給 agent。
const (
	BasisLastDigest = "last_digest" // 上一份 digest 的時間窗終點
	BasisDefault    = "default_24h" // 還沒有任何 digest
	BasisSinceFlag  = "since_flag"
)

// defaultWindowStart：有 digest 就從上一份的終點接續，否則往前 24 小時。
// 這讓每天的 digest 首尾相接，不重複也不漏。
func defaultWindowStart(st *store.Store, now time.Time) (time.Time, string, error) {
	d, err := st.LoadDigest("latest")
	if errors.Is(err, store.ErrNoDigest) {
		return now.Add(-24 * time.Hour), BasisDefault, nil
	}
	if err != nil {
		return time.Time{}, "", err
	}
	return d.WindowTo, BasisLastDigest, nil
}

type candidateSet struct {
	List         []rank.Scored
	MutedHidden  int
	PickedHidden int
}

// collectCandidates 是候選規則的唯一實作；status、fetch、candidates、
// digest create 都呼叫它，避免各自算出不同的數字。
func collectCandidates(state *core.State, w *core.Watchlist, from, now time.Time, includePicked, includeMuted bool, sourceFilter string) candidateSet {
	srcs := sourceIndex(w)
	var cs candidateSet
	for _, it := range state.Items {
		if !rank.InWindow(it, from) {
			continue
		}
		if sourceFilter != "" && !seenBy(it, sourceFilter) {
			continue
		}
		if it.PickedIn != "" && !includePicked {
			cs.PickedHidden++
			continue
		}
		sc := rank.Score(it, srcs, w.Boost, w.Mute, now)
		if sc.Muted != "" && !includeMuted {
			cs.MutedHidden++
			continue
		}
		cs.List = append(cs.List, sc)
	}
	rank.Sort(cs.List)
	return cs
}

func sourceIndex(w *core.Watchlist) map[string]core.Source {
	m := make(map[string]core.Source, len(w.Sources))
	for _, s := range w.Sources {
		m[s.ID] = s
	}
	return m
}

func seenBy(it *core.Item, sourceID string) bool {
	for _, sg := range it.Sightings {
		if sg.SourceID == sourceID {
			return true
		}
	}
	return false
}

func findSource(w *core.Watchlist, id string) int {
	for i, s := range w.Sources {
		if s.ID == id {
			return i
		}
	}
	return -1
}

func sourceNotFound(id string) *Error {
	return &Error{Code: "not_found", Message: "unknown source " + id, Hint: "list ids with `scout sources list --json`", exit: ExitUsage}
}
