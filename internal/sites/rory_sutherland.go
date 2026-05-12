package sites

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/guille/rss-builder/internal/rss"
)

type SutherlandParser struct {
	httpClient *http.Client
}

func (SutherlandParser) Name() string { return "Rory Sutherland (Spectator.co.uk)" }
func (SutherlandParser) URL() string {
	return "https://www.spectator.co.uk/writer/rory-sutherland/?filter=article&edition=uk"
}
func (SutherlandParser) dateFormat() string { return "2 January 2006" }
func (p SutherlandParser) Fetch() ([]rss.Item, error) {
	req, err := http.NewRequest(http.MethodGet, p.URL(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "rss-builder")

	res, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch spectator: %w", err)
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

	doc.Find("div.mosaic").EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			titleSel := s.Find(".article__title")
			if titleSel.Length() == 0 {
				firstErr = fmt.Errorf("missing title selector at index %d", i)
				return false
			}
			title := strings.TrimSpace(titleSel.Text())
			if title == "" {
				firstErr = fmt.Errorf("empty title at index %d", i)
				return false
			}

			dateSel := s.Find("time.archive-entry__timestamp")
			if dateSel.Length() == 0 {
				firstErr = fmt.Errorf("missing date selector at index %d", i)
				return false
			}
			inputDate := strings.TrimSpace(dateSel.Text())
			parsedDate, perr := time.Parse(p.dateFormat(), inputDate)
			if perr != nil {
				firstErr = fmt.Errorf("parse date %q at index %d: %w", inputDate, i, perr)
				return false
			}

			linkSel := s.Find("a.article__title-link")
			if linkSel.Length() == 0 {
				firstErr = fmt.Errorf("missing link selector at index %d", i)
				return false
			}
			link, exists := linkSel.Attr("href")
			if !exists || link == "" {
				firstErr = fmt.Errorf("empty link at index %d", i)
				return false
			}

			descSel := s.Find("p.article__excerpt-text")
			if descSel.Length() == 0 {
				firstErr = fmt.Errorf("missing description selector at index %d", i)
				return false
			}
			description := strings.TrimSpace(descSel.Text())

			items = append(items, rss.Item{
				Title:       title,
				Link:        link,
				Description: description,
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
