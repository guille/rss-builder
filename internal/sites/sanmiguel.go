package sites

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"go.guillerg.dev/rss-builder/internal/rss"
)

type SanmiguelParser struct {
	httpClient *http.Client
}

func (SanmiguelParser) Name() string { return "Jorge San Miguel (Vozpopuli)" }
func (SanmiguelParser) URL() string {
	return "https://www.vozpopuli.com/redaccion/jorge-san-miguel-lobeto"
}
func (SanmiguelParser) dateFormat() string { return time.RFC3339 }
func (p SanmiguelParser) Fetch() ([]rss.Item, error) {
	doc, err := fetchDocument(p.httpClient, p.URL())
	if err != nil {
		return nil, fmt.Errorf("fetch document: %w", err)
	}

	var (
		items    []rss.Item
		firstErr error
	)

	doc.Find("a.block-post-title-link").EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			title := strings.TrimSpace(s.Text())
			if title == "" {
				firstErr = fmt.Errorf("empty title at index %d", i)
				return false
			}

			link, exists := s.Attr("href")
			if !exists || link == "" {
				firstErr = fmt.Errorf("empty link at index %d", i)
				return false
			}

			inputDate, err := p.getDateFromArticle(link)
			if err != nil {
				firstErr = fmt.Errorf("couldn't get date from %s: %v", link, err)
				return false
			}
			parsedDate, perr := time.Parse(p.dateFormat(), inputDate)
			if perr != nil {
				firstErr = fmt.Errorf("parse date %q at index %d: %w", inputDate, i, perr)
				return false
			}

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

// getDateFromArticle extracts the article's date from the given url
func (p SanmiguelParser) getDateFromArticle(url string) (string, error) {
	doc, err := fetchDocument(p.httpClient, url)
	if err != nil {
		return "", fmt.Errorf("fetch document: %w", err)
	}

	publication := doc.Find("div.post-publication-date time")
	if publication.Length() == 0 {
		return "", fmt.Errorf("can't find date text element")
	}

	datetimeAttr, exists := publication.Attr("datetime")
	if !exists {
		return "", fmt.Errorf("time element has no datetime attribute")
	}

	return datetimeAttr, nil
}
