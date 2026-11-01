// Package store 管理 SCOUT_HOME 目錄下的所有檔案：
//
//	watchlist.json   使用者設定（來源、關鍵字），可放進 git
//	state.json       scout 自己維護的文章與抓取狀態
//	digests/*.json   每天一份 digest，自足、永不自動刪除
//	scout.lock       寫入鎖，避免 agent 與排程同時寫檔
//
// 只用標準庫：資料量是「每天數百篇」等級，JSON 檔加原子寫入就夠，
// 不值得為此引入 SQLite 依賴。
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

var (
	// ErrNoWatchlist 代表還沒執行 `scout init`。
	ErrNoWatchlist = errors.New("watchlist not found")
	// ErrNoDigest 代表指定日期（或任何日期）沒有 digest。
	ErrNoDigest = errors.New("digest not found")
	// ErrLocked 代表另一個 scout 程序正在寫入。
	ErrLocked = errors.New("another scout process holds the lock")
)

// lockStaleAfter 之後視為前一個程序異常結束，鎖可以被接手。
// 一次完整 fetch 在正常網路下不到一分鐘，10 分鐘留了足夠餘裕。
const lockStaleAfter = 10 * time.Minute

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// Store 對應一個 SCOUT_HOME 目錄。
type Store struct {
	Home string
}

// DefaultHome 回傳 SCOUT_HOME，未設定時為 ~/.scout。
func DefaultHome() string {
	if h := os.Getenv("SCOUT_HOME"); h != "" {
		return h
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".scout"
	}
	return filepath.Join(home, ".scout")
}

// New 建立 Store，不會碰檔案系統。
func New(home string) *Store { return &Store{Home: home} }

func (s *Store) path(name string) string { return filepath.Join(s.Home, name) }

// WatchlistPath 是設定檔位置，回報給 agent 用。
func (s *Store) WatchlistPath() string { return s.path("watchlist.json") }

// DigestDir 是 digest 存放目錄。
func (s *Store) DigestDir() string { return s.path("digests") }

// Lock 取得寫入鎖，回傳釋放函式。讀取類指令不需要鎖。
func (s *Store) Lock() (func(), error) {
	if err := os.MkdirAll(s.Home, 0o755); err != nil {
		return nil, err
	}
	p := s.path("scout.lock")
	for attempt := 0; attempt < 2; attempt++ {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintf(f, "pid=%d at=%s\n", os.Getpid(), time.Now().Format(time.RFC3339))
			f.Close()
			return func() { os.Remove(p) }, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		fi, statErr := os.Stat(p)
		if statErr == nil && time.Since(fi.ModTime()) > lockStaleAfter {
			os.Remove(p) // 前一個程序異常結束留下的鎖
			continue
		}
		return nil, fmt.Errorf("%w (%s)", ErrLocked, p)
	}
	return nil, fmt.Errorf("%w (%s)", ErrLocked, p)
}

// LoadWatchlist 讀取設定，不存在時回傳 ErrNoWatchlist。
func (s *Store) LoadWatchlist() (*core.Watchlist, error) {
	var w core.Watchlist
	if err := readJSON(s.WatchlistPath(), &w); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoWatchlist
		}
		return nil, fmt.Errorf("read %s: %w", s.WatchlistPath(), err)
	}
	for i := range w.Sources {
		if w.Sources[i].Weight == 0 {
			w.Sources[i].Weight = 1
		}
	}
	return &w, nil
}

// SaveWatchlist 原子寫入設定。
func (s *Store) SaveWatchlist(w *core.Watchlist) error {
	return writeJSON(s.WatchlistPath(), w)
}

// LoadState 讀取狀態，不存在時回傳空狀態（第一次執行是正常情況）。
func (s *Store) LoadState() (*core.State, error) {
	st := &core.State{Items: map[string]*core.Item{}, Sources: map[string]*core.SourceState{}}
	err := readJSON(s.path("state.json"), st)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read state.json: %w", err)
	}
	if st.Items == nil {
		st.Items = map[string]*core.Item{}
	}
	if st.Sources == nil {
		st.Sources = map[string]*core.SourceState{}
	}
	return st, nil
}

// SaveState 原子寫入狀態。
func (s *Store) SaveState(st *core.State) error {
	return writeJSON(s.path("state.json"), st)
}

// LoadDigest 讀取某天的 digest；date 為 "latest" 時取最新一份。
func (s *Store) LoadDigest(date string) (*core.Digest, error) {
	if date == "" || date == "latest" {
		dates, err := s.DigestDates()
		if err != nil {
			return nil, err
		}
		if len(dates) == 0 {
			return nil, ErrNoDigest
		}
		date = dates[len(dates)-1]
	}
	if !dateRe.MatchString(date) {
		return nil, fmt.Errorf("invalid date %q (want YYYY-MM-DD or latest)", date)
	}
	var d core.Digest
	if err := readJSON(filepath.Join(s.DigestDir(), date+".json"), &d); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w for %s", ErrNoDigest, date)
		}
		return nil, err
	}
	return &d, nil
}

// SaveDigest 寫入 digest。
func (s *Store) SaveDigest(d *core.Digest) error {
	if !dateRe.MatchString(d.Date) {
		return fmt.Errorf("invalid digest date %q", d.Date)
	}
	return writeJSON(filepath.Join(s.DigestDir(), d.Date+".json"), d)
}

// DigestDates 回傳所有 digest 日期，由舊到新。
func (s *Store) DigestDates() ([]string, error) {
	entries, err := os.ReadDir(s.DigestDir())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".json")
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") && dateRe.MatchString(name) {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out, nil
}

// IsValidDate 檢查 YYYY-MM-DD 格式。
func IsValidDate(s string) bool {
	if !dateRe.MatchString(s) {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// writeJSON 先寫暫存檔再 rename，避免寫到一半斷電留下壞檔。
func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(b, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
