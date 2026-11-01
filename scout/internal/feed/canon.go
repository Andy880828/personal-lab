package feed

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

// trackingParams 是只用來追蹤、不影響內容的 query 參數。
// 移除它們，才能把同一篇文章從不同來源的連結合併。
var trackingParams = map[string]bool{
	"fbclid": true, "gclid": true, "mc_cid": true, "mc_eid": true,
	"ref_src": true, "igshid": true, "cmpid": true,
}

// Canonical 回傳正規化網址（顯示用）與文章 ID（去重用）。
//
// ID 的 key 不含 scheme 與 www.，所以 http/https、有無 www 視為同一篇；
// 顯示用網址則保留原本的 scheme，不擅自改寫成可能打不開的版本。
func Canonical(raw string) (display, id string, ok bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", "", false
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Fragment = ""
	q := u.Query()
	for k := range q {
		if strings.HasPrefix(strings.ToLower(k), "utm_") || trackingParams[strings.ToLower(k)] {
			q.Del(k)
		}
	}
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		for _, v := range q[k] {
			parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	u.RawQuery = strings.Join(parts, "&")
	if len(u.Path) > 1 {
		u.Path = strings.TrimRight(u.Path, "/")
		u.RawPath = ""
	}
	if u.Path == "/" {
		u.Path = ""
	}
	display = u.String()

	key := strings.TrimPrefix(u.Host, "www.") + u.EscapedPath()
	if u.RawQuery != "" {
		key += "?" + u.RawQuery
	}
	sum := sha256.Sum256([]byte(key))
	return display, hex.EncodeToString(sum[:])[:12], true
}
