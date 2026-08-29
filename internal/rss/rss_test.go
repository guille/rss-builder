package rss

import (
	"bytes"
	"strings"
	"testing"
)

// Go's encoding/xml has no namespace-prefix support: it emits the literal
// names from the struct tags. This pins that behaviour.
func TestWriteContentEncoded(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, Channel{
		Title: "feed",
		Items: []Item{
			{Title: "with", Content: NewCDATA("<p>body & more</p>")},
			{Title: "without", Content: NewCDATA("   ")},
		},
	})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		`xmlns:content="http://purl.org/rss/1.0/modules/content/"`,
		`<content:encoded><![CDATA[<p>body & more</p>]]></content:encoded>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}

	if got := strings.Count(out, "<content:encoded>"); got != 1 {
		t.Errorf("got %d content:encoded elements, want 1:\n%s", got, out)
	}
}
