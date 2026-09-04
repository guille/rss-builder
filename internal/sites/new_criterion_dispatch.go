package sites

import (
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

	var stream feedlyStream
	if err := fetchJSON(p.httpClient, apiURL, &stream); err != nil {
		return nil, err
	}

	var items []rss.Item
	for _, it := range stream.Items {
		if len(it.Alternate) == 0 {
			continue
		}
		link := it.Alternate[0].Href

		content := strings.TrimSpace(it.Content.Content)
		desc := strings.TrimSpace(it.Summary.Content)
		if desc == "" {
			desc = content
		}

		pub := time.UnixMilli(it.Published).UTC().Format(rss.PubDateFormat)

		items = append(items, rss.Item{
			Title:       strings.TrimSpace(it.Title),
			Link:        link,
			Description: desc,
			GUID:        rss.NewGUID(link),
			PubDate:     pub,
			Content:     rss.NewCDATA(content),
		})
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no items found — the feedly response may have changed")
	}
	return items, nil
}
