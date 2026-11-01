// Package rank 用透明、可解釋的規則替候選文章打分。
//
// 分數只是「預設排序」，最後選哪幾篇由 agent 判斷。每一分都附理由
// （reasons），agent 可以直接引用向使用者解釋為什麼某篇排前面。
// 參考 syssec 的做法：離散、公開的規則優於黑盒模型分數。
package rank

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

// 評分規則。改這些數字就是改排序語意，所以集中在這裡並由 describe 輸出。
const (
	SourceWeightPoints = 10.0 // 來源權重 × 10（取所有看到它的來源中最高者）
	FollowedPoints     = 15.0 // 任一來源是你追蹤的人
	CoveragePoints     = 8.0  // 每多一個來源報導 +8
	HNPointsFactor     = 5.0  // 5 × log10(1 + HN points)
	HNCommentsFactor   = 2.0  // 2 × log10(1 + HN comments)
	BoostTitlePoints   = 6.0  // 興趣關鍵字出現在標題
	BoostExcerptPoints = 3.0  // 興趣關鍵字只出現在摘要
	BoostCap           = 18.0 // 關鍵字加分上限，避免標題堆關鍵字的文章霸榜
	HalfLifeHours      = 36.0 // 時間衰減半衰期：36 小時前的文章分數減半

	// LateDiscoveryHours：文章發布時間早於時間窗起點多少小時內仍算候選。
	// 用來接住「昨天發、今天才被 feed 列出」的文章，同時擋掉第一次抓取時
	// feed 裡幾週前的舊文。
	LateDiscoveryHours = 72.0
)

// Rules 是給 describe 輸出的規則說明（單一決定點，文件不另抄一份）。
func Rules() map[string]string {
	return map[string]string{
		"source":    fmt.Sprintf("+%.0f × highest weight among sources that listed the item", SourceWeightPoints),
		"followed":  fmt.Sprintf("+%.0f if any listing source is marked followed (a person you follow)", FollowedPoints),
		"coverage":  fmt.Sprintf("+%.0f per additional distinct source", CoveragePoints),
		"hn":        fmt.Sprintf("+%.0f×log10(1+points) + %.0f×log10(1+comments) from HN sightings", HNPointsFactor, HNCommentsFactor),
		"boost":     fmt.Sprintf("+%.0f per boost keyword in title, +%.0f if only in excerpt, capped at %.0f", BoostTitlePoints, BoostExcerptPoints, BoostCap),
		"mute":      "any mute keyword in title/excerpt excludes the item (shown only with --include-muted)",
		"recency":   fmt.Sprintf("total × 0.5^(age_hours/%.0f); age from published time, or first_seen when unknown", HalfLifeHours),
		"window":    fmt.Sprintf("candidate if first_seen >= window start AND (published unknown OR published >= window start - %.0fh), and not already picked", LateDiscoveryHours),
		"keywords":  "case-insensitive; ASCII keywords match on word boundaries (\"ai\" does not match \"said\"), others by substring",
		"tie_break": "higher score, then newer, then id",
	}
}

// Scored 是一篇候選文章與它的分數。
type Scored struct {
	Item     *core.Item
	Score    float64
	Reasons  []string
	AgeHours float64
	Muted    string // 命中的 mute 關鍵字，空字串代表沒被靜音
}

// Score 計算單篇分數。sources 以 ID 索引；已從 watchlist 移除的來源權重視為 1。
func Score(it *core.Item, sources map[string]core.Source, boost, mute []string, now time.Time) Scored {
	s := Scored{Item: it}
	var total float64
	add := func(points float64, format string, a ...any) {
		if points == 0 {
			return
		}
		total += points
		s.Reasons = append(s.Reasons, fmt.Sprintf("%s (+%.1f)", fmt.Sprintf(format, a...), points))
	}

	maxW, bestSrc, followed := 0.0, "", ""
	distinct := map[string]bool{}
	var hnPoints, hnComments int
	for _, sg := range it.Sightings {
		distinct[sg.SourceID] = true
		src, ok := sources[sg.SourceID]
		w := 1.0
		if ok {
			w = src.Weight
			if src.Followed && followed == "" {
				followed = src.ID
			}
		}
		if w > maxW {
			maxW, bestSrc = w, sg.SourceID
		}
		if sg.Points > hnPoints {
			hnPoints = sg.Points
		}
		if sg.Comments > hnComments {
			hnComments = sg.Comments
		}
	}
	add(maxW*SourceWeightPoints, "source %s weight %.1f", bestSrc, maxW)
	if followed != "" {
		add(FollowedPoints, "followed author via %s", followed)
	}
	if n := len(distinct); n > 1 {
		add(float64(n-1)*CoveragePoints, "covered by %d sources", n)
	}
	if hnPoints > 0 || hnComments > 0 {
		add(HNPointsFactor*math.Log10(1+float64(hnPoints))+HNCommentsFactor*math.Log10(1+float64(hnComments)),
			"hn %d points, %d comments", hnPoints, hnComments)
	}

	title, excerpt := strings.ToLower(it.Title), strings.ToLower(it.Excerpt)
	var boostPts float64
	var hits []string
	for _, k := range boost {
		switch {
		case Match(title, k):
			boostPts += BoostTitlePoints
			hits = append(hits, k)
		case Match(excerpt, k):
			boostPts += BoostExcerptPoints
			hits = append(hits, k+"(excerpt)")
		}
	}
	if boostPts > 0 {
		add(math.Min(boostPts, BoostCap), "keywords %s", strings.Join(hits, ","))
	}
	for _, k := range mute {
		if Match(title, k) || Match(excerpt, k) {
			s.Muted = k
			break
		}
	}

	ref := it.FirstSeen
	if it.Published != nil && it.Published.Before(now) {
		ref = *it.Published
	}
	s.AgeHours = math.Max(0, now.Sub(ref).Hours())
	decay := math.Pow(0.5, s.AgeHours/HalfLifeHours)
	s.Score = math.Round(total*decay*10) / 10
	s.Reasons = append(s.Reasons, fmt.Sprintf("recency %.0fh old (×%.2f)", s.AgeHours, decay))
	s.AgeHours = math.Round(s.AgeHours*10) / 10
	return s
}

// InWindow 判斷文章是否屬於從 from 開始的候選時間窗（規則見 Rules()["window"]）。
func InWindow(it *core.Item, from time.Time) bool {
	if it.FirstSeen.Before(from) {
		return false
	}
	if it.Published == nil {
		return true
	}
	return !it.Published.Before(from.Add(-time.Duration(LateDiscoveryHours * float64(time.Hour))))
}

// Sort 依分數、時間、ID 排序（結果穩定，方便比對）。
func Sort(list []Scored) {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Score != list[j].Score {
			return list[i].Score > list[j].Score
		}
		if list[i].AgeHours != list[j].AgeHours {
			return list[i].AgeHours < list[j].AgeHours
		}
		return list[i].Item.ID < list[j].Item.ID
	})
}

// Match 不分大小寫比對關鍵字。純 ASCII 關鍵字要求前後不是字母或數字，
// 避免 "ai" 命中 "said"；含非 ASCII（例如中文）的關鍵字用子字串比對。
// text 必須已轉小寫。
func Match(text, keyword string) bool {
	k := strings.ToLower(strings.TrimSpace(keyword))
	if k == "" {
		return false
	}
	if !isASCII(k) {
		return strings.Contains(text, k)
	}
	for start := 0; ; {
		i := strings.Index(text[start:], k)
		if i < 0 {
			return false
		}
		i += start
		end := i + len(k)
		if (i == 0 || !isWordByte(text[i-1])) && (end == len(text) || !isWordByte(text[end])) {
			return true
		}
		start = i + 1
	}
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func isWordByte(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}
