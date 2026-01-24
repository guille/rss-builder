package kirshatrov

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/guille/rss-builder/rss"
)

const (
	baseURL    = "https://kirshatrov.com/posts/"
	dateFormat = "January 2006"
)

type Parser struct{}

func getDateFromArticle(httpClient http.Client, url string) string {
	// Get the Date from the article's footer
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

	written_in := doc.Find(".text-base")
	if written_in.Length() == 0 {
		return ""
	}
	// "Written in December 2025." ... Ugh
	return strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(written_in.Text()), "Written in "), ".")
}

func (Parser) Name() string { return "Kir Shatrov" }
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

	// Incredibly cursed HTML structure. Let's only get the ones for this year...
	year := time.Now().Year()
	anchor := fmt.Sprintf("#%d-ref", year)

	doc.Find(anchor).Siblings().Each(func(i int, s *goquery.Selection) {
		if firstErr != nil {
			return
		}

		title := strings.TrimSpace(s.Text())
		if title == "" {
			firstErr = fmt.Errorf("empty title at index %d", i)
			return
		}

		linkSel := s.Find("a")
		if linkSel.Length() == 0 {
			firstErr = fmt.Errorf("missing link selector at index %d", i)
			return
		}
		relativeLink, exists := linkSel.Attr("href")
		if !exists || relativeLink == "" {
			firstErr = fmt.Errorf("empty link at index %d", i)
			return
		}
		link := fmt.Sprintf("https://kirshatrov.com%v", relativeLink)

		inputDate := getDateFromArticle(*httpClient, link)
		parsedDate, perr := time.Parse(dateFormat, inputDate)
		if perr != nil {
			firstErr = fmt.Errorf("parse date %q at index %d: %w", inputDate, i, perr)
			return
		}

		items = append(items, rss.Item{
			Title:       title,
			Link:        link,
			Description: "",
			GUID:        rss.NewGUID(link),
			PubDate:     parsedDate.Format(rss.PubDateFormat),
		})
	})

	if firstErr != nil {
		return nil, firstErr
	}
	return items, nil
}
