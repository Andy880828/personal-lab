package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
	"github.com/Andy880828/personal-lab/scout/internal/digest"
	"github.com/Andy880828/personal-lab/scout/internal/rank"
	"github.com/Andy880828/personal-lab/scout/internal/store"
)

func digestCreateCmd() *Command {
	return &Command{
		Name:    "create",
		Summary: "save the agent's picks (JSON: {items:[{id,summary,why?}],note?}) as the digest for a date",
		Examples: []string{
			`echo '{"items":[{"id":"3f2a9c1d0b7e","summary":"一句話繁中摘要"}]}' | scout digest create --input - --json`,
			"scout digest create --input picks.json --dry-run --json",
		},
		Setup: func(fs *flag.FlagSet) Handler {
			input := fs.String("input", "", "path to the submission JSON, or - for stdin (required)")
			date := fs.String("date", "", "digest date YYYY-MM-DD (default: today, local time)")
			replace := fs.Bool("replace", false, "overwrite an existing digest for the date")
			dry := fs.Bool("dry-run", false, "validate and render without saving")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				if *input == "" {
					return nil, usageError("pass --input file.json or --input - (stdin); format in SKILLS.md §5", "--input is required")
				}
				now := rt.Now()
				day := *date
				if day == "" {
					day = now.Local().Format("2006-01-02")
				}
				if !store.IsValidDate(day) {
					return nil, usageError("format YYYY-MM-DD", "invalid --date %q", day)
				}
				var r io.Reader = rt.Stdin
				if *input != "-" {
					f, err := os.Open(*input)
					if err != nil {
						return nil, usageError("check the --input path", "%v", err)
					}
					defer f.Close()
					r = f
				}
				sub, err := digest.ParseSubmission(r)
				if err != nil {
					return nil, usageError(`expected {"items":[{"id":"...","summary":"...","why":"..."}],"note":"..."}`, "%v", err)
				}

				st := rt.Store()
				if !*dry {
					unlock, err := st.Lock()
					if err != nil {
						return nil, err
					}
					defer unlock()
				}
				w, err := st.LoadWatchlist()
				if err != nil {
					return nil, err
				}
				state, err := st.LoadState()
				if err != nil {
					return nil, err
				}
				if err := digest.Validate(sub, state.Items); err != nil {
					return nil, err
				}
				existing, err := st.LoadDigest(day)
				if err == nil && !*replace {
					return nil, &Error{Code: "conflict", Message: "digest for " + day + " already exists", Hint: "add --replace to overwrite (confirm with the user), or use --date", exit: ExitUsage}
				}
				if err != nil && !errors.Is(err, store.ErrNoDigest) {
					return nil, err
				}

				// 時間窗：若是取代既有 digest，沿用它的起點，避免把前一天已讀的範圍重算一次。
				var from time.Time
				if existing != nil {
					from = existing.WindowFrom
				} else if from, _, err = defaultWindowStart(st, now); err != nil {
					return nil, err
				}
				cs := collectCandidates(state, w, from, now, false, false, "")
				srcs := sourceIndex(w)
				d := &core.Digest{Date: day, CreatedAt: now, WindowFrom: from, WindowTo: now, CandidatesConsidered: len(cs.List), Note: sub.Note}
				var warnings []string
				for _, p := range sub.Items {
					it := state.Items[p.ID]
					sc := rank.Score(it, srcs, w.Boost, w.Mute, now)
					if it.PickedIn != "" && it.PickedIn != day {
						warnings = append(warnings, fmt.Sprintf("%s was already picked in %s", it.ID, it.PickedIn))
					}
					d.Items = append(d.Items, core.DigestItem{
						ID: it.ID, Title: it.Title, URL: it.URL, Summary: p.Summary, Why: p.Why,
						Sources: digest.SourceNames(it, srcs), DiscussionURL: digest.DiscussionURL(it), Score: sc.Score,
					})
				}
				saved := false
				if !*dry {
					if existing != nil {
						for _, old := range existing.Items {
							if it := state.Items[old.ID]; it != nil && it.PickedIn == day {
								it.PickedIn = ""
							}
						}
					}
					for _, p := range sub.Items {
						state.Items[p.ID].PickedIn = day
					}
					if err := st.SaveDigest(d); err != nil {
						return nil, err
					}
					if err := st.SaveState(state); err != nil {
						return nil, err
					}
					saved = true
				}
				md := digest.Markdown(d)
				data := map[string]any{"saved": saved, "path": filepath.Join(st.DigestDir(), day+".json"), "digest": d, "markdown": md,
					"next_steps": []string{"show `markdown` to the user in the conversation", "scout digest export --out <dir> --json  # refresh the static site if one is served"}}
				return &Output{Data: data, Warnings: warnings, Human: func(o io.Writer) { fmt.Fprint(o, md) }}, nil
			}
		},
	}
}

func digestShowCmd() *Command {
	return &Command{
		Name:     "show",
		Summary:  "show a digest (default latest); human output as Markdown or plain text",
		Args:     "[YYYY-MM-DD|latest]",
		Examples: []string{"scout digest show --json", "scout digest show 2026-11-01 --format text"},
		Setup: func(fs *flag.FlagSet) Handler {
			format := fs.String("format", "md", "human output format: md or text (JSON output always includes both)")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				if *format != "md" && *format != "text" {
					return nil, usageError("use md or text", "invalid --format %q", *format)
				}
				day := "latest"
				if len(args) > 0 {
					day = args[0]
				}
				d, err := rt.Store().LoadDigest(day)
				if err != nil {
					if !errors.Is(err, store.ErrNoDigest) {
						return nil, usageError("format YYYY-MM-DD or latest", "%v", err)
					}
					return nil, err
				}
				md, txt := digest.Markdown(d), digest.Text(d)
				return &Output{Data: map[string]any{"digest": d, "markdown": md, "text": txt}, Human: func(o io.Writer) {
					if *format == "text" {
						fmt.Fprint(o, txt)
					} else {
						fmt.Fprint(o, md)
					}
				}}, nil
			}
		},
	}
}

func digestListCmd() *Command {
	return &Command{
		Name:    "list",
		Summary: "list saved digests, newest first",
		Setup: func(fs *flag.FlagSet) Handler {
			limit := fs.Int("limit", 30, "max rows (0 = all)")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				st := rt.Store()
				dates, err := st.DigestDates()
				if err != nil {
					return nil, err
				}
				total := len(dates)
				var rows []map[string]any
				for i := len(dates) - 1; i >= 0 && (*limit == 0 || len(rows) < *limit); i-- {
					d, err := st.LoadDigest(dates[i])
					if err != nil {
						return nil, err
					}
					first := ""
					if len(d.Items) > 0 {
						first = d.Items[0].Title
					}
					rows = append(rows, map[string]any{"date": d.Date, "items": len(d.Items), "created_at": d.CreatedAt, "first_title": first})
				}
				if rows == nil {
					rows = []map[string]any{}
				}
				return &Output{Data: map[string]any{"total": total, "returned": len(rows), "digests": rows}, Human: func(o io.Writer) {
					for _, r := range rows {
						fmt.Fprintf(o, "%s  %2d items  %s\n", r["date"], r["items"], r["first_title"])
					}
				}}, nil
			}
		},
	}
}

func digestExportCmd() *Command {
	return &Command{
		Name:     "export",
		Summary:  "write a self-contained static site (index.html + one page per day) for Caddy file_server or any web host",
		Examples: []string{"scout digest export --out /srv/scout-site --json"},
		Setup: func(fs *flag.FlagSet) Handler {
			out := fs.String("out", "", "output directory (required; created if missing)")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				if *out == "" {
					return nil, usageError("pass --out <dir>", "--out is required")
				}
				st := rt.Store()
				dates, err := st.DigestDates()
				if err != nil {
					return nil, err
				}
				if len(dates) == 0 {
					return nil, fmt.Errorf("%w: nothing to export", store.ErrNoDigest)
				}
				if err := os.MkdirAll(*out, 0o755); err != nil {
					return nil, err
				}
				var all []*core.Digest
				var files []string
				for i := len(dates) - 1; i >= 0; i-- {
					d, err := st.LoadDigest(dates[i])
					if err != nil {
						return nil, err
					}
					all = append(all, d)
					p := filepath.Join(*out, d.Date+".html")
					if err := os.WriteFile(p, []byte(digest.HTMLPage(d)), 0o644); err != nil {
						return nil, err
					}
					files = append(files, p)
				}
				idx := filepath.Join(*out, "index.html")
				if err := os.WriteFile(idx, []byte(digest.HTMLIndex(all)), 0o644); err != nil {
					return nil, err
				}
				files = append([]string{idx}, files...)
				return &Output{Data: map[string]any{"out": *out, "digests": len(all), "files": files}, Human: func(o io.Writer) {
					fmt.Fprintf(o, "wrote %d files to %s\n", len(files), *out)
				}}, nil
			}
		},
	}
}
