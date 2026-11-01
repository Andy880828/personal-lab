package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/Andy880828/personal-lab/scout/internal/digest"
	"github.com/Andy880828/personal-lab/scout/internal/extract"
	"github.com/Andy880828/personal-lab/scout/internal/feed"
)

func readCmd() *Command {
	return &Command{
		Name:     "read",
		Summary:  "fetch an item (by id or URL) and return its main text for the agent to summarise",
		Args:     "<item-id|url>",
		Examples: []string{"scout read 3f2a9c1d0b7e --json", "scout read https://example.com/post --max-chars 6000 --json"},
		Setup: func(fs *flag.FlagSet) Handler {
			maxChars := fs.Int("max-chars", 12000, "truncate text to this many characters (0 = no limit)")
			return func(ctx context.Context, rt *Runtime, args []string) (*Output, error) {
				if len(args) != 1 {
					return nil, usageError("pass exactly one item id or URL", "read needs one argument")
				}
				target := args[0]
				data := map[string]any{}
				url := target
				if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
					state, err := rt.Store().LoadState()
					if err != nil {
						return nil, err
					}
					it := state.Items[target]
					if it == nil {
						return nil, &Error{Code: "not_found", Message: "unknown item " + target, Hint: "use an id from `scout candidates --json`, or pass the URL", exit: ExitUsage}
					}
					url = it.URL
					w, err := rt.Store().LoadWatchlist()
					if err != nil {
						return nil, err
					}
					data["item"] = map[string]any{"id": it.ID, "title": it.Title, "sources": digest.SourceNames(it, sourceIndex(w)), "discussion_url": digest.DiscussionURL(it), "published": it.Published}
				}
				if _, _, ok := feed.Canonical(url); !ok {
					return nil, usageError("pass an http(s) URL", "invalid URL %q", url)
				}
				_, body, hdr, err := rt.Client.Get(ctx, url, nil)
				if err != nil {
					if he, ok := err.(*feed.HTTPError); ok {
						return nil, &Error{Code: "runtime", Message: fmt.Sprintf("GET %s: %v", url, he), Hint: "the site refused or the page is gone; summarise from the candidate excerpt and say so"}
					}
					return nil, err
				}
				ct := hdr.Get("Content-Type")
				if ct == "" {
					ct = http.DetectContentType(body)
				}
				mt, _, _ := mime.ParseMediaType(ct)
				var warnings []string
				var res extract.Result
				switch {
				case strings.Contains(mt, "html") || strings.Contains(mt, "xml"):
					res = extract.HTML(string(body), *maxChars)
				case strings.HasPrefix(mt, "text/"):
					res = extract.HTML("<pre>"+string(body)+"</pre>", *maxChars)
					res.Container = "plain"
				default:
					return nil, &Error{Code: "unsupported_content", Message: fmt.Sprintf("%s returned %s", url, mt), Hint: "not an HTML page (PDF/video/binary); summarise from the excerpt or open it another way"}
				}
				if res.Truncated {
					warnings = append(warnings, fmt.Sprintf("text truncated to %d of %d characters (--max-chars)", res.Returned, res.TotalChars))
				}
				if res.Container == "body" || res.Container == "document" {
					warnings = append(warnings, "no <article>/<main> found; text may include navigation or unrelated page parts")
				}
				if res.TotalChars < 300 {
					warnings = append(warnings, fmt.Sprintf("only %d characters extracted; the page may need JavaScript or block bots — do not summarise beyond what is here", res.TotalChars))
				}
				data["url"], data["content_type"] = url, mt
				data["title"], data["text"], data["container"] = res.Title, res.Text, res.Container
				data["total_chars"], data["returned_chars"], data["truncated"] = res.TotalChars, res.Returned, res.Truncated
				return &Output{Data: data, Warnings: warnings, Human: func(o io.Writer) {
					fmt.Fprintf(o, "# %s\n%s\n\n%s\n", res.Title, url, res.Text)
				}}, nil
			}
		},
	}
}
