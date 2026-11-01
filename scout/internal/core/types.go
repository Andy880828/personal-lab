// Package core 定義跨模組共用的資料型別與語意常數。
// 列舉值只在這裡定義一次（單一決定點），describe 也從這裡輸出。
package core

import "time"

// SchemaVersion 是 JSON envelope 的版本，欄位有不相容變更時才升版。
const SchemaVersion = "scout/v1"

// SourceKind 決定用哪個 parser 處理來源回應。
type SourceKind string

const (
	// KindFeed 是 RSS 2.0、Atom 1.0 或 RSS 1.0 (RDF)，格式自動偵測。
	KindFeed SourceKind = "feed"
	// KindHN 是 Hacker News 的 Algolia search API JSON（帶分數與留言數）。
	KindHN SourceKind = "hn"
)

// SourceKinds 列出所有合法的來源種類與說明。
var SourceKinds = map[SourceKind]string{
	KindFeed: "RSS 2.0 / Atom / RSS 1.0 (RDF), auto-detected",
	KindHN:   "Hacker News Algolia search API JSON (points + comments)",
}

// FetchStatus 是單一來源單次抓取的結果。
type FetchStatus string

const (
	FetchOK          FetchStatus = "ok"
	FetchNotModified FetchStatus = "not_modified" // 304：來源沒有新內容，不是錯誤
	FetchError       FetchStatus = "error"
)

// Source 是使用者自選的一個資訊來源。
type Source struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Kind     SourceKind `json:"kind"`
	URL      string     `json:"url"`
	Weight   float64    `json:"weight"`             // 來源可信度／重要性，1 為一般
	Followed bool       `json:"followed,omitempty"` // 你追蹤的「人」，排序時加分
	Tags     []string   `json:"tags,omitempty"`
	Disabled bool       `json:"disabled,omitempty"`
}

// Watchlist 是使用者可編輯、可放進 git 的設定：來源與關鍵字。
type Watchlist struct {
	Sources []Source `json:"sources"`
	Boost   []string `json:"boost"` // 命中就加分的興趣關鍵字
	Mute    []string `json:"mute"`  // 命中就排除的關鍵字
}

// Sighting 記錄某個來源在某次抓取看到了這篇文章。
type Sighting struct {
	SourceID      string    `json:"source_id"`
	SeenAt        time.Time `json:"seen_at"`
	DiscussionURL string    `json:"discussion_url,omitempty"` // 例如 HN 討論串
	Points        int       `json:"points,omitempty"`
	Comments      int       `json:"comments,omitempty"`
}

// Item 是去重後的一篇文章。ID 由正規化後的網址決定，所以同一篇文章
// 從不同來源進來會合併成一筆，Sightings 記錄所有來源。
type Item struct {
	ID        string     `json:"id"`
	URL       string     `json:"url"`
	Title     string     `json:"title"`
	Excerpt   string     `json:"excerpt,omitempty"` // 來源提供的摘要，已去 HTML、截斷
	Author    string     `json:"author,omitempty"`
	Published *time.Time `json:"published,omitempty"` // nil 代表來源沒給或無法解析，不是「現在」
	FirstSeen time.Time  `json:"first_seen"`
	Sightings []Sighting `json:"sightings"`
	PickedIn  string     `json:"picked_in,omitempty"` // 被選進哪一天的 digest
}

// SourceState 是某個來源的抓取狀態，用來做條件式請求與健康度回報。
type SourceState struct {
	ETag                string      `json:"etag,omitempty"`
	LastModified        string      `json:"last_modified,omitempty"`
	LastAttempt         time.Time   `json:"last_attempt"`
	LastSuccess         *time.Time  `json:"last_success,omitempty"`
	LastStatus          FetchStatus `json:"last_status"`
	LastError           string      `json:"last_error,omitempty"`
	LastItemCount       int         `json:"last_item_count"`
	ConsecutiveFailures int         `json:"consecutive_failures"`
}

// State 是 scout 自己維護的資料（使用者不必手動編輯）。
type State struct {
	Items     map[string]*Item        `json:"items"`
	Sources   map[string]*SourceState `json:"sources"`
	LastFetch *time.Time              `json:"last_fetch,omitempty"`
}

// DigestItem 是 agent 選進 digest 的一篇文章，Summary 由 agent 撰寫。
type DigestItem struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	URL           string   `json:"url"`
	Summary       string   `json:"summary"`
	Why           string   `json:"why,omitempty"`
	Sources       []string `json:"sources"`
	DiscussionURL string   `json:"discussion_url,omitempty"`
	Score         float64  `json:"score"`
}

// Digest 是某一天的精選。它是自足的副本，prune 刪掉 item 後仍可閱讀。
type Digest struct {
	Date                 string       `json:"date"` // YYYY-MM-DD，本機時區
	CreatedAt            time.Time    `json:"created_at"`
	WindowFrom           time.Time    `json:"window_from"`
	WindowTo             time.Time    `json:"window_to"`
	CandidatesConsidered int          `json:"candidates_considered"`
	Note                 string       `json:"note,omitempty"`
	Items                []DigestItem `json:"items"`
}
