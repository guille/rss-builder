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
	ecbBaseURL      = "https://www.ecb.europa.eu"
	ecbMopoIndexURL = ecbBaseURL + "/press/govcdec/mopo/html/index.en.html"
	ecbMopoContent  = "main"
	ecbMopoJunk     = ".related-topics, .address-box"
)

type ECBMopoParser struct {
	httpClient *http.Client
}

func (ECBMopoParser) Name() string { return "ECB monetary policy decisions" }
func (ECBMopoParser) URL() string  { return ecbMopoIndexURL }

func (p ECBMopoParser) Fetch() ([]rss.Item, error) {
	year := time.Now().UTC().Year()

	var (
		items         []rss.Item
		firstErr      error
		fetchedAny    bool
		fragmentYears = []int{year, year - 1}
	)

	for _, y := range fragmentYears {
		fragmentURL := fmt.Sprintf("%s/press/govcdec/mopo/%d/html/index_include.en.html", ecbBaseURL, y)
		doc, err := fetchDocument(p.httpClient, fragmentURL)
		if err != nil {
			// The coming year's fragment 404s until the ECB publishes it;
			// a single missing year must not sink the feed
			log.Printf("%s: fragment for %d: %v", p.Name(), y, err)
			continue
		}
		fetchedAny = true

		doc.Find("dt[isoDate]").Not(".accordion dt").EachWithBreak(
			func(i int, s *goquery.Selection) bool {
				parsedDate, perr := ecbMopoDate(s)
				if perr != nil {
					firstErr = fmt.Errorf("year %d index %d: %w", y, i, perr)
					return false
				}

				dd := s.NextFiltered("dd")
				linkSel := dd.Find("div.title a").First()
				href, exists := linkSel.Attr("href")
				if !exists || strings.TrimSpace(href) == "" {
					firstErr = fmt.Errorf("year %d index %d: empty link", y, i)
					return false
				}
				link := href
				if !strings.HasPrefix(link, "http") {
					link = ecbBaseURL + link
				}

				title := strings.TrimSpace(linkSel.Text())
				if title == "" {
					firstErr = fmt.Errorf("year %d index %d: empty title", y, i)
					return false
				}

				content := p.articleBody(link)

				items = append(items, rss.Item{
					Title:       fmt.Sprintf("%s — %s", title, parsedDate.Format("2 January 2006")),
					Link:        link,
					Description: "",
					GUID:        rss.NewGUID(link),
					PubDate:     parsedDate.Format(rss.PubDateFormat),
					Content:     content,
				})
				return true
			})

		if firstErr != nil {
			return nil, firstErr
		}
	}

	if !fetchedAny {
		return nil, fmt.Errorf("no fragments fetched — the site layout may have changed")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no items found — the site layout may have changed")
	}
	return items, nil
}

// ecbMopoDate prefers the machine-readable isoDate attribute, falling back to
// the rendered date text
func ecbMopoDate(s *goquery.Selection) (time.Time, error) {
	if iso, _ := s.Attr("isoDate"); iso != "" {
		if parsed, err := time.Parse("2006-01-02", iso); err == nil {
			return parsed, nil
		}
	}
	text := strings.TrimSpace(s.Find("div.date").First().Text())
	if text == "" {
		return time.Time{}, fmt.Errorf("no usable date")
	}
	parsed, err := time.Parse("2 January 2006", text)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date %q: %w", text, err)
	}
	return parsed, nil
}

// articleBody fetches the decision page and pulls out its body. Best-effort:
// losing it costs a poorer item, not the feed.
func (p ECBMopoParser) articleBody(url string) *rss.CDATA {
	doc, err := fetchDocument(p.httpClient, url)
	if err != nil {
		log.Printf("%s: fetch %s: %v", p.Name(), url, err)
		return nil
	}

	content, err := articleHTML(doc, ecbMopoContent, ecbBaseURL, ecbMopoJunk)
	if err != nil {
		log.Printf("%s: content for %s: %v", p.Name(), url, err)
	}
	return rss.NewCDATA(content)
}
