package sites

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.guillerg.dev/rss-builder/internal/rss"
)

const newCriterionDispatchFeedURL = "https://newcriterion.com/posts/dispatch/feed/"

// New Criterion puts their /feed behind Cloudflare. Until they fix that, using Feedly's public API
// as a bypass/workaround
type NewCriterionDispatchParser struct {
	httpClient *http.Client
}

func (NewCriterionDispatchParser) Name() string { return "The New Criterion Dispatch" }
func (NewCriterionDispatchParser) URL() string  { return newCriterionDispatchFeedURL }

type feedlyStream struct {
	Title string       `json:"title"`
	Items []feedlyItem `json:"items"`
}

type feedlyItem struct {
	Title     string        `json:"title"`
	Author    string        `json:"author"`
	Published int64         `json:"published"`
	Alternate []feedlyLink  `json:"alternate"`
	Summary   feedlyContent `json:"summary"`
	Content   feedlyContent `json:"content"`
}

type feedlyLink struct {
	Href string `json:"href"`
	Type string `json:"type"`
}

type feedlyContent struct {
	Content string `json:"content"`
}

func (p NewCriterionDispatchParser) Fetch() ([]rss.Item, error) {
	streamID := "feed/" + url.QueryEscape(newCriterionDispatchFeedURL)
	apiURL := "https://cloud.feedly.com/v3/streams/contents?streamId=" + streamID + "&count=20"

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "rss-builder")

	res, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", res.StatusCode)
	}

	var stream feedlyStream
	if err := json.NewDecoder(res.Body).Decode(&stream); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	var items []rss.Item
	for _, it := range stream.Items {
		if len(it.Alternate) == 0 {
			continue
		}
		link := it.Alternate[0].Href

		desc := strings.TrimSpace(it.Summary.Content)
		if desc == "" {
			desc = strings.TrimSpace(it.Content.Content)
		}

		pub := time.UnixMilli(it.Published).UTC().Format(rss.PubDateFormat)

		items = append(items, rss.Item{
			Title:       strings.TrimSpace(it.Title),
			Link:        link,
			Description: desc,
			GUID:        rss.NewGUID(link),
			PubDate:     pub,
		})
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no items found — the feedly response may have changed")
	}
	return items, nil
}
