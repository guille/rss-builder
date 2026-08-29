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
	albiacBaseURL = "https://www.eldebate.com"
	albiacContent = "div.c-detail__body"
	albiacJunk    = "div.content-add, .c-detail__box, .c-detail__tags, .c-detail__comments, aside"
)

type AlbiacParser struct {
	httpClient *http.Client
}

func (AlbiacParser) Name() string       { return "Gabriel Albiac (El Debate)" }
func (AlbiacParser) URL() string        { return albiacBaseURL + "/autor/gabriel-albiac/" }
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
			link := albiacBaseURL + relativeLink
			desc, content := p.articleParts(link)

			items = append(items, rss.Item{
				Title:       title,
				Link:        link,
				Description: desc,
				GUID:        rss.NewGUID(link),
				PubDate:     parsedDate.Format(rss.PubDateFormat),
				Content:     content,
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

// articleParts fetches the article page once and pulls out its subtitle and
// body. Both are best-effort: losing them costs a poorer item, not the feed.
func (p AlbiacParser) articleParts(url string) (string, *rss.CDATA) {
	doc, err := fetchDocument(p.httpClient, url)
	if err != nil {
		log.Printf("%s: fetch %s: %v", p.Name(), url, err)
		return "", nil
	}

	desc := strings.TrimSpace(doc.Find(".c-detail__subtitle").First().Text())

	content, err := articleHTML(doc, albiacContent, albiacBaseURL, albiacJunk)
	if err != nil {
		log.Printf("%s: content for %s: %v", p.Name(), url, err)
	}
	return desc, rss.NewCDATA(content)
}
