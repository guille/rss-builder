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
	req, err := http.NewRequest(http.MethodGet, p.URL(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "rss-builder")

	res, err := p.httpClient.Do(req)
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
				firstErr = fmt.Errorf("couldn't get date from %s", link)
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

func (p KirshatrovParser) getDateFromArticle(url string) (string, error) {
	// Get the Date from the article's footer
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "rss-builder")

	res, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	writtenIn := doc.Find(".text-base")
	if writtenIn.Length() == 0 {
		return "", err
	}
	// "Written in December 2025." ... Ugh
	return strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(writtenIn.Text()), "Written in "), "."), nil
}
