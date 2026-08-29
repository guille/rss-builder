package sites

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// fetchDocument uses the given client to make a Get request to the given url
// and build a goquery Document from a successful response body
func fetchDocument(client *http.Client, url string) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "rss-builder")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", res.StatusCode)
	}

	return goquery.NewDocumentFromReader(res.Body)
}

// nonContent is dropped from every article body
const nonContent = "script, style, noscript, iframe, form, template, button"

// articleHTML returns the inner HTML of the first node matching container, with
// relative links resolved against base. junk holds extra selectors to drop, for
// the ads and recirculation widgets sites like to interleave with the prose.
func articleHTML(doc *goquery.Document, container, base string, junk ...string) (string, error) {
	sel := doc.Find(container).First()
	if sel.Length() == 0 {
		return "", fmt.Errorf("no content matching %q", container)
	}

	sel.Find(strings.Join(append([]string{nonContent}, junk...), ", ")).Remove()
	// srcset values are relative and not worth resolving; readers fall back to
	// the <img> the <picture> wraps
	sel.Find("picture > source").Remove()

	baseURL, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse base %q: %w", base, err)
	}
	absolutise(sel, "a[href]", "href", baseURL)
	absolutise(sel, "img[src]", "src", baseURL)

	return sel.Html()
}

// absolutise rewrites attr on every node matching selector into an absolute URL
func absolutise(sel *goquery.Selection, selector, attr string, base *url.URL) {
	sel.Find(selector).Each(func(_ int, s *goquery.Selection) {
		ref, _ := s.Attr(attr)
		abs, err := base.Parse(ref)
		if err != nil {
			return
		}
		s.SetAttr(attr, abs.String())
	})
}
