package sites

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/guille/rss-builder/internal/rss"
)

type GhosttyParser struct {
	httpClient *http.Client
}

func (GhosttyParser) Name() string         { return "Ghostty release notes" }
func (GhosttyParser) URL() string          { return "https://ghostty.org/docs/install/release-notes" }
func (p GhosttyParser) dateFormat() string { return "January 2, 2006" }
func (p GhosttyParser) Fetch() ([]rss.Item, error) {
	doc, err := fetchDocument(p.httpClient, p.URL())
	if err != nil {
		return nil, fmt.Errorf("fetch document: %w", err)
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

			parsedDate, perr := time.Parse(p.dateFormat(), date)
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
