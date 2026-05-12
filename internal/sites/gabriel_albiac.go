package sites

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/guille/rss-builder/internal/rss"
)

type AlbiacParser struct {
	httpClient *http.Client
}

func (AlbiacParser) Name() string       { return "Gabriel Albiac (El Debate)" }
func (AlbiacParser) URL() string        { return "https://www.eldebate.com/autor/gabriel-albiac/" }
func (AlbiacParser) dateFormat() string { return "02/01/2006" }
func (p AlbiacParser) Fetch() ([]rss.Item, error) {
	doc, err := fetchDocument(p.httpClient, p.URL())
	if err != nil {
		return nil, fmt.Errorf("fetch document: %w", err)
	}

	var (
		items    []rss.Item
		firstErr error
	)

	doc.Find("article.c-article").EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			titleSel := s.Find(".c-article__title")
			if titleSel.Length() == 0 {
				firstErr = fmt.Errorf("missing title selector at index %d", i)
				return false
			}
			title := strings.TrimSpace(titleSel.Text())
			if title == "" {
				firstErr = fmt.Errorf("empty title at index %d", i)
				return false
			}

			dateSel := s.Find("div.date")
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

			linkSel := s.Find("a.page-link")
			if linkSel.Length() == 0 {
				firstErr = fmt.Errorf("missing link selector at index %d", i)
				return false
			}
			relativeLink, exists := linkSel.Attr("href")
			if !exists || relativeLink == "" {
				firstErr = fmt.Errorf("empty link at index %d", i)
				return false
			}
			link := "https://www.eldebate.com" + relativeLink
			desc, _ := p.getDescriptionFromArticle(link)

			items = append(items, rss.Item{
				Title:       title,
				Link:        link,
				Description: desc,
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

// getDescriptionFromArticle tries to extract the description from inside the article in the given url
func (p AlbiacParser) getDescriptionFromArticle(url string) (string, error) {
	doc, err := fetchDocument(p.httpClient, url)
	if err != nil {
		return "", fmt.Errorf("fetch document: %w", err)
	}

	description := doc.Find("h2.c-detail__subtitle").First().Text()
	return strings.TrimSpace(description), nil
}
