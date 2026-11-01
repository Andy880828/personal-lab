package feed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Andy880828/personal-lab/scout/internal/core"
	"github.com/Andy880828/personal-lab/scout/internal/version"
)

const (
	// maxBody 限制單一回應大小。OpenAI 的 feed 約 0.8 MB，10 MB 足夠；
	// 超過時明確報錯，不會只解析前半段。
	maxBody = 10 << 20
	// concurrency 是同時抓取的來源數，對個人用途與來源站都夠客氣。
	concurrency = 6
)

// ErrTooLarge 代表回應超過 maxBody。
var ErrTooLarge = fmt.Errorf("response larger than %d bytes", maxBody)

// HTTPError 是非 2xx/304 的回應。
type HTTPError struct{ Status int }

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http %d %s", e.Status, http.StatusText(e.Status))
}

// Client 包裝 HTTP 設定。
type Client struct {
	HTTP      *http.Client
	UserAgent string
}

// NewClient 依 SCOUT_HTTP_TIMEOUT（秒，預設 20）建立 client。
func NewClient() *Client {
	timeout := 20 * time.Second
	if v, err := strconv.Atoi(os.Getenv("SCOUT_HTTP_TIMEOUT")); err == nil && v > 0 {
		timeout = time.Duration(v) * time.Second
	}
	return &Client{
		HTTP:      &http.Client{Timeout: timeout},
		UserAgent: "scout/" + version.Version + " (personal news radar; +https://github.com/Andy880828/personal-lab)",
	}
}

// Get 發出 GET。prev 不為 nil 時帶上條件式請求標頭；304 時 body 為 nil。
func (c *Client) Get(ctx context.Context, url string, prev *core.SourceState) (status int, body []byte, hdr http.Header, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml, application/json, text/html;q=0.9, */*;q=0.5")
	if prev != nil {
		if prev.ETag != "" {
			req.Header.Set("If-None-Match", prev.ETag)
		}
		if prev.LastModified != "" {
			req.Header.Set("If-Modified-Since", prev.LastModified)
		}
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotModified {
		return resp.StatusCode, nil, resp.Header, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp.StatusCode, nil, resp.Header, &HTTPError{Status: resp.StatusCode}
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return resp.StatusCode, nil, resp.Header, err
	}
	if len(b) > maxBody {
		return resp.StatusCode, nil, resp.Header, ErrTooLarge
	}
	return resp.StatusCode, b, resp.Header, nil
}

// Result 是單一來源的抓取結果。
type Result struct {
	Source       core.Source
	Status       core.FetchStatus
	Format       string
	Entries      []Entry
	Skipped      int
	Err          error
	ETag         string
	LastModified string
	Duration     time.Duration
}

// FetchAll 平行抓取所有來源，結果順序與輸入相同。單一來源失敗不影響其他來源。
func (c *Client) FetchAll(ctx context.Context, sources []core.Source, states map[string]*core.SourceState) []Result {
	out := make([]Result, len(sources))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for i, s := range sources {
		wg.Add(1)
		go func(i int, s core.Source) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out[i] = c.FetchOne(ctx, s, states[s.ID])
		}(i, s)
	}
	wg.Wait()
	return out
}

// FetchOne 抓取並解析單一來源。
func (c *Client) FetchOne(ctx context.Context, s core.Source, prev *core.SourceState) Result {
	start := time.Now()
	r := Result{Source: s}
	status, body, hdr, err := c.Get(ctx, s.URL, prev)
	r.Duration = time.Since(start)
	if err != nil {
		r.Status, r.Err = core.FetchError, err
		return r
	}
	r.ETag, r.LastModified = hdr.Get("ETag"), hdr.Get("Last-Modified")
	if status == http.StatusNotModified {
		r.Status = core.FetchNotModified
		if prev != nil { // 304 不一定回傳驗證標頭，沿用舊值
			if r.ETag == "" {
				r.ETag = prev.ETag
			}
			if r.LastModified == "" {
				r.LastModified = prev.LastModified
			}
		}
		return r
	}
	r.Format, r.Entries, r.Skipped, err = Parse(s.Kind, body)
	if err != nil {
		r.Status, r.Err = core.FetchError, err
		return r
	}
	r.Status = core.FetchOK
	return r
}
