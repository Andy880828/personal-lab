package feed

import (
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
)

// MaxEntryAgeDays：發布時間早於此天數的項目不寫入 store。
// 部分官方部落格（OpenAI、Hugging Face）的 feed 會列出上千篇完整歷史，
// 全部存下來只會讓 state.json 膨脹，而且它們本來就不會成為候選。
// 略過數量透過 MergeStats.Old 回報，不無聲丟棄。
const MaxEntryAgeDays = 14

// MergeStats 是一次合併的計數。
type MergeStats struct {
	New     int // 第一次看到的文章
	Seen    int // 已存在的文章（可能被別的來源先抓到）
	Invalid int // 網址無法正規化而略過
	Old     int // 發布時間早於 MaxEntryAgeDays 而略過（已存在的文章仍會更新）
}

// Merge 把一個來源的抓取結果併入 state。同一篇文章（同 ID）只保留一筆，
// 但每個來源的 sighting 都會記錄，HN 分數與留言數取最新值。
func Merge(st *core.State, src core.Source, entries []Entry, now time.Time) MergeStats {
	var ms MergeStats
	for _, e := range entries {
		display, id, ok := Canonical(e.URL)
		if !ok {
			ms.Invalid++
			continue
		}
		it, exists := st.Items[id]
		if !exists && e.Published != nil && e.Published.Before(now.AddDate(0, 0, -MaxEntryAgeDays)) {
			ms.Old++
			continue
		}
		if !exists {
			it = &core.Item{ID: id, URL: display, Title: e.Title, Excerpt: e.Excerpt, Author: e.Author, Published: e.Published, FirstSeen: now}
			st.Items[id] = it
			ms.New++
		} else {
			ms.Seen++
			if it.Excerpt == "" {
				it.Excerpt = e.Excerpt
			}
			if it.Published == nil {
				it.Published = e.Published
			}
			if it.Author == "" {
				it.Author = e.Author
			}
		}
		updated := false
		for i := range it.Sightings {
			if it.Sightings[i].SourceID == src.ID {
				it.Sightings[i].SeenAt = now
				it.Sightings[i].Points, it.Sightings[i].Comments = e.Points, e.Comments
				if e.DiscussionURL != "" {
					it.Sightings[i].DiscussionURL = e.DiscussionURL
				}
				updated = true
			}
		}
		if !updated {
			it.Sightings = append(it.Sightings, core.Sighting{SourceID: src.ID, SeenAt: now, DiscussionURL: e.DiscussionURL, Points: e.Points, Comments: e.Comments})
		}
	}
	return ms
}
