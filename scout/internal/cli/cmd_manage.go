package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/Andy880828/personal-lab/scout/internal/core"
	"github.com/Andy880828/personal-lab/scout/internal/digest"
	"github.com/Andy880828/personal-lab/scout/internal/feed"
	"github.com/Andy880828/personal-lab/scout/internal/rank"
	"github.com/Andy880828/personal-lab/scout/internal/trends"
	"github.com/Andy880828/personal-lab/scout/internal/version"
)

var sourceIDRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)

func trendsCmd() *Command {
	return &Command{
		Name:     "trends",
		Summary:  "terms rising in titles: this window vs the previous one of equal length (offline)",
		Examples: []string{"scout trends --json", "scout trends --days 30 --limit 30 --json"},
		Setup: func(fs *flag.FlagSet) Handler {
			days := fs.Int("days", 7, "window length in days")
			limit := fs.Int("limit", 20, "max terms per list (0 = all)")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				if *days <= 0 {
					return nil, usageError("use a positive number of days", "invalid --days %d", *days)
				}
				state, err := rt.Store().LoadState()
				if err != nil {
					return nil, err
				}
				if len(state.Items) == 0 {
					return nil, &Error{Code: "no_data", Message: "no items stored yet", Hint: "run `scout fetch --json` (trends need about 2×--days of history)"}
				}
				rep := trends.Compute(state.Items, *days, *limit, rt.Now())
				var warnings []string
				if rep.ItemsPrev == 0 {
					warnings = append(warnings, fmt.Sprintf("no items in the previous %d-day window, so every term looks new; trends become meaningful after %d days of fetching", *days, 2**days))
				}
				if rep.SkippedCJK > 0 {
					warnings = append(warnings, fmt.Sprintf("%d titles containing CJK text were not tokenised (no word boundaries); read them via `scout candidates`", rep.SkippedCJK))
				}
				return &Output{Data: rep, Warnings: warnings, Human: func(o io.Writer) {
					fmt.Fprintf(o, "rising (last %d days vs previous):\n", *days)
					for _, t := range rep.Rising {
						fmt.Fprintf(o, "  %-28s %3d (prev %d)\n", t.Term, t.Current, t.Previous)
					}
				}}, nil
			}
		},
	}
}

func sourcesListCmd() *Command {
	return &Command{
		Name:    "list",
		Summary: "list sources with their fetch health",
		Setup: func(fs *flag.FlagSet) Handler {
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				st := rt.Store()
				w, err := st.LoadWatchlist()
				if err != nil {
					return nil, err
				}
				state, err := st.LoadState()
				if err != nil {
					return nil, err
				}
				counts := map[string]int{}
				for _, it := range state.Items {
					for _, sg := range it.Sightings {
						counts[sg.SourceID]++
					}
				}
				rows := make([]map[string]any, 0, len(w.Sources))
				for _, s := range w.Sources {
					rows = append(rows, map[string]any{"source": s, "health": state.Sources[s.ID], "items_in_store": counts[s.ID]})
				}
				return &Output{Data: map[string]any{"path": st.WatchlistPath(), "sources": rows, "boost": w.Boost, "mute": w.Mute}, Human: func(o io.Writer) {
					for _, s := range w.Sources {
						flag := ""
						if s.Disabled {
							flag = " (disabled)"
						}
						health := "never fetched"
						if h := state.Sources[s.ID]; h != nil {
							health = string(h.LastStatus)
							if h.ConsecutiveFailures > 0 {
								health = fmt.Sprintf("failing ×%d: %s", h.ConsecutiveFailures, h.LastError)
							}
						}
						fmt.Fprintf(o, "%-22s w=%.1f %-5s %s%s\n", s.ID, s.Weight, s.Kind, health, flag)
					}
				}}, nil
			}
		},
	}
}

// testSource 抓一次並解析，不寫入任何東西；add 與 test 共用。
func testSource(ctx context.Context, rt *Runtime, s core.Source) (map[string]any, error) {
	r := rt.Client.FetchOne(ctx, s, nil)
	if r.Err != nil {
		return nil, &Error{Code: "runtime", Message: fmt.Sprintf("source test failed for %s: %v", s.URL, r.Err),
			Hint: "check the URL (many sites expose /feed, /rss, /atom.xml or <link rel=alternate> in their HTML); use --skip-verify only if the user insists"}
	}
	var sample []map[string]any
	for i, e := range r.Entries {
		if i == 5 {
			break
		}
		sample = append(sample, map[string]any{"title": e.Title, "url": e.URL, "published": e.Published})
	}
	return map[string]any{"format": r.Format, "entries": len(r.Entries), "skipped_without_link": r.Skipped, "sample": sample}, nil
}

func parseKind(k string) (core.SourceKind, error) {
	kind := core.SourceKind(k)
	if _, ok := core.SourceKinds[kind]; !ok {
		return "", usageError("valid kinds: feed, hn", "invalid --kind %q", k)
	}
	return kind, nil
}

func sourcesAddCmd() *Command {
	return &Command{
		Name:     "add",
		Summary:  "add a source after a live test fetch (refuses duplicates)",
		Examples: []string{`scout sources add --id simonw --name "Simon Willison" --url https://simonwillison.net/atom/everything/ --weight 1.5 --followed --json`},
		Setup: func(fs *flag.FlagSet) Handler {
			id := fs.String("id", "", "short id: lowercase letters, digits, dashes (required)")
			name := fs.String("name", "", "display name (default: id)")
			kind := fs.String("kind", "feed", "feed or hn")
			url := fs.String("url", "", "feed or API URL (required)")
			weight := fs.Float64("weight", 1, "importance; 1 is normal, 1.5 strong, 0.5 weak")
			followed := fs.Bool("followed", false, "a person you follow (ranking bonus)")
			tags := fs.String("tags", "", "comma-separated tags")
			skip := fs.Bool("skip-verify", false, "add without the live test fetch")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				if !sourceIDRe.MatchString(*id) {
					return nil, usageError("example: --id simonw", "invalid or missing --id %q", *id)
				}
				k, err := parseKind(*kind)
				if err != nil {
					return nil, err
				}
				if _, _, ok := feed.Canonical(*url); !ok {
					return nil, usageError("pass an http(s) URL", "invalid or missing --url %q", *url)
				}
				if *weight <= 0 || *weight > 5 {
					return nil, usageError("use a weight between 0 and 5", "invalid --weight %v", *weight)
				}
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
				if findSource(w, *id) >= 0 {
					return nil, &Error{Code: "conflict", Message: "source " + *id + " already exists", Hint: "choose another id, or `scout sources rm` it first", exit: ExitUsage}
				}
				for _, s := range w.Sources {
					if s.URL == *url {
						return nil, &Error{Code: "conflict", Message: "URL already used by source " + s.ID, exit: ExitUsage}
					}
				}
				nm := *name
				if nm == "" {
					nm = *id
				}
				src := core.Source{ID: *id, Name: nm, Kind: k, URL: *url, Weight: *weight, Followed: *followed, Tags: csv(*tags)}
				var test map[string]any
				if !*skip {
					if test, err = testSource(ctx, rt, src); err != nil {
						return nil, err
					}
				}
				w.Sources = append(w.Sources, src)
				if err := st.SaveWatchlist(w); err != nil {
					return nil, err
				}
				return &Output{Data: map[string]any{"added": src, "test": test, "verified": !*skip}, Human: func(o io.Writer) {
					fmt.Fprintf(o, "added %s (%s)\n", src.ID, src.URL)
				}}, nil
			}
		},
	}
}

func sourcesRmCmd() *Command {
	return &Command{
		Name:    "rm",
		Summary: "remove a source from the watchlist (stored items stay until prune)",
		Args:    "<source-id>",
		Setup: func(fs *flag.FlagSet) Handler {
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				if len(args) != 1 {
					return nil, usageError("pass one source id", "rm needs one argument")
				}
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
				i := findSource(w, args[0])
				if i < 0 {
					return nil, sourceNotFound(args[0])
				}
				removed := w.Sources[i]
				w.Sources = slices.Delete(w.Sources, i, i+1)
				if err := st.SaveWatchlist(w); err != nil {
					return nil, err
				}
				return &Output{Data: map[string]any{"removed": removed}, Human: func(o io.Writer) { fmt.Fprintln(o, "removed", removed.ID) }}, nil
			}
		},
	}
}

func sourcesToggleCmd(enable bool) *Command {
	name, summary := "disable", "stop fetching a source (keeps it in the watchlist)"
	if enable {
		name, summary = "enable", "resume fetching a source"
	}
	return &Command{
		Name:    name,
		Summary: summary,
		Args:    "<source-id>",
		Setup: func(fs *flag.FlagSet) Handler {
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				if len(args) != 1 {
					return nil, usageError("pass one source id", "%s needs one argument", name)
				}
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
				i := findSource(w, args[0])
				if i < 0 {
					return nil, sourceNotFound(args[0])
				}
				changed := w.Sources[i].Disabled == enable
				w.Sources[i].Disabled = !enable
				if err := st.SaveWatchlist(w); err != nil {
					return nil, err
				}
				return &Output{Data: map[string]any{"source": w.Sources[i], "changed": changed}, Human: func(o io.Writer) {
					fmt.Fprintf(o, "%s %sd (changed: %v)\n", args[0], name, changed)
				}}, nil
			}
		},
	}
}

func sourcesTestCmd() *Command {
	return &Command{
		Name:     "test",
		Summary:  "fetch and parse a URL without saving anything (validate a feed before adding it)",
		Examples: []string{"scout sources test --url https://example.com/feed.xml --json"},
		Setup: func(fs *flag.FlagSet) Handler {
			url := fs.String("url", "", "URL to test (required)")
			kind := fs.String("kind", "feed", "feed or hn")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				k, err := parseKind(*kind)
				if err != nil {
					return nil, err
				}
				if _, _, ok := feed.Canonical(*url); !ok {
					return nil, usageError("pass an http(s) URL", "invalid or missing --url %q", *url)
				}
				res, err := testSource(ctx, rt, core.Source{ID: "test", Kind: k, URL: *url, Weight: 1})
				if err != nil {
					return nil, err
				}
				return &Output{Data: res, Human: func(o io.Writer) {
					fmt.Fprintf(o, "%s: %v entries\n", res["format"], res["entries"])
				}}, nil
			}
		},
	}
}

func keywordsListCmd() *Command {
	return &Command{
		Name:    "list",
		Summary: "show boost and mute keywords",
		Setup: func(fs *flag.FlagSet) Handler {
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				w, err := rt.Store().LoadWatchlist()
				if err != nil {
					return nil, err
				}
				return &Output{Data: map[string]any{"boost": w.Boost, "mute": w.Mute, "matching": rank.Rules()["keywords"]}, Human: func(o io.Writer) {
					fmt.Fprintf(o, "boost: %s\nmute:  %s\n", strings.Join(w.Boost, ", "), strings.Join(w.Mute, ", "))
				}}, nil
			}
		},
	}
}

func keywordsEditCmd(add bool) *Command {
	name, summary := "rm", "remove boost and/or mute keywords"
	if add {
		name, summary = "add", "add boost and/or mute keywords (case-insensitive, deduplicated)"
	}
	return &Command{
		Name:     name,
		Summary:  summary,
		Examples: []string{fmt.Sprintf("scout keywords %s --boost rust,wasm --mute crypto --json", name)},
		Setup: func(fs *flag.FlagSet) Handler {
			boost := fs.String("boost", "", "comma-separated boost keywords")
			mute := fs.String("mute", "", "comma-separated mute keywords")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				b, m := csv(*boost), csv(*mute)
				if len(b) == 0 && len(m) == 0 {
					return nil, usageError("pass --boost and/or --mute", "no keywords given")
				}
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
				var bc, mc []string
				if add {
					w.Boost, bc = addKeywords(w.Boost, b)
					w.Mute, mc = addKeywords(w.Mute, m)
				} else {
					w.Boost, bc = rmKeywords(w.Boost, b)
					w.Mute, mc = rmKeywords(w.Mute, m)
				}
				if err := st.SaveWatchlist(w); err != nil {
					return nil, err
				}
				data := map[string]any{"boost": w.Boost, "mute": w.Mute, "changed_boost": bc, "changed_mute": mc}
				var warnings []string
				if len(bc)+len(mc) < len(b)+len(m) {
					warnings = append(warnings, "some keywords were already present/absent and were left unchanged")
				}
				return &Output{Data: data, Warnings: warnings, Human: func(o io.Writer) {
					fmt.Fprintf(o, "boost: %s\nmute:  %s\n", strings.Join(w.Boost, ", "), strings.Join(w.Mute, ", "))
				}}, nil
			}
		},
	}
}

func addKeywords(list, add []string) ([]string, []string) {
	changed := []string{}
	for _, k := range add {
		k = strings.ToLower(k)
		if !slices.Contains(list, k) {
			list = append(list, k)
			changed = append(changed, k)
		}
	}
	return list, changed
}

func rmKeywords(list, rm []string) ([]string, []string) {
	changed := []string{}
	out := []string{}
	for _, k := range list {
		if slices.ContainsFunc(rm, func(r string) bool { return strings.EqualFold(r, k) }) {
			changed = append(changed, k)
			continue
		}
		out = append(out, k)
	}
	return out, changed
}

func pruneCmd() *Command {
	return &Command{
		Name:     "prune",
		Summary:  "delete stored items first seen before a cutoff (digests are never touched)",
		Examples: []string{"scout prune --older-than 60d --dry-run --json"},
		Setup: func(fs *flag.FlagSet) Handler {
			older := fs.String("older-than", "60d", "age cutoff (e.g. 60d, 720h)")
			dry := fs.Bool("dry-run", false, "report what would be deleted without deleting")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				d, err := parseDuration(*older)
				if err != nil {
					return nil, usageError("examples: 60d, 720h", "--older-than: %v", err)
				}
				st := rt.Store()
				if !*dry {
					unlock, err := st.Lock()
					if err != nil {
						return nil, err
					}
					defer unlock()
				}
				state, err := st.LoadState()
				if err != nil {
					return nil, err
				}
				w, err := st.LoadWatchlist()
				if err != nil {
					return nil, err
				}
				cutoff := rt.Now().Add(-d)
				var items, picked int
				for id, it := range state.Items {
					if it.FirstSeen.Before(cutoff) {
						items++
						if it.PickedIn != "" {
							picked++
						}
						if !*dry {
							delete(state.Items, id)
						}
					}
				}
				orphans := []string{}
				for id := range state.Sources {
					if findSource(w, id) < 0 {
						orphans = append(orphans, id)
						if !*dry {
							delete(state.Sources, id)
						}
					}
				}
				slices.Sort(orphans)
				remaining := len(state.Items)
				if *dry {
					remaining -= items
				}
				if !*dry {
					if err := st.SaveState(state); err != nil {
						return nil, err
					}
				}
				data := map[string]any{"dry_run": *dry, "cutoff": cutoff, "items_deleted": items, "of_which_picked": picked,
					"orphan_source_states": orphans, "items_remaining": remaining}
				return &Output{Data: data, Human: func(o io.Writer) {
					verb := "deleted"
					if *dry {
						verb = "would delete"
					}
					fmt.Fprintf(o, "%s %d items first seen before %s (%d were in digests; digests are kept)\n", verb, items, cutoff.Local().Format("2006-01-02"), picked)
				}}, nil
			}
		},
	}
}

func versionCmd() *Command {
	return &Command{
		Name:    "version",
		Summary: "print version information",
		Setup: func(fs *flag.FlagSet) Handler {
			return func(context.Context, *Runtime, []string) (*Output, error) {
				data := map[string]string{"version": version.Version, "commit": version.Commit, "go": runtime.Version(), "os": runtime.GOOS, "arch": runtime.GOARCH}
				return &Output{Data: data, Human: func(w io.Writer) {
					fmt.Fprintf(w, "scout %s (%s) %s %s/%s\n", version.Version, version.Commit, runtime.Version(), runtime.GOOS, runtime.GOARCH)
				}}, nil
			}
		},
	}
}

func describeCmd() *Command {
	return &Command{
		Name:    "describe",
		Summary: "machine-readable manifest: commands, flags, envelope, exit/error codes, ranking rules, env vars",
		Setup: func(fs *flag.FlagSet) Handler {
			return func(context.Context, *Runtime, []string) (*Output, error) {
				var cmds []CommandInfo
				var walk func(c *Command, path []string)
				walk = func(c *Command, path []string) {
					if c.Setup != nil {
						cmds = append(cmds, info(c, path))
					}
					for _, s := range c.Sub {
						walk(s, append(append([]string{}, path...), s.Name))
					}
				}
				walk(Root(), nil)
				kinds := map[string]string{}
				for k, v := range core.SourceKinds {
					kinds[string(k)] = v
				}
				data := map[string]any{
					"name": "scout", "version": version.Version, "schema": core.SchemaVersion,
					"envelope": map[string]string{
						"ok": "bool", "command": "string", "schema": "string", "data": "object (command specific)",
						"warnings": "[]string, non-fatal; always relay relevant ones", "error": "{code, message, hint, details?} when ok=false",
					},
					"exit_codes":   map[string]string{"0": "success", "1": "runtime error", "2": "usage / validation error"},
					"error_codes":  ErrorCodes,
					"fetch_status": map[string]string{string(core.FetchOK): "fetched and parsed", string(core.FetchNotModified): "HTTP 304, nothing new (healthy)", string(core.FetchError): "failed; see error, counted in consecutive_failures"},
					"source_kinds": kinds,
					"ranking":      rank.Rules(),
					"digest":       map[string]any{"max_items": digest.MaxItems, "max_summary_chars": digest.MaxSummaryRunes, "recommended_items": "4-7"},
					"env": map[string]string{
						"SCOUT_HOME":         "data directory (default ~/.scout): watchlist.json, state.json, digests/",
						"SCOUT_OUTPUT":       "set to 'json' to force JSON output",
						"SCOUT_HTTP_TIMEOUT": "HTTP timeout in seconds (default 20)",
					},
					"commands": cmds,
				}
				return &Output{Data: data, Human: func(w io.Writer) {
					for _, c := range cmds {
						fmt.Fprintf(w, "%-26s %s\n", c.Command, c.Summary)
					}
				}}, nil
			}
		},
	}
}
