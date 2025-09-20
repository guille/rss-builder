package gabriel_albiac

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/guille/rss-builder/rss"
)

const (
	baseURL    = "https://www.eldebate.com/autor/gabriel-albiac/"
	dateFormat = "02/01/2006"
)

type Parser struct{}

func getDescriptionFromArticle(httpClient http.Client, url string) string {
	// Try to get the description from inside the article
	// Soft fail to empty string in case of any error
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "rss-builder")

	res, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return ""
	}

	description := doc.Find("h2.c-detail__subtitle")
	if description.Length() == 0 {
		return ""
	}
	return strings.TrimSpace(description.Text())
}

func (Parser) Name() string { return "Gabriel Albiac (El Debate)" }
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
		return nil, fmt.Errorf("fetch eldebate: %w", err)
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

	doc.Find("article.c-article").Each(func(i int, s *goquery.Selection) {
		if firstErr != nil {
			return
		}

		titleSel := s.Find(".c-article__title")
		if titleSel.Length() == 0 {
			firstErr = fmt.Errorf("missing title selector at index %d", i)
			return
		}
		title := strings.TrimSpace(titleSel.Text())
		if title == "" {
			firstErr = fmt.Errorf("empty title at index %d", i)
			return
		}

		dateSel := s.Find("div.date")
		if dateSel.Length() == 0 {
			firstErr = fmt.Errorf("missing date selector at index %d", i)
			return
		}
		inputDate := strings.TrimSpace(dateSel.Text())
		parsedDate, perr := time.Parse(dateFormat, inputDate)
		if perr != nil {
			firstErr = fmt.Errorf("parse date %q at index %d: %w", inputDate, i, perr)
			return
		}

		linkSel := s.Find("a.page-link")
		if linkSel.Length() == 0 {
			firstErr = fmt.Errorf("missing link selector at index %d", i)
			return
		}
		relativeLink, exists := linkSel.Attr("href")
		if !exists || relativeLink == "" {
			firstErr = fmt.Errorf("empty link at index %d", i)
			return
		}
		link := fmt.Sprintf("https://www.eldebate.com%v", relativeLink)

		items = append(items, rss.Item{
			Title:       title,
			Link:        link,
			Description: getDescriptionFromArticle(*httpClient, link),
			GUID:        rss.NewGUID(link),
			PubDate:     parsedDate.Format(rss.PubDateFormat),
		})
	})

	if firstErr != nil {
		return nil, firstErr
	}
	return items, nil
}
