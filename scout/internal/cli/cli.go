// Package cli 是給 agent 操作的指令介面：每個指令回傳同一份資料，
// 依 --json 輸出穩定的 JSON envelope，或精簡的人類可讀文字。
// 架構沿用 syssec：指令樹即文件來源，describe 由同一棵樹產生，不會跟行為脫節。
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
	"github.com/Andy880828/personal-lab/scout/internal/digest"
	"github.com/Andy880828/personal-lab/scout/internal/feed"
	"github.com/Andy880828/personal-lab/scout/internal/store"
)

// Exit codes。
const (
	ExitOK    = 0
	ExitError = 1
	ExitUsage = 2
)

// ErrorCodes 是所有穩定的 error.code 與處理方式（describe 與 SKILLS.md 的唯一來源）。
var ErrorCodes = map[string]string{
	"usage":               "bad flags/args or invalid input JSON; see hint",
	"no_data":             "nothing recorded yet (no watchlist, no fetch, no digest); run the hinted command",
	"not_found":           "unknown item/source/digest id; error.details lists offending ids when relevant",
	"conflict":            "target already exists (watchlist, source id, digest date); see hint for the override flag",
	"locked":              "another scout process is writing; retry in a minute",
	"network":             "remote unreachable or timed out; offline commands still work",
	"unsupported_content": "URL did not return HTML/text (e.g. PDF, video); read it another way",
	"runtime":             "anything else; report message to the user",
}

// Command 是指令樹的一個節點。
type Command struct {
	Name     string
	Summary  string
	Args     string
	Examples []string
	Setup    func(fs *flag.FlagSet) Handler
	Sub      []*Command
}

// Handler 在 flag 解析後執行。
type Handler func(ctx context.Context, rt *Runtime, args []string) (*Output, error)

// Output 是 handler 的產出。
type Output struct {
	Data     any
	Human    func(w io.Writer)
	Warnings []string
}

// Error 是結構化、agent 可據以行動的錯誤。
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
	Details any    `json:"details,omitempty"`
	exit    int
}

func (e *Error) Error() string { return e.Message }

func usageError(hint, format string, a ...any) *Error {
	return &Error{Code: "usage", Message: fmt.Sprintf(format, a...), Hint: hint, exit: ExitUsage}
}

// Runtime 帶著全域選項與依賴；測試可以替換 Now、Stdin、HTTP。
type Runtime struct {
	JSON   bool
	Home   string
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
	Now    func() time.Time
	Client *feed.Client
}

// Store 回傳目前 SCOUT_HOME 的 store。
func (rt *Runtime) Store() *store.Store { return store.New(rt.Home) }

type envelope struct {
	OK       bool     `json:"ok"`
	Command  string   `json:"command"`
	Schema   string   `json:"schema"`
	Data     any      `json:"data,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Error    *Error   `json:"error,omitempty"`
}

// Run 執行 CLI 並回傳 exit code。
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return RunWith(ctx, &Runtime{
		Stdout: stdout, Stderr: stderr, Stdin: os.Stdin,
		JSON: os.Getenv("SCOUT_OUTPUT") == "json", Home: store.DefaultHome(),
		Now: time.Now, Client: feed.NewClient(),
	}, args)
}

// RunWith 讓測試注入 Runtime。
func RunWith(ctx context.Context, rt *Runtime, args []string) int {
	root := Root()
	globals, rest := splitGlobals(args)
	if len(rest) > 0 && rest[0] == "help" { // `scout help fetch` 等同 `scout fetch --help`
		rest = append(rest[1:], "--help")
	}
	cmd, path, rest := resolve(root, rest)
	if cmd == root || cmd.Setup == nil {
		if len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
			fmt.Fprintf(rt.Stderr, "unknown command %q\n\n", strings.Join(append(path, rest[0]), " "))
			printHelp(rt.Stderr, cmd, path)
			return ExitUsage
		}
		printHelp(rt.Stdout, cmd, path)
		return ExitOK
	}
	name := strings.Join(path, " ")

	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&rt.JSON, "json", rt.JSON, "emit a JSON envelope (also: SCOUT_OUTPUT=json)")
	fs.StringVar(&rt.Home, "home", rt.Home, "data directory (also: SCOUT_HOME; default ~/.scout)")
	handler := cmd.Setup(fs)
	positional, err := parseInterleaved(fs, append(globals, rest...))
	if errors.Is(err, flag.ErrHelp) {
		if rt.JSON || slices.Contains(args, "--json") || slices.Contains(args, "-json") {
			writeJSON(rt.Stdout, envelope{OK: true, Command: name, Schema: core.SchemaVersion, Data: info(cmd, path)})
		} else {
			printHelp(rt.Stdout, cmd, path)
		}
		return ExitOK
	}
	if err != nil {
		return emitError(rt, name, usageError("run `scout "+name+" --help`", "%v", err))
	}
	out, err := handler(ctx, rt, positional)
	if err != nil {
		return emitError(rt, name, err)
	}
	if rt.JSON {
		writeJSON(rt.Stdout, envelope{OK: true, Command: name, Schema: core.SchemaVersion, Data: out.Data, Warnings: out.Warnings})
	} else {
		if out.Human != nil {
			out.Human(rt.Stdout)
		} else {
			writeJSON(rt.Stdout, out.Data)
		}
		for _, w := range out.Warnings {
			fmt.Fprintln(rt.Stderr, "warning:", w)
		}
	}
	return ExitOK
}

func emitError(rt *Runtime, name string, err error) int {
	e := classify(err)
	if e.exit == 0 {
		e.exit = ExitError
	}
	if rt.JSON {
		writeJSON(rt.Stdout, envelope{OK: false, Command: name, Schema: core.SchemaVersion, Error: e})
	} else {
		fmt.Fprintf(rt.Stderr, "error: %s\n", e.Message)
		if e.Hint != "" {
			fmt.Fprintf(rt.Stderr, "hint: %s\n", e.Hint)
		}
	}
	return e.exit
}

// classify 把錯誤對應到穩定的 code 與可行動的 hint。
func classify(err error) *Error {
	var e *Error
	var netErr net.Error
	var urlErr *url.Error
	switch {
	case errors.As(err, &e):
		return e
	case errors.Is(err, store.ErrNoWatchlist):
		return &Error{Code: "no_data", Message: err.Error(), Hint: "run `scout init --json` to create the default watchlist"}
	case errors.Is(err, store.ErrNoDigest):
		return &Error{Code: "no_data", Message: err.Error(), Hint: "list digests with `scout digest list --json`, or create one (see SKILLS.md §5)"}
	case errors.Is(err, store.ErrLocked):
		return &Error{Code: "locked", Message: err.Error(), Hint: "another scout command is writing; retry in a minute (stale locks expire after 10 minutes)"}
	case errors.As(err, &urlErr), errors.As(err, &netErr):
		return &Error{Code: "network", Message: err.Error(), Hint: "check connectivity/proxy; offline commands (status, candidates, digest, trends) still work"}
	}
	if ve, ok := digest.IsValidation(err); ok {
		code, hint := "usage", "fix every listed problem and resubmit"
		if len(ve.UnknownIDs) > 0 && len(ve.UnknownIDs) == len(ve.Problems) {
			code, hint = "not_found", "use ids from `scout candidates --json`"
		}
		return &Error{Code: code, Message: ve.Error(), Hint: hint, Details: map[string]any{"problems": ve.Problems, "unknown_ids": ve.UnknownIDs}, exit: ExitUsage}
	}
	return &Error{Code: "runtime", Message: err.Error()}
}

func writeJSON(w io.Writer, v any) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if f, ok := w.(*os.File); ok {
		if fi, err := f.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
			enc.SetIndent("", "  ") // 終端機縮排，pipe 給 agent 時緊湊
		}
	}
	_ = enc.Encode(v)
}

// splitGlobals 讓 --json、--home 可以放在指令名稱前面。
func splitGlobals(args []string) (globals, rest []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") || a == "-h" || a == "--help" || a == "-help" {
			return globals, args[i:]
		}
		globals = append(globals, a)
		if bare := strings.TrimLeft(a, "-"); bare == "home" && i+1 < len(args) {
			i++
			globals = append(globals, args[i])
		}
	}
	return globals, nil
}

func resolve(root *Command, args []string) (*Command, []string, []string) {
	cmd, path := root, []string{}
	for len(args) > 0 {
		next := find(cmd.Sub, args[0])
		if next == nil {
			break
		}
		cmd, path, args = next, append(path, next.Name), args[1:]
	}
	return cmd, path, args
}

func find(cs []*Command, name string) *Command {
	for _, c := range cs {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// parseInterleaved 允許 flag 出現在位置參數之後。
func parseInterleaved(fs *flag.FlagSet, args []string) ([]string, error) {
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		args = fs.Args()
		if len(args) == 0 {
			return pos, nil
		}
		if args[0] == "--" {
			return append(pos, args[1:]...), nil
		}
		pos = append(pos, args[0])
		args = args[1:]
	}
}

func printHelp(w io.Writer, cmd *Command, path []string) {
	name := strings.TrimSpace("scout " + strings.Join(path, " "))
	if cmd.Summary != "" {
		fmt.Fprintf(w, "%s — %s\n\n", name, cmd.Summary)
	}
	if len(cmd.Sub) > 0 {
		fmt.Fprintf(w, "Usage: %s <command> [flags]\n\nCommands:\n", name)
		for _, c := range cmd.Sub {
			fmt.Fprintf(w, "  %-12s %s\n", c.Name, c.Summary)
		}
		fmt.Fprintf(w, "\nEvery command accepts --json. `%s help <command>` for flags; `scout describe --json` for the full manifest.\n", name)
		return
	}
	fmt.Fprintf(w, "Usage: %s %s[flags]\n\nFlags:\n", name, strings.TrimSpace(cmd.Args+" "))
	for _, f := range flagsOf(cmd) {
		def := ""
		if f.Default != "" && f.Default != "false" {
			def = fmt.Sprintf(" (default %s)", f.Default)
		}
		fmt.Fprintf(w, "  --%-16s %s%s\n", f.Name, f.Usage, def)
	}
	if len(cmd.Examples) > 0 {
		fmt.Fprintln(w, "\nExamples:")
		for _, e := range cmd.Examples {
			fmt.Fprintln(w, "  "+e)
		}
	}
}

// FlagInfo 描述一個 flag。
type FlagInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Default string `json:"default,omitempty"`
	Usage   string `json:"usage"`
}

func flagsOf(cmd *Command) []FlagInfo {
	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	fs.Bool("json", false, "emit a JSON envelope (also: SCOUT_OUTPUT=json)")
	fs.String("home", "", "data directory (also: SCOUT_HOME; default ~/.scout)")
	cmd.Setup(fs)
	var out []FlagInfo
	fs.VisitAll(func(f *flag.Flag) {
		typ := strings.TrimSuffix(strings.TrimPrefix(fmt.Sprintf("%T", f.Value), "*flag."), "Value")
		out = append(out, FlagInfo{Name: f.Name, Type: typ, Default: f.DefValue, Usage: f.Usage})
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// CommandInfo 是 manifest 中的一個指令。
type CommandInfo struct {
	Command  string     `json:"command"`
	Summary  string     `json:"summary"`
	Args     string     `json:"args,omitempty"`
	Flags    []FlagInfo `json:"flags"`
	Examples []string   `json:"examples,omitempty"`
}

func info(c *Command, path []string) CommandInfo {
	return CommandInfo{Command: "scout " + strings.Join(path, " "), Summary: c.Summary, Args: c.Args, Flags: flagsOf(c), Examples: c.Examples}
}

// csv 拆逗號分隔的 flag 值。
func csv(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseDuration 支援 Go duration 再加上天數，例如 36h、3d、90m。
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("invalid duration %q (examples: 36h, 3d, 90m)", s)
	}
	return d, nil
}

func ageHours(now time.Time, t *time.Time) *float64 {
	if t == nil {
		return nil
	}
	h := float64(int(now.Sub(*t).Hours()*10)) / 10
	return &h
}
