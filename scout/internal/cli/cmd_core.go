package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
	"github.com/Andy880828/personal-lab/scout/internal/digest"
	"github.com/Andy880828/personal-lab/scout/internal/feed"
	"github.com/Andy880828/personal-lab/scout/internal/store"
)

// staleFetchHours：上次抓取超過這個時數，status 就建議重新抓取。
// 大部分來源一天更新數次，6 小時讓早晚各跑一次就足夠新鮮。
const staleFetchHours = 6.0

// retentionDays：status 在有文章早於此天數時建議 prune（digest 不受影響）。
const retentionDays = 60

func statusCmd() *Command {
	return &Command{
		Name:     "status",
		Summary:  "one-call overview: watchlist, last fetch, failing sources, candidates, last digest, next steps (offline)",
		Examples: []string{"scout status --json"},
		Setup: func(fs *flag.FlagSet) Handler {
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				st, now := rt.Store(), rt.Now()
				data := map[string]any{"home": st.Home}
				var next []string
				w, err := st.LoadWatchlist()
				if errors.Is(err, store.ErrNoWatchlist) {
					data["watchlist"] = nil
					data["next_steps"] = []string{"scout init --json  # create the default watchlist, then review sources with the user"}
					return &Output{Data: data, Human: func(o io.Writer) { fmt.Fprintln(o, "no watchlist yet — run `scout init`") }}, nil
				}
				if err != nil {
					return nil, err
				}
				state, err := st.LoadState()
				if err != nil {
					return nil, err
				}
				enabled := 0
				var failing []map[string]any
				for _, s := range w.Sources {
					if s.Disabled {
						continue
					}
					enabled++
					if ss := state.Sources[s.ID]; ss != nil && ss.ConsecutiveFailures > 0 {
						failing = append(failing, map[string]any{"id": s.ID, "consecutive_failures": ss.ConsecutiveFailures, "last_error": ss.LastError, "last_success": ss.LastSuccess})
					}
				}
				data["watchlist"] = map[string]any{"path": st.WatchlistPath(), "sources": len(w.Sources), "enabled": enabled, "boost": w.Boost, "mute": w.Mute}
				data["failing_sources"] = failing

				lastFetch := map[string]any{"at": state.LastFetch, "age_hours": ageHours(now, state.LastFetch)}
				data["last_fetch"] = lastFetch
				if state.LastFetch == nil || now.Sub(*state.LastFetch).Hours() > staleFetchHours {
					next = append(next, "scout fetch --json  # data is missing or older than 6h")
				}

				from, basis, err := defaultWindowStart(st, now)
				if err != nil {
					return nil, err
				}
				cs := collectCandidates(state, w, from, now, false, false, "")
				oldest := 0
				cutoff := now.AddDate(0, 0, -retentionDays)
				for _, it := range state.Items {
					if it.FirstSeen.Before(cutoff) {
						oldest++
					}
				}
				data["items"] = map[string]any{"total": len(state.Items), "older_than_retention": oldest, "retention_days": retentionDays}
				data["candidates"] = map[string]any{"count": len(cs.List), "window_from": from, "window_basis": basis, "muted_hidden": cs.MutedHidden}

				today := now.Local().Format("2006-01-02")
				var lastDigest any
				hasToday := false
				if d, err := st.LoadDigest("latest"); err == nil {
					lastDigest = map[string]any{"date": d.Date, "items": len(d.Items), "created_at": d.CreatedAt}
					hasToday = d.Date == today
				} else if !errors.Is(err, store.ErrNoDigest) {
					return nil, err
				}
				data["last_digest"] = lastDigest
				data["today"] = today
				data["today_digest_exists"] = hasToday

				if len(cs.List) > 0 && !hasToday {
					next = append(next, "scout candidates --json  # then pick 4-7, summarise, `scout digest create` (SKILLS.md §5)")
				}
				if len(failing) > 0 {
					next = append(next, "scout sources list --json  # some sources keep failing; test or disable them with the user")
				}
				if oldest > 0 {
					next = append(next, fmt.Sprintf("scout prune --older-than %dd --dry-run --json", retentionDays))
				}
				if len(next) == 0 {
					next = []string{"nothing to do; `scout trends --json` for what is heating up"}
				}
				data["next_steps"] = next
				return &Output{Data: data, Human: func(o io.Writer) {
					age := "never"
					if state.LastFetch != nil {
						age = fmt.Sprintf("%.1fh ago", now.Sub(*state.LastFetch).Hours())
					}
					fmt.Fprintf(o, "sources %d/%d enabled, %d failing · last fetch %s · %d items · %d candidates · today's digest: %v\n",
						enabled, len(w.Sources), len(failing), age, len(state.Items), len(cs.List), hasToday)
					for _, n := range next {
						fmt.Fprintln(o, "next:", n)
					}
				}}, nil
			}
		},
	}
}

func initCmd() *Command {
	return &Command{
		Name:    "init",
		Summary: "create the default watchlist in SCOUT_HOME (refuses to overwrite)",
		Setup: func(fs *flag.FlagSet) Handler {
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				st := rt.Store()
				unlock, err := st.Lock()
				if err != nil {
					return nil, err
				}
				defer unlock()
				if _, err := st.LoadWatchlist(); err == nil {
					return nil, &Error{Code: "conflict", Message: "watchlist already exists at " + st.WatchlistPath(), Hint: "edit it with `scout sources` / `scout keywords`; delete the file manually only if the user wants a reset", exit: ExitUsage}
				} else if !errors.Is(err, store.ErrNoWatchlist) {
					return nil, err
				}
				w := store.DefaultWatchlist()
				if err := st.SaveWatchlist(w); err != nil {
					return nil, err
				}
				enabled := 0
				for _, s := range w.Sources {
					if !s.Disabled {
						enabled++
					}
				}
				data := map[string]any{"path": st.WatchlistPath(), "sources": len(w.Sources), "enabled": enabled, "boost": w.Boost,
					"next_steps": []string{"scout sources list --json  # review sources and keywords with the user", "scout fetch --json"}}
				return &Output{Data: data, Human: func(o io.Writer) {
					fmt.Fprintf(o, "created %s with %d sources (%d enabled)\n", st.WatchlistPath(), len(w.Sources), enabled)
				}}, nil
			}
		},
	}
}

type fetchRow struct {
	ID         string           `json:"id"`
	Status     core.FetchStatus `json:"status"`
	Format     string           `json:"format,omitempty"`
	Entries    int              `json:"entries"`
	New        int              `json:"new"`
	Seen       int              `json:"seen"`
	Skipped    int              `json:"skipped,omitempty"`
	Invalid    int              `json:"invalid,omitempty"`
	Old        int              `json:"old_skipped,omitempty"`
	Error      string           `json:"error,omitempty"`
	DurationMS int64            `json:"duration_ms"`
}

func fetchCmd() *Command {
	return &Command{
		Name:     "fetch",
		Summary:  "fetch all enabled sources (conditional GET), dedupe into the local store, report per-source health",
		Examples: []string{"scout fetch --json", "scout fetch --source hn,simonwillison --json"},
		Setup: func(fs *flag.FlagSet) Handler {
			only := fs.String("source", "", "comma-separated source ids (may include disabled ones)")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				st := rt.Store()
				unlock, err := st.Lock()
				if err != nil {
					return nil, err
				}
				defer unlock()
				w, err := st.LoadWatchlist()
				if err != nil {
					return nil, err
				}
				state, err := st.LoadState()
				if err != nil {
					return nil, err
				}
				var srcs []core.Source
				var warnings []string
				if ids := csv(*only); len(ids) > 0 {
					for _, id := range ids {
						i := findSource(w, id)
						if i < 0 {
							return nil, sourceNotFound(id)
						}
						srcs = append(srcs, w.Sources[i])
					}
				} else {
					for _, s := range w.Sources {
						if !s.Disabled {
							srcs = append(srcs, s)
						}
					}
				}
				if len(srcs) == 0 {
					return nil, usageError("enable a source with `scout sources enable <id>`", "no enabled sources to fetch")
				}

				results := rt.Client.FetchAll(ctx, srcs, state.Sources)
				now := rt.Now()
				rows := make([]fetchRow, 0, len(results))
				totals := map[string]int{"sources": len(results), "ok": 0, "not_modified": 0, "failed": 0, "new_items": 0}
				for _, r := range results {
					ss := state.Sources[r.Source.ID]
					if ss == nil {
						ss = &core.SourceState{}
						state.Sources[r.Source.ID] = ss
					}
					ss.LastAttempt, ss.LastStatus = now, r.Status
					row := fetchRow{ID: r.Source.ID, Status: r.Status, Format: r.Format, Entries: len(r.Entries), Skipped: r.Skipped, DurationMS: r.Duration.Milliseconds()}
					switch r.Status {
					case core.FetchError:
						totals["failed"]++
						ss.ConsecutiveFailures++
						ss.LastError = r.Err.Error()
						row.Error = r.Err.Error()
						warnings = append(warnings, fmt.Sprintf("source %s failed (%d in a row): %v", r.Source.ID, ss.ConsecutiveFailures, r.Err))
					default:
						if r.Status == core.FetchOK {
							totals["ok"]++
							ms := feed.Merge(state, r.Source, r.Entries, now)
							row.New, row.Seen, row.Invalid, row.Old = ms.New, ms.Seen, ms.Invalid, ms.Old
							totals["new_items"] += ms.New
							ss.LastItemCount = len(r.Entries)
							if ms.Invalid > 0 {
								warnings = append(warnings, fmt.Sprintf("source %s: %d entries had unusable URLs and were skipped", r.Source.ID, ms.Invalid))
							}
						} else {
							totals["not_modified"]++
						}
						t := now
						ss.LastSuccess, ss.ConsecutiveFailures, ss.LastError = &t, 0, ""
						ss.ETag, ss.LastModified = r.ETag, r.LastModified
					}
					if r.Skipped > 0 {
						warnings = append(warnings, fmt.Sprintf("source %s: %d entries without a link were skipped", r.Source.ID, r.Skipped))
					}
					rows = append(rows, row)
				}
				state.LastFetch = &now
				if err := st.SaveState(state); err != nil {
					return nil, err
				}
				from, basis, err := defaultWindowStart(st, now)
				if err != nil {
					return nil, err
				}
				cs := collectCandidates(state, w, from, now, false, false, "")
				data := map[string]any{
					"fetched_at": now, "sources": rows, "totals": totals,
					"candidates": map[string]any{"count": len(cs.List), "window_from": from, "window_basis": basis},
					"next_steps": []string{"scout candidates --json"},
				}
				return &Output{Data: data, Warnings: warnings, Human: func(o io.Writer) {
					for _, r := range rows {
						extra := ""
						if r.Error != "" {
							extra = " — " + r.Error
						}
						old := ""
						if r.Old > 0 {
							old = fmt.Sprintf(" (%d older than %dd skipped)", r.Old, feed.MaxEntryAgeDays)
						}
						fmt.Fprintf(o, "%-22s %-12s %4d entries %3d new%s%s\n", r.ID, r.Status, r.Entries, r.New, old, extra)
					}
					fmt.Fprintf(o, "\n%d ok, %d not modified, %d failed · %d new items · %d candidates\n",
						totals["ok"], totals["not_modified"], totals["failed"], totals["new_items"], len(cs.List))
				}}, nil
			}
		},
	}
}

type candidateRow struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	URL           string     `json:"url"`
	Score         float64    `json:"score"`
	Reasons       []string   `json:"reasons"`
	AgeHours      float64    `json:"age_hours"`
	Published     *time.Time `json:"published,omitempty"`
	FirstSeen     time.Time  `json:"first_seen"`
	Sources       []string   `json:"sources"`
	SourceIDs     []string   `json:"source_ids"`
	DiscussionURL string     `json:"discussion_url,omitempty"`
	Author        string     `json:"author,omitempty"`
	Excerpt       string     `json:"excerpt,omitempty"`
	Muted         string     `json:"muted_by,omitempty"`
	PickedIn      string     `json:"picked_in,omitempty"`
}

func candidatesCmd() *Command {
	return &Command{
		Name:    "candidates",
		Summary: "ranked, unpicked items in the current window, each score with its reasons (offline)",
		Examples: []string{
			"scout candidates --json",
			"scout candidates --since 3d --limit 50 --json",
			"scout candidates --source hn --json",
		},
		Setup: func(fs *flag.FlagSet) Handler {
			since := fs.String("since", "", "window start as a duration back from now (36h, 3d); default: end of the last digest, or 24h")
			limit := fs.Int("limit", 20, "max rows (0 = all)")
			source := fs.String("source", "", "only items listed by this source id")
			incPicked := fs.Bool("include-picked", false, "include items already picked into a digest")
			incMuted := fs.Bool("include-muted", false, "include items matching mute keywords")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				st, now := rt.Store(), rt.Now()
				w, err := st.LoadWatchlist()
				if err != nil {
					return nil, err
				}
				state, err := st.LoadState()
				if err != nil {
					return nil, err
				}
				if state.LastFetch == nil {
					return nil, &Error{Code: "no_data", Message: "nothing fetched yet", Hint: "run `scout fetch --json`"}
				}
				if *source != "" && findSource(w, *source) < 0 {
					return nil, sourceNotFound(*source)
				}
				var from time.Time
				basis := BasisSinceFlag
				if *since != "" {
					d, err := parseDuration(*since)
					if err != nil {
						return nil, usageError("examples: --since 36h, --since 3d", "--since: %v", err)
					}
					from = now.Add(-d)
				} else if from, basis, err = defaultWindowStart(st, now); err != nil {
					return nil, err
				}
				cs := collectCandidates(state, w, from, now, *incPicked, *incMuted, *source)
				srcs := sourceIndex(w)
				total := len(cs.List)
				list := cs.List
				if *limit > 0 && len(list) > *limit {
					list = list[:*limit]
				}
				rows := make([]candidateRow, 0, len(list))
				for _, sc := range list {
					it := sc.Item
					var ids []string
					for _, sg := range it.Sightings {
						ids = append(ids, sg.SourceID)
					}
					sort.Strings(ids)
					rows = append(rows, candidateRow{
						ID: it.ID, Title: it.Title, URL: it.URL, Score: sc.Score, Reasons: sc.Reasons, AgeHours: sc.AgeHours,
						Published: it.Published, FirstSeen: it.FirstSeen, Sources: digest.SourceNames(it, srcs), SourceIDs: ids,
						DiscussionURL: digest.DiscussionURL(it), Author: it.Author, Excerpt: trimRunes(it.Excerpt, 300),
						Muted: sc.Muted, PickedIn: it.PickedIn,
					})
				}
				var warnings []string
				if age := now.Sub(*state.LastFetch).Hours(); age > staleFetchHours {
					warnings = append(warnings, fmt.Sprintf("last fetch was %.1fh ago; run `scout fetch --json` for fresh items", age))
				}
				data := map[string]any{
					"window_from": from, "window_to": now, "window_basis": basis,
					"total": total, "returned": len(rows), "muted_hidden": cs.MutedHidden, "picked_hidden": cs.PickedHidden,
					"candidates": rows,
				}
				return &Output{Data: data, Warnings: warnings, Human: func(o io.Writer) {
					for _, r := range rows {
						fmt.Fprintf(o, "%5.1f  %s  %s\n       %s · %s\n", r.Score, r.ID, r.Title, strings.Join(r.Sources, ", "), r.URL)
					}
					fmt.Fprintf(o, "\n%d of %d candidates since %s (%s); %d muted, %d already picked hidden\n",
						len(rows), total, from.Local().Format("01/02 15:04"), basis, cs.MutedHidden, cs.PickedHidden)
				}}, nil
			}
		},
	}
}

func trimRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
