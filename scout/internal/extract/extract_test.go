package extract

import (
	"strings"
	"testing"
)

func TestPrefersArticleAndDropsNoise(t *testing.T) {
	body := strings.Repeat("Real content sentence. ", 20)
	doc := `<html><head><title>T &amp; Co</title><meta property="og:title" content="OG Title"><style>.x{}</style></head><body>
<nav>Home | About</nav><script>var secret = 1;</script>
<article><h1>Heading</h1><p>` + body + `</p><p>Second&nbsp;para &lt;ok&gt;</p></article>
<footer>Copyright</footer></body></html>`
	r := HTML(doc, 0)
	if r.Title != "OG Title" || r.Container != "article" {
		t.Fatalf("title=%q container=%q", r.Title, r.Container)
	}
	for _, bad := range []string{"secret", "Home | About", "Copyright", ".x{}"} {
		if strings.Contains(r.Text, bad) {
			t.Errorf("text should not contain %q", bad)
		}
	}
	if !strings.Contains(r.Text, "Second para <ok>") || !strings.HasPrefix(r.Text, "Heading") {
		t.Errorf("text=%q", r.Text[:80])
	}
}

func TestTruncationIsReported(t *testing.T) {
	doc := "<body><p>" + strings.Repeat("字", 500) + "</p></body>"
	r := HTML(doc, 100)
	if !r.Truncated || r.Returned != 100 || r.TotalChars != 500 {
		t.Fatalf("truncated=%v returned=%d total=%d", r.Truncated, r.Returned, r.TotalChars)
	}
}

// 回歸：先 ToLower 再取索引時，'İ' 轉小寫會變長，導致切片錯位。
func TestUnicodeBeforeArticleDoesNotShiftOffsets(t *testing.T) {
	body := strings.Repeat("Body text here. ", 20)
	doc := "<html><head><title>İİİİ</title></head><body><ARTICLE>" + body + "</ARTICLE></body></html>"
	r := HTML(doc, 0)
	if r.Container != "article" || !strings.HasPrefix(r.Text, "Body text here.") {
		t.Fatalf("container=%q text=%q", r.Container, r.Text[:30])
	}
}
