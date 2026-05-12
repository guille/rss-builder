// Package sites implements feed parsers that scrape articles from various websites
package sites

import (
	"github.com/guille/rss-builder/internal/rss"
)

type Parser interface {
	Name() string
	URL() string
	Fetch() ([]rss.Item, error)
}
