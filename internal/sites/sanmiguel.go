package sites

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"go.guillerg.dev/rss-builder/internal/rss"
)

const (
	sanmiguelBaseURL = "https://www.vozpopuli.com"
	sanmiguelContent = "div.post-container"
	sanmiguelJunk    = ".ad-slot-mobile, .ad-slot-content, .recirculation-block, .recirculation-posts, .post-contributions, #comments-content"
)

type SanmiguelParser struct {
	httpClient *http.Client
}

func (SanmiguelParser) Name() string { return "Jorge San Miguel (Vozpopuli)" }
func (SanmiguelParser) URL() string {
	return sanmiguelBaseURL + "/redaccion/jorge-san-miguel-lobeto"
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

			article, err := fetchDocument(p.httpClient, link)
			if err != nil {
				firstErr = fmt.Errorf("fetch %s: %w", link, err)
				return false
			}

			inputDate, err := p.getDateFromArticle(article)
			if err != nil {
				firstErr = fmt.Errorf("couldn't get date from %s: %v", link, err)
				return false
			}
			parsedDate, perr := time.Parse(p.dateFormat(), inputDate)
			if perr != nil {
				firstErr = fmt.Errorf("parse date %q at index %d: %w", inputDate, i, perr)
				return false
			}

			content, cerr := articleHTML(article, sanmiguelContent, sanmiguelBaseURL, sanmiguelJunk)
			if cerr != nil {
				log.Printf("%s: content for %s: %v", p.Name(), link, cerr)
			}

			items = append(items, rss.Item{
				Title:       title,
				Link:        link,
				Description: "",
				GUID:        rss.NewGUID(link),
				PubDate:     parsedDate.Format(rss.PubDateFormat),
				Content:     rss.NewCDATA(content),
			})
			return true
		})

	if firstErr != nil {
		return nil, firstErr
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no items found — the site layout may have changed")
	}
	return items, nil
}

// getDateFromArticle extracts the article's date from its header
func (p SanmiguelParser) getDateFromArticle(doc *goquery.Document) (string, error) {
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
