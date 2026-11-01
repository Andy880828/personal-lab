// Package trends 比較兩個相鄰時間窗的標題詞頻，找出正在升溫的主題。
//
// 這是「不被時代拋下」的雷達：digest 回答「今天該讀什麼」，trends 回答
// 「最近大家開始談什麼」。刻意只做詞頻差異，不做語意分群——語意判讀
// 交給讀這份結果的 agent。
package trends

import (
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

// MinCount：目前時間窗至少要有幾篇文章提到，才列入結果（過濾單篇雜訊）。
const MinCount = 2

// Term 是一個詞（或雙詞片語）在兩個時間窗的出現次數。
type Term struct {
	Term     string   `json:"term"`
	Current  int      `json:"current"`  // 目前時間窗內提到它的文章數
	Previous int      `json:"previous"` // 前一個時間窗的文章數
	Delta    int      `json:"delta"`
	New      bool     `json:"new"`      // 前一個時間窗完全沒出現
	Examples []string `json:"examples"` // 最多 3 個 item ID
}

// Report 是 trends 的結果。
type Report struct {
	CurrentFrom  time.Time `json:"current_from"`
	PreviousFrom time.Time `json:"previous_from"`
	To           time.Time `json:"to"`
	ItemsCurrent int       `json:"items_current"`
	ItemsPrev    int       `json:"items_previous"`
	SkippedCJK   int       `json:"skipped_cjk_titles"` // 含中日韓文字的標題無法用空白斷詞，未納入
	Rising       []Term    `json:"rising"`
	Top          []Term    `json:"top"`
}

// Compute 以 days 天為一個時間窗，比較 [now-days, now) 與前一個時間窗。
func Compute(items map[string]*core.Item, days int, limit int, now time.Time) Report {
	span := time.Duration(days) * 24 * time.Hour
	r := Report{To: now, CurrentFrom: now.Add(-span), PreviousFrom: now.Add(-2 * span)}
	cur, prev := map[string][]string{}, map[string]int{}

	ids := make([]string, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Strings(ids) // 讓 examples 穩定
	for _, id := range ids {
		it := items[id]
		t := it.FirstSeen
		if it.Published != nil {
			t = *it.Published
		}
		inCur := !t.Before(r.CurrentFrom) && t.Before(now)
		inPrev := !t.Before(r.PreviousFrom) && t.Before(r.CurrentFrom)
		if !inCur && !inPrev {
			continue
		}
		if hasCJK(it.Title) {
			r.SkippedCJK++
			continue
		}
		terms := Terms(it.Title)
		if inCur {
			r.ItemsCurrent++
			for _, term := range terms {
				cur[term] = append(cur[term], id)
			}
		} else {
			r.ItemsPrev++
			for _, term := range terms {
				prev[term]++
			}
		}
	}

	var all []Term
	for term, list := range cur {
		if len(list) < MinCount {
			continue
		}
		ex := list
		if len(ex) > 3 {
			ex = ex[:3]
		}
		p := prev[term]
		all = append(all, Term{Term: term, Current: len(list), Previous: p, Delta: len(list) - p, New: p == 0, Examples: ex})
	}

	top := append([]Term(nil), all...)
	sort.Slice(top, func(i, j int) bool {
		if top[i].Current != top[j].Current {
			return top[i].Current > top[j].Current
		}
		return top[i].Term < top[j].Term
	})
	rising := make([]Term, 0, len(all))
	for _, t := range all {
		if t.Delta > 0 {
			rising = append(rising, t)
		}
	}
	sort.Slice(rising, func(i, j int) bool {
		if rising[i].Delta != rising[j].Delta {
			return rising[i].Delta > rising[j].Delta
		}
		if rising[i].Current != rising[j].Current {
			return rising[i].Current > rising[j].Current
		}
		return rising[i].Term < rising[j].Term
	})
	r.Top, r.Rising = clip(top, limit), clip(rising, limit)
	return r
}

func clip(t []Term, n int) []Term {
	if n > 0 && len(t) > n {
		return t[:n]
	}
	if t == nil {
		return []Term{}
	}
	return t
}

// Terms 把標題切成去重後的單詞與雙詞片語（去掉停用詞與純數字）。
func Terms(title string) []string {
	var words []string
	for _, raw := range strings.FieldsFunc(strings.ToLower(title), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '.' || r == '+' || r == '#')
	}) {
		w := strings.Trim(raw, "-.")
		words = append(words, w)
	}
	seen := map[string]bool{}
	var out []string
	addTerm := func(t string) {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	for i, w := range words {
		if keep(w) {
			addTerm(w)
		}
		if i+1 < len(words) && keep(w) && keep(words[i+1]) {
			addTerm(w + " " + words[i+1])
		}
	}
	return out
}

func keep(w string) bool {
	if w == "" || stopwords[w] {
		return false
	}
	digits, letters := 0, 0
	for _, r := range w {
		if unicode.IsDigit(r) {
			digits++
		} else if unicode.IsLetter(r) {
			letters++
		}
	}
	if letters == 0 {
		return false // 純數字或符號
	}
	return len(w) >= 3 || digits > 0 // "ai" 太短被過濾，"o3"、"go1" 保留
}

func hasCJK(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) || unicode.Is(unicode.Hangul, r) {
			return true
		}
	}
	return false
}

var stopwords = func() map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(`the and for with from that this your you are was were will can how why what when who
	into about over after before more most new now not but all its it's our out has have had just than then them they
	their there here his her via use using used get gets got make makes made one two first last best top show ask tell
	hn launch introducing announces announced says said week day year years today vs per off any some way ways like
	being been does did done want need now's let's i'm we're it’s don't can't isn't
	could would should might must may much many very really still back down even also only own other every
	without against under again people thing things time times world big small good bad better worse old
	look looks going goes come comes take takes give gives keep work works working part found finds`) {
		m[w] = true
	}
	return m
}()
