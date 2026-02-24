package ghostty

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/guille/rss-builder/rss"
)

const (
	baseURL    = "https://ghostty.org/docs/install/release-notes"
	dateFormat = "January 2, 2006"
)

type Parser struct{}

func (Parser) Name() string { return "Ghostty release notes" }
func (Parser) URL() string  { return baseURL }
func (Parser) Fetch() ([]rss.Item, error) {
	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "rss-builder")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch ghostty: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	var (
		items    []rss.Item
		firstErr error
	)
	const key = "Released on "

	doc.Find(`div ul li a[href*="release-notes"]`).EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			parentText := strings.TrimSpace(s.Parent().Text())
			if !strings.Contains(parentText, key) {
				// Skip the sidebar links and all the other garbage...
				return true
			}

			linkSel := s
			title := strings.TrimSpace(linkSel.Text())
			if title == "" {
				firstErr = fmt.Errorf("empty title at index %d", i)
				return false
			}

			_, date, ok := strings.Cut(parentText, key)
			if !ok {
				firstErr = fmt.Errorf("couldn't find release date in %s", parentText)
				return false
			}

			parsedDate, perr := time.Parse(dateFormat, date)
			if perr != nil {
				firstErr = fmt.Errorf("parse date %q at index %d: %w", date, i, perr)
				return false
			}

			relativeLink, exists := linkSel.Attr("href")
			if !exists || relativeLink == "" {
				firstErr = fmt.Errorf("empty link at index %d", i)
				return false
			}
			link := "https://ghostty.org" + relativeLink

			items = append(items, rss.Item{
				Title:       title,
				Link:        link,
				Description: "",
				GUID:        rss.NewGUID(link),
				PubDate:     parsedDate.Format(rss.PubDateFormat),
			})

			return true
		})

	if firstErr != nil {
		return nil, firstErr
	}
	return items, nil
}
