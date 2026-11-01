// Package digest 驗證 agent 提交的精選，並輸出 Markdown、純文字與靜態 HTML。
//
// 分工：scout 負責資料正確（ID 存在、不重複、摘要不為空、來源可追溯），
// agent 負責判斷與文字（挑哪幾篇、一句話摘要怎麼寫）。
package digest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

// MaxItems 是一份 digest 的硬上限。建議 4–7 篇（見 SKILLS.md），
// 上限設寬一點，讓特別熱鬧的日子仍可提交，但擋掉「把候選清單整包丟進來」。
const MaxItems = 12

// MaxSummaryRunes 是單篇摘要的字數上限。原文設計是「一句話」，
// 200 字已經很寬；超過代表 agent 寫成了段落。
const MaxSummaryRunes = 200

// Pick 是 agent 提交的單篇精選。
type Pick struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Why     string `json:"why,omitempty"`
}

// Submission 是 `scout digest create` 讀取的 JSON。
type Submission struct {
	Items []Pick `json:"items"`
	Note  string `json:"note,omitempty"`
}

// ValidationError 列出所有問題，讓 agent 一次修完，而不是改一個錯一個。
type ValidationError struct {
	Problems   []string
	UnknownIDs []string
}

func (e *ValidationError) Error() string { return strings.Join(e.Problems, "; ") }

// ParseSubmission 讀取並解析提交內容，拒絕未知欄位以免拼錯欄位名被無聲忽略。
func ParseSubmission(r io.Reader) (*Submission, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var s Submission
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("parse submission json: %w", err)
	}
	return &s, nil
}

// Validate 檢查提交內容；items 是目前 state 中的文章。
func Validate(s *Submission, items map[string]*core.Item) error {
	ve := &ValidationError{}
	if len(s.Items) == 0 {
		ve.Problems = append(ve.Problems, "items is empty")
	}
	if len(s.Items) > MaxItems {
		ve.Problems = append(ve.Problems, fmt.Sprintf("%d items exceeds the maximum of %d", len(s.Items), MaxItems))
	}
	seen := map[string]bool{}
	for i, p := range s.Items {
		switch {
		case p.ID == "":
			ve.Problems = append(ve.Problems, fmt.Sprintf("items[%d].id is empty", i))
		case seen[p.ID]:
			ve.Problems = append(ve.Problems, fmt.Sprintf("items[%d].id %s is duplicated", i, p.ID))
		case items[p.ID] == nil:
			ve.UnknownIDs = append(ve.UnknownIDs, p.ID)
			ve.Problems = append(ve.Problems, fmt.Sprintf("items[%d].id %s not found", i, p.ID))
		}
		seen[p.ID] = true
		if strings.TrimSpace(p.Summary) == "" {
			ve.Problems = append(ve.Problems, fmt.Sprintf("items[%d].summary is empty", i))
		} else if n := utf8.RuneCountInString(p.Summary); n > MaxSummaryRunes {
			ve.Problems = append(ve.Problems, fmt.Sprintf("items[%d].summary has %d characters (max %d; write one sentence)", i, n, MaxSummaryRunes))
		}
	}
	if len(ve.Problems) > 0 {
		return ve
	}
	return nil
}

// IsValidation 判斷錯誤是否為驗證失敗。
func IsValidation(err error) (*ValidationError, bool) {
	var ve *ValidationError
	ok := errors.As(err, &ve)
	return ve, ok
}

// SourceNames 把 sighting 轉成來源名稱清單（依 watchlist 名稱，找不到就用 ID）。
func SourceNames(it *core.Item, sources map[string]core.Source) []string {
	var out []string
	for _, sg := range it.Sightings {
		name := sg.SourceID
		if s, ok := sources[sg.SourceID]; ok && s.Name != "" {
			name = s.Name
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// DiscussionURL 回傳第一個討論串連結（例如 HN），沒有就是空字串。
func DiscussionURL(it *core.Item) string {
	for _, sg := range it.Sightings {
		if sg.DiscussionURL != "" {
			return sg.DiscussionURL
		}
	}
	return ""
}
