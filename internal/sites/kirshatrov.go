package sites

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/guille/rss-builder/internal/rss"
)

type KirshatrovParser struct {
	httpClient *http.Client
}

func (KirshatrovParser) Name() string       { return "Kir Shatrov" }
func (KirshatrovParser) URL() string        { return "https://kirshatrov.com/posts/" }
func (KirshatrovParser) dateFormat() string { return "January 2006" }
func (p KirshatrovParser) Fetch() ([]rss.Item, error) {
	doc, err := fetchDocument(p.httpClient, p.URL())
	if err != nil {
		return nil, fmt.Errorf("fetch document: %w", err)
	}

	var (
		items    []rss.Item
		firstErr error
	)

	// Incredibly cursed HTML structure. Let's only get the ones for this year...
	year := time.Now().Year()
	anchor := fmt.Sprintf("#%d-ref", year)

	doc.Find(anchor).Siblings().EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			title := strings.TrimSpace(s.Text())
			if title == "" {
				firstErr = fmt.Errorf("empty title at index %d", i)
				return false
			}

			linkSel := s.Find("a")
			if linkSel.Length() == 0 {
				firstErr = fmt.Errorf("missing link selector at index %d", i)
				return false
			}
			relativeLink, exists := linkSel.Attr("href")
			if !exists || relativeLink == "" {
				firstErr = fmt.Errorf("empty link at index %d", i)
				return false
			}
			link := "https://kirshatrov.com" + relativeLink

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

// getDateFromArticle extracts the article's date from given url's footer
func (p KirshatrovParser) getDateFromArticle(url string) (string, error) {
	doc, err := fetchDocument(p.httpClient, url)
	if err != nil {
		return "", fmt.Errorf("fetch document: %w", err)
	}

	writtenIn := doc.Find(".text-base")
	if writtenIn.Length() == 0 {
		return "", fmt.Errorf("can't find date text element")
	}
	// "Written in December 2025." ... Ugh
	return strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(writtenIn.Text()), "Written in "), "."), nil
}
