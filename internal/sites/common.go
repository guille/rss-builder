package sites

import (
	"fmt"
	"net/http"

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
