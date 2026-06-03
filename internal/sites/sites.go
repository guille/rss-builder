// Package sites implements feed parsers that scrape articles from various websites
package sites

import (
	"go.guillerg.dev/rss-builder/internal/rss"
)

type Parser interface {
	Name() string
	URL() string
	Fetch() ([]rss.Item, error)
}
