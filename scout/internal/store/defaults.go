package store

import "github.com/Andy880828/personal-lab/scout/internal/core"

// HNFrontPageURL 是 Algolia 的 HN 首頁 API，一次請求就拿到分數與留言數。
const HNFrontPageURL = "https://hn.algolia.com/api/v1/search?tags=front_page&hitsPerPage=50"

// DefaultWatchlist 是 `scout init` 寫入的起始設定。
// 來源組合參考 Threads 原文（HN、TechCrunch、The Verge 加上自己追蹤的作者），
// 再補上 AI 實驗室官方部落格與台灣的 iThome。權重只是起點，請依個人需求調整。
func DefaultWatchlist() *core.Watchlist {
	return &core.Watchlist{
		Sources: []core.Source{
			{ID: "hn", Name: "Hacker News 首頁", Kind: core.KindHN, URL: HNFrontPageURL, Weight: 1.2, Tags: []string{"aggregator"}},
			{ID: "lobsters", Name: "Lobsters", Kind: core.KindFeed, URL: "https://lobste.rs/rss", Weight: 1.0, Tags: []string{"aggregator"}},
			{ID: "techcrunch", Name: "TechCrunch", Kind: core.KindFeed, URL: "https://techcrunch.com/feed/", Weight: 0.8, Tags: []string{"media"}},
			{ID: "theverge", Name: "The Verge", Kind: core.KindFeed, URL: "https://www.theverge.com/rss/index.xml", Weight: 0.7, Tags: []string{"media"}},
			{ID: "arstechnica", Name: "Ars Technica", Kind: core.KindFeed, URL: "https://feeds.arstechnica.com/arstechnica/index", Weight: 0.8, Tags: []string{"media"}},
			{ID: "ithome", Name: "iThome", Kind: core.KindFeed, URL: "https://www.ithome.com.tw/rss", Weight: 0.9, Tags: []string{"media", "tw"}},
			{ID: "simonwillison", Name: "Simon Willison", Kind: core.KindFeed, URL: "https://simonwillison.net/atom/everything/", Weight: 1.5, Followed: true, Tags: []string{"person", "ai"}},
			{ID: "karpathy", Name: "Andrej Karpathy", Kind: core.KindFeed, URL: "https://karpathy.bearblog.dev/feed/", Weight: 1.5, Followed: true, Tags: []string{"person", "ai"}},
			{ID: "ghuntley", Name: "Geoffrey Huntley", Kind: core.KindFeed, URL: "https://ghuntley.com/rss/", Weight: 1.3, Followed: true, Tags: []string{"person", "ai"}},
			{ID: "latentspace", Name: "Latent Space", Kind: core.KindFeed, URL: "https://www.latent.space/feed", Weight: 1.2, Tags: []string{"ai"}},
			{ID: "openai", Name: "OpenAI News", Kind: core.KindFeed, URL: "https://openai.com/news/rss.xml", Weight: 1.2, Tags: []string{"lab", "ai"}},
			{ID: "deepmind", Name: "Google DeepMind Blog", Kind: core.KindFeed, URL: "https://deepmind.google/blog/rss.xml", Weight: 1.1, Tags: []string{"lab", "ai"}},
			{ID: "huggingface", Name: "Hugging Face Blog", Kind: core.KindFeed, URL: "https://huggingface.co/blog/feed.xml", Weight: 1.0, Tags: []string{"ai"}},
			{ID: "claude-code-releases", Name: "Claude Code releases", Kind: core.KindFeed, URL: "https://github.com/anthropics/claude-code/releases.atom", Weight: 1.0, Tags: []string{"release", "ai"}},
			{ID: "arxiv-cs-ai", Name: "arXiv cs.AI", Kind: core.KindFeed, URL: "https://rss.arxiv.org/rss/cs.AI", Weight: 0.5, Tags: []string{"paper", "ai"}, Disabled: true},
		},
		Boost: []string{"claude", "anthropic", "agent", "llm", "mcp", "open source", "security", "vulnerability"},
		Mute:  []string{},
	}
}
