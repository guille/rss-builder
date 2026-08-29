package sites

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"go.guillerg.dev/rss-builder/internal/rss"
)

const (
	kirshatrovBaseURL = "https://kirshatrov.com"
	kirshatrovContent = "div.post"
)

type KirshatrovParser struct {
	httpClient *http.Client
}

func (KirshatrovParser) Name() string       { return "Kir Shatrov" }
func (KirshatrovParser) URL() string        { return kirshatrovBaseURL + "/posts/" }
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

	// Incredibly cursed HTML structure. Let's only get the most recent year...
	anchor := latestYearAnchor(doc)
	if anchor == nil {
		return nil, fmt.Errorf("no year anchor found — the site layout may have changed")
	}

	anchor.Siblings().EachWithBreak(
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
			link := kirshatrovBaseURL + relativeLink

			post, err := fetchDocument(p.httpClient, link)
			if err != nil {
				firstErr = fmt.Errorf("fetch %s: %w", link, err)
				return false
			}

			inputDate, err := p.getDateFromPost(post)
			if err != nil {
				firstErr = fmt.Errorf("couldn't get date from %s: %v", link, err)
				return false
			}
			parsedDate, perr := time.Parse(p.dateFormat(), inputDate)
			if perr != nil {
				firstErr = fmt.Errorf("parse date %q at index %d: %w", inputDate, i, perr)
				return false
			}

			content, cerr := articleHTML(post, kirshatrovContent, kirshatrovBaseURL)
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

// latestYearAnchor returns the heading of the most recent year the index groups
// posts under, or nil if there is none. Anchoring on the current year instead
// would empty the feed until the first post of every January.
func latestYearAnchor(doc *goquery.Document) *goquery.Selection {
	var (
		anchor *goquery.Selection
		latest int
	)
	doc.Find(`h2[id$="-ref"]`).Each(func(_ int, s *goquery.Selection) {
		id, _ := s.Attr("id")
		year, err := strconv.Atoi(strings.TrimSuffix(id, "-ref"))
		if err != nil || year <= latest {
			return
		}
		anchor, latest = s, year
	})
	return anchor
}

// getDateFromPost extracts the post's date from its footer
func (p KirshatrovParser) getDateFromPost(doc *goquery.Document) (string, error) {
	writtenIn := doc.Find(".text-base")
	if writtenIn.Length() == 0 {
		return "", fmt.Errorf("can't find date text element")
	}
	// "Written in December 2025." ... Ugh
	return strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(writtenIn.Text()), "Written in "), "."), nil
}
